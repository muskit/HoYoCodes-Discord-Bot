package db

func AddTicker(messageID uint64, game string, channelID uint64) error {
	_, err := DBCfg.Exec("INSERT INTO Tickers SET message_id = ?, game = ?, channel_id = ?", messageID, game, channelID)
	return err
}

func RemoveTicker(messageID uint64) error {
	_, err := DBCfg.Exec("DELETE FROM Tickers WHERE message_id = ?", messageID)
	return err
}

// returns a slice of channelID,messageID pairs
func GetTickers(game string) ( [][]string, error ) {
	ret := [][]string{}
	sels, err := DBCfg.Query("SELECT channel_id, message_id FROM Tickers WHERE game = ?", game)
	if err != nil {
		return ret, err
	}

	for sels.Next() {
		channelID := ""
		messageID := ""
		sels.Scan(&channelID, &messageID)
		ret = append(ret, []string{channelID, messageID})
	}
	if err = sels.Err(); err != nil {
		return ret, err
	}

	return ret, nil
}