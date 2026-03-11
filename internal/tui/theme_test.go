package tui

import (
	"testing"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

func TestSleyTheme(t *testing.T) {
	theme := sleyTheme(true)

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
		color compat.AdaptiveColor
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
			if tc.color.Light == nil {
				t.Errorf("%s should have a light color defined", tc.name)
			}
		})

		t.Run(tc.name+" has dark color", func(t *testing.T) {
			if tc.color.Dark == nil {
				t.Errorf("%s should have a dark color defined", tc.name)
			}
		})
	}
}

func TestSleyThemeConsistency(t *testing.T) {
	theme := sleyTheme(true)

	t.Run("Focused and blurred buttons have same padding", func(t *testing.T) {
		_, fRight, _, fLeft := theme.Focused.FocusedButton.GetPadding()
		_, bRight, _, bLeft := theme.Focused.BlurredButton.GetPadding()

		if fLeft != bLeft || fRight != bRight {
			t.Error("FocusedButton and BlurredButton should have consistent padding")
		}
	})
}
