package bot

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
	"github.com/muskit/hoyocodes-discord-bot/internal/scraper"
	"github.com/muskit/hoyocodes-discord-bot/pkg/consts"
	"github.com/muskit/hoyocodes-discord-bot/pkg/util"
)

var UpdatingMutex = sync.Mutex{}

type CodeChanges struct {
	Added [][]string
	Removed [][]string
}

func UpdateRoutine(session *discordgo.Session, interruptCh chan<-os.Signal) {
	for {
		slog.Info("---------- Start update loop ----------")

		// check session integrity
		if _, err := session.User("@me"); err != nil {
			log.Fatalf("Error getting me: %v", err)
		}
		// check DB operability
		if err := db.CheckDBs(); err != nil {
			slog.Error("One of the DBs error'd on ping!")
			slog.Error(err.Error())
			interruptCh<-nil
			return
		}

		UpdatingMutex.Lock()
		changes := updateCodesDB()
		updateTickers(session)
		notifySubscribers(session, changes, false)
		UpdatingMutex.Unlock()

		nextUpdateTime := time.Now().Add(consts.UpdateInterval)
		slog.Info("Finished update loop!")
		slog.Info(fmt.Sprintf("Sleeping for %v until %v", consts.UpdateInterval, nextUpdateTime.Format(time.Kitchen)))
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

	if len(changes) > 0 {
		for game, chg := range changes {
			slog.Info(fmt.Sprintf("%v: %v added, %v removed", game, len(chg.Added), len(chg.Removed)))
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
	}

	return changes
}

func updateTickers(session *discordgo.Session) {
	slog.Info("Update Tickers")
	
	for _, g := range consts.Games {
		game := g
		UpdateEmbedTickersGame(session, game)
		// UpdateTextTickersGame(session, game)
	}
}

func ShouldNotify(sub db.Subscription, chg CodeChanges) bool {
	if (!sub.AnnounceAdds && !sub.AnnounceRems) ||
		(len(chg.Added) == 0 && len(chg.Removed) == 0) {
		return false
	}

	return (sub.AnnounceAdds && len(chg.Added) > 0) ||
		(sub.AnnounceRems && len(chg.Removed) > 0)
}

func notifyContent(game string, chgs CodeChanges) string {
	_, updateTime, err := db.GetScrapeTimes(game)
	if err != nil {
		log.Fatalf("Error getting scrape times for %v: %v", game, err)
	}

	content := ""

	content += fmt.Sprintf("## Codes updated for %v!\n", game)
	if len(chgs.Added) > 0 {
		content += "**NEW:**\n"
		content += util.CodeListing(chgs.Added, &game) + "\n"
	}
	if len(chgs.Removed) > 0 {
		content += "**REMOVED:**\n"
		content += util.CodeListing(chgs.Removed, nil) + "\n"
	}

	if link, exists := consts.RedeemURL[game]; exists {
		content += fmt.Sprintf("\n[Redemption page](<%v>)\n", link)
	}

	footer := fmt.Sprintf("-# [source](<%v>) updated <t:%v:R>.", consts.ArticleURL[game], updateTime.Unix())
	content += footer

	return content
}

func notifySubscribers(session *discordgo.Session, gameChanges map[string]*CodeChanges, dryrun bool) {
	if len(gameChanges) == 0 {
		slog.Info("No changes to notify subscribers of")
		return
	}

	slog.Info("Notify Subscribed Channels")


	for game, chgs := range gameChanges {
		subscriptions, err := db.GetGameSubscriptions(game)
		if err != nil {
			log.Fatalf("Error getting subscriptions for %v: %v", game, err)
		}

		// notification message for game
		content := notifyContent(game, *chgs)

		for _, sub := range subscriptions {
			if !ShouldNotify(sub, *chgs) {
				continue
			}

			subMsg := content

			// prepend role mentions
			roles, err := db.GetPingRoles(sub.ChannelID)
			if err != nil {
				log.Fatalf("Error getting ping roles for subscription %v: %v", sub.ChannelID, err)
			}
			if len(roles) > 0 {
				mentions := "||"
				for _, r := range roles {
					mentions += fmt.Sprintf("<@&%v> ", r)
				}
				subMsg = content + strings.Trim(mentions, " ") + "||\n"
			}

			if dryrun { 
				slog.Debug(fmt.Sprintf("for %v:\n%s", sub.ChannelID, subMsg))
				continue
			}

			if _, err := session.ChannelMessageSend(sub.ChannelID, subMsg); err != nil {
				if strings.Contains(err.Error(), "HTTP 403") {
					// Forbidden: no permission to post
					slog.Warn(fmt.Sprintf("HTTP Forbidden 403 sending subscription notification: %v", err))
				} else if strings.Contains(err.Error(), "HTTP 404") {
					// Not found: channel or message
					// TODO: delete subscription from DB?
					slog.Warn(fmt.Sprintf("HTTP Not Found 404 sending subscription notification: %v", err))
				} else {
					log.Fatalf("Error sending subscription notification to %v: %v", sub.ChannelID, err)
				}
			}
		}
	}
}