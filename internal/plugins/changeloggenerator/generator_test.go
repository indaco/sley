package changeloggenerator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	cfg := DefaultConfig()
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g == nil {
		t.Fatal("expected non-nil generator")
	}
	if g.config != cfg {
		t.Error("expected config to match")
	}
}

func TestGetDefaultHost(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{"github", "github.com"},
		{"gitlab", "gitlab.com"},
		{"codeberg", "codeberg.org"},
		{"gitea", "gitea.io"},
		{"bitbucket", "bitbucket.org"},
		{"sourcehut", "sr.ht"},
		{"custom", ""},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := getDefaultHost(tt.provider)
			if got != tt.want {
				t.Errorf("getDefaultHost(%q) = %q, want %q", tt.provider, got, tt.want)
			}
		})
	}
}

func TestGetProviderFromHost(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{"github.com", "github"},
		{"gitlab.com", "gitlab"},
		{"codeberg.org", "codeberg"},
		{"bitbucket.org", "bitbucket"},
		{"custom.server.com", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := getProviderFromHost(tt.host)
			if got != tt.want {
				t.Errorf("getProviderFromHost(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}

func TestBuildCompareURL(t *testing.T) {
	tests := []struct {
		name     string
		remote   *RemoteInfo
		prev     string
		curr     string
		contains string
	}{
		{
			name:     "GitHub",
			remote:   &RemoteInfo{Provider: "github", Host: "github.com", Owner: "owner", Repo: "repo"},
			prev:     "v1.0.0",
			curr:     "v1.1.0",
			contains: "github.com/owner/repo/compare/v1.0.0...v1.1.0",
		},
		{
			name:     "GitLab",
			remote:   &RemoteInfo{Provider: "gitlab", Host: "gitlab.com", Owner: "group", Repo: "project"},
			prev:     "v1.0.0",
			curr:     "v1.1.0",
			contains: "gitlab.com/group/project/-/compare/v1.0.0...v1.1.0",
		},
		{
			name:     "Bitbucket",
			remote:   &RemoteInfo{Provider: "bitbucket", Host: "bitbucket.org", Owner: "team", Repo: "repo"},
			prev:     "v1.0.0",
			curr:     "v1.1.0",
			contains: "bitbucket.org/team/repo/branches/compare",
		},
		{
			name:     "Codeberg",
			remote:   &RemoteInfo{Provider: "codeberg", Host: "codeberg.org", Owner: "user", Repo: "project"},
			prev:     "v1.0.0",
			curr:     "v1.1.0",
			contains: "codeberg.org/user/project/compare/v1.0.0...v1.1.0",
		},
		{
			name:     "Sourcehut",
			remote:   &RemoteInfo{Provider: "sourcehut", Host: "sr.ht", Owner: "~user", Repo: "repo"},
			prev:     "v1.0.0",
			curr:     "v1.1.0",
			contains: "git.sr.ht/~user/repo/log/v1.0.0..v1.1.0",
		},
		{
			name:     "Gitea",
			remote:   &RemoteInfo{Provider: "gitea", Host: "gitea.io", Owner: "org", Repo: "repo"},
			prev:     "v1.0.0",
			curr:     "v1.1.0",
			contains: "gitea.io/org/repo/compare/v1.0.0...v1.1.0",
		},
		{
			name:     "Custom",
			remote:   &RemoteInfo{Provider: "custom", Host: "git.example.com", Owner: "org", Repo: "repo"},
			prev:     "v1.0.0",
			curr:     "v1.1.0",
			contains: "git.example.com/org/repo/compare/v1.0.0...v1.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCompareURL(tt.remote, tt.prev, tt.curr)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("buildCompareURL() = %q, expected to contain %q", got, tt.contains)
			}
		})
	}
}

func TestBuildCommitURL(t *testing.T) {

	tests := []struct {
		name     string
		remote   *RemoteInfo
		hash     string
		contains string
	}{
		{
			name:     "GitHub",
			remote:   &RemoteInfo{Provider: "github", Host: "github.com", Owner: "owner", Repo: "repo"},
			hash:     "abc123",
			contains: "github.com/owner/repo/commit/abc123",
		},
		{
			name:     "GitLab",
			remote:   &RemoteInfo{Provider: "gitlab", Host: "gitlab.com", Owner: "group", Repo: "project"},
			hash:     "def456",
			contains: "gitlab.com/group/project/-/commit/def456",
		},
		{
			name:     "Bitbucket",
			remote:   &RemoteInfo{Provider: "bitbucket", Host: "bitbucket.org", Owner: "team", Repo: "repo"},
			hash:     "ghi789",
			contains: "bitbucket.org/team/repo/commits/ghi789",
		},
		{
			name:     "Sourcehut",
			remote:   &RemoteInfo{Provider: "sourcehut", Host: "sr.ht", Owner: "~user", Repo: "repo"},
			hash:     "jkl012",
			contains: "git.sr.ht/~user/repo/commit/jkl012",
		},
		{
			name:     "Codeberg",
			remote:   &RemoteInfo{Provider: "codeberg", Host: "codeberg.org", Owner: "user", Repo: "project"},
			hash:     "mno345",
			contains: "codeberg.org/user/project/commit/mno345",
		},
		{
			name:     "Gitea",
			remote:   &RemoteInfo{Provider: "gitea", Host: "gitea.io", Owner: "org", Repo: "repo"},
			hash:     "pqr678",
			contains: "gitea.io/org/repo/commit/pqr678",
		},
		{
			name:     "Custom",
			remote:   &RemoteInfo{Provider: "custom", Host: "git.example.com", Owner: "org", Repo: "repo"},
			hash:     "stu901",
			contains: "git.example.com/org/repo/commit/stu901",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCommitURL(tt.remote, tt.hash)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("buildCommitURL() = %q, expected to contain %q", got, tt.contains)
			}
		})
	}
}

func TestBuildPRURL(t *testing.T) {

	tests := []struct {
		name     string
		remote   *RemoteInfo
		prNumber string
		contains string
	}{
		{
			name:     "GitHub",
			remote:   &RemoteInfo{Provider: "github", Host: "github.com", Owner: "owner", Repo: "repo"},
			prNumber: "123",
			contains: "github.com/owner/repo/pull/123",
		},
		{
			name:     "GitLab",
			remote:   &RemoteInfo{Provider: "gitlab", Host: "gitlab.com", Owner: "group", Repo: "project"},
			prNumber: "456",
			contains: "gitlab.com/group/project/-/merge_requests/456",
		},
		{
			name:     "Bitbucket",
			remote:   &RemoteInfo{Provider: "bitbucket", Host: "bitbucket.org", Owner: "team", Repo: "repo"},
			prNumber: "789",
			contains: "bitbucket.org/team/repo/pull-requests/789",
		},
		{
			name:     "Codeberg",
			remote:   &RemoteInfo{Provider: "codeberg", Host: "codeberg.org", Owner: "user", Repo: "project"},
			prNumber: "42",
			contains: "codeberg.org/user/project/pull/42",
		},
		{
			name:     "Gitea",
			remote:   &RemoteInfo{Provider: "gitea", Host: "gitea.io", Owner: "org", Repo: "repo"},
			prNumber: "99",
			contains: "gitea.io/org/repo/pull/99",
		},
		{
			name:     "Custom",
			remote:   &RemoteInfo{Provider: "custom", Host: "git.example.com", Owner: "org", Repo: "repo"},
			prNumber: "77",
			contains: "git.example.com/org/repo/pull/77",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPRURL(tt.remote, tt.prNumber)
			if !strings.Contains(got, tt.contains) {
				t.Errorf("buildPRURL() = %q, expected to contain %q", got, tt.contains)
			}
		})
	}
}

func TestFormatCommitEntry(t *testing.T) {
	remote := &RemoteInfo{Provider: "github", Host: "github.com", Owner: "owner", Repo: "repo"}

	tests := []struct {
		name     string
		commit   *GroupedCommit
		remote   *RemoteInfo
		contains []string
	}{
		{
			name: "Basic commit",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{ShortHash: "abc123"},
					Description: "add feature",
				},
			},
			remote:   remote,
			contains: []string{"- add feature", "abc123"},
		},
		{
			name: "Commit with scope",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{ShortHash: "def456"},
					Description: "update config",
					Scope:       "cli",
				},
			},
			remote:   remote,
			contains: []string{"**cli:**", "update config"},
		},
		{
			name: "Commit with PR number",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{ShortHash: "ghi789"},
					Description: "fix bug",
					PRNumber:    "42",
				},
			},
			remote:   remote,
			contains: []string{"fix bug", "ghi789", "#42", "pull/42"},
		},
		{
			name: "Commit without remote",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{ShortHash: "jkl012"},
					Description: "simple change",
				},
			},
			remote:   nil,
			contains: []string{"- simple change"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCommitEntry(tt.commit, tt.remote)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("formatCommitEntry() = %q, expected to contain %q", got, want)
				}
			}
		})
	}
}

func TestWriteContributorEntry(t *testing.T) {
	g, err := NewGenerator(DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	remote := &RemoteInfo{Provider: "github", Host: "github.com", Owner: "owner", Repo: "repo"}

	tests := []struct {
		name     string
		contrib  Contributor
		remote   *RemoteInfo
		contains []string
	}{
		{
			name:     "With remote",
			contrib:  Contributor{Name: "Alice", Username: "alice", Host: "github.com"},
			remote:   remote,
			contains: []string{"@alice", "github.com/alice"},
		},
		{
			name:     "Without remote",
			contrib:  Contributor{Name: "Bob", Username: "bob"},
			remote:   nil,
			contains: []string{"- @bob"},
		},
		{
			name:     "Contributor with different host",
			contrib:  Contributor{Name: "Charlie", Username: "charlie", Host: "gitlab.com"},
			remote:   remote,
			contains: []string{"@charlie", "gitlab.com/charlie"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			g.writeContributorEntry(&sb, tt.contrib, tt.remote)
			got := sb.String()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("writeContributorEntry() = %q, expected to contain %q", got, want)
				}
			}
		})
	}
}

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

func TestWriteContributorEntry_CustomFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		contrib  Contributor
		remote   *RemoteInfo
		expected string
	}{
		{
			name:     "Custom format with name and username",
			format:   "- {{.Name}} (@{{.Username}})",
			contrib:  Contributor{Name: "Alice Smith", Username: "alice", Host: "github.com"},
			remote:   &RemoteInfo{Host: "github.com", Owner: "test", Repo: "repo"},
			expected: "- Alice Smith (@alice)\n",
		},
		{
			name:     "Custom format username only",
			format:   "- @{{.Username}}",
			contrib:  Contributor{Name: "Bob Jones", Username: "bob", Host: "github.com"},
			remote:   &RemoteInfo{Host: "github.com", Owner: "test", Repo: "repo"},
			expected: "- @bob\n",
		},
		{
			name:     "Custom format with email",
			format:   "- {{.Username}} <{{.Email}}>",
			contrib:  Contributor{Name: "Charlie", Username: "charlie", Email: "charlie@example.com", Host: "github.com"},
			remote:   &RemoteInfo{Host: "github.com", Owner: "test", Repo: "repo"},
			expected: "- charlie <charlie@example.com>\n",
		},
		{
			name:     "Default format when empty",
			format:   "",
			contrib:  Contributor{Name: "Dave", Username: "dave", Host: "github.com"},
			remote:   &RemoteInfo{Host: "github.com", Owner: "test", Repo: "repo"},
			expected: "- [@dave](https://github.com/dave)\n",
		},
		{
			name:     "Fallback on invalid template",
			format:   "- {{.Invalid",
			contrib:  Contributor{Name: "Eve", Username: "eve", Host: "github.com"},
			remote:   &RemoteInfo{Host: "github.com", Owner: "test", Repo: "repo"},
			expected: "- [@eve](https://github.com/eve)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Contributors.Format = tt.format
			g, err := NewGenerator(cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var sb strings.Builder
			g.writeContributorEntry(&sb, tt.contrib, tt.remote)
			got := sb.String()

			if got != tt.expected {
				t.Errorf("writeContributorEntry() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWriteContributorEntry_NoHost(t *testing.T) {
	cfg := DefaultConfig()
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contrib := Contributor{Name: "NoHost User", Username: "nohost", Host: ""}

	var sb strings.Builder
	g.writeContributorEntry(&sb, contrib, nil)
	got := sb.String()

	expected := "- @nohost\n"
	if got != expected {
		t.Errorf("writeContributorEntry() = %q, want %q", got, expected)
	}
}

func TestWriteVersionedFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.ChangesDir = filepath.Join(tmpDir, ".changes")
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := "## v1.0.0\n\nTest content"
	err = g.WriteVersionedFile("v1.0.0", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check file exists
	expectedPath := filepath.Join(cfg.ChangesDir, "v1.0.0.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s", expectedPath)
	}

	// Check content - file should be normalized with single trailing newline
	data, readErr := os.ReadFile(expectedPath)
	if readErr != nil {
		t.Fatalf("failed to read file: %v", readErr)
	}
	expectedContent := "## v1.0.0\n\nTest content\n"
	if string(data) != expectedContent {
		t.Errorf("file content = %q, want %q", string(data), expectedContent)
	}
}

func TestWriteUnifiedChangelog_New(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.ChangelogPath = filepath.Join(tmpDir, "CHANGELOG.md")
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	newContent := "## v1.0.0\n\nNew content"
	err = g.WriteUnifiedChangelog(newContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(cfg.ChangelogPath); os.IsNotExist(err) {
		t.Error("expected CHANGELOG.md to be created")
	}

	// Check content includes header
	data, readErr := os.ReadFile(cfg.ChangelogPath)
	if readErr != nil {
		t.Fatalf("failed to read file: %v", readErr)
	}
	content := string(data)
	if !strings.Contains(content, "# Changelog") {
		t.Error("expected changelog header")
	}
	if !strings.Contains(content, "v1.0.0") {
		t.Error("expected version content")
	}
}

func TestWriteUnifiedChangelog_Existing(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.ChangelogPath = filepath.Join(tmpDir, "CHANGELOG.md")
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Create existing changelog
	existingContent := `# Changelog

## v0.9.0

Previous content
`
	if err := os.WriteFile(cfg.ChangelogPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to create existing changelog: %v", err)
	}

	// Write new content
	newContent := "## v1.0.0\n\nNew content\n\n"
	err = g.WriteUnifiedChangelog(newContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check content
	data, readErr := os.ReadFile(cfg.ChangelogPath)
	if readErr != nil {
		t.Fatalf("failed to read file: %v", readErr)
	}
	content := string(data)

	// New content should come before old
	v1Index := strings.Index(content, "v1.0.0")
	v09Index := strings.Index(content, "v0.9.0")
	if v1Index > v09Index {
		t.Error("expected new version to appear before old version")
	}
}

func TestGetDefaultHeader(t *testing.T) {
	cfg := DefaultConfig()
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	header := g.getDefaultHeader()

	if !strings.Contains(header, "Changelog") {
		t.Error("expected 'Changelog' in header")
	}
	if !strings.Contains(header, "Semantic Versioning") {
		t.Error("expected 'Semantic Versioning' in header")
	}
}

func TestGetDefaultHeader_CustomTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "header.md")
	customHeader := "# Custom Header\n\nCustom description"
	if err := os.WriteFile(templatePath, []byte(customHeader), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	cfg := DefaultConfig()
	cfg.HeaderTemplate = templatePath
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	header := g.getDefaultHeader()

	if header != strings.TrimSpace(customHeader) {
		t.Errorf("header = %q, want %q", header, strings.TrimSpace(customHeader))
	}
}

func TestInsertAfterHeader(t *testing.T) {
	cfg := DefaultConfig()
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	existing := `# Changelog

Some description

## v0.9.0

Old content
`
	newContent := "## v1.0.0\n\nNew content\n\n"

	result := g.insertAfterHeader(existing, newContent)

	// New content should be before v0.9.0
	v1Index := strings.Index(result, "v1.0.0")
	v09Index := strings.Index(result, "v0.9.0")
	if v1Index > v09Index {
		t.Error("expected new version to appear before old version")
	}
}

func TestSortVersionFiles(t *testing.T) {
	files := []string{
		"/tmp/.changes/v0.1.0.md",
		"/tmp/.changes/v1.0.0.md",
		"/tmp/.changes/v0.9.0.md",
	}

	sortVersionFiles(files)

	// Should be in reverse order (newest first)
	if files[0] != "/tmp/.changes/v1.0.0.md" {
		t.Errorf("expected v1.0.0.md first, got %s", files[0])
	}
	if files[1] != "/tmp/.changes/v0.9.0.md" {
		t.Errorf("expected v0.9.0.md second, got %s", files[1])
	}
	if files[2] != "/tmp/.changes/v0.1.0.md" {
		t.Errorf("expected v0.1.0.md third, got %s", files[2])
	}
}

func TestResolveRemote_FromConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Provider: "gitlab",
		Host:     "gitlab.com",
		Owner:    "mygroup",
		Repo:     "myproject",
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remote, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if remote.Provider != "gitlab" {
		t.Errorf("Provider = %q, want 'gitlab'", remote.Provider)
	}
	if remote.Host != "gitlab.com" {
		t.Errorf("Host = %q, want 'gitlab.com'", remote.Host)
	}
	if remote.Owner != "mygroup" {
		t.Errorf("Owner = %q, want 'mygroup'", remote.Owner)
	}
}

func TestResolveRemote_FillDefaults(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Owner:    "owner",
		Repo:     "repo",
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remote, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Host should be filled from provider
	if remote.Host != "github.com" {
		t.Errorf("Host = %q, want 'github.com'", remote.Host)
	}
}

func TestMergeVersionedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("failed to create changes dir: %v", err)
	}

	// Create version files
	files := map[string]string{
		"v0.1.0.md": "## v0.1.0\n\nFirst version\n",
		"v0.2.0.md": "## v0.2.0\n\nSecond version\n",
	}
	for name, content := range files {
		path := filepath.Join(changesDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	cfg := DefaultConfig()
	cfg.ChangesDir = changesDir
	cfg.ChangelogPath = filepath.Join(tmpDir, "CHANGELOG.md")
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = g.MergeVersionedFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check merged file
	data, readErr := os.ReadFile(cfg.ChangelogPath)
	if readErr != nil {
		t.Fatalf("failed to read changelog: %v", readErr)
	}
	content := string(data)

	if !strings.Contains(content, "v0.1.0") {
		t.Error("expected v0.1.0 in merged changelog")
	}
	if !strings.Contains(content, "v0.2.0") {
		t.Error("expected v0.2.0 in merged changelog")
	}
}

func TestMergeVersionedFiles_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("failed to create changes dir: %v", err)
	}

	cfg := DefaultConfig()
	cfg.ChangesDir = changesDir
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not error with empty directory
	err = g.MergeVersionedFiles()
	if err != nil {
		t.Errorf("unexpected error for empty dir: %v", err)
	}
}

func TestMergeVersionedFiles_NonexistentDir(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ChangesDir = "/nonexistent/path"
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should error with non-existent directory
	err = g.MergeVersionedFiles()
	if err == nil {
		t.Error("expected error for non-existent dir")
	}
}

func TestResolveRemote_AutoDetect(t *testing.T) {
	// Save and restore original function
	originalFn := GetRemoteInfoFn
	defer func() { GetRemoteInfoFn = originalFn }()

	// Mock GetRemoteInfoFn
	GetRemoteInfoFn = func() (*RemoteInfo, error) {
		return &RemoteInfo{
			Provider: "github",
			Host:     "github.com",
			Owner:    "autodetected",
			Repo:     "repo",
		}, nil
	}

	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		AutoDetect: true,
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remote, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if remote.Owner != "autodetected" {
		t.Errorf("Owner = %q, want 'autodetected'", remote.Owner)
	}
}

func TestResolveRemote_NoConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = nil
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = g.resolveRemote()
	if err == nil {
		t.Error("expected error when repository config is nil")
	}
}

func TestResolveRemote_Cached(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First call
	remote1, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call should return cached
	remote2, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if remote1 != remote2 {
		t.Error("expected cached remote to be returned")
	}
}

func TestWriteVersionedFile_Error(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ChangesDir = "/nonexistent/readonly/path"
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = g.WriteVersionedFile("v1.0.0", "content")
	if err == nil {
		t.Error("expected error for non-writable path")
	}
}

func TestWriteUnifiedChangelog_Error(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ChangelogPath = "/nonexistent/readonly/CHANGELOG.md"
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = g.WriteUnifiedChangelog("content")
	if err == nil {
		t.Error("expected error for non-writable path")
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

func TestInsertAfterHeader_NoVersionFound(t *testing.T) {
	cfg := DefaultConfig()
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Existing content with no version headers
	existing := `# Changelog

Some description about this project.
`
	newContent := "## v1.0.0\n\nNew content\n\n"

	result := g.insertAfterHeader(existing, newContent)

	// New content should be appended
	if !strings.Contains(result, "v1.0.0") {
		t.Error("expected new version in result")
	}
}

func TestWriteNewContributorsSection(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:             true,
		ShowNewContributors: true,
	}
	g, _ := NewGenerator(cfg)

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}

	newContributors := []NewContributor{
		{
			Contributor: Contributor{
				Name:     "New Dev",
				Username: "newdev",
				Host:     "github.com",
			},
			FirstCommit: CommitInfo{ShortHash: "abc123"},
			PRNumber:    "42",
		},
	}

	var sb strings.Builder
	g.writeNewContributorsSection(&sb, newContributors, remote)
	result := sb.String()

	if !strings.Contains(result, "### New Contributors") {
		t.Error("expected New Contributors header")
	}
	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in output")
	}
	if !strings.Contains(result, "#42") {
		t.Error("expected PR number in output")
	}
}

func TestWriteNewContributorsSection_WithIcon(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:             true,
		ShowNewContributors: true,
		NewContributorsIcon: "🎉",
	}
	g, _ := NewGenerator(cfg)

	newContributors := []NewContributor{
		{
			Contributor: Contributor{
				Name:     "New Dev",
				Username: "newdev",
				Host:     "github.com",
			},
			FirstCommit: CommitInfo{ShortHash: "abc123"},
			PRNumber:    "42",
		},
	}

	var sb strings.Builder
	g.writeNewContributorsSection(&sb, newContributors, nil)
	result := sb.String()

	if !strings.Contains(result, "### 🎉 New Contributors") {
		t.Error("expected New Contributors header with icon")
	}
}

func TestWriteNewContributorEntry_WithRemote(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:             true,
		ShowNewContributors: true,
	}
	g, _ := NewGenerator(cfg)

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, remote)
	result := sb.String()

	if !strings.Contains(result, "newdev") {
		t.Error("expected username in output")
	}
	if !strings.Contains(result, "first contribution") {
		t.Error("expected 'first contribution' text")
	}
}

func TestWriteNewContributorEntry_WithoutPR(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:             true,
		ShowNewContributors: true,
	}
	g, _ := NewGenerator(cfg)

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "", // No PR number
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, remote)
	result := sb.String()

	if !strings.Contains(result, "newdev") {
		t.Error("expected username in output")
	}
	// Should contain commit hash when no PR number
	if !strings.Contains(result, "abc123") {
		t.Error("expected commit hash in output when no PR number")
	}
}

func TestWriteNewContributorEntry_WithoutRemote(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:             true,
		ShowNewContributors: true,
	}
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, nil) // nil remote
	result := sb.String()

	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in output")
	}
	if !strings.Contains(result, "#42") {
		t.Error("expected PR number in output")
	}
}

func TestWriteNewContributorEntry_CustomFormat(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:               true,
		ShowNewContributors:   true,
		NewContributorsFormat: "* {{.Username}} joined in #{{.PRNumber}}",
	}
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, nil)
	result := sb.String()

	if !strings.Contains(result, "newdev joined in #42") {
		t.Errorf("expected custom format output, got: %s", result)
	}
}

func TestWriteNewContributorFallback_WithPR(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorFallback(&sb, &nc, remote)
	result := sb.String()

	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in fallback output")
	}
	if !strings.Contains(result, "#42") {
		t.Error("expected PR number in fallback output")
	}
}

func TestWriteNewContributorFallback_WithoutPR(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "",
	}

	var sb strings.Builder
	g.writeNewContributorFallback(&sb, &nc, &RemoteInfo{Host: "github.com"})
	result := sb.String()

	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in fallback output")
	}
	if !strings.Contains(result, "first contribution") {
		t.Error("expected 'first contribution' in fallback output")
	}
}

func TestWriteNewContributorFallback_WithoutRemote(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorFallback(&sb, &nc, nil)
	result := sb.String()

	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in fallback output")
	}
	if !strings.Contains(result, "#42") {
		t.Error("expected PR number in fallback output")
	}
}

func TestWriteNewContributorFallback_NoRemoteNoPR(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "",
	}

	var sb strings.Builder
	g.writeNewContributorFallback(&sb, &nc, nil)
	result := sb.String()

	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in fallback output")
	}
	if !strings.Contains(result, "first contribution") {
		t.Error("expected 'first contribution' in fallback output")
	}
}

func TestGetDefaultNewContributorFormat_WithRemote(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}

	format := g.getDefaultNewContributorFormat(remote)

	if !strings.Contains(format, "{{.Username}}") {
		t.Error("expected username placeholder in format")
	}
	if !strings.Contains(format, "{{.PRNumber}}") {
		t.Error("expected PR number placeholder in format")
	}
	if !strings.Contains(format, "owner/repo") {
		t.Error("expected owner/repo in format for PR links")
	}
}

func TestGetDefaultNewContributorFormat_WithoutRemote(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	format := g.getDefaultNewContributorFormat(nil)

	if !strings.Contains(format, "{{.Username}}") {
		t.Error("expected username placeholder in format")
	}
	if !strings.Contains(format, "{{.PRNumber}}") {
		t.Error("expected PR number placeholder in format")
	}
	// Should not contain full URL format
	if strings.Contains(format, "https://{{.Host}}") {
		t.Error("expected simpler format without full URLs when no remote")
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

func TestNewGenerator_InvalidFormat(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Format = "invalid-format"

	_, err := NewGenerator(cfg)
	if err == nil {
		t.Error("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "unknown changelog format") {
		t.Errorf("error = %v, expected to contain 'unknown changelog format'", err)
	}
}

func TestResolveRemote_FillProviderFromHost(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Repository = &RepositoryConfig{
		Host:  "github.com",
		Owner: "owner",
		Repo:  "repo",
		// Provider not set - should be filled from host
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remote, err := g.resolveRemote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Provider should be filled from host
	if remote.Provider != "github" {
		t.Errorf("Provider = %q, want 'github'", remote.Provider)
	}
}

func TestWriteContributorEntry_TemplateExecutionError(t *testing.T) {
	cfg := DefaultConfig()
	// Invalid template that parses but fails on execution
	cfg.Contributors = &ContributorsConfig{
		Enabled: true,
		Format:  "- {{.NonExistentMethod}}",
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contrib := Contributor{
		Name:     "Test User",
		Username: "testuser",
		Host:     "github.com",
	}

	var sb strings.Builder
	g.writeContributorEntry(&sb, contrib, &RemoteInfo{Host: "github.com"})
	result := sb.String()

	// Should fallback to default format
	if !strings.Contains(result, "@testuser") {
		t.Error("expected fallback format with username")
	}
}

func TestWriteNewContributorEntry_TemplateParseError(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:               true,
		ShowNewContributors:   true,
		NewContributorsFormat: "- {{.Invalid", // Invalid template syntax
	}
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, &RemoteInfo{Host: "github.com", Owner: "owner", Repo: "repo"})
	result := sb.String()

	// Should fallback
	if !strings.Contains(result, "@newdev") {
		t.Error("expected fallback format with username")
	}
}

func TestWriteNewContributorEntry_TemplateExecutionError(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:               true,
		ShowNewContributors:   true,
		NewContributorsFormat: "- {{.NonExistent.Field}}", // Will fail on execution
	}
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, &RemoteInfo{Host: "github.com", Owner: "owner", Repo: "repo"})
	result := sb.String()

	// Should fallback
	if !strings.Contains(result, "@newdev") {
		t.Error("expected fallback format with username")
	}
}

func TestCollectVersionFiles(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("failed to create changes dir: %v", err)
	}

	// Create version files and other files
	files := map[string]string{
		"v1.0.0.md": "version 1",
		"v0.1.0.md": "version 0.1",
		"README.md": "not a version",
		"other.txt": "not markdown",
		"notes.md":  "not starting with v",
	}
	for name, content := range files {
		path := filepath.Join(changesDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	// Create a subdirectory (should be skipped)
	subdir := filepath.Join(changesDir, "v2.0.0")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	collected, err := collectVersionFiles(changesDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have v1.0.0.md and v0.1.0.md
	if len(collected) != 2 {
		t.Errorf("expected 2 version files, got %d", len(collected))
	}

	// Check files are the right ones
	hasV1 := false
	hasV01 := false
	for _, f := range collected {
		if strings.Contains(f, "v1.0.0.md") {
			hasV1 = true
		}
		if strings.Contains(f, "v0.1.0.md") {
			hasV01 = true
		}
	}
	if !hasV1 || !hasV01 {
		t.Error("expected both v1.0.0.md and v0.1.0.md in collected files")
	}
}

func TestBuildMergedContent(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("failed to create changes dir: %v", err)
	}

	// Create version files
	v1Content := "## v1.0.0\n\nFirst release"
	v2Content := "## v2.0.0\n\nSecond release"
	if err := os.WriteFile(filepath.Join(changesDir, "v1.0.0.md"), []byte(v1Content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changesDir, "v2.0.0.md"), []byte(v2Content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	cfg := DefaultConfig()
	cfg.ChangesDir = changesDir
	g, _ := NewGenerator(cfg)

	files := []string{
		filepath.Join(changesDir, "v2.0.0.md"),
		filepath.Join(changesDir, "v1.0.0.md"),
	}

	content := g.buildMergedContent(files)

	if !strings.Contains(content, "# Changelog") {
		t.Error("expected header in merged content")
	}
	if !strings.Contains(content, "v1.0.0") {
		t.Error("expected v1.0.0 in merged content")
	}
	if !strings.Contains(content, "v2.0.0") {
		t.Error("expected v2.0.0 in merged content")
	}
}
