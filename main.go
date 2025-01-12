package main

import (
	"github.com/muskit/hoyocodes-discord-bot/pkg/bot"
	"github.com/muskit/hoyocodes-discord-bot/pkg/update"
)

func main() {

	// scraper.ScrapeHI3()
	// scraper.ScrapeGI()
	// scraper.ScrapeHSR()
	// scraper.ScrapeHSRLive()
	// scraper.ScrapeZZZ()

	update.UpdateRoutine()

	bot.RunBot()
}