package main

import (
	"log"

	"github.com/joho/godotenv"
	_ "github.com/muskit/hoyocodes-discord-bot/pkg/db"
	"github.com/muskit/hoyocodes-discord-bot/pkg/scraper"
)

func main() {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("could not load .env: %s", err)
	}

	scraper.ScrapeHI3()
	scraper.ScrapeGI()
	scraper.ScrapeHSR()
	scraper.ScrapeHSRLive()
	scraper.ScrapeZZZ()
	
	RunBot()
}