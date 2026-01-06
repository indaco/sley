// Package core defines interfaces and types for dependency injection.
// These interfaces enable proper testing without global mutable state.
package core

import (
	"context"
	"io/fs"
)

// FileSystem abstracts file system operations for testability.
// All methods accept context.Context to support cancellation and timeouts.
type FileSystem interface {
	// ReadFile reads the entire file at path and returns its contents.
	ReadFile(ctx context.Context, path string) ([]byte, error)

	// WriteFile writes data to the file at path with the given permissions.
	WriteFile(ctx context.Context, path string, data []byte, perm fs.FileMode) error

	// Stat returns file info for the path.
	Stat(ctx context.Context, path string) (fs.FileInfo, error)

	// MkdirAll creates a directory path, along with any necessary parents.
	MkdirAll(ctx context.Context, path string, perm fs.FileMode) error

	// Remove removes the file or empty directory at path.
	Remove(ctx context.Context, path string) error

	// RemoveAll removes path and any children it contains.
	RemoveAll(ctx context.Context, path string) error

	// ReadDir reads the directory named by path and returns a list of directory entries.
	ReadDir(ctx context.Context, path string) ([]fs.DirEntry, error)
}

// CommandExecutor abstracts command execution for testability.
type CommandExecutor interface {
	// Run executes a command and returns an error if it fails.
	Run(ctx context.Context, dir string, command string, args ...string) error

	// Output executes a command and returns its combined output.
	Output(ctx context.Context, dir string, command string, args ...string) (string, error)
}

// GitClient abstracts git operations for testability.
type GitClient interface {
	// DescribeTags returns the most recent tag reachable from HEAD.
	DescribeTags(ctx context.Context) (string, error)

	// Clone clones a repository to the given path.
	Clone(ctx context.Context, url string, path string) error

	// Pull updates the repository at path.
	Pull(ctx context.Context, path string) error

	// IsValidRepo checks if path is a valid git repository.
	IsValidRepo(path string) bool
}

// VersionReader abstracts version file reading operations.
type VersionReader interface {
	// Read reads a version from the given path.
	Read(ctx context.Context, path string) (string, error)
}

// VersionWriter abstracts version file writing operations.
type VersionWriter interface {
	// Write writes a version string to the given path.
	Write(ctx context.Context, path string, version string) error
}

// VersionManager combines reading and writing operations.
type VersionManager interface {
	VersionReader
	VersionWriter

	// Initialize creates a version file if it doesn't exist.
	Initialize(ctx context.Context, path string) (created bool, err error)
}

// HookRunner abstracts hook execution.
type HookRunner interface {
	// Run executes a hook and returns an error if it fails.
	Run(ctx context.Context) error

	// Name returns the hook's name for logging/error reporting.
	Name() string
}
