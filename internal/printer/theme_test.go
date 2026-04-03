package printer

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestResolveTheme(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string defaults to sley", ""},
		{"sley theme", "sley"},
		{"dracula theme", "dracula"},
		{"catppuccin theme", "catppuccin"},
		{"base16 theme", "base16"},
		{"base theme", "base"},
		{"charm theme", "charm"},
		{"unknown falls back to sley", "nonexistent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := ResolveTheme(tt.input)
			// Theme should have a non-zero H1 style (has foreground color set).
			h1fg := theme.H1.GetForeground()
			if h1fg == (lipgloss.NoColor{}) {
				t.Errorf("ResolveTheme(%q) returned theme with no H1 foreground color", tt.input)
			}
		})
	}
}

func TestNewTypography(t *testing.T) {
	tests := []struct {
		name  string
		theme string
	}{
		{"default", ""},
		{"sley", "sley"},
		{"dracula", "dracula"},
		{"catppuccin", "catppuccin"},
		{"base16", "base16"},
		{"charm", "charm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ty := NewTypography(tt.theme)
			if ty == nil {
				t.Fatalf("NewTypography(%q) returned nil", tt.theme)
			}
			// H1 should produce non-empty output.
			out := ty.H1("test")
			if out == "" {
				t.Errorf("NewTypography(%q).H1(\"test\") returned empty string", tt.theme)
			}
		})
	}
}

func TestSleyTheme(t *testing.T) {
	theme := SleyTheme()
	// Verify key palette-derived styles are non-zero.
	if theme.H1.GetForeground() == (lipgloss.NoColor{}) {
		t.Error("SleyTheme() H1 has no foreground color")
	}
	if theme.Bold.GetBold() != true {
		t.Error("SleyTheme() Bold style is not bold")
	}
}
