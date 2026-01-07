// Package console provides simple colored console output utilities.
//
// This package offers basic ANSI color support for terminal output with
// success (green) and failure (red) message formatting. It supports
// disabling colors via the NoColor flag for non-TTY environments or
// user preference.
//
// # Usage
//
// Print colored messages to stdout:
//
//	console.PrintSuccess("Version bumped successfully")
//	console.PrintFailure("Failed to read version file")
//
// # Color Control
//
// Disable colors programmatically:
//
//	console.SetNoColor(true)
//
// When NoColor is true, messages are printed without ANSI escape codes.
//
// # Note
//
// For more sophisticated terminal styling with lipgloss support, use the
// printer package instead. This package is intended for simple, lightweight
// color output without external dependencies beyond the standard library.
package console
