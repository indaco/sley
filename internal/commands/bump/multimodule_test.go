package bump

import (
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
