package db

import (
	"database/sql"
	"log"
	"time"
)

func AddCode(code string, game string, description string, livestream bool, foundTime time.Time) error {
	_, err := DBScraper.Exec("INSERT INTO Codes SET code = ?, game = ?, description = ?, is_livestream = ?, found = ?", code, game, description, livestream, foundTime)
	return err
}

func GetCodes(game string, recent bool, livestream bool) [][]string {
	var sels *sql.Rows
	var err error

	if recent {
		// TODO: figure criteria for if a code is "recent"
		return [][]string{}
	} else {
		sels, err = DBScraper.Query("SELECT code, description FROM Codes WHERE game = ? AND is_livestream = ? ORDER BY found ASC", game, livestream)
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