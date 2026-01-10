package changeloggenerator

import (
	"strings"
	"testing"
)

func TestMinimalFormatter_FormatChangelog(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &MinimalFormatter{config: cfg}

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
					CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add new caching layer"},
					Type:        "feat",
					Scope:       "cache",
					Description: "Add new caching layer",
				},
				GroupLabel: "Enhancements",
				GroupOrder: 0,
			},
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "abc456", ShortHash: "abc456", Subject: "feat: implement user settings"},
					Type:        "feat",
					Description: "Implement user settings",
				},
				GroupLabel: "Enhancements",
				GroupOrder: 0,
			},
		},
		"Fixes": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "def456", ShortHash: "def456", Subject: "fix: memory leak in parser"},
					Type:        "fix",
					Description: "Memory leak in parser",
				},
				GroupLabel: "Fixes",
				GroupOrder: 1,
			},
		},
	}
	sortedKeys := []string{"Enhancements", "Fixes"}

	result := formatter.FormatChangelog("v1.2.0", "v1.1.0", grouped, sortedKeys, remote)

	// Check version header without date
	if !strings.Contains(result, "## v1.2.0\n") {
		t.Error("expected version header without date")
	}

	// Should NOT contain date
	if strings.Contains(result, " - 20") {
		t.Error("minimal format should not include date in header")
	}

	// Check type abbreviations
	if !strings.Contains(result, "[Feat] Add new caching layer") {
		t.Error("expected [Feat] prefix for feat commits")
	}
	if !strings.Contains(result, "[Fix] Memory leak in parser") {
		t.Error("expected [Fix] prefix for fix commits")
	}

	// Check bullet points
	if !strings.Contains(result, "- [Feat]") {
		t.Error("expected dash bullet points")
	}

	// Should NOT contain links (no URLs)
	if strings.Contains(result, "https://") {
		t.Error("minimal format should not contain URLs")
	}

	// Should NOT contain section headers
	if strings.Contains(result, "### ") {
		t.Error("minimal format should not contain section headers")
	}

	// Should NOT contain author attribution
	if strings.Contains(result, "by @") {
		t.Error("minimal format should not contain author attribution")
	}

	// Should NOT contain PR references
	if strings.Contains(result, "in #") {
		t.Error("minimal format should not contain PR references")
	}
}

func TestMinimalFormatter_BreakingChanges(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &MinimalFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123", Subject: "feat!: remove deprecated API"},
					Type:        "feat",
					Description: "Remove deprecated API",
					Breaking:    true,
				},
				GroupLabel: "Enhancements",
			},
			{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{Hash: "def456", ShortHash: "def456", Subject: "fix!: change error handling"},
					Type:        "fix",
					Description: "Change error handling",
					Breaking:    true,
				},
				GroupLabel: "Enhancements",
			},
		},
	}
	sortedKeys := []string{"Enhancements"}

	result := formatter.FormatChangelog("v2.0.0", "v1.0.0", grouped, sortedKeys, nil)

	// Breaking changes should use [Breaking] prefix instead of type
	if !strings.Contains(result, "[Breaking] Remove deprecated API") {
		t.Error("expected [Breaking] prefix for breaking feat commit")
	}
	if !strings.Contains(result, "[Breaking] Change error handling") {
		t.Error("expected [Breaking] prefix for breaking fix commit")
	}

	// Should NOT have the original type prefixes for breaking changes
	if strings.Contains(result, "[Feat] Remove deprecated") {
		t.Error("breaking changes should not use [Feat] prefix")
	}
	if strings.Contains(result, "[Fix] Change error") {
		t.Error("breaking changes should not use [Fix] prefix")
	}
}

func TestMinimalFormatter_AllTypes(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &MinimalFormatter{config: cfg}

	tests := []struct {
		commitType     string
		expectedPrefix string
	}{
		{"feat", "[Feat]"},
		{"fix", "[Fix]"},
		{"docs", "[Docs]"},
		{"perf", "[Perf]"},
		{"refactor", "[Refactor]"},
		{"style", "[Style]"},
		{"test", "[Test]"},
		{"chore", "[Chore]"},
		{"ci", "[CI]"},
		{"build", "[Build]"},
		{"revert", "[Revert]"},
	}

	for _, tt := range tests {
		t.Run(tt.commitType, func(t *testing.T) {
			grouped := map[string][]*GroupedCommit{
				"Test": {
					{
						ParsedCommit: &ParsedCommit{
							CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123"},
							Type:        tt.commitType,
							Description: "test description",
						},
						GroupLabel: "Test",
					},
				},
			}

			result := formatter.FormatChangelog("v1.0.0", "", grouped, []string{"Test"}, nil)

			if !strings.Contains(result, tt.expectedPrefix+" test description") {
				t.Errorf("expected %s prefix for %s commits, got: %s", tt.expectedPrefix, tt.commitType, result)
			}
		})
	}
}

func TestMinimalFormatter_UnknownAndEmptyTypes(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &MinimalFormatter{config: cfg}

	tests := []struct {
		name           string
		commitType     string
		expectedPrefix string
	}{
		{"unknown type", "custom", "[Other]"},
		{"empty type", "", "[Other]"},
		{"random type", "xyz", "[Other]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grouped := map[string][]*GroupedCommit{
				"Other": {
					{
						ParsedCommit: &ParsedCommit{
							CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123"},
							Type:        tt.commitType,
							Description: "some change",
						},
						GroupLabel: "Other",
					},
				},
			}

			result := formatter.FormatChangelog("v1.0.0", "", grouped, []string{"Other"}, nil)

			if !strings.Contains(result, tt.expectedPrefix+" some change") {
				t.Errorf("expected %s prefix for %s, got: %s", tt.expectedPrefix, tt.name, result)
			}
		})
	}
}

func TestMinimalFormatter_EmptyChangelog(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &MinimalFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{}
	sortedKeys := []string{}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Should still have version header
	if !strings.Contains(result, "## v1.0.0") {
		t.Error("expected version header even with no commits")
	}

	// Should be minimal output (just header and newlines)
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) > 1 {
		t.Errorf("expected only version header for empty changelog, got %d lines", len(lines))
	}
}

func TestMinimalFormatter_FlatList(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &MinimalFormatter{config: cfg}

	// Multiple groups should be flattened into a single list
	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					Type:        "feat",
					Description: "feature one",
					CommitInfo:  CommitInfo{ShortHash: "aaa"},
				},
			},
		},
		"Fixes": {
			{
				ParsedCommit: &ParsedCommit{
					Type:        "fix",
					Description: "fix one",
					CommitInfo:  CommitInfo{ShortHash: "bbb"},
				},
			},
		},
		"Documentation": {
			{
				ParsedCommit: &ParsedCommit{
					Type:        "docs",
					Description: "doc one",
					CommitInfo:  CommitInfo{ShortHash: "ccc"},
				},
			},
		},
	}
	sortedKeys := []string{"Enhancements", "Fixes", "Documentation"}

	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Count bullet points - should be exactly 3
	bulletCount := strings.Count(result, "- [")
	if bulletCount != 3 {
		t.Errorf("expected 3 bullet points, got %d", bulletCount)
	}

	// Ensure all commits are present
	if !strings.Contains(result, "[Feat] feature one") {
		t.Error("missing feat commit")
	}
	if !strings.Contains(result, "[Fix] fix one") {
		t.Error("missing fix commit")
	}
	if !strings.Contains(result, "[Docs] doc one") {
		t.Error("missing docs commit")
	}
}

func TestMinimalFormatter_NoRemoteInfo(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &MinimalFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					Type:        "feat",
					Description: "add feature",
					CommitInfo:  CommitInfo{Hash: "abc123", ShortHash: "abc123"},
				},
			},
		},
	}
	sortedKeys := []string{"Enhancements"}

	// Remote is nil
	result := formatter.FormatChangelog("v1.0.0", "", grouped, sortedKeys, nil)

	// Should still work without remote info
	if !strings.Contains(result, "## v1.0.0") {
		t.Error("expected version header")
	}
	if !strings.Contains(result, "[Feat] add feature") {
		t.Error("expected commit entry")
	}
}

func TestMinimalFormatter_OutputFormat(t *testing.T) {
	cfg := DefaultConfig()
	formatter := &MinimalFormatter{config: cfg}

	grouped := map[string][]*GroupedCommit{
		"Enhancements": {
			{
				ParsedCommit: &ParsedCommit{
					Type:        "feat",
					Description: "Add new caching layer",
					CommitInfo:  CommitInfo{ShortHash: "aaa"},
				},
			},
			{
				ParsedCommit: &ParsedCommit{
					Type:        "feat",
					Description: "Implement user settings",
					CommitInfo:  CommitInfo{ShortHash: "bbb"},
				},
			},
		},
		"Fixes": {
			{
				ParsedCommit: &ParsedCommit{
					Type:        "fix",
					Description: "Memory leak in parser",
					CommitInfo:  CommitInfo{ShortHash: "ccc"},
				},
			},
		},
	}
	sortedKeys := []string{"Enhancements", "Fixes"}

	result := formatter.FormatChangelog("v1.2.0", "", grouped, sortedKeys, nil)

	// Check the exact expected format structure
	expected := "## v1.2.0\n\n- [Feat] Add new caching layer\n- [Feat] Implement user settings\n- [Fix] Memory leak in parser\n\n"

	if result != expected {
		t.Errorf("output format mismatch.\nExpected:\n%q\nGot:\n%q", expected, result)
	}
}

func TestGetTypeAbbreviation(t *testing.T) {
	tests := []struct {
		name     string
		commit   *GroupedCommit
		expected string
	}{
		{
			name: "feat type",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{Type: "feat"},
			},
			expected: "Feat",
		},
		{
			name: "fix type",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{Type: "fix"},
			},
			expected: "Fix",
		},
		{
			name: "breaking feat",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{Type: "feat", Breaking: true},
			},
			expected: "Breaking",
		},
		{
			name: "breaking fix",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{Type: "fix", Breaking: true},
			},
			expected: "Breaking",
		},
		{
			name: "unknown type",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{Type: "unknown"},
			},
			expected: "Other",
		},
		{
			name: "empty type",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{Type: ""},
			},
			expected: "Other",
		},
		{
			name: "ci type",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{Type: "ci"},
			},
			expected: "CI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTypeAbbreviation(tt.commit)
			if result != tt.expected {
				t.Errorf("getTypeAbbreviation() = %q, want %q", result, tt.expected)
			}
		})
	}
}
