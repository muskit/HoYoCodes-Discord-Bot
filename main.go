package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/muskit/hoyocodes-discord-bot/pkg/db"
)

func TestDB() {
	println("[ADMIN ROLES]")
	roles, err := db.GetAdminRoles()
	if err != nil {
		return
	}

	for _, id := range roles {
		println(id)
	}
	println()

	println("Is 123 an admin role?")
	res, err := db.IsAdminRole(123)
	if err != nil {
		fmt.Errorf("Error checking if is admin: %v\n", err)
	}
	println(res)
	println()

	println("Is 853531051594481715 an admin role?")
	res, err = db.IsAdminRole(853531051594481715)
	if err != nil {
		fmt.Errorf("Error checking if is admin: %v\n", err)
	}
	println(res)
	println()
}

func main() {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("could not load .env: %s", err)
	}

	TestDB()

	// scraper.ScrapeHI3()
	// scraper.ScrapeGI()
	// scraper.ScrapeHSR()
	// scraper.ScrapeHSRLive()
	// scraper.ScrapeZZZ()
	
	RunBot()
}