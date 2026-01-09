package changeloggenerator

import (
	"strings"
	"testing"
)

func TestGenerateVersionChangelog(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}
	cfg.Contributors = &ContributorsConfig{Enabled: false}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commits := []CommitInfo{
		{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add feature", Author: "Alice", AuthorEmail: "alice@example.com"},
		{Hash: "def456", ShortHash: "def456", Subject: "fix: fix bug", Author: "Bob", AuthorEmail: "bob@example.com"},
	}

	content, err := g.GenerateVersionChangelog("v1.0.0", "v0.9.0", commits)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check version header
	if !strings.Contains(content, "## v1.0.0") {
		t.Error("expected version header")
	}

	// Check compare link
	if !strings.Contains(content, "compare/v0.9.0...v1.0.0") {
		t.Error("expected compare link")
	}

	// Check grouped content
	if !strings.Contains(content, "add feature") {
		t.Error("expected feature description")
	}
	if !strings.Contains(content, "fix bug") {
		t.Error("expected fix description")
	}
}

func TestGenerateVersionChangelog_WithContributors(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}
	cfg.Contributors = &ContributorsConfig{Enabled: true}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commits := []CommitInfo{
		{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add feature", Author: "Alice", AuthorEmail: "alice@users.noreply.github.com"},
	}

	content, err := g.GenerateVersionChangelog("v1.0.0", "v0.9.0", commits)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check contributors section
	if !strings.Contains(content, "### Contributors") {
		t.Error("expected contributors section")
	}
	// Username is extracted from noreply email: alice@users.noreply.github.com -> alice
	if !strings.Contains(content, "@alice") {
		t.Error("expected @alice in contributors")
	}
}

func TestGenerateVersionChangelog_WithContributorsIcon(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}
	cfg.Contributors = &ContributorsConfig{Enabled: true, Icon: "❤️"}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commits := []CommitInfo{
		{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add feature", Author: "Alice", AuthorEmail: "alice@users.noreply.github.com"},
	}

	content, err := g.GenerateVersionChangelog("v1.0.0", "v0.9.0", commits)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check contributors section with icon
	if !strings.Contains(content, "### ❤️ Contributors") {
		t.Error("expected contributors section with icon")
	}
	// Username is extracted from noreply email: alice@users.noreply.github.com -> alice
	if !strings.Contains(content, "@alice") {
		t.Error("expected @alice in contributors")
	}
}

func TestGenerateVersionChangelog_WithCustomContributorFormat(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}
	// Custom format that includes both Name and Username
	cfg.Contributors = &ContributorsConfig{
		Enabled: true,
		Format:  "- {{.Name}} ([@{{.Username}}](https://{{.Host}}/{{.Username}}))",
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commits := []CommitInfo{
		{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add feature", Author: "Alice Smith", AuthorEmail: "alice@users.noreply.github.com"},
	}

	content, err := g.GenerateVersionChangelog("v1.0.0", "v0.9.0", commits)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check contributors section includes both Name and Username
	if !strings.Contains(content, "### Contributors") {
		t.Error("expected contributors section")
	}
	if !strings.Contains(content, "Alice Smith") {
		t.Error("expected full name 'Alice Smith' in contributors with custom format")
	}
	if !strings.Contains(content, "@alice") {
		t.Error("expected @alice in contributors")
	}
	if !strings.Contains(content, "github.com/alice") {
		t.Error("expected github.com/alice link in contributors")
	}
}

func TestGenerateVersionChangelog_EmptyCommits(t *testing.T) {
	cfg := DefaultConfig()
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := g.GenerateVersionChangelog("v1.0.0", "v0.9.0", []CommitInfo{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still have version header
	if !strings.Contains(content, "## v1.0.0") {
		t.Error("expected version header even with no commits")
	}
}

func TestGenerateVersionChangelog_NoRemote(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = nil
	cfg.Contributors = &ContributorsConfig{Enabled: false}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commits := []CommitInfo{
		{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add feature", Author: "Alice", AuthorEmail: "alice@example.com"},
	}

	content, err := g.GenerateVersionChangelog("v1.0.0", "v0.9.0", commits)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have version header
	if !strings.Contains(content, "## v1.0.0") {
		t.Error("expected version header")
	}

	// Should NOT have compare link (no remote)
	if strings.Contains(content, "compare") {
		t.Error("did not expect compare link without remote")
	}
}

func TestGenerateVersionChangelog_NoPreviousVersion(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}
	cfg.Contributors = &ContributorsConfig{Enabled: false}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commits := []CommitInfo{
		{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add feature", Author: "Alice", AuthorEmail: "alice@example.com"},
	}

	// Empty previous version
	content, err := g.GenerateVersionChangelog("v1.0.0", "", commits)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have version header
	if !strings.Contains(content, "## v1.0.0") {
		t.Error("expected version header")
	}

	// Should NOT have compare link (no previous version)
	if strings.Contains(content, "compare") {
		t.Error("did not expect compare link without previous version")
	}
}

func TestGenerateVersionChangelog_WithNewContributors(t *testing.T) {
	// Save and restore original functions
	originalGetNewContributorsFn := GetNewContributorsFn
	originalGetContributorsFn := GetContributorsFn
	defer func() {
		GetNewContributorsFn = originalGetNewContributorsFn
		GetContributorsFn = originalGetContributorsFn
	}()

	// Mock new contributors
	GetNewContributorsFn = func(commits []CommitInfo, previousVersion string) ([]NewContributor, error) {
		return []NewContributor{
			{
				Contributor: Contributor{
					Name:     "New Dev",
					Username: "newdev",
					Host:     "github.com",
				},
				FirstCommit: CommitInfo{ShortHash: "abc123"},
				PRNumber:    "42",
			},
		}, nil
	}

	// Mock contributors
	GetContributorsFn = func(commits []CommitInfo) []Contributor {
		return []Contributor{
			{Name: "New Dev", Username: "newdev", Host: "github.com"},
		}
	}

	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:             true,
		ShowNewContributors: true,
	}
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}

	g, _ := NewGenerator(cfg)

	commits := []CommitInfo{
		{Hash: "abc123", ShortHash: "abc123", Subject: "feat: add feature (#42)", Author: "New Dev", AuthorEmail: "newdev@users.noreply.github.com"},
	}

	content, _ := g.GenerateVersionChangelog("v1.0.0", "v0.9.0", commits)

	if !strings.Contains(content, "New Contributors") {
		t.Error("expected New Contributors section in output")
	}
	if !strings.Contains(content, "Full Changelog") {
		t.Error("expected Full Changelog link in output")
	}
	if !strings.Contains(content, "Contributors") {
		t.Error("expected Contributors section in output")
	}
}
