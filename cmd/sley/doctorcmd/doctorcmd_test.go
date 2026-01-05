package doctorcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestCLI_ValidateCommand_ValidCases(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	tests := []struct {
		name           string
		version        string
		expectedOutput string
	}{
		{
			name:    "valid semantic version",
			version: "1.2.3",
		},
		{
			name:    "valid version with build metadata",
			version: "1.2.3+exp.sha.5114f85",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.version)

			output, err := testutils.CaptureStdout(func() {
				testutils.RunCLITest(t, appCli, []string{"sley", "doctor"}, tmpDir)
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			expected := fmt.Sprintf("Valid version file at %s/.version", tmpDir)
			if !strings.Contains(output, expected) {
				t.Errorf("expected output to contain %q, got %q", expected, output)
			}
		})
	}
}

func TestCLI_ValidateCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		expectedError string
	}{
		{"invalid version string", "not-a-version", "invalid version"},
		{"invalid build metadata", "1.0.0+inv@lid-meta", "invalid version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testutils.WriteTempVersionFile(t, tmpDir, tt.version)
			versionPath := filepath.Join(tmpDir, ".version")

			// Prepare and run the CLI command
			cfg := &config.Config{Path: versionPath}
			appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

			err := appCli.Run(context.Background(), []string{"sley", "doctor"})
			if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
				t.Fatalf("expected error containing %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestCLI_ValidateCommand_MultiModule_All(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --all flag
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--all"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify output contains both modules
	if !strings.Contains(output, "module-a") {
		t.Errorf("expected output to contain module-a, got: %q", output)
	}
	if !strings.Contains(output, "module-b") {
		t.Errorf("expected output to contain module-b, got: %q", output)
	}
}

func TestCLI_ValidateCommand_MultiModule_Specific(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --module flag (target specific module)
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--module", "module-a"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify output contains only module-a
	if !strings.Contains(output, "module-a") {
		t.Errorf("expected output to contain module-a, got: %q", output)
	}
}

func TestCLI_ValidateCommand_MultiModule_Quiet(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --quiet flag
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--all", "--quiet"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Quiet mode should show minimal output
	if !strings.Contains(output, "Success:") && !strings.Contains(output, "2 module(s)") {
		t.Errorf("expected quiet summary, got: %q", output)
	}
}

func TestCLI_ValidateCommand_MultiModule_WithInvalidVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace with one invalid version
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "not-a-version")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --all flag - should fail due to invalid version
	// We need to run in the tmpDir context
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	err = appCli.Run(context.Background(), []string{"sley", "doctor", "--all"})
	if err == nil {
		t.Fatal("expected error due to invalid version in one module, got nil")
	}

	if !strings.Contains(err.Error(), "failed validation") {
		t.Errorf("expected error message to contain 'failed validation', got: %v", err)
	}
}

func TestCLI_ValidateCommand_MultiModule_ContinueOnError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace with one invalid version
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	moduleC := filepath.Join(tmpDir, "module-c")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleC, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "not-a-version")
	testutils.WriteTempVersionFile(t, moduleC, "3.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --continue-on-error flag
	// We need to run in the tmpDir context
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	output, _ := testutils.CaptureStdout(func() {
		_ = appCli.Run(context.Background(), []string{"sley", "doctor", "--all", "--continue-on-error"})
	})

	// Should show results for all modules including the failed one
	if !strings.Contains(output, "module-a") {
		t.Errorf("expected output to contain module-a, got: %q", output)
	}
	if !strings.Contains(output, "module-c") {
		t.Errorf("expected output to contain module-c, got: %q", output)
	}
}

func TestCLI_ValidateCommand_MultiModule_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --format json
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--all", "--format", "json"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Output should contain JSON
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "module-b") {
		t.Errorf("expected JSON output with module names, got: %q", output)
	}
}

func TestCLI_ValidateCommand_MultiModule_TableFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace
	moduleA := filepath.Join(tmpDir, "module-a")
	moduleB := filepath.Join(tmpDir, "module-b")
	if err := os.MkdirAll(moduleA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(moduleB, 0755); err != nil {
		t.Fatal(err)
	}

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --format table
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "doctor", "--all", "--format", "table"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Output should contain table-formatted data
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "module-b") {
		t.Errorf("expected table output with module names, got: %q", output)
	}
}
