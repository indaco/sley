package changeloggenerator

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var fakeGitCommands = map[string]string{}

// s is a shorthand alias for fieldSep to keep test fixtures readable.
var s = fieldSep

// commitLogFormat is the git log --pretty=format string used by getCommitsWithMeta.
var commitLogFormat = "git log --pretty=format:%H" + s + "%h" + s + "%s" + s + "%an" + s + "%ae"

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cmdStr := command + " " + strings.Join(args, " ")
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", cmdStr) //nolint:gosec // standard test re-exec pattern

	cmd.Env = append(os.Environ(),
		"GO_TEST_HELPER_PROCESS=1",
		"MOCK_KEY="+cmdStr,
		"MOCK_VAL="+fakeGitCommands[cmdStr],
	)

	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_HELPER_PROCESS") != "1" {
		return
	}

	val := os.Getenv("MOCK_VAL")

	if val == "ERROR" {
		_, _ = os.Stderr.WriteString("mock failure")
		os.Exit(1)
	}

	_, _ = os.Stdout.WriteString(val)
	os.Exit(0)
}

func TestGetCommitsWithMeta(t *testing.T) {
	tests := []struct {
		name            string
		since           string
		until           string
		mockGitCommands map[string]string
		expectedCount   int
		expectErr       bool
	}{
		{
			name:  "with explicit since and until",
			since: "v1.0.0",
			until: "HEAD",
			mockGitCommands: map[string]string{
				commitLogFormat + " v1.0.0..HEAD": "abc123" + s + "abc123" + s + "feat: login" + s + "Alice" + s + "alice@example.com\ndef456" + s + "def456" + s + "fix: bug" + s + "Bob" + s + "bob@example.com",
			},
			expectedCount: 2,
		},
		{
			name:  "fallback to HEAD~10 when no tag and enough commits",
			since: "",
			until: "HEAD",
			mockGitCommands: map[string]string{
				"git describe --tags --abbrev=0":   "", // no tags
				"git rev-list --count HEAD":        "25",
				commitLogFormat + " HEAD~10..HEAD": "abc123" + s + "abc123" + s + "feat: update" + s + "Alice" + s + "alice@example.com",
			},
			expectedCount: 1,
		},
		{
			name:  "fallback to root commit when fewer than 10 commits",
			since: "",
			until: "HEAD",
			mockGitCommands: map[string]string{
				"git describe --tags --abbrev=0":    "", // no tags
				"git rev-list --count HEAD":         "2",
				"git rev-list --max-parents=0 HEAD": "root123",
				commitLogFormat + " root123..HEAD":  "abc123" + s + "abc123" + s + "feat: init" + s + "Alice" + s + "alice@example.com",
			},
			expectedCount: 1,
		},
		{
			name:  "fallback to last tag when tag exists",
			since: "",
			until: "HEAD",
			mockGitCommands: map[string]string{
				"git describe --tags --abbrev=0":  "v2.0.0",
				commitLogFormat + " v2.0.0..HEAD": "abc123" + s + "abc123" + s + "feat: new" + s + "Alice" + s + "alice@example.com",
			},
			expectedCount: 1,
		},
		{
			name:  "git log returns error",
			since: "v1.0.0",
			until: "HEAD",
			mockGitCommands: map[string]string{
				commitLogFormat + " v1.0.0..HEAD": "ERROR",
			},
			expectErr: true,
		},
		{
			name:  "empty commit log",
			since: "v1.0.0",
			until: "HEAD",
			mockGitCommands: map[string]string{
				commitLogFormat + " v1.0.0..HEAD": "",
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeGitCommands = tt.mockGitCommands

			g := &GitOps{ExecCommandFn: fakeExecCommand}
			commits, err := g.getCommitsWithMeta(tt.since, tt.until)

			if (err != nil) != tt.expectErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(commits) != tt.expectedCount {
				t.Fatalf("expected %d commits, got %d", tt.expectedCount, len(commits))
			}
		})
	}
}

func TestGetCommitsWithMeta_PipeInSubject(t *testing.T) {

	tests := []struct {
		name        string
		mockOutput  string
		wantSubject string
		wantAuthor  string
		wantEmail   string
	}{
		{
			name:        "pipe in subject is preserved",
			mockOutput:  "abc123" + s + "abc1" + s + "feat: add A | B support" + s + "Alice" + s + "alice@example.com",
			wantSubject: "feat: add A | B support",
			wantAuthor:  "Alice",
			wantEmail:   "alice@example.com",
		},
		{
			name:        "multiple pipes in subject",
			mockOutput:  "def456" + s + "def4" + s + "fix: handle X | Y | Z" + s + "Bob" + s + "bob@example.com",
			wantSubject: "fix: handle X | Y | Z",
			wantAuthor:  "Bob",
			wantEmail:   "bob@example.com",
		},
		{
			name:        "no pipe in subject",
			mockOutput:  "ghi789" + s + "ghi7" + s + "feat: normal change" + s + "Charlie" + s + "charlie@example.com",
			wantSubject: "feat: normal change",
			wantAuthor:  "Charlie",
			wantEmail:   "charlie@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeGitCommands = map[string]string{
				commitLogFormat + " v1.0.0..HEAD": tt.mockOutput,
			}

			g := &GitOps{ExecCommandFn: fakeExecCommand}
			commits, err := g.getCommitsWithMeta("v1.0.0", "HEAD")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(commits) != 1 {
				t.Fatalf("expected 1 commit, got %d", len(commits))
			}

			c := commits[0]
			if c.Subject != tt.wantSubject {
				t.Errorf("Subject = %q, want %q", c.Subject, tt.wantSubject)
			}
			if c.Author != tt.wantAuthor {
				t.Errorf("Author = %q, want %q", c.Author, tt.wantAuthor)
			}
			if c.AuthorEmail != tt.wantEmail {
				t.Errorf("AuthorEmail = %q, want %q", c.AuthorEmail, tt.wantEmail)
			}
		})
	}
}

func TestParseRemoteURL(t *testing.T) {

	tests := []struct {
		name         string
		url          string
		wantProvider string
		wantHost     string
		wantOwner    string
		wantRepo     string
		wantErr      bool
	}{
		// GitHub
		{
			name:         "GitHub SSH format",
			url:          "git@github.com:indaco/sley.git",
			wantProvider: "github",
			wantHost:     "github.com",
			wantOwner:    "indaco",
			wantRepo:     "sley",
		},
		{
			name:         "GitHub SSH format without .git",
			url:          "git@github.com:owner/repo",
			wantProvider: "github",
			wantHost:     "github.com",
			wantOwner:    "owner",
			wantRepo:     "repo",
		},
		{
			name:         "GitHub HTTPS format",
			url:          "https://github.com/indaco/sley.git",
			wantProvider: "github",
			wantHost:     "github.com",
			wantOwner:    "indaco",
			wantRepo:     "sley",
		},
		{
			name:         "GitHub HTTPS format without .git",
			url:          "https://github.com/owner/repo",
			wantProvider: "github",
			wantHost:     "github.com",
			wantOwner:    "owner",
			wantRepo:     "repo",
		},
		{
			name:         "GitHub Git protocol",
			url:          "git://github.com/owner/repo.git",
			wantProvider: "github",
			wantHost:     "github.com",
			wantOwner:    "owner",
			wantRepo:     "repo",
		},
		// GitLab
		{
			name:         "GitLab SSH format",
			url:          "git@gitlab.com:mygroup/myproject.git",
			wantProvider: "gitlab",
			wantHost:     "gitlab.com",
			wantOwner:    "mygroup",
			wantRepo:     "myproject",
		},
		{
			name:         "GitLab HTTPS format",
			url:          "https://gitlab.com/mygroup/myproject.git",
			wantProvider: "gitlab",
			wantHost:     "gitlab.com",
			wantOwner:    "mygroup",
			wantRepo:     "myproject",
		},
		// Codeberg
		{
			name:         "Codeberg SSH format",
			url:          "git@codeberg.org:user/project.git",
			wantProvider: "codeberg",
			wantHost:     "codeberg.org",
			wantOwner:    "user",
			wantRepo:     "project",
		},
		{
			name:         "Codeberg HTTPS format",
			url:          "https://codeberg.org/user/project",
			wantProvider: "codeberg",
			wantHost:     "codeberg.org",
			wantOwner:    "user",
			wantRepo:     "project",
		},
		// Bitbucket
		{
			name:         "Bitbucket SSH format",
			url:          "git@bitbucket.org:team/repo.git",
			wantProvider: "bitbucket",
			wantHost:     "bitbucket.org",
			wantOwner:    "team",
			wantRepo:     "repo",
		},
		{
			name:         "Bitbucket HTTPS format",
			url:          "https://bitbucket.org/team/repo.git",
			wantProvider: "bitbucket",
			wantHost:     "bitbucket.org",
			wantOwner:    "team",
			wantRepo:     "repo",
		},
		// Custom/self-hosted
		{
			name:         "Self-hosted GitLab SSH",
			url:          "git@git.company.com:team/project.git",
			wantProvider: "custom",
			wantHost:     "git.company.com",
			wantOwner:    "team",
			wantRepo:     "project",
		},
		{
			name:         "Self-hosted Gitea HTTPS",
			url:          "https://gitea.myserver.io/user/repo",
			wantProvider: "custom",
			wantHost:     "gitea.myserver.io",
			wantOwner:    "user",
			wantRepo:     "repo",
		},
		// Error cases
		{
			name:    "Invalid URL",
			url:     "not-a-valid-url",
			wantErr: true,
		},
		{
			name:    "Local path",
			url:     "/path/to/repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := parseRemoteURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got.Provider != tt.wantProvider {
				t.Errorf("Provider = %q, want %q", got.Provider, tt.wantProvider)
			}
			if got.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", got.Host, tt.wantHost)
			}
			if got.Owner != tt.wantOwner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.wantOwner)
			}
			if got.Repo != tt.wantRepo {
				t.Errorf("Repo = %q, want %q", got.Repo, tt.wantRepo)
			}
		})
	}
}

func TestExtractUsername(t *testing.T) {

	tests := []struct {
		name       string
		email      string
		authorName string
		wantUser   string
		wantHost   string
	}{
		{
			name:       "GitHub noreply with ID",
			email:      "12345+testuser@users.noreply.github.com",
			authorName: "Test User",
			wantUser:   "testuser",
			wantHost:   "github.com",
		},
		{
			name:       "GitHub noreply without ID",
			email:      "testuser@users.noreply.github.com",
			authorName: "Test User",
			wantUser:   "testuser",
			wantHost:   "github.com",
		},
		{
			name:       "GitLab noreply",
			email:      "testuser@noreply.gitlab.com",
			authorName: "Test User",
			wantUser:   "testuser",
			wantHost:   "gitlab.com",
		},
		{
			name:       "Codeberg noreply",
			email:      "myuser@noreply.codeberg.org",
			authorName: "My User",
			wantUser:   "myuser",
			wantHost:   "codeberg.org",
		},
		{
			name:       "Regular email - fallback to author name",
			email:      "test@example.com",
			authorName: "Test User",
			wantUser:   "testuser",
			wantHost:   "",
		},
		{
			name:       "Single name author",
			email:      "user@example.com",
			authorName: "Developer",
			wantUser:   "developer",
			wantHost:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotUser, gotHost := extractUsername(tt.email, tt.authorName)
			if gotUser != tt.wantUser {
				t.Errorf("username = %q, want %q", gotUser, tt.wantUser)
			}
			if gotHost != tt.wantHost {
				t.Errorf("host = %q, want %q", gotHost, tt.wantHost)
			}
		})
	}
}

func TestGetContributors(t *testing.T) {

	commits := []CommitInfo{
		{Author: "Alice", AuthorEmail: "alice@example.com"},
		{Author: "Bob", AuthorEmail: "bob@example.com"},
		{Author: "Alice", AuthorEmail: "alice@example.com"}, // Duplicate
		{Author: "Charlie", AuthorEmail: "charlie@users.noreply.github.com"},
	}

	contributors := getContributors(commits)

	if len(contributors) != 3 {
		t.Fatalf("expected 3 unique contributors, got %d", len(contributors))
	}

	// Verify contributor names
	names := make(map[string]bool)
	for _, c := range contributors {
		names[c.Name] = true
	}

	if !names["Alice"] || !names["Bob"] || !names["Charlie"] {
		t.Error("expected Alice, Bob, and Charlie in contributors")
	}

	// Verify Charlie has GitHub host detected
	for _, c := range contributors {
		if c.Name == "Charlie" {
			if c.Host != "github.com" {
				t.Errorf("Charlie's host = %q, want 'github.com'", c.Host)
			}
			if c.Username != "charlie" {
				t.Errorf("Charlie's username = %q, want 'charlie'", c.Username)
			}
		}
	}
}

func TestCommitInfo(t *testing.T) {

	commit := CommitInfo{
		Hash:        "abc123def456",
		ShortHash:   "abc123d",
		Subject:     "feat: add feature",
		Author:      "Test Author",
		AuthorEmail: "test@example.com",
	}

	if commit.Hash != "abc123def456" {
		t.Errorf("Hash = %q, want 'abc123def456'", commit.Hash)
	}
	if commit.ShortHash != "abc123d" {
		t.Errorf("ShortHash = %q, want 'abc123d'", commit.ShortHash)
	}
	if commit.Subject != "feat: add feature" {
		t.Errorf("Subject = %q, want 'feat: add feature'", commit.Subject)
	}
}

func TestBuildRemoteInfo(t *testing.T) {

	tests := []struct {
		name         string
		host         string
		owner        string
		repo         string
		wantProvider string
	}{
		{
			name:         "GitHub",
			host:         "github.com",
			owner:        "owner",
			repo:         "repo",
			wantProvider: "github",
		},
		{
			name:         "GitLab",
			host:         "gitlab.com",
			owner:        "group",
			repo:         "project",
			wantProvider: "gitlab",
		},
		{
			name:         "Codeberg",
			host:         "codeberg.org",
			owner:        "user",
			repo:         "repo",
			wantProvider: "codeberg",
		},
		{
			name:         "Bitbucket",
			host:         "bitbucket.org",
			owner:        "team",
			repo:         "repo",
			wantProvider: "bitbucket",
		},
		{
			name:         "Custom host",
			host:         "git.mycompany.com",
			owner:        "team",
			repo:         "project",
			wantProvider: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := buildRemoteInfo(tt.host, tt.owner, tt.repo)

			if got.Provider != tt.wantProvider {
				t.Errorf("Provider = %q, want %q", got.Provider, tt.wantProvider)
			}
			if got.Host != tt.host {
				t.Errorf("Host = %q, want %q", got.Host, tt.host)
			}
			if got.Owner != tt.owner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.owner)
			}
			if got.Repo != tt.repo {
				t.Errorf("Repo = %q, want %q", got.Repo, tt.repo)
			}
		})
	}
}

func TestKnownProviders(t *testing.T) {

	expected := map[string]string{
		"github.com":    "github",
		"gitlab.com":    "gitlab",
		"codeberg.org":  "codeberg",
		"gitea.io":      "gitea",
		"bitbucket.org": "bitbucket",
		"sr.ht":         "sourcehut",
	}

	for host, provider := range expected {
		if got := KnownProviders[host]; got != provider {
			t.Errorf("KnownProviders[%q] = %q, want %q", host, got, provider)
		}
	}
}

func TestGetCommitsWithMeta_MockSuccess(t *testing.T) {

	gitOps := NewGitOps()
	gitOps.GetCommitsWithMetaFn = func(since, until string) ([]CommitInfo, error) {
		return []CommitInfo{
			{Hash: "abc123", ShortHash: "abc123", Subject: "feat: test", Author: "Test", AuthorEmail: "test@example.com"},
			{Hash: "def456", ShortHash: "def456", Subject: "fix: bug", Author: "User", AuthorEmail: "user@example.com"},
		}, nil
	}

	commits, err := gitOps.GetCommitsWithMetaFn("v1.0.0", "HEAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 2 {
		t.Errorf("expected 2 commits, got %d", len(commits))
	}
}

func TestGetRemoteInfo_MockSuccess(t *testing.T) {

	gitOps := NewGitOps()
	gitOps.GetRemoteInfoFn = func() (*RemoteInfo, error) {
		return &RemoteInfo{
			Provider: "github",
			Host:     "github.com",
			Owner:    "testowner",
			Repo:     "testrepo",
		}, nil
	}

	remote, err := gitOps.GetRemoteInfoFn()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remote.Owner != "testowner" {
		t.Errorf("Owner = %q, want 'testowner'", remote.Owner)
	}
}

func TestGetLatestTag_MockSuccess(t *testing.T) {

	gitOps := NewGitOps()
	gitOps.GetLatestTagFn = func() (string, error) {
		return "v1.0.0", nil
	}

	tag, err := gitOps.GetLatestTagFn()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v1.0.0" {
		t.Errorf("tag = %q, want 'v1.0.0'", tag)
	}
}

func TestGetContributors_MockSuccess(t *testing.T) {

	gitOps := NewGitOps()
	gitOps.GetContributorsFn = func(commits []CommitInfo) []Contributor {
		return []Contributor{
			{Name: "Test User", Username: "testuser", Email: "test@example.com", Host: "github.com"},
		}
	}

	commits := []CommitInfo{{Author: "Test", AuthorEmail: "test@example.com"}}
	contributors := gitOps.GetContributorsFn(commits)
	if len(contributors) != 1 {
		t.Errorf("expected 1 contributor, got %d", len(contributors))
	}
}

func TestRemoteInfo_Fields(t *testing.T) {

	remote := RemoteInfo{
		Provider: "gitlab",
		Host:     "gitlab.example.com",
		Owner:    "mygroup",
		Repo:     "myproject",
	}

	if remote.Provider != "gitlab" {
		t.Errorf("Provider = %q, want 'gitlab'", remote.Provider)
	}
	if remote.Host != "gitlab.example.com" {
		t.Errorf("Host = %q, want 'gitlab.example.com'", remote.Host)
	}
	if remote.Owner != "mygroup" {
		t.Errorf("Owner = %q, want 'mygroup'", remote.Owner)
	}
	if remote.Repo != "myproject" {
		t.Errorf("Repo = %q, want 'myproject'", remote.Repo)
	}
}

func TestContributor_Fields(t *testing.T) {

	contrib := Contributor{
		Name:     "Alice Smith",
		Username: "alicesmith",
		Email:    "alice@example.com",
		Host:     "github.com",
	}

	if contrib.Name != "Alice Smith" {
		t.Errorf("Name = %q, want 'Alice Smith'", contrib.Name)
	}
	if contrib.Username != "alicesmith" {
		t.Errorf("Username = %q, want 'alicesmith'", contrib.Username)
	}
	if contrib.Email != "alice@example.com" {
		t.Errorf("Email = %q, want 'alice@example.com'", contrib.Email)
	}
	if contrib.Host != "github.com" {
		t.Errorf("Host = %q, want 'github.com'", contrib.Host)
	}
}

func TestNewContributor_Fields(t *testing.T) {

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Email:    "newdev@users.noreply.github.com",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{
			Hash:      "abc123",
			ShortHash: "abc123",
			Subject:   "feat: first feature (#42)",
		},
		PRNumber: "42",
	}

	if nc.Name != "New Dev" {
		t.Errorf("Name = %q, want 'New Dev'", nc.Name)
	}
	if nc.Username != "newdev" {
		t.Errorf("Username = %q, want 'newdev'", nc.Username)
	}
	if nc.PRNumber != "42" {
		t.Errorf("PRNumber = %q, want '42'", nc.PRNumber)
	}
	if nc.FirstCommit.ShortHash != "abc123" {
		t.Errorf("FirstCommit.ShortHash = %q, want 'abc123'", nc.FirstCommit.ShortHash)
	}
}

func TestGetNewContributors(t *testing.T) {

	tests := []struct {
		name                string
		commits             []CommitInfo
		historicalUsernames map[string]struct{}
		previousVersion     string
		wantCount           int
		wantUsernames       []string
	}{
		{
			name: "all new contributors (first release)",
			commits: []CommitInfo{
				{Author: "Alice", AuthorEmail: "alice@users.noreply.github.com", ShortHash: "abc123", Subject: "feat: initial (#1)"},
				{Author: "Bob", AuthorEmail: "bob@users.noreply.github.com", ShortHash: "def456", Subject: "docs: readme"},
			},
			historicalUsernames: map[string]struct{}{},
			previousVersion:     "",
			wantCount:           2,
			wantUsernames:       []string{"alice", "bob"},
		},
		{
			name: "mix of new and existing contributors",
			commits: []CommitInfo{
				{Author: "Alice", AuthorEmail: "alice@users.noreply.github.com", ShortHash: "abc123", Subject: "feat: new (#5)"},
				{Author: "Charlie", AuthorEmail: "charlie@users.noreply.github.com", ShortHash: "def456", Subject: "fix: bug (#6)"},
			},
			historicalUsernames: map[string]struct{}{
				"alice": {},
			},
			previousVersion: "v1.0.0",
			wantCount:       1,
			wantUsernames:   []string{"charlie"},
		},
		{
			name: "no new contributors",
			commits: []CommitInfo{
				{Author: "Alice", AuthorEmail: "alice@users.noreply.github.com", ShortHash: "abc123", Subject: "feat: update"},
			},
			historicalUsernames: map[string]struct{}{
				"alice": {},
			},
			previousVersion: "v1.0.0",
			wantCount:       0,
			wantUsernames:   []string{},
		},
		{
			name: "deduplicates same contributor multiple commits",
			commits: []CommitInfo{
				{Author: "NewUser", AuthorEmail: "newuser@users.noreply.github.com", ShortHash: "abc123", Subject: "feat: first (#10)"},
				{Author: "NewUser", AuthorEmail: "newuser@users.noreply.github.com", ShortHash: "def456", Subject: "feat: second (#11)"},
			},
			historicalUsernames: map[string]struct{}{},
			previousVersion:     "v1.0.0",
			wantCount:           1,
			wantUsernames:       []string{"newuser"},
		},
		{
			name: "extracts PR number from commit subject",
			commits: []CommitInfo{
				{Author: "Dev", AuthorEmail: "dev@users.noreply.github.com", ShortHash: "abc123", Subject: "feat: add feature (#123)"},
			},
			historicalUsernames: map[string]struct{}{},
			previousVersion:     "v1.0.0",
			wantCount:           1,
			wantUsernames:       []string{"dev"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gitOps := NewGitOps()
			gitOps.GetHistoricalContributorsFn = func(ref string) (map[string]struct{}, error) {
				return tt.historicalUsernames, nil
			}

			got, err := gitOps.getNewContributors(tt.commits, tt.previousVersion)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != tt.wantCount {
				t.Errorf("got %d new contributors, want %d", len(got), tt.wantCount)
			}

			// Verify usernames
			gotUsernames := make(map[string]bool)
			for _, nc := range got {
				gotUsernames[nc.Username] = true
			}

			for _, wantUsername := range tt.wantUsernames {
				if !gotUsernames[wantUsername] {
					t.Errorf("expected username %q not found in new contributors", wantUsername)
				}
			}
		})
	}
}

func TestGetNewContributors_EmptyPreviousVersionResolvesTag(t *testing.T) {
	// Regression test: when previousVersion is "", getNewContributors must
	// resolve the latest tag and use it for the historical contributor lookup.
	// Without this, getHistoricalContributors receives "" and returns an empty
	// set, making every author appear as a new contributor.

	gitOps := NewGitOps()

	// Simulate a repo where "alice" has commits before v1.0.0 (historical).
	gitOps.GetLatestTagFn = func() (string, error) {
		return "v1.0.0", nil
	}
	gitOps.GetHistoricalContributorsFn = func(beforeRef string) (map[string]struct{}, error) {
		if beforeRef == "" {
			// This should NOT be reached after the fix.
			return map[string]struct{}{}, nil
		}
		// Return alice as a known historical contributor.
		return map[string]struct{}{"alice": {}}, nil
	}

	commits := []CommitInfo{
		{Author: "Alice", AuthorEmail: "alice@users.noreply.github.com", ShortHash: "aaa111", Subject: "fix: update (#10)"},
		{Author: "NewDev", AuthorEmail: "newdev@users.noreply.github.com", ShortHash: "bbb222", Subject: "feat: feature (#11)"},
	}

	got, err := gitOps.getNewContributors(commits, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 new contributor, got %d", len(got))
	}
	if got[0].Username != "newdev" {
		t.Errorf("expected new contributor 'newdev', got %q", got[0].Username)
	}
}

func TestGetNewContributors_EmptyPreviousVersionNoTags(t *testing.T) {
	// When previousVersion is "" and no tags exist, all contributors are
	// genuinely new (first release). This must not error.

	gitOps := NewGitOps()
	gitOps.GetLatestTagFn = func() (string, error) {
		return "", fmt.Errorf("no tags found")
	}
	gitOps.GetHistoricalContributorsFn = func(beforeRef string) (map[string]struct{}, error) {
		return map[string]struct{}{}, nil
	}

	commits := []CommitInfo{
		{Author: "Alice", AuthorEmail: "alice@users.noreply.github.com", ShortHash: "aaa111", Subject: "feat: init (#1)"},
	}

	got, err := gitOps.getNewContributors(commits, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 new contributor, got %d", len(got))
	}
	if got[0].Username != "alice" {
		t.Errorf("expected new contributor 'alice', got %q", got[0].Username)
	}
}

func TestGetNewContributors_PRNumberExtraction(t *testing.T) {

	gitOps := NewGitOps()
	gitOps.GetHistoricalContributorsFn = func(ref string) (map[string]struct{}, error) {
		return map[string]struct{}{}, nil
	}

	tests := []struct {
		name         string
		subject      string
		wantPRNumber string
	}{
		{
			name:         "PR number at end",
			subject:      "feat: add feature (#123)",
			wantPRNumber: "123",
		},
		{
			name:         "PR number in middle",
			subject:      "Merge pull request #456 from branch",
			wantPRNumber: "456",
		},
		{
			name:         "no PR number",
			subject:      "feat: add feature without PR",
			wantPRNumber: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			commits := []CommitInfo{
				{Author: "Dev", AuthorEmail: "dev@users.noreply.github.com", ShortHash: "abc123", Subject: tt.subject},
			}

			got, err := gitOps.getNewContributors(commits, "v1.0.0")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 new contributor, got %d", len(got))
			}

			if got[0].PRNumber != tt.wantPRNumber {
				t.Errorf("PRNumber = %q, want %q", got[0].PRNumber, tt.wantPRNumber)
			}
		})
	}
}

func TestGetHistoricalContributors_MockSuccess(t *testing.T) {

	gitOps := NewGitOps()
	gitOps.GetHistoricalContributorsFn = func(beforeRef string) (map[string]struct{}, error) {
		return map[string]struct{}{
			"alice":   {},
			"bob":     {},
			"charlie": {},
		}, nil
	}

	usernames, err := gitOps.GetHistoricalContributorsFn("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(usernames) != 3 {
		t.Errorf("expected 3 historical contributors, got %d", len(usernames))
	}
	if _, ok := usernames["alice"]; !ok {
		t.Error("expected alice in historical contributors")
	}
}

func TestGetNewContributorsFn_MockSuccess(t *testing.T) {

	gitOps := NewGitOps()
	gitOps.GetNewContributorsFn = func(commits []CommitInfo, previousVersion string) ([]NewContributor, error) {
		return []NewContributor{
			{
				Contributor: Contributor{
					Name:     "New Dev",
					Username: "newdev",
					Host:     "github.com",
				},
				PRNumber: "42",
			},
		}, nil
	}

	commits := []CommitInfo{{Author: "New Dev", AuthorEmail: "newdev@users.noreply.github.com"}}
	newContribs, err := gitOps.GetNewContributorsFn(commits, "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(newContribs) != 1 {
		t.Errorf("expected 1 new contributor, got %d", len(newContribs))
	}
	if newContribs[0].Username != "newdev" {
		t.Errorf("Username = %q, want 'newdev'", newContribs[0].Username)
	}
}
