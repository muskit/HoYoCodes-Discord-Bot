package main

import "github.com/muskit/hoyocodes-discord-bot/pkg/scraper"

func main() {
	scraper.ScrapeHI3()
	scraper.ScrapeGI()
	scraper.ScrapeHSR()
	scraper.ScrapeHSRLive()
	scraper.ScrapeZZZ()
	
	RunBot()
}