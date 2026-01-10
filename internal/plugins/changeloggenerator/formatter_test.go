package changeloggenerator

import (
	"strings"
	"testing"
)

func TestNewFormatter(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "grouped format",
			format:  "grouped",
			wantErr: false,
		},
		{
			name:    "keepachangelog format",
			format:  "keepachangelog",
			wantErr: false,
		},
		{
			name:    "github format",
			format:  "github",
			wantErr: false,
		},
		{
			name:    "minimal format",
			format:  "minimal",
			wantErr: false,
		},
		{
			name:    "unknown format",
			format:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewFormatter(tt.format, cfg)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error for unknown format")
				}
				if !strings.Contains(err.Error(), "unknown changelog format") {
					t.Errorf("error = %v, want error mentioning 'unknown changelog format'", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if formatter == nil {
					t.Error("expected non-nil formatter")
				}
			}
		})
	}
}

func TestGroupedFormatter_FormatChangelog(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GroupedFormatter{config: cfg}

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add feature"},
					Type:        "feat",
					Scope:       "core",
					Description: "add feature",
				},
				GroupLabel: "Enhancements",
				GroupOrder: 0,
			},
		},
		"Fixes": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "def456", ShortHash: "def456", Subject: "fix: fix bug"},
					Type:        "fix",
					Description: "fix bug",
				},
				GroupLabel: "Fixes",
				GroupOrder: 1,
			},
		},
	}
	sortedKeys := []string{"Enhancements", "Fixes"}

	result := formatter.FormatChangelog("v1.0.0", "v0.9.0", grouped, sortedKeys, remote)

	// Check version header with "v" prefix
	if !strings.Contains(result, "## v1.0.0") {
		t.Error("expected version header with 'v' prefix")
	}

	// Note: Full Changelog link is now generated in generator.go, not in formatters

	// Check section headers
	if !strings.Contains(result, "### Enhancements") {
		t.Error("expected Enhancements section")
	}
	if !strings.Contains(result, "### Fixes") {
		t.Error("expected Fixes section")
	}

	// Check commit entries
	if !strings.Contains(result, "**core:** add feature") {
		t.Error("expected commit with scope")
	}
	if !strings.Contains(result, "fix bug") {
		t.Error("expected commit description")
	}

	// Check commit links
	if !strings.Contains(result, "[abc123](https://github.com/testowner/testrepo/commit/abc123)") {
		t.Error("expected commit link")
	}
}

func TestGroupedFormatter_WithIcons(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GroupedFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add feature"},
					Type:        "feat",
					Description: "add feature",
				},
				GroupLabel: "Enhancements",
				GroupIcon:  "🚀",
				GroupOrder: 0,
			},
		},
	}
	sortedKeys := []string{"Enhancements"}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Check section header with icon
	if !strings.Contains(result, "### 🚀 Enhancements") {
		t.Error("expected section header with icon")
	}
}

func TestKeepAChangelogFormatter_FormatChangelog(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &KeepAChangelogFormatter{config: cfg}

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add feature"},
					Type:        "feat",
					Scope:       "core",
					Description: "add feature",
				},
				GroupLabel: "Enhancements",
			},
		},
		"Fixes": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "def456", ShortHash: "def456", Subject: "fix: fix bug"},
					Type:        "fix",
					Description: "fix bug",
				},
				GroupLabel: "Fixes",
			},
		},
	}
	sortedKeys := []string{"Enhancements", "Fixes"}

	result := formatter.FormatChangelog("v1.0.0", "v0.9.0", grouped, sortedKeys, remote)

	// Check version header WITHOUT "v" prefix and WITH brackets
	if !strings.Contains(result, "## [1.0.0]") {
		t.Error("expected version header with brackets and no 'v' prefix")
	}

	// Note: Full Changelog link is now generated in generator.go, not in formatters

	// Check standard Keep a Changelog sections
	if !strings.Contains(result, "### Added") {
		t.Error("expected 'Added' section for feat commits")
	}
	if !strings.Contains(result, "### Fixed") {
		t.Error("expected 'Fixed' section for fix commits")
	}

	// Should NOT have original group labels
	if strings.Contains(result, "Enhancements") {
		t.Error("Keep a Changelog format should not use custom group labels")
	}
	if strings.Contains(result, "Fixes") {
		t.Error("Keep a Changelog format should not use custom group labels")
	}

	// Check commit entries still have links
	if !strings.Contains(result, "[abc123](https://github.com/testowner/testrepo/commit/abc123)") {
		t.Error("expected commit link")
	}
}

func TestKeepAChangelogFormatter_CommitTypeMapping(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &KeepAChangelogFormatter{config: cfg}

	tests := []struct {
		commitType      string
		breaking        bool
		expectedSection string
	}{
		{"feat", false, "Added"},
		{"fix", false, "Fixed"},
		{"refactor", false, "Changed"},
		{"perf", false, "Changed"},
		{"style", false, "Changed"},
		{"docs", false, ""},  // Skipped
		{"test", false, ""},  // Skipped
		{"chore", false, ""}, // Skipped
		{"ci", false, ""},    // Skipped
		{"build", false, ""}, // Skipped
		{"revert", false, "Removed"},
		{"feat", true, "Breaking Changes"},
		{"fix", true, "Breaking Changes"},
	}

	for _, tt := range tests {
		t.Run(tt.commitType, func(t *testing.T) {
			commit := &GroupedCommit{
				ParsedCommit: &ParsedCommit{
					Type:     tt.commitType,
					Breaking: tt.breaking,
				},
			}

			section := formatter.mapTypeToSection(commit)
			if section != tt.expectedSection {
				t.Errorf("mapTypeToSection(%q, breaking=%v) = %q, want %q",
					tt.commitType, tt.breaking, section, tt.expectedSection)
			}
		})
	}
}

func TestKeepAChangelogFormatter_SectionOrder(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &KeepAChangelogFormatter{config: cfg}

	// Create commits of different types
	grouped := map[string][]*GroupedCommit{
		"custom": {
			{
				ParsedCommit: &ParsedCommit{
					Type:        "feat",
					Description: "feature 1",
					Breaking:    true,
					CommitInfo:  CommitInfo{ShortHash: "aaa"},
				},
			},
			{
				ParsedCommit: &ParsedCommit{
					Type:        "fix",
					Description: "fix 1",
					CommitInfo:  CommitInfo{ShortHash: "bbb"},
				},
			},
			{
				ParsedCommit: &ParsedCommit{
					Type:        "feat",
					Description: "feature 2",
					CommitInfo:  CommitInfo{ShortHash: "ccc"},
				},
			},
			{
				ParsedCommit: &ParsedCommit{
					Type:        "refactor",
					Description: "refactor 1",
					CommitInfo:  CommitInfo{ShortHash: "ddd"},
				},
			},
		},
	}
	sortedKeys := []string{"custom"}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Find positions of each section
	breakingPos := strings.Index(result, "### Breaking Changes")
	addedPos := strings.Index(result, "### Added")
	changedPos := strings.Index(result, "### Changed")
	fixedPos := strings.Index(result, "### Fixed")

	// Breaking Changes should come first
	if breakingPos == -1 {
		t.Error("expected Breaking Changes section")
	}
	if addedPos == -1 {
		t.Error("expected Added section")
	}
	if changedPos == -1 {
		t.Error("expected Changed section")
	}
	if fixedPos == -1 {
		t.Error("expected Fixed section")
	}

	// Verify order: Breaking -> Added -> Changed -> Fixed
	if breakingPos > addedPos {
		t.Error("Breaking Changes should come before Added")
	}
	if addedPos > changedPos {
		t.Error("Added should come before Changed")
	}
	if changedPos > fixedPos {
		t.Error("Changed should come before Fixed")
	}
}

func TestKeepAChangelogFormatter_UnknownTypeMapping(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &KeepAChangelogFormatter{config: cfg}

	tests := []struct {
		name            string
		commitType      string
		expectedSection string
	}{
		{
			name:            "unknown type with content",
			commitType:      "custom",
			expectedSection: "Changed",
		},
		{
			name:            "empty type (non-conventional)",
			commitType:      "",
			expectedSection: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commit := &GroupedCommit{
				ParsedCommit: &ParsedCommit{
					Type:        tt.commitType,
					Description: "some change",
				},
			}

			section := formatter.mapTypeToSection(commit)
			if section != tt.expectedSection {
				t.Errorf("mapTypeToSection(%q) = %q, want %q",
					tt.commitType, section, tt.expectedSection)
			}
		})
	}
}

func TestGroupedFormatter_WithoutRemote(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GroupedFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add feature"},
					Type:        "feat",
					Description: "add feature",
				},
				GroupLabel: "Enhancements",
				GroupOrder: 0,
			},
		},
	}
	sortedKeys := []string{"Enhancements"}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Without remote, commit entries should just have the description without links
	if !strings.Contains(result, "add feature") {
		t.Error("expected description in output")
	}
	// Should not contain full URL
	if strings.Contains(result, "https://") {
		t.Error("should not contain URLs without remote")
	}
}

func TestGroupedFormatter_BreakingChangesOnly(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakingChangesIcon = DefaultBreakingChangesIcon
	formatter := &GroupedFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123", Subject: "feat(api)!: breaking API change"},
					Type:        "feat",
					Scope:       "api",
					Description: "breaking API change",
					Breaking:    true,
				},
				GroupLabel: "Enhancements",
				GroupOrder: 0,
			},
		},
	}
	sortedKeys := []string{"Enhancements"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Breaking change should appear in dedicated section with icon
	if !strings.Contains(result, "### "+DefaultBreakingChangesIcon+" Breaking Changes") {
		t.Error("expected Breaking Changes section header with icon")
	}

	// Breaking change commit should be included
	if !strings.Contains(result, "**api:** breaking API change") {
		t.Error("expected breaking change entry with scope")
	}

	// Should NOT have the original group label section when only breaking changes exist
	if strings.Contains(result, "### Enhancements") {
		t.Error("should not have empty 'Enhancements' section when only breaking changes exist")
	}
}

func TestGroupedFormatter_MixedBreakingAndRegular(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakingChangesIcon = DefaultBreakingChangesIcon
	formatter := &GroupedFormatter{config: cfg}

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "break1", ShortHash: "break1", Subject: "feat(api)!: Remove deprecated endpoints"},
					Type:        "feat",
					Scope:       "api",
					Description: "Remove deprecated endpoints",
					Breaking:    true,
				},
				GroupLabel: "Enhancements",
				GroupIcon:  "rocket",
				GroupOrder: 0,
			},
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "feat1", ShortHash: "feat1", Subject: "feat(core): Add new caching layer"},
					Type:        "feat",
					Scope:       "core",
					Description: "Add new caching layer",
					Breaking:    false,
				},
				GroupLabel: "Enhancements",
				GroupIcon:  "rocket",
				GroupOrder: 0,
			},
		},
		"Fixes": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "fix1", ShortHash: "fix1", Subject: "fix: Fix memory leak"},
					Type:        "fix",
					Description: "Fix memory leak",
					Breaking:    false,
				},
				GroupLabel: "Fixes",
				GroupIcon:  "bug",
				GroupOrder: 1,
			},
		},
	}
	sortedKeys := []string{"Enhancements", "Fixes"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, remote)

	// Breaking Changes section should exist
	if !strings.Contains(result, "### "+DefaultBreakingChangesIcon+" Breaking Changes") {
		t.Error("expected Breaking Changes section header")
	}

	// Regular sections should exist
	if !strings.Contains(result, "### rocket Enhancements") {
		t.Error("expected Enhancements section")
	}
	if !strings.Contains(result, "### bug Fixes") {
		t.Error("expected Fixes section")
	}

	// Breaking change should be in Breaking Changes section
	if !strings.Contains(result, "Remove deprecated endpoints") {
		t.Error("expected breaking change entry")
	}

	// Regular changes should be in their sections
	if !strings.Contains(result, "Add new caching layer") {
		t.Error("expected regular enhancement entry")
	}
	if !strings.Contains(result, "Fix memory leak") {
		t.Error("expected fix entry")
	}

	// Breaking Changes should appear before other sections
	breakingIdx := strings.Index(result, "### "+DefaultBreakingChangesIcon+" Breaking Changes")
	enhancementsIdx := strings.Index(result, "### rocket Enhancements")
	fixesIdx := strings.Index(result, "### bug Fixes")

	if breakingIdx > enhancementsIdx {
		t.Error("Breaking Changes section should appear before Enhancements")
	}
	if breakingIdx > fixesIdx {
		t.Error("Breaking Changes section should appear before Fixes")
	}
}

func TestGroupedFormatter_CustomBreakingChangesIcon(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakingChangesIcon = "CUSTOM"
	formatter := &GroupedFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123", Subject: "feat!: breaking change"},
					Type:        "feat",
					Description: "breaking change",
					Breaking:    true,
				},
				GroupLabel: "Enhancements",
			},
		},
	}
	sortedKeys := []string{"Enhancements"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Should use custom icon
	if !strings.Contains(result, "### CUSTOM Breaking Changes") {
		t.Error("expected custom breaking changes icon in header")
	}
}

func TestGroupedFormatter_NoBreakingChangesIcon(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakingChangesIcon = "" // No icon
	formatter := &GroupedFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123", Subject: "feat!: breaking change"},
					Type:        "feat",
					Description: "breaking change",
					Breaking:    true,
				},
				GroupLabel: "Enhancements",
			},
		},
	}
	sortedKeys := []string{"Enhancements"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Header should be plain "Breaking Changes" without icon
	if !strings.Contains(result, "### Breaking Changes\n") {
		t.Error("expected plain 'Breaking Changes' header without icon")
	}
	// Should not have a space before "Breaking"
	if strings.Contains(result, "###  Breaking") {
		t.Error("should not have extra space before 'Breaking'")
	}
}

func TestGroupedFormatter_BreakingChangesExcludedFromRegularGroups(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakingChangesIcon = DefaultBreakingChangesIcon
	formatter := &GroupedFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "break1", ShortHash: "break1", Subject: "feat!: breaking feature"},
					Type:        "feat",
					Description: "breaking feature",
					Breaking:    true,
				},
				GroupLabel: "Enhancements",
			},
		},
	}
	sortedKeys := []string{"Enhancements"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Breaking change should appear in Breaking Changes section
	if !strings.Contains(result, "### "+DefaultBreakingChangesIcon+" Breaking Changes") {
		t.Error("expected Breaking Changes section")
	}
	if !strings.Contains(result, "breaking feature") {
		t.Error("expected breaking feature entry")
	}

	// The Enhancements section should NOT appear (no regular commits)
	if strings.Contains(result, "### Enhancements") {
		t.Error("empty Enhancements section should not appear when all commits are breaking")
	}
}

func TestGroupedFormatter_MultipleBreakingChanges(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakingChangesIcon = DefaultBreakingChangesIcon
	formatter := &GroupedFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "break1", ShortHash: "break1", Subject: "feat(api)!: Remove deprecated endpoints"},
					Type:        "feat",
					Scope:       "api",
					Description: "Remove deprecated endpoints",
					Breaking:    true,
				},
				GroupLabel: "Enhancements",
			},
		},
		"Fixes": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "break2", ShortHash: "break2", Subject: "fix!: Change error response format"},
					Type:        "fix",
					Description: "Change error response format",
					Breaking:    true,
				},
				GroupLabel: "Fixes",
			},
		},
	}
	sortedKeys := []string{"Enhancements", "Fixes"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Both breaking changes should appear in the Breaking Changes section
	if !strings.Contains(result, "Remove deprecated endpoints") {
		t.Error("expected first breaking change entry")
	}
	if !strings.Contains(result, "Change error response format") {
		t.Error("expected second breaking change entry")
	}

	// There should only be ONE Breaking Changes section header
	count := strings.Count(result, "### "+DefaultBreakingChangesIcon+" Breaking Changes")
	if count != 1 {
		t.Errorf("expected exactly one Breaking Changes header, got %d", count)
	}

	// Neither Enhancements nor Fixes sections should appear (all are breaking)
	if strings.Contains(result, "### Enhancements") {
		t.Error("empty Enhancements section should not appear")
	}
	if strings.Contains(result, "### Fixes") {
		t.Error("empty Fixes section should not appear")
	}
}
