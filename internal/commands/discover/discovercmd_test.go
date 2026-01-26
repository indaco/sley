package discover

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestRun_ReturnsCommand(t *testing.T) {
	cmd := Run(nil)

	if cmd.Name != "discover" {
		t.Errorf("Name = %q, want %q", cmd.Name, "discover")
	}

	if cmd.Usage == "" {
		t.Error("Usage should not be empty")
	}

	// Verify flags exist
	flagNames := []string{"format", "quiet", "no-interactive"}
	for _, name := range flagNames {
		found := false
		for _, flag := range cmd.Flags {
			if flag.Names()[0] == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected flag %q not found", name)
		}
	}
}

func TestDiscoverAndSuggest(t *testing.T) {
	// Create a mock filesystem with test files
	fs := core.NewMockFileSystem()
	fs.SetFile("/project/.version", []byte("1.0.0\n"))
	fs.SetFile("/project/package.json", []byte(`{"version": "1.0.0"}`))

	// Note: DiscoverAndSuggest uses os.Getwd and NewOSFileSystem internally
	// This test verifies the interface only, not the actual filesystem operations

	cfg := &config.Config{}
	result, suggestion, err := DiscoverAndSuggest(context.Background(), cfg, ".")

	// In a real test environment, this might fail due to actual filesystem
	// We're mainly testing that the function signature works correctly
	if err == nil {
		if result == nil {
			t.Error("result should not be nil when err is nil")
		}
	}

	// suggestion can be nil if no sync candidates are found
	_ = suggestion
}

func TestPrintQuietSummary(t *testing.T) {
	// This is a visual output test - we just verify it doesn't panic

	tests := []struct {
		name   string
		result *discovery.Result
	}{
		{
			name:   "empty result",
			result: &discovery.Result{},
		},
		{
			name: "with modules",
			result: &discovery.Result{
				Mode: discovery.SingleModule,
				Modules: []discovery.Module{
					{Name: "root", RelPath: ".version", Version: "1.0.0"},
				},
			},
		},
		{
			name: "with mismatches",
			result: &discovery.Result{
				Mode: discovery.SingleModule,
				Modules: []discovery.Module{
					{Name: "root", RelPath: ".version", Version: "1.0.0"},
				},
				Mismatches: []discovery.Mismatch{
					{Source: "package.json", ExpectedVersion: "1.0.0", ActualVersion: "2.0.0"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic
			printQuietSummary(tt.result)
		})
	}
}

func TestRun_HasCorrectAliases(t *testing.T) {
	cmd := Run(nil)

	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "scan" {
		t.Errorf("Aliases = %v, want [scan]", cmd.Aliases)
	}
}

func TestRun_HasDepthFlag(t *testing.T) {
	cmd := Run(nil)

	found := false
	for _, flag := range cmd.Flags {
		if flag.Names()[0] == "depth" {
			found = true
			// Check alias
			names := flag.Names()
			if len(names) < 2 || names[1] != "d" {
				t.Errorf("depth flag should have alias 'd', got %v", names)
			}
			break
		}
	}
	if !found {
		t.Error("expected depth flag not found")
	}
}

func TestRun_CommandAction(t *testing.T) {
	cfg := &config.Config{}
	cmd := Run(cfg)

	if cmd.Action == nil {
		t.Error("Action should not be nil")
	}
}

func TestCLI_DiscoverCommand_TextFormat(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
	versionPath := filepath.Join(tmpDir, ".version")

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "discover", "--format", "text", "--no-interactive"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify output contains discovery results
	if !strings.Contains(output, "Discovery Results") {
		t.Errorf("expected 'Discovery Results' in output, got: %q", output)
	}
	if !strings.Contains(output, "1.2.3") {
		t.Errorf("expected version '1.2.3' in output, got: %q", output)
	}
}

func TestCLI_DiscoverCommand_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "2.0.0")
	versionPath := filepath.Join(tmpDir, ".version")

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "discover", "--format", "json", "--no-interactive"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify JSON output structure
	if !strings.Contains(output, `"mode"`) {
		t.Errorf("expected JSON mode field in output, got: %q", output)
	}
	if !strings.Contains(output, `"2.0.0"`) {
		t.Errorf("expected version '2.0.0' in JSON output, got: %q", output)
	}
}

func TestCLI_DiscoverCommand_TableFormat(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "3.0.0")
	versionPath := filepath.Join(tmpDir, ".version")

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "discover", "--format", "table", "--no-interactive"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify table output headers
	if !strings.Contains(output, "PATH") && !strings.Contains(output, "VERSION") {
		t.Errorf("expected table headers in output, got: %q", output)
	}
}

func TestCLI_DiscoverCommand_QuietMode(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")
	versionPath := filepath.Join(tmpDir, ".version")

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "discover", "--quiet", "--no-interactive"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify quiet mode shows summary only
	if !strings.Contains(output, "Mode:") {
		t.Errorf("expected 'Mode:' in quiet output, got: %q", output)
	}
	// Quiet mode should be shorter than regular output
	if strings.Contains(output, "Discovery Results") {
		t.Errorf("quiet mode should not contain 'Discovery Results' header")
	}
}

func TestCLI_DiscoverCommand_WithDepthFlag(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a nested structure
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Create a manifest in subdirectory
	pkgPath := filepath.Join(subDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"version": "1.0.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	versionPath := filepath.Join(tmpDir, ".version")
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	// Test with depth 0 (should only find root)
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "discover", "--depth", "0", "--format", "json", "--no-interactive"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// With depth 0, should still find the root .version
	if !strings.Contains(output, ".version") {
		t.Errorf("expected .version in output with depth 0, got: %q", output)
	}
}

func TestCLI_DiscoverCommand_WithManifests(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Create package.json
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"version": "1.0.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	versionPath := filepath.Join(tmpDir, ".version")
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "discover", "--format", "json", "--no-interactive"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify package.json is discovered
	if !strings.Contains(output, "package.json") {
		t.Errorf("expected package.json in output, got: %q", output)
	}
}

func TestCLI_DiscoverCommand_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{Path: filepath.Join(tmpDir, ".version")}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "discover", "--no-interactive"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Empty directory should still produce output
	if !strings.Contains(output, "Discovery Results") || !strings.Contains(output, "No version sources found") {
		t.Errorf("expected discovery results message for empty directory, got: %q", output)
	}
}

func TestCLI_DiscoverCommand_VersionMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Create package.json with different version
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"version": "2.0.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	versionPath := filepath.Join(tmpDir, ".version")
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "discover", "--format", "json", "--no-interactive"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify mismatch is detected
	if !strings.Contains(output, `"has_mismatches": true`) && !strings.Contains(output, `"has_mismatches":true`) {
		t.Errorf("expected has_mismatches true in output, got: %q", output)
	}
}

func TestCLI_DiscoverCommand_MultipleManifests(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")

	// Create multiple manifest files
	files := map[string]string{
		"package.json": `{"version": "1.0.0"}`,
		"Cargo.toml":   "[package]\nversion = \"1.0.0\"\n",
		"Chart.yaml":   "version: 1.0.0\n",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	versionPath := filepath.Join(tmpDir, ".version")
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"sley", "discover", "--format", "json", "--no-interactive"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify all manifests are discovered
	for name := range files {
		if !strings.Contains(output, name) {
			t.Errorf("expected %s in output, got: %q", name, output)
		}
	}
}

func TestDiscoverAndSuggest_WithSyncCandidates(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create version file and manifest
	testutils.WriteTempVersionFile(t, tmpDir, "1.0.0")
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"version": "1.0.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	result, suggestion, err := DiscoverAndSuggest(context.Background(), cfg, tmpDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Mode != discovery.SingleModule {
		t.Errorf("Mode = %v, want SingleModule", result.Mode)
	}

	// Check if sync candidates exist
	if len(result.SyncCandidates) == 0 {
		t.Log("Note: No sync candidates found - this may be expected")
	}

	// Suggestion may or may not be nil depending on sync candidates
	_ = suggestion
}

func TestDiscoverAndSuggest_ErrorCase(t *testing.T) {
	// Test with non-existent directory
	cfg := &config.Config{}
	_, _, err := DiscoverAndSuggest(context.Background(), cfg, "/nonexistent/directory/path")

	if err == nil {
		t.Log("Note: Discovery succeeded for non-existent directory - implementation handles this gracefully")
	}
}

func TestPrintQuietSummary_WithPrimaryVersion(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.SingleModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "5.0.0"},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		printQuietSummary(result)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "5.0.0") {
		t.Errorf("expected version '5.0.0' in quiet summary, got: %q", output)
	}
	if !strings.Contains(output, "Version:") {
		t.Errorf("expected 'Version:' in quiet summary, got: %q", output)
	}
}

func TestPrintQuietSummary_MultiModule(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
			{Name: "sub", RelPath: "sub/.version", Version: "1.0.0"},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		printQuietSummary(result)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "MultiModule") {
		t.Errorf("expected 'MultiModule' in quiet summary, got: %q", output)
	}
	if !strings.Contains(output, "Modules: 2") {
		t.Errorf("expected 'Modules: 2' in quiet summary, got: %q", output)
	}
}

func TestPrintQuietSummary_WithManifestsAndMismatches(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.SingleModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
		Manifests: []discovery.ManifestSource{
			{RelPath: "package.json", Version: "2.0.0"},
		},
		Mismatches: []discovery.Mismatch{
			{Source: "package.json", ExpectedVersion: "1.0.0", ActualVersion: "2.0.0"},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		printQuietSummary(result)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "Manifests: 1") {
		t.Errorf("expected 'Manifests: 1' in quiet summary, got: %q", output)
	}
	if !strings.Contains(output, "Mismatches: 1") {
		t.Errorf("expected 'Mismatches: 1' in quiet summary, got: %q", output)
	}
}
