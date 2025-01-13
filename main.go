package main

import (
	"flag"
	"log/slog"

	"github.com/muskit/hoyocodes-discord-bot/pkg/bot"
)

func main() {
	// var debugFlag bool
	dbgFlag := flag.Bool("debug", false, "enable debug output")
	flag.Parse()

	if *dbgFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	bot.RunBot()
}