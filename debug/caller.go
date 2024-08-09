package debug

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

const (
	maxCallerDepth = 1000
)

var (
	goRootForFilter = filepath.ToSlash(runtime.GOROOT())
	stackFilterKey  = "/nexframe/gdebug"
)

// SetStackFilterKey sets the stack filter key used for filtering debug information.
func SetStackFilterKey(key string) {
	stackFilterKey = key
}

// Caller returns the function name, absolute file path, and line number of the caller.
func Caller(skip ...int) (function string, path string, line int, err error) {
	return CallerWithFilter(nil, skip...)
}

// CallerWithFilter returns the caller information with path filtering.
func CallerWithFilter(filters []string, skip ...int) (function string, path string, line int, err error) {
	skipFrames := 0
	if len(skip) > 0 {
		skipFrames = skip[0]
	}

	pc, file, line, ok := callerFromIndex(filters, skipFrames)
	if !ok {
		return "", "", -1, fmt.Errorf("unable to get caller information")
	}

	if fn := runtime.FuncForPC(pc); fn == nil {
		function = "unknown"
	} else {
		function = fn.Name()
	}

	return function, filepath.ToSlash(file), line, nil
}

func callerFromIndex(filters []string, skip int) (pc uintptr, file string, line int, ok bool) {
	for i := skip; i < maxCallerDepth; i++ {
		if pc, file, line, ok = runtime.Caller(i); ok {
			if filterFileByFilters(file, filters) {
				continue
			}
			return pc, file, line, true
		}
	}
	return 0, "", -1, false
}

func filterFileByFilters(file string, filters []string) bool {
	if file == "" || strings.Contains(file, stackFilterKey) {
		return true
	}

	for _, filter := range filters {
		if filter != "" && strings.Contains(file, filter) {
			return true
		}
	}

	return strings.HasPrefix(file, goRootForFilter)
}

// CallerPackage returns the package name of the caller.
func CallerPackage() string {
	function, _, _, _ := Caller(1)
	return extractPackageName(function)
}

func extractPackageName(function string) string {
	lastSlashIndex := strings.LastIndexByte(function, '/')
	if lastSlashIndex == -1 {
		return function[:strings.IndexByte(function, '.')]
	}

	dotIndex := strings.IndexByte(function[lastSlashIndex+1:], '.')
	return function[:lastSlashIndex+1+dotIndex]
}

// CallerFunction returns the function name of the caller.
func CallerFunction() string {
	function, _, _, _ := Caller(1)
	return function[strings.LastIndexByte(function, '.')+1:]
}

// CallerFilePath returns the file path of the caller.
func CallerFilePath() string {
	_, path, _, _ := Caller(1)
	return filepath.ToSlash(path)
}

// CallerDirectory returns the directory of the caller.
func CallerDirectory() string {
	_, path, _, _ := Caller(1)
	return filepath.ToSlash(filepath.Dir(path))
}

// CallerFileLine returns the file path along with the line number of the caller.
func CallerFileLine() string {
	_, path, line, _ := Caller(1)
	return fmt.Sprintf("%s:%d", filepath.ToSlash(path), line)
}

// CallerFileLineShort returns the file name along with the line number of the caller.
func CallerFileLineShort() string {
	_, path, line, _ := Caller(1)
	return fmt.Sprintf("%s:%d", filepath.Base(path), line)
}

// FuncPath returns the complete function path of given `f`.
func FuncPath(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// FuncName returns the function name of given `f`.
func FuncName(f interface{}) string {
	path := FuncPath(f)
	if path == "" {
		return ""
	}
	return filepath.Base(path)
}
