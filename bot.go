package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/bwmarrin/discordgo"
)

func RunBot() {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		// error'd trying to read .env!
		log.Fatal("could not load .env!")
	}

	// init bot
	discord, err := discordgo.New("Bot " + "authentication token")
}
