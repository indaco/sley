package set

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestCLI_SetVersionCommandVariants(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"set version", []string{"sley", "set", "2.5.0"}, "2.5.0"},
		{"set with pre-release", []string{"sley", "set", "3.0.0", "--pre", "beta.2"}, "3.0.0-beta.2"},
		{"set with metadata", []string{"sley", "set", "1.0.0", "--meta", "001"}, "1.0.0+001"},
		{"set with pre and meta", []string{"sley", "set", "1.0.0", "--pre", "alpha.1", "--meta", "exp.sha.5114f85"}, "1.0.0-alpha.1+exp.sha.5114f85"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)
			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCLI_SetVersionCommand_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{"sley", "set", "invalid.version"})
	if err == nil {
		t.Fatal("expected error due to invalid version format, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_SetVersionCommand_MissingArgument(t *testing.T) {
	if os.Getenv("TEST_SLEY_SET_MISSING_ARG") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		// Prepare and run the CLI command
		cfg := &config.Config{Path: versionPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

		err := appCli.Run(context.Background(), []string{"sley", "set", "--path", versionPath})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1) // expected non-zero exit
		}
		os.Exit(0) // should not happen
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_SetVersionCommand_MissingArgument")
	cmd.Env = append(os.Environ(), "TEST_SLEY_SET_MISSING_ARG=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "missing required version argument"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got %q", expected, string(output))
	}
}

func TestCLI_SetVersionCommand_SaveError(t *testing.T) {
	tmp := t.TempDir()

	protectedDir := filepath.Join(tmp, "protected")
	if err := os.Mkdir(protectedDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(protectedDir, 0755)
	})

	versionPath := filepath.Join(protectedDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"sley", "set", "3.0.0", "--path", versionPath,
	})
	if err == nil {
		t.Fatal("expected error due to save failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to save version") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCLI_SetVersionCommand_MultiModule_All(t *testing.T) {
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
				Enabled:        &enabled,
				Recursive:      &recursive,
				ModuleMaxDepth: &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --all flag (non-interactive)
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "set", "3.0.0", "--all"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify both modules were updated
	gotA := testutils.ReadTempVersionFile(t, moduleA)
	gotB := testutils.ReadTempVersionFile(t, moduleB)

	if gotA != "3.0.0" {
		t.Errorf("module-a version = %q, want %q", gotA, "3.0.0")
	}
	if gotB != "3.0.0" {
		t.Errorf("module-b version = %q, want %q", gotB, "3.0.0")
	}

	// Verify output contains success message
	if !strings.Contains(output, "Set version to 3.0.0") || !strings.Contains(output, "module-a") {
		t.Errorf("expected success output, got: %q", output)
	}
}

func TestCLI_SetVersionCommand_MultiModule_Specific(t *testing.T) {
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
				Enabled:        &enabled,
				Recursive:      &recursive,
				ModuleMaxDepth: &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --module flag (target specific module)
	testutils.RunCLITest(t, appCli, []string{"sley", "set", "5.0.0", "--module", "module-a"}, tmpDir)

	// Verify only module-a was updated
	gotA := testutils.ReadTempVersionFile(t, moduleA)
	gotB := testutils.ReadTempVersionFile(t, moduleB)

	if gotA != "5.0.0" {
		t.Errorf("module-a version = %q, want %q", gotA, "5.0.0")
	}
	if gotB != "2.0.0" {
		t.Errorf("module-b version should remain %q, got %q", "2.0.0", gotB)
	}
}

func TestCLI_SetVersionCommand_MultiModule_WithPreAndMeta(t *testing.T) {
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
				Enabled:        &enabled,
				Recursive:      &recursive,
				ModuleMaxDepth: &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --all flag and pre-release + metadata
	testutils.RunCLITest(t, appCli, []string{
		"sley", "set", "4.0.0",
		"--pre", "beta.1",
		"--meta", "build.123",
		"--all",
	}, tmpDir)

	// Verify both modules were updated with pre-release and metadata
	gotA := testutils.ReadTempVersionFile(t, moduleA)
	gotB := testutils.ReadTempVersionFile(t, moduleB)

	expected := "4.0.0-beta.1+build.123"
	if gotA != expected {
		t.Errorf("module-a version = %q, want %q", gotA, expected)
	}
	if gotB != expected {
		t.Errorf("module-b version = %q, want %q", gotB, expected)
	}
}

func TestCLI_SetVersionCommand_MultiModule_Quiet(t *testing.T) {
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
				Enabled:        &enabled,
				Recursive:      &recursive,
				ModuleMaxDepth: &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --quiet flag
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "set", "6.0.0", "--all", "--quiet"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Quiet mode should show minimal output
	if !strings.Contains(output, "Success:") && !strings.Contains(output, "2 module(s)") {
		t.Errorf("expected quiet summary, got: %q", output)
	}

	// Verify versions were still updated
	gotA := testutils.ReadTempVersionFile(t, moduleA)
	gotB := testutils.ReadTempVersionFile(t, moduleB)

	if gotA != "6.0.0" || gotB != "6.0.0" {
		t.Errorf("modules not updated correctly: module-a=%q, module-b=%q", gotA, gotB)
	}
}

func TestCLI_SetVersionCommand_MultiModule_JSONFormat(t *testing.T) {
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
				Enabled:        &enabled,
				Recursive:      &recursive,
				ModuleMaxDepth: &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --format json
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "set", "7.0.0", "--all", "--format", "json"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Output should contain JSON
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "module-b") {
		t.Errorf("expected JSON output with module names, got: %q", output)
	}
}

func TestCLI_SetVersionCommand_MultiModule_Parallel(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-module workspace with 3 modules
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
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0")
	testutils.WriteTempVersionFile(t, moduleC, "3.0.0")

	// Create config with workspace discovery enabled
	enabled := true
	recursive := true
	maxDepth := 10
	cfg := &config.Config{
		Path: ".version",
		Workspace: &config.WorkspaceConfig{
			Discovery: &config.DiscoveryConfig{
				Enabled:        &enabled,
				Recursive:      &recursive,
				ModuleMaxDepth: &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --parallel flag
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "set", "10.0.0", "--all", "--parallel"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify all modules were updated
	gotA := testutils.ReadTempVersionFile(t, moduleA)
	gotB := testutils.ReadTempVersionFile(t, moduleB)
	gotC := testutils.ReadTempVersionFile(t, moduleC)

	if gotA != "10.0.0" || gotB != "10.0.0" || gotC != "10.0.0" {
		t.Errorf("modules not all updated to 10.0.0: a=%q, b=%q, c=%q", gotA, gotB, gotC)
	}

	// Should show success message
	if !strings.Contains(output, "Set version to 10.0.0") {
		t.Errorf("expected success message, got: %q", output)
	}
}

func TestCLI_SetVersionCommand_MultiModule_TextFormat(t *testing.T) {
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
				Enabled:        &enabled,
				Recursive:      &recursive,
				ModuleMaxDepth: &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --format text (default)
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "set", "11.0.0", "--all", "--format", "text"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify text format output
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "module-b") {
		t.Errorf("expected text output with module names, got: %q", output)
	}
}
