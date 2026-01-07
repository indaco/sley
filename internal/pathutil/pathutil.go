package pathutil

import (
	"path/filepath"
	"strings"

	"github.com/indaco/sley/internal/apperrors"
)

// ValidatePath ensures a path is safe and within expected boundaries.
// It rejects paths with directory traversal attempts and cleans the path.
func ValidatePath(path string, baseDir string) (string, error) {
	if path == "" {
		return "", &apperrors.PathValidationError{Path: path, Reason: "path cannot be empty"}
	}

	// Clean the path to resolve . and .. components
	cleanPath := filepath.Clean(path)

	// If baseDir is provided, ensure path stays within it
	if baseDir != "" {
		absBase, err := filepath.Abs(baseDir)
		if err != nil {
			return "", &apperrors.PathValidationError{Path: path, Reason: "invalid base directory"}
		}

		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return "", &apperrors.PathValidationError{Path: path, Reason: "invalid path"}
		}

		// Check for directory traversal
		if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) && absPath != absBase {
			return "", &apperrors.PathValidationError{Path: path, Reason: "path traversal detected"}
		}
	}

	return cleanPath, nil
}

// IsWithinDir checks if a path is within a given directory.
func IsWithinDir(path string, dir string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}

	return strings.HasPrefix(absPath, absDir+string(filepath.Separator)) || absPath == absDir
}
