package printer

import (
	"fmt"
	"os"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/indaco/herald"
)

var (
	mu sync.RWMutex
	ty *herald.Typography

	// Standalone styles for semantic text coloring.
	// These use the SemanticPalette from the active theme so that
	// Success/Error/Warning/Info render in the expected colors without
	// needing a full herald element (H1, P, Badge, etc.).
	successStyle lipgloss.Style
	errorStyle   lipgloss.Style
	warningStyle lipgloss.Style
	infoStyle    lipgloss.Style
)

// Init sets the package-level Typography from the named theme.
// Call once at startup (e.g. in cli.go Before hook).
func Init(theme string) {
	mu.Lock()
	defer mu.Unlock()
	ty = NewTypography(theme)
	rebuildSemanticStyles()
}

// typography returns the package-level Typography, lazily initializing
// with the default sley theme if Init was never called.
func typography() *herald.Typography {
	mu.RLock()
	t := ty
	mu.RUnlock()
	if t != nil {
		return t
	}
	// Lazy init with default theme.
	Init("")
	mu.RLock()
	defer mu.RUnlock()
	return ty
}

// Typography returns the package-level herald Typography instance.
// Use this when you need direct access to herald elements like H1, H2, HR, etc.
func Typography() *herald.Typography {
	return typography()
}

// rebuildSemanticStyles derives colored text styles from the active theme's
// semantic badge styles. We extract the background color of each badge
// (which is the semantic color) and use it as a foreground color for text.
func rebuildSemanticStyles() {
	th := ty.Theme()
	successStyle = lipgloss.NewStyle().Foreground(th.SuccessBadge.GetBackground())
	errorStyle = lipgloss.NewStyle().Foreground(th.ErrorBadge.GetBackground())
	warningStyle = lipgloss.NewStyle().Foreground(th.WarningBadge.GetBackground())
	infoStyle = lipgloss.NewStyle().Foreground(th.InfoBadge.GetBackground())
}

// SetNoColor controls whether the printer uses colors.
// This respects the --no-color flag and NO_COLOR environment variable.
func SetNoColor(disabled bool) {
	if disabled || os.Getenv("NO_COLOR") != "" {
		lipgloss.Writer.Profile = colorprofile.ASCII
	}
}

// ---------------------------------------------------------------------------
// Render functions - return styled strings without printing
// ---------------------------------------------------------------------------

// Faint returns text with faint styling.
func Faint(text string) string {
	return typography().Small(text)
}

// Bold returns text with bold styling.
func Bold(text string) string {
	return typography().Bold(text)
}

// Success returns text with success (green) styling.
func Success(text string) string {
	mu.RLock()
	defer mu.RUnlock()
	return lipgloss.Sprint(successStyle.Render(text))
}

// Error returns text with error (red) styling.
func Error(text string) string {
	mu.RLock()
	defer mu.RUnlock()
	return lipgloss.Sprint(errorStyle.Render(text))
}

// Warning returns text with warning (yellow) styling.
func Warning(text string) string {
	mu.RLock()
	defer mu.RUnlock()
	return lipgloss.Sprint(warningStyle.Render(text))
}

// Info returns text with info (cyan) styling.
func Info(text string) string {
	mu.RLock()
	defer mu.RUnlock()
	return lipgloss.Sprint(infoStyle.Render(text))
}

// ---------------------------------------------------------------------------
// Print functions - output styled text to stdout with a newline
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Badge functions
// ---------------------------------------------------------------------------

// SuccessBadge returns a bold, green styled badge.
func SuccessBadge(text string) string {
	return typography().SuccessBadge(text)
}

// ErrorBadge returns a bold, red styled badge.
func ErrorBadge(text string) string {
	return typography().ErrorBadge(text)
}

// WarningBadge returns a bold, yellow styled badge.
func WarningBadge(text string) string {
	return typography().WarningBadge(text)
}

// ---------------------------------------------------------------------------
// Validation format functions
// ---------------------------------------------------------------------------

// FormatValidationPass formats a validation result with PASS status.
// Uses a subtle green tag, category in normal text, message in faint.
func FormatValidationPass(category, message string) string {
	t := typography()
	return fmt.Sprintf("%s %s: %s", t.SuccessTag("PASS"), category, t.Small(message))
}

// FormatValidationFail formats a validation result with FAIL status.
// Uses a subtle red tag, category in normal text, message in faint.
func FormatValidationFail(category, message string) string {
	t := typography()
	return fmt.Sprintf("%s %s: %s", t.ErrorTag("FAIL"), category, t.Small(message))
}

// FormatValidationWarn formats a validation result with WARN status.
// Uses a subtle yellow tag, category in normal text, message in faint.
func FormatValidationWarn(category, message string) string {
	t := typography()
	return fmt.Sprintf("%s %s: %s", t.WarningTag("WARN"), category, t.Small(message))
}

// FormatValidationFaint formats a validation result with all faint styling.
// Used for disabled or inactive items.
func FormatValidationFaint(category, message string) string {
	t := typography()
	return fmt.Sprintf("%s %s: %s", t.Small("[OFF]"), t.Small(category), t.Small(message))
}
