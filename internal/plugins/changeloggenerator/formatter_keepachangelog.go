package changeloggenerator

import (
	"fmt"
	"strings"
	"time"
)

// KeepAChangelogFormatter implements the "Keep a Changelog" format.
// This formatter follows the Keep a Changelog specification:
// https://keepachangelog.com/
//
// Key differences from grouped format:
// - Version header: ## [version] - date (no "v" prefix, brackets around version)
// - Standard sections: Added, Changed, Deprecated, Removed, Fixed, Security
// - No compare links (not part of the spec)
// - No custom icons (strict section names)
type KeepAChangelogFormatter struct {
	config *Config
}

// FormatChangelog generates the changelog in Keep a Changelog format.
func (f *KeepAChangelogFormatter) FormatChangelog(
	version string,
	previousVersion string,
	grouped map[string][]*GroupedCommit,
	sortedKeys []string,
	remote *RemoteInfo,
) string {
	var sb strings.Builder

	// Version header without "v" prefix, with brackets
	date := time.Now().Format("2006-01-02")
	versionNumber := strings.TrimPrefix(version, "v")
	sb.WriteString(fmt.Sprintf("## [%s] - %s\n\n", versionNumber, date))

	// Regroup commits according to Keep a Changelog sections
	sections := f.regroupCommits(grouped)

	// Write sections in Keep a Changelog order
	sectionOrder := []string{"Breaking Changes", "Added", "Changed", "Deprecated", "Removed", "Fixed", "Security"}

	for _, sectionName := range sectionOrder {
		commits, exists := sections[sectionName]
		if !exists || len(commits) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s\n\n", sectionName))

		for _, c := range commits {
			entry := formatCommitEntry(c, remote)
			sb.WriteString(entry)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// regroupCommits maps conventional commit types to Keep a Changelog sections.
func (f *KeepAChangelogFormatter) regroupCommits(
	grouped map[string][]*GroupedCommit,
) map[string][]*GroupedCommit {
	result := make(map[string][]*GroupedCommit)

	// Iterate through all grouped commits and remap them
	for _, commits := range grouped {
		for _, commit := range commits {
			section := f.mapTypeToSection(commit)
			if section != "" {
				result[section] = append(result[section], commit)
			}
		}
	}

	return result
}

// mapTypeToSection maps a commit type to a Keep a Changelog section.
func (f *KeepAChangelogFormatter) mapTypeToSection(commit *GroupedCommit) string {
	// Breaking changes get their own section at the top
	if commit.Breaking {
		return "Breaking Changes"
	}

	// Map conventional commit types to Keep a Changelog sections
	switch commit.Type {
	case "feat":
		return "Added"
	case "fix":
		return "Fixed"
	case "refactor", "perf", "style":
		return "Changed"
	case "docs", "test", "chore", "ci", "build":
		// These types are typically not included in Keep a Changelog
		// as they don't affect the user-facing functionality
		return ""
	case "revert":
		return "Removed"
	default:
		// For unknown types, include in Changed if they have content
		if commit.Type != "" {
			return "Changed"
		}
		// Non-conventional commits are skipped unless explicitly included
		return ""
	}
}
