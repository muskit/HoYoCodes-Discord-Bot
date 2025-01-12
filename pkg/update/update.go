package update

import (
	"log"
	"time"

	"github.com/muskit/hoyocodes-discord-bot/pkg/db"
	"github.com/muskit/hoyocodes-discord-bot/pkg/scraper"
)

// - Scrape for new codes
// - Update Codes DB
// - Execute Discord tasks (concurrency within limits!)
func UpdateRoutine() {
	updateDB()
}

func updateDB() {
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