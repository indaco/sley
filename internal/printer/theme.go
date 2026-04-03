package printer

import (
	"os"

	"charm.land/lipgloss/v2"
	"github.com/indaco/herald"
)

// SleyTheme returns a herald Theme using the sley brand teal palette
// with adaptive light/dark support.
func SleyTheme() herald.Theme {
	return herald.ThemeFromPalette(sleyPalette())
}

// sleyPalette maps sley brand colors to herald's ColorPalette.
// Colors match the TUI theme in tui/theme_sley.go.
func sleyPalette() herald.ColorPalette {
	lightDark := lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))

	return herald.ColorPalette{
		Primary:   lightDark(lipgloss.Color("#0d9488"), lipgloss.Color("#14b8a6")), // sley teal
		Secondary: lightDark(lipgloss.Color("#0f766e"), lipgloss.Color("#2dd4bf")), // sley teal bright
		Tertiary:  lightDark(lipgloss.Color("#115e59"), lipgloss.Color("#5eead4")), // sley teal accent
		Accent:    lightDark(lipgloss.Color("#b07d2b"), lipgloss.Color("#f6c177")), // warning/amber
		Highlight: lightDark(lipgloss.Color("#c44040"), lipgloss.Color("#ea9a97")), // error/red
		Muted:     lightDark(lipgloss.Color("#64748b"), lipgloss.Color("#94a3b8")), // sley text muted
		Text:      lightDark(lipgloss.Color("#334155"), lipgloss.Color("#cbd5e1")), // sley text normal
		Surface:   lightDark(lipgloss.Color("#e2e8f0"), lipgloss.Color("#334155")), // sley button bg blurred
		Base:      lightDark(lipgloss.Color("#f1f5f9"), lipgloss.Color("#0f172a")), // sley text strong (inv)
	}
}

// ResolveTheme returns the herald Theme for the given name.
// Unknown names silently fall back to SleyTheme.
func ResolveTheme(name string) herald.Theme {
	switch name {
	case "", "sley":
		return SleyTheme()
	case "dracula":
		return herald.DraculaTheme()
	case "catppuccin":
		return herald.CatppuccinTheme()
	case "base", "base16":
		return herald.Base16Theme()
	case "charm":
		return herald.CharmTheme()
	default:
		return SleyTheme()
	}
}

// NewTypography creates a herald Typography with the named theme.
func NewTypography(theme string) *herald.Typography {
	return herald.New(herald.WithTheme(ResolveTheme(theme)))
}
