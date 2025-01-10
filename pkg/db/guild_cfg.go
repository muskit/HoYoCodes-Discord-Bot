package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/hashicorp/go-set/v3"
)

type Subscription struct {
	ChannelID uint64
	Active bool
	PingOnAdds bool
	PingOnRems bool
}

func CreateSubscription(channelID uint64, guildID uint64, additions bool, removals bool) error {
	_, err := DBCfg.Exec("INSERT INTO Subscriptions SET channel_id = ?, guild_id = ?, ping_on_code_add = ?, ping_on_code_remove = ?", channelID, guildID, additions, removals)
	return err
}

func UpdateSubscription(channelID uint64, additions bool, removals bool) error {
	_, err := DBCfg.Exec("UPDATE Subscriptions SET ping_on_code_add = ?, ping_on_code_remove = ?, active = true WHERE channel_id = ?", additions, removals, channelID)
	return err
}

func DeactivateSubscription(channelID uint64) error {
	// _, err := DBCfg.Exec("DELETE FROM Subscriptions WHERE channel_id = ?", channelID)
	_, err := DBCfg.Exec("UPDATE Subscriptions SET active = false WHERE channel_id = ?", channelID)
	return err
}

func GetSubscription(channelID uint64) (*Subscription, error) {
	var pingOnAdd bool
	var pingOnRem bool
	var active bool
	s := DBCfg.QueryRow("SELECT ping_on_code_add, ping_on_code_remove, active FROM Subscriptions WHERE channel_id = ?", channelID)
	if err := s.Scan(&pingOnAdd, &pingOnRem, &active); err != nil {
		return nil, err
	}

	return &Subscription{
		ChannelID: channelID,
		Active: active,
		PingOnAdds: pingOnAdd,
		PingOnRems: pingOnRem,
	}, nil
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
		if rows.Err() != nil {
			return results, err
		}
		rows.Scan(&val)
		results = append(results, val)
	}
	return results, nil
}

func SetGameFilters(channelID uint64, games *set.Set[string]) error {
	if _, err := DBCfg.Exec("DELETE FROM SubscriptionGames WHERE channel_id = ?", channelID); err != nil {
		return err
	}
	
	for _, game := range games.Slice() {
		_, err := DBCfg.Exec("INSERT INTO SubscriptionGames VALUES (?, ?)", channelID, game)
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
		if rows.Err() != nil {
			return results, err
		}
		rows.Scan(&val)
		results = append(results, val)
	}
	return results, nil
}

//// REMOVE ////
type GuildAdmin struct {
	GuildID uint64
	RoleID uint64
}

func AddGuildAdmin(guildID uint64, roleID uint64) error {
	// TODO: handle duplicate
	_, err := DBCfg.Exec("INSERT INTO GuildAdmins VALUES (?, ?)", guildID, roleID)
	if err != nil {
		log.Printf("ERROR: could not add GuildAdmin: %v", err)
	}
	return err
}

func IsGuildAdmin(guildID uint64, roleID uint64) (bool, error) {
	row := DBCfg.QueryRow("SELECT * FROM GuildAdmins WHERE guild_id = ? AND role_id = ?", guildID, roleID)
	
	var foundGID uint64
	var foundRID uint64
	if err := row.Scan(&foundGID, &foundRID); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func GetGuildAdmins() ([]GuildAdmin, error) {
	var Ret []GuildAdmin
	rows, err := DBCfg.Query("SELECT * FROM GuildAdmins")
	if err != nil {
        return nil, fmt.Errorf("error reading GuildAdmins: %v", err)
    }

	for rows.Next() {
		var guildID uint64
		var roleId uint64
		rows.Scan(&guildID, &roleId)
		Ret = append(Ret, GuildAdmin{guildID, roleId})
	}

	return Ret, nil
}
