package tui

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Sley brand colors with adaptive light/dark support
var (
	// Primary teal - use darker shades on light bg, lighter on dark bg
	sleyTealPrimary = lipgloss.AdaptiveColor{Light: "#0d9488", Dark: "#14b8a6"}
	sleyTealBright  = lipgloss.AdaptiveColor{Light: "#0f766e", Dark: "#2dd4bf"}
	sleyTealAccent  = lipgloss.AdaptiveColor{Light: "#115e59", Dark: "#5eead4"}

	// Text colors - high contrast for readability
	sleyTextStrong = lipgloss.AdaptiveColor{Light: "#0f172a", Dark: "#f1f5f9"}
	sleyTextNormal = lipgloss.AdaptiveColor{Light: "#334155", Dark: "#cbd5e1"}
	sleyTextMuted  = lipgloss.AdaptiveColor{Light: "#64748b", Dark: "#94a3b8"}
	sleyTextFaint  = lipgloss.AdaptiveColor{Light: "#94a3b8", Dark: "#64748b"}

	// Borders and separators
	sleyBorderFocused = lipgloss.AdaptiveColor{Light: "#0d9488", Dark: "#14b8a6"}
	sleyBorderNormal  = lipgloss.AdaptiveColor{Light: "#cbd5e1", Dark: "#475569"}

	// Button backgrounds
	sleyButtonBg        = lipgloss.AdaptiveColor{Light: "#0d9488", Dark: "#14b8a6"}
	sleyButtonBgBlurred = lipgloss.AdaptiveColor{Light: "#e2e8f0", Dark: "#334155"}

	// Button text
	sleyButtonText        = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#0f172a"}
	sleyButtonTextBlurred = lipgloss.AdaptiveColor{Light: "#64748b", Dark: "#94a3b8"}
)

// sleyTheme returns a huh theme with sley brand colors
func sleyTheme() *huh.Theme {
	t := huh.ThemeBase()

	// Focused state styles
	t.Focused.Title = t.Focused.Title.
		Foreground(sleyTealBright).
		Bold(true)

	t.Focused.Description = t.Focused.Description.
		Foreground(sleyTextMuted)

	t.Focused.Base = t.Focused.Base.
		BorderForeground(sleyBorderFocused).
		BorderStyle(lipgloss.RoundedBorder())

	t.Focused.SelectedOption = t.Focused.SelectedOption.
		Foreground(sleyTealBright)

	t.Focused.SelectSelector = t.Focused.SelectSelector.
		Foreground(sleyTealPrimary)

	t.Focused.Option = t.Focused.Option.
		Foreground(sleyTextNormal)

	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(sleyButtonText).
		Background(sleyButtonBg).
		Bold(true).
		Padding(0, 1)

	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(sleyButtonTextBlurred).
		Background(sleyButtonBgBlurred).
		Padding(0, 1)

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.
		Foreground(sleyTealAccent)

	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.
		Foreground(sleyTextFaint)

	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.
		Foreground(sleyTealPrimary)

	// Blurred state styles
	t.Blurred.Title = t.Blurred.Title.
		Foreground(sleyTextStrong)

	t.Blurred.Description = t.Blurred.Description.
		Foreground(sleyTextFaint)

	t.Blurred.Base = t.Blurred.Base.
		BorderForeground(sleyBorderNormal)

	t.Blurred.SelectedOption = t.Blurred.SelectedOption.
		Foreground(sleyTextMuted)

	t.Blurred.Option = t.Blurred.Option.
		Foreground(sleyTextFaint)

	// Help styles
	t.Help.ShortKey = t.Help.ShortKey.
		Foreground(sleyTealPrimary)

	t.Help.ShortDesc = t.Help.ShortDesc.
		Foreground(sleyTextFaint)

	t.Help.ShortSeparator = t.Help.ShortSeparator.
		Foreground(sleyBorderNormal)

	t.Help.FullKey = t.Help.FullKey.
		Foreground(sleyTealPrimary)

	t.Help.FullDesc = t.Help.FullDesc.
		Foreground(sleyTextMuted)

	t.Help.FullSeparator = t.Help.FullSeparator.
		Foreground(sleyBorderNormal)

	return t
}
