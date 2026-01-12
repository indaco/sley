package changeloggenerator

import (
	"fmt"
	"os"

	"github.com/indaco/sley/internal/tui"
)

// ChangelogGenerator defines the interface for changelog generation.
type ChangelogGenerator interface {
	Name() string
	Description() string
	Version() string

	// GenerateForVersion generates changelog for a specific version bump.
	GenerateForVersion(version, previousVersion, bumpType string) error

	// IsEnabled returns whether the plugin is enabled.
	IsEnabled() bool

	// GetConfig returns the plugin configuration.
	GetConfig() *Config
}

// ChangelogGeneratorPlugin implements the ChangelogGenerator interface.
type ChangelogGeneratorPlugin struct {
	config    *Config
	generator *Generator
}

// Ensure ChangelogGeneratorPlugin implements ChangelogGenerator.
var _ ChangelogGenerator = (*ChangelogGeneratorPlugin)(nil)

// IsInteractiveFn is a function variable for checking if the environment is interactive.
// Can be overridden in tests.
var IsInteractiveFn = tui.IsInteractive

// ConfirmMergeFn is a function variable for prompting user confirmation.
// Can be overridden in tests.
var ConfirmMergeFn = func(message string) (bool, error) {
	prompter := tui.NewModulePrompt(nil)
	return prompter.ConfirmOperation(message)
}

// NewChangelogGenerator creates a new changelog generator plugin.
func NewChangelogGenerator(cfg *Config) (*ChangelogGeneratorPlugin, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	generator, err := NewGenerator(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create generator: %w", err)
	}

	return &ChangelogGeneratorPlugin{
		config:    cfg,
		generator: generator,
	}, nil
}

// Name returns the plugin name.
func (p *ChangelogGeneratorPlugin) Name() string { return "changelog-generator" }

// Description returns the plugin description.
func (p *ChangelogGeneratorPlugin) Description() string {
	return "Generates changelog from conventional commits"
}

// Version returns the plugin version.
func (p *ChangelogGeneratorPlugin) Version() string { return "v0.1.0" }

// IsEnabled returns whether the plugin is enabled.
func (p *ChangelogGeneratorPlugin) IsEnabled() bool {
	return p.config.Enabled
}

// GetConfig returns the plugin configuration.
func (p *ChangelogGeneratorPlugin) GetConfig() *Config {
	return p.config
}

// GenerateForVersion generates changelog for a version bump.
func (p *ChangelogGeneratorPlugin) GenerateForVersion(version, previousVersion, bumpType string) error {
	if !p.config.Enabled {
		return nil
	}

	// Get commits between versions
	commits, err := GetCommitsWithMetaFn(previousVersion, "HEAD")
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(commits) == 0 {
		return nil // No commits to process
	}

	// Generate changelog content with result
	result := p.generator.GenerateVersionChangelogWithResult(version, previousVersion, commits)

	// Print warning about skipped non-conventional commits
	if len(result.SkippedNonConventional) > 0 {
		fmt.Fprintf(os.Stderr, "\nWarning: %d non-conventional commit(s) skipped:\n", len(result.SkippedNonConventional))
		for _, c := range result.SkippedNonConventional {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", c.ShortHash, c.Subject)
		}
		fmt.Fprintf(os.Stderr, "Tip: Use conventional commit format (type: description) or set 'include-non-conventional: true' in config.\n\n")
	}

	// Write based on mode
	return p.writeChangelog(version, result.Content)
}

// writeChangelog writes the changelog based on configured mode.
func (p *ChangelogGeneratorPlugin) writeChangelog(version, content string) error {
	mode := p.config.Mode

	switch mode {
	case "versioned":
		if err := p.generator.WriteVersionedFile(version, content); err != nil {
			return err
		}
		// Handle merge-after for versioned mode
		return p.handleMergeAfter()
	case "unified":
		return p.generator.WriteUnifiedChangelog(content)
	case "both":
		if err := p.generator.WriteVersionedFile(version, content); err != nil {
			return err
		}
		return p.generator.WriteUnifiedChangelog(content)
	default:
		return fmt.Errorf("unknown mode: %s", mode)
	}
}

// handleMergeAfter handles the merge-after behavior for versioned changelog files.
func (p *ChangelogGeneratorPlugin) handleMergeAfter() error {
	switch p.config.MergeAfter {
	case "immediate":
		if err := p.generator.MergeVersionedFiles(); err != nil {
			return fmt.Errorf("failed to merge changelog files: %w", err)
		}
		fmt.Fprintf(os.Stdout, "Merged versioned changelog files into %s\n", p.config.ChangelogPath)
		return nil

	case "prompt":
		// Skip prompt if not in interactive environment
		if !IsInteractiveFn() {
			fmt.Fprintf(os.Stdout, "Non-interactive environment detected, skipping changelog merge prompt.\n")
			return nil
		}
		confirmed, err := ConfirmMergeFn(fmt.Sprintf("Merge versioned changelog files into %s?", p.config.ChangelogPath))
		if err != nil {
			// Treat prompt errors as declined
			return nil
		}
		if confirmed {
			if err := p.generator.MergeVersionedFiles(); err != nil {
				return fmt.Errorf("failed to merge changelog files: %w", err)
			}
			fmt.Fprintf(os.Stdout, "Merged versioned changelog files into %s\n", p.config.ChangelogPath)
		}
		return nil

	case "manual":
		// No automatic merge
		return nil

	default:
		// Unknown value, treat as manual (no merge)
		return nil
	}
}
