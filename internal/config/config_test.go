package config

import (
	"os"
	"testing"
)

func TestLoadConfigFromDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(t *testing.T, dir string)
		wantNil    bool
		wantErr    bool
		wantPrefix string
	}{
		{
			name: "valid .sley.yaml",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				content := "plugins:\n  tag-manager:\n    enabled: true\n    prefix: \"redis-\"\n"
				if err := os.WriteFile(dir+"/.sley.yaml", []byte(content), 0644); err != nil {
					t.Fatalf("failed to write config: %v", err)
				}
			},
			wantNil:    false,
			wantErr:    false,
			wantPrefix: "redis-",
		},
		{
			name:    "no .sley.yaml in dir",
			setup:   func(t *testing.T, dir string) {},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "malformed YAML",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				content := "plugins:\n  tag-manager:\n    enabled: [invalid yaml\n"
				if err := os.WriteFile(dir+"/.sley.yaml", []byte(content), 0644); err != nil {
					t.Fatalf("failed to write config: %v", err)
				}
			},
			wantNil: false,
			wantErr: true,
		},
		{
			name: "empty prefix falls back to default",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				content := "plugins:\n  tag-manager:\n    enabled: true\n    prefix: \"\"\n"
				if err := os.WriteFile(dir+"/.sley.yaml", []byte(content), 0644); err != nil {
					t.Fatalf("failed to write config: %v", err)
				}
			},
			wantNil:    false,
			wantErr:    false,
			wantPrefix: "v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			tt.setup(t, dir)

			cfg, err := LoadConfigFromDir(dir)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if cfg != nil {
					t.Fatalf("expected nil config, got %+v", cfg)
				}
				return
			}

			if cfg == nil {
				t.Fatal("expected non-nil config, got nil")
			}

			if tt.wantPrefix != "" {
				if cfg.Plugins == nil || cfg.Plugins.TagManager == nil {
					t.Fatal("expected tag-manager plugin config to be set")
				}
				got := cfg.Plugins.TagManager.GetPrefix()
				if got != tt.wantPrefix {
					t.Errorf("GetPrefix() = %q, want %q", got, tt.wantPrefix)
				}
			}
		})
	}
}

func TestMergeConfig_Theme(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		rootTheme string
		modTheme  string
		want      string
	}{
		{"module_has_theme", "sley", "dracula", "dracula"},
		{"module_has_no_theme", "sley", "", "sley"},
		{"both_empty", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := &Config{Theme: tt.rootTheme}
			mod := &Config{Theme: tt.modTheme}
			merged := MergeConfig(root, mod)
			if merged.Theme != tt.want {
				t.Errorf("Theme = %q, want %q", merged.Theme, tt.want)
			}
		})
	}
}

func TestMergeConfig_Extensions_Additive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		root    []ExtensionConfig
		module  []ExtensionConfig
		want    []string // expected names in order
		wantNil bool
	}{
		{
			"root_A_module_B",
			[]ExtensionConfig{{Name: "docker-sync"}},
			[]ExtensionConfig{{Name: "cargo-check"}},
			[]string{"docker-sync", "cargo-check"},
			false,
		},
		{
			"duplicate_by_name_module_wins",
			[]ExtensionConfig{{Name: "docker-sync", Path: "a"}},
			[]ExtensionConfig{{Name: "docker-sync", Path: "b"}},
			[]string{"docker-sync"},
			false,
		},
		{
			"module_empty",
			[]ExtensionConfig{{Name: "docker-sync"}},
			nil,
			[]string{"docker-sync"},
			false,
		},
		{
			"root_empty",
			nil,
			[]ExtensionConfig{{Name: "cargo-check"}},
			[]string{"cargo-check"},
			false,
		},
		{
			"both_empty",
			nil,
			nil,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := &Config{Extensions: tt.root}
			mod := &Config{Extensions: tt.module}
			merged := MergeConfig(root, mod)
			if tt.wantNil {
				if merged.Extensions != nil {
					t.Errorf("expected nil extensions, got %v", merged.Extensions)
				}
				return
			}
			if len(merged.Extensions) != len(tt.want) {
				t.Fatalf("expected %d extensions, got %d: %v", len(tt.want), len(merged.Extensions), merged.Extensions)
			}
			for i, name := range tt.want {
				if merged.Extensions[i].Name != name {
					t.Errorf("ext[%d] = %q, want %q", i, merged.Extensions[i].Name, name)
				}
			}
			// Check duplicate: module path wins
			if tt.name == "duplicate_by_name_module_wins" {
				if merged.Extensions[0].Path != "b" {
					t.Errorf("duplicate: expected module path 'b', got %q", merged.Extensions[0].Path)
				}
			}
		})
	}
}

func TestMergeConfig_PreReleaseHooks_Additive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		root   []map[string]PreReleaseHookConfig
		module []map[string]PreReleaseHookConfig
		want   int // expected total hooks
	}{
		{"root_plus_module", []map[string]PreReleaseHookConfig{{"lint": {Command: "lint"}}}, []map[string]PreReleaseHookConfig{{"test": {Command: "test"}}}, 2},
		{"module_only", nil, []map[string]PreReleaseHookConfig{{"test": {Command: "test"}}}, 1},
		{"both_empty", nil, nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := &Config{PreReleaseHooks: tt.root}
			mod := &Config{PreReleaseHooks: tt.module}
			merged := MergeConfig(root, mod)
			got := len(merged.PreReleaseHooks)
			if got != tt.want {
				t.Errorf("expected %d hooks, got %d", tt.want, got)
			}
		})
	}
}

func TestMergeConfig_DeepCopy(t *testing.T) {
	t.Parallel()

	t.Run("extensions_deep_copy", func(t *testing.T) {
		t.Parallel()
		root := &Config{Extensions: []ExtensionConfig{{Name: "original", Path: "/root"}}}
		mod := &Config{Extensions: []ExtensionConfig{{Name: "added"}}}
		merged := MergeConfig(root, mod)

		// Mutate merged
		merged.Extensions[0].Name = "mutated"

		// Root must be unchanged
		if root.Extensions[0].Name != "original" {
			t.Errorf("root was mutated: got %q, want 'original'", root.Extensions[0].Name)
		}
	})

	t.Run("hooks_deep_copy", func(t *testing.T) {
		t.Parallel()
		root := &Config{PreReleaseHooks: []map[string]PreReleaseHookConfig{{"lint": {Command: "lint"}}}}
		mod := &Config{}
		merged := MergeConfig(root, mod)

		// Mutate merged
		merged.PreReleaseHooks[0]["lint"] = PreReleaseHookConfig{Command: "mutated"}

		// Root must be unchanged
		if root.PreReleaseHooks[0]["lint"].Command != "lint" {
			t.Errorf("root was mutated: got %q, want 'lint'", root.PreReleaseHooks[0]["lint"].Command)
		}
	})
}

func TestMergeConfig_PathAlwaysRoot(t *testing.T) {
	t.Parallel()
	root := &Config{Path: ".version"}
	mod := &Config{Path: "custom.version"}
	merged := MergeConfig(root, mod)
	if merged.Path != ".version" {
		t.Errorf("Path = %q, want '.version' (root always)", merged.Path)
	}
}

func TestMergeConfig_WorkspaceAlwaysRoot(t *testing.T) {
	t.Parallel()
	rootWs := &WorkspaceConfig{Modules: []ModuleConfig{{Name: "root-mod"}}}
	modWs := &WorkspaceConfig{Modules: []ModuleConfig{{Name: "module-mod"}}}
	root := &Config{Workspace: rootWs}
	mod := &Config{Workspace: modWs}
	merged := MergeConfig(root, mod)
	if merged.Workspace != rootWs {
		t.Error("expected workspace from root, not module")
	}
}

func TestMergePluginConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		root       *Config
		module     *Config
		wantNil    bool
		wantPrefix string
		checkWs    bool // if true, check that workspace is from root
	}{
		{
			name:       "nil module returns root",
			root:       &Config{Plugins: &PluginConfig{TagManager: &TagManagerConfig{Prefix: "v"}}},
			module:     nil,
			wantPrefix: "v",
		},
		{
			name:       "nil root returns module",
			root:       nil,
			module:     &Config{Plugins: &PluginConfig{TagManager: &TagManagerConfig{Prefix: "mod-"}}},
			wantPrefix: "mod-",
		},
		{
			name:       "module prefix overrides root",
			root:       &Config{Plugins: &PluginConfig{TagManager: &TagManagerConfig{Prefix: "v"}}},
			module:     &Config{Plugins: &PluginConfig{TagManager: &TagManagerConfig{Prefix: "redis-"}}},
			wantPrefix: "redis-",
		},
		{
			name:       "module has no tag-manager",
			root:       &Config{Plugins: &PluginConfig{TagManager: &TagManagerConfig{Prefix: "v"}}},
			module:     &Config{Plugins: &PluginConfig{}},
			wantPrefix: "v",
		},
		{
			name:    "both nil",
			root:    nil,
			module:  nil,
			wantNil: true,
		},
		{
			name: "workspace not merged from module",
			root: &Config{
				Plugins:   &PluginConfig{TagManager: &TagManagerConfig{Prefix: "v"}},
				Workspace: &WorkspaceConfig{Modules: []ModuleConfig{{Name: "root-mod"}}},
			},
			module:     &Config{Plugins: &PluginConfig{}, Workspace: &WorkspaceConfig{Modules: []ModuleConfig{{Name: "module-mod"}}}},
			wantPrefix: "v",
			checkWs:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			merged := MergePluginConfig(tt.root, tt.module)

			if tt.wantNil {
				if merged != nil {
					t.Fatalf("expected nil, got %+v", merged)
				}
				return
			}

			if merged == nil {
				t.Fatal("expected non-nil result, got nil")
			}

			if tt.wantPrefix != "" {
				if merged.Plugins == nil || merged.Plugins.TagManager == nil {
					t.Fatal("expected tag-manager in merged config")
				}
				got := merged.Plugins.TagManager.GetPrefix()
				if got != tt.wantPrefix {
					t.Errorf("merged prefix = %q, want %q", got, tt.wantPrefix)
				}
			}

			if tt.checkWs {
				if merged.Workspace == nil {
					t.Fatal("expected workspace from root to be retained")
				}
				if len(merged.Workspace.Modules) != 1 || merged.Workspace.Modules[0].Name != "root-mod" {
					t.Errorf("workspace should come from root, got modules: %+v", merged.Workspace.Modules)
				}
			}
		})
	}
}
