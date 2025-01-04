package db

import (
	"testing"

	_ "modernc.org/sqlite"
)

func TestNonexistentAdmin(t *testing.T) {
	var id uint64 = 123
	res, err := IsGuildAdmin(id, id)
	if err != nil {
		t.Fatalf("Error'd checking if %d is admin: %v\n", id, err)
	}
	if res {
		t.Fatalf("result for %d is true! expected false", id)
	}
}

func TestAdminCheck(t *testing.T) {
	var guildID uint64 = 820255165864738816
	var roleID uint64 = 853531051594481715
	res, err := IsGuildAdmin(guildID, roleID)
	if err != nil {
		t.Fatalf("Error'd checking if %d is admin: %v\n", roleID, err)
	}
	if !res {
		t.Fatalf("result for %d is false! expected true", roleID)
	}
}