package initcmd

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
)

// VersionSource represents a detected version from an existing file.
type VersionSource struct {
	// File is the source file path.
	File string

	// Version is the extracted version string.
	Version string

	// Format describes the file format (e.g., "package.json", "Cargo.toml").
	Format string
}

// DetectExistingVersions searches for version information in common project files.
// Returns all detected versions, allowing the user to choose which to use.
func DetectExistingVersions() []VersionSource {
	var sources []VersionSource

	// Check each supported file type
	detectors := []struct {
		file     string
		format   string
		detector func(string) (string, error)
	}{
		{"package.json", "Node.js (package.json)", detectPackageJSONVersion},
		{"Cargo.toml", "Rust (Cargo.toml)", detectCargoVersion},
		{"pyproject.toml", "Python (pyproject.toml)", detectPyprojectVersion},
		{"Chart.yaml", "Helm (Chart.yaml)", detectChartVersion},
		{"version.txt", "Plain text (version.txt)", detectPlainTextVersion},
		{"VERSION", "Plain text (VERSION)", detectPlainTextVersion},
	}

	for _, d := range detectors {
		if version, err := d.detector(d.file); err == nil && version != "" {
			if isValidSemver(version) {
				sources = append(sources, VersionSource{
					File:    d.file,
					Version: version,
					Format:  d.format,
				})
			}
		}
	}

	return sources
}

// detectPackageJSONVersion extracts version from package.json.
func detectPackageJSONVersion(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var pkg struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", err
	}

	return pkg.Version, nil
}

// detectCargoVersion extracts version from Cargo.toml.
func detectCargoVersion(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Simple regex to extract version from [package] section
	// Cargo.toml format: version = "x.y.z"
	re := regexp.MustCompile(`(?m)^\s*version\s*=\s*"([^"]+)"`)
	matches := re.FindSubmatch(data)
	if len(matches) >= 2 {
		return string(matches[1]), nil
	}

	return "", nil
}

// detectPyprojectVersion extracts version from pyproject.toml.
func detectPyprojectVersion(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Check for [project] section version (PEP 621)
	// or [tool.poetry] section version
	patterns := []string{
		`(?m)^\s*version\s*=\s*"([^"]+)"`,
		`(?m)^\s*version\s*=\s*'([^']+)'`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindSubmatch(data)
		if len(matches) >= 2 {
			return string(matches[1]), nil
		}
	}

	return "", nil
}

// detectChartVersion extracts version from Chart.yaml (Helm).
func detectChartVersion(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var chart struct {
		Version string `yaml:"version"`
	}

	if err := yaml.Unmarshal(data, &chart); err != nil {
		return "", err
	}

	return chart.Version, nil
}

// detectPlainTextVersion reads version from a plain text file.
func detectPlainTextVersion(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(data))
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	return version, nil
}

// isValidSemver performs a basic check if the string looks like a semver version.
func isValidSemver(version string) bool {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Basic semver pattern: x.y.z with optional pre-release and build metadata
	re := regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`)
	return re.MatchString(version)
}

// GetBestVersionSource returns the most appropriate version source.
// Priority: package.json > Cargo.toml > pyproject.toml > Chart.yaml > plain text
func GetBestVersionSource(sources []VersionSource) *VersionSource {
	if len(sources) == 0 {
		return nil
	}

	// Priority order
	priority := map[string]int{
		"package.json":   1,
		"Cargo.toml":     2,
		"pyproject.toml": 3,
		"Chart.yaml":     4,
		"version.txt":    5,
		"VERSION":        6,
	}

	best := &sources[0]
	bestPriority := priority[best.File]
	if bestPriority == 0 {
		bestPriority = 100
	}

	for i := 1; i < len(sources); i++ {
		p := priority[sources[i].File]
		if p == 0 {
			p = 100
		}
		if p < bestPriority {
			best = &sources[i]
			bestPriority = p
		}
	}

	return best
}

// FormatVersionSources formats the detected sources for display.
func FormatVersionSources(sources []VersionSource) string {
	if len(sources) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, s := range sources {
		sb.WriteString("  - ")
		sb.WriteString(s.Version)
		sb.WriteString(" from ")
		sb.WriteString(s.File)
		sb.WriteString(" (")
		sb.WriteString(s.Format)
		sb.WriteString(")\n")
	}

	return sb.String()
}
