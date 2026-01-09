package changeloggenerator

import (
	"strings"
	"testing"
)

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
