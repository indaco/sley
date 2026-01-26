package discover

import (
	"strings"
	"testing"

	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/parser"
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
