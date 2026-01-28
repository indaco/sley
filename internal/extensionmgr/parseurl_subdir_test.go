package extensionmgr

import (
	"testing"
)

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
