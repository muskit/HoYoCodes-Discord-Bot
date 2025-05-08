package util_test

import (
	"testing"

	"github.com/muskit/hoyocodes-discord-bot/pkg/util"
)

func TestCodeListing(t *testing.T) {
	codes := [][]string{
		{"ABC123", "This is a test code description that is quite long"},
		{"XYZ789", "Short desc"},
	}
	game := "Genshin Impact"
	expected :=
		"- [`ABC123`](<https://genshin.hoyoverse.com/en/gift?code=ABC123>) - This is a test code description that is quite long\n"+
		"- [`XYZ789`](<https://genshin.hoyoverse.com/en/gift?code=XYZ789>) - Short desc"
	result := util.CodeListing(codes, &game)
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// Test without game
	expected =
		"- `ABC123` - This is a test code description that is quite long\n"+
		"- `XYZ789` - Short desc"
	result = util.CodeListing(codes, nil)
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCodeRedeemURL(t *testing.T) {
	code := "ABC123"
	game := "Genshin Impact"
	expected := "https://genshin.hoyoverse.com/en/gift?code=ABC123"
	result := util.CodeRedeemURL(code, game)
	if result == nil || *result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	game = "Unknown Game"
	result = util.CodeRedeemURL(code, game)
	if result != nil {
		t.Errorf("expected nil, got %v", *result)
	}
}

func TestDownstackIntoSlices(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	capacity := 3
	expected := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	result := util.DownstackIntoSlices(slice, capacity)
	for i := range expected {
		for j := range expected[i] {
			if result[i][j] != expected[i][j] {
				t.Errorf("expected %v, got %v", expected, result)
			}
		}
	}
}

func TestDownstackIntoSlicesNonDivisible(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8}
	capacity := 3
	expected := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8},
	}
	result := util.DownstackIntoSlices(slice, capacity)
	for i := range expected {
		for j := range expected[i] {
			if result[i][j] != expected[i][j] {
				t.Errorf("expected %v, got %v", expected, result)
			}
		}
	}
}

func TestAlphaNumStrip(t *testing.T) {
	s := " GA3M24Z5C5KT "
	expected := "GA3M24Z5C5KT"
	if util.AlphaNumStrip(s) != expected {
		t.Fail()
	}

	s = "– 60 stellar jade (new!)"
	expected = "60 stellar jade (new!)"
	if util.AlphaNumStrip(s) != expected {
		t.Fail()
	}

	s = " – 60 stellar jade (new!) "
	expected = "60 stellar jade (new!)"
	if util.AlphaNumStrip(s) != expected {
		t.Fail()
	}

}