package db

import (
	"testing"
)

func TestIsAdminGood(t *testing.T) {
	id := 853531051594481715
	res, err := IsAdminRole(853531051594481715)
	if err != nil {
		t.Fatalf("Error'd checking if %d is admin: %v\n", id, err)
	}
	if !res {
		t.Fatalf("result for %d is false! expected true", id)
	}
}