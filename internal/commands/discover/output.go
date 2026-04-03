package discover

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/printer"
)

// Formatter handles display of discovery results.
type Formatter struct {
	format OutputFormat
	cfg    *config.Config
}

// NewFormatter creates a new Formatter with the specified output format.
func NewFormatter(format OutputFormat) *Formatter {
	return &Formatter{format: format}
}

// NewFormatterWithConfig creates a new Formatter with the specified output format and config.
// The config is used to adjust output severity (e.g., independent versioning shows info instead of warnings).
func NewFormatterWithConfig(format OutputFormat, cfg *config.Config) *Formatter {
	return &Formatter{format: format, cfg: cfg}
}

// isIndependentVersioning returns true if the config indicates independent versioning mode.
func (f *Formatter) isIndependentVersioning() bool {
	return f.cfg != nil && f.cfg.Workspace != nil && f.cfg.Workspace.IsIndependentVersioning()
}

// versioningMode returns the effective versioning mode string from the config.
func (f *Formatter) versioningMode() string {
	if f.cfg != nil && f.cfg.Workspace != nil {
		return f.cfg.Workspace.VersioningMode()
	}
	return "coordinated"
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
	ty := printer.Typography()
	var blocks []string

	// Header + mode
	blocks = append(blocks, ty.H2("Discovery Results"))
	if result.Mode != discovery.NoModules {
		blocks = append(blocks, ty.KV("Project Type", printer.Bold(getModeDescription(result.Mode))))
	}

	// Modules section
	if len(result.Modules) > 0 {
		items := make([]string, len(result.Modules))
		for i, m := range result.Modules {
			items[i] = fmt.Sprintf("%s %s %s", printer.Success("✓"), m.RelPath, printer.Faint(fmt.Sprintf("(%s)", m.Version)))
		}
		blocks = append(blocks, ty.Section(ty.H4("Version Files (.version)"), ty.UL(items...)))
	}

	// Manifests section
	if len(result.Manifests) > 0 {
		items := make([]string, len(result.Manifests))
		for i, m := range result.Manifests {
			items[i] = fmt.Sprintf("%s %s %s", printer.Success("✓"), m.RelPath, printer.Faint(fmt.Sprintf("(%s: %s)", m.Description, m.Version)))
		}
		blocks = append(blocks, ty.Section(ty.H4("Manifest Files"), ty.UL(items...)))
	}

	// Mismatches section
	if len(result.Mismatches) > 0 {
		if f.isIndependentVersioning() {
			items := make([]string, len(result.Mismatches))
			for i, m := range result.Mismatches {
				items[i] = fmt.Sprintf("%s %s: %s (root: %s)", printer.Info("ℹ"), m.Source, m.ActualVersion, m.ExpectedVersion)
			}
			blocks = append(blocks, ty.Section(
				ty.H4(fmt.Sprintf("Version Summary (independent versioning): %d module(s) at different versions", len(result.Mismatches))),
				ty.UL(items...),
			))
		} else {
			items := make([]string, len(result.Mismatches))
			for i, m := range result.Mismatches {
				items[i] = fmt.Sprintf("%s %s: expected %s, found %s", printer.Warning("⚠"), m.Source, m.ExpectedVersion, m.ActualVersion)
			}
			blocks = append(blocks, ty.Section(ty.H4("Version Mismatches"), ty.UL(items...)))
		}
	}

	// Sync candidates section
	if len(result.SyncCandidates) > 0 && !result.HasModules() {
		items := make([]string, len(result.SyncCandidates))
		for i, c := range result.SyncCandidates {
			items[i] = fmt.Sprintf("%s %s", c.Path, printer.Faint(fmt.Sprintf("(%s)", c.Description)))
		}
		blocks = append(blocks, ty.Section(ty.H4("Sync Candidates (for dependency-check plugin)"), ty.UL(items...)))
	}

	// Summary
	blocks = append(blocks, f.formatSummary(result))

	return ty.Compose(blocks...)
}

// formatTable formats the result as a table.
func (f *Formatter) formatTable(result *discovery.Result) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(printer.Typography().H2("Discovery Results"))
	sb.WriteString("\n")

	// Modules table
	if len(result.Modules) > 0 {
		sb.WriteString("Version Files:\n")
		fmt.Fprintf(&sb, "%-30s %-15s %-20s\n", "PATH", "VERSION", "MODULE")
		sb.WriteString(printer.Typography().HR() + "\n")
		for _, m := range result.Modules {
			fmt.Fprintf(&sb, "%-30s %-15s %-20s\n", m.RelPath, m.Version, m.Name)
		}
		sb.WriteString("\n")
	}

	// Manifests table
	if len(result.Manifests) > 0 {
		sb.WriteString("Manifest Files:\n")
		fmt.Fprintf(&sb, "%-30s %-15s %-25s\n", "PATH", "VERSION", "TYPE")
		sb.WriteString(printer.Typography().HR() + "\n")
		for _, m := range result.Manifests {
			fmt.Fprintf(&sb, "%-30s %-15s %-25s\n", m.RelPath, m.Version, m.Description)
		}
		sb.WriteString("\n")
	}

	// Mismatches table
	if len(result.Mismatches) > 0 {
		if f.isIndependentVersioning() {
			fmt.Fprintf(&sb, "Version Summary (independent versioning): %d module(s) at different versions\n", len(result.Mismatches))
		} else {
			sb.WriteString("Version Mismatches:\n")
		}
		fmt.Fprintf(&sb, "%-30s %-15s %-15s\n", "SOURCE", "EXPECTED", "ACTUAL")
		sb.WriteString(printer.Typography().HR() + "\n")
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
		VersioningMode string              `json:"versioning_mode"`
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
		VersioningMode: f.versioningMode(),
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
	ty := printer.Typography()
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
		if f.isIndependentVersioning() {
			parts = append(parts, printer.Info(fmt.Sprintf("%d version difference(s) (independent)", mismatchCount)))
		} else {
			parts = append(parts, printer.Warning(fmt.Sprintf("%d mismatch(es)", mismatchCount)))
		}
	}

	if len(parts) == 0 {
		return ty.Small("No version sources found")
	}

	pairs := [][2]string{
		{"Found", strings.Join(parts, ", ")},
	}
	if result.PrimaryVersion() != "" {
		pairs = append(pairs, [2]string{"Primary version", ty.Bold(result.PrimaryVersion())})
	}

	return ty.KVGroup(pairs)
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
	fmt.Println(f.FormatResult(result))
}
