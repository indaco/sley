package tagmanager

import (
	"fmt"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/semver"
)

// TagManager defines the interface for git tag operations.
type TagManager interface {
	Name() string
	Description() string
	Version() string

	// CreateTag creates a git tag for the given version.
	CreateTag(version semver.SemVersion, message string) error

	// TagExists checks if a tag for the given version already exists.
	TagExists(version semver.SemVersion) (bool, error)

	// GetLatestTag returns the latest semver tag from git.
	GetLatestTag() (semver.SemVersion, error)

	// ValidateTagAvailable ensures a tag can be created for the version.
	ValidateTagAvailable(version semver.SemVersion) error

	// FormatTagName formats a version as a tag name.
	FormatTagName(version semver.SemVersion) string
}

// Config holds configuration for the tag manager plugin.
type Config struct {
	// Enabled controls whether the plugin is active.
	Enabled bool

	// AutoCreate automatically creates tags after version bumps.
	AutoCreate bool

	// Prefix is the tag prefix (default: "v").
	Prefix string

	// Annotate creates annotated tags instead of lightweight tags.
	Annotate bool

	// Push automatically pushes tags to remote after creation.
	Push bool

	// TagPrereleases controls whether tags are created for pre-release versions.
	// When false, tags are only created for stable releases (major/minor/patch).
	// Default: false (opt-in for pre-release tagging).
	TagPrereleases bool

	// Sign creates GPG-signed tags using git tag -s.
	// Requires git to be configured with a GPG signing key.
	// Default: false.
	Sign bool

	// SigningKey specifies the GPG key ID to use for signing.
	// If empty, git uses the default signing key from user.signingkey config.
	// Only used when Sign is true.
	SigningKey string

	// MessageTemplate is a template for the tag message.
	// Supports placeholders: {version}, {tag}, {prefix}, {date}, {major}, {minor}, {patch}, {prerelease}, {build}
	// Default: "Release {version}" for annotated/signed tags.
	MessageTemplate string

	// CommitMessageTemplate is a template for the commit message created before tagging.
	// Only used when AutoCreate is true. Supports the same placeholders as MessageTemplate.
	// Default: "chore(release): {tag}"
	CommitMessageTemplate string
}

// DefaultConfig returns the default tag manager configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:               false,
		AutoCreate:            false,
		Prefix:                "v",
		Annotate:              true,
		Push:                  false,
		TagPrereleases:        false,
		Sign:                  false,
		SigningKey:            "",
		MessageTemplate:       "Release {version}",
		CommitMessageTemplate: "chore(release): {tag}",
	}
}

// TagManagerPlugin implements the TagManager interface.
type TagManagerPlugin struct {
	config    *Config
	gitOps    core.GitTagOperations
	commitOps core.GitCommitOperations
}

// Ensure TagManagerPlugin implements TagManager.
var _ TagManager = (*TagManagerPlugin)(nil)

func (p *TagManagerPlugin) Name() string { return "tag-manager" }
func (p *TagManagerPlugin) Description() string {
	return "Manages git tags synchronized with version bumps"
}
func (p *TagManagerPlugin) Version() string { return "v0.1.0" }

// NewTagManager creates a new tag manager plugin with the given configuration.
// Uses the default OSGitTagOperations and OSGitCommitOperations for git operations.
func NewTagManager(cfg *Config) *TagManagerPlugin {
	return NewTagManagerWithOps(cfg, NewOSGitTagOperations(), NewOSGitCommitOperations())
}

// NewTagManagerWithOps creates a new tag manager plugin with custom git operations.
// This constructor enables dependency injection for testing.
func NewTagManagerWithOps(cfg *Config, gitOps core.GitTagOperations, commitOps core.GitCommitOperations) *TagManagerPlugin {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if gitOps == nil {
		gitOps = NewOSGitTagOperations()
	}
	if commitOps == nil {
		commitOps = NewOSGitCommitOperations()
	}
	return &TagManagerPlugin{
		config:    cfg,
		gitOps:    gitOps,
		commitOps: commitOps,
	}
}

// FormatTagName formats a version as a tag name using the configured prefix.
func (p *TagManagerPlugin) FormatTagName(version semver.SemVersion) string {
	return p.config.Prefix + version.String()
}

// CreateTag creates a git tag for the given version.
func (p *TagManagerPlugin) CreateTag(version semver.SemVersion, message string) error {
	tagName := p.FormatTagName(version)

	// Check if tag already exists
	exists, err := p.TagExists(version)
	if err != nil {
		return fmt.Errorf("failed to check tag existence: %w", err)
	}
	if exists {
		return fmt.Errorf("tag %s already exists", tagName)
	}

	// Format the message using template if no explicit message provided
	if message == "" {
		template := p.config.MessageTemplate
		if template == "" {
			template = "Release {version}"
		}
		data := NewTemplateData(version, p.config.Prefix)
		message = FormatMessage(template, data)
	}

	// Create the tag based on configuration
	switch {
	case p.config.Sign:
		// GPG-signed tag (implies annotated)
		if err := p.gitOps.CreateSignedTag(tagName, message, p.config.SigningKey); err != nil {
			return fmt.Errorf("failed to create signed tag: %w", err)
		}
	case p.config.Annotate:
		// Annotated tag (not signed)
		if err := p.gitOps.CreateAnnotatedTag(tagName, message); err != nil {
			return fmt.Errorf("failed to create annotated tag: %w", err)
		}
	default:
		// Lightweight tag (no message)
		if err := p.gitOps.CreateLightweightTag(tagName); err != nil {
			return fmt.Errorf("failed to create lightweight tag: %w", err)
		}
	}

	// Optionally push the tag
	if p.config.Push {
		if err := p.gitOps.PushTag(tagName); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}
	}

	return nil
}

// FormatTagMessage formats a tag message using the configured template.
func (p *TagManagerPlugin) FormatTagMessage(version semver.SemVersion) string {
	template := p.config.MessageTemplate
	if template == "" {
		template = "Release {version}"
	}
	data := NewTemplateData(version, p.config.Prefix)
	return FormatMessage(template, data)
}

// TagExists checks if a tag for the given version already exists.
func (p *TagManagerPlugin) TagExists(version semver.SemVersion) (bool, error) {
	tagName := p.FormatTagName(version)
	return p.gitOps.TagExists(tagName)
}

// GetLatestTag returns the latest semver tag from git.
func (p *TagManagerPlugin) GetLatestTag() (semver.SemVersion, error) {
	tag, err := p.gitOps.GetLatestTag()
	if err != nil {
		return semver.SemVersion{}, err
	}

	// Strip prefix if present
	versionStr := tag
	if len(tag) > len(p.config.Prefix) && tag[:len(p.config.Prefix)] == p.config.Prefix {
		versionStr = tag[len(p.config.Prefix):]
	}

	version, err := semver.ParseVersion(versionStr)
	if err != nil {
		return semver.SemVersion{}, fmt.Errorf("failed to parse tag %s as version: %w", tag, err)
	}

	return version, nil
}

// ValidateTagAvailable ensures a tag can be created for the version.
func (p *TagManagerPlugin) ValidateTagAvailable(version semver.SemVersion) error {
	exists, err := p.TagExists(version)
	if err != nil {
		return fmt.Errorf("failed to check tag availability: %w", err)
	}
	if exists {
		tagName := p.FormatTagName(version)
		return fmt.Errorf("tag %s already exists", tagName)
	}
	return nil
}

// IsAutoCreateEnabled returns whether the plugin is enabled with auto-create on.
// This gates automatic tag validation and creation during bumps.
func (p *TagManagerPlugin) IsAutoCreateEnabled() bool {
	return p.config.Enabled && p.config.AutoCreate
}

// CommitChanges stages modified files and creates a commit before tagging.
// It detects modified files via git status plus any explicitly provided extraFiles,
// then commits with a message formatted from the CommitMessageTemplate.
func (p *TagManagerPlugin) CommitChanges(version semver.SemVersion, extraFiles []string) error {
	var filesToStage []string

	// Start with explicitly provided files (e.g., the .version file path)
	filesToStage = append(filesToStage, extraFiles...)

	// Also detect any other modified files via git status
	modified, err := p.commitOps.GetModifiedFiles()
	if err != nil {
		return fmt.Errorf("failed to detect modified files: %w", err)
	}
	filesToStage = append(filesToStage, modified...)

	// Deduplicate files
	filesToStage = deduplicateStrings(filesToStage)

	if len(filesToStage) == 0 {
		return nil
	}

	// Stage files
	if err := p.commitOps.StageFiles(filesToStage...); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Format commit message using template
	template := p.config.CommitMessageTemplate
	if template == "" {
		template = "chore(release): {tag}"
	}
	data := NewTemplateData(version, p.config.Prefix)
	message := FormatMessage(template, data)

	// Create commit
	if err := p.commitOps.Commit(message); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// deduplicateStrings removes duplicate strings while preserving order.
func deduplicateStrings(items []string) []string {
	seen := make(map[string]bool, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// GetConfig returns the plugin configuration.
func (p *TagManagerPlugin) GetConfig() *Config {
	return p.config
}

// ShouldCreateTag determines if a tag should be created for the given version.
// Returns true if tagging is enabled and either:
// - The version is a stable release (no pre-release), or
// - The version is a pre-release and TagPrereleases is true.
func (p *TagManagerPlugin) ShouldCreateTag(version semver.SemVersion) bool {
	if !p.IsAutoCreateEnabled() {
		return false
	}

	// If it's a pre-release version, check if pre-release tagging is enabled
	if version.PreRelease != "" {
		return p.config.TagPrereleases
	}

	// Stable releases are always tagged when the plugin is enabled
	return true
}
