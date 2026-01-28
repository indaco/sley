package extensionmgr

import (
	"testing"
)

/* ------------------------------------------------------------------------- */
/* TABLE-DRIVEN TESTS FOR REPO URL METHODS                                  */
/* ------------------------------------------------------------------------- */

func TestRepoURL_Methods(t *testing.T) {
	tests := []struct {
		name         string
		repoURL      *RepoURL
		wantIsGitHub bool
		wantIsGitLab bool
		wantCloneURL string
		wantString   string
	}{
		{
			name: "GitHub repository",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "user",
				Repo:  "awesome-ext",
				Raw:   "https://github.com/user/awesome-ext",
			},
			wantIsGitHub: true,
			wantIsGitLab: false,
			wantCloneURL: "https://github.com/user/awesome-ext.git",
			wantString:   "user/awesome-ext",
		},
		{
			name: "GitLab repository",
			repoURL: &RepoURL{
				Host:  "gitlab.com",
				Owner: "organization",
				Repo:  "cool-project",
				Raw:   "https://gitlab.com/organization/cool-project",
			},
			wantIsGitHub: false,
			wantIsGitLab: true,
			wantCloneURL: "https://gitlab.com/organization/cool-project.git",
			wantString:   "organization/cool-project",
		},
		{
			name: "other git hosting",
			repoURL: &RepoURL{
				Host:  "git.example.com",
				Owner: "team",
				Repo:  "extension",
				Raw:   "https://git.example.com/team/extension",
			},
			wantIsGitHub: false,
			wantIsGitLab: false,
			wantCloneURL: "https://git.example.com/team/extension.git",
			wantString:   "team/extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.repoURL.IsGitHubURL(); got != tt.wantIsGitHub {
				t.Errorf("IsGitHubURL() = %v, want %v", got, tt.wantIsGitHub)
			}
			if got := tt.repoURL.IsGitLabURL(); got != tt.wantIsGitLab {
				t.Errorf("IsGitLabURL() = %v, want %v", got, tt.wantIsGitLab)
			}
			if got := tt.repoURL.CloneURL(); got != tt.wantCloneURL {
				t.Errorf("CloneURL() = %v, want %v", got, tt.wantCloneURL)
			}
			if got := tt.repoURL.String(); got != tt.wantString {
				t.Errorf("String() = %v, want %v", got, tt.wantString)
			}
		})
	}
}

// TestRepoURL_CloneURL tests clone URL generation
func TestRepoURL_CloneURL(t *testing.T) {
	tests := []struct {
		name    string
		repoURL *RepoURL
		wantURL string
	}{
		{
			name: "GitHub standard",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "user",
				Repo:  "repo",
			},
			wantURL: "https://github.com/user/repo.git",
		},
		{
			name: "GitLab standard",
			repoURL: &RepoURL{
				Host:  "gitlab.com",
				Owner: "org",
				Repo:  "project",
			},
			wantURL: "https://gitlab.com/org/project.git",
		},
		{
			name: "Custom git host",
			repoURL: &RepoURL{
				Host:  "git.company.com",
				Owner: "team",
				Repo:  "extension",
			},
			wantURL: "https://git.company.com/team/extension.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.repoURL.CloneURL()
			if got != tt.wantURL {
				t.Errorf("CloneURL() = %v, want %v", got, tt.wantURL)
			}
		})
	}
}

// TestRepoURL_String tests string representation
func TestRepoURL_String(t *testing.T) {
	tests := []struct {
		name    string
		repoURL *RepoURL
		want    string
	}{
		{
			name: "basic repo",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "user",
				Repo:  "myrepo",
			},
			want: "user/myrepo",
		},
		{
			name: "org repo",
			repoURL: &RepoURL{
				Host:  "gitlab.com",
				Owner: "my-org",
				Repo:  "awesome-project",
			},
			want: "my-org/awesome-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.repoURL.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRepoURL_HostChecks tests host checking methods
func TestRepoURL_HostChecks(t *testing.T) {
	tests := []struct {
		name       string
		repoURL    *RepoURL
		wantGitHub bool
		wantGitLab bool
	}{
		{
			name: "GitHub.com",
			repoURL: &RepoURL{
				Host: "github.com",
			},
			wantGitHub: true,
			wantGitLab: false,
		},
		{
			name: "GitLab.com",
			repoURL: &RepoURL{
				Host: "gitlab.com",
			},
			wantGitHub: false,
			wantGitLab: true,
		},
		{
			name: "GitHub Enterprise (not github.com)",
			repoURL: &RepoURL{
				Host: "github.enterprise.com",
			},
			wantGitHub: false,
			wantGitLab: false,
		},
		{
			name: "Self-hosted GitLab",
			repoURL: &RepoURL{
				Host: "gitlab.company.com",
			},
			wantGitHub: false,
			wantGitLab: false,
		},
		{
			name: "Bitbucket",
			repoURL: &RepoURL{
				Host: "bitbucket.org",
			},
			wantGitHub: false,
			wantGitLab: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.repoURL.IsGitHubURL(); got != tt.wantGitHub {
				t.Errorf("IsGitHubURL() = %v, want %v", got, tt.wantGitHub)
			}
			if got := tt.repoURL.IsGitLabURL(); got != tt.wantGitLab {
				t.Errorf("IsGitLabURL() = %v, want %v", got, tt.wantGitLab)
			}
		})
	}
}

// TestRepoURL_String_WithRef tests the String() method includes ref when present
func TestRepoURL_String_WithRef(t *testing.T) {
	tests := []struct {
		name    string
		repoURL *RepoURL
		want    string
	}{
		{
			name: "without ref",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "user",
				Repo:  "repo",
			},
			want: "user/repo",
		},
		{
			name: "with version tag",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "user",
				Repo:  "repo",
				Ref:   "v1.0.0",
			},
			want: "user/repo@v1.0.0",
		},
		{
			name: "with branch",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "user",
				Repo:  "repo",
				Ref:   "develop",
			},
			want: "user/repo@develop",
		},
		{
			name: "with commit hash",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "user",
				Repo:  "repo",
				Ref:   "abc123",
			},
			want: "user/repo@abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.repoURL.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
