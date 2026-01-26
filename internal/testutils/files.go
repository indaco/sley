// Package testutils provides helper functions for testing CLI applications.
// It includes utilities for reading/writing temp files, capturing CLI output,
// and running CLI commands in isolated working directories.
package testutils

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/core"
)

// ReadFile reads the contents of a file and fails the test on error.
func ReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// WriteFile writes content to a file with the given permissions and fails the test on error.
func WriteFile(t *testing.T, path, content string, perm fs.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatalf("failed to write file %q: %v", path, err)
	}
}

// ReadTempVersionFile reads and returns the trimmed contents of the `.version` file in the given directory.
func ReadTempVersionFile(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".version"))
	if err != nil {
		t.Fatalf("failed to read .version file: %v", err)
	}
	return strings.TrimSpace(string(data))
}

// WriteTempVersionFile writes a `.version` file with the given content and returns its path.
func WriteTempVersionFile(t *testing.T, dir, version string) string {
	t.Helper()
	path := filepath.Join(dir, ".version")
	WriteFile(t, path, version, core.PermPublicRead)

	return path
}

// WriteTempConfig writes a temporary `.sley.yaml` file with the given content and returns its path.
func WriteTempConfig(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, ".sley.yaml")

	WriteFile(t, tmpPath, content, core.PermPublicRead)
	return tmpPath
}
