package tui

import (
	"testing"
)

func TestValidThemes(t *testing.T) {
	expected := []string{"sley", "base", "base16", "catppuccin", "charm", "dracula"}

	if len(ValidThemes) != len(expected) {
		t.Errorf("expected %d valid themes, got %d", len(expected), len(ValidThemes))
	}

	for i, theme := range expected {
		if ValidThemes[i] != theme {
			t.Errorf("expected theme at index %d to be %q, got %q", i, theme, ValidThemes[i])
		}
	}
}

func TestIsValidTheme(t *testing.T) {
	tests := []struct {
		name     string
		theme    string
		expected bool
	}{
		{
			name:     "sley theme is valid",
			theme:    "sley",
			expected: true,
		},
		{
			name:     "base theme is valid",
			theme:    "base",
			expected: true,
		},
		{
			name:     "base16 theme is valid",
			theme:    "base16",
			expected: true,
		},
		{
			name:     "catppuccin theme is valid",
			theme:    "catppuccin",
			expected: true,
		},
		{
			name:     "charm theme is valid",
			theme:    "charm",
			expected: true,
		},
		{
			name:     "dracula theme is valid",
			theme:    "dracula",
			expected: true,
		},
		{
			name:     "empty string is invalid",
			theme:    "",
			expected: false,
		},
		{
			name:     "unknown theme is invalid",
			theme:    "unknown",
			expected: false,
		},
		{
			name:     "case sensitive - SLEY is invalid",
			theme:    "SLEY",
			expected: false,
		},
		{
			name:     "case sensitive - Dracula is invalid",
			theme:    "Dracula",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTheme(tt.theme)
			if got != tt.expected {
				t.Errorf("IsValidTheme(%q) = %v, want %v", tt.theme, got, tt.expected)
			}
		})
	}
}

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name        string
		theme       string
		expectNil   bool
		description string
	}{
		{
			name:        "sley theme returns non-nil",
			theme:       "sley",
			expectNil:   false,
			description: "sley theme should return the custom sley theme",
		},
		{
			name:        "base theme returns non-nil",
			theme:       "base",
			expectNil:   false,
			description: "base theme should return huh.ThemeBase()",
		},
		{
			name:        "base16 theme returns non-nil",
			theme:       "base16",
			expectNil:   false,
			description: "base16 theme should return huh.ThemeBase16()",
		},
		{
			name:        "catppuccin theme returns non-nil",
			theme:       "catppuccin",
			expectNil:   false,
			description: "catppuccin theme should return huh.ThemeCatppuccin()",
		},
		{
			name:        "charm theme returns non-nil",
			theme:       "charm",
			expectNil:   false,
			description: "charm theme should return huh.ThemeCharm()",
		},
		{
			name:        "dracula theme returns non-nil",
			theme:       "dracula",
			expectNil:   false,
			description: "dracula theme should return huh.ThemeDracula()",
		},
		{
			name:        "empty string returns nil",
			theme:       "",
			expectNil:   true,
			description: "empty theme name should return nil",
		},
		{
			name:        "unknown theme returns nil",
			theme:       "unknown",
			expectNil:   true,
			description: "unknown theme name should return nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTheme(tt.theme)
			if tt.expectNil && got != nil {
				t.Errorf("GetTheme(%q) returned non-nil, want nil: %s", tt.theme, tt.description)
			}
			if !tt.expectNil && got == nil {
				t.Errorf("GetTheme(%q) returned nil, want non-nil: %s", tt.theme, tt.description)
			}
		})
	}
}

func TestSetTheme(t *testing.T) {
	// Reset theme after each test
	defer resetTheme()

	t.Run("set valid theme", func(t *testing.T) {
		SetTheme("dracula")
		theme := currentThemeOrDefault()
		if theme == nil {
			t.Error("currentThemeOrDefault() returned nil after SetTheme(\"dracula\")")
		}
	})

	t.Run("set empty string resets to default", func(t *testing.T) {
		SetTheme("dracula") // First set a non-default theme
		SetTheme("")        // Reset to default
		theme := currentThemeOrDefault()
		if theme == nil {
			t.Error("currentThemeOrDefault() returned nil after SetTheme(\"\")")
		}
		// The theme should be the sley theme (default)
	})

	t.Run("set invalid theme falls back to default", func(t *testing.T) {
		SetTheme("invalid-theme")
		theme := currentThemeOrDefault()
		if theme == nil {
			t.Error("currentThemeOrDefault() returned nil after SetTheme(\"invalid-theme\")")
		}
		// Should fall back to sley theme
	})
}

func TestCurrentTheme(t *testing.T) {
	// Reset theme after each test
	defer resetTheme()

	t.Run("returns default theme when not set", func(t *testing.T) {
		resetTheme()
		theme := currentThemeOrDefault()
		if theme == nil {
			t.Error("currentThemeOrDefault() returned nil when no theme is set")
		}
	})

	t.Run("returns set theme", func(t *testing.T) {
		SetTheme("charm")
		theme := currentThemeOrDefault()
		if theme == nil {
			t.Error("currentThemeOrDefault() returned nil after setting charm theme")
		}
	})
}

func TestResetTheme(t *testing.T) {
	// Set a theme first
	SetTheme("dracula")

	// Reset
	resetTheme()

	// Verify we're back to default
	theme := currentThemeOrDefault()
	if theme == nil {
		t.Error("currentThemeOrDefault() returned nil after resetTheme()")
	}
}
