package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/sley/internal/testutils"
)

/* ------------------------------------------------------------------------- */
/* LOAD CONFIG                                                               */
/* ------------------------------------------------------------------------- */

func TestLoadConfig(t *testing.T) {
	t.Run("from env", func(t *testing.T) {
		os.Setenv("SLEY_PATH", "env-defined/.version")
		defer os.Unsetenv("SLEY_PATH")

		cfg, err := LoadConfigFn()
		checkError(t, err, false)
		checkConfigNil(t, cfg, false)
		checkConfigPath(t, cfg, false, "env-defined/.version")
	})

	t.Run("valid yaml file with path", func(t *testing.T) {
		content := "path: ./my-folder/.version\n"
		tmpPath := testutils.WriteTempConfig(t, content)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			checkError(t, err, false)
			checkConfigNil(t, cfg, false)
			checkConfigPath(t, cfg, false, "./my-folder/.version")
		})
	})

	t.Run("missing file fallback", func(t *testing.T) {
		tmpDir := t.TempDir()
		runInTempDir(t, filepath.Join(tmpDir, "dummy"), func() {
			cfg, err := LoadConfigFn()
			checkError(t, err, false)
			checkConfigNil(t, cfg, true)
		})
	})

	t.Run("empty config falls back to default path", func(t *testing.T) {
		content := "{}\n"
		tmpPath := testutils.WriteTempConfig(t, content)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			checkError(t, err, false)
			checkConfigNil(t, cfg, false)
			checkConfigPath(t, cfg, false, ".version")
		})
	})

	t.Run("invalid yaml (bad format)", func(t *testing.T) {
		content := "not_yaml::: true"
		tmpPath := testutils.WriteTempConfig(t, content)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			checkError(t, err, true)
			checkConfigNil(t, cfg, true)
		})
	})

	t.Run("unmarshal error (syntax)", func(t *testing.T) {
		content := ": this is invalid"
		tmpPath := testutils.WriteTempConfig(t, content)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			checkError(t, err, true)
			checkConfigNil(t, cfg, true)
		})
	})

	t.Run("read file error (directory instead of file)", func(t *testing.T) {
		tmpDir := t.TempDir()
		runInTempDir(t, filepath.Join(tmpDir, "dummy"), func() {
			if err := os.Mkdir(".sley.yaml", 0755); err != nil {
				t.Fatal(err)
			}
			cfg, err := LoadConfigFn()
			checkError(t, err, true)
			checkConfigNil(t, cfg, true)
		})
	})
}

/* ------------------------------------------------------------------------- */
/* NORMALIZE VERSION PATH                                                    */
/* ------------------------------------------------------------------------- */

func TestNormalizeVersionPath(t *testing.T) {
	// Case 1: path is a file
	got := NormalizeVersionPath("foo/.version")
	if got != "foo/.version" {
		t.Errorf("expected unchanged path, got %q", got)
	}

	// Case 2: path is a directory
	tmp := t.TempDir()
	got = NormalizeVersionPath(tmp)
	expected := filepath.Join(tmp, ".version")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
