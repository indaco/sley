package changeloggenerator

import (
	"fmt"
	"strings"
	"time"
)

// GroupedFormatter implements the default "grouped" changelog format.
// This formatter groups commits by their configured labels and supports
// custom icons, compare links, and flexible grouping rules.
type GroupedFormatter struct {
	config *Config
}

// FormatChangelog generates the changelog in grouped format.
func (f *GroupedFormatter) FormatChangelog(
	version string,
	previousVersion string,
	grouped map[string][]*GroupedCommit,
	sortedKeys []string,
	remote *RemoteInfo,
) string {
	var sb strings.Builder

	// Version header with "v" prefix
	date := time.Now().Format("2006-01-02")
	sb.WriteString(fmt.Sprintf("## %s - %s\n\n", version, date))

	// Separate breaking changes from regular changes
	var breakingChanges []*GroupedCommit
	regularGrouped := make(map[string][]*GroupedCommit)

	for _, label := range sortedKeys {
		commits := grouped[label]
		for _, c := range commits {
			if c.Breaking {
				breakingChanges = append(breakingChanges, c)
			} else {
				regularGrouped[label] = append(regularGrouped[label], c)
			}
		}
	}

	// Write breaking changes section first if there are any
	if len(breakingChanges) > 0 {
		sb.WriteString(f.formatBreakingChangesHeader())
		for _, c := range breakingChanges {
			entry := formatCommitEntry(c, remote)
			sb.WriteString(entry)
		}
		sb.WriteString("\n")
	}

	// Grouped commits (regular, non-breaking)
	for _, label := range sortedKeys {
		commits := regularGrouped[label]
		if len(commits) == 0 {
			continue
		}

		// Section header with optional icon
		icon := commits[0].GroupIcon
		if icon != "" {
			sb.WriteString(fmt.Sprintf("### %s %s\n\n", icon, label))
		} else {
			sb.WriteString(fmt.Sprintf("### %s\n\n", label))
		}

		// Commit entries
		for _, c := range commits {
			entry := formatCommitEntry(c, remote)
			sb.WriteString(entry)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatBreakingChangesHeader returns the breaking changes section header.
// If a custom icon is configured, it is used; otherwise just "Breaking Changes" is shown.
func (f *GroupedFormatter) formatBreakingChangesHeader() string {
	if f.config.BreakingChangesIcon != "" {
		return fmt.Sprintf("### %s Breaking Changes\n\n", f.config.BreakingChangesIcon)
	}
	return "### Breaking Changes\n\n"
}

// buildCompareURL generates a compare URL for the provider.
func buildCompareURL(remote *RemoteInfo, prev, curr string) string {
	switch remote.Provider {
	case "github", "gitea", "codeberg":
		return fmt.Sprintf("https://%s/%s/%s/compare/%s...%s",
			remote.Host, remote.Owner, remote.Repo, prev, curr)
	case "gitlab":
		return fmt.Sprintf("https://%s/%s/%s/-/compare/%s...%s",
			remote.Host, remote.Owner, remote.Repo, prev, curr)
	case "bitbucket":
		return fmt.Sprintf("https://%s/%s/%s/branches/compare/%s%%0D%s",
			remote.Host, remote.Owner, remote.Repo, curr, prev)
	case "sourcehut":
		return fmt.Sprintf("https://git.%s/%s/%s/log/%s..%s",
			remote.Host, remote.Owner, remote.Repo, prev, curr)
	default:
		// For custom providers, use GitHub-style URL as best effort
		return fmt.Sprintf("https://%s/%s/%s/compare/%s...%s",
			remote.Host, remote.Owner, remote.Repo, prev, curr)
	}
}

// formatCommitEntry formats a single commit entry.
func formatCommitEntry(c *GroupedCommit, remote *RemoteInfo) string {
	var sb strings.Builder

	sb.WriteString("- ")

	// Add scope if present
	if c.Scope != "" {
		sb.WriteString(fmt.Sprintf("**%s:** ", c.Scope))
	}

	// Add description
	sb.WriteString(c.Description)

	// Add commit link (always) and PR link (if present)
	if remote != nil {
		commitURL := buildCommitURL(remote, c.ShortHash)
		sb.WriteString(fmt.Sprintf(" ([%s](%s))", c.ShortHash, commitURL))

		// Add PR link if present
		if c.PRNumber != "" {
			prURL := buildPRURL(remote, c.PRNumber)
			sb.WriteString(fmt.Sprintf(" ([#%s](%s))", c.PRNumber, prURL))
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

// buildCommitURL generates a commit URL for the provider.
func buildCommitURL(remote *RemoteInfo, hash string) string {
	switch remote.Provider {
	case "github", "gitea", "codeberg":
		return fmt.Sprintf("https://%s/%s/%s/commit/%s",
			remote.Host, remote.Owner, remote.Repo, hash)
	case "gitlab":
		return fmt.Sprintf("https://%s/%s/%s/-/commit/%s",
			remote.Host, remote.Owner, remote.Repo, hash)
	case "bitbucket":
		return fmt.Sprintf("https://%s/%s/%s/commits/%s",
			remote.Host, remote.Owner, remote.Repo, hash)
	case "sourcehut":
		return fmt.Sprintf("https://git.%s/%s/%s/commit/%s",
			remote.Host, remote.Owner, remote.Repo, hash)
	default:
		return fmt.Sprintf("https://%s/%s/%s/commit/%s",
			remote.Host, remote.Owner, remote.Repo, hash)
	}
}

// buildPRURL generates a PR/MR URL for the provider.
func buildPRURL(remote *RemoteInfo, prNumber string) string {
	switch remote.Provider {
	case "github", "gitea", "codeberg":
		return fmt.Sprintf("https://%s/%s/%s/pull/%s",
			remote.Host, remote.Owner, remote.Repo, prNumber)
	case "gitlab":
		return fmt.Sprintf("https://%s/%s/%s/-/merge_requests/%s",
			remote.Host, remote.Owner, remote.Repo, prNumber)
	case "bitbucket":
		return fmt.Sprintf("https://%s/%s/%s/pull-requests/%s",
			remote.Host, remote.Owner, remote.Repo, prNumber)
	default:
		return fmt.Sprintf("https://%s/%s/%s/pull/%s",
			remote.Host, remote.Owner, remote.Repo, prNumber)
	}
}
