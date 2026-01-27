package extensionmgr

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/indaco/sley/internal/core"
)

// FilePermissionError indicates insufficient permissions for file operations
type FilePermissionError struct {
	Src string
	Dst string
	Op  string // operation: "open", "create", "copy"
	Err error
}

func (e *FilePermissionError) Error() string {
	return fmt.Sprintf("permission denied: cannot %s file from %q to %q: %v", e.Op, e.Src, e.Dst, e.Err)
}

func (e *FilePermissionError) Unwrap() error {
	return e.Err
}

// DiskFullError indicates no space left on device
type DiskFullError struct {
	Path string
	Err  error
}

func (e *DiskFullError) Error() string {
	return fmt.Sprintf("no space left on device at %q: %v", e.Path, e.Err)
}

func (e *DiskFullError) Unwrap() error {
	return e.Err
}

// OSFileCopier implements core.FileCopier using OS file operations.
type OSFileCopier struct {
	walkFn      func(root string, fn filepath.WalkFunc) error
	relFn       func(basepath, targpath string) (string, error)
	openSrcFile func(name string) (*os.File, error)
	openDstFile func(name string, flag int, perm os.FileMode) (*os.File, error)
	copyFn      func(dst io.Writer, src io.Reader) (int64, error)
}

// NewOSFileCopier creates an OSFileCopier with default OS implementations.
func NewOSFileCopier() *OSFileCopier {
	return &OSFileCopier{
		walkFn:      filepath.Walk,
		relFn:       filepath.Rel,
		openSrcFile: os.Open,
		openDstFile: os.OpenFile,
		copyFn:      io.Copy,
	}
}

// Verify OSFileCopier implements core.FileCopier.
var _ core.FileCopier = (*OSFileCopier)(nil)

// CopyDir recursively copies all files and subdirectories from src to dst.
func (c *OSFileCopier) CopyDir(src, dst string) error {
	return c.walkFn(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk error at %q: %w", path, err)
		}

		skipFile, skipDir := shouldSkipEntry(info)
		if skipDir {
			return filepath.SkipDir
		}
		if skipFile {
			return nil
		}

		rel, err := c.relFn(src, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path from %q to %q: %w", src, path, err)
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return c.CopyFile(path, target, info.Mode())
	})
}

// CopyFile copies a single file from src to dst with given permissions.
// Returns context-aware errors for common failure scenarios.
func (c *OSFileCopier) CopyFile(src, dst string, perm core.FileMode) error {
	in, err := c.openSrcFile(src)
	if err != nil {
		return classifyFileCopyError(err, src, dst, "open")
	}
	defer in.Close()

	out, err := c.openDstFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return classifyFileCopyError(err, src, dst, "create")
	}
	defer out.Close()

	if _, err := c.copyFn(out, in); err != nil {
		return classifyFileCopyError(err, src, dst, "copy")
	}

	return nil
}

// classifyFileCopyError analyzes file system errors and provides context.
// It detects specific error types and returns structured errors with helpful information.
func classifyFileCopyError(err error, src, dst, operation string) error {
	if err == nil {
		return nil
	}

	// Check for permission errors
	if os.IsPermission(err) {
		return &FilePermissionError{
			Src: src,
			Dst: dst,
			Op:  operation,
			Err: err,
		}
	}

	// Check for file not found errors
	if os.IsNotExist(err) {
		return fmt.Errorf("source path not found: %q: %w", src, err)
	}

	// Check for disk full errors (ENOSPC error code)
	// Common patterns: "no space left on device", "disk full"
	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "no space left on device") || strings.Contains(errMsg, "disk full") {
		return &DiskFullError{
			Path: dst,
			Err:  err,
		}
	}

	// Return generic error with context
	return fmt.Errorf("failed to %s from %q to %q: %w", operation, src, dst, err)
}

// defaultFileCopier is the default file copier for backward compatibility.
var defaultFileCopier = NewOSFileCopier()

// skipNames defines a set of directory or file names excluded during directory copying.
var skipNames = map[string]struct{}{
	".git":         {},
	".DS_Store":    {},
	"node_modules": {},
}

// copyDirFn is kept for backward compatibility during migration.
var copyDirFn = func(src, dst string) error { return defaultFileCopier.CopyDir(src, dst) }

// shouldSkipEntry determines whether a file should be skipped or a directory subtree should be skipped.
func shouldSkipEntry(info os.FileInfo) (skipFile bool, skipDir bool) {
	_, skip := skipNames[info.Name()]
	if !skip {
		return false, false
	}
	if info.IsDir() {
		return false, true // skip entire directory
	}
	return true, false // skip just the file
}
