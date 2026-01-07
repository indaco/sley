// Package pathutil provides utilities for safe path handling and validation.
//
// This package helps prevent path traversal attacks and ensures paths stay
// within expected directory boundaries. It is used throughout sley to
// validate user-provided paths and extension script locations.
//
// # Path Validation
//
// ValidatePath ensures a path is safe and within a base directory:
//
//	cleanPath, err := pathutil.ValidatePath("./scripts/hook.sh", "/app/extensions")
//	if err != nil {
//	    // Handle path traversal attempt or invalid path
//	}
//
// The function:
//   - Cleans the path using filepath.Clean
//   - Resolves to absolute paths for comparison
//   - Rejects paths that escape the base directory
//   - Returns structured errors via apperrors.PathValidationError
//
// # Directory Containment
//
// IsWithinDir checks if a path is contained within a directory:
//
//	if pathutil.IsWithinDir("/app/data/file.txt", "/app/data") {
//	    // Path is safe to use
//	}
//
// # Security
//
// This package is critical for security when handling:
//   - Extension script paths
//   - User-provided file paths
//   - Configuration file locations
//
// Always validate paths before performing file operations on user input.
package pathutil
