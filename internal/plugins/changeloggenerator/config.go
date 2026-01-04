package changeloggenerator

import "github.com/indaco/verso/internal/config"

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

	// GroupIcons maps default group labels to icons (used when Groups is empty).
	GroupIcons map[string]string

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
	Enabled bool
	Format  string
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
			Enabled: true,
			Format:  "- {{.Name}} ([@{{.Username}}](https://{{.Host}}/{{.Username}}))",
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
	}

	// Convert repository config
	if cfg.Repository != nil {
		result.Repository = &RepositoryConfig{
			Provider:   cfg.Repository.Provider,
			Host:       cfg.Repository.Host,
			Owner:      cfg.Repository.Owner,
			Repo:       cfg.Repository.Repo,
			AutoDetect: cfg.Repository.AutoDetect,
		}
	} else {
		result.Repository = &RepositoryConfig{AutoDetect: true}
	}

	// Convert groups
	if len(cfg.Groups) > 0 {
		// Full custom groups provided - use them directly
		result.Groups = make([]GroupConfig, len(cfg.Groups))
		for i, g := range cfg.Groups {
			result.Groups[i] = GroupConfig{
				Pattern: g.Pattern,
				Label:   g.Label,
				Icon:    g.Icon,
				Order:   g.Order,
			}
		}
	} else {
		// Use defaults, optionally with icons from GroupIcons
		result.Groups = DefaultGroups()
		if len(cfg.GroupIcons) > 0 {
			result.GroupIcons = cfg.GroupIcons
			// Apply icons to default groups by label
			for i, g := range result.Groups {
				if icon, ok := cfg.GroupIcons[g.Label]; ok {
					result.Groups[i].Icon = icon
				}
			}
		}
	}

	// Use default exclude patterns if none specified
	if len(result.ExcludePatterns) == 0 {
		result.ExcludePatterns = DefaultExcludePatterns()
	}

	// Convert contributors config
	if cfg.Contributors != nil {
		result.Contributors = &ContributorsConfig{
			Enabled: cfg.Contributors.Enabled,
			Format:  cfg.Contributors.Format,
		}
	} else {
		result.Contributors = &ContributorsConfig{
			Enabled: true,
			Format:  "- {{.Name}} ([@{{.Username}}](https://{{.Host}}/{{.Username}}))",
		}
	}

	return result
}
