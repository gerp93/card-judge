package util

import (
	"os"
	"os/user"
	"path/filepath"
)

// FindPath searches for a file/directory across multiple possible locations
func FindPath(relativePaths ...string) string {
	possiblePaths := []string{}

	// Add the relative paths first
	possiblePaths = append(possiblePaths, relativePaths...)

	// Also try relative to executable
	if ex, err := os.Executable(); err == nil {
		exePath := filepath.Dir(ex)
		for _, relative := range relativePaths {
			possiblePaths = append(possiblePaths,
				filepath.Join(exePath, "..", relative),
				filepath.Join(exePath, "..", "..", relative),
				filepath.Join(exePath, "..", "..", "..", relative),
			)
		}
	}

	// Try relative to user home dir
	if home, err := user.Current(); err == nil {
		for _, relative := range relativePaths {
			possiblePaths = append(possiblePaths,
				filepath.Join(home.HomeDir, "source/repos/card-judge", relative),
			)
		}
	}

	// Return the first path that exists
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Return the first path if none exist (will fail with clear error when trying to use it)
	if len(possiblePaths) > 0 {
		return possiblePaths[0]
	}

	return relativePaths[0]
}
