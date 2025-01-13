package bot

import (
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/muskit/hoyocodes-discord-bot/pkg/db"
	"github.com/muskit/hoyocodes-discord-bot/pkg/scraper"
)

var UpdatingMutex sync.Mutex

// - Scrape for new codes and update DB
// - Execute Discord tasks (concurrency within limits!)
func UpdateLoop(session *discordgo.Session, waitFor time.Duration) {
	for {
		UpdatingMutex.Lock()
		updateCodesDB()
		updateEmbeds(session)
		notifySubscribers(session)
		UpdatingMutex.Unlock()

		nextUpdateTime := time.Now().Add(waitFor)
		slog.Info(fmt.Sprintf("Running next update loop %v from now at %v", waitFor, nextUpdateTime.Format(time.Kitchen)))
		<-time.After(4*time.Hour)
	}
}

func updateCodesDB() {
	slog.Debug("--- [Update Code Database] ---")
	for _, cfg := range scraper.Configs {
		livestream := false
		for i := 0; i < 2; i++ {
			codes, timeStr := scraper.ScrapePJT(cfg)
			time, _ := time.Parse(time.RFC3339, timeStr)
			for code, desc := range codes {
				if err := db.AddCode(code, cfg.Game, desc, livestream, time); err != nil {
					if !db.IsDuplicateErr(err) {
						log.Fatalf("Error adding code to database: %v\n", err)
					}
				}
			}
			cfg.Heading = "livestream codes"
			livestream = true
		}
	}
}

func updateEmbeds(session *discordgo.Session) {
	slog.Debug("--- [Update Embeds] ---")
	for _, ch := range GameChoices {
		game := ch.Name
		go UpdateGameEmbeds(session, game)
	}
}

func notifySubscribers(session *discordgo.Session) {
	slog.Debug("--- [Notify Subscribed Channels] ---")
	slog.Warn("TODO: update.notifySubscribers")
}