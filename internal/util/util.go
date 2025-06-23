package util

import (
	"os/user"
	"path/filepath"
	"strings"
)

// Expand resolves a file path, expanding the tilde (~) symbol to the user's home directory if present.
//
// Parameters:
//   - path: A string representing the file path to be expanded.
//
// Returns:
//
//	A string containing the expanded file path. If the original path starts with "~",
//	it is replaced with the user's home directory. Otherwise, the original path is returned unchanged.
func Expand(path string) string {
	if strings.HasPrefix(path, "~") {
		u, _ := user.Current()
		return filepath.Join(u.HomeDir, path[1:])
	}
	return path
}
