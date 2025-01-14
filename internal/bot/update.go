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
	Added [][]string
	Removed [][]string
}

func UpdateLoop(session *discordgo.Session, waitFor time.Duration) {
	for {
		slog.Info("Beginning update loop...")
		UpdatingMutex.Lock()
		changes := updateCodesDB()
		updateTickers(session)
		notifySubscribers(session, changes, false)
		UpdatingMutex.Unlock()

		nextUpdateTime := time.Now().Add(waitFor)
		slog.Info("Finished update loop!")
		slog.Info(fmt.Sprintf("Sleeping for %v until %v", waitFor, nextUpdateTime.Format(time.Kitchen)))
		<-time.After(time.Until(nextUpdateTime))
	}
}

func updateCodesDB() map[string]*CodeChanges {
	slog.Info("Update Codes Database")
	changes := map[string]*CodeChanges{}

	for _, cfg := range scraper.Configs {
		checkTime := time.Now()
		var updateTime time.Time
		pageCodes := []string{}

		livestream := false
		for i := 0; i < 2; i++ { // get w/o, then w/ livestream
			codes, updateTimeStr := scraper.ScrapePJT(cfg)
			updateTime, _ = time.Parse(time.RFC3339, updateTimeStr)
			for code, desc := range codes {
				pageCodes = append(pageCodes, code)
				if err := db.AddCode(code, cfg.Game, desc, livestream, updateTime); err != nil {
					if !db.IsDuplicateErr(err) {
						log.Fatalf("Error adding code to database: %v\n", err)
					}
				} else {
					// new code added
					slog.Debug("Found new code!", "game", cfg.Game, "code", code)
					if _, exists := changes[cfg.Game]; !exists {
						changes[cfg.Game] = &CodeChanges{}
					}
					changes[cfg.Game].Added = append(changes[cfg.Game].Added, []string{code, desc})
				}
			}
			// set for next check
			cfg.Heading = "livestream codes"
			livestream = true
		}


		// TODO: code removals
		removed, err := db.GetRemovedCodes(pageCodes, cfg.Game, true)
		if err != nil {
			log.Fatalf("Error getting removed codes for %v: %v", cfg.Game, err)
		}
		if len(removed) > 0 {
			if _, exists := changes[cfg.Game]; !exists {
				changes[cfg.Game] = &CodeChanges{}
			}
			for _, elem := range removed {
				code, desc := elem[0], elem[1]
				changes[cfg.Game].Removed = append(changes[cfg.Game].Removed, []string{code, desc})
			}
			
			if err := db.RemoveCodes(removed, cfg.Game); err != nil {
				log.Fatalf("Error deleting removed codes from db: %v", err)
			}
		}

		if err := db.SetScrapeTimes(cfg.Game, updateTime, checkTime); err != nil {
			log.Fatalf("Error updating scrape times for %v: %v", cfg.Game, err)
		}
	}

	// slog.Debug(fmt.Sprintf("%v", changes))
	for game, chg := range changes {
		slog.Debug(fmt.Sprintf("Changes for %v:", game))
		if len(chg.Added) > 0 {
			slog.Debug("Added:")
			for _, elem := range chg.Added {
				code, desc := elem[0], elem[1]
				slog.Debug("", "code", code, "desc", desc)
			}
		}
		if len(chg.Removed) > 0 {
			slog.Debug("Removed:")
			for _, elem := range chg.Removed {
				code, desc := elem[0], elem[1]
				slog.Debug("", "code", code, "desc", desc)
			}
		}
	}

	return changes
}

func updateTickersRoutine(s *discordgo.Session, game string) {
	tickers, err := db.GetTickers(game)
	if err != nil {
		log.Fatalf("Error getting tickers: %v", err)
	}

	slog.Debug("", "game", game)
	content := createCodePrint(game, true)
	slog.Debug("", "len(content)", len(content))
	// slog.Debug(fmt.Sprintf("Content is as follows:\n%v", content))

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
	slog.Info("Update Tickers")
	for _, ch := range GameChoices {
		game := ch.Name
		updateTickersRoutine(session, game)
	}
}

func notifySubscribers(session *discordgo.Session, changes map[string]*CodeChanges, dryrun bool) {
	slog.Info("Notify Subscribed Channels")

	for game, chg := range changes {
		checkTime, updateTime, err := db.GetScrapeTimes(game)
		if err != nil {
			log.Fatalf("Error getting scrape times for %v: %v", game, err)
		}

		subscriptions, err := db.GetGameSubscriptions(game)
		if err != nil {
			log.Fatalf("Error getting subscriptions for %v: %v", game, err)
		}

		for _, sub := range subscriptions {
			if !sub.AnnounceAdds && !sub.AnnounceRems ||
				(sub.AnnounceAdds && len(chg.Added) <= 0) ||
				(sub.AnnounceRems && len(chg.Removed) <= 0) {
				continue
			}

			content := ""

			roles, err := db.GetPingRoles(sub.ChannelID)
			if err != nil {
				log.Fatalf("Error getting ping roles for subscription %v: %v", sub.ChannelID, err)
			}

			if len(roles) > 0 {
				content = "||"
				for _, r := range roles {
					content += fmt.Sprintf("<@&%v> ", r)
				}
				content = strings.Trim(content, " ") + "||\n"
			}

			content += fmt.Sprintf("## Codes updated for %v!\n", game)
			if len(chg.Added) > 0 {
				content += "**NEW:**\n"
				content += CodeListing(chg.Added) + "\n"
			}
			if len(chg.Removed) > 0 {
				content += "**REMOVED:**\n"
				content += CodeListing(chg.Removed) + "\n"
			}

			if link, exists := redeemURL[game]; exists {
				content += fmt.Sprintf("\n[Redemption page](%v)\n", link)
			}

			footer := fmt.Sprintf("-# Checked <t:%v:R>; [source](<%v>) updated <t:%v:R>.\n", checkTime.Unix(), articleURL[game], updateTime.Unix())
			content += footer

			slog.Debug(fmt.Sprintf("for %v:\n%s", sub.ChannelID, content))
			if dryrun { continue }

			if _, err := session.ChannelMessageSend(strconv.FormatUint(sub.ChannelID, 10), content); err != nil {
				log.Fatalf("Error sending subscription notification to %v: %v", sub.ChannelID, err)
			}
		}
	}
}