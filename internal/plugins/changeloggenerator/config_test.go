package changeloggenerator

import (
	"testing"

	"github.com/indaco/sley/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("expected Enabled to be false by default")
	}
	if cfg.Mode != "versioned" {
		t.Errorf("Mode = %q, want 'versioned'", cfg.Mode)
	}
	if cfg.ChangesDir != ".changes" {
		t.Errorf("ChangesDir = %q, want '.changes'", cfg.ChangesDir)
	}
	if cfg.ChangelogPath != "CHANGELOG.md" {
		t.Errorf("ChangelogPath = %q, want 'CHANGELOG.md'", cfg.ChangelogPath)
	}
	if cfg.Repository == nil {
		t.Fatal("expected Repository to be non-nil")
	}
	if !cfg.Repository.AutoDetect {
		t.Error("expected Repository.AutoDetect to be true by default")
	}
	if len(cfg.Groups) == 0 {
		t.Error("expected Groups to have default values")
	}
	if len(cfg.ExcludePatterns) == 0 {
		t.Error("expected ExcludePatterns to have default values")
	}
	if cfg.Contributors == nil {
		t.Fatal("expected Contributors to be non-nil")
	}
	if !cfg.Contributors.Enabled {
		t.Error("expected Contributors.Enabled to be true by default")
	}
}

func TestDefaultGroups(t *testing.T) {
	groups := DefaultGroups()

	if len(groups) == 0 {
		t.Fatal("expected at least one default group")
	}

	// Check that feat and fix are in the groups
	patterns := make(map[string]string)
	for _, g := range groups {
		patterns[g.Pattern] = g.Label
	}

	if _, ok := patterns["^feat"]; !ok {
		t.Error("expected '^feat' pattern in default groups")
	}
	if _, ok := patterns["^fix"]; !ok {
		t.Error("expected '^fix' pattern in default groups")
	}
}

func TestDefaultExcludePatterns(t *testing.T) {
	patterns := DefaultExcludePatterns()

	if len(patterns) == 0 {
		t.Fatal("expected at least one default exclude pattern")
	}

	// Check for common patterns
	found := make(map[string]bool)
	for _, p := range patterns {
		found[p] = true
	}

	if !found["^Merge"] {
		t.Error("expected '^Merge' in default exclude patterns")
	}
	if !found["^WIP"] {
		t.Error("expected '^WIP' in default exclude patterns")
	}
}

func TestFromConfigStruct_Nil(t *testing.T) {
	cfg := FromConfigStruct(nil)

	if cfg == nil {
		t.Fatal("expected non-nil config from nil input")
	}
	if cfg.Mode != "versioned" {
		t.Errorf("Mode = %q, want 'versioned'", cfg.Mode)
	}
}

func TestFromConfigStruct_Full(t *testing.T) {
	input := &config.ChangelogGeneratorConfig{
		Enabled:       true,
		Mode:          "unified",
		ChangesDir:    "custom-changes",
		ChangelogPath: "CHANGES.md",
		Repository: &config.RepositoryConfig{
			Provider: "gitlab",
			Host:     "gitlab.com",
			Owner:    "mygroup",
			Repo:     "myproject",
		},
		Groups: []config.CommitGroupConfig{
			{Pattern: "^feat", Label: "Features", Icon: "rocket", Order: 0},
			{Pattern: "^fix", Label: "Bug Fixes", Order: 1},
		},
		ExcludePatterns: []string{"^WIP", "^SKIP"},
		Contributors: &config.ContributorsConfig{
			Enabled: false,
			Format:  "custom format",
		},
	}

	cfg := FromConfigStruct(input)

	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.Mode != "unified" {
		t.Errorf("Mode = %q, want 'unified'", cfg.Mode)
	}
	if cfg.ChangesDir != "custom-changes" {
		t.Errorf("ChangesDir = %q, want 'custom-changes'", cfg.ChangesDir)
	}
	if cfg.ChangelogPath != "CHANGES.md" {
		t.Errorf("ChangelogPath = %q, want 'CHANGES.md'", cfg.ChangelogPath)
	}

	// Repository
	if cfg.Repository == nil {
		t.Fatal("expected Repository to be non-nil")
	}
	if cfg.Repository.Provider != "gitlab" {
		t.Errorf("Repository.Provider = %q, want 'gitlab'", cfg.Repository.Provider)
	}
	if cfg.Repository.Host != "gitlab.com" {
		t.Errorf("Repository.Host = %q, want 'gitlab.com'", cfg.Repository.Host)
	}
	if cfg.Repository.Owner != "mygroup" {
		t.Errorf("Repository.Owner = %q, want 'mygroup'", cfg.Repository.Owner)
	}

	// Groups
	if len(cfg.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(cfg.Groups))
	}
	if cfg.Groups[0].Icon != "rocket" {
		t.Errorf("Groups[0].Icon = %q, want 'rocket'", cfg.Groups[0].Icon)
	}

	// ExcludePatterns
	if len(cfg.ExcludePatterns) != 2 {
		t.Errorf("expected 2 exclude patterns, got %d", len(cfg.ExcludePatterns))
	}

	// Contributors
	if cfg.Contributors == nil {
		t.Fatal("expected Contributors to be non-nil")
	}
	if cfg.Contributors.Enabled {
		t.Error("expected Contributors.Enabled to be false")
	}
	if cfg.Contributors.Format != "custom format" {
		t.Errorf("Contributors.Format = %q, want 'custom format'", cfg.Contributors.Format)
	}
}

func TestFromConfigStruct_Defaults(t *testing.T) {
	// Test with minimal config - should use defaults
	input := &config.ChangelogGeneratorConfig{
		Enabled: true,
	}

	cfg := FromConfigStruct(input)

	// Should use default mode
	if cfg.Mode != "versioned" {
		t.Errorf("Mode = %q, want 'versioned' (default)", cfg.Mode)
	}

	// Should use default changes dir
	if cfg.ChangesDir != ".changes" {
		t.Errorf("ChangesDir = %q, want '.changes' (default)", cfg.ChangesDir)
	}

	// Should have default groups
	if len(cfg.Groups) == 0 {
		t.Error("expected default groups")
	}

	// Should have default exclude patterns
	if len(cfg.ExcludePatterns) == 0 {
		t.Error("expected default exclude patterns")
	}

	// Repository should default to AutoDetect
	if cfg.Repository == nil {
		t.Fatal("expected Repository to be non-nil")
	}
	if !cfg.Repository.AutoDetect {
		t.Error("expected Repository.AutoDetect to be true by default")
	}
}

func TestGroupConfig(t *testing.T) {
	g := GroupConfig{
		Pattern: "^feat",
		Label:   "Features",
		Icon:    "rocket",
		Order:   0,
	}

	if g.Pattern != "^feat" {
		t.Errorf("Pattern = %q, want '^feat'", g.Pattern)
	}
	if g.Label != "Features" {
		t.Errorf("Label = %q, want 'Features'", g.Label)
	}
	if g.Icon != "rocket" {
		t.Errorf("Icon = %q, want 'rocket'", g.Icon)
	}
	if g.Order != 0 {
		t.Errorf("Order = %d, want 0", g.Order)
	}
}

func TestContributorsConfig(t *testing.T) {
	c := ContributorsConfig{
		Enabled: true,
		Format:  "custom",
	}

	if !c.Enabled {
		t.Error("expected Enabled to be true")
	}
	if c.Format != "custom" {
		t.Errorf("Format = %q, want 'custom'", c.Format)
	}
}

func TestRepositoryConfig(t *testing.T) {
	r := RepositoryConfig{
		Provider:   "github",
		Host:       "github.com",
		Owner:      "owner",
		Repo:       "repo",
		AutoDetect: false,
	}

	if r.Provider != "github" {
		t.Errorf("Provider = %q, want 'github'", r.Provider)
	}
	if r.Host != "github.com" {
		t.Errorf("Host = %q, want 'github.com'", r.Host)
	}
	if r.Owner != "owner" {
		t.Errorf("Owner = %q, want 'owner'", r.Owner)
	}
	if r.Repo != "repo" {
		t.Errorf("Repo = %q, want 'repo'", r.Repo)
	}
	if r.AutoDetect {
		t.Error("expected AutoDetect to be false")
	}
}

func TestFromConfigStruct_GroupIcons(t *testing.T) {
	// Test with GroupIcons - should use defaults and apply icons
	input := &config.ChangelogGeneratorConfig{
		Enabled: true,
		GroupIcons: map[string]string{
			"Enhancements":  "rocket",
			"Fixes":         "bug",
			"Documentation": "book",
		},
	}

	cfg := FromConfigStruct(input)

	// Should use default groups (not empty)
	if len(cfg.Groups) == 0 {
		t.Fatal("expected default groups")
	}

	// Verify icons are applied to matching labels
	iconsByLabel := make(map[string]string)
	for _, g := range cfg.Groups {
		iconsByLabel[g.Label] = g.Icon
	}

	if iconsByLabel["Enhancements"] != "rocket" {
		t.Errorf("Enhancements icon = %q, want 'rocket'", iconsByLabel["Enhancements"])
	}
	if iconsByLabel["Fixes"] != "bug" {
		t.Errorf("Fixes icon = %q, want 'bug'", iconsByLabel["Fixes"])
	}
	if iconsByLabel["Documentation"] != "book" {
		t.Errorf("Documentation icon = %q, want 'book'", iconsByLabel["Documentation"])
	}

	// Labels without icons should have empty icon
	if iconsByLabel["Refactors"] != "" {
		t.Errorf("Refactors icon = %q, want empty", iconsByLabel["Refactors"])
	}

	// GroupIcons should be stored in config
	if len(cfg.GroupIcons) != 3 {
		t.Errorf("expected 3 GroupIcons, got %d", len(cfg.GroupIcons))
	}
}

func TestFromConfigStruct_GroupsOverridesGroupIcons(t *testing.T) {
	// Test that Groups takes precedence over GroupIcons
	input := &config.ChangelogGeneratorConfig{
		Enabled: true,
		Groups: []config.CommitGroupConfig{
			{Pattern: "^feat", Label: "New Features"},
		},
		GroupIcons: map[string]string{
			"Enhancements": "rocket", // Should be ignored
		},
	}

	cfg := FromConfigStruct(input)

	// Should use provided groups, not defaults
	if len(cfg.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(cfg.Groups))
	}
	if cfg.Groups[0].Label != "New Features" {
		t.Errorf("Groups[0].Label = %q, want 'New Features'", cfg.Groups[0].Label)
	}

	// GroupIcons should not be applied when Groups is provided
	if cfg.Groups[0].Icon != "" {
		t.Errorf("expected empty icon when Groups provided, got %q", cfg.Groups[0].Icon)
	}
}

func TestFromConfigStruct_UseDefaultIcons(t *testing.T) {
	// Test that UseDefaultIcons applies default icons to all groups
	input := &config.ChangelogGeneratorConfig{
		Enabled:         true,
		UseDefaultIcons: true,
	}

	cfg := FromConfigStruct(input)

	// Should use default groups with icons applied
	if len(cfg.Groups) == 0 {
		t.Fatal("expected default groups")
	}

	// Verify default icons are applied to groups
	iconsByLabel := make(map[string]string)
	for _, g := range cfg.Groups {
		iconsByLabel[g.Label] = g.Icon
	}

	// Check specific default icons are applied
	if iconsByLabel["Enhancements"] != DefaultGroupIcons["Enhancements"] {
		t.Errorf("Enhancements icon = %q, want %q", iconsByLabel["Enhancements"], DefaultGroupIcons["Enhancements"])
	}
	if iconsByLabel["Fixes"] != DefaultGroupIcons["Fixes"] {
		t.Errorf("Fixes icon = %q, want %q", iconsByLabel["Fixes"], DefaultGroupIcons["Fixes"])
	}
	if iconsByLabel["Refactors"] != DefaultGroupIcons["Refactors"] {
		t.Errorf("Refactors icon = %q, want %q", iconsByLabel["Refactors"], DefaultGroupIcons["Refactors"])
	}
	if iconsByLabel["Documentation"] != DefaultGroupIcons["Documentation"] {
		t.Errorf("Documentation icon = %q, want %q", iconsByLabel["Documentation"], DefaultGroupIcons["Documentation"])
	}

	// Verify UseDefaultIcons flag is set
	if !cfg.UseDefaultIcons {
		t.Error("expected UseDefaultIcons to be true")
	}

	// Verify default contributor icon is applied
	if cfg.Contributors == nil {
		t.Fatal("expected Contributors to be non-nil")
	}
	if cfg.Contributors.Icon != DefaultContributorIcon {
		t.Errorf("Contributors.Icon = %q, want %q", cfg.Contributors.Icon, DefaultContributorIcon)
	}
}

func TestFromConfigStruct_UseDefaultIconsWithOverrides(t *testing.T) {
	// Test that user-defined GroupIcons override defaults when UseDefaultIcons is true
	input := &config.ChangelogGeneratorConfig{
		Enabled:         true,
		UseDefaultIcons: true,
		GroupIcons: map[string]string{
			"Enhancements": "custom-rocket", // Override default
			"Fixes":        "custom-bug",    // Override default
		},
	}

	cfg := FromConfigStruct(input)

	iconsByLabel := make(map[string]string)
	for _, g := range cfg.Groups {
		iconsByLabel[g.Label] = g.Icon
	}

	// User-defined icons should override defaults
	if iconsByLabel["Enhancements"] != "custom-rocket" {
		t.Errorf("Enhancements icon = %q, want 'custom-rocket'", iconsByLabel["Enhancements"])
	}
	if iconsByLabel["Fixes"] != "custom-bug" {
		t.Errorf("Fixes icon = %q, want 'custom-bug'", iconsByLabel["Fixes"])
	}

	// Non-overridden groups should use defaults
	if iconsByLabel["Refactors"] != DefaultGroupIcons["Refactors"] {
		t.Errorf("Refactors icon = %q, want %q (default)", iconsByLabel["Refactors"], DefaultGroupIcons["Refactors"])
	}
	if iconsByLabel["Documentation"] != DefaultGroupIcons["Documentation"] {
		t.Errorf("Documentation icon = %q, want %q (default)", iconsByLabel["Documentation"], DefaultGroupIcons["Documentation"])
	}
}

func TestFromConfigStruct_UseDefaultIconsContributorOverride(t *testing.T) {
	// Test that user-defined contributor icon overrides default
	input := &config.ChangelogGeneratorConfig{
		Enabled:         true,
		UseDefaultIcons: true,
		Contributors: &config.ContributorsConfig{
			Enabled: true,
			Icon:    "custom-heart",
		},
	}

	cfg := FromConfigStruct(input)

	// User-defined contributor icon should override default
	if cfg.Contributors.Icon != "custom-heart" {
		t.Errorf("Contributors.Icon = %q, want 'custom-heart'", cfg.Contributors.Icon)
	}
}

func TestFromConfigStruct_UseDefaultIconsFalse(t *testing.T) {
	// Test that UseDefaultIcons=false doesn't apply icons
	input := &config.ChangelogGeneratorConfig{
		Enabled:         true,
		UseDefaultIcons: false,
	}

	cfg := FromConfigStruct(input)

	// Groups should have no icons
	for _, g := range cfg.Groups {
		if g.Icon != "" {
			t.Errorf("Group %q should have no icon when UseDefaultIcons is false, got %q", g.Label, g.Icon)
		}
	}

	// Contributors should have no icon
	if cfg.Contributors.Icon != "" {
		t.Errorf("Contributors.Icon should be empty when UseDefaultIcons is false, got %q", cfg.Contributors.Icon)
	}
}

func TestDefaultGroupIcons(t *testing.T) {
	// Verify all default group labels have corresponding icons
	groups := DefaultGroups()
	for _, g := range groups {
		if _, ok := DefaultGroupIcons[g.Label]; !ok {
			t.Errorf("missing default icon for group label %q", g.Label)
		}
	}

	// Verify expected number of default icons
	expectedCount := len(groups)
	if len(DefaultGroupIcons) != expectedCount {
		t.Errorf("DefaultGroupIcons has %d entries, want %d", len(DefaultGroupIcons), expectedCount)
	}
}

func TestDefaultContributorIcon(t *testing.T) {
	// Verify default contributor icon is set
	if DefaultContributorIcon == "" {
		t.Error("DefaultContributorIcon should not be empty")
	}
}

func TestDefaultBreakingChangesIcon(t *testing.T) {
	// Verify default breaking changes icon is set
	if DefaultBreakingChangesIcon == "" {
		t.Error("DefaultBreakingChangesIcon should not be empty")
	}
}

func TestFromConfigStruct_BreakingChangesIcon(t *testing.T) {
	tests := []struct {
		name     string
		input    *config.ChangelogGeneratorConfig
		wantIcon string
	}{
		{
			name: "explicit icon overrides default",
			input: &config.ChangelogGeneratorConfig{
				Enabled:             true,
				UseDefaultIcons:     true,
				BreakingChangesIcon: "custom-warning",
			},
			wantIcon: "custom-warning",
		},
		{
			name: "UseDefaultIcons applies default icon",
			input: &config.ChangelogGeneratorConfig{
				Enabled:         true,
				UseDefaultIcons: true,
			},
			wantIcon: DefaultBreakingChangesIcon,
		},
		{
			name: "no UseDefaultIcons and no explicit icon results in empty",
			input: &config.ChangelogGeneratorConfig{
				Enabled:         true,
				UseDefaultIcons: false,
			},
			wantIcon: "",
		},
		{
			name: "explicit icon without UseDefaultIcons",
			input: &config.ChangelogGeneratorConfig{
				Enabled:             true,
				UseDefaultIcons:     false,
				BreakingChangesIcon: "explicit-icon",
			},
			wantIcon: "explicit-icon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := FromConfigStruct(tt.input)
			if cfg.BreakingChangesIcon != tt.wantIcon {
				t.Errorf("BreakingChangesIcon = %q, want %q", cfg.BreakingChangesIcon, tt.wantIcon)
			}
		})
	}
}
