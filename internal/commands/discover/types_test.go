package discover

import (
	"testing"
)

func TestNewPrompter(t *testing.T) {
	prompter := NewPrompter()

	if prompter == nil {
		t.Fatal("expected non-nil prompter")
	}

	// Verify it returns a TUIPrompter
	_, ok := prompter.(*TUIPrompter)
	if !ok {
		t.Error("expected TUIPrompter type")
	}
}

func TestTUIPrompter_Interface(t *testing.T) {
	// Verify TUIPrompter implements Prompter interface
	var _ Prompter = &TUIPrompter{}
}

func TestOutputFormat_Constants(t *testing.T) {
	if FormatText != "text" {
		t.Errorf("FormatText = %q, want %q", FormatText, "text")
	}
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}
	if FormatTable != "table" {
		t.Errorf("FormatTable = %q, want %q", FormatTable, "table")
	}
}

func TestParseOutputFormat_AllCases(t *testing.T) {
	tests := []struct {
		input    string
		expected OutputFormat
	}{
		{"text", FormatText},
		{"json", FormatJSON},
		{"table", FormatTable},
		{"", FormatText},
		{"invalid", FormatText},
		{"unknown", FormatText},
		{"  ", FormatText}, // whitespace
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseOutputFormat(tt.input)
			if result != tt.expected {
				t.Errorf("ParseOutputFormat(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
