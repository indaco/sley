package pathutil

import (
	"path/filepath"
	"strings"

	"github.com/indaco/sley/internal/apperrors"
)

// resolveSymlinks attempts to resolve symlinks for both paths.
// On case-insensitive filesystems (e.g., macOS), /var may be a symlink to /private/var.
// To avoid false mismatches, both paths are resolved together:
// if either resolution fails, the originals are returned unchanged.
func resolveSymlinks(absBase, absPath string) (string, string) {
	resolvedBase, baseErr := filepath.EvalSymlinks(absBase)
	if baseErr != nil {
		return absBase, absPath
	}

	// Try resolving the full path first
	if resolvedPath, err := filepath.EvalSymlinks(absPath); err == nil {
		return resolvedBase, resolvedPath
	}

	// Path doesn't exist yet; try resolving parent directory and reattach the filename
	parent := filepath.Dir(absPath)
	if resolvedParent, err := filepath.EvalSymlinks(parent); err == nil {
		return resolvedBase, filepath.Join(resolvedParent, filepath.Base(absPath))
	}

	// Cannot resolve path side - fall back to both unresolved to avoid mismatch
	return absBase, absPath
}

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

		// Resolve symlinks to normalize paths on case-insensitive filesystems
		absBase, absPath = resolveSymlinks(absBase, absPath)

		// Check for directory traversal
		if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) && absPath != absBase {
			return "", &apperrors.PathValidationError{Path: path, Reason: "path traversal detected"}
		}
	}

	return cleanPath, nil
}

// IsWithinDir checks if a path is within a given directory.
// Resolves symlinks to handle case-insensitive filesystems correctly.
func IsWithinDir(path string, dir string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}

	// Resolve symlinks to normalize paths on case-insensitive filesystems
	absDir, absPath = resolveSymlinks(absDir, absPath)

	return strings.HasPrefix(absPath, absDir+string(filepath.Separator)) || absPath == absDir
}
