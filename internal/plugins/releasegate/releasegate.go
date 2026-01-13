package releasegate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/indaco/sley/internal/semver"
)

// ReleaseGate defines the interface for release gate validation.
type ReleaseGate interface {
	Name() string
	Description() string
	Version() string
	ValidateRelease(newVersion, previousVersion semver.SemVersion, bumpType string) error
}

// ReleaseGatePlugin implements the ReleaseGate interface.
type ReleaseGatePlugin struct {
	cfg    *Config
	gitOps GitOperations
}

// Ensure ReleaseGatePlugin implements ReleaseGate.
var _ ReleaseGate = (*ReleaseGatePlugin)(nil)

// NewReleaseGate creates a new ReleaseGatePlugin instance.
// Uses the default OSGitOperations for git operations.
func NewReleaseGate(cfg *Config) *ReleaseGatePlugin {
	return NewReleaseGateWithOps(cfg, NewOSGitOperations())
}

// NewReleaseGateWithOps creates a new ReleaseGatePlugin with custom git operations.
// This constructor enables dependency injection for testing.
func NewReleaseGateWithOps(cfg *Config, gitOps GitOperations) *ReleaseGatePlugin {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if gitOps == nil {
		gitOps = NewOSGitOperations()
	}
	return &ReleaseGatePlugin{
		cfg:    cfg,
		gitOps: gitOps,
	}
}

// Name returns the plugin name.
func (p *ReleaseGatePlugin) Name() string {
	return "release-gate"
}

// Description returns a brief description of the plugin.
func (p *ReleaseGatePlugin) Description() string {
	return "Enforces quality gates before allowing version bumps"
}

// Version returns the plugin version.
func (p *ReleaseGatePlugin) Version() string {
	return "v0.1.0"
}

// GetConfig returns the plugin configuration.
func (p *ReleaseGatePlugin) GetConfig() *Config {
	return p.cfg
}

// IsEnabled returns true if the plugin is enabled.
func (p *ReleaseGatePlugin) IsEnabled() bool {
	return p.cfg != nil && p.cfg.Enabled
}

// ValidateRelease checks if a version bump is allowed based on configured gates.
func (p *ReleaseGatePlugin) ValidateRelease(newVersion, previousVersion semver.SemVersion, bumpType string) error {
	if !p.IsEnabled() {
		return nil
	}

	// Check worktree cleanliness
	if p.cfg.RequireCleanWorktree {
		if err := p.checkWorktreeClean(); err != nil {
			return err
		}
	}

	// Check branch constraints
	if err := p.checkBranchConstraints(); err != nil {
		return err
	}

	// Check for WIP commits
	if p.cfg.BlockedOnWIPCommits {
		if err := p.checkWIPCommits(); err != nil {
			return err
		}
	}

	// CI status check is not yet implemented
	// When enabled, this will check CI status before allowing bumps
	// For now, we skip this check even if enabled

	return nil
}

// checkWorktreeClean verifies that the git working tree is clean.
func (p *ReleaseGatePlugin) checkWorktreeClean() error {
	clean, err := p.gitOps.IsWorktreeClean()
	if err != nil {
		// If we can't check git status, we should fail safe
		return fmt.Errorf("release-gate: failed to check git status: %w", err)
	}

	if !clean {
		return fmt.Errorf("release-gate: uncommitted changes detected. Commit or stash changes before bumping")
	}

	return nil
}

// checkBranchConstraints validates that the current branch is allowed for bumps.
func (p *ReleaseGatePlugin) checkBranchConstraints() error {
	branch, err := p.gitOps.GetCurrentBranch()
	if err != nil {
		// If we can't get the branch, skip this check
		return nil
	}

	// Check blocked branches first (takes precedence)
	if len(p.cfg.BlockedBranches) > 0 {
		for _, blocked := range p.cfg.BlockedBranches {
			matched, err := matchBranchPattern(blocked, branch)
			if err != nil {
				return fmt.Errorf("release-gate: invalid blocked branch pattern %q: %w", blocked, err)
			}
			if matched {
				return fmt.Errorf("release-gate: bumps not allowed from branch %q (blocked branches: %v)", branch, p.cfg.BlockedBranches)
			}
		}
	}

	// Check allowed branches (if configured)
	if len(p.cfg.AllowedBranches) > 0 {
		allowed := false
		for _, pattern := range p.cfg.AllowedBranches {
			matched, err := matchBranchPattern(pattern, branch)
			if err != nil {
				return fmt.Errorf("release-gate: invalid allowed branch pattern %q: %w", pattern, err)
			}
			if matched {
				allowed = true
				break
			}
		}

		if !allowed {
			return fmt.Errorf("release-gate: bumps not allowed from branch %q. Allowed branches: %v", branch, p.cfg.AllowedBranches)
		}
	}

	return nil
}

// checkWIPCommits checks if recent commits contain WIP markers.
func (p *ReleaseGatePlugin) checkWIPCommits() error {
	commits, err := p.gitOps.GetRecentCommits(10)
	if err != nil {
		// If we can't get commits, skip this check
		return nil
	}

	wipPatterns := []string{
		`(?i)\bWIP\b`,
		`(?i)\bfixup!`,
		`(?i)\bsquash!`,
		`(?i)\bdo not merge\b`,
		`(?i)\bDNM\b`,
	}

	for _, commit := range commits {
		for _, pattern := range wipPatterns {
			matched, err := regexp.MatchString(pattern, commit)
			if err != nil {
				continue
			}
			if matched {
				// Extract just the commit message (remove hash)
				parts := strings.SplitN(commit, " ", 2)
				message := commit
				if len(parts) == 2 {
					message = parts[1]
				}
				return fmt.Errorf("release-gate: WIP commit detected in recent history: %q. Complete your work before releasing", message)
			}
		}
	}

	return nil
}

// matchBranchPattern checks if a branch name matches a glob-like pattern.
// Supports * as wildcard (e.g., "release/*" matches "release/v1.0").
func matchBranchPattern(pattern, branch string) (bool, error) {
	// Convert glob pattern to regex
	// Escape all regex special characters except *
	regexPattern := regexp.QuoteMeta(pattern)
	// Replace escaped \* with .* for wildcard matching
	regexPattern = strings.ReplaceAll(regexPattern, `\*`, ".*")
	// Anchor the pattern
	regexPattern = "^" + regexPattern + "$"

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return false, err
	}

	return re.MatchString(branch), nil
}
