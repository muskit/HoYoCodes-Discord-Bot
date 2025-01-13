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

func UpdateLoop(session *discordgo.Session, waitFor time.Duration) {
	for {
		UpdatingMutex.Lock()
		slog.Info("Running update loop...")
		updateCodesDB()
		updateEmbeds(session)
		notifySubscribers(session)
		UpdatingMutex.Unlock()

		nextUpdateTime := time.Now().Add(waitFor)
		slog.Info(fmt.Sprintf("Finished update loop! Running again %v from now at %v", waitFor, nextUpdateTime.Format(time.Kitchen)))
		<-time.After(4*time.Hour)
	}
}

func updateCodesDB() {
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
}

func updateEmbedsRoutine(s *discordgo.Session, game string) {
	embeds, err := db.GetEmbeds(game)
	if err != nil {
		log.Fatalf("Error getting embeds to update: %v", err)
	}

	embedContent := createEmbed(game, true)

	for _, emb := range embeds {
		channelID, messageID := emb[0], emb[1]
		_, err = s.ChannelMessageEditEmbed(channelID, messageID, embedContent)
		if err != nil {
			if strings.Contains(err.Error(), "HTTP 404 Not Found") {
				// embed message no longer exists -- delete from db
				msgNum, _ := strconv.ParseUint(messageID, 10, 64)
				err := db.RemoveEmbed(msgNum)
				if err != nil {
					slog.Error(fmt.Sprintf("404'd removing embed from db during update: %s", err))
				}
			} else {
				log.Fatalf("Error updating embed: %v", err)
			}
		}
	}
}

func updateEmbeds(session *discordgo.Session) {
	slog.Debug("--- [Update Embeds] ---")
	for _, ch := range GameChoices {
		game := ch.Name
		go updateEmbedsRoutine(session, game)
	}
}

func notifySubscribers(session *discordgo.Session) {
	slog.Debug("--- [Notify Subscribed Channels] ---")
	slog.Warn("TODO: update.notifySubscribers")
}