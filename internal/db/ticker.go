package db

type Ticker struct {
	MessageID string
	Game string
	GuildID string
	ChannelID string
}

func AddTicker(messageID string, game string, channelID string, guildID string) error {
	_, err := DBCfg.Exec("INSERT INTO Tickers SET message_id = ?, game = ?, channel_id = ?, guild_id = ?", messageID, game, channelID, guildID)
	return err
}

func RemoveTicker(messageID string) error {
	_, err := DBCfg.Exec("DELETE FROM Tickers WHERE message_id = ?", messageID)
	return err
}

// returns a slice of channelID,messageID pairs
func GetGameTickers(game string) ( [][]string, error ) {
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

// returns a slice of channelID,messageID pairs
func GetGuildTickers(guildID string) ( []Ticker, error ) {
	ret := []Ticker{}
	sels, err := DBCfg.Query("SELECT game, channel_id, message_id FROM Tickers WHERE guild_id = ?", guildID)
	if err != nil {
		return ret, err
	}

	for sels.Next() {
		game := ""
		channelID := ""
		messageID := ""
		sels.Scan(&game, &channelID, &messageID)
		ret = append(ret, Ticker{
			MessageID: messageID,
			Game: game,
			GuildID: guildID,
			ChannelID: channelID,
		})
	}
	if err = sels.Err(); err != nil {
		return ret, err
	}

	return ret, nil
}