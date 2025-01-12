package bot

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/muskit/hoyocodes-discord-bot/pkg/db"
	"github.com/muskit/hoyocodes-discord-bot/pkg/scraper"
)

// - Scrape for new codes and update DB
// - Execute Discord tasks (concurrency within limits!)
func UpdateLoop(session *discordgo.Session) {
	for {
		updateCodesDB()
		updateEmbeds(session)
		notifySubscribers(session)

		nextUpdateTime := time.Now().Add(4*time.Hour)
		log.Printf("Running next update loop 4hrs from now at %v", nextUpdateTime.Format(time.Kitchen))
		<-time.After(4*time.Hour)
	}
}

func updateCodesDB() {
	for _, cfg := range scraper.Configs {
		livestream := false
		for i := 0; i < 2; i++ {
			codes, timeStr := scraper.ScrapeGame(cfg)
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
	for _, ch := range GameChoices {
		game := ch.Name
		go UpdateGameEmbeds(session, game)
	}
}

func notifySubscribers(session *discordgo.Session) {
	log.Println("TODO: update.notifySubscribers")
}