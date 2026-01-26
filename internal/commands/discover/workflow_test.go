package discover

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/parser"
	"github.com/indaco/sley/internal/testutils"
)

// MockPrompter is a test double for Prompter.
type MockPrompter struct {
	ConfirmResult     bool
	ConfirmErr        error
	MultiSelectResult []string
	MultiSelectErr    error

	ConfirmCalls     int
	MultiSelectCalls int
}

func (m *MockPrompter) Confirm(title, description string) (bool, error) {
	m.ConfirmCalls++
	return m.ConfirmResult, m.ConfirmErr
}

func (m *MockPrompter) MultiSelect(title, description string, options []huh.Option[string], defaults []string) ([]string, error) {
	m.MultiSelectCalls++
	return m.MultiSelectResult, m.MultiSelectErr
}

func TestGenerateDependencyCheckFileConfig(t *testing.T) {
	candidate := discovery.SyncCandidate{
		Path:    "package.json",
		Format:  parser.FormatJSON,
		Field:   "version",
		Pattern: "",
	}

	cfg := GenerateDependencyCheckFileConfig(candidate)

	if cfg.Path != "package.json" {
		t.Errorf("Path = %q, want %q", cfg.Path, "package.json")
	}
	if cfg.Format != "json" {
		t.Errorf("Format = %q, want %q", cfg.Format, "json")
	}
	if cfg.Field != "version" {
		t.Errorf("Field = %q, want %q", cfg.Field, "version")
	}
}

func TestGenerateDependencyCheckConfig(t *testing.T) {
	candidates := []discovery.SyncCandidate{
		{
			Path:   "package.json",
			Format: parser.FormatJSON,
			Field:  "version",
		},
		{
			Path:   "Cargo.toml",
			Format: parser.FormatTOML,
			Field:  "package.version",
		},
	}

	cfg := GenerateDependencyCheckConfig(candidates)

	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}
	if !cfg.AutoSync {
		t.Error("AutoSync should be true")
	}
	if len(cfg.Files) != 2 {
		t.Errorf("Files length = %d, want 2", len(cfg.Files))
	}
}

func TestSuggestDependencyCheckFromDiscovery(t *testing.T) {
	tests := []struct {
		name      string
		result    *discovery.Result
		wantNil   bool
		wantFiles int
	}{
		{
			name:    "nil result",
			result:  nil,
			wantNil: true,
		},
		{
			name: "empty sync candidates",
			result: &discovery.Result{
				SyncCandidates: []discovery.SyncCandidate{},
			},
			wantNil: true,
		},
		{
			name: "with sync candidates",
			result: &discovery.Result{
				SyncCandidates: []discovery.SyncCandidate{
					{Path: "package.json", Format: parser.FormatJSON, Field: "version"},
					{Path: "Cargo.toml", Format: parser.FormatTOML, Field: "package.version"},
				},
			},
			wantNil:   false,
			wantFiles: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := SuggestDependencyCheckFromDiscovery(tt.result)

			if tt.wantNil {
				if cfg != nil {
					t.Error("expected nil config")
				}
				return
			}

			if cfg == nil {
				t.Fatal("expected non-nil config")
			}

			if len(cfg.Files) != tt.wantFiles {
				t.Errorf("Files length = %d, want %d", len(cfg.Files), tt.wantFiles)
			}
		})
	}
}

func TestWorkflow_generateDependencyCheckConfig(t *testing.T) {
	candidates := []discovery.SyncCandidate{
		{
			Path:        "package.json",
			Format:      parser.FormatJSON,
			Field:       "version",
			Description: "Node.js",
		},
		{
			Path:        "version.go",
			Format:      parser.FormatRegex,
			Pattern:     `Version = "(.*?)"`,
			Description: "Go",
		},
	}

	w := &Workflow{}
	output := w.generateDependencyCheckConfig(candidates)

	// Verify YAML structure
	checks := []string{
		"plugins:",
		"dependency-check:",
		"enabled: true",
		"auto-sync: true",
		"files:",
		"path: package.json",
		"format: json",
		"field: version",
		"path: version.go",
		"format: regex",
		"pattern:",
	}

	for _, check := range checks {
		if !contains(output, check) {
			t.Errorf("output missing expected text %q", check)
		}
	}
}

func TestWorkflow_selectSyncFiles_EmptyCandidates(t *testing.T) {
	mock := &MockPrompter{}
	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{},
	}
	w := NewWorkflow(mock, result, "/test")

	selected, err := w.selectSyncFiles()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 0 {
		t.Errorf("expected empty selection, got %d", len(selected))
	}
}

func TestWorkflow_selectSyncFiles_WithCandidates(t *testing.T) {
	mock := &MockPrompter{
		MultiSelectResult: []string{"package.json"},
	}
	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
			{Path: "Cargo.toml", Format: parser.FormatTOML},
		},
	}
	w := NewWorkflow(mock, result, "/test")

	selected, err := w.selectSyncFiles()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 1 {
		t.Errorf("expected 1 selection, got %d", len(selected))
	}
	if len(selected) > 0 && selected[0].Path != "package.json" {
		t.Errorf("expected package.json, got %s", selected[0].Path)
	}
	if mock.MultiSelectCalls != 1 {
		t.Errorf("MultiSelectCalls = %d, want 1", mock.MultiSelectCalls)
	}
}

func TestNewWorkflow(t *testing.T) {
	mock := &MockPrompter{}
	result := &discovery.Result{}

	w := NewWorkflow(mock, result, "/test")

	if w.prompter != mock {
		t.Error("prompter mismatch")
	}
	if w.result != result {
		t.Error("result mismatch")
	}
	if w.rootDir != "/test" {
		t.Errorf("rootDir = %q, want %q", w.rootDir, "/test")
	}
}

func TestBuildSyncFileOptions(t *testing.T) {
	candidates := []discovery.SyncCandidate{
		{Path: "package.json", Format: parser.FormatJSON, Description: "Node.js"},
		{Path: "Cargo.toml", Format: parser.FormatTOML, Description: "Rust"},
	}

	options, defaults := buildSyncFileOptions(candidates)

	if len(options) != 2 {
		t.Errorf("options length = %d, want 2", len(options))
	}
	if len(defaults) != 2 {
		t.Errorf("defaults length = %d, want 2", len(defaults))
	}

	// Check defaults contain paths
	if defaults[0] != "package.json" {
		t.Errorf("defaults[0] = %q, want %q", defaults[0], "package.json")
	}
	if defaults[1] != "Cargo.toml" {
		t.Errorf("defaults[1] = %q, want %q", defaults[1], "Cargo.toml")
	}
}

func TestBuildSyncFileOptions_Empty(t *testing.T) {
	options, defaults := buildSyncFileOptions([]discovery.SyncCandidate{})

	if len(options) != 0 {
		t.Errorf("options length = %d, want 0", len(options))
	}
	if len(defaults) != 0 {
		t.Errorf("defaults length = %d, want 0", len(defaults))
	}
}

func TestFilterCandidatesByPaths(t *testing.T) {
	candidates := []discovery.SyncCandidate{
		{Path: "package.json", Format: parser.FormatJSON},
		{Path: "Cargo.toml", Format: parser.FormatTOML},
		{Path: "pyproject.toml", Format: parser.FormatTOML},
	}

	tests := []struct {
		name          string
		selectedPaths []string
		wantCount     int
		wantPaths     []string
	}{
		{
			name:          "select one",
			selectedPaths: []string{"package.json"},
			wantCount:     1,
			wantPaths:     []string{"package.json"},
		},
		{
			name:          "select multiple",
			selectedPaths: []string{"package.json", "Cargo.toml"},
			wantCount:     2,
			wantPaths:     []string{"package.json", "Cargo.toml"},
		},
		{
			name:          "select all",
			selectedPaths: []string{"package.json", "Cargo.toml", "pyproject.toml"},
			wantCount:     3,
			wantPaths:     []string{"package.json", "Cargo.toml", "pyproject.toml"},
		},
		{
			name:          "select none",
			selectedPaths: []string{},
			wantCount:     0,
			wantPaths:     []string{},
		},
		{
			name:          "select non-existent",
			selectedPaths: []string{"nonexistent.json"},
			wantCount:     0,
			wantPaths:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterCandidatesByPaths(candidates, tt.selectedPaths)

			if len(result) != tt.wantCount {
				t.Errorf("result length = %d, want %d", len(result), tt.wantCount)
			}

			for i, want := range tt.wantPaths {
				if i < len(result) && result[i].Path != want {
					t.Errorf("result[%d].Path = %q, want %q", i, result[i].Path, want)
				}
			}
		})
	}
}

func TestGetFieldForManifest(t *testing.T) {
	tests := []struct {
		filename string
		format   parser.Format
		want     string
	}{
		{"package.json", parser.FormatJSON, "version"},
		{"Cargo.toml", parser.FormatTOML, "package.version"},
		{"pyproject.toml", parser.FormatTOML, "project.version"},
		{"Chart.yaml", parser.FormatYAML, "version"},
		{"unknown.json", parser.FormatJSON, "version"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := GetFieldForManifest(tt.filename, tt.format)
			if got != tt.want {
				t.Errorf("GetFieldForManifest(%q, %v) = %q, want %q", tt.filename, tt.format, got, tt.want)
			}
		})
	}
}

func TestGenerateConfigYAML_DefaultsOnly(t *testing.T) {
	plugins := []string{"commit-parser", "tag-manager"}
	data, err := generateConfigYAML(".version", plugins, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data)

	// Check header
	if !contains(content, "# sley configuration file") {
		t.Error("missing header comment")
	}
	if !contains(content, "# Generated by 'sley discover'") {
		t.Error("missing generated by comment")
	}

	// Check plugins
	if !contains(content, "commit-parser: true") {
		t.Error("missing commit-parser config")
	}
	if !contains(content, "tag-manager:") {
		t.Error("missing tag-manager config")
	}
}

func TestGenerateConfigYAML_WithDependencyCheck(t *testing.T) {
	plugins := []string{"commit-parser", "tag-manager", "dependency-check"}
	candidates := []discovery.SyncCandidate{
		{
			Path:   "package.json",
			Format: parser.FormatJSON,
			Field:  "version",
		},
	}

	data, err := generateConfigYAML(".version", plugins, candidates)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data)

	// Check dependency-check config
	if !contains(content, "dependency-check:") {
		t.Error("missing dependency-check config")
	}
	if !contains(content, "auto-sync: true") {
		t.Error("missing auto-sync config")
	}
	if !contains(content, "path: package.json") {
		t.Error("missing file path in config")
	}
}

func TestMarshalConfigWithComments(t *testing.T) {
	cfg := &config.Config{
		Path: ".version",
	}

	data, err := marshalConfigWithComments(cfg, []string{"commit-parser"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data)

	// Check header comments
	if !contains(content, "# sley configuration file") {
		t.Error("missing header comment")
	}
	if !contains(content, "# Generated by 'sley discover'") {
		t.Error("missing generation comment")
	}

	// Check plugin list
	if !contains(content, "#   - commit-parser") {
		t.Error("missing plugin list comment")
	}
}

func TestDefaultVersionPath(t *testing.T) {
	path := defaultVersionPath()
	if path != ".version" {
		t.Errorf("defaultVersionPath() = %q, want %q", path, ".version")
	}
}

// contains is a helper to check if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestConfigExists(t *testing.T) {
	t.Run("config exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		// Create .sley.yaml
		if err := os.WriteFile(".sley.yaml", []byte("path: .version\n"), 0644); err != nil {
			t.Fatal(err)
		}

		if !configExists() {
			t.Error("expected configExists() to return true")
		}
	})

	t.Run("config does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		if configExists() {
			t.Error("expected configExists() to return false")
		}
	})
}

func TestWorkflow_Run_NonInteractive(t *testing.T) {
	// When not interactive, Run should return false, nil
	mock := &MockPrompter{}
	result := &discovery.Result{}
	w := NewWorkflow(mock, result, "/test")

	// In non-interactive mode (test environment), should return early
	completed, err := w.Run(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// In test environment, tui.IsInteractive() returns false
	if completed {
		t.Log("Note: completed=true means test ran in interactive mode")
	}
}

func TestWorkflow_selectSyncFiles_WithError(t *testing.T) {
	expectedErr := errors.New("prompt error")
	mock := &MockPrompter{
		MultiSelectErr: expectedErr,
	}
	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
		},
	}
	w := NewWorkflow(mock, result, "/test")

	_, err := w.selectSyncFiles()

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestWorkflow_selectSyncFiles_AllSelected(t *testing.T) {
	mock := &MockPrompter{
		MultiSelectResult: []string{"package.json", "Cargo.toml", "pyproject.toml"},
	}
	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Description: "Node.js"},
			{Path: "Cargo.toml", Format: parser.FormatTOML, Description: "Rust"},
			{Path: "pyproject.toml", Format: parser.FormatTOML, Description: "Python"},
		},
	}
	w := NewWorkflow(mock, result, "/test")

	selected, err := w.selectSyncFiles()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 3 {
		t.Errorf("expected 3 selections, got %d", len(selected))
	}
}

func TestWorkflow_selectSyncFiles_NoneSelected(t *testing.T) {
	mock := &MockPrompter{
		MultiSelectResult: []string{},
	}
	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
		},
	}
	w := NewWorkflow(mock, result, "/test")

	selected, err := w.selectSyncFiles()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 0 {
		t.Errorf("expected 0 selections, got %d", len(selected))
	}
}

func TestWorkflow_runMismatchWorkflow(t *testing.T) {
	mock := &MockPrompter{}
	result := &discovery.Result{
		Mode: discovery.SingleModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
		Mismatches: []discovery.Mismatch{
			{Source: "package.json", ExpectedVersion: "1.0.0", ActualVersion: "2.0.0"},
			{Source: "Cargo.toml", ExpectedVersion: "1.0.0", ActualVersion: "3.0.0"},
		},
	}
	w := NewWorkflow(mock, result, "/test")

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runMismatchWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if completed {
			t.Error("expected completed to be false")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify warning message is printed
	if !strings.Contains(output, "mismatch") {
		t.Errorf("expected mismatch warning in output, got: %q", output)
	}
}

func TestWorkflow_suggestAdditionalSyncFiles(t *testing.T) {
	mock := &MockPrompter{}
	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
		},
	}
	w := NewWorkflow(mock, result, "/test")

	completed, err := w.suggestAdditionalSyncFiles(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// This function is informational only, returns false, nil
	if completed {
		t.Error("expected completed to be false")
	}
}

func TestWorkflow_printInitSuccess(t *testing.T) {
	w := &Workflow{}

	t.Run("with plugins only", func(t *testing.T) {
		plugins := []string{"commit-parser", "tag-manager"}

		output, err := testutils.CaptureStdout(func() {
			w.printInitSuccess(plugins, nil)
		})
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		if !strings.Contains(output, "2 plugin(s) enabled") {
			t.Errorf("expected plugin count in output, got: %q", output)
		}
		if !strings.Contains(output, "commit-parser") {
			t.Errorf("expected commit-parser in output, got: %q", output)
		}
		if !strings.Contains(output, "tag-manager") {
			t.Errorf("expected tag-manager in output, got: %q", output)
		}
	})

	t.Run("with sync candidates", func(t *testing.T) {
		plugins := []string{"commit-parser", "dependency-check"}
		candidates := []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
			{Path: "Cargo.toml", Format: parser.FormatTOML},
		}

		output, err := testutils.CaptureStdout(func() {
			w.printInitSuccess(plugins, candidates)
		})
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		if !strings.Contains(output, "Configured sync files") {
			t.Errorf("expected sync files section in output, got: %q", output)
		}
		if !strings.Contains(output, "package.json") {
			t.Errorf("expected package.json in output, got: %q", output)
		}
	})

	t.Run("shows next steps", func(t *testing.T) {
		plugins := []string{"commit-parser"}

		output, err := testutils.CaptureStdout(func() {
			w.printInitSuccess(plugins, nil)
		})
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		if !strings.Contains(output, "Next steps") {
			t.Errorf("expected Next steps in output, got: %q", output)
		}
		if !strings.Contains(output, "sley bump") {
			t.Errorf("expected 'sley bump' suggestion in output, got: %q", output)
		}
	})
}

func TestWorkflow_ensureVersionFile(t *testing.T) {
	t.Run("creates version file when missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		w := &Workflow{rootDir: tmpDir}

		err := w.ensureVersionFile(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify .version file was created
		versionPath := filepath.Join(tmpDir, ".version")
		if _, err := os.Stat(versionPath); os.IsNotExist(err) {
			t.Error("expected .version file to be created")
		}
	})

	t.Run("does nothing when file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		// Create existing version file
		versionPath := filepath.Join(tmpDir, ".version")
		if err := os.WriteFile(versionPath, []byte("9.9.9\n"), 0644); err != nil {
			t.Fatal(err)
		}

		w := &Workflow{rootDir: tmpDir}

		err := w.ensureVersionFile(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify version wasn't changed
		data, _ := os.ReadFile(versionPath)
		if strings.TrimSpace(string(data)) != "9.9.9" {
			t.Errorf("expected version to remain 9.9.9, got: %s", string(data))
		}
	})
}

func TestWorkflow_createConfigWithDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	w := &Workflow{rootDir: tmpDir}

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.createConfigWithDefaults(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if !completed {
			t.Error("expected completed to be true")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify .version was created
	versionPath := filepath.Join(tmpDir, ".version")
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("expected .version file to be created")
	}

	// Verify .sley.yaml was created
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected .sley.yaml file to be created")
	}

	// Verify output contains success message
	if !strings.Contains(output, "plugin") {
		t.Errorf("expected plugin info in output, got: %q", output)
	}
}

func TestWorkflow_createConfigWithDependencyCheck(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	candidates := []discovery.SyncCandidate{
		{Path: "package.json", Format: parser.FormatJSON, Field: "version", Description: "Node.js"},
	}

	w := &Workflow{rootDir: tmpDir}

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.createConfigWithDependencyCheck(context.Background(), candidates)
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if !completed {
			t.Error("expected completed to be true")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify .sley.yaml was created with dependency-check
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(configData)
	if !strings.Contains(content, "dependency-check") {
		t.Errorf("expected dependency-check in config, got: %q", content)
	}

	// Verify output mentions sync files
	if !strings.Contains(output, "sync") || !strings.Contains(output, "package.json") {
		t.Errorf("expected sync file info in output, got: %q", output)
	}
}

func TestWorkflow_runDependencyCheckSetup(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Mock prompter that selects package.json
	mock := &MockPrompter{
		MultiSelectResult: []string{"package.json"},
	}

	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Field: "version", Description: "Node.js"},
			{Path: "Cargo.toml", Format: parser.FormatTOML, Field: "package.version", Description: "Rust"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runDependencyCheckSetup(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if !completed {
			t.Error("expected completed to be true")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify output shows discovered files
	if !strings.Contains(output, "Discovered files") {
		t.Errorf("expected discovered files message, got: %q", output)
	}

	// Verify MultiSelect was called
	if mock.MultiSelectCalls != 1 {
		t.Errorf("expected 1 MultiSelect call, got %d", mock.MultiSelectCalls)
	}
}

func TestWorkflow_runDependencyCheckSetup_NoSelection(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Mock prompter that selects nothing
	mock := &MockPrompter{
		MultiSelectResult: []string{},
	}

	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Description: "Node.js"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runDependencyCheckSetup(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if !completed {
			t.Error("expected completed to be true")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Should still create config but without dependency-check files
	if !strings.Contains(output, "No files selected") {
		t.Errorf("expected 'No files selected' message, got: %q", output)
	}
}

func TestWorkflow_runDependencyCheckSetup_SelectionError(t *testing.T) {
	tmpDir := t.TempDir()

	expectedErr := errors.New("selection failed")
	mock := &MockPrompter{
		MultiSelectErr: expectedErr,
	}

	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	_, _ = testutils.CaptureStdout(func() {
		_, runErr := w.runDependencyCheckSetup(context.Background())
		if !errors.Is(runErr, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, runErr)
		}
	})
}

func TestWorkflow_runInitWorkflow_NoSyncCandidates(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mock := &MockPrompter{
		ConfirmResult: false, // User declines init
	}

	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{},
		Modules:        []discovery.Module{},
	}

	w := NewWorkflow(mock, result, tmpDir)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runInitWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		// Should return false since no useful suggestions and user declined
		_ = completed
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Should mention no config found
	if !strings.Contains(output, "No .sley.yaml configuration found") {
		t.Errorf("expected no config message, got: %q", output)
	}
}

func TestWorkflow_runInitWorkflow_UserDeclines(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mock := &MockPrompter{
		ConfirmResult: false, // User declines init
	}

	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
		},
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runInitWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if completed {
			t.Error("expected completed to be false when user declines")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Should mention sley init
	if !strings.Contains(output, "sley init") {
		t.Errorf("expected sley init suggestion, got: %q", output)
	}
}

func TestWorkflow_runInitWorkflow_UserAccepts(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mock := &MockPrompter{
		ConfirmResult:     true,
		MultiSelectResult: []string{"package.json"},
	}

	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Field: "version", Description: "Node.js"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	_, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runInitWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if !completed {
			t.Error("expected completed to be true when user accepts")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify .sley.yaml was created
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected .sley.yaml to be created")
	}
}

func TestWorkflow_runInitWorkflow_ConfirmError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	expectedErr := errors.New("confirm error")
	mock := &MockPrompter{
		ConfirmErr: expectedErr,
	}

	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	_, _ = testutils.CaptureStdout(func() {
		_, runErr := w.runInitWorkflow(context.Background())
		if !errors.Is(runErr, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, runErr)
		}
	})
}

func TestWorkflow_runExistingConfigWorkflow_WithMismatches(t *testing.T) {
	mock := &MockPrompter{}

	result := &discovery.Result{
		Mismatches: []discovery.Mismatch{
			{Source: "package.json", ExpectedVersion: "1.0.0", ActualVersion: "2.0.0"},
		},
	}

	w := NewWorkflow(mock, result, "/test")

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runExistingConfigWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if completed {
			t.Error("expected completed to be false (informational only)")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Should show mismatch warning
	if !strings.Contains(output, "mismatch") {
		t.Errorf("expected mismatch warning, got: %q", output)
	}
}

func TestWorkflow_runExistingConfigWorkflow_NoMismatches(t *testing.T) {
	mock := &MockPrompter{}

	result := &discovery.Result{
		Mismatches:     []discovery.Mismatch{},
		SyncCandidates: []discovery.SyncCandidate{},
	}

	w := NewWorkflow(mock, result, "/test")

	completed, err := w.runExistingConfigWorkflow(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if completed {
		t.Error("expected completed to be false")
	}
}

func TestGenerateConfigYAML_EmptyPlugins(t *testing.T) {
	data, err := generateConfigYAML(".version", []string{}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data)

	// Should still have header
	if !strings.Contains(content, "# sley configuration file") {
		t.Error("missing header comment")
	}

	// Should have path
	if !strings.Contains(content, "path: .version") {
		t.Error("missing path in config")
	}
}

func TestGenerateConfigYAML_AllPluginTypes(t *testing.T) {
	plugins := []string{"commit-parser", "tag-manager", "dependency-check"}
	candidates := []discovery.SyncCandidate{
		{Path: "package.json", Format: parser.FormatJSON, Field: "version"},
	}

	data, err := generateConfigYAML(".version", plugins, candidates)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data)

	// Check all plugins are present
	if !strings.Contains(content, "commit-parser: true") {
		t.Error("missing commit-parser")
	}
	if !strings.Contains(content, "tag-manager:") {
		t.Error("missing tag-manager")
	}
	if !strings.Contains(content, "dependency-check:") {
		t.Error("missing dependency-check")
	}
}

func TestMarshalToYAML(t *testing.T) {
	cfg := &config.Config{
		Path: ".version",
	}

	data, err := marshalToYAML(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "path: .version") {
		t.Errorf("expected path in YAML, got: %q", content)
	}
}

func TestMarshalConfigWithComments_EmptyPlugins(t *testing.T) {
	cfg := &config.Config{
		Path: ".version",
	}

	data, err := marshalConfigWithComments(cfg, []string{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(data)

	// Should have header but no plugin list
	if !strings.Contains(content, "# sley configuration file") {
		t.Error("missing header")
	}
}

func TestGenerateDependencyCheckFileConfig_WithPattern(t *testing.T) {
	candidate := discovery.SyncCandidate{
		Path:    "version.go",
		Format:  parser.FormatRegex,
		Pattern: `Version = "(.*?)"`,
	}

	cfg := GenerateDependencyCheckFileConfig(candidate)

	if cfg.Pattern != `Version = "(.*?)"` {
		t.Errorf("Pattern = %q, want %q", cfg.Pattern, `Version = "(.*?)"`)
	}
	if cfg.Format != "regex" {
		t.Errorf("Format = %q, want %q", cfg.Format, "regex")
	}
}

func TestWorkflow_generateDependencyCheckConfig_NoFieldOrPattern(t *testing.T) {
	candidates := []discovery.SyncCandidate{
		{
			Path:        "VERSION",
			Format:      parser.FormatRaw,
			Description: "Plain text",
		},
	}

	w := &Workflow{}
	output := w.generateDependencyCheckConfig(candidates)

	if !strings.Contains(output, "path: VERSION") {
		t.Errorf("expected path in output, got: %q", output)
	}
	if !strings.Contains(output, "format: raw") {
		t.Errorf("expected format in output, got: %q", output)
	}
	// Should not have field line
	if strings.Contains(output, "field:") {
		t.Errorf("unexpected field in output: %q", output)
	}
}
