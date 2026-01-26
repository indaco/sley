package config

import (
	"testing"

	"github.com/indaco/sley/internal/testutils"
)

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

	if defaults.ModuleMaxDepth == nil || *defaults.ModuleMaxDepth != 10 {
		t.Errorf("expected MaxDepth to be 10, got %v", defaults.ModuleMaxDepth)
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
    module_max_depth: 5
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
    module_max_depth: 3
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
