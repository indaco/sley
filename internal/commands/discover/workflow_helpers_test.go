package discover

import (
	"errors"
	"testing"

	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/parser"
)

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
