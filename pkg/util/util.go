package util

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/muskit/hoyocodes-discord-bot/pkg/consts"
)

// input: slice of code,description pairs
//
// returns: unordered list in markdown
func CodeListing(codes [][]string, game *string) string {
	ret := ""
	addURL := false

	if game != nil {
		_, addURL = consts.RedeemURL[*game]
	}

	for _, elem := range codes {
		var line string
		code, description := elem[0], elem[1]

		if addURL {
			url := CodeRedeemURL(code, *game);
			line = fmt.Sprintf("- [`%v`](<%v>) - %v", code, *url, description)
		} else {
			line = fmt.Sprintf("- `%v` - %v", code, description)
		}

		ret += line + "\n"
	}
	return strings.Trim(ret, " \n	")
}

// returns nil if game doesn't have redeem URL
func CodeRedeemURL(code string, game string) *string {
	url, exists := consts.RedeemURL[game]
	if !exists {
		return nil
	}

	url += "?code=" + code
	return &url
}

// Return a slice of slices where each slice has a max specified capacity.
func DownstackIntoSlices[T any](slice []T, cap int) [][]T {
	slices := [][]T{}
	for len(slice) > cap {
		overflow := slice[cap:]
		slice = slice[:cap]
		slices = append(slices, slice)
		slice = overflow
	}
	return append(slices, slice)
}

func AlphaNumStrip(s string) string {
	ret := strings.TrimLeftFunc(s, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsNumber(r))
	})
	
	ret = strings.TrimSpace(ret)
	
	return ret
}