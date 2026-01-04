package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/testutils"
)

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */

// runInTempDir runs a function in a temporary directory, then restores to a safe directory.
// This handles the case where the CWD has been deleted by previous test cleanup.
func runInTempDir(t *testing.T, tmpPath string, fn func()) {
	t.Helper()

	// First, ensure we're in a valid directory. The CWD might have been
	// deleted by a previous test's cleanup. Use /tmp as a safe fallback.
	origDir, err := os.Getwd()
	if err != nil {
		// CWD doesn't exist - use /tmp as fallback
		origDir = os.TempDir()
		if chErr := os.Chdir(origDir); chErr != nil {
			t.Fatalf("failed to chdir to temp dir: %v", chErr)
		}
	}

	targetDir := filepath.Dir(tmpPath)
	if err := os.Chdir(targetDir); err != nil {
		t.Fatalf("failed to chdir to %s: %v", targetDir, err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	fn()
}

func checkError(t *testing.T, err error, wantErr bool) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Fatalf("expected err=%v, got err=%v", wantErr, err)
	}
}

func checkConfigNil(t *testing.T, cfg *Config, wantNil bool) {
	t.Helper()
	if wantNil && cfg != nil {
		t.Errorf("expected nil config, got %+v", cfg)
	}
	if !wantNil && cfg == nil {
		t.Fatal("expected non-nil config, got nil")
	}
}

func checkConfigPath(t *testing.T, cfg *Config, wantNil bool, wantPath string) {
	t.Helper()
	if !wantNil && cfg.Path != wantPath {
		t.Errorf("expected path %q, got %q", wantPath, cfg.Path)
	}
}

func requireNonNilWorkspace(t *testing.T, cfg *Config) {
	t.Helper()
	if cfg.Workspace == nil {
		t.Fatal("expected Workspace to be non-nil")
	}
}

func requireNonNilDiscovery(t *testing.T, cfg *Config) {
	t.Helper()
	requireNonNilWorkspace(t, cfg)
	if cfg.Workspace.Discovery == nil {
		t.Fatal("expected Discovery to be non-nil")
	}
}

func assertBoolPtr(t *testing.T, name string, ptr *bool, expected bool) {
	t.Helper()
	if ptr == nil {
		t.Errorf("expected %s to be non-nil", name)
		return
	}
	if *ptr != expected {
		t.Errorf("expected %s to be %v, got %v", name, expected, *ptr)
	}
}

func assertIntPtr(t *testing.T, name string, ptr *int, expected int) {
	t.Helper()
	if ptr == nil {
		t.Errorf("expected %s to be non-nil", name)
		return
	}
	if *ptr != expected {
		t.Errorf("expected %s to be %d, got %d", name, expected, *ptr)
	}
}

func assertDiscoveryEnabled(t *testing.T, disc *DiscoveryConfig, expected bool) {
	t.Helper()
	assertBoolPtr(t, "Enabled", disc.Enabled, expected)
}

func assertDiscoveryRecursive(t *testing.T, disc *DiscoveryConfig, expected bool) {
	t.Helper()
	assertBoolPtr(t, "Recursive", disc.Recursive, expected)
}

func assertDiscoveryMaxDepth(t *testing.T, disc *DiscoveryConfig, expected int) {
	t.Helper()
	assertIntPtr(t, "MaxDepth", disc.MaxDepth, expected)
}

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

/* ------------------------------------------------------------------------- */
/* SAVE CONFIG                                                               */
/* ------------------------------------------------------------------------- */

func TestSaveConfigFn(t *testing.T) {
	t.Run("basic save scenarios", func(t *testing.T) {
		defer func() {
			marshalFn = yaml.Marshal
			openFileFn = os.OpenFile
		}()

		tests := []struct {
			name               string
			cfg                *Config
			wantErr            bool
			overwriteMarshalFn bool
			mockMarshalErr     error
			overwriteOpenFile  bool
		}{
			{
				name:    "save minimal config",
				cfg:     &Config{Path: "my.version"},
				wantErr: false,
			},
			{
				name: "save config with plugins",
				cfg: &Config{
					Path: "custom.version",
					Extensions: []ExtensionConfig{
						{Name: "example", Path: "/plugin/path", Enabled: true},
					},
				},
				wantErr: false,
			},
			{
				name:               "marshal failure",
				cfg:                &Config{Path: "fail.version"},
				wantErr:            true,
				overwriteMarshalFn: true,
				mockMarshalErr:     fmt.Errorf("mock marshal failure"),
			},
			{
				name:              "write fails due to file permission",
				cfg:               &Config{Path: "fail-write.version"},
				wantErr:           true,
				overwriteOpenFile: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmp := t.TempDir()
				runInTempDir(t, filepath.Join(tmp, "dummy"), func() {
					if tt.overwriteMarshalFn {
						marshalFn = func(any) ([]byte, error) {
							return nil, tt.mockMarshalErr
						}
					}

					if tt.overwriteOpenFile {
						openFileFn = func(name string, flag int, perm os.FileMode) (*os.File, error) {
							// Simulate permission denied by opening read-only
							path := filepath.Join(t.TempDir(), "readonly.yaml")
							f, err := os.Create(path)
							if err != nil {
								t.Fatal(err)
							}
							f.Close()
							_ = os.Chmod(path, 0400)
							return os.OpenFile(path, os.O_WRONLY, 0400)
						}
					}

					err := SaveConfigFn(tt.cfg)
					if (err != nil) != tt.wantErr {
						t.Fatalf("SaveConfigFn() error = %v, wantErr = %v", err, tt.wantErr)
					}

					if !tt.wantErr {
						if _, err := os.Stat(".sley.yaml"); err != nil {
							t.Errorf(".sley.yaml was not created: %v", err)
						}
					}
				})
			})
		}
	})

	t.Run("write fails due to directory permission", func(t *testing.T) {
		tmp := t.TempDir()
		badDir := filepath.Join(tmp, "readonly")
		if err := os.Mkdir(badDir, 0500); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Chmod(badDir, 0755); err != nil {
				t.Logf("cleanup warning: failed to chmod %q: %v", badDir, err)
			}
		}()

		runInTempDir(t, filepath.Join(badDir, "dummy"), func() {
			err := SaveConfigFn(&Config{Path: "blocked.version"})
			if err == nil {
				t.Error("expected error due to write permission, got nil")
			}
		})
	})
}

func TestSaveConfigFn_WriteFileFn_Error(t *testing.T) {
	origWriteFn := writeFileFn
	defer func() {
		writeFileFn = origWriteFn
	}()

	tmp := t.TempDir()
	runInTempDir(t, filepath.Join(tmp, "dummy"), func() {
		writeFileFn = func(f *os.File, data []byte) (int, error) {
			fmt.Println(">>> writeFileFn invoked")
			return 0, fmt.Errorf("simulated write failure")
		}

		cfg := &Config{Path: "whatever"}
		err := SaveConfigFn(cfg)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		want := "failed to write config data: simulated write failure"
		if err.Error() != want {
			t.Errorf("unexpected error. got: %q, want: %q", err.Error(), want)
		}
	})
}

/* ------------------------------------------------------------------------- */
/* WORKSPACE CONFIG - DISCOVERY DEFAULTS                                     */
/* ------------------------------------------------------------------------- */

func TestDiscoveryDefaults(t *testing.T) {
	defaults := DiscoveryDefaults()

	if defaults == nil {
		t.Fatal("expected non-nil DiscoveryConfig")
	}

	if defaults.Enabled == nil || !*defaults.Enabled {
		t.Error("expected Enabled to be true by default")
	}

	if defaults.Recursive == nil || !*defaults.Recursive {
		t.Error("expected Recursive to be true by default")
	}

	if defaults.MaxDepth == nil || *defaults.MaxDepth != 10 {
		t.Errorf("expected MaxDepth to be 10, got %v", defaults.MaxDepth)
	}

	expectedExcludes := []string{
		"node_modules", ".git", "vendor", "tmp",
		"build", "dist", ".cache", "__pycache__",
	}

	if len(defaults.Exclude) != len(expectedExcludes) {
		t.Errorf("expected %d exclude patterns, got %d", len(expectedExcludes), len(defaults.Exclude))
	}

	for i, pattern := range expectedExcludes {
		if i >= len(defaults.Exclude) || defaults.Exclude[i] != pattern {
			t.Errorf("expected exclude[%d] to be %q, got %q", i, pattern, defaults.Exclude[i])
		}
	}
}

/* ------------------------------------------------------------------------- */
/* WORKSPACE CONFIG - PARSING FROM YAML                                      */
/* ------------------------------------------------------------------------- */

func TestLoadConfig_WorkspaceWithDiscovery(t *testing.T) {
	t.Run("workspace with discovery enabled", func(t *testing.T) {
		yamlContent := `path: .version
workspace:
  discovery:
    enabled: true
    recursive: true
    max_depth: 5
`
		tmpPath := testutils.WriteTempConfig(t, yamlContent)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			requireNonNilDiscovery(t, cfg)
			disc := cfg.Workspace.Discovery
			assertDiscoveryEnabled(t, disc, true)
			assertDiscoveryRecursive(t, disc, true)
			assertDiscoveryMaxDepth(t, disc, 5)
		})
	})

	t.Run("workspace with discovery disabled", func(t *testing.T) {
		yamlContent := `path: .version
workspace:
  discovery:
    enabled: false
`
		tmpPath := testutils.WriteTempConfig(t, yamlContent)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			requireNonNilDiscovery(t, cfg)
			assertDiscoveryEnabled(t, cfg.Workspace.Discovery, false)
		})
	})

	t.Run("workspace with custom excludes", func(t *testing.T) {
		yamlContent := `path: .version
workspace:
  discovery:
    exclude:
      - custom_exclude
      - another_exclude
`
		tmpPath := testutils.WriteTempConfig(t, yamlContent)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			requireNonNilDiscovery(t, cfg)
			excludes := cfg.Workspace.Discovery.Exclude
			if len(excludes) != 2 {
				t.Fatalf("expected 2 excludes, got %d", len(excludes))
			}
			if excludes[0] != "custom_exclude" {
				t.Errorf("expected excludes[0] to be 'custom_exclude', got %q", excludes[0])
			}
			if excludes[1] != "another_exclude" {
				t.Errorf("expected excludes[1] to be 'another_exclude', got %q", excludes[1])
			}
		})
	})
}

func assertModuleConfig(t *testing.T, mod ModuleConfig, name, path string) {
	t.Helper()
	if mod.Name != name {
		t.Errorf("expected module.Name to be %q, got %q", name, mod.Name)
	}
	if mod.Path != path {
		t.Errorf("expected module.Path to be %q, got %q", path, mod.Path)
	}
}

func TestLoadConfig_WorkspaceWithModules(t *testing.T) {
	t.Run("explicit modules defined", func(t *testing.T) {
		yamlContent := `path: .version
workspace:
  modules:
    - name: module1
      path: services/module1
      enabled: true
    - name: module2
      path: services/module2
      enabled: false
`
		tmpPath := testutils.WriteTempConfig(t, yamlContent)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			requireNonNilWorkspace(t, cfg)
			modules := cfg.Workspace.Modules
			if len(modules) != 2 {
				t.Fatalf("expected 2 modules, got %d", len(modules))
			}

			assertModuleConfig(t, modules[0], "module1", "services/module1")
			assertBoolPtr(t, "module[0].Enabled", modules[0].Enabled, true)

			assertModuleConfig(t, modules[1], "module2", "services/module2")
			assertBoolPtr(t, "module[1].Enabled", modules[1].Enabled, false)
		})
	})

	t.Run("modules without enabled field defaults to enabled", func(t *testing.T) {
		yamlContent := `path: .version
workspace:
  modules:
    - name: default-enabled
      path: services/default
`
		tmpPath := testutils.WriteTempConfig(t, yamlContent)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			requireNonNilWorkspace(t, cfg)
			if len(cfg.Workspace.Modules) != 1 {
				t.Fatalf("expected 1 module, got %d", len(cfg.Workspace.Modules))
			}

			module := cfg.Workspace.Modules[0]
			if module.Enabled != nil {
				t.Error("expected Enabled to be nil (unset)")
			}
		})
	})
}

/* ------------------------------------------------------------------------- */
/* WORKSPACE CONFIG - DEFAULT VALUES                                         */
/* ------------------------------------------------------------------------- */

func requireGetDiscoveryConfig(t *testing.T, cfg *Config) *DiscoveryConfig {
	t.Helper()
	discovery := cfg.GetDiscoveryConfig()
	if discovery == nil {
		t.Fatal("expected GetDiscoveryConfig to return non-nil defaults")
	}
	return discovery
}

func assertDefaultDiscoveryValues(t *testing.T, discovery *DiscoveryConfig) {
	t.Helper()
	assertDiscoveryEnabled(t, discovery, true)
	assertDiscoveryRecursive(t, discovery, true)
	assertDiscoveryMaxDepth(t, discovery, 10)
}

func TestConfig_WorkspaceDefaults(t *testing.T) {
	t.Run("no workspace section returns defaults", func(t *testing.T) {
		yamlContent := `path: .version`
		tmpPath := testutils.WriteTempConfig(t, yamlContent)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.Workspace != nil {
				t.Error("expected Workspace to be nil when not configured")
			}

			discovery := requireGetDiscoveryConfig(t, cfg)
			assertDefaultDiscoveryValues(t, discovery)
		})
	})

	t.Run("workspace without discovery section returns defaults", func(t *testing.T) {
		yamlContent := `path: .version
workspace:
  modules:
    - name: test
      path: test
`
		tmpPath := testutils.WriteTempConfig(t, yamlContent)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			requireNonNilWorkspace(t, cfg)
			if cfg.Workspace.Discovery != nil {
				t.Error("expected Discovery to be nil when not configured")
			}

			discovery := requireGetDiscoveryConfig(t, cfg)
			assertDiscoveryEnabled(t, discovery, true)
		})
	})

	t.Run("partial discovery config uses defaults for missing fields", func(t *testing.T) {
		yamlContent := `path: .version
workspace:
  discovery:
    max_depth: 3
`
		tmpPath := testutils.WriteTempConfig(t, yamlContent)
		runInTempDir(t, tmpPath, func() {
			cfg, err := LoadConfigFn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			discovery := requireGetDiscoveryConfig(t, cfg)
			assertDiscoveryEnabled(t, discovery, true)
			assertDiscoveryRecursive(t, discovery, true)
			assertDiscoveryMaxDepth(t, discovery, 3)
		})
	})
}

/* ------------------------------------------------------------------------- */
/* WORKSPACE CONFIG - HELPER METHODS                                         */
/* ------------------------------------------------------------------------- */

func TestConfig_GetExcludePatterns(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{
			name:     "no workspace config - returns defaults",
			config:   &Config{},
			expected: DefaultExcludePatterns,
		},
		{
			name: "workspace with custom excludes - merges with defaults",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Discovery: &DiscoveryConfig{
						Exclude: []string{"custom1", "custom2"},
					},
				},
			},
			expected: append(DefaultExcludePatterns, "custom1", "custom2"),
		},
		{
			name: "workspace with overlapping excludes - no duplicates",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Discovery: &DiscoveryConfig{
						Exclude: []string{".git", "custom_only"},
					},
				},
			},
			expected: append(DefaultExcludePatterns, "custom_only"),
		},
		{
			name: "workspace without discovery config - returns defaults",
			config: &Config{
				Workspace: &WorkspaceConfig{},
			},
			expected: DefaultExcludePatterns,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := tt.config.GetExcludePatterns()

			if len(patterns) != len(tt.expected) {
				t.Errorf("expected %d patterns, got %d", len(tt.expected), len(patterns))
			}

			// Convert to map for easier comparison
			patternMap := make(map[string]bool)
			for _, p := range patterns {
				patternMap[p] = true
			}

			for _, expected := range tt.expected {
				if !patternMap[expected] {
					t.Errorf("expected pattern %q not found in result", expected)
				}
			}
		})
	}
}

func TestConfig_HasExplicitModules(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name:     "no workspace - returns false",
			config:   &Config{},
			expected: false,
		},
		{
			name: "workspace with no modules - returns false",
			config: &Config{
				Workspace: &WorkspaceConfig{},
			},
			expected: false,
		},
		{
			name: "workspace with empty modules list - returns false",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Modules: []ModuleConfig{},
				},
			},
			expected: false,
		},
		{
			name: "workspace with modules - returns true",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Modules: []ModuleConfig{
						{Name: "test", Path: "test"},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.HasExplicitModules()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConfig_IsModuleEnabled(t *testing.T) {
	enabled := true
	disabled := false

	tests := []struct {
		name       string
		config     *Config
		moduleName string
		expected   bool
	}{
		{
			name:       "no workspace - returns false",
			config:     &Config{},
			moduleName: "any",
			expected:   false,
		},
		{
			name: "module not found - returns false",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Modules: []ModuleConfig{
						{Name: "other", Path: "other"},
					},
				},
			},
			moduleName: "notfound",
			expected:   false,
		},
		{
			name: "module found and enabled explicitly",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Modules: []ModuleConfig{
						{Name: "test", Path: "test", Enabled: &enabled},
					},
				},
			},
			moduleName: "test",
			expected:   true,
		},
		{
			name: "module found and disabled explicitly",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Modules: []ModuleConfig{
						{Name: "test", Path: "test", Enabled: &disabled},
					},
				},
			},
			moduleName: "test",
			expected:   false,
		},
		{
			name: "module found with nil enabled (defaults to true)",
			config: &Config{
				Workspace: &WorkspaceConfig{
					Modules: []ModuleConfig{
						{Name: "test", Path: "test", Enabled: nil},
					},
				},
			},
			moduleName: "test",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsModuleEnabled(tt.moduleName)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestModuleConfig_IsEnabled(t *testing.T) {
	enabled := true
	disabled := false

	tests := []struct {
		name     string
		module   *ModuleConfig
		expected bool
	}{
		{
			name:     "nil enabled field - defaults to true",
			module:   &ModuleConfig{Name: "test", Path: "test", Enabled: nil},
			expected: true,
		},
		{
			name:     "explicitly enabled",
			module:   &ModuleConfig{Name: "test", Path: "test", Enabled: &enabled},
			expected: true,
		},
		{
			name:     "explicitly disabled",
			module:   &ModuleConfig{Name: "test", Path: "test", Enabled: &disabled},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.module.IsEnabled()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

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

/* ------------------------------------------------------------------------- */
/* BACKWARD COMPATIBILITY                                                    */
/* ------------------------------------------------------------------------- */

/* ------------------------------------------------------------------------- */
/* TAG MANAGER CONFIG                                                        */
/* ------------------------------------------------------------------------- */

func TestTagManagerConfig_GetAutoCreate(t *testing.T) {
	tests := []struct {
		name     string
		config   *TagManagerConfig
		expected bool
	}{
		{
			name:     "nil AutoCreate - defaults to true",
			config:   &TagManagerConfig{AutoCreate: nil},
			expected: true,
		},
		{
			name:     "explicit true",
			config:   &TagManagerConfig{AutoCreate: boolPtr(true)},
			expected: true,
		},
		{
			name:     "explicit false",
			config:   &TagManagerConfig{AutoCreate: boolPtr(false)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetAutoCreate()
			if result != tt.expected {
				t.Errorf("GetAutoCreate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTagManagerConfig_GetAnnotate(t *testing.T) {
	tests := []struct {
		name     string
		config   *TagManagerConfig
		expected bool
	}{
		{
			name:     "nil Annotate - defaults to true",
			config:   &TagManagerConfig{Annotate: nil},
			expected: true,
		},
		{
			name:     "explicit true",
			config:   &TagManagerConfig{Annotate: boolPtr(true)},
			expected: true,
		},
		{
			name:     "explicit false",
			config:   &TagManagerConfig{Annotate: boolPtr(false)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetAnnotate()
			if result != tt.expected {
				t.Errorf("GetAnnotate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTagManagerConfig_GetPrefix(t *testing.T) {
	tests := []struct {
		name     string
		config   *TagManagerConfig
		expected string
	}{
		{
			name:     "empty prefix - defaults to v",
			config:   &TagManagerConfig{Prefix: ""},
			expected: "v",
		},
		{
			name:     "custom prefix",
			config:   &TagManagerConfig{Prefix: "release-"},
			expected: "release-",
		},
		{
			name:     "v prefix explicit",
			config:   &TagManagerConfig{Prefix: "v"},
			expected: "v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPrefix()
			if result != tt.expected {
				t.Errorf("GetPrefix() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// boolPtr is a helper to create a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}

func TestConfig_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		check     func(t *testing.T, cfg *Config)
	}{
		{
			name: "existing config without workspace - still works",
			yamlInput: `path: .version
plugins:
  commit-parser: true
`,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				if cfg.Path != ".version" {
					t.Errorf("expected path to be '.version', got %q", cfg.Path)
				}
				if cfg.Plugins == nil || !cfg.Plugins.CommitParser {
					t.Error("expected plugins.commit-parser to be true")
				}
				if cfg.Workspace != nil {
					t.Error("expected Workspace to be nil for legacy configs")
				}
			},
		},
		{
			name: "config with extensions and no workspace",
			yamlInput: `path: custom.version
extensions:
  - name: test-ext
    path: /path/to/ext
    enabled: true
`,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				if cfg.Path != "custom.version" {
					t.Errorf("expected path to be 'custom.version', got %q", cfg.Path)
				}
				if len(cfg.Extensions) != 1 {
					t.Fatalf("expected 1 extension, got %d", len(cfg.Extensions))
				}
				if cfg.Extensions[0].Name != "test-ext" {
					t.Errorf("expected extension name 'test-ext', got %q", cfg.Extensions[0].Name)
				}
			},
		},
		{
			name:      "minimal config - just path",
			yamlInput: `path: minimal.version`,
			check: func(t *testing.T, cfg *Config) {
				t.Helper()
				if cfg.Path != "minimal.version" {
					t.Errorf("expected path to be 'minimal.version', got %q", cfg.Path)
				}
				if cfg.Workspace != nil {
					t.Error("expected Workspace to be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpPath := testutils.WriteTempConfig(t, tt.yamlInput)
			runInTempDir(t, tmpPath, func() {
				cfg, err := LoadConfigFn()
				if err != nil {
					t.Fatalf("unexpected error loading legacy config: %v", err)
				}

				tt.check(t, cfg)
			})
		})
	}
}

/* ------------------------------------------------------------------------- */
/* AUDIT LOG CONFIG GETTER TESTS                                             */
/* ------------------------------------------------------------------------- */

func TestAuditLogConfig_GetPath(t *testing.T) {
	tests := []struct {
		name     string
		config   *AuditLogConfig
		expected string
	}{
		{
			name:     "empty path returns default",
			config:   &AuditLogConfig{Path: ""},
			expected: ".version-history.json",
		},
		{
			name:     "custom path returns custom",
			config:   &AuditLogConfig{Path: "custom-history.json"},
			expected: "custom-history.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPath()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestAuditLogConfig_GetFormat(t *testing.T) {
	tests := []struct {
		name     string
		config   *AuditLogConfig
		expected string
	}{
		{
			name:     "empty format returns default",
			config:   &AuditLogConfig{Format: ""},
			expected: "json",
		},
		{
			name:     "custom format returns custom",
			config:   &AuditLogConfig{Format: "yaml"},
			expected: "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetFormat()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* CHANGELOG GENERATOR CONFIG GETTER TESTS                                   */
/* ------------------------------------------------------------------------- */

func TestChangelogGeneratorConfig_GetChangesDir(t *testing.T) {
	tests := []struct {
		name     string
		config   *ChangelogGeneratorConfig
		expected string
	}{
		{
			name:     "empty changes dir returns default",
			config:   &ChangelogGeneratorConfig{ChangesDir: ""},
			expected: ".changes",
		},
		{
			name:     "custom changes dir returns custom",
			config:   &ChangelogGeneratorConfig{ChangesDir: "changelog-entries"},
			expected: "changelog-entries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetChangesDir()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestChangelogGeneratorConfig_GetChangelogPath(t *testing.T) {
	tests := []struct {
		name     string
		config   *ChangelogGeneratorConfig
		expected string
	}{
		{
			name:     "empty changelog path returns default",
			config:   &ChangelogGeneratorConfig{ChangelogPath: ""},
			expected: "CHANGELOG.md",
		},
		{
			name:     "custom changelog path returns custom",
			config:   &ChangelogGeneratorConfig{ChangelogPath: "docs/HISTORY.md"},
			expected: "docs/HISTORY.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetChangelogPath()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestChangelogGeneratorConfig_GetMode(t *testing.T) {
	tests := []struct {
		name     string
		config   *ChangelogGeneratorConfig
		expected string
	}{
		{
			name:     "empty mode returns default",
			config:   &ChangelogGeneratorConfig{Mode: ""},
			expected: "versioned",
		},
		{
			name:     "unified mode returns unified",
			config:   &ChangelogGeneratorConfig{Mode: "unified"},
			expected: "unified",
		},
		{
			name:     "both mode returns both",
			config:   &ChangelogGeneratorConfig{Mode: "both"},
			expected: "both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetMode()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* CONTRIBUTORS CONFIG TESTS                                                 */
/* ------------------------------------------------------------------------- */

func TestContributorsConfig_Icon(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		wantIcon  string
	}{
		{
			name: "contributors with icon",
			yamlInput: `path: .version
plugins:
  changelog-generator:
    enabled: true
    contributors:
      enabled: true
      icon: "❤️"
`,
			wantIcon: "❤️",
		},
		{
			name: "contributors without icon",
			yamlInput: `path: .version
plugins:
  changelog-generator:
    enabled: true
    contributors:
      enabled: true
`,
			wantIcon: "",
		},
		{
			name: "contributors with custom format and icon",
			yamlInput: `path: .version
plugins:
  changelog-generator:
    enabled: true
    contributors:
      enabled: true
      format: "- {{.Name}}"
      icon: "👥"
`,
			wantIcon: "👥",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpPath := testutils.WriteTempConfig(t, tt.yamlInput)
			runInTempDir(t, tmpPath, func() {
				cfg, err := LoadConfigFn()
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if cfg.Plugins == nil || cfg.Plugins.ChangelogGenerator == nil {
					t.Fatal("expected changelog-generator plugin config")
				}

				if cfg.Plugins.ChangelogGenerator.Contributors == nil {
					t.Fatal("expected contributors config")
				}

				if cfg.Plugins.ChangelogGenerator.Contributors.Icon != tt.wantIcon {
					t.Errorf("expected icon %q, got %q",
						tt.wantIcon, cfg.Plugins.ChangelogGenerator.Contributors.Icon)
				}
			})
		})
	}
}
