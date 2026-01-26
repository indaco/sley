package discover

import (
	"github.com/charmbracelet/huh"
	"github.com/indaco/sley/internal/tui"
)

// Prompter abstracts interactive prompts for testability.
type Prompter interface {
	Confirm(title, description string) (bool, error)
	MultiSelect(title, description string, options []huh.Option[string], defaults []string) ([]string, error)
	Select(title, description string, options []huh.Option[string]) (string, error)
}

// TUIPrompter implements Prompter using the tui package.
type TUIPrompter struct{}

// NewPrompter creates a new TUIPrompter.
func NewPrompter() Prompter {
	return &TUIPrompter{}
}

// Confirm shows a yes/no confirmation prompt.
func (p *TUIPrompter) Confirm(title, description string) (bool, error) {
	return tui.Confirm(title, description)
}

// MultiSelect shows a multi-select prompt.
func (p *TUIPrompter) MultiSelect(title, description string, options []huh.Option[string], defaults []string) ([]string, error) {
	return tui.MultiSelect(title, description, options, defaults)
}

// Select shows a single-select prompt.
func (p *TUIPrompter) Select(title, description string, options []huh.Option[string]) (string, error) {
	return tui.Select(title, description, options)
}

// OutputFormat controls how discovery results are displayed.
type OutputFormat string

const (
	// FormatText outputs human-readable text.
	FormatText OutputFormat = "text"

	// FormatJSON outputs machine-readable JSON.
	FormatJSON OutputFormat = "json"

	// FormatTable outputs tabular data.
	FormatTable OutputFormat = "table"
)

// ParseOutputFormat converts a string to OutputFormat.
func ParseOutputFormat(s string) OutputFormat {
	switch s {
	case "json":
		return FormatJSON
	case "table":
		return FormatTable
	default:
		return FormatText
	}
}
