// Package printer provides rich terminal styling for CLI output.
//
// This package uses lipgloss for consistent, attractive console output
// across the sley CLI. It provides both render functions (returning styled
// strings) and print functions (outputting to stdout).
//
// # Styling Functions
//
// Render functions return styled strings without printing:
//
//	styled := printer.Success("Operation completed")
//	styled := printer.Error("Something went wrong")
//	styled := printer.Warning("Deprecated feature")
//	styled := printer.Info("Processing...")
//	styled := printer.Bold("Important")
//	styled := printer.Faint("Secondary info")
//
// # Print Functions
//
// Print functions output styled text to stdout with a newline:
//
//	printer.PrintSuccess("Version bumped to 1.2.3")
//	printer.PrintError("Failed to read config")
//	printer.PrintWarning("No changelog entries found")
//	printer.PrintInfo("Checking dependencies...")
//
// # Badges
//
// Badge functions create bold, colored status indicators:
//
//	pass := printer.SuccessBadge("PASS")
//	fail := printer.ErrorBadge("FAIL")
//	warn := printer.WarningBadge("WARN")
//
// # Validation Formatting
//
// Specialized functions for doctor command validation output:
//
//	line := printer.FormatValidationPass("*", "PASS", "Config", "Valid YAML")
//	line := printer.FormatValidationFail("x", "FAIL", "Version", "Invalid format")
//	line := printer.FormatValidationWarn("!", "WARN", "Plugin", "Deprecated option")
//
// # Color Control
//
// Disable colors for non-TTY or user preference:
//
//	printer.SetNoColor(true)
//
// This also respects the NO_COLOR environment variable.
package printer
