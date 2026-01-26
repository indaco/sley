package discover

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/parser"
	"github.com/indaco/sley/internal/testutils"
)

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
