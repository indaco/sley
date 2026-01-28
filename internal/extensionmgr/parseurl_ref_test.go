package extensionmgr

import (
	"testing"
)

/* ------------------------------------------------------------------------- */
/* TABLE-DRIVEN TESTS FOR VERSION/BRANCH/COMMIT REF PARSING               */
/* ------------------------------------------------------------------------- */

// TestParseRepoURL_WithRef tests parsing URLs with version/branch/commit refs
func TestParseRepoURL_WithRef(t *testing.T) {
	tests := []struct {
		name        string
		urlStr      string
		wantHost    string
		wantOwner   string
		wantRepo    string
		wantSubdir  string
		wantRef     string
		wantErr     bool
		wantErrText string
	}{
		{
			name:       "URL with version tag",
			urlStr:     "https://github.com/user/repo@v1.0.0",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "v1.0.0",
			wantErr:    false,
		},
		{
			name:       "URL with branch name",
			urlStr:     "https://github.com/user/repo@develop",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "develop",
			wantErr:    false,
		},
		{
			name:       "URL with commit hash (short)",
			urlStr:     "https://github.com/user/repo@abc123",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "abc123",
			wantErr:    false,
		},
		{
			name:       "URL with commit hash (long)",
			urlStr:     "https://github.com/user/repo@abc123def456789012345678901234567890",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "abc123def456789012345678901234567890",
			wantErr:    false,
		},
		{
			name:       "URL with subdirectory and version tag",
			urlStr:     "https://github.com/indaco/sley/contrib/extensions/changelog-generator@v2.0.0",
			wantHost:   "github.com",
			wantOwner:  "indaco",
			wantRepo:   "sley",
			wantSubdir: "contrib/extensions/changelog-generator",
			wantRef:    "v2.0.0",
			wantErr:    false,
		},
		{
			name:       "URL without protocol with version",
			urlStr:     "github.com/user/repo@v1.2.3",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "v1.2.3",
			wantErr:    false,
		},
		{
			name:       "URL with main branch",
			urlStr:     "github.com/user/repo@main",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "main",
			wantErr:    false,
		},
		{
			name:       "URL with master branch",
			urlStr:     "github.com/user/repo@master",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "master",
			wantErr:    false,
		},
		{
			name:       "URL with ref containing slash (release branch)",
			urlStr:     "github.com/user/repo@release/v1.0",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "release/v1.0",
			wantErr:    false,
		},
		{
			name:       "URL with ref containing dash",
			urlStr:     "github.com/user/repo@feature-branch",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "feature-branch",
			wantErr:    false,
		},
		{
			name:       "URL with ref containing underscore",
			urlStr:     "github.com/user/repo@feature_branch",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "feature_branch",
			wantErr:    false,
		},
		{
			name:       "GitLab URL with version",
			urlStr:     "https://gitlab.com/org/project@v2.5.0",
			wantHost:   "gitlab.com",
			wantOwner:  "org",
			wantRepo:   "project",
			wantSubdir: "",
			wantRef:    "v2.5.0",
			wantErr:    false,
		},
		{
			name:       "Bitbucket URL with branch",
			urlStr:     "https://bitbucket.org/team/repo@develop",
			wantHost:   "bitbucket.org",
			wantOwner:  "team",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "develop",
			wantErr:    false,
		},
		{
			name:       "Self-hosted with version",
			urlStr:     "https://git.company.com/team/ext@v1.0.0",
			wantHost:   "git.company.com",
			wantOwner:  "team",
			wantRepo:   "ext",
			wantSubdir: "",
			wantRef:    "v1.0.0",
			wantErr:    false,
		},
		{
			name:        "URL with empty ref after @",
			urlStr:      "github.com/user/repo@",
			wantErr:     true,
			wantErrText: "empty ref specified after @",
		},
		{
			name:       "URL with .git suffix and version",
			urlStr:     "https://github.com/user/repo.git@v1.0.0",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "v1.0.0",
			wantErr:    false,
		},
		{
			name:       "URL with subdirectory and branch with slash",
			urlStr:     "github.com/org/repo/extensions/my-ext@release/1.0",
			wantHost:   "github.com",
			wantOwner:  "org",
			wantRepo:   "repo",
			wantSubdir: "extensions/my-ext",
			wantRef:    "release/1.0",
			wantErr:    false,
		},
		{
			name:       "URL without ref (backward compatibility)",
			urlStr:     "github.com/user/repo",
			wantHost:   "github.com",
			wantOwner:  "user",
			wantRepo:   "repo",
			wantSubdir: "",
			wantRef:    "",
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
			if got.Ref != tt.wantRef {
				t.Errorf("ParseRepoURL() Ref = %q, want %q", got.Ref, tt.wantRef)
			}
		})
	}
}
