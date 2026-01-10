package config

// PluginConfig holds configuration for all built-in plugins.
type PluginConfig struct {
	CommitParser       bool                      `yaml:"commit-parser"`
	TagManager         *TagManagerConfig         `yaml:"tag-manager,omitempty"`
	VersionValidator   *VersionValidatorConfig   `yaml:"version-validator,omitempty"`
	DependencyCheck    *DependencyCheckConfig    `yaml:"dependency-check,omitempty"`
	ChangelogParser    *ChangelogParserConfig    `yaml:"changelog-parser,omitempty"`
	ChangelogGenerator *ChangelogGeneratorConfig `yaml:"changelog-generator,omitempty"`
	ReleaseGate        *ReleaseGateConfig        `yaml:"release-gate,omitempty"`
	AuditLog           *AuditLogConfig           `yaml:"audit-log,omitempty"`
}

// TagManagerConfig holds configuration for the tag manager plugin.
type TagManagerConfig struct {
	// Enabled controls whether the plugin is active.
	Enabled bool `yaml:"enabled"`

	// AutoCreate automatically creates tags after version bumps.
	AutoCreate *bool `yaml:"auto-create,omitempty"`

	// Prefix is the tag prefix (default: "v").
	Prefix string `yaml:"prefix,omitempty"`

	// Annotate creates annotated tags instead of lightweight tags.
	Annotate *bool `yaml:"annotate,omitempty"`

	// Push automatically pushes tags to remote after creation.
	Push bool `yaml:"push,omitempty"`
}

// GetAutoCreate returns the auto-create setting with default true.
func (c *TagManagerConfig) GetAutoCreate() bool {
	if c.AutoCreate == nil {
		return true
	}
	return *c.AutoCreate
}

// GetAnnotate returns the annotate setting with default true.
func (c *TagManagerConfig) GetAnnotate() bool {
	if c.Annotate == nil {
		return true
	}
	return *c.Annotate
}

// GetPrefix returns the prefix with default "v".
func (c *TagManagerConfig) GetPrefix() string {
	if c.Prefix == "" {
		return "v"
	}
	return c.Prefix
}

// VersionValidatorConfig holds configuration for the version validator plugin.
type VersionValidatorConfig struct {
	// Enabled controls whether the plugin is active.
	Enabled bool `yaml:"enabled"`

	// Rules defines the validation rules to apply.
	Rules []ValidationRule `yaml:"rules,omitempty"`
}

// ValidationRule defines a single validation rule.
type ValidationRule struct {
	// Type is the rule type (e.g., "pre-release-format", "major-version-max").
	Type string `yaml:"type"`

	// Pattern is a regex pattern for format validation rules.
	Pattern string `yaml:"pattern,omitempty"`

	// Value is a numeric limit for max-version rules.
	Value int `yaml:"value,omitempty"`

	// Enabled controls whether this specific rule is active.
	Enabled bool `yaml:"enabled,omitempty"`

	// Branch is a glob pattern for branch-constraint rules.
	Branch string `yaml:"branch,omitempty"`

	// Allowed lists allowed bump types for branch-constraint rules.
	Allowed []string `yaml:"allowed,omitempty"`
}

// DependencyCheckConfig holds configuration for the dependency check plugin.
type DependencyCheckConfig struct {
	// Enabled controls whether the plugin is active.
	Enabled bool `yaml:"enabled"`

	// AutoSync automatically syncs versions after bumps.
	AutoSync bool `yaml:"auto-sync,omitempty"`

	// Files lists the files to check and sync.
	Files []DependencyFileConfig `yaml:"files,omitempty"`
}

// DependencyFileConfig defines a single file to check/sync.
type DependencyFileConfig struct {
	// Path is the file path relative to repository root.
	Path string `yaml:"path"`

	// Field is the dot-notation path to the version field (for JSON/YAML/TOML).
	Field string `yaml:"field,omitempty"`

	// Format specifies the file format: json, yaml, toml, raw, regex
	Format string `yaml:"format"`

	// Pattern is the regex pattern for "regex" format.
	Pattern string `yaml:"pattern,omitempty"`
}

// ChangelogParserConfig holds configuration for the changelog parser plugin.
type ChangelogParserConfig struct {
	// Enabled controls whether the plugin is active.
	Enabled bool `yaml:"enabled"`

	// Path is the path to the changelog file (default: "CHANGELOG.md").
	Path string `yaml:"path,omitempty"`

	// RequireUnreleasedSection enforces presence of Unreleased section.
	RequireUnreleasedSection bool `yaml:"require-unreleased-section,omitempty"`

	// InferBumpType enables automatic bump type inference from changelog.
	InferBumpType bool `yaml:"infer-bump-type,omitempty"`

	// Priority determines which parser takes precedence: "changelog" or "commits"
	Priority string `yaml:"priority,omitempty"`
}

// ChangelogGeneratorConfig holds configuration for the changelog generator plugin.
type ChangelogGeneratorConfig struct {
	// Enabled controls whether the plugin is active.
	Enabled bool `yaml:"enabled"`

	// Mode determines output style: "versioned", "unified", or "both".
	// "versioned" writes each version to a separate file (e.g., .changes/v1.2.3.md)
	// "unified" writes to a single CHANGELOG.md file
	// "both" writes to both locations
	Mode string `yaml:"mode,omitempty"`

	// Format determines the changelog format: "grouped" or "keepachangelog".
	// "grouped" (default): Custom group labels with commit type grouping
	// "keepachangelog": Keep a Changelog specification format with standard sections
	Format string `yaml:"format,omitempty"`

	// ChangesDir is the directory for version-specific changelog files (versioned mode).
	ChangesDir string `yaml:"changes-dir,omitempty"`

	// ChangelogPath is the path to the unified changelog file.
	ChangelogPath string `yaml:"changelog-path,omitempty"`

	// HeaderTemplate is the path to a custom header template file.
	HeaderTemplate string `yaml:"header-template,omitempty"`

	// Repository contains git repository settings for link generation.
	// Supports GitHub, GitLab, Codeberg, Gitea, Bitbucket, and custom hosts.
	Repository *RepositoryConfig `yaml:"repository,omitempty"`

	// Groups defines commit grouping rules.
	Groups []CommitGroupConfig `yaml:"groups,omitempty"`

	// ExcludePatterns lists regex patterns for commits to exclude.
	ExcludePatterns []string `yaml:"exclude-patterns,omitempty"`

	// IncludeNonConventional includes commits that don't follow conventional commit format
	// in an "Other Changes" section. When false (default), these commits are skipped
	// and a warning is printed listing the skipped commits.
	IncludeNonConventional bool `yaml:"include-non-conventional,omitempty"`

	// UseDefaultIcons enables predefined icons for all commit groups and contributors.
	// When true, default icons (emojis) are automatically applied to commit group headers
	// and the contributors section. User-defined GroupIcons or Contributors.Icon values
	// will override the defaults for specific entries. This is a convenient shorthand
	// instead of manually specifying all icons via group-icons.
	UseDefaultIcons bool `yaml:"use-default-icons,omitempty"`

	// GroupIcons maps default group labels to icons. Use this to add icons to default
	// groups without redefining patterns and labels. Ignored if Groups is specified.
	// Keys must match default labels: Enhancements, Fixes, Refactors, Documentation,
	// Performance, Styling, Tests, Chores, CI, Build, Reverts.
	GroupIcons map[string]string `yaml:"group-icons,omitempty"`

	// BreakingChangesIcon is the icon/emoji for the breaking changes section header.
	// Used by formatters with a dedicated breaking changes section (e.g., GitHub format).
	// When UseDefaultIcons is true and this is empty, the default icon is used.
	BreakingChangesIcon string `yaml:"breaking-changes-icon,omitempty"`

	// Contributors configures the contributors section.
	Contributors *ContributorsConfig `yaml:"contributors,omitempty"`
}

// GetChangesDir returns the changes directory with default ".changes".
func (c *ChangelogGeneratorConfig) GetChangesDir() string {
	if c.ChangesDir == "" {
		return ".changes"
	}
	return c.ChangesDir
}

// GetChangelogPath returns the changelog path with default "CHANGELOG.md".
func (c *ChangelogGeneratorConfig) GetChangelogPath() string {
	if c.ChangelogPath == "" {
		return "CHANGELOG.md"
	}
	return c.ChangelogPath
}

// GetMode returns the mode with default "versioned".
func (c *ChangelogGeneratorConfig) GetMode() string {
	if c.Mode == "" {
		return "versioned"
	}
	return c.Mode
}

// GetFormat returns the format with default "grouped".
func (c *ChangelogGeneratorConfig) GetFormat() string {
	if c.Format == "" {
		return "grouped"
	}
	return c.Format
}

// RepositoryConfig holds git repository settings for changelog links.
// Supports multiple providers: github, gitlab, codeberg, gitea, bitbucket, custom.
type RepositoryConfig struct {
	// Provider is the git hosting provider: github, gitlab, codeberg, gitea, bitbucket, custom.
	// Default: auto-detected from git remote URL.
	Provider string `yaml:"provider,omitempty"`

	// Host is the git server hostname (e.g., "github.com", "gitlab.com", "codeberg.org").
	// Required for custom providers or when auto-detect is disabled.
	Host string `yaml:"host,omitempty"`

	// Owner is the repository owner/organization.
	Owner string `yaml:"owner,omitempty"`

	// Repo is the repository name.
	Repo string `yaml:"repo,omitempty"`

	// AutoDetect enables automatic detection from git remote.
	AutoDetect bool `yaml:"auto-detect,omitempty"`
}

// CommitGroupConfig defines a grouping rule for commits.
type CommitGroupConfig struct {
	// Pattern is the regex pattern to match commit types.
	Pattern string `yaml:"pattern"`

	// Label is the section header label.
	Label string `yaml:"label"`

	// Icon is the icon/emoji for the section (optional).
	Icon string `yaml:"icon,omitempty"`

	// Order determines the display order (lower = higher priority).
	Order int `yaml:"order,omitempty"`
}

// ContributorsConfig configures the contributors section in changelog.
type ContributorsConfig struct {
	// Enabled controls whether to include contributors section.
	Enabled bool `yaml:"enabled,omitempty"`

	// Format is a Go template for contributor formatting.
	Format string `yaml:"format,omitempty"`

	// Icon is the icon/emoji for the contributors section header (optional).
	Icon string `yaml:"icon,omitempty"`

	// ShowNewContributors enables the "New Contributors" section showing first-time contributors.
	// Default: true when contributors are enabled.
	ShowNewContributors *bool `yaml:"show-new-contributors,omitempty"`

	// NewContributorsFormat is a Go template for new contributor entries.
	// Available fields: {{.Name}}, {{.Username}}, {{.Host}}, {{.PRNumber}}, {{.CommitHash}}
	// Default: "* [@{{.Username}}](https://{{.Host}}/{{.Username}}) made their first contribution in [#{{.PRNumber}}](...)"
	NewContributorsFormat string `yaml:"new-contributors-format,omitempty"`

	// NewContributorsIcon is the icon for the "New Contributors" section header.
	NewContributorsIcon string `yaml:"new-contributors-icon,omitempty"`
}

// GetShowNewContributors returns the show-new-contributors setting with default true.
func (c *ContributorsConfig) GetShowNewContributors() bool {
	if c.ShowNewContributors == nil {
		return true
	}
	return *c.ShowNewContributors
}

// ReleaseGateConfig holds configuration for the release gate plugin.
type ReleaseGateConfig struct {
	// Enabled controls whether the plugin is active.
	Enabled bool `yaml:"enabled"`

	// RequireCleanWorktree blocks bumps if git has uncommitted changes.
	RequireCleanWorktree bool `yaml:"require-clean-worktree,omitempty"`

	// RequireCIPass checks CI status before allowing bumps (disabled by default).
	RequireCIPass bool `yaml:"require-ci-pass,omitempty"`

	// BlockedOnWIPCommits blocks if recent commits contain WIP/fixup/squash.
	BlockedOnWIPCommits bool `yaml:"blocked-on-wip-commits,omitempty"`

	// AllowedBranches lists branches where bumps are allowed (empty = all allowed).
	AllowedBranches []string `yaml:"allowed-branches,omitempty"`

	// BlockedBranches lists branches where bumps are never allowed.
	BlockedBranches []string `yaml:"blocked-branches,omitempty"`
}

// AuditLogConfig holds configuration for the audit log plugin.
type AuditLogConfig struct {
	// Enabled controls whether the plugin is active.
	Enabled bool `yaml:"enabled"`

	// Path is the path to the audit log file.
	Path string `yaml:"path,omitempty"`

	// Format specifies the output format: json or yaml.
	Format string `yaml:"format,omitempty"`

	// IncludeAuthor includes git author in log entries.
	IncludeAuthor bool `yaml:"include-author,omitempty"`

	// IncludeTimestamp includes ISO 8601 timestamp in log entries.
	IncludeTimestamp bool `yaml:"include-timestamp,omitempty"`

	// IncludeCommitSHA includes current commit SHA in log entries.
	IncludeCommitSHA bool `yaml:"include-commit-sha,omitempty"`

	// IncludeBranch includes current branch name in log entries.
	IncludeBranch bool `yaml:"include-branch,omitempty"`
}

// GetPath returns the path with default ".version-history.json".
func (c *AuditLogConfig) GetPath() string {
	if c.Path == "" {
		return ".version-history.json"
	}
	return c.Path
}

// GetFormat returns the format with default "json".
func (c *AuditLogConfig) GetFormat() string {
	if c.Format == "" {
		return "json"
	}
	return c.Format
}
