package printer

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// TestRenderFunctions verifies that all render functions return non-empty styled strings.
func TestRenderFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(string) string
		input    string
	}{
		{"Faint", Faint, "test text"},
		{"Bold", Bold, "test text"},
		{"Success", Success, "test text"},
		{"Error", Error, "test text"},
		{"Warning", Warning, "test text"},
		{"Info", Info, "test text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)

			// Verify result is not empty
			if result == "" {
				t.Errorf("%s() returned empty string", tt.name)
			}

			// Verify result contains the original text
			// The styled output may or may not contain ANSI codes depending on terminal detection,
			// but it should at minimum contain the original text
			if !strings.Contains(result, tt.input) {
				t.Errorf("%s() result does not contain input text. got %q, want to contain %q", tt.name, result, tt.input)
			}
		})
	}
}

// TestPrintFunctions verifies that print functions output to stdout.
func TestPrintFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(string)
		input    string
	}{
		{"PrintFaint", PrintFaint, "test text"},
		{"PrintBold", PrintBold, "test text"},
		{"PrintSuccess", PrintSuccess, "test text"},
		{"PrintError", PrintError, "test text"},
		{"PrintWarning", PrintWarning, "test text"},
		{"PrintInfo", PrintInfo, "test text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the print function
			tt.function(tt.input)

			// Restore stdout and read the captured output
			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			// Verify output is not empty
			if output == "" {
				t.Errorf("%s() produced no output", tt.name)
			}

			// Verify output contains the original text
			if !strings.Contains(output, tt.input) {
				t.Errorf("%s() output does not contain input text. got %q, want to contain %q", tt.name, output, tt.input)
			}

			// Verify output ends with newline
			if !strings.HasSuffix(output, "\n") {
				t.Errorf("%s() output does not end with newline", tt.name)
			}
		})
	}
}

// TestEmptyInput verifies that functions handle empty strings gracefully.
func TestEmptyInput(t *testing.T) {
	tests := []struct {
		name     string
		function func(string) string
	}{
		{"Faint", Faint},
		{"Bold", Bold},
		{"Success", Success},
		{"Error", Error},
		{"Warning", Warning},
		{"Info", Info},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function("")

			// Even with empty input, the function should return a string
			// (it may contain ANSI codes even for empty text)
			if result == "" {
				// This is acceptable - empty input may yield empty output
				return
			}
		})
	}
}

// TestBadgeFunctions verifies that badge functions return styled strings.
func TestBadgeFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(string) string
		input    string
	}{
		{"SuccessBadge", SuccessBadge, "[PASS]"},
		{"ErrorBadge", ErrorBadge, "[FAIL]"},
		{"WarningBadge", WarningBadge, "[WARN]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)

			// Verify result is not empty
			if result == "" {
				t.Errorf("%s() returned empty string", tt.name)
			}

			// Verify result contains the original text
			if !strings.Contains(result, tt.input) {
				t.Errorf("%s() result does not contain input text. got %q, want to contain %q", tt.name, result, tt.input)
			}
		})
	}
}

// TestValidationFormatFunctions verifies validation formatting functions.
func TestValidationFormatFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(string, string, string, string) string
		symbol   string
		badge    string
		category string
		message  string
	}{
		{
			name:     "FormatValidationPass",
			function: FormatValidationPass,
			symbol:   "✓",
			badge:    "[PASS]",
			category: "YAML Syntax",
			message:  "Configuration file is valid YAML",
		},
		{
			name:     "FormatValidationFail",
			function: FormatValidationFail,
			symbol:   "✗",
			badge:    "[FAIL]",
			category: "Plugin: audit-log",
			message:  "Invalid format 'xml': must be 'json' or 'yaml'",
		},
		{
			name:     "FormatValidationWarn",
			function: FormatValidationWarn,
			symbol:   "⚠",
			badge:    "[WARN]",
			category: "Plugin: tag-manager",
			message:  "Tag prefix 'v' is valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.symbol, tt.badge, tt.category, tt.message)

			// Verify result is not empty
			if result == "" {
				t.Errorf("%s() returned empty string", tt.name)
			}

			// Verify result contains all components
			if !strings.Contains(result, tt.symbol) {
				t.Errorf("%s() result does not contain symbol. got %q, want to contain %q", tt.name, result, tt.symbol)
			}
			if !strings.Contains(result, tt.badge) {
				t.Errorf("%s() result does not contain badge. got %q, want to contain %q", tt.name, result, tt.badge)
			}
			if !strings.Contains(result, tt.category) {
				t.Errorf("%s() result does not contain category. got %q, want to contain %q", tt.name, result, tt.category)
			}
			if !strings.Contains(result, tt.message) {
				t.Errorf("%s() result does not contain message. got %q, want to contain %q", tt.name, result, tt.message)
			}

			// Verify the format follows the expected pattern: "symbol badge category: message"
			// We check that category comes before message
			categoryIndex := strings.Index(result, tt.category)
			messageIndex := strings.Index(result, tt.message)
			if categoryIndex == -1 || messageIndex == -1 {
				t.Fatalf("%s() missing category or message in result", tt.name)
			}
			if categoryIndex >= messageIndex {
				t.Errorf("%s() category should come before message", tt.name)
			}
		})
	}
}
