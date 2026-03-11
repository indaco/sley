package tui

import (
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

// Sley brand colors with adaptive light/dark support
var (
	// Primary teal - use darker shades on light bg, lighter on dark bg
	sleyTealPrimary = compat.AdaptiveColor{Light: lipgloss.Color("#0d9488"), Dark: lipgloss.Color("#14b8a6")}
	sleyTealBright  = compat.AdaptiveColor{Light: lipgloss.Color("#0f766e"), Dark: lipgloss.Color("#2dd4bf")}
	sleyTealAccent  = compat.AdaptiveColor{Light: lipgloss.Color("#115e59"), Dark: lipgloss.Color("#5eead4")}

	// Text colors - high contrast for readability
	sleyTextStrong = compat.AdaptiveColor{Light: lipgloss.Color("#0f172a"), Dark: lipgloss.Color("#f1f5f9")}
	sleyTextNormal = compat.AdaptiveColor{Light: lipgloss.Color("#334155"), Dark: lipgloss.Color("#cbd5e1")}
	sleyTextMuted  = compat.AdaptiveColor{Light: lipgloss.Color("#64748b"), Dark: lipgloss.Color("#94a3b8")}
	sleyTextFaint  = compat.AdaptiveColor{Light: lipgloss.Color("#94a3b8"), Dark: lipgloss.Color("#64748b")}

	// Borders and separators
	sleyBorderFocused = compat.AdaptiveColor{Light: lipgloss.Color("#0d9488"), Dark: lipgloss.Color("#14b8a6")}
	sleyBorderNormal  = compat.AdaptiveColor{Light: lipgloss.Color("#cbd5e1"), Dark: lipgloss.Color("#475569")}

	// Button backgrounds
	sleyButtonBg        = compat.AdaptiveColor{Light: lipgloss.Color("#0d9488"), Dark: lipgloss.Color("#14b8a6")}
	sleyButtonBgBlurred = compat.AdaptiveColor{Light: lipgloss.Color("#e2e8f0"), Dark: lipgloss.Color("#334155")}

	// Button text
	sleyButtonText        = compat.AdaptiveColor{Light: lipgloss.Color("#ffffff"), Dark: lipgloss.Color("#0f172a")}
	sleyButtonTextBlurred = compat.AdaptiveColor{Light: lipgloss.Color("#64748b"), Dark: lipgloss.Color("#94a3b8")}
)

// sleyTheme returns a huh theme with sley brand colors
func sleyTheme(isDark bool) *huh.Styles {
	t := huh.ThemeBase(isDark)

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
