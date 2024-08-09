package debug

import (
	"fmt"
	"runtime"
	"strings"
)

// PrintStack prints the stack trace to standard error.
func PrintStack(skip ...int) {
	fmt.Print(Stack(skip...))
}

// Stack returns a formatted stack trace of the calling goroutine.
func Stack(skip ...int) string {
	return StackWithFilters(nil, skip...)
}

// StackWithFilter returns a formatted stack trace with a single filter.
func StackWithFilter(filter string, skip ...int) string {
	return StackWithFilters([]string{filter}, skip...)
}

// StackWithFilters returns a formatted stack trace with multiple filters.
func StackWithFilters(filters []string, skip ...int) string {
	number := 0
	if len(skip) > 0 {
		number = skip[0]
	}

	var builder strings.Builder
	pc, file, line, ok := runtime.Caller(number)

	for i := 0; ok && i < maxCallerDepth; i++ {
		if filterFileByFilters(file, filters) {
			pc, file, line, ok = runtime.Caller(number + i + 1)
			continue
		}

		name := "unknown"
		if fn := runtime.FuncForPC(pc); fn != nil {
			name = fn.Name()
		}

		fmt.Fprintf(&builder, "%d.%s%s\n    %s:%d\n", i+1, getSpace(i), name, file, line)
		pc, file, line, ok = runtime.Caller(number + i + 1)
	}

	return builder.String()
}

func getSpace(index int) string {
	if index > 9 {
		return " "
	}
	return "  "
}
