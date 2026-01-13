package initcmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/indaco/sley/internal/tui"
)

// PluginOption represents a selectable plugin with metadata.
type PluginOption struct {
	Name        string
	Description string
	Default     bool
}

// AllPluginOptions returns all available plugins with their descriptions.
// The order matches the desired display order in the prompt.
func AllPluginOptions() []PluginOption {
	return []PluginOption{
		{
			Name:        "commit-parser",
			Description: "Analyze conventional commits to determine bump type",
			Default:     true,
		},
		{
			Name:        "tag-manager",
			Description: "Auto-create git tags after version bumps",
			Default:     true,
		},
		{
			Name:        "version-validator",
			Description: "Enforce versioning policies and constraints",
			Default:     false,
		},
		{
			Name:        "dependency-check",
			Description: "Sync version to package.json and other files",
			Default:     false,
		},
		{
			Name:        "changelog-parser",
			Description: "Infer bump type from CHANGELOG.md",
			Default:     false,
		},
		{
			Name:        "changelog-generator",
			Description: "Generate changelogs from commits",
			Default:     false,
		},
		{
			Name:        "release-gate",
			Description: "Pre-bump validation (clean worktree, CI status)",
			Default:     false,
		},
		{
			Name:        "audit-log",
			Description: "Record version history with metadata",
			Default:     false,
		},
	}
}

// DefaultPluginNames returns the names of plugins that are enabled by default.
func DefaultPluginNames() []string {
	defaults := []string{}
	for _, opt := range AllPluginOptions() {
		if opt.Default {
			defaults = append(defaults, opt.Name)
		}
	}
	return defaults
}

// PromptPluginSelection shows an interactive multi-select prompt for plugin selection.
// Returns the list of selected plugin names or an error if the user cancels.
func PromptPluginSelection(detectionSummary string) ([]string, error) {
	// Build options from available plugins
	options := make([]huh.Option[string], 0, len(AllPluginOptions()))
	for _, plugin := range AllPluginOptions() {
		label := fmt.Sprintf("%s - %s", plugin.Name, plugin.Description)
		options = append(options, huh.NewOption(label, plugin.Name))
	}

	// Set default selections
	defaults := DefaultPluginNames()

	// Build description with detection summary if available
	description := "Space: toggle | Enter: confirm | Esc: cancel"
	if detectionSummary != "" {
		description = detectionSummary + "\n" + description
	}

	return tui.MultiSelect("Select plugins to enable:", description, options, defaults)
}

// ConfirmOverwrite asks the user if they want to overwrite an existing .sley.yaml file.
func ConfirmOverwrite() (bool, error) {
	return tui.Confirm(
		".sley.yaml already exists. Overwrite?",
		"This will replace your existing configuration.",
	)
}

// confirmVersionMigration asks the user if they want to use the detected version.
func confirmVersionMigration(version, file string) (bool, error) {
	return tui.Confirm(
		fmt.Sprintf("Use version %s from %s?", version, file),
		"This will initialize .version with the detected version.",
	)
}

// selectVersionSource prompts the user to select a version source from multiple options.
func selectVersionSource(sources []VersionSource) (*VersionSource, error) {
	if len(sources) == 0 {
		return nil, nil
	}

	options := make([]huh.Option[int], len(sources))
	for i, s := range sources {
		label := fmt.Sprintf("%s from %s (%s)", s.Version, s.File, s.Format)
		options[i] = huh.NewOption(label, i)
	}

	selected, err := tui.Select(
		"Select version to use:",
		"Choose which version to migrate to .version file",
		options,
	)
	if err != nil {
		return nil, fmt.Errorf("selection failed: %w", err)
	}

	return &sources[selected], nil
}
