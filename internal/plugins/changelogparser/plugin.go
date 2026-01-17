package changelogparser

import (
	"errors"
	"fmt"
	"os"
)

// ChangelogInferrer defines the interface for parsing changelog files.
type ChangelogInferrer interface {
	Name() string
	Description() string
	Version() string
	InferBumpType() (string, error)
	ValidateHasEntries() error
}

// ChangelogParserPlugin implements the ChangelogInferrer interface.
type ChangelogParserPlugin struct {
	config *Config
}

// Config holds configuration for the changelog parser plugin.
type Config struct {
	Enabled                  bool
	Path                     string
	RequireUnreleasedSection bool
	InferBumpType            bool
	Priority                 string
	Format                   string
	GroupedSectionMap        map[string]string
}

var _ ChangelogInferrer = (*ChangelogParserPlugin)(nil)

func (p *ChangelogParserPlugin) Name() string { return "changelog-parser" }
func (p *ChangelogParserPlugin) Description() string {
	return "Parses CHANGELOG.md to infer bump type and validate changelog completeness"
}
func (p *ChangelogParserPlugin) Version() string { return "v0.2.0" }

// NewChangelogParser creates a new changelog parser plugin.
func NewChangelogParser(cfg *Config) *ChangelogParserPlugin {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if cfg.Path == "" {
		cfg.Path = "CHANGELOG.md"
	}
	if cfg.Priority == "" {
		cfg.Priority = "changelog"
	}
	if cfg.Format == "" {
		cfg.Format = "keepachangelog"
	}
	return &ChangelogParserPlugin{config: cfg}
}

// DefaultConfig returns the default changelog parser configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:                  false,
		Path:                     "CHANGELOG.md",
		RequireUnreleasedSection: true,
		InferBumpType:            true,
		Priority:                 "changelog",
		Format:                   "keepachangelog",
	}
}

func (p *ChangelogParserPlugin) IsEnabled() bool {
	return p.config.Enabled
}

func (p *ChangelogParserPlugin) GetConfig() *Config {
	return p.config
}

// InferBumpType parses the changelog and infers the bump type.
func (p *ChangelogParserPlugin) InferBumpType() (string, error) {
	if !p.IsEnabled() || !p.config.InferBumpType {
		return "", errors.New("changelog parser not enabled or inference disabled")
	}

	section, err := p.parseUnreleasedWithFormat()
	if err != nil {
		return "", fmt.Errorf("failed to parse unreleased section: %w", err)
	}

	if section.InferredBumpType == "" {
		return "", errors.New("failed to infer bump type: no bump type could be determined")
	}

	return section.InferredBumpType, nil
}

// InferBumpTypeWithConfidence returns bump type and confidence level.
func (p *ChangelogParserPlugin) InferBumpTypeWithConfidence() (bumpType, confidence string, err error) {
	if !p.IsEnabled() || !p.config.InferBumpType {
		return "", "", errors.New("changelog parser not enabled or inference disabled")
	}

	section, err := p.parseUnreleasedWithFormat()
	if err != nil {
		return "", "", fmt.Errorf("failed to parse unreleased section: %w", err)
	}

	return section.InferredBumpType, section.BumpTypeConfidence, nil
}

// ValidateHasEntries validates that the Unreleased section has entries.
func (p *ChangelogParserPlugin) ValidateHasEntries() error {
	if !p.IsEnabled() || !p.config.RequireUnreleasedSection {
		return nil
	}

	section, err := p.parseUnreleasedWithFormat()
	if err != nil {
		return fmt.Errorf("changelog validation failed: %w", err)
	}

	if !section.HasEntries {
		return errors.New("changelog validation failed: Unreleased section has no entries")
	}

	return nil
}

// ShouldTakePrecedence returns true if changelog parser should take precedence.
func (p *ChangelogParserPlugin) ShouldTakePrecedence() bool {
	return p.IsEnabled() && p.config.Priority == "changelog"
}

// parseUnreleasedWithFormat creates the appropriate parser and parses.
func (p *ChangelogParserPlugin) parseUnreleasedWithFormat() (*ParsedSection, error) {
	parser, err := NewParser(p.config.Format, p.config)
	if err != nil {
		return nil, err
	}

	file, err := openFileFn(p.config.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("changelog file not found")
		}
		return nil, err
	}
	defer file.Close()

	return parser.ParseUnreleased(file)
}

// GetFormat returns the configured format.
func (p *ChangelogParserPlugin) GetFormat() string {
	return p.config.Format
}
