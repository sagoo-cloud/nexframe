package gstr

import "strings"

const (
	// NotFoundIndex is the position index for string not found in searching functions.
	NotFoundIndex = -1
)

// Contains reports whether `substr` is within `str`, case-sensitively.
func Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}
