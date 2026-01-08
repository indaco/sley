package changeloggenerator

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Pre-compiled regexes for URL parsing (compiled once at package init).
var (
	// Remote URL formats
	sshRemoteRe   = regexp.MustCompile(`git@([^:]+):([^/]+)/([^/]+?)(?:\.git)?$`)
	httpsRemoteRe = regexp.MustCompile(`https?://([^/]+)/([^/]+)/([^/]+?)(?:\.git)?$`)
	gitRemoteRe   = regexp.MustCompile(`git://([^/]+)/([^/]+)/([^/]+?)(?:\.git)?$`)

	// Noreply email formats for username extraction
	githubNoreplyRe   = regexp.MustCompile(`(?:\d+\+)?([^@]+)@users\.noreply\.github\.com`)
	gitlabNoreplyRe   = regexp.MustCompile(`([^@]+)@noreply\.gitlab\.com`)
	codebergNoreplyRe = regexp.MustCompile(`([^@]+)@noreply\.codeberg\.org`)
)

// CommitInfo represents a git commit with metadata.
type CommitInfo struct {
	Hash        string
	ShortHash   string
	Subject     string
	Author      string
	AuthorEmail string
}

// RemoteInfo holds parsed git remote information.
// Supports multiple providers: github, gitlab, codeberg, gitea, bitbucket, etc.
type RemoteInfo struct {
	Provider string // github, gitlab, codeberg, gitea, bitbucket, custom
	Host     string // e.g., github.com, gitlab.com, codeberg.org
	Owner    string
	Repo     string
}

// KnownProviders maps hostnames to provider names.
var KnownProviders = map[string]string{
	"github.com":    "github",
	"gitlab.com":    "gitlab",
	"codeberg.org":  "codeberg",
	"gitea.io":      "gitea",
	"bitbucket.org": "bitbucket",
	"sr.ht":         "sourcehut",
}

// Mockable functions for testing.
var (
	execCommand                 = exec.Command
	GetCommitsWithMetaFn        = getCommitsWithMeta
	GetRemoteInfoFn             = getRemoteInfo
	GetLatestTagFn              = getLatestTag
	GetContributorsFn           = getContributors
	GetHistoricalContributorsFn = getHistoricalContributors
	GetNewContributorsFn        = getNewContributors
)

// getCommitsWithMeta retrieves commits between two refs with full metadata.
// Format: hash|short_hash|subject|author|email
func getCommitsWithMeta(since, until string) ([]CommitInfo, error) {
	if until == "" {
		until = "HEAD"
	}

	if since == "" {
		lastTag, err := getLatestTag()
		if err != nil {
			since = "HEAD~10"
		} else {
			since = lastTag
		}
	}

	revRange := since + ".." + until
	// Use a delimiter that's unlikely to appear in commit messages
	format := "%H|%h|%s|%an|%ae"
	cmd := execCommand("git", "log", "--pretty=format:"+format, revRange)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return nil, fmt.Errorf("git log failed: %s: %w", stderrMsg, err)
		}
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []CommitInfo{}, nil
	}

	commits := make([]CommitInfo, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue // Skip malformed lines
		}
		commits = append(commits, CommitInfo{
			Hash:        parts[0],
			ShortHash:   parts[1],
			Subject:     parts[2],
			Author:      parts[3],
			AuthorEmail: parts[4],
		})
	}

	return commits, nil
}

// getLatestTag returns the most recent git tag.
func getLatestTag() (string, error) {
	cmd := execCommand("git", "describe", "--tags", "--abbrev=0")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return "", fmt.Errorf("git describe failed: %s: %w", stderrMsg, err)
		}
		return "", fmt.Errorf("git describe failed: %w", err)
	}

	tag := strings.TrimSpace(string(out))
	if tag == "" {
		return "", fmt.Errorf("no tags found")
	}

	return tag, nil
}

// getRemoteInfo parses the owner/repo from git remote origin.
// Supports multiple git hosting providers.
func getRemoteInfo() (*RemoteInfo, error) {
	cmd := execCommand("git", "remote", "get-url", "origin")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return nil, fmt.Errorf("git remote get-url failed: %s: %w", stderrMsg, err)
		}
		return nil, fmt.Errorf("git remote get-url failed: %w", err)
	}

	url := strings.TrimSpace(string(out))
	return parseRemoteURL(url)
}

// parseRemoteURL extracts host, owner and repo from various git URL formats.
// Supports any git hosting provider with standard URL formats:
// - git@host:owner/repo.git (SSH)
// - https://host/owner/repo.git (HTTPS)
// - git://host/owner/repo.git (Git protocol)
func parseRemoteURL(url string) (*RemoteInfo, error) {
	// SSH format: git@host:owner/repo.git
	if matches := sshRemoteRe.FindStringSubmatch(url); len(matches) == 4 {
		return buildRemoteInfo(matches[1], matches[2], matches[3]), nil
	}

	// HTTPS format: https://host/owner/repo.git or https://host/owner/repo
	if matches := httpsRemoteRe.FindStringSubmatch(url); len(matches) == 4 {
		return buildRemoteInfo(matches[1], matches[2], matches[3]), nil
	}

	// Git protocol: git://host/owner/repo.git
	if matches := gitRemoteRe.FindStringSubmatch(url); len(matches) == 4 {
		return buildRemoteInfo(matches[1], matches[2], matches[3]), nil
	}

	return nil, fmt.Errorf("could not parse remote URL: %s", url)
}

// buildRemoteInfo creates a RemoteInfo with provider detection.
func buildRemoteInfo(host, owner, repo string) *RemoteInfo {
	provider := "custom"
	if p, ok := KnownProviders[host]; ok {
		provider = p
	}
	return &RemoteInfo{
		Provider: provider,
		Host:     host,
		Owner:    owner,
		Repo:     repo,
	}
}

// Contributor represents a unique contributor.
type Contributor struct {
	Name     string
	Username string
	Email    string
	Host     string // The git host for URL generation
}

// getContributors extracts unique contributors from commits.
func getContributors(commits []CommitInfo) []Contributor {
	seen := make(map[string]bool)
	contributors := make([]Contributor, 0)

	for _, c := range commits {
		key := c.AuthorEmail
		if seen[key] {
			continue
		}
		seen[key] = true

		// Extract username from email if it follows known patterns
		username, host := extractUsername(c.AuthorEmail, c.Author)
		contributors = append(contributors, Contributor{
			Name:     c.Author,
			Username: username,
			Email:    c.AuthorEmail,
			Host:     host,
		})
	}

	return contributors
}

// extractUsername tries to extract username and host from noreply email addresses.
// Supports GitHub, GitLab, and other providers.
func extractUsername(email, authorName string) (username string, host string) {
	// GitHub noreply format: 12345+username@users.noreply.github.com
	// or: username@users.noreply.github.com
	if matches := githubNoreplyRe.FindStringSubmatch(email); len(matches) == 2 {
		return matches[1], "github.com"
	}

	// GitLab noreply format: username@noreply.gitlab.com
	if matches := gitlabNoreplyRe.FindStringSubmatch(email); len(matches) == 2 {
		return matches[1], "gitlab.com"
	}

	// Codeberg noreply format: username@noreply.codeberg.org
	if matches := codebergNoreplyRe.FindStringSubmatch(email); len(matches) == 2 {
		return matches[1], "codeberg.org"
	}

	// Fall back to using author name converted to lowercase with spaces removed
	// This is a best-effort guess
	return strings.ToLower(strings.ReplaceAll(authorName, " ", "")), ""
}

// NewContributor represents a first-time contributor with their first contribution details.
type NewContributor struct {
	Contributor
	FirstCommit CommitInfo // The commit that introduced this contributor
	PRNumber    string     // PR number if available (extracted from commit message)
}

// getHistoricalContributors returns all unique contributor usernames before a given ref.
// This is used to determine if a contributor is new (first-time) in this release.
func getHistoricalContributors(beforeRef string) (map[string]struct{}, error) {
	if beforeRef == "" {
		return make(map[string]struct{}), nil
	}

	// git log --format="%ae|%an" beforeRef
	// Returns all author emails and names from the beginning of history up to beforeRef
	cmd := execCommand("git", "log", "--format=%ae|%an", beforeRef)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		// If the ref doesn't exist (e.g., first release), return empty set
		if strings.Contains(stderrMsg, "unknown revision") ||
			strings.Contains(stderrMsg, "bad revision") {
			return make(map[string]struct{}), nil
		}
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	usernames := make(map[string]struct{})
	for line := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) < 2 {
			continue
		}
		email, authorName := parts[0], parts[1]
		username, _ := extractUsername(email, authorName)
		if username != "" {
			usernames[username] = struct{}{}
		}
	}

	return usernames, nil
}

// prNumberExtractRe extracts PR number from commit messages like "(#123)" or "Merge pull request #123".
var prNumberExtractRe = regexp.MustCompile(`#(\d+)`)

// getNewContributors identifies first-time contributors in a set of commits.
// It checks if the contributor has any commits before previousVersion.
func getNewContributors(commits []CommitInfo, previousVersion string) ([]NewContributor, error) {
	// Get historical contributor usernames (before this release)
	historicalUsernames, err := GetHistoricalContributorsFn(previousVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical contributors: %w", err)
	}

	// Track which usernames we've already processed in this release
	seenInRelease := make(map[string]bool)
	var newContributors []NewContributor

	for _, commit := range commits {
		username, host := extractUsername(commit.AuthorEmail, commit.Author)
		if username == "" || seenInRelease[username] {
			continue
		}
		seenInRelease[username] = true

		// Check if this is a new contributor (not in historical set)
		if _, existed := historicalUsernames[username]; !existed {
			// Extract PR number from commit subject
			prNumber := ""
			if matches := prNumberExtractRe.FindStringSubmatch(commit.Subject); len(matches) == 2 {
				prNumber = matches[1]
			}

			newContributors = append(newContributors, NewContributor{
				Contributor: Contributor{
					Name:     commit.Author,
					Username: username,
					Email:    commit.AuthorEmail,
					Host:     host,
				},
				FirstCommit: commit,
				PRNumber:    prNumber,
			})
		}
	}

	return newContributors, nil
}
