package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestSleyTheme(t *testing.T) {
	theme := sleyTheme()

	if theme == nil {
		t.Fatal("SleyTheme() returned nil")
	}

	t.Run("Focused styles are configured", func(t *testing.T) {
		// Title should be bold
		if !theme.Focused.Title.GetBold() {
			t.Error("Focused.Title should be bold")
		}

		// Base should have rounded border
		if theme.Focused.Base.GetBorderStyle() != lipgloss.RoundedBorder() {
			t.Error("Focused.Base should have rounded border")
		}

		// FocusedButton should be bold with padding
		if !theme.Focused.FocusedButton.GetBold() {
			t.Error("Focused.FocusedButton should be bold")
		}

		// FocusedButton should have horizontal padding
		_, right, _, left := theme.Focused.FocusedButton.GetPadding()
		if left != 1 || right != 1 {
			t.Errorf("Focused.FocusedButton should have horizontal padding of 1, got left=%d right=%d", left, right)
		}
	})

	t.Run("Blurred styles are configured", func(t *testing.T) {
		// BlurredButton should have padding matching FocusedButton
		_, right, _, left := theme.Focused.BlurredButton.GetPadding()
		if left != 1 || right != 1 {
			t.Errorf("Focused.BlurredButton should have horizontal padding of 1, got left=%d right=%d", left, right)
		}
	})

	t.Run("Help styles are configured", func(t *testing.T) {
		// Verify help styles can render text (styles are configured)
		shortKeyRendered := theme.Help.ShortKey.Render("key")
		shortDescRendered := theme.Help.ShortDesc.Render("description")
		shortSepRendered := theme.Help.ShortSeparator.Render(" - ")
		fullKeyRendered := theme.Help.FullKey.Render("key")
		fullDescRendered := theme.Help.FullDesc.Render("description")
		fullSepRendered := theme.Help.FullSeparator.Render(" - ")

		// Verify all help styles can render non-empty output
		if shortKeyRendered == "" {
			t.Error("Help.ShortKey should render non-empty output")
		}
		if shortDescRendered == "" {
			t.Error("Help.ShortDesc should render non-empty output")
		}
		if shortSepRendered == "" {
			t.Error("Help.ShortSeparator should render non-empty output")
		}
		if fullKeyRendered == "" {
			t.Error("Help.FullKey should render non-empty output")
		}
		if fullDescRendered == "" {
			t.Error("Help.FullDesc should render non-empty output")
		}
		if fullSepRendered == "" {
			t.Error("Help.FullSeparator should render non-empty output")
		}
	})
}

func TestSleyThemeColors(t *testing.T) {
	// Verify adaptive colors are properly defined
	testCases := []struct {
		name  string
		color lipgloss.AdaptiveColor
	}{
		{"sleyTealPrimary", sleyTealPrimary},
		{"sleyTealBright", sleyTealBright},
		{"sleyTealAccent", sleyTealAccent},
		{"sleyTextStrong", sleyTextStrong},
		{"sleyTextNormal", sleyTextNormal},
		{"sleyTextMuted", sleyTextMuted},
		{"sleyTextFaint", sleyTextFaint},
		{"sleyBorderFocused", sleyBorderFocused},
		{"sleyBorderNormal", sleyBorderNormal},
		{"sleyButtonBg", sleyButtonBg},
		{"sleyButtonBgBlurred", sleyButtonBgBlurred},
		{"sleyButtonText", sleyButtonText},
		{"sleyButtonTextBlurred", sleyButtonTextBlurred},
	}

	for _, tc := range testCases {
		t.Run(tc.name+" has light color", func(t *testing.T) {
			if tc.color.Light == "" {
				t.Errorf("%s should have a light color defined", tc.name)
			}
		})

		t.Run(tc.name+" has dark color", func(t *testing.T) {
			if tc.color.Dark == "" {
				t.Errorf("%s should have a dark color defined", tc.name)
			}
		})

		t.Run(tc.name+" has valid hex colors", func(t *testing.T) {
			if !isValidHexColor(tc.color.Light) {
				t.Errorf("%s light color %q is not a valid hex color", tc.name, tc.color.Light)
			}
			if !isValidHexColor(tc.color.Dark) {
				t.Errorf("%s dark color %q is not a valid hex color", tc.name, tc.color.Dark)
			}
		})
	}
}

func TestSleyThemeConsistency(t *testing.T) {
	theme := sleyTheme()

	t.Run("Focused and blurred buttons have same padding", func(t *testing.T) {
		_, fRight, _, fLeft := theme.Focused.FocusedButton.GetPadding()
		_, bRight, _, bLeft := theme.Focused.BlurredButton.GetPadding()

		if fLeft != bLeft || fRight != bRight {
			t.Error("FocusedButton and BlurredButton should have consistent padding")
		}
	})
}

// isValidHexColor checks if a string is a valid hex color (e.g., "#0d9488")
func isValidHexColor(s string) bool {
	if len(s) != 7 {
		return false
	}
	if s[0] != '#' {
		return false
	}
	for _, c := range s[1:] {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
