package ao3

import (
	"strings"
	"strconv"
)

// AtoiWithComma performs strconv.Atoi, removing commas from the string
func AtoiWithComma(s string) (int, error) {
	s = strings.Replace(s, ",", "", -1)
	return strconv.Atoi(s)
}