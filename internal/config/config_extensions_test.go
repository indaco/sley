package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/sley/internal/testutils"
)

/* ------------------------------------------------------------------------- */
/* EXTENSION CONFIGURATION                                                   */
/* ------------------------------------------------------------------------- */

// checkExtension is a helper to verify a single extension's configuration
func checkExtension(t *testing.T, ext ExtensionConfig, wantName, wantPath string, wantEnabled bool) {
	t.Helper()
	if ext.Name != wantName {
		t.Errorf("expected name %q, got %q", wantName, ext.Name)
	}
	if ext.Path != wantPath {
		t.Errorf("expected path %q, got %q", wantPath, ext.Path)
	}
	if ext.Enabled != wantEnabled {
		t.Errorf("expected enabled=%v, got %v", wantEnabled, ext.Enabled)
	}
}

// checkExtensionCount is a helper to verify extension count
func checkExtensionCount(t *testing.T, cfg *Config, want int) {
	t.Helper()
	if len(cfg.Extensions) != want {
		t.Fatalf("expected %d extension(s), got %d", want, len(cfg.Extensions))
	}
}

func TestLoadConfig_ExtensionConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		wantErr   bool
		check     func(t *testing.T, cfg *Config)
	}{
		{
			name: "single extension with all fields",
			yamlInput: `path: .version
extensions:
  - name: git-hook
    path: /home/user/.sley-extensions/git-hook
    enabled: true
`,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				checkExtensionCount(t, cfg, 1)
				checkExtension(t, cfg.Extensions[0], "git-hook", "/home/user/.sley-extensions/git-hook", true)
			},
		},
		{
			name: "multiple extensions with mixed enabled states",
			yamlInput: `path: .version
extensions:
  - name: changelog
    path: ./extensions/changelog
    enabled: true
  - name: git-tag
    path: ./extensions/git-tag
    enabled: false
  - name: notify
    path: ./extensions/notify
    enabled: true
`,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				checkExtensionCount(t, cfg, 3)
				checkExtension(t, cfg.Extensions[0], "changelog", "./extensions/changelog", true)
				checkExtension(t, cfg.Extensions[1], "git-tag", "./extensions/git-tag", false)
				checkExtension(t, cfg.Extensions[2], "notify", "./extensions/notify", true)
			},
		},
		{
			name: "extensions with plugins and workspace",
			yamlInput: `path: .version
plugins:
  commit-parser: true
extensions:
  - name: test-ext
    path: ./ext
    enabled: true
workspace:
  discovery:
    enabled: true
`,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				checkExtensionCount(t, cfg, 1)
				if cfg.Plugins == nil || !cfg.Plugins.CommitParser {
					t.Error("expected plugins.commit-parser to be true")
				}
				if cfg.Workspace == nil {
					t.Error("expected Workspace to be non-nil")
				}
			},
		},
		{
			name: "empty extensions list",
			yamlInput: `path: .version
extensions: []
`,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				checkExtensionCount(t, cfg, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpPath := testutils.WriteTempConfig(t, tt.yamlInput)
			runInTempDir(t, tmpPath, func() {
				cfg, err := LoadConfigFn()
				if (err != nil) != tt.wantErr {
					t.Fatalf("LoadConfigFn() error = %v, wantErr = %v", err, tt.wantErr)
				}
				if !tt.wantErr && cfg != nil {
					tt.check(t, cfg)
				}
			})
		})
	}
}

func TestSaveConfig_WithExtensions(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "save config with single extension",
			cfg: &Config{
				Path: ".version",
				Extensions: []ExtensionConfig{
					{
						Name:    "test-ext",
						Path:    "/path/to/ext",
						Enabled: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "save config with multiple extensions",
			cfg: &Config{
				Path: ".version",
				Extensions: []ExtensionConfig{
					{
						Name:    "changelog",
						Path:    "./ext/changelog",
						Enabled: true,
					},
					{
						Name:    "git-hook",
						Path:    "./ext/git-hook",
						Enabled: false,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "save config with extensions and plugins",
			cfg: &Config{
				Path: "custom.version",
				Plugins: &PluginConfig{
					CommitParser: true,
				},
				Extensions: []ExtensionConfig{
					{
						Name:    "notify",
						Path:    "/ext/notify",
						Enabled: true,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			runInTempDir(t, filepath.Join(tmp, "dummy"), func() {
				err := SaveConfigFn(tt.cfg)
				if (err != nil) != tt.wantErr {
					t.Fatalf("SaveConfigFn() error = %v, wantErr = %v", err, tt.wantErr)
				}

				if !tt.wantErr {
					// Verify file was created
					if _, err := os.Stat(".sley.yaml"); err != nil {
						t.Errorf(".sley.yaml was not created: %v", err)
						return
					}

					// Reload and verify
					reloaded, err := LoadConfigFn()
					if err != nil {
						t.Fatalf("failed to reload config: %v", err)
					}

					if len(reloaded.Extensions) != len(tt.cfg.Extensions) {
						t.Errorf("expected %d extensions after reload, got %d",
							len(tt.cfg.Extensions), len(reloaded.Extensions))
					}

					for i, ext := range tt.cfg.Extensions {
						if i >= len(reloaded.Extensions) {
							break
						}
						if reloaded.Extensions[i].Name != ext.Name {
							t.Errorf("extension[%d] name mismatch: got %q, want %q",
								i, reloaded.Extensions[i].Name, ext.Name)
						}
						if reloaded.Extensions[i].Enabled != ext.Enabled {
							t.Errorf("extension[%d] enabled mismatch: got %v, want %v",
								i, reloaded.Extensions[i].Enabled, ext.Enabled)
						}
					}
				}
			})
		})
	}
}
