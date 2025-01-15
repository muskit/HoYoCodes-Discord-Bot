package db

import (
	"github.com/hashicorp/go-set/v3"
)

type Subscription struct {
	ChannelID string
	Active bool
	AnnounceAdds bool
	AnnounceRems bool
}

func CreateSubscription(channelID string, guildID string, additions bool, removals bool) error {
	_, err := DBCfg.Exec("INSERT INTO Subscriptions SET channel_id = ?, guild_id = ?, announce_additions = ?, announce_removals = ?", channelID, guildID, additions, removals)
	return err
}

func UpdateSubscription(channelID string, additions bool, removals bool) error {
	_, err := DBCfg.Exec("UPDATE Subscriptions SET announce_additions = ?, announce_removals = ?, active = true WHERE channel_id = ?", additions, removals, channelID)
	return err
}

func DeactivateSubscription(channelID string) error {
	// _, err := DBCfg.Exec("DELETE FROM Subscriptions WHERE channel_id = ?", channelID)
	_, err := DBCfg.Exec("UPDATE Subscriptions SET active = false WHERE channel_id = ?", channelID)
	return err
}

func DeleteSubscription(channelID string) error {
	_, err := DBCfg.Exec("DELETE FROM Subscriptions WHERE channel_id = ?", channelID)
	return err
}

func GetSubscription(channelID string) (*Subscription, error) {
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

func GetGuildSubscriptions(guildID string) ([]Subscription, error) {
	result := []Subscription{}

	sels, err := DBCfg.Query("SELECT channel_id, active, announce_additions, announce_removals FROM Subscriptions WHERE guild_id = ?", guildID)
	if err != nil {
		return result, err
	}

	var channel_id string
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
		var channel_id string
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
		var channel_id string
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

func AddPingRole(channelID string, pingRole string) error {
	_, err := DBCfg.Exec("INSERT INTO SubscriptionPingRoles SET channel_id = ?, role_id = ?", channelID, pingRole)
	return err
}

func RemovePingRole(channelID string, pingRole string) error {
	_, err := DBCfg.Exec("DELETE FROM SubscriptionPingRoles WHERE channel_id = ? AND role_id = ?", channelID, pingRole)
	return err
}

func GetPingRoles(channelID string) ([]string, error) {
	rows, err := DBCfg.Query("SELECT role_id FROM SubscriptionPingRoles WHERE channel_id = ?", channelID)
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

func SetGameFilters(channelID string, games *set.Set[string]) error {
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

func GetSubscriptionGames(channelID string) ([]string, error) {
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
