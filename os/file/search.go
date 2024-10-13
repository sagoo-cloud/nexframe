package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Search searches for a file by name in the given priority search paths,
// current working directory, and executable directory.
// It returns the absolute file path if found, or an empty string and error if not found.
func Search(name string, prioritySearchPaths ...string) (string, error) {
	// Check if it's an absolute path
	if filepath.IsAbs(name) {
		if _, err := os.Stat(name); err == nil {
			return name, nil
		}
		return "", nil
	}

	// Prepare search paths
	searchPaths := append([]string{}, prioritySearchPaths...)
	if pwd, err := os.Getwd(); err == nil {
		searchPaths = append(searchPaths, pwd)
	}
	if exePath, err := os.Executable(); err == nil {
		searchPaths = append(searchPaths, filepath.Dir(exePath))
	}

	// Remove duplicates
	uniquePaths := make([]string, 0, len(searchPaths))
	seen := make(map[string]bool)
	for _, path := range searchPaths {
		if !seen[path] {
			seen[path] = true
			uniquePaths = append(uniquePaths, path)
		}
	}

	// Search for the file
	for _, path := range uniquePaths {
		fullPath := filepath.Join(path, name)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	// File not found, return error
	var pathsStr strings.Builder
	for i, path := range uniquePaths {
		fmt.Fprintf(&pathsStr, "\n%d. %s", i+1, path)
	}
	return "", fmt.Errorf("cannot find %q in following paths:%s", name, pathsStr.String())
}
