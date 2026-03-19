package dependencycheck

import (
	"fmt"
	"strings"
)

// DependencyChecker defines the interface for dependency version checking.
type DependencyChecker interface {
	Name() string
	Description() string
	Version() string

	// CheckConsistency validates all configured files match the current version.
	CheckConsistency(currentVersion string) ([]Inconsistency, error)

	// SyncVersions updates all configured files to the new version.
	SyncVersions(newVersion string) error

	// IsEnabled returns whether the plugin is active.
	IsEnabled() bool

	// GetConfig returns the plugin configuration.
	GetConfig() *Config
}

// Config holds configuration for the dependency checker plugin.
type Config struct {
	// Enabled controls whether the plugin is active.
	Enabled bool

	// AutoSync automatically syncs versions after bumps.
	AutoSync bool

	// Files lists the files to check and sync.
	Files []FileConfig
}

// FileConfig defines a single file to check/sync.
type FileConfig struct {
	// Path is the file path relative to repository root.
	Path string

	// Field is the dot-notation path to the version field (for JSON/YAML/TOML).
	// Example: "version", "tool.poetry.version", "metadata.version"
	Field string

	// Format specifies the file format: json, yaml, toml, raw, regex
	Format string

	// Pattern is the regex pattern for "regex" format.
	// Use capturing group for version: e.g., `version = "(.*?)"`
	Pattern string
}

// Inconsistency represents a version mismatch in a file.
type Inconsistency struct {
	Path     string
	Expected string
	Found    string
	Format   string
}

// String returns a formatted string representation of the inconsistency.
func (i Inconsistency) String() string {
	return fmt.Sprintf("%s: expected %s, found %s (format: %s)", i.Path, i.Expected, i.Found, i.Format)
}

// DependencyCheckerPlugin implements the DependencyChecker interface.
type DependencyCheckerPlugin struct {
	config *Config

	// Format-specific read functions (injected for testability).
	readJSONVersionFn  func(path, field string) (string, error)
	readYAMLVersionFn  func(path, field string) (string, error)
	readTOMLVersionFn  func(path, field string) (string, error)
	readRawVersionFn   func(path string) (string, error)
	readRegexVersionFn func(path, pattern string) (string, error)

	// Format-specific write functions (injected for testability).
	writeJSONVersionFn  func(path, field, version string) error
	writeYAMLVersionFn  func(path, field, version string) error
	writeTOMLVersionFn  func(path, field, version string) error
	writeRawVersionFn   func(path, version string) error
	writeRegexVersionFn func(path, pattern, version string) error
}

// Ensure DependencyCheckerPlugin implements DependencyChecker.
var _ DependencyChecker = (*DependencyCheckerPlugin)(nil)

func (p *DependencyCheckerPlugin) Name() string { return "dependency-check" }
func (p *DependencyCheckerPlugin) Description() string {
	return "Validates and syncs version numbers across multiple files"
}
func (p *DependencyCheckerPlugin) Version() string { return "v0.1.0" }

// NewDependencyChecker creates a new dependency checker plugin with the given configuration.
func NewDependencyChecker(cfg *Config) *DependencyCheckerPlugin {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &DependencyCheckerPlugin{
		config:              cfg,
		readJSONVersionFn:   readJSONVersion,
		readYAMLVersionFn:   readYAMLVersion,
		readTOMLVersionFn:   readTOMLVersion,
		readRawVersionFn:    readRawVersion,
		readRegexVersionFn:  readRegexVersion,
		writeJSONVersionFn:  writeJSONVersion,
		writeYAMLVersionFn:  writeYAMLVersion,
		writeTOMLVersionFn:  writeTOMLVersion,
		writeRawVersionFn:   writeRawVersion,
		writeRegexVersionFn: writeRegexVersion,
	}
}

// DefaultConfig returns the default dependency checker configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:  false,
		AutoSync: false,
		Files:    []FileConfig{},
	}
}

// IsEnabled returns whether the plugin is active.
func (p *DependencyCheckerPlugin) IsEnabled() bool {
	return p.config.Enabled
}

// GetConfig returns the plugin configuration.
func (p *DependencyCheckerPlugin) GetConfig() *Config {
	return p.config
}

// CheckConsistency validates all configured files match the current version.
func (p *DependencyCheckerPlugin) CheckConsistency(currentVersion string) ([]Inconsistency, error) {
	if !p.IsEnabled() {
		return nil, nil
	}

	var inconsistencies []Inconsistency
	normalizedExpected := normalizeVersion(currentVersion)

	for _, file := range p.config.Files {
		version, err := p.readVersionFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read version from %s: %w", file.Path, err)
		}

		normalizedFound := normalizeVersion(version)
		if normalizedFound != normalizedExpected {
			inconsistencies = append(inconsistencies, Inconsistency{
				Path:     file.Path,
				Expected: currentVersion,
				Found:    version,
				Format:   file.Format,
			})
		}
	}

	return inconsistencies, nil
}

// SyncVersions updates all configured files to the new version.
func (p *DependencyCheckerPlugin) SyncVersions(newVersion string) error {
	if !p.IsEnabled() {
		return nil
	}

	for _, file := range p.config.Files {
		if err := p.writeVersionToFile(file, newVersion); err != nil {
			return fmt.Errorf("failed to write version to %s: %w", file.Path, err)
		}
	}

	return nil
}

// readVersionFromFile reads the version from a file based on its format.
func (p *DependencyCheckerPlugin) readVersionFromFile(file FileConfig) (string, error) {
	switch file.Format {
	case "json":
		return p.readJSONVersionFn(file.Path, file.Field)
	case "yaml":
		return p.readYAMLVersionFn(file.Path, file.Field)
	case "toml":
		return p.readTOMLVersionFn(file.Path, file.Field)
	case "raw":
		return p.readRawVersionFn(file.Path)
	case "regex":
		if file.Pattern == "" {
			return "", fmt.Errorf("regex format requires a pattern")
		}
		return p.readRegexVersionFn(file.Path, file.Pattern)
	default:
		return "", fmt.Errorf("unsupported format: %s", file.Format)
	}
}

// writeVersionToFile writes the version to a file based on its format.
func (p *DependencyCheckerPlugin) writeVersionToFile(file FileConfig, version string) error {
	switch file.Format {
	case "json":
		return p.writeJSONVersionFn(file.Path, file.Field, version)
	case "yaml":
		return p.writeYAMLVersionFn(file.Path, file.Field, version)
	case "toml":
		return p.writeTOMLVersionFn(file.Path, file.Field, version)
	case "raw":
		return p.writeRawVersionFn(file.Path, version)
	case "regex":
		if file.Pattern == "" {
			return fmt.Errorf("regex format requires a pattern")
		}
		return p.writeRegexVersionFn(file.Path, file.Pattern, version)
	default:
		return fmt.Errorf("unsupported format: %s", file.Format)
	}
}

// normalizeVersion strips the "v" prefix for comparison.
func normalizeVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}
