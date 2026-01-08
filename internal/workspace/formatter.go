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

// formatTextVersionInfo returns the version info suffix for a result.
func formatTextVersionInfo(result ExecutionResult) string {
	switch {
	case result.OldVersion != "" && result.NewVersion != "" && result.OldVersion != result.NewVersion:
		return fmt.Sprintf(": %s -> %s (%s)", result.OldVersion, result.NewVersion, formatDuration(result.Duration))
	case result.NewVersion != "":
		return fmt.Sprintf(": %s (%s)", result.NewVersion, formatDuration(result.Duration))
	default:
		return fmt.Sprintf(" (%s)", formatDuration(result.Duration))
	}
}

// writeTextResultLine writes a single result line to the builder.
func writeTextResultLine(sb *strings.Builder, result ExecutionResult) {
	if result.Success {
		fmt.Fprintf(sb, "  %s %s", printer.SuccessBadge("✓"), result.Module.Name)
		sb.WriteString(printer.Faint(formatTextVersionInfo(result)))
	} else {
		fmt.Fprintf(sb, "  %s %s: ", printer.ErrorBadge("✗"), result.Module.Name)
		sb.WriteString(printer.Faint(result.Error.Error()))
	}
	sb.WriteString("\n")
}

// writeTextSummary writes the summary line to the builder.
func (f *TextFormatter) writeTextSummary(sb *strings.Builder, results []ExecutionResult, successCount int) {
	sb.WriteString("\n")
	if successCount == len(results) {
		msg := fmt.Sprintf("Success: %d module%s %s in %s",
			len(results), pluralize(len(results)), f.actionVerb, formatDuration(TotalDuration(results)))
		sb.WriteString(printer.Success(msg))
	} else {
		msg := fmt.Sprintf("Completed: %d succeeded, %d failed", successCount, len(results)-successCount)
		sb.WriteString(printer.Warning(msg))
	}
	sb.WriteString("\n")
}

// FormatResults formats execution results as text.
func (f *TextFormatter) FormatResults(results []ExecutionResult) string {
	if len(results) == 0 {
		return "No results to display."
	}

	var sb strings.Builder
	if f.operation != "" {
		sb.WriteString(fmt.Sprintf("%s\n", f.operation))
	}

	successCount := 0
	for _, result := range results {
		writeTextResultLine(&sb, result)
		if result.Success {
			successCount++
		}
	}

	f.writeTextSummary(&sb, results, successCount)
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

// formatResultVersion returns the display version for a result.
func formatResultVersion(result ExecutionResult) string {
	if result.OldVersion != "" && result.NewVersion != "" && result.OldVersion != result.NewVersion {
		return fmt.Sprintf("%s -> %s", result.OldVersion, result.NewVersion)
	}
	return result.NewVersion
}

// buildTableDivider creates a table divider line for the given column widths.
func buildTableDivider(widths ...int) string {
	var sb strings.Builder
	for _, w := range widths {
		sb.WriteString("+" + strings.Repeat("-", w+2))
	}
	sb.WriteString("+\n")
	return sb.String()
}

// calculateResultColumnWidths computes column widths for result table formatting.
func calculateResultColumnWidths(results []ExecutionResult) (name, version, status, duration int) {
	name, version, status, duration = len("Module"), len("Version"), len("Status"), len("Duration")
	for _, result := range results {
		if len(result.Module.Name) > name {
			name = len(result.Module.Name)
		}
		if v := formatResultVersion(result); len(v) > version {
			version = len(v)
		}
	}
	return name + 2, version + 2, status + 2, duration + 2
}

// FormatResults formats execution results as a table.
func (f *TableFormatter) FormatResults(results []ExecutionResult) string {
	if len(results) == 0 {
		return "No results to display."
	}

	var sb strings.Builder
	if f.operation != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", f.operation))
	}

	nameW, versionW, statusW, durationW := calculateResultColumnWidths(results)
	headerFmt := fmt.Sprintf("| %%-%ds | %%-%ds | %%-%ds | %%-%ds |\n", nameW, versionW, statusW, durationW)
	divider := buildTableDivider(nameW, versionW, statusW, durationW)

	sb.WriteString(divider)
	sb.WriteString(fmt.Sprintf(headerFmt, "Module", "Version", "Status", "Duration"))
	sb.WriteString(divider)

	successCount := 0
	for _, result := range results {
		status := "FAILED"
		if result.Success {
			status = "OK"
			successCount++
		}
		sb.WriteString(fmt.Sprintf(headerFmt, result.Module.Name, formatResultVersion(result), status, formatDuration(result.Duration)))
	}
	sb.WriteString(divider)

	sb.WriteString("\n")
	f.writeResultSummary(&sb, results, successCount)
	return sb.String()
}

// writeResultSummary appends a summary line to the builder.
func (f *TableFormatter) writeResultSummary(sb *strings.Builder, results []ExecutionResult, successCount int) {
	total := len(results)
	if successCount == total {
		fmt.Fprintf(sb, "Success: %d module%s %s in %s\n",
			total, pluralize(total), f.actionVerb, formatDuration(TotalDuration(results)))
	} else {
		fmt.Fprintf(sb, "Completed: %d succeeded, %d failed\n", successCount, total-successCount)
	}
}

// moduleDisplayPath returns the path to display for a module.
func moduleDisplayPath(mod *Module) string {
	if mod.RelPath != "" {
		return mod.RelPath
	}
	return mod.Dir
}

// truncatePath shortens a path if it exceeds maxWidth.
func truncatePath(path string, maxWidth int) string {
	if len(path) > maxWidth {
		return "..." + path[len(path)-maxWidth+3:]
	}
	return path
}

// calculateModuleColumnWidths computes column widths for module table formatting.
func calculateModuleColumnWidths(modules []*Module) (name, version, path int) {
	const maxPathWidth = 50
	name, version, path = len("Module"), len("Version"), len("Path")
	for _, mod := range modules {
		if len(mod.Name) > name {
			name = len(mod.Name)
		}
		if len(mod.CurrentVersion) > version {
			version = len(mod.CurrentVersion)
		}
		if p := moduleDisplayPath(mod); len(p) > path {
			path = len(p)
		}
	}
	name += 2
	version += 2
	path += 2
	if path > maxPathWidth {
		path = maxPathWidth
	}
	return name, version, path
}

// FormatModuleList formats a list of modules as a table.
func (f *TableFormatter) FormatModuleList(modules []*Module) string {
	if len(modules) == 0 {
		return "No modules found."
	}

	var sb strings.Builder
	nameW, versionW, pathW := calculateModuleColumnWidths(modules)
	headerFmt := fmt.Sprintf("| %%-%ds | %%-%ds | %%-%ds |\n", nameW, versionW, pathW)
	divider := buildTableDivider(nameW, versionW, pathW)

	sb.WriteString(fmt.Sprintf("Found %d module%s:\n\n", len(modules), pluralize(len(modules))))
	sb.WriteString(divider)
	sb.WriteString(fmt.Sprintf(headerFmt, "Module", "Version", "Path"))
	sb.WriteString(divider)

	for _, mod := range modules {
		version := mod.CurrentVersion
		if version == "" {
			version = "-"
		}
		sb.WriteString(fmt.Sprintf(headerFmt, mod.Name, version, truncatePath(moduleDisplayPath(mod), pathW)))
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
