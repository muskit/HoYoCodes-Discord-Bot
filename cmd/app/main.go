package main

import (
	"flag"
	"log/slog"

	"github.com/muskit/hoyocodes-discord-bot/internal/bot"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
)

func main() {
	dbgFlag := flag.Bool("debug", false, "enable debug output")
	flag.Parse()

	if *dbgFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	db.Init()
	bot.RunBot()
}