package main

import (
	"log"

	"github.com/joho/godotenv"
)

func RunBot() {
	err := godotenv.Load()
	if err != nil {
		// error'd trying to read .env!
		log.Fatal("could not load .env!")
	}
}