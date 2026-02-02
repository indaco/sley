package tui

import (
	"github.com/charmbracelet/huh"
)

// currentTheme holds the currently configured theme for TUI components.
// When nil, currentThemeOrDefault() returns the default sleyTheme.
var currentTheme *huh.Theme

// SetTheme sets the current theme by name.
// If the name is invalid or empty, the sley theme is used.
func SetTheme(name string) {
	if name == "" {
		currentTheme = nil
		return
	}
	theme := GetTheme(name)
	if theme != nil {
		currentTheme = theme
	} else {
		// Fall back to sley theme for invalid names
		currentTheme = nil
	}
}

// currentThemeOrDefault returns the current theme for TUI components.
// Returns the sley theme if no theme has been set.
func currentThemeOrDefault() *huh.Theme {
	if currentTheme == nil {
		return sleyTheme()
	}
	return currentTheme
}

// resetTheme resets the current theme to the default (sley).
// This is primarily useful for testing.
func resetTheme() {
	currentTheme = nil
}
