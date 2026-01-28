package extensionmgr

import (
	"testing"
)

/* ------------------------------------------------------------------------- */
/* TABLE-DRIVEN TESTS FOR URL PARSING                                       */
/* ------------------------------------------------------------------------- */

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name        string
		urlStr      string
		wantHost    string
		wantOwner   string
		wantRepo    string
		wantErr     bool
		wantErrText string
	}{
		{
			name:      "full GitHub HTTPS URL",
			urlStr:    "https://github.com/user/repo",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "full GitLab HTTPS URL",
			urlStr:    "https://gitlab.com/organization/project",
			wantHost:  "gitlab.com",
			wantOwner: "organization",
			wantRepo:  "project",
			wantErr:   false,
		},
		{
			name:      "GitHub URL without protocol",
			urlStr:    "github.com/user/extension-repo",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "extension-repo",
			wantErr:   false,
		},
		{
			name:      "GitLab URL without protocol",
			urlStr:    "gitlab.com/team/awesome-ext",
			wantHost:  "gitlab.com",
			wantOwner: "team",
			wantRepo:  "awesome-ext",
			wantErr:   false,
		},
		{
			name:      "URL with .git suffix",
			urlStr:    "https://github.com/user/repo.git",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "URL without protocol and with .git",
			urlStr:    "github.com/user/repo.git",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:        "empty URL",
			urlStr:      "",
			wantErr:     true,
			wantErrText: "empty URL",
		},
		{
			name:        "URL with only host",
			urlStr:      "https://github.com/",
			wantErr:     true,
			wantErrText: "invalid repository URL format",
		},
		{
			name:        "URL with only owner",
			urlStr:      "https://github.com/user",
			wantErr:     true,
			wantErrText: "invalid repository URL format",
		},
		{
			name:        "whitespace only",
			urlStr:      "   ",
			wantErr:     true,
			wantErrText: "empty URL",
		},
		{
			name:      "URL with trailing slash",
			urlStr:    "https://github.com/user/repo/",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTP URL (less common)",
			urlStr:    "http://github.com/user/repo",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRepoURL(tt.urlStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.wantErrText != "" && !contains(err.Error(), tt.wantErrText) {
					t.Errorf("Expected error containing %q, got %q", tt.wantErrText, err.Error())
				}
				return
			}

			if got.Host != tt.wantHost {
				t.Errorf("ParseRepoURL() Host = %v, want %v", got.Host, tt.wantHost)
			}
			if got.Owner != tt.wantOwner {
				t.Errorf("ParseRepoURL() Owner = %v, want %v", got.Owner, tt.wantOwner)
			}
			if got.Repo != tt.wantRepo {
				t.Errorf("ParseRepoURL() Repo = %v, want %v", got.Repo, tt.wantRepo)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* ADDITIONAL TABLE-DRIVEN TESTS FOR EDGE CASES                            */
/* ------------------------------------------------------------------------- */

// TestParseRepoURL_AdvancedCases tests advanced URL parsing scenarios
func TestParseRepoURL_AdvancedCases(t *testing.T) {
	tests := []struct {
		name        string
		urlStr      string
		wantHost    string
		wantOwner   string
		wantRepo    string
		wantErr     bool
		wantErrText string
	}{
		{
			name:      "URL with multiple path segments",
			urlStr:    "https://github.com/owner/repo/extra/path",
			wantHost:  "github.com",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "URL with dash in owner",
			urlStr:    "https://github.com/my-org/repo",
			wantHost:  "github.com",
			wantOwner: "my-org",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "URL with underscore in repo",
			urlStr:    "https://github.com/user/my_repo",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "my_repo",
			wantErr:   false,
		},
		{
			name:      "URL with numbers",
			urlStr:    "https://github.com/user123/repo456",
			wantHost:  "github.com",
			wantOwner: "user123",
			wantRepo:  "repo456",
			wantErr:   false,
		},
		{
			name:      "URL with mixed case",
			urlStr:    "https://GitHub.com/User/Repo",
			wantHost:  "GitHub.com",
			wantOwner: "User",
			wantRepo:  "Repo",
			wantErr:   false,
		},
		{
			name:        "URL with no path",
			urlStr:      "https://github.com",
			wantErr:     true,
			wantErrText: "invalid repository URL format",
		},
		{
			name:        "URL with single path component",
			urlStr:      "github.com/justowner",
			wantErr:     true,
			wantErrText: "invalid repository URL format",
		},
		{
			name:      "GitLab subgroups URL",
			urlStr:    "https://gitlab.com/group/subgroup/repo",
			wantHost:  "gitlab.com",
			wantOwner: "group",
			wantRepo:  "subgroup", // Takes first two segments
			wantErr:   false,
		},
		{
			name:      "URL with query parameters (should ignore)",
			urlStr:    "https://github.com/user/repo?ref=main",
			wantHost:  "github.com",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRepoURL(tt.urlStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.wantErrText != "" && !contains(err.Error(), tt.wantErrText) {
					t.Errorf("Expected error containing %q, got %q", tt.wantErrText, err.Error())
				}
				return
			}

			if got.Host != tt.wantHost {
				t.Errorf("ParseRepoURL() Host = %v, want %v", got.Host, tt.wantHost)
			}
			if got.Owner != tt.wantOwner {
				t.Errorf("ParseRepoURL() Owner = %v, want %v", got.Owner, tt.wantOwner)
			}
			if got.Repo != tt.wantRepo {
				t.Errorf("ParseRepoURL() Repo = %v, want %v", got.Repo, tt.wantRepo)
			}
		})
	}
}

// TestParseRepoURL_VariousGitHosts tests URL parsing for various git hosting services
func TestParseRepoURL_VariousGitHosts(t *testing.T) {
	tests := []struct {
		name      string
		urlStr    string
		wantHost  string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "Bitbucket",
			urlStr:    "https://bitbucket.org/user/repo",
			wantHost:  "bitbucket.org",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "Self-hosted GitLab",
			urlStr:    "https://gitlab.company.com/team/extension",
			wantHost:  "gitlab.company.com",
			wantOwner: "team",
			wantRepo:  "extension",
			wantErr:   false,
		},
		{
			name:      "GitHub Enterprise",
			urlStr:    "https://github.enterprise.com/org/project",
			wantHost:  "github.enterprise.com",
			wantOwner: "org",
			wantRepo:  "project",
			wantErr:   false,
		},
		{
			name:      "Gitea instance",
			urlStr:    "https://gitea.example.com/developer/tool",
			wantHost:  "gitea.example.com",
			wantOwner: "developer",
			wantRepo:  "tool",
			wantErr:   false,
		},
		{
			name:      "Custom git server",
			urlStr:    "https://git.example.org/team/app",
			wantHost:  "git.example.org",
			wantOwner: "team",
			wantRepo:  "app",
			wantErr:   false,
		},
		{
			name:      "Self-hosted with port",
			urlStr:    "https://git.company.com:8443/user/repo",
			wantHost:  "git.company.com:8443",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "Bitbucket without protocol",
			urlStr:    "bitbucket.org/workspace/repository",
			wantHost:  "bitbucket.org",
			wantOwner: "workspace",
			wantRepo:  "repository",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRepoURL(tt.urlStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Host != tt.wantHost {
				t.Errorf("ParseRepoURL() Host = %v, want %v", got.Host, tt.wantHost)
			}
			if got.Owner != tt.wantOwner {
				t.Errorf("ParseRepoURL() Owner = %v, want %v", got.Owner, tt.wantOwner)
			}
			if got.Repo != tt.wantRepo {
				t.Errorf("ParseRepoURL() Repo = %v, want %v", got.Repo, tt.wantRepo)
			}
		})
	}
}
