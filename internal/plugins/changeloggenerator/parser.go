package changeloggenerator

import (
	"regexp"
	"strings"
)

// ParsedCommit represents a fully parsed conventional commit.
type ParsedCommit struct {
	CommitInfo
	Type        string // feat, fix, docs, etc.
	Scope       string // Optional scope in parentheses
	Description string // The commit description after the colon
	Breaking    bool   // Has breaking change indicator (! or BREAKING CHANGE footer)
	PRNumber    string // Extracted PR/MR number if present
}

// Regex patterns for conventional commit parsing.
var (
	// Matches: type(scope)!: description or type!: description or type: description
	conventionalCommitRe = regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?(!)?:\s*(.+)$`)

	// Matches: (#123) or (closes #123) etc at end of message
	prNumberRe = regexp.MustCompile(`\(?#(\d+)\)?`)
)

// ParseConventionalCommit parses a commit message into its components.
// Returns nil if the commit doesn't follow conventional commit format.
func ParseConventionalCommit(commit *CommitInfo) *ParsedCommit {
	matches := conventionalCommitRe.FindStringSubmatch(commit.Subject)
	if matches == nil {
		// Not a conventional commit, return with just the subject as description
		return &ParsedCommit{
			CommitInfo:  *commit,
			Type:        "",
			Description: commit.Subject,
		}
	}

	parsed := &ParsedCommit{
		CommitInfo:  *commit,
		Type:        strings.ToLower(matches[1]),
		Scope:       matches[2],
		Breaking:    matches[3] == "!",
		Description: matches[4],
	}

	// Extract PR number from description and remove it from the description text
	if prMatches := prNumberRe.FindStringSubmatch(parsed.Description); len(prMatches) == 2 {
		parsed.PRNumber = prMatches[1]
		// Remove the PR reference from description to avoid duplication in output
		parsed.Description = strings.TrimSpace(prNumberRe.ReplaceAllString(parsed.Description, ""))
	}

	return parsed
}

// ParseCommits parses a slice of CommitInfo into ParsedCommits.
func ParseCommits(commits []CommitInfo) []*ParsedCommit {
	parsed := make([]*ParsedCommit, 0, len(commits))
	for i := range commits {
		parsed = append(parsed, ParseConventionalCommit(&commits[i]))
	}
	return parsed
}

// FilterCommits filters out commits matching exclude patterns.
func FilterCommits(commits []*ParsedCommit, excludePatterns []string) []*ParsedCommit {
	if len(excludePatterns) == 0 {
		return commits
	}

	// Compile patterns
	patterns := make([]*regexp.Regexp, 0, len(excludePatterns))
	for _, p := range excludePatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			continue // Skip invalid patterns
		}
		patterns = append(patterns, re)
	}

	filtered := make([]*ParsedCommit, 0, len(commits))
	for _, c := range commits {
		excluded := false
		for _, re := range patterns {
			if re.MatchString(c.Subject) {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

// GroupedCommit represents a commit with its group assignment.
type GroupedCommit struct {
	*ParsedCommit
	GroupLabel string
	GroupIcon  string
	GroupOrder int
}

// GroupCommitsResult contains grouped commits and any skipped non-conventional commits.
type GroupCommitsResult struct {
	Grouped                map[string][]*GroupedCommit
	SkippedNonConventional []*ParsedCommit
}

// GroupCommits groups parsed commits by their type using the configured groups.
// The order is derived from the position in the groups slice (index) unless
// explicitly overridden by the Order field (if > 0).
// If includeNonConventional is true, commits without a type are included in "Other Changes".
// If false, they are returned in SkippedNonConventional for warning purposes.
func GroupCommits(commits []*ParsedCommit, groups []GroupConfig) map[string][]*GroupedCommit {
	result := GroupCommitsWithOptions(commits, groups, false)
	return result.Grouped
}

// compiledGroup holds a group config with its compiled regex and order.
type compiledGroup struct {
	GroupConfig
	re    *regexp.Regexp
	order int
}

// compileGroupPatterns compiles group patterns into regexes with order.
func compileGroupPatterns(groups []GroupConfig) []compiledGroup {
	compiled := make([]compiledGroup, 0, len(groups))
	for i, g := range groups {
		re, err := regexp.Compile(g.Pattern)
		if err != nil {
			continue
		}
		order := i
		if g.Order > 0 {
			order = g.Order
		}
		compiled = append(compiled, compiledGroup{GroupConfig: g, re: re, order: order})
	}
	return compiled
}

// matchCommitToGroup attempts to match a commit to a group, returning the match or nil.
func matchCommitToGroup(commit *ParsedCommit, groups []compiledGroup) *GroupedCommit {
	matchTarget := commit.Type
	if matchTarget == "" {
		matchTarget = commit.Subject
	}
	for _, group := range groups {
		if group.re.MatchString(matchTarget) {
			return &GroupedCommit{
				ParsedCommit: commit,
				GroupLabel:   group.Label,
				GroupIcon:    group.Icon,
				GroupOrder:   group.order,
			}
		}
	}
	return nil
}

// handleUnmatchedCommit returns a grouped commit for unmatched commits or nil if skipped.
func handleUnmatchedCommit(commit *ParsedCommit, includeNonConventional bool) (*GroupedCommit, bool) {
	if commit.Type != "" {
		return &GroupedCommit{ParsedCommit: commit, GroupLabel: "Other", GroupOrder: 999}, false
	}
	if includeNonConventional {
		return &GroupedCommit{ParsedCommit: commit, GroupLabel: "Other Changes", GroupOrder: 1000}, false
	}
	return nil, true // skipped
}

// GroupCommitsWithOptions groups commits by configured patterns with options.
func GroupCommitsWithOptions(commits []*ParsedCommit, groups []GroupConfig, includeNonConventional bool) GroupCommitsResult {
	result := GroupCommitsResult{
		Grouped:                make(map[string][]*GroupedCommit),
		SkippedNonConventional: make([]*ParsedCommit, 0),
	}

	compiledGroups := compileGroupPatterns(groups)

	for _, commit := range commits {
		if gc := matchCommitToGroup(commit, compiledGroups); gc != nil {
			result.Grouped[gc.GroupLabel] = append(result.Grouped[gc.GroupLabel], gc)
			continue
		}
		if gc, skipped := handleUnmatchedCommit(commit, includeNonConventional); skipped {
			result.SkippedNonConventional = append(result.SkippedNonConventional, commit)
		} else {
			result.Grouped[gc.GroupLabel] = append(result.Grouped[gc.GroupLabel], gc)
		}
	}

	return result
}

// SortedGroupKeys returns group labels sorted by their order.
func SortedGroupKeys(grouped map[string][]*GroupedCommit) []string {
	type groupInfo struct {
		label string
		order int
	}

	infos := make([]groupInfo, 0, len(grouped))
	for label, commits := range grouped {
		if len(commits) > 0 {
			infos = append(infos, groupInfo{label: label, order: commits[0].GroupOrder})
		}
	}

	// Simple insertion sort (groups are small)
	for i := 1; i < len(infos); i++ {
		for j := i; j > 0 && infos[j].order < infos[j-1].order; j-- {
			infos[j], infos[j-1] = infos[j-1], infos[j]
		}
	}

	keys := make([]string, len(infos))
	for i, info := range infos {
		keys[i] = info.label
	}
	return keys
}
