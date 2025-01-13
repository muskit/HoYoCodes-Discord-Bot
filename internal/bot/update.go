package bot

import (
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
	"github.com/muskit/hoyocodes-discord-bot/internal/scraper"
)

var UpdatingMutex sync.Mutex

type CodeChanges struct {
	Added []string
	Removed []string
}

func UpdateLoop(session *discordgo.Session, waitFor time.Duration) {
	for {
		slog.Info("Beginning update loop...")
		UpdatingMutex.Lock()
		updateCodesDB()
		updateTickers(session)
		notifySubscribers(session)
		UpdatingMutex.Unlock()

		nextUpdateTime := time.Now().Add(waitFor)
		slog.Info("Finished update loop!")
		slog.Info(fmt.Sprintf("Sleeping for %v until %v", waitFor, nextUpdateTime.Format(time.Kitchen)))
		<-time.After(time.Until(nextUpdateTime))
	}
}

func updateCodesDB() map[string]CodeChanges {
	slog.Debug("--- [Update Code Database] ---")
	for _, cfg := range scraper.Configs {
		checkTime := time.Now()
		var updateTime time.Time

		livestream := false
		for i := 0; i < 2; i++ {
			codes, updateTimeStr := scraper.ScrapePJT(cfg)
			updateTime, _ = time.Parse(time.RFC3339, updateTimeStr)
			for code, desc := range codes {
				if err := db.AddCode(code, cfg.Game, desc, livestream, updateTime); err != nil {
					if !db.IsDuplicateErr(err) {
						log.Fatalf("Error adding code to database: %v\n", err)
					}
				}
			}
			cfg.Heading = "livestream codes"
			livestream = true
		}

		if err := db.SetScrapeStats(cfg.Game, updateTime, checkTime); err != nil {
			log.Fatalf("Error updating scrape stats for %v: %v", cfg.Game, err)
		}
	}
	// TODO
	return map[string]CodeChanges{}
}

func updateTickersRoutine(s *discordgo.Session, game string) {
	tickers, err := db.GetEmbeds(game)
	if err != nil {
		log.Fatalf("Error getting tickers: %v", err)
	}

	slog.Debug("", "game", game)
	content := createCodePrint(game, true)
	slog.Debug("", "length", len(content))
	slog.Debug(fmt.Sprintf("Content is as follows:\n%v", content))

	for _, emb := range tickers {
		channelID, messageID := emb[0], emb[1]
		
		// update ticker
		edit := discordgo.MessageEdit{
			Channel: channelID,
			ID: messageID,
			Content: &content,
			Embeds: &[]*discordgo.MessageEmbed{}, // delete old embed if still there
		}
		// if _, err := s.ChannelMessageEdit(channelID, messageID, content, ); err != nil {
		if _, err := s.ChannelMessageEditComplex(&edit); err != nil {
			if strings.Contains(err.Error(), "HTTP 404 Not Found") {
				// message no longer exists -- delete from db
				msgNum, _ := strconv.ParseUint(messageID, 10, 64)
				err := db.RemoveTicker(msgNum)
				if err != nil {
					slog.Error(fmt.Sprintf("404'd removing ticker from db during update: %s", err))
				}
			} else {
				log.Fatalf("Error updating ticker: %v", err)
			}
		}
	}
}

func updateTickers(session *discordgo.Session) {
	slog.Debug("--- [Update Tickers] ---")
	for _, ch := range GameChoices {
		game := ch.Name
		updateTickersRoutine(session, game)
	}
}

func notifySubscribers(session *discordgo.Session) {
	slog.Debug("--- [Notify Subscribed Channels] ---")
	slog.Warn("TODO: update.notifySubscribers")
}