package printer

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

var (
	// Style definitions for consistent console output across the application.
	faintStyle   = lipgloss.NewStyle().Faint(true)
	boldStyle    = lipgloss.NewStyle().Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(2)) // Green
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(1)) // Red
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(3)) // Yellow
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(6)) // Cyan

	// Combined styles for status badges
	successBadgeStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(2)) // Bold green
	errorBadgeStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(1)) // Bold red
	warningBadgeStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.ANSIColor(3)) // Bold yellow
)

// SetNoColor controls whether the printer uses colors.
// This respects the --no-color flag and NO_COLOR environment variable.
func SetNoColor(disabled bool) {
	if disabled || os.Getenv("NO_COLOR") != "" {
		lipgloss.Writer.Profile = colorprofile.ASCII
	}
}

// render applies a style and downsamples ANSI codes through lipgloss.Writer
// so the output respects the detected color profile (e.g., no colors in non-TTY).
func render(style lipgloss.Style, text string) string {
	return lipgloss.Sprint(style.Render(text))
}

// Render functions return styled strings without printing.

// Faint returns text with faint styling.
func Faint(text string) string {
	return render(faintStyle, text)
}

// Bold returns text with bold styling.
func Bold(text string) string {
	return render(boldStyle, text)
}

// Success returns text with success (green) styling.
func Success(text string) string {
	return render(successStyle, text)
}

// Error returns text with error (red) styling.
func Error(text string) string {
	return render(errorStyle, text)
}

// Warning returns text with warning (yellow) styling.
func Warning(text string) string {
	return render(warningStyle, text)
}

// Info returns text with info (cyan) styling.
func Info(text string) string {
	return render(infoStyle, text)
}

// Print functions output styled text to stdout with a newline.

// PrintFaint prints text with faint styling.
func PrintFaint(text string) {
	fmt.Println(Faint(text))
}

// PrintBold prints text with bold styling.
func PrintBold(text string) {
	fmt.Println(Bold(text))
}

// PrintSuccess prints text with success (green) styling.
func PrintSuccess(text string) {
	fmt.Println(Success(text))
}

// PrintError prints text with error (red) styling.
func PrintError(text string) {
	fmt.Println(Error(text))
}

// PrintWarning prints text with warning (yellow) styling.
func PrintWarning(text string) {
	fmt.Println(Warning(text))
}

// PrintInfo prints text with info (cyan) styling.
func PrintInfo(text string) {
	fmt.Println(Info(text))
}

// SuccessBadge returns a bold, green styled badge.
func SuccessBadge(text string) string {
	return render(successBadgeStyle, text)
}

// ErrorBadge returns a bold, red styled badge.
func ErrorBadge(text string) string {
	return render(errorBadgeStyle, text)
}

// WarningBadge returns a bold, yellow styled badge.
func WarningBadge(text string) string {
	return render(warningBadgeStyle, text)
}

// FormatValidationPass formats a validation result with PASS status.
// Symbol and badge are bold green, category is normal, message is faint.
func FormatValidationPass(symbol, badge, category, message string) string {
	styledSymbol := render(successBadgeStyle, symbol)
	styledBadge := render(successBadgeStyle, badge)
	styledMessage := render(faintStyle, message)
	return fmt.Sprintf("%s %s %s: %s", styledSymbol, styledBadge, category, styledMessage)
}

// FormatValidationFail formats a validation result with FAIL status.
// Symbol and badge are bold red, category is normal, message is faint.
func FormatValidationFail(symbol, badge, category, message string) string {
	styledSymbol := render(errorBadgeStyle, symbol)
	styledBadge := render(errorBadgeStyle, badge)
	styledMessage := render(faintStyle, message)
	return fmt.Sprintf("%s %s %s: %s", styledSymbol, styledBadge, category, styledMessage)
}

// FormatValidationWarn formats a validation result with WARN status.
// Symbol and badge are bold yellow, category is normal, message is faint.
func FormatValidationWarn(symbol, badge, category, message string) string {
	styledSymbol := render(warningBadgeStyle, symbol)
	styledBadge := render(warningBadgeStyle, badge)
	styledMessage := render(faintStyle, message)
	return fmt.Sprintf("%s %s %s: %s", styledSymbol, styledBadge, category, styledMessage)
}

// FormatValidationFaint formats a validation result with all faint styling.
// Used for disabled or inactive items.
func FormatValidationFaint(symbol, badge, category, message string) string {
	styledSymbol := render(faintStyle, symbol)
	styledBadge := render(faintStyle, badge)
	styledCategory := render(faintStyle, category)
	styledMessage := render(faintStyle, message)
	return fmt.Sprintf("%s %s %s: %s", styledSymbol, styledBadge, styledCategory, styledMessage)
}
