// Package printer provides rich terminal styling for CLI output.
//
// This package uses herald (github.com/indaco/herald) for consistent,
// themed console output across the sley CLI. It provides both render
// functions (returning styled strings) and print functions (outputting
// to stdout).
//
// # Initialization
//
// Call Init with a theme name at startup to configure colors:
//
//	printer.Init("sley")   // sley brand teal palette
//	printer.Init("dracula") // dracula dark theme
//
// If Init is not called, the default sley theme is used.
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
//	line := printer.FormatValidationPass("Config", "Valid YAML")
//	line := printer.FormatValidationFail("Version", "Invalid format")
//	line := printer.FormatValidationWarn("Plugin", "Deprecated option")
//
// # Color Control
//
// Disable colors for non-TTY or user preference:
//
//	printer.SetNoColor(true)
//
// This also respects the NO_COLOR environment variable.
package printer
