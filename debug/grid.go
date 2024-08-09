package debug

import (
	"errors"
	"regexp"
	"runtime"
	"strconv"
)

var (
	gridRegex = regexp.MustCompile(`^\w+\s+(\d+)\s+`)
)

// GoroutineId retrieves and returns the current goroutine id from stack information.
// Note: This function has low performance due to its use of runtime.Stack.
// It is primarily intended for debugging purposes.
func GoroutineId() (int, error) {
	buf := make([]byte, 26)
	runtime.Stack(buf, false)
	match := gridRegex.FindSubmatch(buf)
	if len(match) < 2 {
		return 0, errors.New("failed to extract goroutine id")
	}
	return strconv.Atoi(string(match[1]))
}
