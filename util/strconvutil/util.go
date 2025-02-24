package strconvutil

import (
	"slices"
)

func BasicUnquote(s string) (string, bool) {
	if len(s) < 2 {
		return "", false
	}

	quotes := []byte("\"'`") // allowed quotes
	if !slices.Contains(quotes, s[0]) {
		return "", false
	}

	// end quote must equal start
	if s[len(s)-1] != s[0] {
		return "", false
	}

	return s[1 : len(s)-2], true
}
