package commitparser

import (
	"strings"
	"testing"
)

// FuzzCommitParse tests the commit parser with random commit messages.
// Run with: go test -fuzz=FuzzCommitParse -fuzztime=30s
func FuzzCommitParse(f *testing.F) {
	// Seed corpus with valid conventional commits
	seeds := []string{
		// Breaking changes
		"feat!: add new API",
		"fix!: breaking fix",
		"feat(api)!: breaking feature",
		"fix(core)!: breaking bug fix",
		"chore: update deps\n\nBREAKING CHANGE: removed old API",
		"feat: new feature\n\nBREAKING-CHANGE: old behavior removed",
		// Features
		"feat: add new feature",
		"feat(ui): add button component",
		"feat(api): implement new endpoint",
		// Fixes
		"fix: resolve bug",
		"fix(auth): fix login issue",
		"fix(db): correct query",
		// Other types
		"chore: update dependencies",
		"docs: update README",
		"test: add unit tests",
		"refactor: clean up code",
		"perf: optimize query",
		"style: format code",
		"ci: update workflow",
		"build: update config",
		// Edge cases
		"",
		"random message",
		"FEAT: uppercase type",
		"Fix: mixed case",
		"feat : space before colon",
		"feat(scope: unclosed paren",
		"feat(): empty scope",
		"breaking change in message",
		"This commit contains breaking change somewhere",
		// Multi-line commits
		"feat: title\n\nbody text here",
		"fix: title\n\n- bullet 1\n- bullet 2",
		"chore: title\n\nCo-authored-by: Someone",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, commit string) {
		fuzzSingleCommit(t, commit)
	})
}

// fuzzSingleCommit tests parsing a single commit message.
func fuzzSingleCommit(t *testing.T, commit string) {
	t.Helper()

	parser := NewCommitParser()

	// Should never panic
	result, err := parser.Parse([]string{commit})

	if err == nil {
		// Result should be one of the valid bump types
		validTypes := map[string]bool{"major": true, "minor": true, "patch": true}
		if !validTypes[result] {
			t.Errorf("invalid bump type %q for commit %q", result, commit)
		}

		// Verify the result makes sense
		verifyBumpTypeForCommit(t, commit, result)
	}
}

// verifyBumpTypeForCommit checks that the bump type is appropriate for the commit.
func verifyBumpTypeForCommit(t *testing.T, commit, bumpType string) {
	t.Helper()

	lowerCommit := strings.ToLower(commit)

	// If it detected breaking change, bump type should be major
	hasBreakingIndicator := strings.Contains(lowerCommit, "breaking change") ||
		breakingExclamationRe.MatchString(lowerCommit) ||
		breakingFooterRe.MatchString(commit)

	if hasBreakingIndicator && bumpType != "major" {
		t.Errorf("commit %q has breaking indicator but bump type is %q (expected major)",
			commit, bumpType)
	}
}

// FuzzCommitParseMultiple tests parsing multiple commits together.
func FuzzCommitParseMultiple(f *testing.F) {
	type seedInput struct {
		commit1 string
		commit2 string
		commit3 string
	}

	seeds := []seedInput{
		{"feat: add feature", "fix: bug fix", "chore: update"},
		{"feat!: breaking", "feat: regular", "fix: bug"},
		{"fix: bug1", "fix: bug2", "fix: bug3"},
		{"", "", ""},
		{"feat(api): endpoint", "docs: readme", "test: add tests"},
	}

	for _, seed := range seeds {
		f.Add(seed.commit1, seed.commit2, seed.commit3)
	}

	f.Fuzz(func(t *testing.T, c1, c2, c3 string) {
		fuzzMultipleCommits(t, c1, c2, c3)
	})
}

// fuzzMultipleCommits tests parsing multiple commits.
func fuzzMultipleCommits(t *testing.T, c1, c2, c3 string) {
	t.Helper()

	parser := NewCommitParser()
	commits := []string{c1, c2, c3}

	// Should never panic
	result, err := parser.Parse(commits)

	if err == nil {
		// Verify priority: breaking > feat > fix
		verifyBumpTypePriority(t, commits, result)
	}
}

// verifyBumpTypePriority checks that bump type priority is correct.
func verifyBumpTypePriority(t *testing.T, commits []string, result string) {
	t.Helper()

	hasBreaking := false
	hasFeat := false

	for _, commit := range commits {
		lowerCommit := strings.ToLower(commit)

		if breakingExclamationRe.MatchString(lowerCommit) ||
			breakingFooterRe.MatchString(commit) ||
			strings.Contains(lowerCommit, "breaking change") {
			hasBreaking = true
		}

		if featRe.MatchString(lowerCommit) {
			hasFeat = true
		}
	}

	// Breaking should result in major
	if hasBreaking && result != "major" {
		t.Errorf("has breaking change but result is %q (expected major)", result)
	}

	// If no breaking but has feat, should be minor
	if !hasBreaking && hasFeat && result != "minor" {
		t.Errorf("has feat but no breaking, result is %q (expected minor)", result)
	}
}

// FuzzCommitParserConsistency verifies parsing is deterministic.
func FuzzCommitParserConsistency(f *testing.F) {
	seeds := []string{
		"feat: test",
		"fix: bug",
		"feat!: breaking",
		"random text",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, commit string) {
		parser := NewCommitParser()

		// Parse twice
		result1, err1 := parser.Parse([]string{commit})
		result2, err2 := parser.Parse([]string{commit})

		// Results should be identical
		if (err1 == nil) != (err2 == nil) {
			t.Errorf("inconsistent error for %q: %v vs %v", commit, err1, err2)
		}

		if err1 == nil && result1 != result2 {
			t.Errorf("inconsistent result for %q: %q vs %q", commit, result1, result2)
		}
	})
}
