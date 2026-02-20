package discover

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/printer"
)

// Formatter handles display of discovery results.
type Formatter struct {
	format OutputFormat
}

// NewFormatter creates a new Formatter with the specified output format.
func NewFormatter(format OutputFormat) *Formatter {
	return &Formatter{format: format}
}

// FormatResult formats the discovery result for display.
func (f *Formatter) FormatResult(result *discovery.Result) string {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(result)
	case FormatTable:
		return f.formatTable(result)
	default:
		return f.formatText(result)
	}
}

// formatText formats the result as human-readable text.
func (f *Formatter) formatText(result *discovery.Result) string {
	var sb strings.Builder

	// Header
	sb.WriteString("\n")
	sb.WriteString(printer.Info("Discovery Results"))
	sb.WriteString("\n")
	sb.WriteString(printer.Faint(strings.Repeat("-", 70)))
	sb.WriteString("\n")

	// Mode summary
	modeStr := getModeDescription(result.Mode)
	fmt.Fprintf(&sb, "Project Type: %s\n", printer.Bold(modeStr))
	sb.WriteString("\n")

	// Modules section
	if len(result.Modules) > 0 {
		sb.WriteString(printer.Info("Version Files (.version):"))
		sb.WriteString("\n")
		for _, m := range result.Modules {
			status := printer.Success("✓")
			fmt.Fprintf(&sb, "  %s %s %s\n", status, m.RelPath, printer.Faint(fmt.Sprintf("(%s)", m.Version)))
		}
		sb.WriteString("\n")
	}

	// Manifests section
	if len(result.Manifests) > 0 {
		sb.WriteString(printer.Info("Manifest Files:"))
		sb.WriteString("\n")
		for _, m := range result.Manifests {
			status := printer.Success("✓")
			fmt.Fprintf(&sb, "  %s %s %s\n", status, m.RelPath, printer.Faint(fmt.Sprintf("(%s: %s)", m.Description, m.Version)))
		}
		sb.WriteString("\n")
	}

	// Mismatches section
	if len(result.Mismatches) > 0 {
		sb.WriteString(printer.Warning("Version Mismatches:"))
		sb.WriteString("\n")
		for _, m := range result.Mismatches {
			status := printer.Warning("⚠")
			fmt.Fprintf(&sb, "  %s %s: expected %s, found %s\n",
				status, m.Source, m.ExpectedVersion, m.ActualVersion)
		}
		sb.WriteString("\n")
	}

	// Sync candidates section
	if len(result.SyncCandidates) > 0 && !result.HasModules() {
		sb.WriteString(printer.Info("Sync Candidates (for dependency-check plugin):"))
		sb.WriteString("\n")
		for _, c := range result.SyncCandidates {
			fmt.Fprintf(&sb, "  - %s %s\n", c.Path, printer.Faint(fmt.Sprintf("(%s)", c.Description)))
		}
		sb.WriteString("\n")
	}

	// Summary
	sb.WriteString(printer.Faint(strings.Repeat("-", 70)))
	sb.WriteString("\n")
	sb.WriteString(f.formatSummary(result))
	sb.WriteString("\n")

	return sb.String()
}

// formatTable formats the result as a table.
func (f *Formatter) formatTable(result *discovery.Result) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(printer.Info("Discovery Results"))
	sb.WriteString("\n\n")

	// Modules table
	if len(result.Modules) > 0 {
		sb.WriteString("Version Files:\n")
		fmt.Fprintf(&sb, "%-30s %-15s %-20s\n", "PATH", "VERSION", "MODULE")
		sb.WriteString(strings.Repeat("-", 65) + "\n")
		for _, m := range result.Modules {
			fmt.Fprintf(&sb, "%-30s %-15s %-20s\n", m.RelPath, m.Version, m.Name)
		}
		sb.WriteString("\n")
	}

	// Manifests table
	if len(result.Manifests) > 0 {
		sb.WriteString("Manifest Files:\n")
		fmt.Fprintf(&sb, "%-30s %-15s %-25s\n", "PATH", "VERSION", "TYPE")
		sb.WriteString(strings.Repeat("-", 70) + "\n")
		for _, m := range result.Manifests {
			fmt.Fprintf(&sb, "%-30s %-15s %-25s\n", m.RelPath, m.Version, m.Description)
		}
		sb.WriteString("\n")
	}

	// Mismatches table
	if len(result.Mismatches) > 0 {
		sb.WriteString("Version Mismatches:\n")
		fmt.Fprintf(&sb, "%-30s %-15s %-15s\n", "SOURCE", "EXPECTED", "ACTUAL")
		sb.WriteString(strings.Repeat("-", 60) + "\n")
		for _, m := range result.Mismatches {
			fmt.Fprintf(&sb, "%-30s %-15s %-15s\n", m.Source, m.ExpectedVersion, m.ActualVersion)
		}
		sb.WriteString("\n")
	}

	sb.WriteString(f.formatSummary(result))
	sb.WriteString("\n")

	return sb.String()
}

// formatJSON formats the result as JSON.
func (f *Formatter) formatJSON(result *discovery.Result) string {
	type jsonModule struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Version string `json:"version"`
	}

	type jsonManifest struct {
		Path        string `json:"path"`
		Filename    string `json:"filename"`
		Version     string `json:"version"`
		Format      string `json:"format"`
		Description string `json:"description"`
	}

	type jsonMismatch struct {
		Source   string `json:"source"`
		Expected string `json:"expected"`
		Actual   string `json:"actual"`
	}

	type jsonSyncCandidate struct {
		Path        string `json:"path"`
		Format      string `json:"format"`
		Field       string `json:"field,omitempty"`
		Pattern     string `json:"pattern,omitempty"`
		Description string `json:"description"`
	}

	output := struct {
		Mode           string              `json:"mode"`
		Modules        []jsonModule        `json:"modules"`
		Manifests      []jsonManifest      `json:"manifests"`
		Mismatches     []jsonMismatch      `json:"mismatches"`
		SyncCandidates []jsonSyncCandidate `json:"sync_candidates"`
		Summary        struct {
			ModuleCount         int    `json:"module_count"`
			ManifestCount       int    `json:"manifest_count"`
			MismatchCount       int    `json:"mismatch_count"`
			HasMismatches       bool   `json:"has_mismatches"`
			PrimaryVersion      string `json:"primary_version"`
			IsVersionConsistent bool   `json:"is_version_consistent"`
		} `json:"summary"`
	}{
		Mode:           result.Mode.String(),
		Modules:        make([]jsonModule, len(result.Modules)),
		Manifests:      make([]jsonManifest, len(result.Manifests)),
		Mismatches:     make([]jsonMismatch, len(result.Mismatches)),
		SyncCandidates: make([]jsonSyncCandidate, len(result.SyncCandidates)),
	}

	for i, m := range result.Modules {
		output.Modules[i] = jsonModule{
			Name:    m.Name,
			Path:    m.RelPath,
			Version: m.Version,
		}
	}

	for i, m := range result.Manifests {
		output.Manifests[i] = jsonManifest{
			Path:        m.RelPath,
			Filename:    m.Filename,
			Version:     m.Version,
			Format:      m.Format.String(),
			Description: m.Description,
		}
	}

	for i, m := range result.Mismatches {
		output.Mismatches[i] = jsonMismatch{
			Source:   m.Source,
			Expected: m.ExpectedVersion,
			Actual:   m.ActualVersion,
		}
	}

	for i, c := range result.SyncCandidates {
		output.SyncCandidates[i] = jsonSyncCandidate{
			Path:        c.Path,
			Format:      c.Format.String(),
			Field:       c.Field,
			Pattern:     c.Pattern,
			Description: c.Description,
		}
	}

	output.Summary.ModuleCount = len(result.Modules)
	output.Summary.ManifestCount = len(result.Manifests)
	output.Summary.MismatchCount = len(result.Mismatches)
	output.Summary.HasMismatches = result.HasMismatches()
	output.Summary.PrimaryVersion = result.PrimaryVersion()
	output.Summary.IsVersionConsistent = discovery.IsVersionConsistent(result)

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		return ""
	}

	return string(data)
}

// formatSummary returns a summary line for the result.
func (f *Formatter) formatSummary(result *discovery.Result) string {
	moduleCount := len(result.Modules)
	manifestCount := len(result.Manifests)
	mismatchCount := len(result.Mismatches)

	parts := []string{}

	if moduleCount > 0 {
		parts = append(parts, fmt.Sprintf("%d version file(s)", moduleCount))
	}

	if manifestCount > 0 {
		parts = append(parts, fmt.Sprintf("%d manifest(s)", manifestCount))
	}

	if mismatchCount > 0 {
		parts = append(parts, printer.Warning(fmt.Sprintf("%d mismatch(es)", mismatchCount)))
	}

	if len(parts) == 0 {
		return printer.Faint("No version sources found")
	}

	summary := "Found: " + strings.Join(parts, ", ")

	if result.PrimaryVersion() != "" {
		summary += fmt.Sprintf(" | Primary version: %s", printer.Bold(result.PrimaryVersion()))
	}

	return summary
}

// getModeDescription returns a human-readable description of the detection mode.
func getModeDescription(mode discovery.DetectionMode) string {
	switch mode {
	case discovery.SingleModule:
		return "Single Module"
	case discovery.MultiModule:
		return "Multi-Module (Monorepo)"
	case discovery.NoModules:
		return "No .version files found"
	default:
		return "Unknown"
	}
}

// PrintResult prints the formatted result to stdout.
func (f *Formatter) PrintResult(result *discovery.Result) {
	fmt.Print(f.FormatResult(result))
}
