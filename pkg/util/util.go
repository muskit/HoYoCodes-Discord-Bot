package util

import (
	"fmt"
	"strings"

	"github.com/cdfmlr/ellipsis"
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