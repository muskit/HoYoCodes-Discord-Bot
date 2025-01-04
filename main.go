package main

import (
	"log"

	"github.com/muskit/hoyocodes-discord-bot/internal"
	"github.com/muskit/hoyocodes-discord-bot/pkg/db"
)

func TestDB() {
	println("[ADMIN ROLES]")
	entries, err := db.GetGuildAdmins()
	if err != nil {
		return
	}

	for _, entry := range entries {
		log.Printf("%v : %v", entry.GuildID, entry.RoleID)
	}
	println()
}

func main() {
	TestDB()

	// scraper.ScrapeHI3()
	// scraper.ScrapeGI()
	// scraper.ScrapeHSR()
	// scraper.ScrapeHSRLive()
	// scraper.ScrapeZZZ()
	
	internal.RunBot()
}