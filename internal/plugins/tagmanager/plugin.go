package tagmanager

import (
	"fmt"

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
	// Default: true (for backward compatibility).
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
}

// DefaultConfig returns the default tag manager configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:         false,
		AutoCreate:      true,
		Prefix:          "v",
		Annotate:        true,
		Push:            false,
		TagPrereleases:  true,
		Sign:            false,
		SigningKey:      "",
		MessageTemplate: "Release {version}",
	}
}

// TagManagerPlugin implements the TagManager interface.
type TagManagerPlugin struct {
	config *Config
}

// Ensure TagManagerPlugin implements TagManager.
var _ TagManager = (*TagManagerPlugin)(nil)

func (p *TagManagerPlugin) Name() string { return "tag-manager" }
func (p *TagManagerPlugin) Description() string {
	return "Manages git tags synchronized with version bumps"
}
func (p *TagManagerPlugin) Version() string { return "v0.1.0" }

// NewTagManager creates a new tag manager plugin with the given configuration.
func NewTagManager(cfg *Config) *TagManagerPlugin {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &TagManagerPlugin{config: cfg}
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
		if err := createSignedTagFn(tagName, message, p.config.SigningKey); err != nil {
			return fmt.Errorf("failed to create signed tag: %w", err)
		}
	case p.config.Annotate:
		// Annotated tag (not signed)
		if err := createAnnotatedTagFn(tagName, message); err != nil {
			return fmt.Errorf("failed to create annotated tag: %w", err)
		}
	default:
		// Lightweight tag (no message)
		if err := createLightweightTagFn(tagName); err != nil {
			return fmt.Errorf("failed to create lightweight tag: %w", err)
		}
	}

	// Optionally push the tag
	if p.config.Push {
		if err := pushTagFn(tagName); err != nil {
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
	return tagExistsFn(tagName)
}

// GetLatestTag returns the latest semver tag from git.
func (p *TagManagerPlugin) GetLatestTag() (semver.SemVersion, error) {
	tag, err := getLatestTagFn()
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

// IsEnabled returns whether auto-create is enabled.
func (p *TagManagerPlugin) IsEnabled() bool {
	return p.config.Enabled && p.config.AutoCreate
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
	if !p.IsEnabled() {
		return false
	}

	// If it's a pre-release version, check if pre-release tagging is enabled
	if version.PreRelease != "" {
		return p.config.TagPrereleases
	}

	// Stable releases are always tagged when the plugin is enabled
	return true
}
