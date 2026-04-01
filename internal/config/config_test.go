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
