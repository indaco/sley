package discover

import (
	"strings"
	"testing"

	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/parser"
	"github.com/indaco/sley/internal/testutils"
)

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		input string
		want  OutputFormat
	}{
		{"text", FormatText},
		{"json", FormatJSON},
		{"table", FormatTable},
		{"", FormatText},
		{"invalid", FormatText},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseOutputFormat(tt.input)
			if got != tt.want {
				t.Errorf("ParseOutputFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatter_FormatResult_Text(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.SingleModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
		Manifests: []discovery.ManifestSource{
			{
				RelPath:     "package.json",
				Filename:    "package.json",
				Version:     "1.0.0",
				Format:      parser.FormatJSON,
				Description: "Node.js (package.json)",
			},
		},
	}

	formatter := NewFormatter(FormatText)
	output := formatter.FormatResult(result)

	// Verify key elements are present
	checks := []string{
		"Discovery Results",
		"Version Files",
		".version",
		"1.0.0",
		"Manifest Files",
		"package.json",
		"Node.js",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing expected text %q", check)
		}
	}
}

func TestFormatter_FormatResult_JSON(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.SingleModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
	}

	formatter := NewFormatter(FormatJSON)
	output := formatter.FormatResult(result)

	// Verify JSON structure
	checks := []string{
		`"mode"`,
		`"SingleModule"`,
		`"modules"`,
		`"version"`,
		`"1.0.0"`,
		`"summary"`,
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("JSON output missing expected text %q", check)
		}
	}
}

func TestFormatter_FormatResult_Table(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.SingleModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
	}

	formatter := NewFormatter(FormatTable)
	output := formatter.FormatResult(result)

	// Verify table headers
	checks := []string{
		"PATH",
		"VERSION",
		".version",
		"1.0.0",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("table output missing expected text %q", check)
		}
	}
}

func TestFormatter_FormatResult_WithMismatches(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.SingleModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
		Mismatches: []discovery.Mismatch{
			{Source: "package.json", ExpectedVersion: "1.0.0", ActualVersion: "2.0.0"},
		},
	}

	formatter := NewFormatter(FormatText)
	output := formatter.FormatResult(result)

	// Verify mismatch is shown
	checks := []string{
		"Mismatch",
		"package.json",
		"expected 1.0.0",
		"found 2.0.0",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing expected text %q", check)
		}
	}
}

func TestFormatter_FormatResult_Empty(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.NoModules,
	}

	formatter := NewFormatter(FormatText)
	output := formatter.FormatResult(result)

	if !strings.Contains(output, "No version sources found") {
		t.Error("empty result should indicate no sources found")
	}
}

func TestGetModeDescription(t *testing.T) {
	tests := []struct {
		mode discovery.DetectionMode
		want string
	}{
		{discovery.SingleModule, "Single Module"},
		{discovery.MultiModule, "Multi-Module (Monorepo)"},
		{discovery.NoModules, "No .version files found"},
		{discovery.DetectionMode(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			got := getModeDescription(tt.mode)
			if got != tt.want {
				t.Errorf("getModeDescription(%v) = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}

func TestFormatter_formatSummary(t *testing.T) {
	tests := []struct {
		name           string
		result         *discovery.Result
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "with modules and manifests",
			result: &discovery.Result{
				Modules: []discovery.Module{
					{Version: "1.0.0", RelPath: ".version"},
				},
				Manifests: []discovery.ManifestSource{
					{RelPath: "package.json"},
				},
			},
			wantContains: []string{"1 version file", "1 manifest"},
		},
		{
			name: "with mismatches",
			result: &discovery.Result{
				Modules: []discovery.Module{
					{Version: "1.0.0", RelPath: ".version"},
				},
				Mismatches: []discovery.Mismatch{
					{Source: "test"},
				},
			},
			wantContains: []string{"1 mismatch"},
		},
		{
			name:         "empty",
			result:       &discovery.Result{},
			wantContains: []string{"No version sources found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(FormatText)
			summary := formatter.formatSummary(tt.result)

			for _, want := range tt.wantContains {
				if !strings.Contains(summary, want) {
					t.Errorf("summary %q should contain %q", summary, want)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(summary, notWant) {
					t.Errorf("summary %q should not contain %q", summary, notWant)
				}
			}
		})
	}
}

func TestFormatter_PrintResult(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.SingleModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
	}

	formatter := NewFormatter(FormatText)

	output, err := testutils.CaptureStdout(func() {
		formatter.PrintResult(result)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify output is printed
	if !strings.Contains(output, "Discovery Results") {
		t.Errorf("expected Discovery Results in output, got: %q", output)
	}
	if !strings.Contains(output, "1.0.0") {
		t.Errorf("expected version in output, got: %q", output)
	}
}

func TestFormatter_PrintResult_JSON(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "2.0.0"},
			{Name: "sub", RelPath: "sub/.version", Version: "2.0.0"},
		},
	}

	formatter := NewFormatter(FormatJSON)

	output, err := testutils.CaptureStdout(func() {
		formatter.PrintResult(result)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify JSON is printed
	if !strings.Contains(output, `"mode"`) {
		t.Errorf("expected JSON mode field in output, got: %q", output)
	}
}

func TestFormatter_PrintResult_Table(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.NoModules,
		Manifests: []discovery.ManifestSource{
			{
				RelPath:     "package.json",
				Filename:    "package.json",
				Version:     "3.0.0",
				Format:      parser.FormatJSON,
				Description: "Node.js",
			},
		},
	}

	formatter := NewFormatter(FormatTable)

	output, err := testutils.CaptureStdout(func() {
		formatter.PrintResult(result)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify table is printed
	if !strings.Contains(output, "package.json") {
		t.Errorf("expected package.json in output, got: %q", output)
	}
}

func TestFormatter_formatTable_WithManifests(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.NoModules,
		Manifests: []discovery.ManifestSource{
			{
				RelPath:     "package.json",
				Filename:    "package.json",
				Version:     "1.0.0",
				Format:      parser.FormatJSON,
				Description: "Node.js (package.json)",
			},
			{
				RelPath:     "Cargo.toml",
				Filename:    "Cargo.toml",
				Version:     "1.0.0",
				Format:      parser.FormatTOML,
				Description: "Rust (Cargo.toml)",
			},
		},
	}

	formatter := NewFormatter(FormatTable)
	output := formatter.formatTable(result)

	// Verify manifest table structure
	checks := []string{
		"Manifest Files:",
		"PATH",
		"VERSION",
		"TYPE",
		"package.json",
		"Cargo.toml",
		"Node.js",
		"Rust",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("table output missing expected text %q", check)
		}
	}
}

func TestFormatter_formatTable_WithMismatches(t *testing.T) {
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

	formatter := NewFormatter(FormatTable)
	output := formatter.formatTable(result)

	// Verify mismatch table structure
	checks := []string{
		"Version Mismatches:",
		"SOURCE",
		"EXPECTED",
		"ACTUAL",
		"package.json",
		"Cargo.toml",
		"1.0.0",
		"2.0.0",
		"3.0.0",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("table output missing expected text %q", check)
		}
	}
}

func TestFormatter_formatTable_Empty(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.NoModules,
	}

	formatter := NewFormatter(FormatTable)
	output := formatter.formatTable(result)

	// Should have header and summary
	if !strings.Contains(output, "Discovery Results") {
		t.Error("expected Discovery Results header")
	}
	if !strings.Contains(output, "No version sources found") {
		t.Error("expected no sources message")
	}
}

func TestFormatter_formatText_WithSyncCandidates(t *testing.T) {
	// Test when there are sync candidates but no modules
	result := &discovery.Result{
		Mode: discovery.NoModules,
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Description: "Node.js"},
			{Path: "Cargo.toml", Format: parser.FormatTOML, Description: "Rust"},
		},
	}

	formatter := NewFormatter(FormatText)
	output := formatter.formatText(result)

	// Should show sync candidates
	if !strings.Contains(output, "Sync Candidates") {
		t.Error("expected Sync Candidates section")
	}
	if !strings.Contains(output, "package.json") {
		t.Error("expected package.json in sync candidates")
	}
}

func TestFormatter_formatText_WithModulesNoSyncCandidates(t *testing.T) {
	// Test when there are modules - sync candidates section should be hidden
	result := &discovery.Result{
		Mode: discovery.SingleModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Description: "Node.js"},
		},
	}

	formatter := NewFormatter(FormatText)
	output := formatter.formatText(result)

	// Should NOT show sync candidates when modules exist
	// (based on the condition !result.HasModules())
	if strings.Contains(output, "Sync Candidates") {
		t.Error("should not show Sync Candidates when modules exist")
	}
}

func TestFormatter_formatJSON_Complete(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
			{Name: "sub", RelPath: "sub/.version", Version: "1.0.0"},
		},
		Manifests: []discovery.ManifestSource{
			{
				RelPath:     "package.json",
				Filename:    "package.json",
				Version:     "1.0.0",
				Format:      parser.FormatJSON,
				Field:       "version",
				Description: "Node.js",
			},
		},
		SyncCandidates: []discovery.SyncCandidate{
			{
				Path:        "package.json",
				Format:      parser.FormatJSON,
				Field:       "version",
				Description: "Node.js",
			},
		},
		Mismatches: []discovery.Mismatch{
			{Source: "Cargo.toml", ExpectedVersion: "1.0.0", ActualVersion: "2.0.0"},
		},
	}

	formatter := NewFormatter(FormatJSON)
	output := formatter.formatJSON(result)

	// Verify all JSON sections
	checks := []string{
		`"mode"`,
		`"MultiModule"`,
		`"modules"`,
		`"manifests"`,
		`"sync_candidates"`,
		`"mismatches"`,
		`"summary"`,
		`"module_count"`,
		`"manifest_count"`,
		`"mismatch_count"`,
		`"has_mismatches": true`,
		`"primary_version"`,
		`"is_version_consistent"`,
	}

	for _, check := range checks {
		if !strings.Contains(output, check) && !strings.Contains(output, strings.ReplaceAll(check, " ", "")) {
			t.Errorf("JSON output missing expected text %q", check)
		}
	}
}

func TestFormatter_formatJSON_WithPattern(t *testing.T) {
	result := &discovery.Result{
		Mode: discovery.NoModules,
		SyncCandidates: []discovery.SyncCandidate{
			{
				Path:        "version.go",
				Format:      parser.FormatRegex,
				Pattern:     `Version = "(.*?)"`,
				Description: "Go version",
			},
		},
	}

	formatter := NewFormatter(FormatJSON)
	output := formatter.formatJSON(result)

	// Verify pattern is included
	if !strings.Contains(output, `"pattern"`) {
		t.Error("expected pattern field in JSON output")
	}
}

func TestFormatter_formatSummary_WithPrimaryVersion(t *testing.T) {
	result := &discovery.Result{
		Modules: []discovery.Module{
			{Version: "5.0.0", RelPath: ".version"},
		},
	}

	formatter := NewFormatter(FormatText)
	summary := formatter.formatSummary(result)

	if !strings.Contains(summary, "Primary version") {
		t.Errorf("expected Primary version in summary, got: %q", summary)
	}
	if !strings.Contains(summary, "5.0.0") {
		t.Errorf("expected version 5.0.0 in summary, got: %q", summary)
	}
}

func TestFormatter_formatSummary_MultipleModulesAndManifests(t *testing.T) {
	result := &discovery.Result{
		Modules: []discovery.Module{
			{Version: "1.0.0", RelPath: ".version"},
			{Version: "1.0.0", RelPath: "sub/.version"},
		},
		Manifests: []discovery.ManifestSource{
			{RelPath: "package.json"},
			{RelPath: "Cargo.toml"},
			{RelPath: "pyproject.toml"},
		},
	}

	formatter := NewFormatter(FormatText)
	summary := formatter.formatSummary(result)

	if !strings.Contains(summary, "2 version file") {
		t.Errorf("expected '2 version file' in summary, got: %q", summary)
	}
	if !strings.Contains(summary, "3 manifest") {
		t.Errorf("expected '3 manifest' in summary, got: %q", summary)
	}
}

func TestFormatter_formatSummary_MultipleMismatches(t *testing.T) {
	result := &discovery.Result{
		Modules: []discovery.Module{
			{Version: "1.0.0", RelPath: ".version"},
		},
		Mismatches: []discovery.Mismatch{
			{Source: "a"},
			{Source: "b"},
			{Source: "c"},
		},
	}

	formatter := NewFormatter(FormatText)
	summary := formatter.formatSummary(result)

	if !strings.Contains(summary, "3 mismatch") {
		t.Errorf("expected '3 mismatch' in summary, got: %q", summary)
	}
}

func TestParseOutputFormat_CaseInsensitive(t *testing.T) {
	// The current implementation is case-sensitive
	// These tests document that behavior
	tests := []struct {
		input string
		want  OutputFormat
	}{
		{"JSON", FormatText}, // Case-sensitive, returns default
		{"Json", FormatText},
		{"TABLE", FormatText},
		{"Table", FormatText},
		{"TEXT", FormatText},
		{"Text", FormatText},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseOutputFormat(tt.input)
			if got != tt.want {
				t.Errorf("ParseOutputFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		format OutputFormat
	}{
		{FormatText},
		{FormatJSON},
		{FormatTable},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			f := NewFormatter(tt.format)
			if f == nil {
				t.Fatal("expected non-nil formatter")
			}
			if f.format != tt.format {
				t.Errorf("format = %v, want %v", f.format, tt.format)
			}
		})
	}
}

func TestFormatter_FormatResult_DefaultCase(t *testing.T) {
	// Test that an invalid/unknown format defaults to text
	result := &discovery.Result{
		Mode: discovery.SingleModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
	}

	formatter := &Formatter{format: OutputFormat("invalid")}
	output := formatter.FormatResult(result)

	// Should use text format as default
	if !strings.Contains(output, "Discovery Results") {
		t.Error("expected text format output for invalid format")
	}
}
