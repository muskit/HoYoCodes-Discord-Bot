package db

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"
)

func AddCode(code string, game string, description string, livestream bool, foundTime time.Time) error {
	_, err := DBScraper.Exec("INSERT INTO Codes SET code = ?, game = ?, description = ?, is_livestream = ?, added = ?", code, game, description, livestream, foundTime)
	return err
}

func RemoveCode(code string) error {
	_, err := DBScraper.Exec("DELETE FROM FROM Codes WHERE code = ?", code)
	return err
}

func GetCodes(game string, recent bool, livestream bool) [][]string {
	var sels *sql.Rows
	var err error

	if recent {
		// TODO: figure criteria for if a code is "recent"
		return [][]string{}
	} else {
		sels, err = DBScraper.Query("SELECT code, description FROM Codes WHERE game = ? AND is_livestream = ? ORDER BY added ASC", game, livestream)
	}
	if err != nil {
		log.Fatalf("Error trying to get codes for %v: %v", game, err)
	}

	codes := [][]string{}
	var code string
	var description string
	for sels.Next() {
		if err = sels.Err(); err != nil {
			log.Fatalf("Error reading code row for %v: %v", game, err)
		}
		sels.Scan(&code, &description)
		codes = append(codes, []string{code, description})
	}

	return codes
}

func SetScrapeStats(game string, updated time.Time, checked time.Time) error {
	row := DBScraper.QueryRow("SELECT game FROM ScrapeStats WHERE game = ?", game)
	
	var z string // temp unused var for existence checking
	err := row.Scan(&z)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Info(fmt.Sprintf("Adding %v to ScrapeStats", game))
			_, err := DBScraper.Exec("INSERT INTO ScrapeStats SET game = ?", game)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	_, err = DBScraper.Exec("UPDATE ScrapeStats SET updated = ?, checked = ? WHERE game = ?", updated, checked, game)
	return err
}

func GetScrapeStats(game string) (time.Time, time.Time, error) {
	var checked time.Time
	var updated time.Time
	row := DBScraper.QueryRow("SELECT checked, updated FROM ScrapeStats WHERE game = ?", game)
	err := row.Scan(&checked, &updated)
	return checked, updated, err
}