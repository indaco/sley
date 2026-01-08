package printer

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	// noColor is set by SetNoColor and controls whether colors are rendered.
	noColor bool

	// Style definitions for consistent console output across the application.
	faintStyle   = lipgloss.NewStyle().Faint(true)
	boldStyle    = lipgloss.NewStyle().Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // Red
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // Cyan

	// Combined styles for status badges
	successBadgeStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")) // Bold green
	errorBadgeStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")) // Bold red
	warningBadgeStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3")) // Bold yellow
)

// SetNoColor controls whether the printer uses colors.
// This respects the --no-color flag and NO_COLOR environment variable.
func SetNoColor(disabled bool) {
	noColor = disabled
	if noColor || os.Getenv("NO_COLOR") != "" {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
}

// Render functions return styled strings without printing.

// Faint returns text with faint styling.
func Faint(text string) string {
	return faintStyle.Render(text)
}

// Bold returns text with bold styling.
func Bold(text string) string {
	return boldStyle.Render(text)
}

// Success returns text with success (green) styling.
func Success(text string) string {
	return successStyle.Render(text)
}

// Error returns text with error (red) styling.
func Error(text string) string {
	return errorStyle.Render(text)
}

// Warning returns text with warning (yellow) styling.
func Warning(text string) string {
	return warningStyle.Render(text)
}

// Info returns text with info (cyan) styling.
func Info(text string) string {
	return infoStyle.Render(text)
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
	return successBadgeStyle.Render(text)
}

// ErrorBadge returns a bold, red styled badge.
func ErrorBadge(text string) string {
	return errorBadgeStyle.Render(text)
}

// WarningBadge returns a bold, yellow styled badge.
func WarningBadge(text string) string {
	return warningBadgeStyle.Render(text)
}

// FormatValidationPass formats a validation result with PASS status.
// Symbol and badge are bold green, category is normal, message is faint.
func FormatValidationPass(symbol, badge, category, message string) string {
	styledSymbol := successBadgeStyle.Render(symbol)
	styledBadge := successBadgeStyle.Render(badge)
	styledMessage := faintStyle.Render(message)
	return fmt.Sprintf("%s %s %s: %s", styledSymbol, styledBadge, category, styledMessage)
}

// FormatValidationFail formats a validation result with FAIL status.
// Symbol and badge are bold red, category is normal, message is faint.
func FormatValidationFail(symbol, badge, category, message string) string {
	styledSymbol := errorBadgeStyle.Render(symbol)
	styledBadge := errorBadgeStyle.Render(badge)
	styledMessage := faintStyle.Render(message)
	return fmt.Sprintf("%s %s %s: %s", styledSymbol, styledBadge, category, styledMessage)
}

// FormatValidationWarn formats a validation result with WARN status.
// Symbol and badge are bold yellow, category is normal, message is faint.
func FormatValidationWarn(symbol, badge, category, message string) string {
	styledSymbol := warningBadgeStyle.Render(symbol)
	styledBadge := warningBadgeStyle.Render(badge)
	styledMessage := faintStyle.Render(message)
	return fmt.Sprintf("%s %s %s: %s", styledSymbol, styledBadge, category, styledMessage)
}

// FormatValidationFaint formats a validation result with all faint styling.
// Used for disabled or inactive items.
func FormatValidationFaint(symbol, badge, category, message string) string {
	styledSymbol := faintStyle.Render(symbol)
	styledBadge := faintStyle.Render(badge)
	styledCategory := faintStyle.Render(category)
	styledMessage := faintStyle.Render(message)
	return fmt.Sprintf("%s %s %s: %s", styledSymbol, styledBadge, styledCategory, styledMessage)
}
