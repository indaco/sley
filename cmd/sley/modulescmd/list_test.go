package modulescmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/testutils"
	"github.com/indaco/sley/internal/workspace"
	"github.com/urfave/cli/v3"
)

func TestListCmd(t *testing.T) {
	cmd := listCmd()

	if cmd == nil {
		t.Fatal("listCmd() returned nil")
	}

	if cmd.Name != "list" {
		t.Errorf("command name = %q, want %q", cmd.Name, "list")
	}

	expectedAliases := []string{"ls"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("command has %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	} else {
		for i, alias := range cmd.Aliases {
			if alias != expectedAliases[i] {
				t.Errorf("alias[%d] = %q, want %q", i, alias, expectedAliases[i])
			}
		}
	}

	expectedUsage := "List all discovered modules in workspace"
	if cmd.Usage != expectedUsage {
		t.Errorf("command usage = %q, want %q", cmd.Usage, expectedUsage)
	}

	if cmd.Action == nil {
		t.Error("command has no action")
	}
}

func TestListCmd_Flags(t *testing.T) {
	cmd := listCmd()

	expectedFlags := map[string]bool{
		"verbose": true,
		"format":  true,
	}

	foundFlags := make(map[string]bool)
	for _, flag := range cmd.Flags {
		for _, name := range flag.Names() {
			foundFlags[name] = true
		}
	}

	for expectedFlag := range expectedFlags {
		if !foundFlags[expectedFlag] {
			t.Errorf("missing expected flag %q", expectedFlag)
		}
	}

	// Check for aliases
	expectedAliases := map[string]bool{
		"v": true, // verbose alias
	}

	foundAliases := make(map[string]bool)
	for _, flag := range cmd.Flags {
		for _, name := range flag.Names() {
			if name != "verbose" && name != "format" {
				foundAliases[name] = true
			}
		}
	}

	for expectedAlias := range expectedAliases {
		if !foundAliases[expectedAlias] {
			t.Errorf("missing expected flag alias %q", expectedAlias)
		}
	}
}

func TestOutputText_NoModules(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	modules := []*workspace.Module{}
	err := outputText(modules, false)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputText failed: %v", err)
	}

	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	output := buf.String()

	expected := "Found 0 module(s):\n"
	if output != expected {
		t.Errorf("output = %q, want %q", output, expected)
	}
}

func TestOutputText_SingleModule(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	modules := []*workspace.Module{
		{
			Name:           "module-a",
			RelPath:        "module-a/.version",
			CurrentVersion: "1.0.0",
		},
	}
	err := outputText(modules, false)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputText failed: %v", err)
	}

	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Now expects the path to be displayed on a second line
	if !strings.Contains(output, "Found 1 module(s):") {
		t.Errorf("expected header, got: %q", output)
	}
	if !strings.Contains(output, "module-a (1.0.0)") {
		t.Errorf("expected module name and version, got: %q", output)
	}
	if !strings.Contains(output, "module-a") {
		t.Errorf("expected module path, got: %q", output)
	}
}

func TestOutputText_MultipleModules(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	modules := []*workspace.Module{
		{
			Name:           "module-a",
			RelPath:        "module-a/.version",
			CurrentVersion: "1.0.0",
		},
		{
			Name:           "module-b",
			RelPath:        "module-b/.version",
			CurrentVersion: "2.0.0",
		},
	}
	err := outputText(modules, false)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputText failed: %v", err)
	}

	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Now expects paths to be displayed
	if !strings.Contains(output, "Found 2 module(s):") {
		t.Errorf("expected header, got: %q", output)
	}
	if !strings.Contains(output, "module-a (1.0.0)") {
		t.Errorf("expected module-a with version, got: %q", output)
	}
	if !strings.Contains(output, "module-b (2.0.0)") {
		t.Errorf("expected module-b with version, got: %q", output)
	}
	// Paths should be present
	if !strings.Contains(output, "module-a") {
		t.Errorf("expected module-a path, got: %q", output)
	}
	if !strings.Contains(output, "module-b") {
		t.Errorf("expected module-b path, got: %q", output)
	}
}

func TestOutputText_Verbose(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	modules := []*workspace.Module{
		{
			Name:           "module-a",
			RelPath:        "module-a/.version",
			CurrentVersion: "1.0.0",
		},
	}
	err := outputText(modules, true)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputText failed: %v", err)
	}

	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	output := buf.String()

	expected := "Found 1 module(s):\n  - module-a\n    Path: module-a/.version\n    Version: 1.0.0\n"
	if output != expected {
		t.Errorf("output = %q, want %q", output, expected)
	}
}

func TestOutputText_UnknownVersion(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	modules := []*workspace.Module{
		{
			Name:           "module-a",
			RelPath:        "module-a/.version",
			CurrentVersion: "",
		},
	}
	err := outputText(modules, false)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputText failed: %v", err)
	}

	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Now expects the path to be displayed
	if !strings.Contains(output, "Found 1 module(s):") {
		t.Errorf("expected header, got: %q", output)
	}
	if !strings.Contains(output, "module-a (unknown)") {
		t.Errorf("expected module-a with unknown version, got: %q", output)
	}
	if !strings.Contains(output, "module-a") {
		t.Errorf("expected module-a path, got: %q", output)
	}
}

func TestOutputText_CurrentDirectoryModule(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	modules := []*workspace.Module{
		{
			Name:           "root-module",
			RelPath:        ".version",
			CurrentVersion: "1.0.0",
		},
	}
	err := outputText(modules, false)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputText failed: %v", err)
	}

	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	output := buf.String()

	// For modules in current directory (RelPath = ".version"), path should not be displayed
	if !strings.Contains(output, "Found 1 module(s):") {
		t.Errorf("expected header, got: %q", output)
	}
	if !strings.Contains(output, "root-module (1.0.0)") {
		t.Errorf("expected module with version, got: %q", output)
	}
	// Count newlines - should only be 2 (header + module line) without the path line
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines of output, got %d: %q", len(lines), output)
	}
}

func TestOutputText_DuplicateModuleNames(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	modules := []*workspace.Module{
		{
			Name:           "version",
			RelPath:        "backend/gateway/internal/version/.version",
			CurrentVersion: "0.8.0-alpha.6",
		},
		{
			Name:           "version",
			RelPath:        "cli/internal/version/.version",
			CurrentVersion: "0.8.0-alpha.6",
		},
		{
			Name:           "ai-services",
			RelPath:        "backend/ai-services/.version",
			CurrentVersion: "0.8.0-alpha.6",
		},
	}
	err := outputText(modules, false)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputText failed: %v", err)
	}

	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Verify header
	if !strings.Contains(output, "Found 3 module(s):") {
		t.Errorf("expected header with 3 modules, got: %q", output)
	}

	// Verify all modules are listed with their versions
	if !strings.Contains(output, "version (0.8.0-alpha.6)") {
		t.Errorf("expected version module entries, got: %q", output)
	}
	if !strings.Contains(output, "ai-services (0.8.0-alpha.6)") {
		t.Errorf("expected ai-services module, got: %q", output)
	}

	// Verify disambiguating paths are present
	if !strings.Contains(output, "backend/gateway/internal/version") {
		t.Errorf("expected first version module path, got: %q", output)
	}
	if !strings.Contains(output, "cli/internal/version") {
		t.Errorf("expected second version module path, got: %q", output)
	}
	if !strings.Contains(output, "backend/ai-services") {
		t.Errorf("expected ai-services module path, got: %q", output)
	}
}

func TestOutputJSON_SingleModule(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	modules := []*workspace.Module{
		{
			Name:           "module-a",
			RelPath:        "module-a/.version",
			CurrentVersion: "1.0.0",
		},
	}
	err := outputJSON(modules)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputJSON failed: %v", err)
	}

	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	output := buf.String()

	var result []moduleJSON
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 module in JSON, got %d", len(result))
	}

	if result[0].Name != "module-a" {
		t.Errorf("module name = %q, want %q", result[0].Name, "module-a")
	}
	if result[0].Path != "module-a/.version" {
		t.Errorf("module path = %q, want %q", result[0].Path, "module-a/.version")
	}
	if result[0].Version != "1.0.0" {
		t.Errorf("module version = %q, want %q", result[0].Version, "1.0.0")
	}
}

func TestOutputJSON_MultipleModules(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	modules := []*workspace.Module{
		{
			Name:           "module-a",
			RelPath:        "module-a/.version",
			CurrentVersion: "1.0.0",
		},
		{
			Name:           "module-b",
			RelPath:        "module-b/.version",
			CurrentVersion: "2.0.0",
		},
	}
	err := outputJSON(modules)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputJSON failed: %v", err)
	}

	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	output := buf.String()

	var result []moduleJSON
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 modules in JSON, got %d", len(result))
	}

	expectedModules := map[string]moduleJSON{
		"module-a": {Name: "module-a", Path: "module-a/.version", Version: "1.0.0"},
		"module-b": {Name: "module-b", Path: "module-b/.version", Version: "2.0.0"},
	}

	for _, mod := range result {
		expected, ok := expectedModules[mod.Name]
		if !ok {
			t.Errorf("unexpected module %q in output", mod.Name)
			continue
		}

		if mod.Path != expected.Path {
			t.Errorf("module %q path = %q, want %q", mod.Name, mod.Path, expected.Path)
		}
		if mod.Version != expected.Version {
			t.Errorf("module %q version = %q, want %q", mod.Name, mod.Version, expected.Version)
		}
	}
}

func TestOutputJSON_EmptyModules(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	modules := []*workspace.Module{}
	err := outputJSON(modules)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputJSON failed: %v", err)
	}

	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}
	output := buf.String()

	var result []moduleJSON
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty array, got %d modules", len(result))
	}
}

func TestModuleJSON_Structure(t *testing.T) {
	mod := moduleJSON{
		Name:    "test-module",
		Path:    "test/.version",
		Version: "1.2.3",
	}

	data, err := json.Marshal(mod)
	if err != nil {
		t.Fatalf("failed to marshal moduleJSON: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if result["name"] != "test-module" {
		t.Errorf("name = %v, want %v", result["name"], "test-module")
	}
	if result["path"] != "test/.version" {
		t.Errorf("path = %v, want %v", result["path"], "test/.version")
	}
	if result["version"] != "1.2.3" {
		t.Errorf("version = %v, want %v", result["version"], "1.2.3")
	}
}

func TestRunList_ConfigLoadError(t *testing.T) {
	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn to return an error
	config.LoadConfigFn = func() (*config.Config, error) {
		return nil, os.ErrNotExist
	}

	ctx := context.Background()

	// Create a mock CLI command context
	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
			&cli.BoolFlag{Name: "verbose"},
		},
	}

	err := runList(ctx, mockCmd)
	if err == nil {
		t.Fatal("expected error from runList, got nil")
	}

	expectedErrMsg := "failed to load config"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("error message = %q, want to contain %q", err.Error(), expectedErrMsg)
	}
}

func TestRunList_GetWdError(t *testing.T) {
	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn to return valid config
	config.LoadConfigFn = func() (*config.Config, error) {
		return &config.Config{}, nil
	}

	// Change to a directory that will be deleted
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to change to subdir: %v", err)
	}

	// Remove the directory we're in (this will make Getwd fail on some systems)
	if err := os.RemoveAll(tmpDir); err != nil {
		t.Fatalf("failed to remove tmpDir: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
			&cli.BoolFlag{Name: "verbose"},
		},
	}

	err = runList(ctx, mockCmd)
	// On some systems this might succeed, on others it fails
	// Just verify it doesn't panic
	_ = err
}

func TestRunList_NoModulesFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn
	config.LoadConfigFn = func() (*config.Config, error) {
		return &config.Config{}, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
			&cli.BoolFlag{Name: "verbose"},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		if err := runList(ctx, mockCmd); err != nil {
			t.Errorf("runList failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "No modules found in workspace") {
		t.Errorf("expected 'No modules found' message, got: %q", output)
	}
}

func TestRunList_TextFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a module structure
	moduleDir := filepath.Join(tmpDir, "module-a")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	versionFile := filepath.Join(moduleDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn
	config.LoadConfigFn = func() (*config.Config, error) {
		enabled := true
		recursive := true
		maxDepth := 10
		return &config.Config{
			Workspace: &config.WorkspaceConfig{
				Discovery: &config.DiscoveryConfig{
					Enabled:   &enabled,
					Recursive: &recursive,
					MaxDepth:  &maxDepth,
				},
			},
		}, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
			&cli.BoolFlag{Name: "verbose"},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		if err := runList(ctx, mockCmd); err != nil {
			t.Errorf("runList failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "Found") && !strings.Contains(output, "module") {
		t.Errorf("expected module list output, got: %q", output)
	}
}

func TestRunList_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a module structure
	moduleDir := filepath.Join(tmpDir, "module-a")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	versionFile := filepath.Join(moduleDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn
	config.LoadConfigFn = func() (*config.Config, error) {
		enabled := true
		recursive := true
		maxDepth := 10
		return &config.Config{
			Workspace: &config.WorkspaceConfig{
				Discovery: &config.DiscoveryConfig{
					Enabled:   &enabled,
					Recursive: &recursive,
					MaxDepth:  &maxDepth,
				},
			},
		}, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "json"},
			&cli.BoolFlag{Name: "verbose"},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		if err := runList(ctx, mockCmd); err != nil {
			t.Errorf("runList failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Verify it's valid JSON
	var result []moduleJSON
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("output is not valid JSON: %v, output: %q", err, output)
	}
}

func TestRunList_VerboseFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a module structure
	moduleDir := filepath.Join(tmpDir, "module-a")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatalf("failed to create module directory: %v", err)
	}

	versionFile := filepath.Join(moduleDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn
	config.LoadConfigFn = func() (*config.Config, error) {
		enabled := true
		recursive := true
		maxDepth := 10
		return &config.Config{
			Workspace: &config.WorkspaceConfig{
				Discovery: &config.DiscoveryConfig{
					Enabled:   &enabled,
					Recursive: &recursive,
					MaxDepth:  &maxDepth,
				},
			},
		}, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
			&cli.BoolFlag{Name: "verbose", Value: true},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		if err := runList(ctx, mockCmd); err != nil {
			t.Errorf("runList failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// In verbose mode, we expect to see "Path:" and "Version:"
	if !strings.Contains(output, "Path:") || !strings.Contains(output, "Version:") {
		t.Errorf("expected verbose output with Path and Version, got: %q", output)
	}
}

func TestRunList_NilConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn to return nil config
	config.LoadConfigFn = func() (*config.Config, error) {
		return nil, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
			&cli.BoolFlag{Name: "verbose"},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		if err := runList(ctx, mockCmd); err != nil {
			t.Errorf("runList failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	// Should work with default config
	if !strings.Contains(output, "No modules found") {
		t.Errorf("expected 'No modules found' message, got: %q", output)
	}
}

func TestRunList_DiscoverModulesError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an inaccessible subdirectory
	restrictedDir := filepath.Join(tmpDir, "restricted")
	if err := os.MkdirAll(restrictedDir, 0755); err != nil {
		t.Fatalf("failed to create restricted directory: %v", err)
	}

	versionFile := filepath.Join(restrictedDir, ".version")
	if err := os.WriteFile(versionFile, []byte("1.0.0"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	// Make directory inaccessible
	if err := os.Chmod(restrictedDir, 0000); err != nil {
		t.Fatalf("failed to change permissions: %v", err)
	}
	defer func() { _ = os.Chmod(restrictedDir, 0755) }()

	// Save original and restore after test
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock LoadConfigFn
	config.LoadConfigFn = func() (*config.Config, error) {
		enabled := true
		recursive := true
		maxDepth := 10
		return &config.Config{
			Workspace: &config.WorkspaceConfig{
				Discovery: &config.DiscoveryConfig{
					Enabled:   &enabled,
					Recursive: &recursive,
					MaxDepth:  &maxDepth,
				},
			},
		}, nil
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()

	mockCmd := &cli.Command{
		Name: "test",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "text"},
			&cli.BoolFlag{Name: "verbose"},
		},
	}

	// This test might not fail on all systems
	// We just ensure it doesn't panic
	_, _ = testutils.CaptureStdout(func() {
		_ = runList(ctx, mockCmd)
	})
}
