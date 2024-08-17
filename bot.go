package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
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
	fmt.Println(discord)
	fmt.Println(err)
}
