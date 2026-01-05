package printer

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Style definitions for consistent console output across the application.
var (
	faintStyle   = lipgloss.NewStyle().Faint(true)
	boldStyle    = lipgloss.NewStyle().Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // Red
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // Cyan
)

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
