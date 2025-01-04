package db

import (
	"database/sql"
	"fmt"
	"log"
)

func AddAdminRole(roleID int) error {
	_, err := DBCfg.Exec("INSERT INTO AdminRoles VALUES (?)", roleID)
	if err != nil {
		log.Printf("ERROR: could not add AdminRole: %v", err)
	}
	return err
}

func IsAdminRole(id uint64) (bool, error) {
	row := DBCfg.QueryRow("SELECT * FROM AdminRoles WHERE role_id = ?", id)
	
	var val uint64
	if err := row.Scan(&val); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func GetAdminRoles() ([]uint64, error) {
	var Ret []uint64
	rows, err := DBCfg.Query("SELECT * FROM AdminRoles")
	if err != nil {
        return nil, fmt.Errorf("error reading AdminRoles: %v", err)
    }

	for rows.Next() {
		var roleId uint64
		rows.Scan(&roleId)
		Ret = append(Ret, roleId)
	}

	return Ret, nil
}