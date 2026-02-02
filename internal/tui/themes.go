package tui

import (
	"slices"

	"github.com/charmbracelet/huh"
)

// ValidThemes is the list of supported theme names.
var ValidThemes = []string{
	"sley",
	"base",
	"base16",
	"catppuccin",
	"charm",
	"dracula",
}

// IsValidTheme returns true if the given theme name is valid.
func IsValidTheme(name string) bool {
	return slices.Contains(ValidThemes, name)
}

// GetTheme returns the huh.Theme for the given theme name.
// Returns nil if the theme name is not recognized.
func GetTheme(name string) *huh.Theme {
	switch name {
	case "sley":
		return sleyTheme()
	case "base":
		return huh.ThemeBase()
	case "base16":
		return huh.ThemeBase16()
	case "catppuccin":
		return huh.ThemeCatppuccin()
	case "charm":
		return huh.ThemeCharm()
	case "dracula":
		return huh.ThemeDracula()
	default:
		return nil
	}
}
