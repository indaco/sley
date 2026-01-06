package workspace

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewTextFormatter(t *testing.T) {
	formatter := NewTextFormatter("test operation")
	if formatter == nil {
		t.Fatal("NewTextFormatter() returned nil")
	}
	if formatter.operation != "test operation" {
		t.Errorf("operation = %q, want %q", formatter.operation, "test operation")
	}
}

func TestTextFormatter_FormatResults_Empty(t *testing.T) {
	formatter := NewTextFormatter("bump")
	result := formatter.FormatResults([]ExecutionResult{})

	expected := "No results to display."
	if result != expected {
		t.Errorf("FormatResults([]) = %q, want %q", result, expected)
	}
}

func TestTextFormatter_FormatResults_Success(t *testing.T) {
	formatter := NewTextFormatter("Version Bump")

	modules := []*Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
	}

	results := []ExecutionResult{
		{
			Module:     modules[0],
			OldVersion: "1.0.0",
			NewVersion: "1.1.0",
			Success:    true,
			Duration:   100 * time.Millisecond,
		},
		{
			Module:     modules[1],
			OldVersion: "2.0.0",
			NewVersion: "2.1.0",
			Success:    true,
			Duration:   150 * time.Millisecond,
		},
	}

	output := formatter.FormatResults(results)

	// Verify key elements are present
	if !strings.Contains(output, "Version Bump") {
		t.Error("Output should contain operation name")
	}
	if !strings.Contains(output, "module-a") {
		t.Error("Output should contain module-a")
	}
	if !strings.Contains(output, "1.0.0 -> 1.1.0") {
		t.Error("Output should contain version change")
	}
	if !strings.Contains(output, "Success: 2 modules updated") {
		t.Error("Output should contain success summary")
	}
}

func TestTextFormatter_FormatResults_MixedResults(t *testing.T) {
	formatter := NewTextFormatter("bump")

	modules := []*Module{
		{Name: "module-success", Path: "/path/to/success/.version"},
		{Name: "module-fail", Path: "/path/to/fail/.version"},
	}

	results := []ExecutionResult{
		{
			Module:     modules[0],
			OldVersion: "1.0.0",
			NewVersion: "1.1.0",
			Success:    true,
			Duration:   50 * time.Millisecond,
		},
		{
			Module:   modules[1],
			Success:  false,
			Error:    errors.New("test error"),
			Duration: 10 * time.Millisecond,
		},
	}

	output := formatter.FormatResults(results)

	if !strings.Contains(output, "✓ module-success") {
		t.Error("Output should contain success indicator")
	}
	if !strings.Contains(output, "✗ module-fail") {
		t.Error("Output should contain error indicator")
	}
	if !strings.Contains(output, "test error") {
		t.Error("Output should contain error message")
	}
	if !strings.Contains(output, "Completed: 1 succeeded, 1 failed") {
		t.Error("Output should contain mixed results summary")
	}
}

func TestTextFormatter_FormatResults_NoVersionChange(t *testing.T) {
	formatter := NewTextFormatter("")

	module := &Module{Name: "module-a", Path: "/path/.version"}
	results := []ExecutionResult{
		{
			Module:     module,
			NewVersion: "1.0.0",
			Success:    true,
			Duration:   10 * time.Millisecond,
		},
	}

	output := formatter.FormatResults(results)

	if !strings.Contains(output, "module-a: 1.0.0") {
		t.Error("Output should show version without arrow when no change")
	}
}

func TestTextFormatter_FormatModuleList_Empty(t *testing.T) {
	formatter := NewTextFormatter("")
	result := formatter.FormatModuleList([]*Module{})

	expected := "No modules found."
	if result != expected {
		t.Errorf("FormatModuleList([]) = %q, want %q", result, expected)
	}
}

func TestTextFormatter_FormatModuleList(t *testing.T) {
	formatter := NewTextFormatter("")

	modules := []*Module{
		{Name: "api", CurrentVersion: "1.0.0", RelPath: "services/api/.version"},
		{Name: "web", CurrentVersion: "2.1.0", RelPath: "apps/web/.version"},
		{Name: "lib", CurrentVersion: "", RelPath: "packages/lib/.version"},
	}

	output := formatter.FormatModuleList(modules)

	if !strings.Contains(output, "Found 3 modules:") {
		t.Error("Output should contain module count")
	}
	if !strings.Contains(output, "api (1.0.0)") {
		t.Error("Output should contain module with version")
	}
	if !strings.Contains(output, "services/api/.version") {
		t.Error("Output should contain relative path")
	}
	if !strings.Contains(output, "lib") {
		t.Error("Output should contain module without version")
	}
}

func TestTextFormatter_FormatModuleList_SingleModule(t *testing.T) {
	formatter := NewTextFormatter("")

	modules := []*Module{
		{Name: "single", CurrentVersion: "1.0.0"},
	}

	output := formatter.FormatModuleList(modules)

	if !strings.Contains(output, "Found 1 module:") {
		t.Error("Output should use singular 'module'")
	}
}

func TestNewJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter()
	if formatter == nil {
		t.Fatal("NewJSONFormatter() returned nil")
	}
}

func TestJSONFormatter_FormatResults(t *testing.T) {
	formatter := NewJSONFormatter()

	modules := []*Module{
		{Name: "module-a", Path: "/path/a/.version"},
		{Name: "module-b", Path: "/path/b/.version"},
	}

	results := []ExecutionResult{
		{
			Module:     modules[0],
			OldVersion: "1.0.0",
			NewVersion: "1.1.0",
			Success:    true,
			Duration:   100 * time.Millisecond,
		},
		{
			Module:   modules[1],
			Success:  false,
			Error:    errors.New("failed"),
			Duration: 50 * time.Millisecond,
		},
	}

	output := formatter.FormatResults(results)

	// Verify it's valid JSON
	var parsed resultsJSON
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("FormatResults() produced invalid JSON: %v", err)
	}

	if parsed.Total != 2 {
		t.Errorf("total = %d, want 2", parsed.Total)
	}
	if parsed.SuccessCount != 1 {
		t.Errorf("success_count = %d, want 1", parsed.SuccessCount)
	}
	if parsed.ErrorCount != 1 {
		t.Errorf("error_count = %d, want 1", parsed.ErrorCount)
	}
	if len(parsed.Results) != 2 {
		t.Errorf("len(results) = %d, want 2", len(parsed.Results))
	}

	// Check first result
	if parsed.Results[0].Module != "module-a" {
		t.Errorf("results[0].module = %q, want %q", parsed.Results[0].Module, "module-a")
	}
	if parsed.Results[0].OldVersion != "1.0.0" {
		t.Errorf("results[0].old_version = %q, want %q", parsed.Results[0].OldVersion, "1.0.0")
	}
	if parsed.Results[0].NewVersion != "1.1.0" {
		t.Errorf("results[0].new_version = %q, want %q", parsed.Results[0].NewVersion, "1.1.0")
	}
	if !parsed.Results[0].Success {
		t.Error("results[0].success should be true")
	}

	// Check second result
	if parsed.Results[1].Success {
		t.Error("results[1].success should be false")
	}
	if parsed.Results[1].Error != "failed" {
		t.Errorf("results[1].error = %q, want %q", parsed.Results[1].Error, "failed")
	}
}

func TestJSONFormatter_FormatModuleList(t *testing.T) {
	formatter := NewJSONFormatter()

	modules := []*Module{
		{
			Name:           "api",
			Path:           "/workspace/api/.version",
			RelPath:        "api/.version",
			CurrentVersion: "1.0.0",
			Dir:            "/workspace/api",
		},
		{
			Name:           "web",
			Path:           "/workspace/web/.version",
			RelPath:        "web/.version",
			CurrentVersion: "2.0.0",
			Dir:            "/workspace/web",
		},
	}

	output := formatter.FormatModuleList(modules)

	// Verify it's valid JSON
	var parsed modulesJSON
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("FormatModuleList() produced invalid JSON: %v", err)
	}

	if parsed.Total != 2 {
		t.Errorf("total = %d, want 2", parsed.Total)
	}
	if len(parsed.Modules) != 2 {
		t.Fatalf("len(modules) = %d, want 2", len(parsed.Modules))
	}

	// Verify first module
	if parsed.Modules[0].Name != "api" {
		t.Errorf("modules[0].name = %q, want %q", parsed.Modules[0].Name, "api")
	}
	if parsed.Modules[0].CurrentVersion != "1.0.0" {
		t.Errorf("modules[0].current_version = %q, want %q", parsed.Modules[0].CurrentVersion, "1.0.0")
	}
	if parsed.Modules[0].RelPath != "api/.version" {
		t.Errorf("modules[0].rel_path = %q, want %q", parsed.Modules[0].RelPath, "api/.version")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "microseconds",
			duration: 500 * time.Microsecond,
			expected: "500µs",
		},
		{
			name:     "milliseconds",
			duration: 150 * time.Millisecond,
			expected: "150ms",
		},
		{
			name:     "seconds",
			duration: 2500 * time.Millisecond,
			expected: "2.50s",
		},
		{
			name:     "sub-millisecond",
			duration: 100 * time.Nanosecond,
			expected: "0µs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.expected)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{0, "s"},
		{1, ""},
		{2, "s"},
		{100, "s"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.count)), func(t *testing.T) {
			got := pluralize(tt.count)
			if got != tt.expected {
				t.Errorf("pluralize(%d) = %q, want %q", tt.count, got, tt.expected)
			}
		})
	}
}

func TestGetFormatter(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		operation string
		wantType  string
	}{
		{
			name:      "json formatter",
			format:    "json",
			operation: "test",
			wantType:  "*workspace.JSONFormatter",
		},
		{
			name:      "text formatter",
			format:    "text",
			operation: "bump",
			wantType:  "*workspace.TextFormatter",
		},
		{
			name:      "table formatter",
			format:    "table",
			operation: "show",
			wantType:  "*workspace.TableFormatter",
		},
		{
			name:      "default formatter",
			format:    "",
			operation: "show",
			wantType:  "*workspace.TextFormatter",
		},
		{
			name:      "unknown formatter defaults to text",
			format:    "xml",
			operation: "test",
			wantType:  "*workspace.TextFormatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := GetFormatter(tt.format, tt.operation)
			if formatter == nil {
				t.Fatal("GetFormatter() returned nil")
			}

			// Type assertion to verify correct formatter type
			switch tt.wantType {
			case "*workspace.JSONFormatter":
				if _, ok := formatter.(*JSONFormatter); !ok {
					t.Errorf("GetFormatter(%q) returned %T, want JSONFormatter", tt.format, formatter)
				}
			case "*workspace.TextFormatter":
				if _, ok := formatter.(*TextFormatter); !ok {
					t.Errorf("GetFormatter(%q) returned %T, want TextFormatter", tt.format, formatter)
				}
			case "*workspace.TableFormatter":
				if _, ok := formatter.(*TableFormatter); !ok {
					t.Errorf("GetFormatter(%q) returned %T, want TableFormatter", tt.format, formatter)
				}
			}
		})
	}
}

func TestTextFormatter_ImplementsOutputFormatter(t *testing.T) {
	var _ OutputFormatter = (*TextFormatter)(nil)
}

func TestJSONFormatter_ImplementsOutputFormatter(t *testing.T) {
	var _ OutputFormatter = (*JSONFormatter)(nil)
}

func TestTableFormatter_ImplementsOutputFormatter(t *testing.T) {
	var _ OutputFormatter = (*TableFormatter)(nil)
}

func TestNewTableFormatter(t *testing.T) {
	formatter := NewTableFormatter("test operation")
	if formatter == nil {
		t.Fatal("NewTableFormatter() returned nil")
	}
	if formatter.operation != "test operation" {
		t.Errorf("operation = %q, want %q", formatter.operation, "test operation")
	}
}

func TestTableFormatter_FormatResults_Empty(t *testing.T) {
	formatter := NewTableFormatter("bump")
	result := formatter.FormatResults([]ExecutionResult{})

	expected := "No results to display."
	if result != expected {
		t.Errorf("FormatResults([]) = %q, want %q", result, expected)
	}
}

func TestTableFormatter_FormatResults_Success(t *testing.T) {
	formatter := NewTableFormatter("Version Bump")

	modules := []*Module{
		{Name: "module-a", Path: "/path/to/module-a/.version"},
		{Name: "module-b", Path: "/path/to/module-b/.version"},
	}

	results := []ExecutionResult{
		{
			Module:     modules[0],
			OldVersion: "1.0.0",
			NewVersion: "1.1.0",
			Success:    true,
			Duration:   100 * time.Millisecond,
		},
		{
			Module:     modules[1],
			OldVersion: "2.0.0",
			NewVersion: "2.1.0",
			Success:    true,
			Duration:   150 * time.Millisecond,
		},
	}

	output := formatter.FormatResults(results)

	// Verify key elements are present
	if !strings.Contains(output, "Version Bump") {
		t.Error("Output should contain operation name")
	}
	if !strings.Contains(output, "module-a") {
		t.Error("Output should contain module-a")
	}
	if !strings.Contains(output, "1.0.0 -> 1.1.0") {
		t.Error("Output should contain version change")
	}
	if !strings.Contains(output, "Module") {
		t.Error("Output should contain table header 'Module'")
	}
	if !strings.Contains(output, "Version") {
		t.Error("Output should contain table header 'Version'")
	}
	if !strings.Contains(output, "Status") {
		t.Error("Output should contain table header 'Status'")
	}
	if !strings.Contains(output, "Duration") {
		t.Error("Output should contain table header 'Duration'")
	}
	if !strings.Contains(output, "OK") {
		t.Error("Output should contain OK status")
	}
	if !strings.Contains(output, "+") && !strings.Contains(output, "|") {
		t.Error("Output should contain table borders")
	}
	if !strings.Contains(output, "Success: 2 modules updated") {
		t.Error("Output should contain success summary")
	}
}

func TestTableFormatter_FormatResults_MixedResults(t *testing.T) {
	formatter := NewTableFormatter("bump")

	modules := []*Module{
		{Name: "module-success", Path: "/path/to/success/.version"},
		{Name: "module-fail", Path: "/path/to/fail/.version"},
	}

	results := []ExecutionResult{
		{
			Module:     modules[0],
			OldVersion: "1.0.0",
			NewVersion: "1.1.0",
			Success:    true,
			Duration:   50 * time.Millisecond,
		},
		{
			Module:   modules[1],
			Success:  false,
			Error:    errors.New("test error"),
			Duration: 10 * time.Millisecond,
		},
	}

	output := formatter.FormatResults(results)

	if !strings.Contains(output, "OK") {
		t.Error("Output should contain OK status")
	}
	if !strings.Contains(output, "FAILED") {
		t.Error("Output should contain FAILED status")
	}
	if !strings.Contains(output, "Completed: 1 succeeded, 1 failed") {
		t.Error("Output should contain mixed results summary")
	}
}

func TestTableFormatter_FormatModuleList_Empty(t *testing.T) {
	formatter := NewTableFormatter("")
	result := formatter.FormatModuleList([]*Module{})

	expected := "No modules found."
	if result != expected {
		t.Errorf("FormatModuleList([]) = %q, want %q", result, expected)
	}
}

func TestTableFormatter_FormatModuleList(t *testing.T) {
	formatter := NewTableFormatter("")

	modules := []*Module{
		{Name: "api", CurrentVersion: "1.0.0", RelPath: "services/api/.version"},
		{Name: "web", CurrentVersion: "2.1.0", RelPath: "apps/web/.version"},
		{Name: "lib", CurrentVersion: "", RelPath: "packages/lib/.version"},
	}

	output := formatter.FormatModuleList(modules)

	if !strings.Contains(output, "Found 3 modules:") {
		t.Error("Output should contain module count")
	}
	if !strings.Contains(output, "Module") {
		t.Error("Output should contain table header 'Module'")
	}
	if !strings.Contains(output, "Version") {
		t.Error("Output should contain table header 'Version'")
	}
	if !strings.Contains(output, "Path") {
		t.Error("Output should contain table header 'Path'")
	}
	if !strings.Contains(output, "api") {
		t.Error("Output should contain module name")
	}
	if !strings.Contains(output, "1.0.0") {
		t.Error("Output should contain version")
	}
	if !strings.Contains(output, "-") {
		t.Error("Output should contain '-' for empty version")
	}
	if !strings.Contains(output, "+") && !strings.Contains(output, "|") {
		t.Error("Output should contain table borders")
	}
}

func TestTableFormatter_FormatModuleList_LongPath(t *testing.T) {
	formatter := NewTableFormatter("")

	modules := []*Module{
		{
			Name:           "api",
			CurrentVersion: "1.0.0",
			RelPath:        "very/long/path/to/some/deeply/nested/directory/structure/that/exceeds/fifty/characters/api/.version",
		},
	}

	output := formatter.FormatModuleList(modules)

	// Long paths should be truncated
	if !strings.Contains(output, "...") {
		t.Error("Long paths should be truncated with ...")
	}
}

func TestNewTextFormatterWithVerb(t *testing.T) {
	tests := []struct {
		name       string
		operation  string
		actionVerb string
		wantVerb   string
	}{
		{"validated verb", "Validation Summary", "validated", "validated"},
		{"checked verb", "Version Summary", "checked", "checked"},
		{"custom verb", "Custom Op", "processed", "processed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewTextFormatterWithVerb(tt.operation, tt.actionVerb)

			module := &Module{Name: "test-module", Path: "/path/.version"}
			results := []ExecutionResult{
				{
					Module:     module,
					NewVersion: "1.0.0",
					Success:    true,
					Duration:   10 * time.Millisecond,
				},
			}

			output := formatter.FormatResults(results)

			if !strings.Contains(output, tt.operation) {
				t.Errorf("Output should contain operation %q", tt.operation)
			}
			if !strings.Contains(output, tt.wantVerb) {
				t.Errorf("Output should contain verb %q, got: %s", tt.wantVerb, output)
			}
		})
	}
}

func TestNewTableFormatterWithVerb(t *testing.T) {
	formatter := NewTableFormatterWithVerb("Validation Summary", "validated")

	module := &Module{Name: "test-module", Path: "/path/.version"}
	results := []ExecutionResult{
		{
			Module:     module,
			NewVersion: "1.0.0",
			Success:    true,
			Duration:   10 * time.Millisecond,
		},
	}

	output := formatter.FormatResults(results)

	if !strings.Contains(output, "Validation Summary") {
		t.Error("Output should contain operation name")
	}
	if !strings.Contains(output, "validated") {
		t.Errorf("Output should contain 'validated', got: %s", output)
	}
}

func TestGetFormatterWithVerb(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		operation  string
		actionVerb string
	}{
		{"text formatter", "text", "Version Summary", "checked"},
		{"table formatter", "table", "Validation Summary", "validated"},
		{"default to text", "", "Bump Summary", "updated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := GetFormatterWithVerb(tt.format, tt.operation, tt.actionVerb)

			if formatter == nil {
				t.Error("GetFormatterWithVerb should not return nil")
			}

			module := &Module{Name: "test", Path: "/path/.version"}
			results := []ExecutionResult{
				{
					Module:     module,
					NewVersion: "1.0.0",
					Success:    true,
					Duration:   10 * time.Millisecond,
				},
			}

			output := formatter.FormatResults(results)
			if !strings.Contains(output, tt.actionVerb) {
				t.Errorf("Output should contain verb %q", tt.actionVerb)
			}
		})
	}
}

func TestGetFormatterWithVerb_JSON(t *testing.T) {
	// JSON formatter doesn't use actionVerb, just verify it returns JSONFormatter
	formatter := GetFormatterWithVerb("json", "Test Op", "tested")

	if formatter == nil {
		t.Error("GetFormatterWithVerb should not return nil for json")
	}

	module := &Module{Name: "test", Path: "/path/.version"}
	results := []ExecutionResult{
		{
			Module:     module,
			NewVersion: "1.0.0",
			Success:    true,
			Duration:   10 * time.Millisecond,
		},
	}

	output := formatter.FormatResults(results)
	// JSON formatter output should be valid JSON object
	if !strings.HasPrefix(output, "{") {
		t.Error("JSON formatter should output JSON object")
	}
}
