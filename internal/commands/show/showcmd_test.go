package show

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

func TestCLI_ShowCommand(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "9.8.7")
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "show"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if output != "9.8.7" {
		t.Errorf("expected output '9.8.7', got %q", output)
	}
}

func TestCLI_ShowCommand_Strict_MissingFile(t *testing.T) {
	if os.Getenv("TEST_SLEY_STRICT") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		// Prepare and run the CLI command
		cfg := &config.Config{Path: versionPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

		err := appCli.Run(context.Background(), []string{"sley", "show", "--strict"})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1) // Exit with non-zero status to signal error to parent process
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_ShowCommand_Strict_MissingFile")
	cmd.Env = append(os.Environ(), "TEST_SLEY_STRICT=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "version file not found"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got %q", expected, string(output))
	}
}

func TestCLI_ShowCommand_Strict_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "show", "--strict"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if output != "1.2.3" {
		t.Errorf("expected output '1.2.3', got %q", output)
	}
}

func TestCLI_ShowCommand_InvalidVersionContent(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	// Write an invalid version string
	if err := os.WriteFile(versionPath, []byte("not-a-semver\n"), 0644); err != nil {
		t.Fatalf("failed to write invalid version: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{"sley", "show"})
	if err == nil {
		t.Fatal("expected error due to invalid version, got nil")
	}

	if !strings.Contains(err.Error(), "failed to read version file at") {
		t.Errorf("unexpected error message: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("error does not mention 'invalid version format': %v", err)
	}
}

func TestCLI_ShowCommand_MultiModule_All(t *testing.T) {
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
		testutils.RunCLITest(t, appCli, []string{"sley", "show", "--all"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify output contains both module versions
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "1.0.0") {
		t.Errorf("expected output to contain module-a with version 1.0.0, got: %q", output)
	}
	if !strings.Contains(output, "module-b") || !strings.Contains(output, "2.0.0") {
		t.Errorf("expected output to contain module-b with version 2.0.0, got: %q", output)
	}
}

func TestCLI_ShowCommand_MultiModule_Specific(t *testing.T) {
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
		testutils.RunCLITest(t, appCli, []string{"sley", "show", "--module", "module-a"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify output contains only module-a
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "1.0.0") {
		t.Errorf("expected output to contain module-a with version 1.0.0, got: %q", output)
	}
}

func TestCLI_ShowCommand_MultiModule_Quiet(t *testing.T) {
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
		testutils.RunCLITest(t, appCli, []string{"sley", "show", "--all", "--quiet"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Quiet mode should show minimal output
	if !strings.Contains(output, "Success:") && !strings.Contains(output, "2 module(s)") {
		t.Errorf("expected quiet summary, got: %q", output)
	}
}

func TestCLI_ShowCommand_MultiModule_JSONFormat(t *testing.T) {
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
		testutils.RunCLITest(t, appCli, []string{"sley", "show", "--all", "--format", "json"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Output should contain JSON
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "module-b") {
		t.Errorf("expected JSON output with module names, got: %q", output)
	}
}

func TestCLI_ShowCommand_MultiModule_TableFormat(t *testing.T) {
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
		testutils.RunCLITest(t, appCli, []string{"sley", "show", "--all", "--format", "table"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Output should contain table-formatted data
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "module-b") {
		t.Errorf("expected table output with module names, got: %q", output)
	}
}

func TestCLI_ShowCommand_MultiModule_TableFormat2(t *testing.T) {
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

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0-alpha.1")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0+build.123")
	testutils.WriteTempVersionFile(t, moduleC, "3.5.7-rc.2+meta")

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
		testutils.RunCLITest(t, appCli, []string{"sley", "show", "--all", "--format", "table"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Output should contain all module versions
	if !strings.Contains(output, "1.0.0-alpha.1") {
		t.Errorf("expected table output with module-a version, got: %q", output)
	}
	if !strings.Contains(output, "2.0.0+build.123") {
		t.Errorf("expected table output with module-b version, got: %q", output)
	}
	if !strings.Contains(output, "3.5.7-rc.2+meta") {
		t.Errorf("expected table output with module-c version, got: %q", output)
	}
}

func TestCLI_ShowCommand_MultiModule_Parallel(t *testing.T) {
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
				Enabled:   &enabled,
				Recursive: &recursive,
				MaxDepth:  &maxDepth,
			},
		},
	}

	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with --parallel flag
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "show", "--all", "--parallel"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify all modules are shown
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "1.0.0") {
		t.Errorf("expected output to contain module-a, got: %q", output)
	}
	if !strings.Contains(output, "module-b") || !strings.Contains(output, "2.0.0") {
		t.Errorf("expected output to contain module-b, got: %q", output)
	}
	if !strings.Contains(output, "module-c") || !strings.Contains(output, "3.0.0") {
		t.Errorf("expected output to contain module-c, got: %q", output)
	}
}

func TestCLI_ShowCommand_MultiModule_SuccessQuiet(t *testing.T) {
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

	testutils.WriteTempVersionFile(t, moduleA, "1.2.3")
	testutils.WriteTempVersionFile(t, moduleB, "4.5.6")

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

	// Test with --quiet flag showing success
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "show", "--all", "--quiet"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Quiet success mode should show module count
	if !strings.Contains(output, "Success:") || !strings.Contains(output, "2 module(s)") {
		t.Errorf("expected quiet success summary with 2 modules, got: %q", output)
	}
}
