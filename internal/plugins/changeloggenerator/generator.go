package changeloggenerator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Generator handles changelog content generation.
type Generator struct {
	config    *Config
	remote    *RemoteInfo
	formatter Formatter
}

// NewGenerator creates a new changelog generator.
func NewGenerator(config *Config) (*Generator, error) {
	formatter, err := NewFormatter(config.Format, config)
	if err != nil {
		return nil, err
	}

	return &Generator{
		config:    config,
		formatter: formatter,
	}, nil
}

// resolveRemote resolves repository info from config or git remote.
func (g *Generator) resolveRemote() (*RemoteInfo, error) {
	if g.remote != nil {
		return g.remote, nil
	}

	if g.config.Repository != nil {
		if g.config.Repository.Owner != "" && g.config.Repository.Repo != "" {
			g.remote = &RemoteInfo{
				Provider: g.config.Repository.Provider,
				Host:     g.config.Repository.Host,
				Owner:    g.config.Repository.Owner,
				Repo:     g.config.Repository.Repo,
			}
			// Fill in defaults if not specified
			if g.remote.Host == "" {
				g.remote.Host = getDefaultHost(g.remote.Provider)
			}
			if g.remote.Provider == "" {
				g.remote.Provider = getProviderFromHost(g.remote.Host)
			}
			return g.remote, nil
		}

		if g.config.Repository.AutoDetect {
			remote, err := GetRemoteInfoFn()
			if err != nil {
				return nil, err
			}
			g.remote = remote
			return g.remote, nil
		}
	}

	return nil, fmt.Errorf("repository configuration not available")
}

// getDefaultHost returns the default host for a provider.
func getDefaultHost(provider string) string {
	switch provider {
	case "github":
		return "github.com"
	case "gitlab":
		return "gitlab.com"
	case "codeberg":
		return "codeberg.org"
	case "gitea":
		return "gitea.io"
	case "bitbucket":
		return "bitbucket.org"
	case "sourcehut":
		return "sr.ht"
	default:
		return ""
	}
}

// getProviderFromHost returns the provider name for a known host.
func getProviderFromHost(host string) string {
	if p, ok := KnownProviders[host]; ok {
		return p
	}
	return "custom"
}

// GenerateResult contains the generated changelog and any warnings.
type GenerateResult struct {
	Content                string
	SkippedNonConventional []*ParsedCommit
}

// GenerateVersionChangelog generates the changelog content for a version.
func (g *Generator) GenerateVersionChangelog(version, previousVersion string, commits []CommitInfo) (string, error) {
	result := g.GenerateVersionChangelogWithResult(version, previousVersion, commits)
	return result.Content, nil
}

// GenerateVersionChangelogWithResult generates the changelog content and returns detailed result.
func (g *Generator) GenerateVersionChangelogWithResult(version, previousVersion string, commits []CommitInfo) GenerateResult {
	// Parse and filter commits
	parsed := ParseCommits(commits)
	filtered := FilterCommits(parsed, g.config.ExcludePatterns)

	// Group commits with options
	groupResult := GroupCommitsWithOptions(filtered, g.config.Groups, g.config.IncludeNonConventional)
	grouped := groupResult.Grouped
	sortedKeys := SortedGroupKeys(grouped)

	// Resolve remote for links
	remote, _ := g.resolveRemote() // Ignore error, just won't have links

	// Use formatter to generate the main changelog content
	var sb strings.Builder
	content := g.formatter.FormatChangelog(version, previousVersion, grouped, sortedKeys, remote)
	sb.WriteString(content)

	// Contributors section (applies to both formats)
	if g.config.Contributors != nil && g.config.Contributors.Enabled {
		contributors := GetContributorsFn(commits)
		if len(contributors) > 0 {
			if g.config.Contributors.Icon != "" {
				sb.WriteString(fmt.Sprintf("### %s Contributors\n\n", g.config.Contributors.Icon))
			} else {
				sb.WriteString("### Contributors\n\n")
			}
			for _, contrib := range contributors {
				g.writeContributorEntry(&sb, contrib, remote)
			}
			sb.WriteString("\n")
		}
	}

	return GenerateResult{
		Content:                sb.String(),
		SkippedNonConventional: groupResult.SkippedNonConventional,
	}
}

// contributorTemplateData holds data for contributor template rendering.
type contributorTemplateData struct {
	Name     string
	Username string
	Email    string
	Host     string
}

// writeContributorEntry writes a contributor entry to the builder using the configured format template.
func (g *Generator) writeContributorEntry(sb *strings.Builder, contrib Contributor, remote *RemoteInfo) {
	// Determine the host for URL generation
	host := ""
	if remote != nil {
		host = remote.Host
	}
	if contrib.Host != "" {
		host = contrib.Host
	}

	// If no host available, fall back to simple format
	if host == "" {
		fmt.Fprintf(sb, "- @%s\n", contrib.Username)
		return
	}

	// Get format template
	format := g.config.Contributors.Format
	if format == "" {
		format = "- [@{{.Username}}](https://{{.Host}}/{{.Username}})"
	}

	// Parse and execute template
	tmpl, err := template.New("contributor").Parse(format)
	if err != nil {
		// Fallback on template error
		fmt.Fprintf(sb, "- [@%s](https://%s/%s)\n", contrib.Username, host, contrib.Username)
		return
	}

	data := contributorTemplateData{
		Name:     contrib.Name,
		Username: contrib.Username,
		Email:    contrib.Email,
		Host:     host,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		// Fallback on execution error
		fmt.Fprintf(sb, "- [@%s](https://%s/%s)\n", contrib.Username, host, contrib.Username)
		return
	}

	sb.WriteString(buf.String())
	sb.WriteString("\n")
}

// WriteVersionedFile writes the changelog to a version-specific file.
func (g *Generator) WriteVersionedFile(version, content string) error {
	dir := g.config.ChangesDir
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create changes directory: %w", err)
	}

	filename := fmt.Sprintf("%s.md", version)
	path := filepath.Join(dir, filename)

	// Normalize content: trim trailing whitespace and ensure single trailing newline
	normalizedContent := strings.TrimRight(content, "\n\r\t ") + "\n"

	if err := os.WriteFile(path, []byte(normalizedContent), 0644); err != nil {
		return fmt.Errorf("failed to write changelog file: %w", err)
	}

	return nil
}

// WriteUnifiedChangelog writes to the unified CHANGELOG.md file.
func (g *Generator) WriteUnifiedChangelog(newContent string) error {
	path := g.config.ChangelogPath

	var existingContent string

	// Read existing content if file exists
	if data, err := os.ReadFile(path); err == nil {
		existingContent = string(data)
	}

	var finalContent string

	if existingContent == "" {
		// Create new changelog with header
		header := g.getDefaultHeader()
		finalContent = header + "\n\n" + strings.TrimRight(newContent, "\n\r\t ")
	} else {
		// Insert new content after header
		finalContent = g.insertAfterHeader(existingContent, newContent)
	}

	// Normalize: trim trailing whitespace and ensure single trailing newline
	finalContent = strings.TrimRight(finalContent, "\n\r\t ") + "\n"

	if err := os.WriteFile(path, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write changelog: %w", err)
	}

	return nil
}

// getDefaultHeader returns the default changelog header.
func (g *Generator) getDefaultHeader() string {
	// Try to read custom header template
	if g.config.HeaderTemplate != "" {
		if data, err := os.ReadFile(g.config.HeaderTemplate); err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	return `# Changelog

All notable changes to this project will be documented in this file.

The format adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).`
}

// insertAfterHeader inserts new content after the changelog header.
func (g *Generator) insertAfterHeader(existing, newContent string) string {
	lines := strings.Split(existing, "\n")

	// Find the first version header (## v... or ## [v...)
	insertIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			insertIdx = i
			break
		}
	}

	if insertIdx == -1 {
		// No existing version found, append after header
		return existing + "\n" + newContent
	}

	// Insert new content before the first version
	before := strings.Join(lines[:insertIdx], "\n")
	after := strings.Join(lines[insertIdx:], "\n")

	// Ensure proper spacing
	before = strings.TrimRight(before, "\n") + "\n\n"

	return before + newContent + after
}

// MergeVersionedFiles merges all versioned changelog files into a unified CHANGELOG.md.
func (g *Generator) MergeVersionedFiles() error {
	dir := g.config.ChangesDir

	// Read all version files
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read changes directory: %w", err)
	}

	// Collect version files (excluding header template and directories)
	var versionFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "v") && strings.HasSuffix(name, ".md") {
			versionFiles = append(versionFiles, filepath.Join(dir, name))
		}
	}

	if len(versionFiles) == 0 {
		return nil // Nothing to merge
	}

	// Sort files by version (newest first)
	sortVersionFiles(versionFiles)

	// Build merged content
	var sb strings.Builder

	// Add header
	sb.WriteString(g.getDefaultHeader())
	sb.WriteString("\n\n")

	// Add each version's content
	for _, file := range versionFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip unreadable files
		}
		// Trim trailing whitespace and add consistent spacing
		content := strings.TrimRight(string(data), "\n\r\t ")
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	// Write to unified changelog
	path := g.config.ChangelogPath
	// Normalize: trim trailing whitespace and ensure single trailing newline
	finalContent := strings.TrimRight(sb.String(), "\n\r\t ") + "\n"
	if err := os.WriteFile(path, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write unified changelog: %w", err)
	}

	return nil
}

// sortVersionFiles sorts version files by semantic version (newest first).
func sortVersionFiles(files []string) {
	// Simple reverse lexicographic sort (works for vX.Y.Z format)
	for i := 1; i < len(files); i++ {
		for j := i; j > 0 && files[j] > files[j-1]; j-- {
			files[j], files[j-1] = files[j-1], files[j]
		}
	}
}
