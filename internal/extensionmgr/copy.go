package extensionmgr

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/indaco/sley/internal/core"
)

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
func (c *OSFileCopier) CopyFile(src, dst string, perm core.FileMode) error {
	in, err := c.openSrcFile(src)
	if err != nil {
		return fmt.Errorf("failed to open source %q: %w", src, err)
	}
	defer in.Close()

	out, err := c.openDstFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("failed to create destination %q: %w", dst, err)
	}
	defer out.Close()

	if _, err := c.copyFn(out, in); err != nil {
		return fmt.Errorf("failed to copy %q to %q: %w", src, dst, err)
	}

	return nil
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
