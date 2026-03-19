package clix

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestGetOrInitVersionFile(t *testing.T) {
	t.Parallel()
	t.Run("strict=true and file exists", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		tmpFile := testutils.WriteTempVersionFile(t, tmpDir, "0.1.0")

		created, err := GetOrInitVersionFile(tmpFile, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created {
			t.Errorf("expected created=false, got true")
		}
	})

	t.Run("strict=true and file missing", func(t *testing.T) {
		t.Parallel()
		missingPath := filepath.Join(t.TempDir(), "missing.version")

		created, err := GetOrInitVersionFile(missingPath, true)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if created {
			t.Errorf("expected created=false, got true")
		}
	})

	t.Run("strict=false and initialization succeeds", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		targetPath := filepath.Join(tmpDir, ".version")

		created, err := GetOrInitVersionFile(targetPath, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Errorf("expected created=true, got false")
		}

		// Check file actually created
		if _, err := os.Stat(targetPath); err != nil {
			t.Errorf("expected file to exist, got error: %v", err)
		}
	})
}

func TestGetOrInitVersionFile_InitError(t *testing.T) {
	t.Parallel()

	// Use MockFileSystem with write error to simulate initialization failure.
	// Call getOrInitVersionFileWith directly with the mock manager to avoid
	// mutating the package-level global defaultManager.
	mockFS := core.NewMockFileSystem()
	mockFS.WriteErr = errors.New("mock init failure")

	mgr := semver.NewVersionManager(mockFS, nil)

	targetPath := "/test/.version"

	created, err := getOrInitVersionFileWith(targetPath, false, mgr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if created {
		t.Errorf("expected created=false, got true")
	}
	if !strings.Contains(err.Error(), "mock init failure") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestFromCommand(t *testing.T) {
	t.Parallel()
	t.Run("path exists, strict false", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		tmpFile := testutils.WriteTempVersionFile(t, tmpDir, "0.1.0")

		cfg := &config.Config{Path: tmpFile}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{})

		created, err := FromCommand(appCli)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created {
			t.Errorf("expected created=false, got true")
		}
	})

	t.Run("path init success, strict false", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		targetPath := filepath.Join(tmpDir, ".version")

		cfg := &config.Config{Path: targetPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{})

		created, err := FromCommand(appCli)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Errorf("expected created=true, got false")
		}
	})
}
