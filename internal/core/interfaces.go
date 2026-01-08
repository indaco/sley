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

// Marshaler abstracts data marshaling operations (YAML, JSON, etc.).
type Marshaler interface {
	Marshal(v any) ([]byte, error)
}

// Unmarshaler abstracts data unmarshaling operations.
type Unmarshaler interface {
	Unmarshal(data []byte, v any) error
}

// MarshalUnmarshaler combines marshaling and unmarshaling.
type MarshalUnmarshaler interface {
	Marshaler
	Unmarshaler
}

// GitTagOperations abstracts git tag operations for testability.
type GitTagOperations interface {
	// CreateAnnotatedTag creates an annotated git tag with the given name and message.
	CreateAnnotatedTag(name, message string) error

	// CreateLightweightTag creates a lightweight git tag with the given name.
	CreateLightweightTag(name string) error

	// TagExists checks if a git tag with the given name exists.
	TagExists(name string) (bool, error)

	// GetLatestTag returns the most recent semver tag from git.
	GetLatestTag() (string, error)

	// PushTag pushes a specific tag to the remote.
	PushTag(name string) error
}

// GitCommitReader reads git commit information.
type GitCommitReader interface {
	// GetCommits returns commits between two references.
	GetCommits(since, until string) ([]string, error)
}

// GitBranchReader reads git branch information.
type GitBranchReader interface {
	// GetCurrentBranch returns the current git branch name.
	GetCurrentBranch(ctx context.Context) (string, error)
}

// FileCopier abstracts file and directory copy operations.
type FileCopier interface {
	// CopyDir recursively copies a directory from src to dst.
	CopyDir(src, dst string) error

	// CopyFile copies a single file from src to dst with given permissions.
	CopyFile(src, dst string, perm FileMode) error
}

// FileMode is an alias for fs.FileMode to avoid import cycles.
type FileMode = fs.FileMode

// UserDirProvider provides user directory information.
type UserDirProvider interface {
	// HomeDir returns the current user's home directory.
	HomeDir() (string, error)
}
