package changeloggenerator

import (
	"strings"
	"testing"
)

func TestGitHubFormatter_FormatChangelog(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GitHubFormatter{config: cfg}

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
					CommitInfo: CommitInfo{
						Hash:        "abc123full",
						ShortHash:   "abc123",
						Subject:     "feat(core): add feature",
						Author:      "John Doe",
						AuthorEmail: "johndoe@users.noreply.github.com",
					},
					Type:        "feat",
					Scope:       "core",
					Description: "add feature",
					PRNumber:    "123",
				},
				GroupLabel: "Enhancements",
				GroupOrder: 0,
			},
		},
		"Fixes": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "def456full",
						ShortHash:   "def456",
						Subject:     "fix: fix bug",
						Author:      "Jane Doe",
						AuthorEmail: "janedoe@users.noreply.github.com",
					},
					Type:        "fix",
					Description: "fix bug",
					PRNumber:    "456",
				},
				GroupLabel: "Fixes",
				GroupOrder: 1,
			},
		},
	}
	sortedKeys := []string{"Enhancements", "Fixes"}

	result := formatter.FormatChangelog("v1.2.0", "v1.1.0", grouped, sortedKeys, remote)

	// Check version header with "v" prefix
	if !strings.Contains(result, "## v1.2.0") {
		t.Error("expected version header with 'v' prefix")
	}

	// Check single "What's Changed" section
	if !strings.Contains(result, "### What's Changed") {
		t.Error("expected 'What's Changed' section header")
	}

	// Should NOT have individual group headers
	if strings.Contains(result, "### Enhancements") {
		t.Error("GitHub format should not have Enhancements section")
	}
	if strings.Contains(result, "### Fixes") {
		t.Error("GitHub format should not have Fixes section")
	}

	// Check * bullet style (not -)
	if !strings.Contains(result, "* **core:**") {
		t.Error("expected * bullet with scope")
	}

	// Check author attribution
	if !strings.Contains(result, "by @johndoe") {
		t.Error("expected author attribution '@johndoe'")
	}
	if !strings.Contains(result, "by @janedoe") {
		t.Error("expected author attribution '@janedoe'")
	}

	// Check PR references
	if !strings.Contains(result, "in #123") {
		t.Error("expected PR reference '#123'")
	}
	if !strings.Contains(result, "in #456") {
		t.Error("expected PR reference '#456'")
	}
}

func TestGitHubFormatter_WithScope(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Features": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "abc123",
						ShortHash:   "abc123",
						Subject:     "feat(api): update rate limiting",
						Author:      "Contributor",
						AuthorEmail: "contributor@example.com",
					},
					Type:        "feat",
					Scope:       "api",
					Description: "update rate limiting",
				},
				GroupLabel: "Features",
			},
		},
	}
	sortedKeys := []string{"Features"}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Scope should be bold
	if !strings.Contains(result, "* **api:** update rate limiting") {
		t.Error("expected scope to be bold: **api:**")
	}
}

func TestGitHubFormatter_WithoutScope(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Features": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "abc123",
						ShortHash:   "abc123",
						Subject:     "feat: add new feature",
						Author:      "Contributor",
						AuthorEmail: "contributor@example.com",
					},
					Type:        "feat",
					Scope:       "", // No scope
					Description: "add new feature",
				},
				GroupLabel: "Features",
			},
		},
	}
	sortedKeys := []string{"Features"}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Should have the entry without scope
	if !strings.Contains(result, "* add new feature") {
		t.Error("expected entry without scope")
	}
	// Should not have **:** pattern for empty scope
	if strings.Contains(result, "**:**") {
		t.Error("should not have empty scope pattern")
	}
}

func TestGitHubFormatter_WithPRNumber(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Features": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "abc123",
						ShortHash:   "abc123",
						Subject:     "feat: add feature (#789)",
						Author:      "Dev",
						AuthorEmail: "12345+devuser@users.noreply.github.com",
					},
					Type:        "feat",
					Description: "add feature",
					PRNumber:    "789",
				},
				GroupLabel: "Features",
			},
		},
	}
	sortedKeys := []string{"Features"}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Check PR reference is included
	if !strings.Contains(result, "in #789") {
		t.Error("expected PR reference 'in #789'")
	}

	// Check username extracted from GitHub noreply email with ID prefix
	if !strings.Contains(result, "by @devuser") {
		t.Error("expected author '@devuser' extracted from noreply email")
	}
}

func TestGitHubFormatter_WithoutPRNumber(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Features": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "abc123",
						ShortHash:   "abc123",
						Subject:     "feat: add feature",
						Author:      "Dev User",
						AuthorEmail: "dev@example.com",
					},
					Type:        "feat",
					Description: "add feature",
					PRNumber:    "", // No PR number
				},
				GroupLabel: "Features",
			},
		},
	}
	sortedKeys := []string{"Features"}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Should not have "in #" if no PR number
	if strings.Contains(result, "in #") {
		t.Error("should not have PR reference without PR number")
	}

	// Should still have author (fallback to lowercased name without spaces)
	if !strings.Contains(result, "by @devuser") {
		t.Error("expected fallback author '@devuser'")
	}
}

func TestGitHubFormatter_WithoutRemote(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Features": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "abc123",
						ShortHash:   "abc123",
						Subject:     "feat(core): add feature",
						Author:      "John Doe",
						AuthorEmail: "john@example.com",
					},
					Type:        "feat",
					Scope:       "core",
					Description: "add feature",
					PRNumber:    "123",
				},
				GroupLabel: "Features",
			},
		},
	}
	sortedKeys := []string{"Features"}

	// Pass nil remote - format should still work
	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Should still include basic formatting
	if !strings.Contains(result, "## v1.0.0") {
		t.Error("expected version header")
	}
	if !strings.Contains(result, "### What's Changed") {
		t.Error("expected What's Changed section")
	}
	if !strings.Contains(result, "* **core:** add feature") {
		t.Error("expected formatted entry")
	}
	// Author should still be included (using fallback)
	if !strings.Contains(result, "by @johndoe") {
		t.Error("expected author attribution")
	}
	// PR should still be included
	if !strings.Contains(result, "in #123") {
		t.Error("expected PR reference")
	}
}

func TestGitHubFormatter_FlattenedGroups(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GitHubFormatter{config: cfg}

	// Multiple groups with multiple commits each
	grouped := map[string][]*GroupedCommit{
		"Features": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "aaa111",
						ShortHash:   "aaa111",
						Subject:     "feat: feature 1",
						Author:      "User A",
						AuthorEmail: "usera@users.noreply.github.com",
					},
					Type:        "feat",
					Description: "feature 1",
					PRNumber:    "1",
				},
				GroupLabel: "Features",
				GroupOrder: 0,
			},
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "bbb222",
						ShortHash:   "bbb222",
						Subject:     "feat: feature 2",
						Author:      "User B",
						AuthorEmail: "userb@users.noreply.github.com",
					},
					Type:        "feat",
					Description: "feature 2",
					PRNumber:    "2",
				},
				GroupLabel: "Features",
				GroupOrder: 0,
			},
		},
		"Fixes": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "ccc333",
						ShortHash:   "ccc333",
						Subject:     "fix: bug fix 1",
						Author:      "User C",
						AuthorEmail: "userc@users.noreply.github.com",
					},
					Type:        "fix",
					Description: "bug fix 1",
					PRNumber:    "3",
				},
				GroupLabel: "Fixes",
				GroupOrder: 1,
			},
		},
		"Other": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "ddd444",
						ShortHash:   "ddd444",
						Subject:     "chore: cleanup",
						Author:      "User D",
						AuthorEmail: "userd@users.noreply.github.com",
					},
					Type:        "chore",
					Description: "cleanup",
					PRNumber:    "4",
				},
				GroupLabel: "Other",
				GroupOrder: 2,
			},
		},
	}
	sortedKeys := []string{"Features", "Fixes", "Other"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Count "What's Changed" sections - should be exactly 1
	whatsChangedCount := strings.Count(result, "### What's Changed")
	if whatsChangedCount != 1 {
		t.Errorf("expected exactly 1 'What's Changed' section, got %d", whatsChangedCount)
	}

	// All commits should be present in a single flat list
	expectedEntries := []string{
		"* feature 1 by @usera in #1",
		"* feature 2 by @userb in #2",
		"* bug fix 1 by @userc in #3",
		"* cleanup by @userd in #4",
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(result, entry) {
			t.Errorf("expected entry %q in output", entry)
		}
	}

	// Count total * bullets - should be 4
	bulletCount := strings.Count(result, "\n* ")
	if bulletCount != 4 {
		t.Errorf("expected 4 bullet entries, got %d", bulletCount)
	}
}

func TestGitHubFormatter_GitLabNoreplyEmail(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Features": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "abc123",
						ShortHash:   "abc123",
						Subject:     "feat: gitlab feature",
						Author:      "GitLab User",
						AuthorEmail: "gitlabuser@noreply.gitlab.com",
					},
					Type:        "feat",
					Description: "gitlab feature",
				},
				GroupLabel: "Features",
			},
		},
	}
	sortedKeys := []string{"Features"}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Check username extracted from GitLab noreply email
	if !strings.Contains(result, "by @gitlabuser") {
		t.Error("expected author '@gitlabuser' extracted from GitLab noreply email")
	}
}

func TestGitHubFormatter_CodebergNoreplyEmail(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Features": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "abc123",
						ShortHash:   "abc123",
						Subject:     "feat: codeberg feature",
						Author:      "Codeberg User",
						AuthorEmail: "codeberguser@noreply.codeberg.org",
					},
					Type:        "feat",
					Description: "codeberg feature",
				},
				GroupLabel: "Features",
			},
		},
	}
	sortedKeys := []string{"Features"}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Check username extracted from Codeberg noreply email
	if !strings.Contains(result, "by @codeberguser") {
		t.Error("expected author '@codeberguser' extracted from Codeberg noreply email")
	}
}

func TestGitHubFormatter_BreakingChange(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakingChangesIcon = DefaultBreakingChangesIcon
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Breaking": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "abc123",
						ShortHash:   "abc123",
						Subject:     "feat(api)!: breaking API change",
						Author:      "Dev",
						AuthorEmail: "dev@users.noreply.github.com",
					},
					Type:        "feat",
					Scope:       "api",
					Description: "breaking API change",
					Breaking:    true,
					PRNumber:    "999",
				},
				GroupLabel: "Breaking",
			},
		},
	}
	sortedKeys := []string{"Breaking"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Breaking change should appear in dedicated section
	if !strings.Contains(result, "### ⚠️ Breaking Changes") {
		t.Error("expected '⚠️ Breaking Changes' section header")
	}

	// Breaking change commit should be included
	if !strings.Contains(result, "* **api:** breaking API change by @dev in #999") {
		t.Error("expected breaking change entry")
	}

	// Should NOT have "What's Changed" section when only breaking changes exist
	if strings.Contains(result, "### What's Changed") {
		t.Error("should not have 'What's Changed' section when only breaking changes exist")
	}
}

func TestGitHubFormatter_MixedBreakingAndRegular(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakingChangesIcon = DefaultBreakingChangesIcon
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Breaking": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "break1",
						ShortHash:   "break1",
						Subject:     "feat(api)!: Remove deprecated endpoints",
						Author:      "Maintainer",
						AuthorEmail: "maintainer@users.noreply.github.com",
					},
					Type:        "feat",
					Scope:       "api",
					Description: "Remove deprecated endpoints",
					Breaking:    true,
					PRNumber:    "100",
				},
				GroupLabel: "Breaking",
				GroupOrder: 0,
			},
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "break2",
						ShortHash:   "break2",
						Subject:     "feat!: Change authentication flow",
						Author:      "Dev",
						AuthorEmail: "dev@users.noreply.github.com",
					},
					Type:        "feat",
					Description: "Change authentication flow",
					Breaking:    true,
					PRNumber:    "101",
				},
				GroupLabel: "Breaking",
				GroupOrder: 0,
			},
		},
		"Features": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "feat1",
						ShortHash:   "feat1",
						Subject:     "feat(core): Add new caching layer",
						Author:      "John Doe",
						AuthorEmail: "johndoe@users.noreply.github.com",
					},
					Type:        "feat",
					Scope:       "core",
					Description: "Add new caching layer",
					Breaking:    false,
					PRNumber:    "123",
				},
				GroupLabel: "Features",
				GroupOrder: 1,
			},
		},
		"Fixes": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "fix1",
						ShortHash:   "fix1",
						Subject:     "fix: Fix memory leak in parser",
						Author:      "Jane Doe",
						AuthorEmail: "janedoe@users.noreply.github.com",
					},
					Type:        "fix",
					Description: "Fix memory leak in parser",
					Breaking:    false,
					PRNumber:    "456",
				},
				GroupLabel: "Fixes",
				GroupOrder: 2,
			},
		},
	}
	sortedKeys := []string{"Breaking", "Features", "Fixes"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Extract sections for verification
	breakingSection, changedSection := extractGitHubSections(result)

	t.Run("has both section headers", func(t *testing.T) {
		assertContains(t, result, "### ⚠️ Breaking Changes", "expected Breaking Changes header")
		assertContains(t, result, "### What's Changed", "expected What's Changed header")
	})

	t.Run("section order", func(t *testing.T) {
		breakingIdx := strings.Index(result, "### ⚠️ Breaking Changes")
		whatsChangedIdx := strings.Index(result, "### What's Changed")
		if breakingIdx >= whatsChangedIdx {
			t.Error("Breaking Changes section should appear before What's Changed section")
		}
	})

	t.Run("breaking changes in correct section", func(t *testing.T) {
		assertContains(t, breakingSection, "Remove deprecated endpoints", "breaking change missing from section")
		assertContains(t, breakingSection, "Change authentication flow", "breaking change missing from section")
		assertNotContains(t, changedSection, "Remove deprecated endpoints", "breaking change in wrong section")
		assertNotContains(t, changedSection, "Change authentication flow", "breaking change in wrong section")
	})

	t.Run("regular changes in correct section", func(t *testing.T) {
		assertContains(t, changedSection, "Add new caching layer", "regular change missing from section")
		assertContains(t, changedSection, "Fix memory leak in parser", "regular change missing from section")
	})
}

// extractGitHubSections splits the result into breaking and changed sections.
func extractGitHubSections(result string) (breakingSection, changedSection string) {
	sections := strings.SplitSeq(result, "###")
	for section := range sections {
		if strings.Contains(section, "⚠️ Breaking Changes") {
			breakingSection = section
		}
		if strings.Contains(section, "What's Changed") {
			changedSection = section
		}
	}
	return
}

// assertContains checks if haystack contains needle.
func assertContains(t *testing.T, haystack, needle, msg string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("%s: expected %q in output", msg, needle)
	}
}

// assertNotContains checks if haystack does not contain needle.
func assertNotContains(t *testing.T, haystack, needle, msg string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("%s: unexpected %q in output", msg, needle)
	}
}

func TestGitHubFormatter_EmptyGroups(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &GitHubFormatter{config: cfg}

	// Empty groups
	grouped := map[string][]*GroupedCommit{}
	sortedKeys := []string{}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Should still have header and section
	if !strings.Contains(result, "## v1.0.0") {
		t.Error("expected version header")
	}
	if !strings.Contains(result, "### What's Changed") {
		t.Error("expected What's Changed section")
	}

	// No bullet entries
	if strings.Contains(result, "* ") {
		t.Error("should not have bullet entries for empty groups")
	}
}

func TestGitHubFormatter_CustomBreakingChangesIcon(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakingChangesIcon = "CUSTOM"
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Breaking": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "abc123",
						ShortHash:   "abc123",
						Subject:     "feat!: breaking change",
						Author:      "Dev",
						AuthorEmail: "dev@users.noreply.github.com",
					},
					Type:        "feat",
					Description: "breaking change",
					Breaking:    true,
					PRNumber:    "999",
				},
				GroupLabel: "Breaking",
			},
		},
	}
	sortedKeys := []string{"Breaking"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Custom icon should appear in header
	if !strings.Contains(result, "### CUSTOM Breaking Changes") {
		t.Error("expected custom breaking changes icon in header")
	}
	// Default icon should NOT appear
	if strings.Contains(result, "⚠️") {
		t.Error("should not have default icon when custom icon is set")
	}
}

func TestGitHubFormatter_DefaultBreakingChangesIcon(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakingChangesIcon = DefaultBreakingChangesIcon
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Breaking": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "abc123",
						ShortHash:   "abc123",
						Subject:     "feat!: breaking change",
						Author:      "Dev",
						AuthorEmail: "dev@users.noreply.github.com",
					},
					Type:        "feat",
					Description: "breaking change",
					Breaking:    true,
					PRNumber:    "999",
				},
				GroupLabel: "Breaking",
			},
		},
	}
	sortedKeys := []string{"Breaking"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Default icon (warning sign) should appear in header
	if !strings.Contains(result, "### "+DefaultBreakingChangesIcon+" Breaking Changes") {
		t.Errorf("expected default breaking changes icon in header, got: %s", result)
	}
}

func TestGitHubFormatter_NoBreakingChangesIcon(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakingChangesIcon = "" // No icon
	formatter := &GitHubFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Breaking": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo: CommitInfo{
						Hash:        "abc123",
						ShortHash:   "abc123",
						Subject:     "feat!: breaking change",
						Author:      "Dev",
						AuthorEmail: "dev@users.noreply.github.com",
					},
					Type:        "feat",
					Description: "breaking change",
					Breaking:    true,
					PRNumber:    "999",
				},
				GroupLabel: "Breaking",
			},
		},
	}
	sortedKeys := []string{"Breaking"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Header should be plain "Breaking Changes" without icon
	if !strings.Contains(result, "### Breaking Changes\n") {
		t.Error("expected plain 'Breaking Changes' header without icon")
	}
	// No icon should appear
	if strings.Contains(result, "⚠️") {
		t.Error("should not have default icon when icon is empty")
	}
}
