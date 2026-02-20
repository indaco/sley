package changeloggenerator

import (
	"fmt"
	"strings"
)

// MinimalFormatter implements a condensed changelog format using short type abbreviations.
// Key features:
// - Version header without date (just `## v1.2.0`)
// - Abbreviated type prefixes in brackets: [Feat], [Fix], etc.
// - No sections/grouping - flat list of all commits
// - No author attribution, no PR links, no commit links
// - Breaking changes marked with [Breaking] prefix instead of type
// - Simple `-` bullet points
type MinimalFormatter struct {
	config *Config
}

// typeAbbreviations maps conventional commit types to their abbreviated display form.
var typeAbbreviations = map[string]string{
	"feat":     "Feat",
	"fix":      "Fix",
	"docs":     "Docs",
	"perf":     "Perf",
	"refactor": "Refactor",
	"style":    "Style",
	"test":     "Test",
	"chore":    "Chore",
	"ci":       "CI",
	"build":    "Build",
	"revert":   "Revert",
}

// FormatChangelog generates the changelog in minimal format.
func (f *MinimalFormatter) FormatChangelog(
	version string,
	previousVersion string,
	grouped map[string][]*GroupedCommit,
	sortedKeys []string,
	remote *RemoteInfo,
) string {
	var sb strings.Builder

	// Version header without date
	fmt.Fprintf(&sb, "## %s\n\n", version)

	// Collect all commits from all groups into a flat list
	var allCommits []*GroupedCommit
	for _, label := range sortedKeys {
		commits := grouped[label]
		allCommits = append(allCommits, commits...)
	}

	// Write each commit as a simple bullet point
	for _, c := range allCommits {
		entry := formatMinimalCommitEntry(c)
		sb.WriteString(entry)
	}

	// Add trailing newline if there were commits
	if len(allCommits) > 0 {
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatMinimalCommitEntry formats a single commit entry in minimal style.
// Format: - [Type] description
func formatMinimalCommitEntry(c *GroupedCommit) string {
	var sb strings.Builder

	sb.WriteString("- ")

	// Determine the type abbreviation
	typeAbbr := getTypeAbbreviation(c)
	fmt.Fprintf(&sb, "[%s] ", typeAbbr)

	// Add description
	sb.WriteString(c.Description)

	sb.WriteString("\n")
	return sb.String()
}

// getTypeAbbreviation returns the abbreviated type prefix for a commit.
// Breaking changes always return "Breaking" regardless of the original type.
// Unknown or empty types return "Other".
func getTypeAbbreviation(c *GroupedCommit) string {
	// Breaking changes take precedence
	if c.Breaking {
		return "Breaking"
	}

	// Look up the type abbreviation
	if abbr, ok := typeAbbreviations[c.Type]; ok {
		return abbr
	}

	// Unknown or empty type
	return "Other"
}
