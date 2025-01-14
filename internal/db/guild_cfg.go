package db

import (
	"github.com/hashicorp/go-set/v3"
)

type Subscription struct {
	ChannelID uint64
	Active bool
	AnnounceAdds bool
	AnnounceRems bool
}

func CreateSubscription(channelID uint64, guildID uint64, additions bool, removals bool) error {
	_, err := DBCfg.Exec("INSERT INTO Subscriptions SET channel_id = ?, guild_id = ?, announce_additions = ?, announce_removals = ?", channelID, guildID, additions, removals)
	return err
}

func UpdateSubscription(channelID uint64, additions bool, removals bool) error {
	_, err := DBCfg.Exec("UPDATE Subscriptions SET announce_additions = ?, announce_removals = ?, active = true WHERE channel_id = ?", additions, removals, channelID)
	return err
}

func DeactivateSubscription(channelID uint64) error {
	// _, err := DBCfg.Exec("DELETE FROM Subscriptions WHERE channel_id = ?", channelID)
	_, err := DBCfg.Exec("UPDATE Subscriptions SET active = false WHERE channel_id = ?", channelID)
	return err
}

func GetSubscription(channelID uint64) (*Subscription, error) {
	var announceAdds bool
	var announceRems bool
	var active bool
	s := DBCfg.QueryRow("SELECT announce_additions, announce_removals, active FROM Subscriptions WHERE channel_id = ?", channelID)
	if err := s.Scan(&announceAdds, &announceRems, &active); err != nil {
		return nil, err
	}

	return &Subscription{
		ChannelID: channelID,
		Active: active,
		AnnounceAdds: announceAdds,
		AnnounceRems: announceRems,
	}, nil
}

func GetSubscriptionsFromGuild(guildID uint64) ([]Subscription, error) {
	result := []Subscription{}

	sels, err := DBCfg.Query("SELECT channel_id, active, announce_additions, announce_removals FROM Subscriptions WHERE guild_id = ?", guildID)
	if err != nil {
		return result, err
	}

	var channel_id uint64
	var active bool
	var announceAdds bool
	var announceRems bool
	for sels.Next() {
		sels.Scan(&channel_id, &active, &announceAdds, &announceRems)
		result = append(result, Subscription{
			ChannelID: channel_id,
			Active: active,
			AnnounceAdds: announceAdds,
			AnnounceRems: announceRems,
		})
	}
	err = sels.Err()
	if err != nil {
		return result, err
	}

	return result, nil
}

func GetGameSubscriptions(game string) ([]Subscription, error) {
	result := []Subscription{}

	filteredQ := `
	SELECT Subscriptions.channel_id, active, announce_additions, announce_removals FROM Subscriptions
	JOIN SubscriptionGames ON SubscriptionGames.channel_id=Subscriptions.channel_id
	WHERE SubscriptionGames.game = ? AND active = TRUE;
	`
	sels, err := DBCfg.Query(filteredQ, game)
	if err != nil {
		return result, err
	}

	for sels.Next() {
		var channel_id uint64
		var active bool
		var announceAdds bool
		var announceRems bool

		sels.Scan(&channel_id, &active, &announceAdds, &announceRems)
		result = append(result, Subscription{
			ChannelID: channel_id,
			Active: active,
			AnnounceAdds: announceAdds,
			AnnounceRems: announceRems,
		})
	}
	if sels.Err() != nil {
		return result, sels.Err()
	}

	nofilterQ := `
	SELECT Subscriptions.channel_id, active, announce_additions, announce_removals FROM Subscriptions
	LEFT JOIN SubscriptionGames ON SubscriptionGames.channel_id = Subscriptions.channel_id
	WHERE game IS NULL AND Subscriptions.active = TRUE;
	`
	sels, err = DBCfg.Query(nofilterQ)
	if err != nil {
		return result, err
	}

	for sels.Next() {
		var channel_id uint64
		var active bool
		var announceAdds bool
		var announceRems bool

		sels.Scan(&channel_id, &active, &announceAdds, &announceRems)
		result = append(result, Subscription{
			ChannelID: channel_id,
			Active: active,
			AnnounceAdds: announceAdds,
			AnnounceRems: announceRems,
		})
	}
	if sels.Err() != nil {
		return result, sels.Err()
	}

	return result, nil
}

func AddPingRole(channelID uint64, pingRole uint64) error {
	_, err := DBCfg.Exec("INSERT INTO SubscriptionPingRoles SET channel_id = ?, role_id = ?", channelID, pingRole)
	return err
}

func RemovePingRole(channelID uint64, pingRole uint64) error {
	_, err := DBCfg.Exec("DELETE FROM SubscriptionPingRoles WHERE channel_id = ? AND role_id = ?", channelID, pingRole)
	return err
}

func GetPingRoles(channelID uint64) ([]uint64, error) {
	rows, err := DBCfg.Query("SELECT role_id FROM SubscriptionPingRoles WHERE channel_id = ?", channelID)
	if err != nil {
		return nil, err
	}

	results := []uint64{}
	var val uint64
	for rows.Next() {
		rows.Scan(&val)
		results = append(results, val)
	}
	if rows.Err() != nil {
		return results, rows.Err()
	}

	return results, nil
}

func SetGameFilters(channelID uint64, games *set.Set[string]) error {
	if _, err := DBCfg.Exec("DELETE FROM SubscriptionGames WHERE channel_id = ?", channelID); err != nil {
		return err
	}
	
	for _, game := range games.Slice() {
		_, err := DBCfg.Exec("INSERT INTO SubscriptionGames SET channel_id = ?, game = ?", channelID, game)
		if err != nil && !IsDuplicateErr(err) {
			return err
		}
	}
	return nil
}

func GetSubscriptionGames(channelID uint64) ([]string, error) {
	rows, err := DBCfg.Query("SELECT game FROM SubscriptionGames WHERE channel_id = ?", channelID)
	if err != nil {
		return nil, err
	}

	results := []string{}
	var val string
	for rows.Next() {
		
		rows.Scan(&val)
		results = append(results, val)
	}
	if rows.Err() != nil {
		return results, rows.Err()
	}

	return results, nil
}

func AddEmbed(messageID uint64, game string, channelID uint64) error {
	_, err := DBCfg.Exec("INSERT INTO Embeds SET message_id = ?, game = ?, channel_id = ?", messageID, game, channelID)
	return err
}

func RemoveTicker(messageID uint64) error {
	_, err := DBCfg.Exec("DELETE FROM Embeds WHERE message_id = ?", messageID)
	return err
}

// elem[0] = channel ID
// elem[1] = message ID
func GetEmbeds(game string) ( [][]string, error ) {
	ret := [][]string{}
	sels, err := DBCfg.Query("SELECT channel_id, message_id FROM Embeds WHERE game = ?", game)
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
