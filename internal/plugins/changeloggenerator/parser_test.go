package changeloggenerator

import (
	"testing"
)

func TestParseConventionalCommit(t *testing.T) {
	tests := []struct {
		name         string
		subject      string
		wantType     string
		wantScope    string
		wantDesc     string
		wantBreaking bool
		wantPR       string
	}{
		{
			name:      "feat with scope",
			subject:   "feat(cli): add new command",
			wantType:  "feat",
			wantScope: "cli",
			wantDesc:  "add new command",
		},
		{
			name:     "fix without scope",
			subject:  "fix: resolve timeout issue",
			wantType: "fix",
			wantDesc: "resolve timeout issue",
		},
		{
			name:         "breaking change with exclamation",
			subject:      "feat!: redesign API",
			wantType:     "feat",
			wantBreaking: true,
			wantDesc:     "redesign API",
		},
		{
			name:         "breaking change with scope and exclamation",
			subject:      "refactor(core)!: change interface",
			wantType:     "refactor",
			wantScope:    "core",
			wantBreaking: true,
			wantDesc:     "change interface",
		},
		{
			name:      "commit with PR number",
			subject:   "feat(api): add endpoint (#123)",
			wantType:  "feat",
			wantScope: "api",
			wantDesc:  "add endpoint",
			wantPR:    "123",
		},
		{
			name:     "docs commit",
			subject:  "docs: update README",
			wantType: "docs",
			wantDesc: "update README",
		},
		{
			name:     "non-conventional commit",
			subject:  "Update dependencies",
			wantType: "",
			wantDesc: "Update dependencies",
		},
		{
			name:      "chore with hyphenated scope",
			subject:   "chore(ci-cd): update pipeline",
			wantType:  "chore",
			wantScope: "ci-cd",
			wantDesc:  "update pipeline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commit := CommitInfo{
				Hash:      "abc123",
				ShortHash: "abc",
				Subject:   tt.subject,
				Author:    "Test Author",
			}

			got := ParseConventionalCommit(&commit)

			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
			if got.Scope != tt.wantScope {
				t.Errorf("Scope = %q, want %q", got.Scope, tt.wantScope)
			}
			if got.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", got.Description, tt.wantDesc)
			}
			if got.Breaking != tt.wantBreaking {
				t.Errorf("Breaking = %v, want %v", got.Breaking, tt.wantBreaking)
			}
			if got.PRNumber != tt.wantPR {
				t.Errorf("PRNumber = %q, want %q", got.PRNumber, tt.wantPR)
			}
		})
	}
}

func TestParseCommits(t *testing.T) {
	commits := []CommitInfo{
		{Hash: "a", ShortHash: "a", Subject: "feat: add feature"},
		{Hash: "b", ShortHash: "b", Subject: "fix: fix bug"},
		{Hash: "c", ShortHash: "c", Subject: "docs: update docs"},
	}

	parsed := ParseCommits(commits)

	if len(parsed) != 3 {
		t.Fatalf("expected 3 parsed commits, got %d", len(parsed))
	}

	if parsed[0].Type != "feat" {
		t.Errorf("first commit type = %q, want 'feat'", parsed[0].Type)
	}
	if parsed[1].Type != "fix" {
		t.Errorf("second commit type = %q, want 'fix'", parsed[1].Type)
	}
	if parsed[2].Type != "docs" {
		t.Errorf("third commit type = %q, want 'docs'", parsed[2].Type)
	}
}

func TestFilterCommits(t *testing.T) {
	commits := []*ParsedCommit{
		{CommitInfo: CommitInfo{Subject: "feat: add feature"}, Type: "feat"},
		{CommitInfo: CommitInfo{Subject: "Merge branch 'main'"}, Type: ""},
		{CommitInfo: CommitInfo{Subject: "WIP: work in progress"}, Type: ""},
		{CommitInfo: CommitInfo{Subject: "fix: fix bug"}, Type: "fix"},
	}

	patterns := []string{"^Merge", "^WIP"}
	filtered := FilterCommits(commits, patterns)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered commits, got %d", len(filtered))
	}

	if filtered[0].Type != "feat" {
		t.Errorf("first filtered commit type = %q, want 'feat'", filtered[0].Type)
	}
	if filtered[1].Type != "fix" {
		t.Errorf("second filtered commit type = %q, want 'fix'", filtered[1].Type)
	}
}

func TestFilterCommits_EmptyPatterns(t *testing.T) {
	commits := []*ParsedCommit{
		{CommitInfo: CommitInfo{Subject: "feat: add feature"}, Type: "feat"},
		{CommitInfo: CommitInfo{Subject: "fix: fix bug"}, Type: "fix"},
	}

	filtered := FilterCommits(commits, nil)

	if len(filtered) != 2 {
		t.Errorf("expected all commits when no patterns, got %d", len(filtered))
	}
}

func TestGroupCommits(t *testing.T) {
	commits := []*ParsedCommit{
		{CommitInfo: CommitInfo{Hash: "a"}, Type: "feat", Description: "add feature"},
		{CommitInfo: CommitInfo{Hash: "b"}, Type: "feat", Description: "another feature"},
		{CommitInfo: CommitInfo{Hash: "c"}, Type: "fix", Description: "fix bug"},
		{CommitInfo: CommitInfo{Hash: "d"}, Type: "docs", Description: "update docs"},
	}

	groups := []GroupConfig{
		{Pattern: "^feat", Label: "Features"},
		{Pattern: "^fix", Label: "Bug Fixes"},
		{Pattern: "^docs", Label: "Documentation"},
	}

	grouped := GroupCommits(commits, groups)

	if len(grouped["Features"]) != 2 {
		t.Errorf("Features group should have 2 commits, got %d", len(grouped["Features"]))
	}
	if len(grouped["Bug Fixes"]) != 1 {
		t.Errorf("Bug Fixes group should have 1 commit, got %d", len(grouped["Bug Fixes"]))
	}
	if len(grouped["Documentation"]) != 1 {
		t.Errorf("Documentation group should have 1 commit, got %d", len(grouped["Documentation"]))
	}
}

func TestGroupCommits_OrderFromPosition(t *testing.T) {
	commits := []*ParsedCommit{
		{CommitInfo: CommitInfo{Hash: "a"}, Type: "docs"},
		{CommitInfo: CommitInfo{Hash: "b"}, Type: "feat"},
		{CommitInfo: CommitInfo{Hash: "c"}, Type: "fix"},
	}

	// Groups in specific order - order should be derived from position
	groups := []GroupConfig{
		{Pattern: "^feat", Label: "Features"}, // position 0
		{Pattern: "^fix", Label: "Bug Fixes"}, // position 1
		{Pattern: "^docs", Label: "Docs"},     // position 2
	}

	grouped := GroupCommits(commits, groups)
	keys := SortedGroupKeys(grouped)

	// Should be sorted by position order
	expected := []string{"Features", "Bug Fixes", "Docs"}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("key[%d] = %q, want %q", i, key, expected[i])
		}
	}
}

func TestGroupCommits_UnmatchedGoesToOther(t *testing.T) {
	commits := []*ParsedCommit{
		{CommitInfo: CommitInfo{Hash: "a"}, Type: "feat", Description: "add feature"},
		{CommitInfo: CommitInfo{Hash: "b"}, Type: "unknown", Description: "unknown type"},
	}

	groups := []GroupConfig{
		{Pattern: "^feat", Label: "Features"},
	}

	grouped := GroupCommits(commits, groups)

	if len(grouped["Features"]) != 1 {
		t.Errorf("Features group should have 1 commit")
	}
	if len(grouped["Other"]) != 1 {
		t.Errorf("Other group should have 1 commit for unmatched type")
	}
}

func TestSortedGroupKeys(t *testing.T) {
	grouped := map[string][]*GroupedCommit{
		"Chores":    {{GroupOrder: 3}},
		"Features":  {{GroupOrder: 0}},
		"Bug Fixes": {{GroupOrder: 1}},
		"Docs":      {{GroupOrder: 2}},
	}

	keys := SortedGroupKeys(grouped)

	expected := []string{"Features", "Bug Fixes", "Docs", "Chores"}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}

	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("key[%d] = %q, want %q", i, key, expected[i])
		}
	}
}

func TestSortedGroupKeys_Empty(t *testing.T) {
	grouped := map[string][]*GroupedCommit{}

	keys := SortedGroupKeys(grouped)

	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

func TestFilterCommits_InvalidPattern(t *testing.T) {
	commits := []*ParsedCommit{
		{CommitInfo: CommitInfo{Subject: "feat: add feature"}, Type: "feat"},
		{CommitInfo: CommitInfo{Subject: "fix: fix bug"}, Type: "fix"},
	}

	// Invalid regex pattern should be skipped
	patterns := []string{"[invalid"}
	filtered := FilterCommits(commits, patterns)

	// All commits should remain because invalid pattern is skipped
	if len(filtered) != 2 {
		t.Errorf("expected 2 commits, got %d", len(filtered))
	}
}

func TestGroupCommits_InvalidPattern(t *testing.T) {
	commits := []*ParsedCommit{
		{CommitInfo: CommitInfo{Hash: "a"}, Type: "feat", Description: "add feature"},
	}

	// Invalid regex pattern should be skipped
	groups := []GroupConfig{
		{Pattern: "[invalid", Label: "Invalid"},
		{Pattern: "^feat", Label: "Features"},
	}

	grouped := GroupCommits(commits, groups)

	if len(grouped["Features"]) != 1 {
		t.Errorf("Features group should have 1 commit, got %d", len(grouped["Features"]))
	}
}

func TestGroupCommits_WithExplicitOrder(t *testing.T) {
	commits := []*ParsedCommit{
		{CommitInfo: CommitInfo{Hash: "a"}, Type: "fix", Description: "fix bug"},
		{CommitInfo: CommitInfo{Hash: "b"}, Type: "feat", Description: "add feature"},
	}

	// Groups with explicit Order values
	groups := []GroupConfig{
		{Pattern: "^feat", Label: "Features", Order: 10}, // Explicit order 10
		{Pattern: "^fix", Label: "Bug Fixes", Order: 5},  // Explicit order 5
	}

	grouped := GroupCommits(commits, groups)
	keys := SortedGroupKeys(grouped)

	// Bug Fixes should come first (order 5)
	if keys[0] != "Bug Fixes" {
		t.Errorf("expected 'Bug Fixes' first, got %q", keys[0])
	}
	if keys[1] != "Features" {
		t.Errorf("expected 'Features' second, got %q", keys[1])
	}
}

func TestGroupCommits_NonConventionalCommit(t *testing.T) {
	commits := []*ParsedCommit{
		{CommitInfo: CommitInfo{Hash: "a", Subject: "Update README"}, Type: "", Description: "Update README"},
	}

	groups := []GroupConfig{
		{Pattern: "^feat", Label: "Features"},
	}

	grouped := GroupCommits(commits, groups)

	// Non-conventional commit without type should not be in any group
	if len(grouped["Features"]) != 0 {
		t.Errorf("Features should have 0 commits, got %d", len(grouped["Features"]))
	}
	if len(grouped["Other"]) != 0 {
		// Non-conventional commits (empty type) don't go to Other
		t.Errorf("Other should have 0 commits for empty type, got %d", len(grouped["Other"]))
	}
}

func TestParsedCommit_Fields(t *testing.T) {
	pc := ParsedCommit{
		CommitInfo: CommitInfo{
			Hash:        "abc123",
			ShortHash:   "abc",
			Subject:     "feat(cli): add feature (#42)",
			Author:      "Test",
			AuthorEmail: "test@example.com",
		},
		Type:        "feat",
		Scope:       "cli",
		Description: "add feature",
		Breaking:    false,
		PRNumber:    "42",
	}

	if pc.Type != "feat" {
		t.Errorf("Type = %q, want 'feat'", pc.Type)
	}
	if pc.Scope != "cli" {
		t.Errorf("Scope = %q, want 'cli'", pc.Scope)
	}
	if pc.PRNumber != "42" {
		t.Errorf("PRNumber = %q, want '42'", pc.PRNumber)
	}
}

func TestGroupedCommit_Fields(t *testing.T) {
	gc := GroupedCommit{
		ParsedCommit: &ParsedCommit{
			CommitInfo: CommitInfo{Hash: "abc"},
			Type:       "feat",
		},
		GroupLabel: "Features",
		GroupIcon:  "rocket",
		GroupOrder: 0,
	}

	if gc.GroupLabel != "Features" {
		t.Errorf("GroupLabel = %q, want 'Features'", gc.GroupLabel)
	}
	if gc.GroupIcon != "rocket" {
		t.Errorf("GroupIcon = %q, want 'rocket'", gc.GroupIcon)
	}
	if gc.GroupOrder != 0 {
		t.Errorf("GroupOrder = %d, want 0", gc.GroupOrder)
	}
}

func TestGroupCommitsWithOptions_IncludeNonConventional(t *testing.T) {
	commits := []*ParsedCommit{
		{CommitInfo: CommitInfo{Hash: "a", Subject: "feat: add feature"}, Type: "feat", Description: "add feature"},
		{CommitInfo: CommitInfo{Hash: "b", Subject: "Update README"}, Type: "", Description: "Update README"},
		{CommitInfo: CommitInfo{Hash: "c", Subject: "fix: fix bug"}, Type: "fix", Description: "fix bug"},
	}

	groups := []GroupConfig{
		{Pattern: "^feat", Label: "Features"},
		{Pattern: "^fix", Label: "Bug Fixes"},
	}

	// With includeNonConventional = true
	result := GroupCommitsWithOptions(commits, groups, true)

	if len(result.Grouped["Features"]) != 1 {
		t.Errorf("Features should have 1 commit, got %d", len(result.Grouped["Features"]))
	}
	if len(result.Grouped["Bug Fixes"]) != 1 {
		t.Errorf("Bug Fixes should have 1 commit, got %d", len(result.Grouped["Bug Fixes"]))
	}
	if len(result.Grouped["Other Changes"]) != 1 {
		t.Errorf("Other Changes should have 1 commit, got %d", len(result.Grouped["Other Changes"]))
	}
	if len(result.SkippedNonConventional) != 0 {
		t.Errorf("SkippedNonConventional should be empty when includeNonConventional=true")
	}
}

func TestGroupCommitsWithOptions_ExcludeNonConventional(t *testing.T) {
	commits := []*ParsedCommit{
		{CommitInfo: CommitInfo{Hash: "a", ShortHash: "a", Subject: "feat: add feature"}, Type: "feat", Description: "add feature"},
		{CommitInfo: CommitInfo{Hash: "b", ShortHash: "b", Subject: "Update README"}, Type: "", Description: "Update README"},
		{CommitInfo: CommitInfo{Hash: "c", ShortHash: "c", Subject: "Bump version"}, Type: "", Description: "Bump version"},
		{CommitInfo: CommitInfo{Hash: "d", ShortHash: "d", Subject: "fix: fix bug"}, Type: "fix", Description: "fix bug"},
	}

	groups := []GroupConfig{
		{Pattern: "^feat", Label: "Features"},
		{Pattern: "^fix", Label: "Bug Fixes"},
	}

	// With includeNonConventional = false
	result := GroupCommitsWithOptions(commits, groups, false)

	if len(result.Grouped["Features"]) != 1 {
		t.Errorf("Features should have 1 commit, got %d", len(result.Grouped["Features"]))
	}
	if len(result.Grouped["Bug Fixes"]) != 1 {
		t.Errorf("Bug Fixes should have 1 commit, got %d", len(result.Grouped["Bug Fixes"]))
	}
	if _, ok := result.Grouped["Other Changes"]; ok {
		t.Errorf("Other Changes should not exist when includeNonConventional=false")
	}
	if len(result.SkippedNonConventional) != 2 {
		t.Errorf("SkippedNonConventional should have 2 commits, got %d", len(result.SkippedNonConventional))
	}

	// Verify skipped commits content
	skippedSubjects := make([]string, len(result.SkippedNonConventional))
	for i, c := range result.SkippedNonConventional {
		skippedSubjects[i] = c.Subject
	}
	if skippedSubjects[0] != "Update README" {
		t.Errorf("First skipped commit should be 'Update README', got %q", skippedSubjects[0])
	}
	if skippedSubjects[1] != "Bump version" {
		t.Errorf("Second skipped commit should be 'Bump version', got %q", skippedSubjects[1])
	}
}
