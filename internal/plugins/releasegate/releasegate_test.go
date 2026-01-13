package releasegate

import (
	"errors"
	"testing"

	"github.com/indaco/sley/internal/semver"
)

func TestNewReleaseGate(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want *Config
	}{
		{
			name: "with custom config",
			cfg: &Config{
				Enabled:              true,
				RequireCleanWorktree: true,
				RequireCIPass:        false,
				BlockedOnWIPCommits:  true,
				AllowedBranches:      []string{"main"},
				BlockedBranches:      []string{"dev"},
			},
			want: &Config{
				Enabled:              true,
				RequireCleanWorktree: true,
				RequireCIPass:        false,
				BlockedOnWIPCommits:  true,
				AllowedBranches:      []string{"main"},
				BlockedBranches:      []string{"dev"},
			},
		},
		{
			name: "with nil config uses defaults",
			cfg:  nil,
			want: DefaultConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewReleaseGate(tt.cfg)
			if got == nil {
				t.Fatal("NewReleaseGate returned nil")
			}

			if got.cfg.Enabled != tt.want.Enabled {
				t.Errorf("Enabled = %v, want %v", got.cfg.Enabled, tt.want.Enabled)
			}
			if got.cfg.RequireCleanWorktree != tt.want.RequireCleanWorktree {
				t.Errorf("RequireCleanWorktree = %v, want %v", got.cfg.RequireCleanWorktree, tt.want.RequireCleanWorktree)
			}
			if got.cfg.RequireCIPass != tt.want.RequireCIPass {
				t.Errorf("RequireCIPass = %v, want %v", got.cfg.RequireCIPass, tt.want.RequireCIPass)
			}
			if got.cfg.BlockedOnWIPCommits != tt.want.BlockedOnWIPCommits {
				t.Errorf("BlockedOnWIPCommits = %v, want %v", got.cfg.BlockedOnWIPCommits, tt.want.BlockedOnWIPCommits)
			}
		})
	}
}

func TestReleaseGatePlugin_Metadata(t *testing.T) {
	plugin := NewReleaseGate(DefaultConfig())

	if got := plugin.Name(); got != "release-gate" {
		t.Errorf("Name() = %q, want %q", got, "release-gate")
	}

	if got := plugin.Description(); got == "" {
		t.Error("Description() returned empty string")
	}

	if got := plugin.Version(); got == "" {
		t.Error("Version() returned empty string")
	}
}

func TestReleaseGatePlugin_IsEnabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want bool
	}{
		{
			name: "enabled",
			cfg:  &Config{Enabled: true},
			want: true,
		},
		{
			name: "disabled",
			cfg:  &Config{Enabled: false},
			want: false,
		},
		{
			name: "nil config",
			cfg:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewReleaseGate(tt.cfg)
			if got := plugin.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReleaseGatePlugin_GetConfig(t *testing.T) {
	t.Run("returns config when set", func(t *testing.T) {
		config := &Config{Enabled: true, RequireCleanWorktree: true}
		plugin := NewReleaseGate(config)
		got := plugin.GetConfig()
		if got != config {
			t.Errorf("GetConfig() = %v, want %v", got, config)
		}
	})

	t.Run("returns default config when nil passed", func(t *testing.T) {
		plugin := NewReleaseGate(nil)
		got := plugin.GetConfig()
		if got == nil {
			t.Error("GetConfig() should return default config, got nil")
		}
	})
}

func TestReleaseGatePlugin_ValidateRelease_Disabled(t *testing.T) {
	plugin := NewReleaseGate(&Config{Enabled: false})

	newVersion := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}
	prevVersion := semver.SemVersion{Major: 0, Minor: 1, Patch: 0}

	err := plugin.ValidateRelease(newVersion, prevVersion, "major")
	if err != nil {
		t.Errorf("ValidateRelease() with disabled plugin returned error: %v", err)
	}
}

func TestReleaseGatePlugin_CheckWorktreeClean(t *testing.T) {
	tests := []struct {
		name        string
		clean       bool
		gitErr      error
		wantErr     bool
		errContains string
	}{
		{
			name:    "clean worktree",
			clean:   true,
			gitErr:  nil,
			wantErr: false,
		},
		{
			name:        "dirty worktree",
			clean:       false,
			gitErr:      nil,
			wantErr:     true,
			errContains: "uncommitted changes detected",
		},
		{
			name:        "git status error",
			clean:       false,
			gitErr:      errors.New("not a git repository"),
			wantErr:     true,
			errContains: "failed to check git status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOps := &MockGitOperations{
				IsWorktreeCleanFn: func() (bool, error) {
					return tt.clean, tt.gitErr
				},
			}

			plugin := NewReleaseGateWithOps(&Config{
				Enabled:              true,
				RequireCleanWorktree: true,
			}, mockOps)

			err := plugin.checkWorktreeClean()

			if tt.wantErr {
				if err == nil {
					t.Error("checkWorktreeClean() expected error, got nil")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("checkWorktreeClean() error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("checkWorktreeClean() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestReleaseGatePlugin_CheckBranchConstraints(t *testing.T) {
	tests := []struct {
		name            string
		currentBranch   string
		branchErr       error
		allowedBranches []string
		blockedBranches []string
		wantErr         bool
		errContains     string
	}{
		{
			name:            "no constraints",
			currentBranch:   "main",
			allowedBranches: []string{},
			blockedBranches: []string{},
			wantErr:         false,
		},
		{
			name:            "allowed branch exact match",
			currentBranch:   "main",
			allowedBranches: []string{"main", "release/*"},
			wantErr:         false,
		},
		{
			name:            "allowed branch wildcard match",
			currentBranch:   "release/v1.0",
			allowedBranches: []string{"main", "release/*"},
			wantErr:         false,
		},
		{
			name:            "not allowed branch",
			currentBranch:   "feature/test",
			allowedBranches: []string{"main", "release/*"},
			wantErr:         true,
			errContains:     "not allowed from branch",
		},
		{
			name:            "blocked branch exact match",
			currentBranch:   "dev",
			blockedBranches: []string{"dev", "experimental/*"},
			wantErr:         true,
			errContains:     "not allowed from branch",
		},
		{
			name:            "blocked branch wildcard match",
			currentBranch:   "experimental/feature",
			blockedBranches: []string{"dev", "experimental/*"},
			wantErr:         true,
			errContains:     "not allowed from branch",
		},
		{
			name:            "blocked takes precedence over allowed",
			currentBranch:   "main",
			allowedBranches: []string{"main"},
			blockedBranches: []string{"main"},
			wantErr:         true,
			errContains:     "not allowed from branch",
		},
		{
			name:            "git error skips check",
			currentBranch:   "",
			branchErr:       errors.New("not a git repository"),
			allowedBranches: []string{"main"},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOps := &MockGitOperations{
				GetCurrentBranchFn: func() (string, error) {
					return tt.currentBranch, tt.branchErr
				},
			}

			plugin := NewReleaseGateWithOps(&Config{
				Enabled:         true,
				AllowedBranches: tt.allowedBranches,
				BlockedBranches: tt.blockedBranches,
			}, mockOps)

			err := plugin.checkBranchConstraints()

			if tt.wantErr {
				if err == nil {
					t.Error("checkBranchConstraints() expected error, got nil")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("checkBranchConstraints() error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("checkBranchConstraints() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestReleaseGatePlugin_CheckWIPCommits(t *testing.T) {
	tests := []struct {
		name        string
		commits     []string
		commitsErr  error
		wantErr     bool
		errContains string
	}{
		{
			name: "no WIP commits",
			commits: []string{
				"abc123 feat: add new feature",
				"def456 fix: resolve bug",
			},
			wantErr: false,
		},
		{
			name: "WIP commit uppercase",
			commits: []string{
				"abc123 WIP: work in progress",
				"def456 fix: resolve bug",
			},
			wantErr:     true,
			errContains: "WIP commit detected",
		},
		{
			name: "WIP commit lowercase",
			commits: []string{
				"abc123 wip: work in progress",
				"def456 fix: resolve bug",
			},
			wantErr:     true,
			errContains: "WIP commit detected",
		},
		{
			name: "fixup commit",
			commits: []string{
				"abc123 fixup! previous commit",
				"def456 fix: resolve bug",
			},
			wantErr:     true,
			errContains: "WIP commit detected",
		},
		{
			name: "squash commit",
			commits: []string{
				"abc123 squash! combine commits",
				"def456 fix: resolve bug",
			},
			wantErr:     true,
			errContains: "WIP commit detected",
		},
		{
			name: "do not merge commit",
			commits: []string{
				"abc123 DO NOT MERGE: testing",
				"def456 fix: resolve bug",
			},
			wantErr:     true,
			errContains: "WIP commit detected",
		},
		{
			name: "DNM commit",
			commits: []string{
				"abc123 DNM: testing",
				"def456 fix: resolve bug",
			},
			wantErr:     true,
			errContains: "WIP commit detected",
		},
		{
			name:       "git error skips check",
			commits:    []string{},
			commitsErr: errors.New("not a git repository"),
			wantErr:    false,
		},
		{
			name:    "empty commit history",
			commits: []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOps := &MockGitOperations{
				GetRecentCommitsFn: func(count int) ([]string, error) {
					return tt.commits, tt.commitsErr
				},
			}

			plugin := NewReleaseGateWithOps(&Config{
				Enabled:             true,
				BlockedOnWIPCommits: true,
			}, mockOps)

			err := plugin.checkWIPCommits()

			if tt.wantErr {
				if err == nil {
					t.Error("checkWIPCommits() expected error, got nil")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("checkWIPCommits() error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("checkWIPCommits() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestReleaseGatePlugin_ValidateRelease_Integration(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *Config
		worktreeClean bool
		worktreeErr   error
		currentBranch string
		branchErr     error
		commits       []string
		commitsErr    error
		wantErr       bool
		errContains   string
	}{
		{
			name: "all checks pass",
			cfg: &Config{
				Enabled:              true,
				RequireCleanWorktree: true,
				BlockedOnWIPCommits:  true,
				AllowedBranches:      []string{"main"},
			},
			worktreeClean: true,
			currentBranch: "main",
			commits: []string{
				"abc123 feat: add feature",
			},
			wantErr: false,
		},
		{
			name: "dirty worktree fails",
			cfg: &Config{
				Enabled:              true,
				RequireCleanWorktree: true,
			},
			worktreeClean: false,
			wantErr:       true,
			errContains:   "uncommitted changes detected",
		},
		{
			name: "blocked branch fails",
			cfg: &Config{
				Enabled:         true,
				BlockedBranches: []string{"dev"},
			},
			worktreeClean: true,
			currentBranch: "dev",
			wantErr:       true,
			errContains:   "not allowed from branch",
		},
		{
			name: "WIP commit fails",
			cfg: &Config{
				Enabled:             true,
				BlockedOnWIPCommits: true,
			},
			worktreeClean: true,
			currentBranch: "main",
			commits: []string{
				"abc123 WIP: testing",
			},
			wantErr:     true,
			errContains: "WIP commit detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOps := &MockGitOperations{
				IsWorktreeCleanFn: func() (bool, error) {
					return tt.worktreeClean, tt.worktreeErr
				},
				GetCurrentBranchFn: func() (string, error) {
					return tt.currentBranch, tt.branchErr
				},
				GetRecentCommitsFn: func(count int) ([]string, error) {
					return tt.commits, tt.commitsErr
				},
			}

			plugin := NewReleaseGateWithOps(tt.cfg, mockOps)
			newVersion := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}
			prevVersion := semver.SemVersion{Major: 0, Minor: 1, Patch: 0}

			err := plugin.ValidateRelease(newVersion, prevVersion, "major")

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateRelease() expected error, got nil")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("ValidateRelease() error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateRelease() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestMatchBranchPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		branch  string
		want    bool
		wantErr bool
	}{
		{
			name:    "exact match",
			pattern: "main",
			branch:  "main",
			want:    true,
		},
		{
			name:    "exact no match",
			pattern: "main",
			branch:  "develop",
			want:    false,
		},
		{
			name:    "wildcard prefix match",
			pattern: "release/*",
			branch:  "release/v1.0",
			want:    true,
		},
		{
			name:    "wildcard prefix no match",
			pattern: "release/*",
			branch:  "hotfix/v1.0",
			want:    false,
		},
		{
			name:    "wildcard suffix match",
			pattern: "*/v1.0",
			branch:  "release/v1.0",
			want:    true,
		},
		{
			name:    "multiple wildcards",
			pattern: "feature/*/test/*",
			branch:  "feature/foo/test/bar",
			want:    true,
		},
		{
			name:    "special characters in branch name",
			pattern: "fix/bug-123",
			branch:  "fix/bug-123",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matchBranchPattern(tt.pattern, tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("matchBranchPattern() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("matchBranchPattern(%q, %q) = %v, want %v", tt.pattern, tt.branch, got, tt.want)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	// Save original state
	origGetter := GetReleaseGateFn
	origRegisterer := RegisterReleaseGateFn
	defer func() {
		GetReleaseGateFn = origGetter
		RegisterReleaseGateFn = origRegisterer
		Unregister()
	}()

	// Reset to defaults
	GetReleaseGateFn = func() ReleaseGate {
		return defaultReleaseGate
	}
	RegisterReleaseGateFn = func(rg ReleaseGate) {
		defaultReleaseGate = rg
	}

	t.Run("Register", func(t *testing.T) {
		cfg := &Config{Enabled: true}
		Register(cfg)

		rg := GetReleaseGateFn()
		if rg == nil {
			t.Error("Register() did not set the default release gate")
		}

		plugin, ok := rg.(*ReleaseGatePlugin)
		if !ok {
			t.Error("Register() did not set a ReleaseGatePlugin instance")
		}

		if !plugin.IsEnabled() {
			t.Error("Register() plugin is not enabled")
		}
	})

	t.Run("Unregister", func(t *testing.T) {
		Register(&Config{Enabled: true})
		Unregister()

		rg := GetReleaseGateFn()
		if rg != nil {
			t.Error("Unregister() did not clear the default release gate")
		}
	})
}

func TestNewReleaseGateWithOps_NilGitOps(t *testing.T) {
	// When gitOps is nil, it should default to OSGitOperations
	plugin := NewReleaseGateWithOps(nil, nil)

	if plugin == nil {
		t.Fatal("NewReleaseGateWithOps() returned nil")
	}
	if plugin.gitOps == nil {
		t.Error("NewReleaseGateWithOps() should set default gitOps when nil")
	}
}

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
