package util

import (
	"fmt"
	"strings"

	"github.com/cdfmlr/ellipsis"
	"github.com/muskit/hoyocodes-discord-bot/pkg/consts"
)

// input: slice of code,description pairs
//
// returns: unordered list in markdown
func CodeListing(codes [][]string) string {
	ret := ""
	for _, elem := range codes {
		var line string
		code, description := elem[0], ellipsis.Ending(elem[1], 20)
		line = fmt.Sprintf("- `%v` - %v", code, description)
		ret += line + "\n"
	}
	return strings.Trim(ret, " \n	")
}

// returns nil if game doesn't have redeem URL
func CodeRedeemURL(game string, code string) *string {
	url, exists := consts.RedeemURL[game]
	if !exists {
		return nil
	}

	url += "/?code=" + code
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