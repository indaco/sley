package initcmd

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
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

// customKeyMap returns a KeyMap with Esc added as a quit key.
func customKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("esc", "cancel"),
	)
	return km
}

// PromptPluginSelection shows an interactive multi-select prompt for plugin selection.
// Returns the list of selected plugin names or an error if the user cancels.
func PromptPluginSelection(detectionSummary string) ([]string, error) {
	var selected []string

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

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select plugins to enable:").
				Description(description).
				Options(options...).
				Value(&selected).
				Filterable(false),
		),
	).WithKeyMap(customKeyMap())

	// Set defaults after form creation
	selected = defaults

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, fmt.Errorf("plugin selection canceled by user")
		}
		return nil, fmt.Errorf("plugin selection failed: %w", err)
	}

	return selected, nil
}

// ConfirmOverwrite asks the user if they want to overwrite an existing .sley.yaml file.
func ConfirmOverwrite() (bool, error) {
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(".sley.yaml already exists. Overwrite?").
				Description("This will replace your existing configuration.").
				Value(&confirmed),
		),
	).WithKeyMap(customKeyMap())

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, nil
		}
		return false, fmt.Errorf("confirmation failed: %w", err)
	}

	return confirmed, nil
}
