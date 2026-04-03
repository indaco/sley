package bump

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* GET FIRST SUCCESSFUL VERSION TESTS                                        */
/* ------------------------------------------------------------------------- */

func TestGetFirstSuccessfulVersion(t *testing.T) {

	tests := []struct {
		name     string
		results  []workspace.ExecutionResult
		expected string
	}{
		{
			name:     "empty results returns empty",
			results:  []workspace.ExecutionResult{},
			expected: "",
		},
		{
			name: "all failures returns empty",
			results: []workspace.ExecutionResult{
				{Success: false, NewVersion: "1.0.0"},
				{Success: false, NewVersion: "2.0.0"},
			},
			expected: "",
		},
		{
			name: "first success returns version",
			results: []workspace.ExecutionResult{
				{Success: true, NewVersion: "1.0.1"},
				{Success: true, NewVersion: "2.0.1"},
			},
			expected: "1.0.1",
		},
		{
			name: "mixed results returns first success",
			results: []workspace.ExecutionResult{
				{Success: false, NewVersion: "1.0.0"},
				{Success: true, NewVersion: "2.0.1"},
				{Success: true, NewVersion: "3.0.1"},
			},
			expected: "2.0.1",
		},
		{
			name: "success with empty version",
			results: []workspace.ExecutionResult{
				{Success: true, NewVersion: ""},
				{Success: true, NewVersion: "2.0.1"},
			},
			expected: "2.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result := getFirstSuccessfulVersion(tt.results)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* TEST HELPERS                                                              */
/* ------------------------------------------------------------------------- */

// setupMultiModuleWorkspaceWithVersion creates a workspace with multiple modules.
func setupMultiModuleWorkspaceWithVersion(t *testing.T, dir string, modules map[string]string) {
	t.Helper()
	for modulePath, version := range modules {
		moduleDir := filepath.Join(dir, modulePath)
		if err := os.MkdirAll(moduleDir, 0755); err != nil {
			t.Fatalf("failed to create module dir %s: %v", moduleDir, err)
		}
		versionFile := filepath.Join(moduleDir, ".version")
		if err := os.WriteFile(versionFile, []byte(version+"\n"), 0644); err != nil {
			t.Fatalf("failed to write version file %s: %v", versionFile, err)
		}
	}
}

// readModuleVersionFromDir reads the version of a specific module.
func readModuleVersionFromDir(t *testing.T, dir, modulePath string) string {
	t.Helper()
	versionFile := filepath.Join(dir, modulePath, ".version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("failed to read version file %s: %v", versionFile, err)
	}
	return strings.TrimSpace(string(data))
}

// buildMultiModuleCLI creates a CLI app configured for multi-module mode.
// It uses an empty default path so the detector can discover modules.
func buildMultiModuleCLI(cfg *config.Config, registry *plugins.PluginRegistry) *cli.Command {
	return &cli.Command{
		Name: "sley",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "Path to .version file",
				Value:   ".version", // Default, will be auto-detected
			},
			&cli.BoolFlag{
				Name:    "strict",
				Aliases: []string{"no-auto-init"},
				Usage:   "Fail if .version file is missing (disable auto-initialization)",
			},
		},
		Commands: []*cli.Command{Run(cfg, registry)},
	}
}

/* ------------------------------------------------------------------------- */
/* MODULE PATH DERIVATION TESTS                                              */
/* ------------------------------------------------------------------------- */

func TestModulePathDerivation(t *testing.T) {
	tests := []struct {
		name     string
		relPath  string
		wantDir  string
		wantPath string // after the "." → "" normalization
	}{
		{
			name:     "nested module returns parent dir",
			relPath:  "cobra/.version",
			wantDir:  "cobra",
			wantPath: "cobra",
		},
		{
			name:     "root version file returns dot then empty",
			relPath:  ".version",
			wantDir:  ".",
			wantPath: "",
		},
		{
			name:     "deep nested packages/core",
			relPath:  "packages/core/.version",
			wantDir:  "packages/core",
			wantPath: "packages/core",
		},
		{
			name:     "deep nested services/api",
			relPath:  "services/api/.version",
			wantDir:  "services/api",
			wantPath: "services/api",
		},
		{
			name:     "three levels deep",
			relPath:  "a/b/c/.version",
			wantDir:  "a/b/c",
			wantPath: "a/b/c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Dir(tt.relPath)
			if dir != tt.wantDir {
				t.Errorf("filepath.Dir(%q) = %q, want %q", tt.relPath, dir, tt.wantDir)
			}

			// Apply the same normalization as runMultiModuleBump
			modulePath := dir
			if modulePath == "." {
				modulePath = ""
			}
			if modulePath != tt.wantPath {
				t.Errorf("normalized modulePath = %q, want %q", modulePath, tt.wantPath)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* PER-MODULE CONFIG MERGE TESTS                                             */
/* ------------------------------------------------------------------------- */

// writeConfigFile writes a .sley.yaml file to the given directory.
func writeConfigFile(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".sley.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .sley.yaml in %s: %v", dir, err)
	}
}

func TestPerModuleConfigMerge(t *testing.T) {
	tests := []struct {
		name        string
		rootYAML    string
		moduleYAML  string // empty means no module config file
		hasModule   bool
		checkMerged func(t *testing.T, merged *config.Config)
	}{
		{
			name:       "module overrides changelog generator config",
			rootYAML:   "plugins:\n  changelog-generator:\n    enabled: true\n    mode: unified\n",
			moduleYAML: "plugins:\n  changelog-generator:\n    enabled: true\n    mode: versioned\n",
			hasModule:  true,
			checkMerged: func(t *testing.T, merged *config.Config) {
				t.Helper()
				if merged.Plugins == nil {
					t.Fatal("expected Plugins to be non-nil")
				}
				if merged.Plugins.ChangelogGenerator == nil {
					t.Fatal("expected ChangelogGenerator to be non-nil")
				}
				if merged.Plugins.ChangelogGenerator.Mode != "versioned" {
					t.Errorf("expected changelog mode %q, got %q", "versioned", merged.Plugins.ChangelogGenerator.Mode)
				}
			},
		},
		{
			name:      "module has no config file so root is used as-is",
			rootYAML:  "plugins:\n  tag-manager:\n    enabled: true\n    prefix: \"v\"\n",
			hasModule: false,
			checkMerged: func(t *testing.T, merged *config.Config) {
				t.Helper()
				if merged.Plugins == nil {
					t.Fatal("expected Plugins to be non-nil")
				}
				if merged.Plugins.TagManager == nil {
					t.Fatal("expected TagManager to be non-nil")
				}
				if merged.Plugins.TagManager.Prefix != "v" {
					t.Errorf("expected tag prefix %q, got %q", "v", merged.Plugins.TagManager.Prefix)
				}
			},
		},
		{
			name:       "module overrides tag prefix",
			rootYAML:   "plugins:\n  tag-manager:\n    enabled: true\n    prefix: \"v\"\n",
			moduleYAML: "plugins:\n  tag-manager:\n    enabled: true\n    prefix: \"mymod/v\"\n",
			hasModule:  true,
			checkMerged: func(t *testing.T, merged *config.Config) {
				t.Helper()
				if merged.Plugins == nil {
					t.Fatal("expected Plugins to be non-nil")
				}
				if merged.Plugins.TagManager == nil {
					t.Fatal("expected TagManager to be non-nil")
				}
				if merged.Plugins.TagManager.Prefix != "mymod/v" {
					t.Errorf("expected tag prefix %q, got %q", "mymod/v", merged.Plugins.TagManager.Prefix)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := loadAndMergeConfigs(t, tt.rootYAML, tt.moduleYAML, tt.hasModule)
			tt.checkMerged(t, merged)
		})
	}
}

// loadAndMergeConfigs creates temp dirs, writes config files, loads and merges them.
func loadAndMergeConfigs(t *testing.T, rootYAML, moduleYAML string, hasModule bool) *config.Config {
	t.Helper()
	tmpDir := t.TempDir()

	rootDir := filepath.Join(tmpDir, "root")
	writeConfigFile(t, rootDir, rootYAML)

	rootCfg, err := config.LoadConfigFromDir(rootDir)
	if err != nil {
		t.Fatalf("failed to load root config: %v", err)
	}
	if rootCfg == nil {
		t.Fatal("root config is nil")
	}

	if !hasModule {
		moduleDir := filepath.Join(tmpDir, "module-no-config")
		if err := os.MkdirAll(moduleDir, 0755); err != nil {
			t.Fatalf("failed to create module dir: %v", err)
		}
		moduleCfg, err := config.LoadConfigFromDir(moduleDir)
		if err != nil {
			t.Fatalf("unexpected error loading absent module config: %v", err)
		}
		if moduleCfg != nil {
			return config.MergeConfig(rootCfg, moduleCfg)
		}
		return rootCfg
	}

	moduleDir := filepath.Join(tmpDir, "module")
	writeConfigFile(t, moduleDir, moduleYAML)

	moduleCfg, err := config.LoadConfigFromDir(moduleDir)
	if err != nil {
		t.Fatalf("failed to load module config: %v", err)
	}
	if moduleCfg == nil {
		t.Fatal("module config is nil but expected to exist")
	}
	return config.MergeConfig(rootCfg, moduleCfg)
}

/* ------------------------------------------------------------------------- */
/* POST-BUMP ERROR COLLECTION TESTS                                          */
/* ------------------------------------------------------------------------- */

func TestPostBumpErrorCollection(t *testing.T) {
	tests := []struct {
		name         string
		moduleErrors []struct {
			moduleName string
			err        error
		}
		wantCount    int
		wantFirstMsg string // substring expected in the wrapped error
	}{
		{
			name:         "no errors produces nil",
			moduleErrors: nil,
			wantCount:    0,
		},
		{
			name: "single module error",
			moduleErrors: []struct {
				moduleName string
				err        error
			}{
				{moduleName: "core", err: fmt.Errorf("changelog generation failed")},
			},
			wantCount:    1,
			wantFirstMsg: "core",
		},
		{
			name: "multiple module errors all collected",
			moduleErrors: []struct {
				moduleName string
				err        error
			}{
				{moduleName: "core", err: fmt.Errorf("changelog failed")},
				{moduleName: "api", err: fmt.Errorf("tag creation failed")},
				{moduleName: "web", err: fmt.Errorf("audit log failed")},
			},
			wantCount:    3,
			wantFirstMsg: "core",
		},
		{
			name: "errors from different phases are collected together",
			moduleErrors: []struct {
				moduleName string
				err        error
			}{
				{moduleName: "alpha", err: fmt.Errorf("post-bump actions: dep-sync error")},
				{moduleName: "beta", err: fmt.Errorf("commit/tag: git tag failed")},
			},
			wantCount:    2,
			wantFirstMsg: "alpha",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the error collection pattern from runMultiModuleBump
			var postBumpErrors []error
			for _, me := range tt.moduleErrors {
				postBumpErrors = append(postBumpErrors,
					fmt.Errorf("module %s: %w", me.moduleName, me.err))
			}

			if tt.wantCount == 0 {
				if len(postBumpErrors) != 0 {
					t.Fatalf("expected no errors, got %d", len(postBumpErrors))
				}
				return
			}

			if len(postBumpErrors) != tt.wantCount {
				t.Fatalf("expected %d errors, got %d", tt.wantCount, len(postBumpErrors))
			}

			// Replicate the final error wrapping from runMultiModuleBump
			finalErr := fmt.Errorf("%d module(s) had post-bump errors: %w",
				len(postBumpErrors), postBumpErrors[0])

			if finalErr == nil {
				t.Fatal("expected non-nil final error")
			}

			errMsg := finalErr.Error()
			wantPrefix := fmt.Sprintf("%d module(s) had post-bump errors", tt.wantCount)
			if !strings.Contains(errMsg, wantPrefix) {
				t.Errorf("error message %q does not contain %q", errMsg, wantPrefix)
			}
			if !strings.Contains(errMsg, tt.wantFirstMsg) {
				t.Errorf("error message %q does not contain first module name %q", errMsg, tt.wantFirstMsg)
			}

			// Verify all individual errors were collected (not lost)
			for i, me := range tt.moduleErrors {
				if !strings.Contains(postBumpErrors[i].Error(), me.moduleName) {
					t.Errorf("error[%d] = %q, expected it to contain module name %q",
						i, postBumpErrors[i].Error(), me.moduleName)
				}
			}
		})
	}
}
