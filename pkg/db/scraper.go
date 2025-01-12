package db

import "time"

func AddCode(code string, game string, description string, livestream bool, foundTime time.Time) error {
	_, err := DBScraper.Exec("INSERT INTO Codes SET code = ?, game = ?, description = ?, is_livestream = ?, found = ?", code, game, description, livestream, foundTime)
	return err
}