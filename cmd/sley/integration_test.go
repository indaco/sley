//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Integration tests for the sley CLI.
// Run with: go test -tags=integration ./cmd/sley/...
//
// These tests build the actual binary and run it against real files,
// providing end-to-end verification of CLI behavior.

var binaryPath string

func TestMain(m *testing.M) {
	// Get the directory containing this test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}
	testDir := filepath.Dir(filename)

	// Build the binary once for all tests
	tmpDir, err := os.MkdirTemp("", "sley-integration-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}

	binaryPath = filepath.Join(tmpDir, "sley")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = testDir
	if output, err := cmd.CombinedOutput(); err != nil {
		panic("failed to build binary: " + string(output))
	}

	code := m.Run()

	// Cleanup
	os.RemoveAll(tmpDir)
	os.Exit(code)
}

// runSley executes the sley binary with the given arguments.
func runSley(t *testing.T, workdir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = workdir
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// writeVersionFile creates a .version file with the given content.
func writeVersionFile(t *testing.T, dir, version string) string {
	t.Helper()
	path := filepath.Join(dir, ".version")
	if err := os.WriteFile(path, []byte(version+"\n"), 0644); err != nil {
		t.Fatalf("failed to write version file: %v", err)
	}
	return path
}

// readVersionFile reads the content of the .version file.
func readVersionFile(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, ".version")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read version file: %v", err)
	}
	return strings.TrimSpace(string(data))
}

// TestIntegration_Init tests the init command.
func TestIntegration_Init(t *testing.T) {
	t.Run("creates default version file", func(t *testing.T) {
		dir := t.TempDir()

		output, err := runSley(t, dir, "init")
		if err != nil {
			t.Fatalf("init failed: %v, output: %s", err, output)
		}

		version := readVersionFile(t, dir)
		if version != "0.0.0" {
			t.Errorf("expected version 0.0.0, got %s", version)
		}
	})

	t.Run("creates version file at custom path", func(t *testing.T) {
		dir := t.TempDir()
		customPath := filepath.Join(dir, "custom", ".version")
		if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
			t.Fatal(err)
		}

		output, err := runSley(t, dir, "init", "--path", customPath)
		if err != nil {
			t.Fatalf("init failed: %v, output: %s", err, output)
		}

		data, err := os.ReadFile(customPath)
		if err != nil {
			t.Fatalf("failed to read custom version file: %v", err)
		}
		if strings.TrimSpace(string(data)) != "0.0.0" {
			t.Errorf("expected version 0.0.0, got %s", string(data))
		}
	})
}

// TestIntegration_Show tests the show command.
func TestIntegration_Show(t *testing.T) {
	t.Run("displays current version", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3")

		output, err := runSley(t, dir, "show")
		if err != nil {
			t.Fatalf("show failed: %v, output: %s", err, output)
		}

		if output != "1.2.3" {
			t.Errorf("expected output '1.2.3', got %q", output)
		}
	})

	t.Run("displays version with pre-release", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "2.0.0-beta.1")

		output, err := runSley(t, dir, "show")
		if err != nil {
			t.Fatalf("show failed: %v, output: %s", err, output)
		}

		if output != "2.0.0-beta.1" {
			t.Errorf("expected output '2.0.0-beta.1', got %q", output)
		}
	})

	t.Run("strict mode fails when file missing", func(t *testing.T) {
		dir := t.TempDir()

		_, err := runSley(t, dir, "show", "--strict")
		if err == nil {
			t.Error("expected error with --strict and missing file")
		}
	})

	t.Run("auto-init when file missing", func(t *testing.T) {
		dir := t.TempDir()

		output, err := runSley(t, dir, "show")
		if err != nil {
			t.Fatalf("show failed: %v, output: %s", err, output)
		}

		// Should auto-initialize and show the version
		if !strings.Contains(output, "0.0.0") {
			t.Errorf("expected output to contain '0.0.0', got %q", output)
		}
	})
}

// TestIntegration_Set tests the set command.
func TestIntegration_Set(t *testing.T) {
	t.Run("sets version", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.0.0")

		output, err := runSley(t, dir, "set", "2.0.0")
		if err != nil {
			t.Fatalf("set failed: %v, output: %s", err, output)
		}

		version := readVersionFile(t, dir)
		if version != "2.0.0" {
			t.Errorf("expected version 2.0.0, got %s", version)
		}
	})

	t.Run("sets version with pre-release", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.0.0")

		_, err := runSley(t, dir, "set", "2.0.0", "--pre", "alpha.1")
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "2.0.0-alpha.1" {
			t.Errorf("expected version 2.0.0-alpha.1, got %s", version)
		}
	})

	t.Run("sets version with build metadata", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.0.0")

		_, err := runSley(t, dir, "set", "2.0.0", "--meta", "build.123")
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "2.0.0+build.123" {
			t.Errorf("expected version 2.0.0+build.123, got %s", version)
		}
	})

	t.Run("sets version with pre-release and metadata", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.0.0")

		_, err := runSley(t, dir, "set", "2.0.0", "--pre", "rc.1", "--meta", "ci.456")
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "2.0.0-rc.1+ci.456" {
			t.Errorf("expected version 2.0.0-rc.1+ci.456, got %s", version)
		}
	})
}

// TestIntegration_BumpPatch tests bump patch command.
func TestIntegration_BumpPatch(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		expected string
	}{
		{"simple bump", "1.2.3", "1.2.4"},
		{"from zero", "0.0.0", "0.0.1"},
		{"large numbers", "10.20.30", "10.20.31"},
		{"clears pre-release", "1.2.3-alpha", "1.2.4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSley(t, dir, "bump", "patch")
			if err != nil {
				t.Fatalf("bump patch failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

// TestIntegration_BumpMinor tests bump minor command.
func TestIntegration_BumpMinor(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		expected string
	}{
		{"simple bump", "1.2.3", "1.3.0"},
		{"resets patch", "1.2.9", "1.3.0"},
		{"from zero", "0.0.5", "0.1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSley(t, dir, "bump", "minor")
			if err != nil {
				t.Fatalf("bump minor failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

// TestIntegration_BumpMajor tests bump major command.
func TestIntegration_BumpMajor(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		expected string
	}{
		{"simple bump", "1.2.3", "2.0.0"},
		{"resets minor and patch", "1.9.9", "2.0.0"},
		{"from zero", "0.5.5", "1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSley(t, dir, "bump", "major")
			if err != nil {
				t.Fatalf("bump major failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

// TestIntegration_BumpRelease tests bump release command.
func TestIntegration_BumpRelease(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		expected string
	}{
		{"removes pre-release", "1.2.3-alpha.1", "1.2.3"},
		{"removes pre-release and metadata", "1.2.3-beta+build.123", "1.2.3"},
		{"no-op for release version", "1.2.3", "1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSley(t, dir, "bump", "release")
			if err != nil {
				t.Fatalf("bump release failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

// TestIntegration_BumpAuto tests bump auto command.
func TestIntegration_BumpAuto(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		args     []string
		expected string
	}{
		{"promotes pre-release", "1.2.3-alpha.1", []string{"bump", "auto"}, "1.2.3"},
		{"bumps patch for release", "1.2.3", []string{"bump", "auto", "--no-infer"}, "1.2.4"},
		{"with label minor", "1.2.3", []string{"bump", "auto", "--label", "minor"}, "1.3.0"},
		{"with label major", "1.2.3", []string{"bump", "auto", "--label", "major"}, "2.0.0"},
		{"preserves metadata", "1.2.3-alpha+build.1", []string{"bump", "auto", "--preserve-meta"}, "1.2.3+build.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSley(t, dir, tt.args...)
			if err != nil {
				t.Fatalf("bump auto failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}

	t.Run("next alias works", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3-alpha")

		_, err := runSley(t, dir, "bump", "next")
		if err != nil {
			t.Fatalf("bump next (alias) failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "1.2.3" {
			t.Errorf("expected version 1.2.3, got %s", version)
		}
	})
}

// TestIntegration_Pre tests the pre command.
func TestIntegration_Pre(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		args     []string
		expected string
	}{
		{"sets alpha", "1.2.3", []string{"pre", "--label", "alpha"}, "1.2.4-alpha"},
		{"sets beta", "1.2.3", []string{"pre", "--label", "beta"}, "1.2.4-beta"},
		{"replaces pre-release", "1.2.3-alpha", []string{"pre", "--label", "beta"}, "1.2.3-beta"},
		{"increments alpha", "1.2.3", []string{"pre", "--label", "alpha", "--inc"}, "1.2.3-alpha.1"},
		{"increments existing", "1.2.3-alpha.1", []string{"pre", "--label", "alpha", "--inc"}, "1.2.3-alpha.2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSley(t, dir, tt.args...)
			if err != nil {
				t.Fatalf("pre failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

// TestIntegration_Validate tests the validate command.
func TestIntegration_Validate(t *testing.T) {
	t.Run("valid version passes", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3")

		output, err := runSley(t, dir, "validate")
		if err != nil {
			t.Fatalf("validate failed: %v, output: %s", err, output)
		}

		if !strings.Contains(output, "Valid") {
			t.Errorf("expected output to contain 'Valid', got %q", output)
		}
	})

	t.Run("invalid version fails", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "not-a-version")

		_, err := runSley(t, dir, "validate")
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})

	t.Run("missing file fails", func(t *testing.T) {
		dir := t.TempDir()

		_, err := runSley(t, dir, "validate", "--strict")
		if err == nil {
			t.Error("expected error for missing file with --strict")
		}
	})
}

// TestIntegration_BumpWithFlags tests bump commands with various flags.
func TestIntegration_BumpWithFlags(t *testing.T) {
	t.Run("bump patch with pre-release", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3")

		_, err := runSley(t, dir, "bump", "patch", "--pre", "beta.1")
		if err != nil {
			t.Fatalf("bump failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "1.2.4-beta.1" {
			t.Errorf("expected version 1.2.4-beta.1, got %s", version)
		}
	})

	t.Run("bump minor with metadata", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3")

		_, err := runSley(t, dir, "bump", "minor", "--meta", "ci.456")
		if err != nil {
			t.Fatalf("bump failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "1.3.0+ci.456" {
			t.Errorf("expected version 1.3.0+ci.456, got %s", version)
		}
	})

	t.Run("bump preserves metadata", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3+build.789")

		_, err := runSley(t, dir, "bump", "patch", "--preserve-meta")
		if err != nil {
			t.Fatalf("bump failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "1.2.4+build.789" {
			t.Errorf("expected version 1.2.4+build.789, got %s", version)
		}
	})
}

// TestIntegration_CustomPath tests --path flag across commands.
func TestIntegration_CustomPath(t *testing.T) {
	dir := t.TempDir()
	customPath := filepath.Join(dir, "versions", "app.version")
	if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Init at custom path
	_, err := runSley(t, dir, "init", "--path", customPath)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Show from custom path
	output, err := runSley(t, dir, "show", "--path", customPath)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if output != "0.0.0" {
		t.Errorf("expected 0.0.0, got %s", output)
	}

	// Bump at custom path
	_, err = runSley(t, dir, "bump", "minor", "--path", customPath)
	if err != nil {
		t.Fatalf("bump failed: %v", err)
	}

	// Verify bump
	output, err = runSley(t, dir, "show", "--path", customPath)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if output != "0.2.0" {
		t.Errorf("expected 0.2.0, got %s", output)
	}
}

// TestIntegration_VersionFlag tests --version flag.
func TestIntegration_VersionFlag(t *testing.T) {
	dir := t.TempDir()

	output, err := runSley(t, dir, "--version")
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}

	if !strings.HasPrefix(output, "sley version v") {
		t.Errorf("expected version output, got %q", output)
	}
}

// TestIntegration_HelpFlag tests --help flag.
func TestIntegration_HelpFlag(t *testing.T) {
	dir := t.TempDir()

	output, err := runSley(t, dir, "--help")
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}

	requiredStrings := []string{"sley", "show", "set", "bump", "pre", "validate", "init"}
	for _, s := range requiredStrings {
		if !strings.Contains(output, s) {
			t.Errorf("expected help to contain %q", s)
		}
	}
}

// ============================================================================
// Multi-Module Integration Tests
// ============================================================================

// setupMultiModuleWorkspace creates a workspace with multiple modules.
func setupMultiModuleWorkspace(t *testing.T, dir string, modules map[string]string) {
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

// readModuleVersion reads the version of a specific module.
func readModuleVersion(t *testing.T, dir, modulePath string) string {
	t.Helper()
	versionFile := filepath.Join(dir, modulePath, ".version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("failed to read version file %s: %v", versionFile, err)
	}
	return strings.TrimSpace(string(data))
}

// TestIntegration_ModulesList tests the modules list command.
func TestIntegration_ModulesList(t *testing.T) {
	t.Run("lists discovered modules", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"services/api": "1.0.0",
			"services/web": "2.0.0",
			"packages/lib": "0.5.0",
		})

		output, err := runSley(t, dir, "modules", "list")
		if err != nil {
			t.Fatalf("modules list failed: %v, output: %s", err, output)
		}

		// Should list all modules
		if !strings.Contains(output, "api") {
			t.Error("expected output to contain 'api'")
		}
		if !strings.Contains(output, "web") {
			t.Error("expected output to contain 'web'")
		}
		if !strings.Contains(output, "lib") {
			t.Error("expected output to contain 'lib'")
		}
	})

	t.Run("lists modules in JSON format", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"module-a": "1.0.0",
			"module-b": "2.0.0",
		})

		output, err := runSley(t, dir, "modules", "list", "--format", "json")
		if err != nil {
			t.Fatalf("modules list --format json failed: %v, output: %s", err, output)
		}

		// JSON output is an array of module objects
		if !strings.Contains(output, `"name"`) {
			t.Error("expected JSON output with 'name' key")
		}
		if !strings.Contains(output, `"version"`) {
			t.Error("expected JSON output with 'version' key")
		}
		if !strings.Contains(output, "module-a") {
			t.Error("expected JSON output to contain 'module-a'")
		}
	})

	t.Run("handles empty workspace", func(t *testing.T) {
		dir := t.TempDir()

		output, err := runSley(t, dir, "modules", "list")
		// Should not error but show no modules found
		if err != nil {
			// Check if it's a "no modules found" type message
			if !strings.Contains(output, "No modules") && !strings.Contains(output, "no modules") {
				t.Fatalf("modules list failed unexpectedly: %v, output: %s", err, output)
			}
		}
	})
}

// TestIntegration_ModulesDiscover tests the modules discover command.
func TestIntegration_ModulesDiscover(t *testing.T) {
	t.Run("discovers modules in workspace", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"app":      "1.0.0",
			"lib/core": "0.1.0",
		})

		output, err := runSley(t, dir, "modules", "discover")
		if err != nil {
			t.Fatalf("modules discover failed: %v, output: %s", err, output)
		}

		// Should show discovery information
		if !strings.Contains(output, "app") && !strings.Contains(output, "core") {
			t.Error("expected output to contain discovered module names")
		}
	})
}

// TestIntegration_MultiModule_ShowAll tests show command with --all flag.
func TestIntegration_MultiModule_ShowAll(t *testing.T) {
	t.Run("shows all module versions", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"api": "1.0.0",
			"web": "2.0.0",
		})

		output, err := runSley(t, dir, "show", "--all", "--non-interactive")
		if err != nil {
			t.Fatalf("show --all failed: %v, output: %s", err, output)
		}

		// Should contain both module versions
		if !strings.Contains(output, "1.0.0") {
			t.Error("expected output to contain '1.0.0'")
		}
		if !strings.Contains(output, "2.0.0") {
			t.Error("expected output to contain '2.0.0'")
		}
	})

	t.Run("shows versions in JSON format", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"api": "1.0.0",
			"web": "2.0.0",
		})

		output, err := runSley(t, dir, "show", "--all", "--non-interactive", "--format", "json")
		if err != nil {
			t.Fatalf("show --all --format json failed: %v, output: %s", err, output)
		}

		// Should be valid JSON with results
		if !strings.Contains(output, `"results"`) {
			t.Error("expected JSON output to contain 'results'")
		}
		if !strings.Contains(output, `"success": true`) {
			t.Error("expected JSON output to show success")
		}
	})
}

// TestIntegration_MultiModule_BumpAll tests bump command with --all flag.
func TestIntegration_MultiModule_BumpAll(t *testing.T) {
	t.Run("bumps all module versions", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"api": "1.0.0",
			"web": "2.0.0",
		})

		output, err := runSley(t, dir, "bump", "patch", "--all", "--non-interactive")
		if err != nil {
			t.Fatalf("bump patch --all failed: %v, output: %s", err, output)
		}

		// Verify both modules were bumped
		apiVersion := readModuleVersion(t, dir, "api")
		if apiVersion != "1.0.1" {
			t.Errorf("expected api version '1.0.1', got %q", apiVersion)
		}

		webVersion := readModuleVersion(t, dir, "web")
		if webVersion != "2.0.1" {
			t.Errorf("expected web version '2.0.1', got %q", webVersion)
		}
	})

	t.Run("bumps minor across all modules", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"api": "1.0.0",
			"web": "2.0.0",
		})

		_, err := runSley(t, dir, "bump", "minor", "--all", "--non-interactive")
		if err != nil {
			t.Fatalf("bump minor --all failed: %v", err)
		}

		apiVersion := readModuleVersion(t, dir, "api")
		if apiVersion != "1.1.0" {
			t.Errorf("expected api version '1.1.0', got %q", apiVersion)
		}

		webVersion := readModuleVersion(t, dir, "web")
		if webVersion != "2.1.0" {
			t.Errorf("expected web version '2.1.0', got %q", webVersion)
		}
	})
}

// TestIntegration_MultiModule_SpecificModule tests --module flag.
func TestIntegration_MultiModule_SpecificModule(t *testing.T) {
	t.Run("bumps only specified module", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"api": "1.0.0",
			"web": "2.0.0",
		})

		_, err := runSley(t, dir, "bump", "patch", "--module", "api", "--non-interactive")
		if err != nil {
			t.Fatalf("bump patch --module api failed: %v", err)
		}

		// Only api should be bumped
		apiVersion := readModuleVersion(t, dir, "api")
		if apiVersion != "1.0.1" {
			t.Errorf("expected api version '1.0.1', got %q", apiVersion)
		}

		// web should remain unchanged
		webVersion := readModuleVersion(t, dir, "web")
		if webVersion != "2.0.0" {
			t.Errorf("expected web version '2.0.0', got %q", webVersion)
		}
	})

	t.Run("shows only specified module", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"api": "1.0.0",
			"web": "2.0.0",
		})

		output, err := runSley(t, dir, "show", "--module", "api", "--non-interactive")
		if err != nil {
			t.Fatalf("show --module api failed: %v, output: %s", err, output)
		}

		// Should show only api version
		if !strings.Contains(output, "1.0.0") {
			t.Error("expected output to contain '1.0.0'")
		}
	})
}

// TestIntegration_MultiModule_SetAll tests set command with --all flag.
func TestIntegration_MultiModule_SetAll(t *testing.T) {
	t.Run("sets all module versions", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"api": "1.0.0",
			"web": "2.0.0",
		})

		output, err := runSley(t, dir, "set", "3.0.0", "--all", "--non-interactive")
		if err != nil {
			t.Fatalf("set 3.0.0 --all failed: %v, output: %s", err, output)
		}

		// Both modules should have version 3.0.0
		apiVersion := readModuleVersion(t, dir, "api")
		if apiVersion != "3.0.0" {
			t.Errorf("expected api version '3.0.0', got %q", apiVersion)
		}

		webVersion := readModuleVersion(t, dir, "web")
		if webVersion != "3.0.0" {
			t.Errorf("expected web version '3.0.0', got %q", webVersion)
		}
	})
}

// TestIntegration_MultiModule_YesFlag tests --yes flag for auto-selection.
func TestIntegration_MultiModule_YesFlag(t *testing.T) {
	t.Run("auto-selects all modules with --yes", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"api": "1.0.0",
			"web": "2.0.0",
		})

		output, err := runSley(t, dir, "bump", "patch", "--yes")
		if err != nil {
			t.Fatalf("bump patch --yes failed: %v, output: %s", err, output)
		}

		// Both modules should be bumped
		apiVersion := readModuleVersion(t, dir, "api")
		if apiVersion != "1.0.1" {
			t.Errorf("expected api version '1.0.1', got %q", apiVersion)
		}

		webVersion := readModuleVersion(t, dir, "web")
		if webVersion != "2.0.1" {
			t.Errorf("expected web version '2.0.1', got %q", webVersion)
		}
	})
}

// TestIntegration_MultiModule_QuietFlag tests --quiet flag.
func TestIntegration_MultiModule_QuietFlag(t *testing.T) {
	t.Run("shows only summary with --quiet", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"api": "1.0.0",
			"web": "2.0.0",
		})

		output, err := runSley(t, dir, "bump", "patch", "--all", "--non-interactive", "--quiet")
		if err != nil {
			t.Fatalf("bump patch --all --quiet failed: %v, output: %s", err, output)
		}

		// Should show summary, not individual module results
		if !strings.Contains(output, "Success") && !strings.Contains(output, "module") {
			t.Error("expected summary output")
		}
	})
}

// TestIntegration_MultiModule_NestedDirectories tests modules in nested directories.
func TestIntegration_MultiModule_NestedDirectories(t *testing.T) {
	t.Run("discovers modules in nested directories", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"services/api/v1":      "1.0.0",
			"services/api/v2":      "2.0.0",
			"packages/shared/core": "0.1.0",
			"packages/shared/util": "0.2.0",
		})

		output, err := runSley(t, dir, "modules", "list")
		if err != nil {
			t.Fatalf("modules list failed: %v, output: %s", err, output)
		}

		// Should discover all nested modules
		expectedModules := []string{"v1", "v2", "core", "util"}
		for _, mod := range expectedModules {
			if !strings.Contains(output, mod) {
				t.Errorf("expected output to contain '%s'", mod)
			}
		}
	})
}

// TestIntegration_MultiModule_SingleModuleFallback tests fallback to single module mode.
func TestIntegration_MultiModule_SingleModuleFallback(t *testing.T) {
	t.Run("uses single module mode when only root .version exists", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.0.0")

		output, err := runSley(t, dir, "show")
		if err != nil {
			t.Fatalf("show failed: %v, output: %s", err, output)
		}

		if output != "1.0.0" {
			t.Errorf("expected '1.0.0', got %q", output)
		}
	})

	t.Run("prefers --path over multi-module discovery", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"api": "1.0.0",
			"web": "2.0.0",
		})

		// Also create a root .version file
		writeVersionFile(t, dir, "0.0.1")

		output, err := runSley(t, dir, "show", "--path", filepath.Join(dir, ".version"))
		if err != nil {
			t.Fatalf("show --path failed: %v, output: %s", err, output)
		}

		if output != "0.0.1" {
			t.Errorf("expected '0.0.1' from explicit path, got %q", output)
		}
	})
}

// TestIntegration_MultiModule_ErrorHandling tests error scenarios.
func TestIntegration_MultiModule_ErrorHandling(t *testing.T) {
	t.Run("fails when module not found", func(t *testing.T) {
		dir := t.TempDir()
		setupMultiModuleWorkspace(t, dir, map[string]string{
			"api": "1.0.0",
			"web": "2.0.0",
		})

		_, err := runSley(t, dir, "bump", "patch", "--module", "nonexistent")
		if err == nil {
			t.Error("expected error for non-existent module")
		}
	})

	t.Run("handles no modules gracefully", func(t *testing.T) {
		dir := t.TempDir()
		// Empty directory with no .version files

		_, err := runSley(t, dir, "bump", "patch", "--all")
		if err == nil {
			t.Error("expected error when no modules found")
		}
	})
}

// TestIntegration_Doctor tests the doctor command.
func TestIntegration_Doctor(t *testing.T) {
	t.Run("runs doctor check successfully", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.0.0")

		output, err := runSley(t, dir, "doctor")
		if err != nil {
			t.Fatalf("doctor failed: %v, output: %s", err, output)
		}

		// Doctor should output some diagnostic info
		if len(output) == 0 {
			t.Error("expected doctor to produce output")
		}
	})
}

// TestIntegration_NoColorFlag tests --no-color flag.
func TestIntegration_NoColorFlag(t *testing.T) {
	t.Run("disables colored output", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.0.0")

		output, err := runSley(t, dir, "show", "--no-color")
		if err != nil {
			t.Fatalf("show --no-color failed: %v, output: %s", err, output)
		}

		// Output should not contain ANSI escape codes
		if strings.Contains(output, "\033[") {
			t.Error("expected no ANSI escape codes with --no-color")
		}
	})
}
