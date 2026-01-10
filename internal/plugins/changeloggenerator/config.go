package changeloggenerator

import "github.com/indaco/sley/internal/config"

// DefaultGroupIcons maps default group labels to their icons.
// These are used when UseDefaultIcons is enabled.
var DefaultGroupIcons = map[string]string{
	"Enhancements":  "\U0001F680",   // rocket 🚀
	"Fixes":         "\U0001FA79",   // adhesive bandage 🩹
	"Refactors":     "\U0001F485",   // nail polish 💅
	"Documentation": "\U0001F4D6",   // open book 📖
	"Performance":   "\u26A1",       // high voltage ⚡
	"Styling":       "\U0001F3A8",   // artist palette 🎨
	"Tests":         "\u2705",       // check mark button  ✅
	"Chores":        "\U0001F3E1",   // house with garden 🏡
	"CI":            "\U0001F916",   // robot 🤖
	"Build":         "\U0001F4E6",   // package 📦
	"Reverts":       "\u25C0\uFE0F", // reverse button ◀️
}

// DefaultContributorIcon is the default icon for the contributors section.
const DefaultContributorIcon = "\u2764\uFE0F" // red heart ❤️

// DefaultNewContributorsIcon is the default icon for the new contributors section.
const DefaultNewContributorsIcon = "\U0001F389" // party popper 🎉

// DefaultBreakingChangesIcon is the default icon for the breaking changes section.
const DefaultBreakingChangesIcon = "\u26A0\uFE0F" // warning sign ⚠️

// Config holds the internal configuration for the changelog generator plugin.
type Config struct {
	// Enabled controls whether the plugin is active.
	Enabled bool

	// Mode determines output style: "versioned", "unified", or "both".
	Mode string

	// Format determines the changelog format: "grouped" or "keepachangelog".
	// "grouped" (default): Current behavior with custom group labels
	// "keepachangelog": Keep a Changelog specification format
	Format string

	// ChangesDir is the directory for version-specific changelog files.
	ChangesDir string

	// ChangelogPath is the path to the unified changelog file.
	ChangelogPath string

	// HeaderTemplate is the path to a custom header template file.
	HeaderTemplate string

	// Repository contains git repository settings for link generation.
	Repository *RepositoryConfig

	// Groups defines commit grouping rules.
	Groups []GroupConfig

	// ExcludePatterns lists regex patterns for commits to exclude.
	ExcludePatterns []string

	// IncludeNonConventional includes non-conventional commits in "Other Changes" section.
	// When false (default), these commits are skipped with a warning.
	IncludeNonConventional bool

	// UseDefaultIcons enables default icons for commit groups and contributors.
	// When true, predefined icons are applied to default groups and the contributors section.
	// User-defined GroupIcons or Contributors.Icon can override specific defaults.
	UseDefaultIcons bool

	// GroupIcons maps default group labels to icons (used when Groups is empty).
	GroupIcons map[string]string

	// BreakingChangesIcon is the icon for the breaking changes section header.
	// This is used by formatters that have a dedicated breaking changes section.
	BreakingChangesIcon string

	// Contributors configures the contributors section.
	Contributors *ContributorsConfig
}

// RepositoryConfig holds git repository settings for changelog links.
// Supports multiple providers: github, gitlab, codeberg, gitea, bitbucket, custom.
type RepositoryConfig struct {
	// Provider is the git hosting provider: github, gitlab, codeberg, gitea, bitbucket, custom.
	Provider string
	// Host is the git server hostname (e.g., "github.com", "gitlab.com").
	Host string
	// Owner is the repository owner/organization.
	Owner string
	// Repo is the repository name.
	Repo string
	// AutoDetect enables automatic detection from git remote.
	AutoDetect bool
}

// GroupConfig defines a grouping rule for commits.
type GroupConfig struct {
	Pattern string
	Label   string
	Icon    string
	Order   int
}

// ContributorsConfig configures the contributors section.
type ContributorsConfig struct {
	Enabled               bool
	Format                string
	Icon                  string
	ShowNewContributors   bool
	NewContributorsFormat string
	NewContributorsIcon   string
}

// DefaultConfig returns the default changelog generator configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:       false,
		Mode:          "versioned",
		Format:        "grouped",
		ChangesDir:    ".changes",
		ChangelogPath: "CHANGELOG.md",
		Repository: &RepositoryConfig{
			AutoDetect: true,
		},
		Groups:          DefaultGroups(),
		ExcludePatterns: DefaultExcludePatterns(),
		Contributors: &ContributorsConfig{
			Enabled:             true,
			Format:              "- [@{{.Username}}](https://{{.Host}}/{{.Username}})",
			ShowNewContributors: true,
		},
	}
}

// DefaultGroups returns the default commit grouping rules (git-cliff style).
// Order is derived from array position (first = 0, second = 1, etc.)
func DefaultGroups() []GroupConfig {
	return []GroupConfig{
		{Pattern: "^feat", Label: "Enhancements"},
		{Pattern: "^fix", Label: "Fixes"},
		{Pattern: "^refactor", Label: "Refactors"},
		{Pattern: "^docs?", Label: "Documentation"},
		{Pattern: "^perf", Label: "Performance"},
		{Pattern: "^style", Label: "Styling"},
		{Pattern: "^test", Label: "Tests"},
		{Pattern: "^chore", Label: "Chores"},
		{Pattern: "^ci", Label: "CI"},
		{Pattern: "^build", Label: "Build"},
		{Pattern: "^revert", Label: "Reverts"},
	}
}

// DefaultExcludePatterns returns the default patterns to exclude from changelog.
func DefaultExcludePatterns() []string {
	return []string{
		"^Merge",
		"^WIP",
		"^wip",
	}
}

// FromConfigStruct converts the config package struct to internal config.
func FromConfigStruct(cfg *config.ChangelogGeneratorConfig) *Config {
	if cfg == nil {
		return DefaultConfig()
	}

	result := &Config{
		Enabled:                cfg.Enabled,
		Mode:                   cfg.GetMode(),
		Format:                 cfg.GetFormat(),
		ChangesDir:             cfg.GetChangesDir(),
		ChangelogPath:          cfg.GetChangelogPath(),
		HeaderTemplate:         cfg.HeaderTemplate,
		ExcludePatterns:        cfg.ExcludePatterns,
		IncludeNonConventional: cfg.IncludeNonConventional,
		UseDefaultIcons:        cfg.UseDefaultIcons,
	}

	result.Repository = convertRepositoryConfig(cfg.Repository)
	result.Groups, result.GroupIcons = convertGroupsConfig(cfg)

	if len(result.ExcludePatterns) == 0 {
		result.ExcludePatterns = DefaultExcludePatterns()
	}

	result.BreakingChangesIcon = convertBreakingChangesIcon(cfg)
	result.Contributors = convertContributorsConfig(cfg)

	return result
}

// convertBreakingChangesIcon returns the breaking changes icon based on configuration.
// Priority: user-defined icon > default icon (when UseDefaultIcons is true) > empty string.
func convertBreakingChangesIcon(cfg *config.ChangelogGeneratorConfig) string {
	if cfg.BreakingChangesIcon != "" {
		return cfg.BreakingChangesIcon
	}
	if cfg.UseDefaultIcons {
		return DefaultBreakingChangesIcon
	}
	return ""
}

// convertRepositoryConfig converts repository configuration.
func convertRepositoryConfig(repo *config.RepositoryConfig) *RepositoryConfig {
	if repo == nil {
		return &RepositoryConfig{AutoDetect: true}
	}
	return &RepositoryConfig{
		Provider:   repo.Provider,
		Host:       repo.Host,
		Owner:      repo.Owner,
		Repo:       repo.Repo,
		AutoDetect: repo.AutoDetect,
	}
}

// convertGroupsConfig converts groups configuration with icon handling.
func convertGroupsConfig(cfg *config.ChangelogGeneratorConfig) ([]GroupConfig, map[string]string) {
	if len(cfg.Groups) > 0 {
		groups := make([]GroupConfig, len(cfg.Groups))
		for i, g := range cfg.Groups {
			groups[i] = GroupConfig{
				Pattern: g.Pattern,
				Label:   g.Label,
				Icon:    g.Icon,
				Order:   g.Order,
			}
		}
		return groups, nil
	}

	// Use defaults, optionally with icons
	groups := DefaultGroups()
	applyGroupIcons(groups, cfg.GroupIcons, cfg.UseDefaultIcons)
	return groups, cfg.GroupIcons
}

// applyGroupIcons applies icons to groups based on configuration.
func applyGroupIcons(groups []GroupConfig, groupIcons map[string]string, useDefaults bool) {
	for i, g := range groups {
		if icon, ok := groupIcons[g.Label]; ok {
			groups[i].Icon = icon
		} else if useDefaults {
			if icon, ok := DefaultGroupIcons[g.Label]; ok {
				groups[i].Icon = icon
			}
		}
	}
}

// convertContributorsConfig converts contributors configuration with icon handling.
func convertContributorsConfig(cfg *config.ChangelogGeneratorConfig) *ContributorsConfig {
	if cfg.Contributors == nil {
		return defaultContributorsConfig(cfg.UseDefaultIcons)
	}

	contrib := &ContributorsConfig{
		Enabled:               cfg.Contributors.Enabled,
		Format:                cfg.Contributors.Format,
		Icon:                  cfg.Contributors.Icon,
		ShowNewContributors:   cfg.Contributors.GetShowNewContributors(),
		NewContributorsFormat: cfg.Contributors.NewContributorsFormat,
		NewContributorsIcon:   cfg.Contributors.NewContributorsIcon,
	}

	applyDefaultContributorIcons(contrib, cfg.UseDefaultIcons)
	return contrib
}

// defaultContributorsConfig returns the default contributors configuration.
func defaultContributorsConfig(useDefaultIcons bool) *ContributorsConfig {
	contrib := &ContributorsConfig{
		Enabled:             true,
		Format:              "- {{.Name}} ([@{{.Username}}](https://{{.Host}}/{{.Username}}))",
		ShowNewContributors: true,
	}
	if useDefaultIcons {
		contrib.Icon = DefaultContributorIcon
		contrib.NewContributorsIcon = DefaultNewContributorsIcon
	}
	return contrib
}

// applyDefaultContributorIcons applies default icons if UseDefaultIcons is enabled.
func applyDefaultContributorIcons(contrib *ContributorsConfig, useDefaultIcons bool) {
	if !useDefaultIcons {
		return
	}
	if contrib.Icon == "" {
		contrib.Icon = DefaultContributorIcon
	}
	if contrib.NewContributorsIcon == "" {
		contrib.NewContributorsIcon = DefaultNewContributorsIcon
	}
}
