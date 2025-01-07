package db

import (
	"database/sql"
	"fmt"
	"log"
)

type GuildAdmin struct {
	GuildID uint64
	RoleID uint64
}

func CreateSubscription(channelID uint64) error {
	_, err := DBCfg.Exec("INSERT INTO Subscriptions VALUES (?)", channelID)
	return err
}

//// REMOVE ////
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