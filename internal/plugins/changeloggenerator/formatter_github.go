package changeloggenerator

import (
	"fmt"
	"strings"
	"time"
)

// GitHubFormatter implements the GitHub release changelog format.
// Key features:
// - Single "What's Changed" section (no grouping by type)
// - Inline contributor attribution per commit (@username)
// - PR references inline (in #123)
// - Simple, flat list style using * instead of -
type GitHubFormatter struct {
	config *Config
}

// FormatChangelog generates the changelog in GitHub release format.
func (f *GitHubFormatter) FormatChangelog(
	version string,
	previousVersion string,
	grouped map[string][]*GroupedCommit,
	sortedKeys []string,
	remote *RemoteInfo,
) string {
	var sb strings.Builder

	// Version header with "v" prefix (like grouped format)
	date := time.Now().Format("2006-01-02")
	sb.WriteString(fmt.Sprintf("## %s - %s\n\n", version, date))

	// Separate breaking changes from regular changes
	var breakingChanges []*GroupedCommit
	var regularChanges []*GroupedCommit

	for _, label := range sortedKeys {
		commits := grouped[label]
		for _, c := range commits {
			if c.Breaking {
				breakingChanges = append(breakingChanges, c)
			} else {
				regularChanges = append(regularChanges, c)
			}
		}
	}

	// Write breaking changes section first if there are any
	if len(breakingChanges) > 0 {
		sb.WriteString("### ⚠️ Breaking Changes\n\n")
		for _, c := range breakingChanges {
			entry := formatGitHubCommitEntry(c, remote)
			sb.WriteString(entry)
		}
		sb.WriteString("\n")
	}

	// Write regular changes section if there are any
	if len(regularChanges) > 0 {
		sb.WriteString("### What's Changed\n\n")
		for _, c := range regularChanges {
			entry := formatGitHubCommitEntry(c, remote)
			sb.WriteString(entry)
		}
		sb.WriteString("\n")
	}

	// If there are no changes at all, still show the What's Changed header
	if len(breakingChanges) == 0 && len(regularChanges) == 0 {
		sb.WriteString("### What's Changed\n\n")
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatGitHubCommitEntry formats a single commit entry in GitHub style.
// Format: * **{scope}:** {description} by @{username} in #{pr_number}
func formatGitHubCommitEntry(c *GroupedCommit, remote *RemoteInfo) string {
	var sb strings.Builder

	// Use * bullet (GitHub style)
	sb.WriteString("* ")

	// Add scope if present (bold)
	if c.Scope != "" {
		sb.WriteString(fmt.Sprintf("**%s:** ", c.Scope))
	}

	// Add description
	sb.WriteString(c.Description)

	// Add author attribution using extractUsername
	username, _ := extractUsername(c.AuthorEmail, c.Author)
	if username != "" {
		sb.WriteString(fmt.Sprintf(" by @%s", username))
	}

	// Add PR reference if present
	if c.PRNumber != "" {
		sb.WriteString(fmt.Sprintf(" in #%s", c.PRNumber))
	}

	sb.WriteString("\n")
	return sb.String()
}
