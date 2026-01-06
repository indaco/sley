package precmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/testutils"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

func TestCLI_PreCommand_StaticLabel(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
	testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", "beta.1"}, tmpDir)
	content := testutils.ReadTempVersionFile(t, tmpDir)
	if got := content; got != "1.2.4-beta.1" {
		t.Errorf("expected 1.2.4-beta.1, got %q", got)
	}
}

func TestCLI_PreCommand_Increment(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3-beta.3")
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", "beta", "--inc"}, tmpDir)
	content := testutils.ReadTempVersionFile(t, tmpDir)
	if got := content; got != "1.2.3-beta.4" {
		t.Errorf("expected 1.2.3-beta.4, got %q", got)
	}
}

func TestCLI_PreCommand_AutoInitFeedback(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", "alpha"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expected := fmt.Sprintf("Auto-initialized %s with default version", versionPath)
	if !strings.Contains(output, expected) {
		t.Errorf("expected feedback %q, got %q", expected, output)
	}
}

func TestCLI_PreCommand_InvalidVersion(t *testing.T) {
	tmp := t.TempDir()
	customPath := filepath.Join(tmp, "bad.version")

	// Write invalid version string before CLI setup
	_ = os.WriteFile(customPath, []byte("not-a-version\n"), semver.VersionFilePerm)

	defaultPath := filepath.Join(tmp, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: defaultPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"sley", "pre", "--label", "alpha", "--path", customPath,
	})
	if err == nil {
		t.Fatal("expected error due to invalid version, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_PreCommand_SaveVersionFails(t *testing.T) {
	if os.Getenv("TEST_SLEY_PRE_SAVE_FAIL") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		if err := os.WriteFile(versionPath, []byte("1.2.3\n"), 0444); err != nil {
			fmt.Fprintln(os.Stderr, "failed to write .version file:", err)
			os.Exit(1)
		}

		if err := os.Chmod(versionPath, 0444); err != nil {
			fmt.Fprintln(os.Stderr, "failed to chmod .version file:", err)
			os.Exit(1)
		}

		// Prepare and run the CLI command
		cfg := &config.Config{Path: versionPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

		err := appCli.Run(context.Background(), []string{
			"sley", "pre", "--label", "rc", "--path", versionPath,
		})

		_ = os.Chmod(versionPath, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		os.Exit(0) // Unexpected success
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_PreCommand_SaveVersionFails")
	cmd.Env = append(os.Environ(), "TEST_SLEY_PRE_SAVE_FAIL=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected error due to save failure, got nil")
	}

	if !strings.Contains(string(output), "failed to save version") {
		t.Errorf("expected wrapped error message, got: %q", string(output))
	}
}

// TestCLI_PreCommand_FromCommandFails is obsolete - the pre command now uses
// GetExecutionContext instead of FromCommandFn, and auto-initializes missing files.
// Keeping this as a placeholder for potential future execution context error testing.

/* ------------------------------------------------------------------------- */
/* SINGLE-MODULE OPERATIONS - ADDITIONAL TESTS                              */
/* ------------------------------------------------------------------------- */

func TestCLI_PreCommand_StaticLabel_Variants(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	tests := []struct {
		name     string
		initial  string
		label    string
		expected string
	}{
		{"set alpha", "1.2.3", "alpha", "1.2.4-alpha"},
		{"set beta.1", "1.2.3", "beta.1", "1.2.4-beta.1"},
		{"set rc.1", "2.0.0", "rc.1", "2.0.1-rc.1"},
		{"replace existing pre-release", "1.2.3-alpha.1", "beta.1", "1.2.3-beta.1"},
		{"replace with same label increments patch", "1.0.0-rc.1", "rc.2", "1.0.0-rc.2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", tt.label}, tmpDir)
			content := testutils.ReadTempVersionFile(t, tmpDir)
			if content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, content)
			}
		})
	}
}

func TestCLI_PreCommand_Increment_Variants(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	tests := []struct {
		name     string
		initial  string
		label    string
		expected string
	}{
		{"increment beta.1 to beta.2", "1.2.3-beta.1", "beta", "1.2.3-beta.2"},
		{"increment rc.5 to rc.6", "2.0.0-rc.5", "rc", "2.0.0-rc.6"},
		{"increment alpha to alpha.1", "1.0.0-alpha", "alpha", "1.0.0-alpha.1"},
		{"start increment on final version", "1.2.3", "beta", "1.2.3-beta.1"},
		{"increment with no separator rc1 to rc2", "1.2.3-rc1", "rc", "1.2.3-rc2"},
		{"increment with dash separator rc-1 to rc-2", "1.2.3-rc-1", "rc", "1.2.3-rc-2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", tt.label, "--inc"}, tmpDir)
			content := testutils.ReadTempVersionFile(t, tmpDir)
			if content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, content)
			}
		})
	}
}

func TestCLI_PreCommand_LabelSwitchWithIncrement(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	tests := []struct {
		name     string
		initial  string
		label    string
		expected string
	}{
		{"switch from alpha to beta with inc", "1.2.3-alpha.5", "beta", "1.2.3-beta.1"},
		{"switch from rc to beta with inc", "1.2.3-rc.2", "beta", "1.2.3-beta.1"},
		{"same label increments", "1.2.3-beta.1", "beta", "1.2.3-beta.2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", tt.label, "--inc"}, tmpDir)
			content := testutils.ReadTempVersionFile(t, tmpDir)
			if content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, content)
			}
		})
	}
}

func TestCLI_PreCommand_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	tests := []struct {
		name     string
		initial  string
		args     []string
		expected string
	}{
		{
			name:     "complex label",
			initial:  "1.2.3",
			args:     []string{"sley", "pre", "--label", "alpha.beta.1"},
			expected: "1.2.4-alpha.beta.1",
		},
		{
			name:     "numeric only label",
			initial:  "1.2.3",
			args:     []string{"sley", "pre", "--label", "123"},
			expected: "1.2.4-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)
			content := testutils.ReadTempVersionFile(t, tmpDir)
			if content != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, content)
			}
		})
	}
}

func TestCLI_PreCommand_PermissionErrors(t *testing.T) {
	if testutils.IsWindows() {
		t.Skip("Skipping permission test on Windows")
	}

	tmp := t.TempDir()
	protectedDir := filepath.Join(tmp, "protected")
	if err := os.Mkdir(protectedDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(protectedDir, 0755) })

	versionPath := filepath.Join(protectedDir, ".version")

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"sley", "pre", "--label", "alpha", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected permission denied error, got: %v", err)
	}
}

/* ------------------------------------------------------------------------- */
/* MULTI-MODULE OPERATIONS TESTS                                            */
/* ------------------------------------------------------------------------- */

func TestCLI_PreCommand_MultiModule_All(t *testing.T) {
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
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0-alpha.1")

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
		testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", "beta", "--all"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify both modules were updated
	gotA := testutils.ReadTempVersionFile(t, moduleA)
	gotB := testutils.ReadTempVersionFile(t, moduleB)

	// Module A should have patch incremented then beta set
	if gotA != "1.0.1-beta" {
		t.Errorf("module-a version = %q, want %q", gotA, "1.0.1-beta")
	}
	// Module B should have existing pre-release replaced
	if gotB != "2.0.0-beta" {
		t.Errorf("module-b version = %q, want %q", gotB, "2.0.0-beta")
	}

	// Verify output contains operation name
	if !strings.Contains(output, "Set pre-release to beta") {
		t.Errorf("expected operation name in output, got: %q", output)
	}
}

func TestCLI_PreCommand_MultiModule_Increment(t *testing.T) {
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

	testutils.WriteTempVersionFile(t, moduleA, "1.0.0-rc.1")
	testutils.WriteTempVersionFile(t, moduleB, "2.0.0-rc.5")

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

	// Test with --all and --inc flags
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", "rc", "--inc", "--all"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify both modules were incremented
	gotA := testutils.ReadTempVersionFile(t, moduleA)
	gotB := testutils.ReadTempVersionFile(t, moduleB)

	if gotA != "1.0.0-rc.2" {
		t.Errorf("module-a version = %q, want %q", gotA, "1.0.0-rc.2")
	}
	if gotB != "2.0.0-rc.6" {
		t.Errorf("module-b version = %q, want %q", gotB, "2.0.0-rc.6")
	}

	// Verify output contains increment operation name
	if !strings.Contains(output, "Increment pre-release with rc") {
		t.Errorf("expected increment operation name in output, got: %q", output)
	}
}

func TestCLI_PreCommand_MultiModule_Specific(t *testing.T) {
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

	// Test with --module flag to target specific module
	testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", "alpha", "--module", "module-a"}, tmpDir)

	// Verify only module-a was updated
	gotA := testutils.ReadTempVersionFile(t, moduleA)
	gotB := testutils.ReadTempVersionFile(t, moduleB)

	if gotA != "1.0.1-alpha" {
		t.Errorf("module-a version = %q, want %q", gotA, "1.0.1-alpha")
	}
	if gotB != "2.0.0" {
		t.Errorf("module-b should remain %q, got %q", "2.0.0", gotB)
	}
}

func TestCLI_PreCommand_MultiModule_Quiet(t *testing.T) {
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
		testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", "rc", "--all", "--quiet"}, tmpDir)
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

	if gotA != "1.0.1-rc" || gotB != "2.0.1-rc" {
		t.Errorf("modules not updated correctly: module-a=%q, module-b=%q", gotA, gotB)
	}
}

func TestCLI_PreCommand_MultiModule_JSONFormat(t *testing.T) {
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
		testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", "beta", "--all", "--format", "json"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Output should contain JSON with module names
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "module-b") {
		t.Errorf("expected JSON output with module names, got: %q", output)
	}
}

func TestCLI_PreCommand_MultiModule_Parallel(t *testing.T) {
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
	testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", "alpha", "--all", "--parallel"}, tmpDir)

	// Verify all modules were updated
	gotA := testutils.ReadTempVersionFile(t, moduleA)
	gotB := testutils.ReadTempVersionFile(t, moduleB)
	gotC := testutils.ReadTempVersionFile(t, moduleC)

	if gotA != "1.0.1-alpha" || gotB != "2.0.1-alpha" || gotC != "3.0.1-alpha" {
		t.Errorf("modules not all updated: a=%q, b=%q, c=%q", gotA, gotB, gotC)
	}
}

func TestCLI_PreCommand_MultiModule_TextFormat(t *testing.T) {
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

	// Test with --format text
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "pre", "--label", "rc", "--all", "--format", "text"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify text format output
	if !strings.Contains(output, "module-a") || !strings.Contains(output, "module-b") {
		t.Errorf("expected text output with module names, got: %q", output)
	}
}

/* ------------------------------------------------------------------------- */
/* PRINT QUIET SUMMARY TESTS                                                */
/* ------------------------------------------------------------------------- */

func TestPrintQuietSummary(t *testing.T) {
	tests := []struct {
		name     string
		results  []workspace.ExecutionResult
		expected string
	}{
		{
			name: "all success",
			results: []workspace.ExecutionResult{
				{Module: &workspace.Module{Name: "mod1"}, Success: true},
				{Module: &workspace.Module{Name: "mod2"}, Success: true},
			},
			expected: "Success: 2 module(s) updated",
		},
		{
			name: "with failures",
			results: []workspace.ExecutionResult{
				{Module: &workspace.Module{Name: "mod1"}, Success: true},
				{Module: &workspace.Module{Name: "mod2"}, Success: false, Error: fmt.Errorf("failed")},
			},
			expected: "Completed: 1 succeeded, 1 failed",
		},
		{
			name: "all failures",
			results: []workspace.ExecutionResult{
				{Module: &workspace.Module{Name: "mod1"}, Success: false, Error: fmt.Errorf("error1")},
				{Module: &workspace.Module{Name: "mod2"}, Success: false, Error: fmt.Errorf("error2")},
			},
			expected: "Completed: 0 succeeded, 2 failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, _ := testutils.CaptureStdout(func() {
				printQuietSummary(tt.results)
			})
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}
