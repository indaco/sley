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

	t.Run("from env with path traversal rejected", func(t *testing.T) {
		os.Setenv("SLEY_PATH", "../../../etc/.version")
		defer os.Unsetenv("SLEY_PATH")

		cfg, err := LoadConfigFn()
		checkError(t, err, true)
		checkConfigNil(t, cfg, true)
		if err != nil && err.Error() != "invalid SLEY_PATH: path traversal not allowed, use absolute path instead" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("from env with absolute path allowed", func(t *testing.T) {
		os.Setenv("SLEY_PATH", "/tmp/project/.version")
		defer os.Unsetenv("SLEY_PATH")

		cfg, err := LoadConfigFn()
		checkError(t, err, false)
		checkConfigNil(t, cfg, false)
		checkConfigPath(t, cfg, false, "/tmp/project/.version")
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
/* THEME CONFIGURATION                                                       */
/* ------------------------------------------------------------------------- */

func TestLoadConfigWithTheme(t *testing.T) {
	t.Run("valid yaml file with theme", func(t *testing.T) {
		content := "path: .version\ntheme: dracula\n"
		tmpPath := testutils.WriteTempConfig(t, content)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			checkError(t, err, false)
			checkConfigNil(t, cfg, false)
			if cfg.Theme != "dracula" {
				t.Errorf("expected theme 'dracula', got %q", cfg.Theme)
			}
		})
	})

	t.Run("empty theme in config", func(t *testing.T) {
		content := "path: .version\n"
		tmpPath := testutils.WriteTempConfig(t, content)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			checkError(t, err, false)
			checkConfigNil(t, cfg, false)
			if cfg.Theme != "" {
				t.Errorf("expected empty theme, got %q", cfg.Theme)
			}
		})
	})

	t.Run("explicit empty theme", func(t *testing.T) {
		content := "path: .version\ntheme: \"\"\n"
		tmpPath := testutils.WriteTempConfig(t, content)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			checkError(t, err, false)
			checkConfigNil(t, cfg, false)
			if cfg.Theme != "" {
				t.Errorf("expected empty theme, got %q", cfg.Theme)
			}
		})
	})
}

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name     string
		theme    string
		expected string
	}{
		{
			name:     "empty theme returns default",
			theme:    "",
			expected: "sley",
		},
		{
			name:     "custom theme is preserved",
			theme:    "dracula",
			expected: "dracula",
		},
		{
			name:     "sley theme is preserved",
			theme:    "sley",
			expected: "sley",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Theme: tt.theme}
			got := cfg.GetTheme()
			if got != tt.expected {
				t.Errorf("GetTheme() = %q, want %q", got, tt.expected)
			}
		})
	}
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
