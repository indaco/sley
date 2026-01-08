package versionvalidator

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/semver"
)

// VersionValidator defines the interface for version validation plugins.
type VersionValidator interface {
	Name() string
	Description() string
	Version() string
	Validate(newVersion, previousVersion semver.SemVersion, bumpType string) error
	ValidateSet(version semver.SemVersion) error
}

// RuleType defines the type of validation rule.
type RuleType string

const (
	RulePreReleaseFormat    RuleType = "pre-release-format"
	RuleMajorVersionMax     RuleType = "major-version-max"
	RuleMinorVersionMax     RuleType = "minor-version-max"
	RulePatchVersionMax     RuleType = "patch-version-max"
	RuleRequirePreRelease0x RuleType = "require-pre-release-for-0x"
	RuleBranchConstraint    RuleType = "branch-constraint"
	RuleNoMajorBump         RuleType = "no-major-bump"
	RuleNoMinorBump         RuleType = "no-minor-bump"
	RuleNoPatchBump         RuleType = "no-patch-bump"
)

// Rule represents a single validation rule.
type Rule struct {
	Type    RuleType `yaml:"type"`
	Pattern string   `yaml:"pattern,omitempty"`
	Value   int      `yaml:"value,omitempty"`
	Enabled bool     `yaml:"enabled,omitempty"`
	Branch  string   `yaml:"branch,omitempty"`
	Allowed []string `yaml:"allowed,omitempty"`
}

// Config holds the configuration for the version validator plugin.
type Config struct {
	Enabled bool   `yaml:"enabled"`
	Rules   []Rule `yaml:"rules,omitempty"`
}

// DefaultConfig returns the default configuration for the version validator.
func DefaultConfig() *Config {
	return &Config{
		Enabled: false,
		Rules:   []Rule{},
	}
}

// VersionValidatorPlugin implements the VersionValidator interface.
type VersionValidatorPlugin struct {
	cfg *Config
}

// NewVersionValidator creates a new VersionValidatorPlugin instance.
func NewVersionValidator(cfg *Config) *VersionValidatorPlugin {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &VersionValidatorPlugin{cfg: cfg}
}

// Name returns the plugin name.
func (p *VersionValidatorPlugin) Name() string {
	return "version-validator"
}

// Description returns a brief description of the plugin.
func (p *VersionValidatorPlugin) Description() string {
	return "Enforces versioning policies and constraints beyond basic SemVer syntax validation"
}

// Version returns the plugin version.
func (p *VersionValidatorPlugin) Version() string {
	return "v0.1.0"
}

// GetConfig returns the plugin configuration.
func (p *VersionValidatorPlugin) GetConfig() *Config {
	return p.cfg
}

// IsEnabled returns true if the plugin is enabled.
func (p *VersionValidatorPlugin) IsEnabled() bool {
	return p.cfg != nil && p.cfg.Enabled
}

// Validate checks if the version transition is valid according to configured rules.
func (p *VersionValidatorPlugin) Validate(newVersion, previousVersion semver.SemVersion, bumpType string) error {
	if !p.IsEnabled() {
		return nil
	}

	for _, rule := range p.cfg.Rules {
		if err := p.applyRule(rule, newVersion, previousVersion, bumpType); err != nil {
			return err
		}
	}

	return nil
}

// ValidateSet checks if a manually set version is valid according to configured rules.
func (p *VersionValidatorPlugin) ValidateSet(version semver.SemVersion) error {
	if !p.IsEnabled() {
		return nil
	}

	for _, rule := range p.cfg.Rules {
		if err := p.applySetRule(rule, version); err != nil {
			return err
		}
	}

	return nil
}

// applyRule applies a single rule to a version bump operation.
func (p *VersionValidatorPlugin) applyRule(rule Rule, newVersion, previousVersion semver.SemVersion, bumpType string) error {
	switch rule.Type {
	case RulePreReleaseFormat:
		return p.validatePreReleaseFormat(rule, newVersion)
	case RuleMajorVersionMax:
		return p.validateMaxVersion(rule, newVersion.Major, "major")
	case RuleMinorVersionMax:
		return p.validateMaxVersion(rule, newVersion.Minor, "minor")
	case RulePatchVersionMax:
		return p.validateMaxVersion(rule, newVersion.Patch, "patch")
	case RuleRequirePreRelease0x:
		return p.validateRequirePreRelease0x(rule, newVersion)
	case RuleBranchConstraint:
		return p.validateBranchConstraint(rule, bumpType)
	case RuleNoMajorBump:
		return p.validateNoBumpType(rule, bumpType, "major")
	case RuleNoMinorBump:
		return p.validateNoBumpType(rule, bumpType, "minor")
	case RuleNoPatchBump:
		return p.validateNoBumpType(rule, bumpType, "patch")
	default:
		return fmt.Errorf("unknown rule type: %s", rule.Type)
	}
}

// applySetRule applies rules applicable to set operations.
func (p *VersionValidatorPlugin) applySetRule(rule Rule, version semver.SemVersion) error {
	switch rule.Type {
	case RulePreReleaseFormat:
		return p.validatePreReleaseFormat(rule, version)
	case RuleMajorVersionMax:
		return p.validateMaxVersion(rule, version.Major, "major")
	case RuleMinorVersionMax:
		return p.validateMaxVersion(rule, version.Minor, "minor")
	case RulePatchVersionMax:
		return p.validateMaxVersion(rule, version.Patch, "patch")
	case RuleRequirePreRelease0x:
		return p.validateRequirePreRelease0x(rule, version)
	default:
		// Other rules don't apply to set operations
		return nil
	}
}

// validatePreReleaseFormat checks if the pre-release label matches the configured pattern.
func (p *VersionValidatorPlugin) validatePreReleaseFormat(rule Rule, version semver.SemVersion) error {
	if version.PreRelease == "" {
		return nil // No pre-release to validate
	}

	if rule.Pattern == "" {
		return nil // No pattern configured
	}

	re, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return fmt.Errorf("invalid pre-release pattern %q: %w", rule.Pattern, err)
	}

	if !re.MatchString(version.PreRelease) {
		return fmt.Errorf("pre-release label %q does not match required pattern %q", version.PreRelease, rule.Pattern)
	}

	return nil
}

// validateMaxVersion checks if a version component exceeds the maximum allowed value.
func (p *VersionValidatorPlugin) validateMaxVersion(rule Rule, value int, component string) error {
	if rule.Value <= 0 {
		return nil // No max configured
	}

	if value > rule.Value {
		return fmt.Errorf("%s version %d exceeds maximum allowed value %d", component, value, rule.Value)
	}

	return nil
}

// validateRequirePreRelease0x checks if 0.x versions require a pre-release label.
func (p *VersionValidatorPlugin) validateRequirePreRelease0x(rule Rule, version semver.SemVersion) error {
	if !rule.Enabled {
		return nil
	}

	if version.Major == 0 && version.PreRelease == "" {
		return fmt.Errorf("version 0.x.x requires a pre-release label (e.g., 0.%d.%d-alpha)", version.Minor, version.Patch)
	}

	return nil
}

// defaultCurrentBranchReader is the default branch reader for backward compatibility.
var defaultCurrentBranchReader core.GitBranchReader = defaultBranchReader

// getCurrentBranchFn is kept for backward compatibility during migration.
var getCurrentBranchFn = func() (string, error) {
	return defaultCurrentBranchReader.GetCurrentBranch()
}

// validateBranchConstraint checks if the bump type is allowed on the current branch.
func (p *VersionValidatorPlugin) validateBranchConstraint(rule Rule, bumpType string) error {
	if rule.Branch == "" || len(rule.Allowed) == 0 {
		return nil
	}

	branch, err := getCurrentBranchFn()
	if err != nil {
		// If we can't get the branch, skip this validation
		return nil
	}

	// Check if the branch matches the pattern
	matched, err := matchBranchPattern(rule.Branch, branch)
	if err != nil {
		return fmt.Errorf("invalid branch pattern %q: %w", rule.Branch, err)
	}

	if !matched {
		return nil // Rule doesn't apply to this branch
	}

	// Check if the bump type is allowed
	if slices.Contains(rule.Allowed, bumpType) {
		return nil
	}

	return fmt.Errorf("bump type %q is not allowed on branch %q (allowed: %v)", bumpType, branch, rule.Allowed)
}

// validateNoBumpType checks if a specific bump type is disallowed.
func (p *VersionValidatorPlugin) validateNoBumpType(rule Rule, actualBumpType, restrictedType string) error {
	if !rule.Enabled {
		return nil
	}

	if actualBumpType == restrictedType {
		return fmt.Errorf("%s bumps are not allowed by policy", restrictedType)
	}

	return nil
}

// matchBranchPattern checks if a branch name matches a glob-like pattern.
func matchBranchPattern(pattern, branch string) (bool, error) {
	// Convert glob pattern to regex
	// Support * as wildcard
	regexPattern := "^" + regexp.QuoteMeta(pattern) + "$"
	regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return false, err
	}

	return re.MatchString(branch), nil
}
