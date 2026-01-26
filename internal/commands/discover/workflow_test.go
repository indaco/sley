package discover

import (
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/parser"
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
