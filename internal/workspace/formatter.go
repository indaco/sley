package workspace

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/indaco/sley/internal/printer"
)

// OutputFormatter formats execution results for display.
type OutputFormatter interface {
	// FormatResults formats a slice of execution results into a string.
	FormatResults(results []ExecutionResult) string

	// FormatModuleList formats a list of modules for display.
	FormatModuleList(modules []*Module) string
}

// TextFormatter formats output as human-readable text.
type TextFormatter struct {
	// Operation name for display (e.g., "Version Summary")
	operation string
	// Action verb for success message (e.g., "validated", "updated")
	actionVerb string
}

// NewTextFormatter creates a new text formatter with default "updated" action verb.
func NewTextFormatter(operation string) *TextFormatter {
	return &TextFormatter{
		operation:  operation,
		actionVerb: "updated",
	}
}

// NewTextFormatterWithVerb creates a new text formatter with a custom action verb.
func NewTextFormatterWithVerb(operation, actionVerb string) *TextFormatter {
	return &TextFormatter{
		operation:  operation,
		actionVerb: actionVerb,
	}
}

// FormatResults formats execution results as text.
func (f *TextFormatter) FormatResults(results []ExecutionResult) string {
	if len(results) == 0 {
		return "No results to display."
	}

	var sb strings.Builder

	// Header
	if f.operation != "" {
		sb.WriteString(fmt.Sprintf("%s\n", f.operation))
	}

	// Individual results
	successCount := 0
	for _, result := range results {
		if result.Success {
			// Bold green checkmark
			sb.WriteString(fmt.Sprintf("  %s %s", printer.SuccessBadge("✓"), result.Module.Name))

			// Version info in faint
			var versionInfo string
			switch {
			case result.OldVersion != "" && result.NewVersion != "" && result.OldVersion != result.NewVersion:
				versionInfo = fmt.Sprintf(": %s -> %s (%s)", result.OldVersion, result.NewVersion, formatDuration(result.Duration))
			case result.NewVersion != "":
				versionInfo = fmt.Sprintf(": %s (%s)", result.NewVersion, formatDuration(result.Duration))
			default:
				versionInfo = fmt.Sprintf(" (%s)", formatDuration(result.Duration))
			}
			sb.WriteString(printer.Faint(versionInfo))
			sb.WriteString("\n")
			successCount++
		} else {
			// Bold red X mark
			sb.WriteString(fmt.Sprintf("  %s %s: ", printer.ErrorBadge("✗"), result.Module.Name))
			sb.WriteString(printer.Faint(result.Error.Error()))
			sb.WriteString("\n")
		}
	}

	// Summary
	sb.WriteString("\n")
	if successCount == len(results) {
		totalDuration := TotalDuration(results)
		msg := fmt.Sprintf("Success: %d module%s %s in %s",
			len(results),
			pluralize(len(results)),
			f.actionVerb,
			formatDuration(totalDuration),
		)
		sb.WriteString(printer.Success(msg))
		sb.WriteString("\n")
	} else {
		errorCount := len(results) - successCount
		msg := fmt.Sprintf("Completed: %d succeeded, %d failed",
			successCount,
			errorCount,
		)
		sb.WriteString(printer.Warning(msg))
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatModuleList formats a list of modules as text.
func (f *TextFormatter) FormatModuleList(modules []*Module) string {
	if len(modules) == 0 {
		return "No modules found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d module%s:\n", len(modules), pluralize(len(modules))))

	for _, mod := range modules {
		sb.WriteString(fmt.Sprintf("  • %s", mod.Name))
		if mod.CurrentVersion != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", mod.CurrentVersion))
		}
		if mod.RelPath != "" {
			sb.WriteString(fmt.Sprintf(" - %s", mod.RelPath))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// JSONFormatter formats output as JSON.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// resultJSON is the JSON representation of an execution result.
type resultJSON struct {
	Module     string `json:"module"`
	Path       string `json:"path"`
	OldVersion string `json:"old_version,omitempty"`
	NewVersion string `json:"new_version,omitempty"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	Duration   string `json:"duration"`
}

// resultsJSON is the JSON representation of all results.
type resultsJSON struct {
	Results       []resultJSON `json:"results"`
	Total         int          `json:"total"`
	SuccessCount  int          `json:"success_count"`
	ErrorCount    int          `json:"error_count"`
	TotalDuration string       `json:"total_duration"`
}

// FormatResults formats execution results as JSON.
func (f *JSONFormatter) FormatResults(results []ExecutionResult) string {
	jsonResults := make([]resultJSON, len(results))
	successCount := 0

	for i, result := range results {
		r := resultJSON{
			Module:     result.Module.Name,
			Path:       result.Module.Path,
			OldVersion: result.OldVersion,
			NewVersion: result.NewVersion,
			Success:    result.Success,
			Duration:   result.Duration.String(),
		}
		if result.Error != nil {
			r.Error = result.Error.Error()
		}
		jsonResults[i] = r

		if result.Success {
			successCount++
		}
	}

	output := resultsJSON{
		Results:       jsonResults,
		Total:         len(results),
		SuccessCount:  successCount,
		ErrorCount:    len(results) - successCount,
		TotalDuration: TotalDuration(results).String(),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal JSON: %s"}`, err.Error())
	}

	return string(data)
}

// moduleJSON is the JSON representation of a module.
type moduleJSON struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	RelPath        string `json:"rel_path,omitempty"`
	CurrentVersion string `json:"current_version,omitempty"`
	Dir            string `json:"dir"`
}

// modulesJSON is the JSON representation of a module list.
type modulesJSON struct {
	Modules []moduleJSON `json:"modules"`
	Total   int          `json:"total"`
}

// FormatModuleList formats a list of modules as JSON.
func (f *JSONFormatter) FormatModuleList(modules []*Module) string {
	jsonModules := make([]moduleJSON, len(modules))

	for i, mod := range modules {
		jsonModules[i] = moduleJSON{
			Name:           mod.Name,
			Path:           mod.Path,
			RelPath:        mod.RelPath,
			CurrentVersion: mod.CurrentVersion,
			Dir:            mod.Dir,
		}
	}

	output := modulesJSON{
		Modules: jsonModules,
		Total:   len(modules),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal JSON: %s"}`, err.Error())
	}

	return string(data)
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// pluralize returns "s" if count != 1, empty string otherwise.
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// TableFormatter formats output as an ASCII table.
type TableFormatter struct {
	// Operation name for display
	operation string
	// Action verb for success message (e.g., "validated", "updated")
	actionVerb string
}

// NewTableFormatter creates a new table formatter with default "updated" action verb.
func NewTableFormatter(operation string) *TableFormatter {
	return &TableFormatter{
		operation:  operation,
		actionVerb: "updated",
	}
}

// NewTableFormatterWithVerb creates a new table formatter with a custom action verb.
func NewTableFormatterWithVerb(operation, actionVerb string) *TableFormatter {
	return &TableFormatter{
		operation:  operation,
		actionVerb: actionVerb,
	}
}

// FormatResults formats execution results as a table.
func (f *TableFormatter) FormatResults(results []ExecutionResult) string {
	if len(results) == 0 {
		return "No results to display."
	}

	var sb strings.Builder

	// Header
	if f.operation != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", f.operation))
	}

	// Calculate column widths
	nameWidth := len("Module")
	versionWidth := len("Version")
	statusWidth := len("Status")
	durationWidth := len("Duration")

	for _, result := range results {
		if len(result.Module.Name) > nameWidth {
			nameWidth = len(result.Module.Name)
		}
		version := result.NewVersion
		if result.OldVersion != "" && result.NewVersion != "" && result.OldVersion != result.NewVersion {
			version = fmt.Sprintf("%s -> %s", result.OldVersion, result.NewVersion)
		}
		if len(version) > versionWidth {
			versionWidth = len(version)
		}
	}

	// Add padding
	nameWidth += 2
	versionWidth += 2
	statusWidth += 2
	durationWidth += 2

	// Draw header
	headerFormat := fmt.Sprintf("| %%-%ds | %%-%ds | %%-%ds | %%-%ds |\n",
		nameWidth, versionWidth, statusWidth, durationWidth)
	divider := "+" + strings.Repeat("-", nameWidth+2) +
		"+" + strings.Repeat("-", versionWidth+2) +
		"+" + strings.Repeat("-", statusWidth+2) +
		"+" + strings.Repeat("-", durationWidth+2) + "+\n"

	sb.WriteString(divider)
	sb.WriteString(fmt.Sprintf(headerFormat, "Module", "Version", "Status", "Duration"))
	sb.WriteString(divider)

	// Draw rows
	successCount := 0
	for _, result := range results {
		version := result.NewVersion
		if result.OldVersion != "" && result.NewVersion != "" && result.OldVersion != result.NewVersion {
			version = fmt.Sprintf("%s -> %s", result.OldVersion, result.NewVersion)
		}

		status := "OK"
		if !result.Success {
			status = "FAILED"
		} else {
			successCount++
		}

		sb.WriteString(fmt.Sprintf(headerFormat,
			result.Module.Name,
			version,
			status,
			formatDuration(result.Duration),
		))
	}
	sb.WriteString(divider)

	// Summary
	sb.WriteString("\n")
	if successCount == len(results) {
		totalDuration := TotalDuration(results)
		sb.WriteString(fmt.Sprintf("Success: %d module%s %s in %s\n",
			len(results),
			pluralize(len(results)),
			f.actionVerb,
			formatDuration(totalDuration),
		))
	} else {
		errorCount := len(results) - successCount
		sb.WriteString(fmt.Sprintf("Completed: %d succeeded, %d failed\n",
			successCount,
			errorCount,
		))
	}

	return sb.String()
}

// FormatModuleList formats a list of modules as a table.
func (f *TableFormatter) FormatModuleList(modules []*Module) string {
	if len(modules) == 0 {
		return "No modules found."
	}

	var sb strings.Builder

	// Calculate column widths
	nameWidth := len("Module")
	versionWidth := len("Version")
	pathWidth := len("Path")

	for _, mod := range modules {
		if len(mod.Name) > nameWidth {
			nameWidth = len(mod.Name)
		}
		if len(mod.CurrentVersion) > versionWidth {
			versionWidth = len(mod.CurrentVersion)
		}
		displayPath := mod.RelPath
		if displayPath == "" {
			displayPath = mod.Dir
		}
		if len(displayPath) > pathWidth {
			pathWidth = len(displayPath)
		}
	}

	// Add padding
	nameWidth += 2
	versionWidth += 2
	pathWidth += 2

	// Limit path width to reasonable size
	maxPathWidth := 50
	if pathWidth > maxPathWidth {
		pathWidth = maxPathWidth
	}

	// Draw header
	headerFormat := fmt.Sprintf("| %%-%ds | %%-%ds | %%-%ds |\n",
		nameWidth, versionWidth, pathWidth)
	divider := "+" + strings.Repeat("-", nameWidth+2) +
		"+" + strings.Repeat("-", versionWidth+2) +
		"+" + strings.Repeat("-", pathWidth+2) + "+\n"

	sb.WriteString(fmt.Sprintf("Found %d module%s:\n\n", len(modules), pluralize(len(modules))))
	sb.WriteString(divider)
	sb.WriteString(fmt.Sprintf(headerFormat, "Module", "Version", "Path"))
	sb.WriteString(divider)

	// Draw rows
	for _, mod := range modules {
		displayPath := mod.RelPath
		if displayPath == "" {
			displayPath = mod.Dir
		}
		// Truncate long paths
		if len(displayPath) > pathWidth {
			displayPath = "..." + displayPath[len(displayPath)-pathWidth+3:]
		}

		version := mod.CurrentVersion
		if version == "" {
			version = "-"
		}

		sb.WriteString(fmt.Sprintf(headerFormat, mod.Name, version, displayPath))
	}
	sb.WriteString(divider)

	return sb.String()
}

// GetFormatter returns the appropriate formatter based on the format string.
// Uses "updated" as the default action verb.
func GetFormatter(format string, operation string) OutputFormatter {
	return GetFormatterWithVerb(format, operation, "updated")
}

// GetFormatterWithVerb returns the appropriate formatter with a custom action verb.
// Use this for read-only operations like "show" (checked) or "doctor" (validated).
func GetFormatterWithVerb(format, operation, actionVerb string) OutputFormatter {
	switch format {
	case "json":
		return NewJSONFormatter()
	case "table":
		return NewTableFormatterWithVerb(operation, actionVerb)
	default:
		return NewTextFormatterWithVerb(operation, actionVerb)
	}
}
