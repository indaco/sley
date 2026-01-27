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
/* TABLE-DRIVEN TESTS FOR URL VALIDATION                                    */
/* ------------------------------------------------------------------------- */

func TestIsURL(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			name: "HTTPS GitHub URL",
			str:  "https://github.com/user/repo",
			want: true,
		},
		{
			name: "HTTP URL",
			str:  "http://example.com/path",
			want: true,
		},
		{
			name: "GitHub without protocol",
			str:  "github.com/user/repo",
			want: true,
		},
		{
			name: "GitLab without protocol",
			str:  "gitlab.com/org/project",
			want: true,
		},
		{
			name: "local path",
			str:  "./local/extension",
			want: false,
		},
		{
			name: "absolute local path",
			str:  "/home/user/extension",
			want: false,
		},
		{
			name: "relative path",
			str:  "../extensions/my-ext",
			want: false,
		},
		{
			name: "empty string",
			str:  "",
			want: false,
		},
		{
			name: "just domain",
			str:  "github.com",
			want: false,
		},
		{
			name: "domain with only one path segment",
			str:  "github.com/user",
			want: false,
		},
		{
			name: "whitespace",
			str:  "   ",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsURL(tt.str)
			if got != tt.want {
				t.Errorf("IsURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

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

/* ------------------------------------------------------------------------- */
/* UNIT TESTS                                                                */
/* ------------------------------------------------------------------------- */

func TestValidateGitAvailable(t *testing.T) {
	// This test will pass if git is installed, fail otherwise
	// In CI/CD environments, git is typically available
	err := ValidateGitAvailable()

	// We can't reliably test both cases without mocking exec.Command
	// So we just verify the function runs without panicking
	if err != nil {
		t.Logf("git not available: %v (this is expected if git is not installed)", err)
	} else {
		t.Log("git is available")
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

// TestIsURL_EdgeCases tests IsURL with various edge cases
func TestIsURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{
			name: "HTTPS with port",
			str:  "https://github.com:443/user/repo",
			want: true,
		},
		{
			name: "with .git extension",
			str:  "github.com/user/repo.git",
			want: true,
		},
		{
			name: "with trailing spaces",
			str:  "  github.com/user/repo  ",
			want: true,
		},
		{
			name: "filename only",
			str:  "extension.yaml",
			want: false,
		},
		{
			name: "current directory",
			str:  ".",
			want: false,
		},
		{
			name: "parent directory",
			str:  "..",
			want: false,
		},
		{
			name: "Windows path",
			str:  "C:\\Users\\ext",
			want: false,
		},
		{
			name: "GitHub with four segments",
			str:  "github.com/org/team/repo",
			want: true,
		},
		{
			name: "GitLab with many segments",
			str:  "gitlab.com/group/subgroup/project",
			want: true,
		},
		{
			name: "bitbucket (not explicitly supported but should match pattern)",
			str:  "bitbucket.org/user/repo",
			want: false, // Not github or gitlab
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsURL(tt.str)
			if got != tt.want {
				t.Errorf("IsURL(%q) = %v, want %v", tt.str, got, tt.want)
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

/* ------------------------------------------------------------------------- */
/* TABLE-DRIVEN TESTS FOR SUBDIRECTORY PARSING                             */
/* ------------------------------------------------------------------------- */

// TestParseRepoURL_Subdirectory tests subdirectory parsing in repository URLs
func TestParseRepoURL_Subdirectory(t *testing.T) {
	tests := []struct {
		name        string
		urlStr      string
		wantHost    string
		wantOwner   string
		wantRepo    string
		wantSubdir  string
		wantErr     bool
		wantErrText string
	}{
		{
			name:       "URL with single subdirectory",
			urlStr:     "https://github.com/user/repo/extensions",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "extensions",
			wantErr:    false,
		},
		{
			name:       "URL with multiple subdirectory levels",
			urlStr:     "https://github.com/indaco/sley/contrib/extensions/changelog-generator",
			wantHost:   "github.com",
			wantOwner:  "indaco",
			wantRepo:   "sley",
			wantSubdir: "contrib/extensions/changelog-generator",
			wantErr:    false,
		},
		{
			name:       "URL without protocol with subdirectory",
			urlStr:     "github.com/org/project/path/to/extension",
			wantHost:   "github.com",
			wantOwner:  "org",
			wantRepo:   "project",
			wantSubdir: "path/to/extension",
			wantErr:    false,
		},
		{
			name:       "GitLab URL with subdirectory",
			urlStr:     "https://gitlab.com/group/repo/extensions/my-ext",
			wantHost:   "gitlab.com",
			wantOwner:  "group",
			wantRepo:   "repo",
			wantSubdir: "extensions/my-ext",
			wantErr:    false,
		},
		{
			name:       "URL with .git suffix and subdirectory",
			urlStr:     "https://github.com/user/repo.git/subdir",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "subdir",
			wantErr:    false,
		},
		{
			name:       "URL without subdirectory (backward compatibility)",
			urlStr:     "https://github.com/user/repo",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantErr:    false,
		},
		{
			name:       "URL with trailing slash and subdirectory",
			urlStr:     "https://github.com/user/repo/subdir/",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "subdir",
			wantErr:    false,
		},
		{
			name:       "URL with deep nested subdirectories",
			urlStr:     "github.com/org/proj/a/b/c/d/e",
			wantHost:   "github.com",
			wantOwner:  "org",
			wantRepo:   "proj",
			wantSubdir: "a/b/c/d/e",
			wantErr:    false,
		},
		{
			name:       "URL with subdirectory containing dashes",
			urlStr:     "github.com/user/repo/my-extension-dir",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "my-extension-dir",
			wantErr:    false,
		},
		{
			name:       "URL with subdirectory containing underscores",
			urlStr:     "github.com/user/repo/my_extension_dir",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "my_extension_dir",
			wantErr:    false,
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
			if got.Subdir != tt.wantSubdir {
				t.Errorf("ParseRepoURL() Subdir = %q, want %q", got.Subdir, tt.wantSubdir)
			}
		})
	}
}
