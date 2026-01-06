package core

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
)

// OSFileSystem implements FileSystem using the standard os package.
type OSFileSystem struct{}

// NewOSFileSystem returns a new OSFileSystem.
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

func (f *OSFileSystem) ReadFile(ctx context.Context, path string) ([]byte, error) {
	// Check if context is cancelled before performing I/O
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

func (f *OSFileSystem) WriteFile(ctx context.Context, path string, data []byte, perm fs.FileMode) error {
	// Check if context is cancelled before performing I/O
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}

func (f *OSFileSystem) Stat(ctx context.Context, path string) (fs.FileInfo, error) {
	// Check if context is cancelled before performing I/O
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.Stat(path)
}

func (f *OSFileSystem) MkdirAll(ctx context.Context, path string, perm fs.FileMode) error {
	// Check if context is cancelled before performing I/O
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.MkdirAll(path, perm)
}

func (f *OSFileSystem) Remove(ctx context.Context, path string) error {
	// Check if context is cancelled before performing I/O
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.Remove(path)
}

func (f *OSFileSystem) RemoveAll(ctx context.Context, path string) error {
	// Check if context is cancelled before performing I/O
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func (f *OSFileSystem) ReadDir(ctx context.Context, path string) ([]fs.DirEntry, error) {
	// Check if context is cancelled before performing I/O
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.ReadDir(path)
}

// EnsureParentDir creates the parent directory for a file path if it doesn't exist.
func EnsureParentDir(ctx context.Context, fs FileSystem, path string, perm fs.FileMode) error {
	return fs.MkdirAll(ctx, filepath.Dir(path), perm)
}
