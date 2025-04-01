package main

import (
	"log/slog"

	"github.com/muskit/hoyocodes-discord-bot/internal/scraper"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	for _, game := range scraper.Configs {
		scraper.ScrapePJT(game)
	}
}