package main

import (
	"flag"
	"log/slog"

	"github.com/muskit/hoyocodes-discord-bot/internal/bot"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
)

func main() {
	// desc := "30 crystals, one Hi♪ Love Elf♥ trial card (three-day), and 2,888 asterite"
	// filtered := util.ReplaceNonAlphanumeric(desc)
	// fmt.Printf("before=%s\nfiltered=%s\n", desc, filtered)
	// return

	dbgFlag := flag.Bool("debug", false, "enable debug output")
	flag.Parse()

	if *dbgFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	db.Init()
	bot.RunBot()
}