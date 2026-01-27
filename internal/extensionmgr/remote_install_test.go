//go:build !short

package extensionmgr

import (
	"testing"
)

/* ------------------------------------------------------------------------- */
/* INTEGRATION TESTS FOR REMOTE INSTALLATION                                */
/* ------------------------------------------------------------------------- */

// Note: These tests make actual network calls and clone real repositories.
// They are skipped when running with -short flag.
// They may be flaky depending on network conditions and should be used
// primarily for manual verification.

// TestRemoteInstallation_VariousGitHosts tests URL parsing for various git hosting services
// This test doesn't make network calls, it just validates that various hosts are accepted
func TestRemoteInstallation_VariousGitHosts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "GitHub",
			url:     "github.com/user/repo",
			wantErr: false,
		},
		{
			name:    "GitLab",
			url:     "gitlab.com/user/repo",
			wantErr: false,
		},
		{
			name:    "Bitbucket",
			url:     "bitbucket.org/user/repo",
			wantErr: false,
		},
		{
			name:    "Self-hosted GitLab",
			url:     "gitlab.company.com/team/extension",
			wantErr: false,
		},
		{
			name:    "GitHub Enterprise",
			url:     "github.enterprise.com/org/repo",
			wantErr: false,
		},
		{
			name:    "Custom git host",
			url:     "git.example.com/user/project",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test URL parsing - should accept any valid git URL format
			repoURL, err := ParseRepoURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && repoURL == nil {
				t.Error("ParseRepoURL() returned nil for valid URL")
			}
		})
	}
}

// TestParseRepoURL_WithSubdirectory tests URL parsing with subdirectories (unit test)
func TestParseRepoURL_WithSubdirectoryIntegration(t *testing.T) {
	// This is a unit test but grouped with integration tests for organization
	tests := []struct {
		name       string
		url        string
		wantSubdir string
		wantErr    bool
	}{
		{
			name:       "no subdirectory",
			url:        "github.com/user/repo",
			wantSubdir: "",
			wantErr:    false,
		},
		{
			name:       "single level subdirectory",
			url:        "github.com/user/repo/extensions",
			wantSubdir: "extensions",
			wantErr:    false,
		},
		{
			name:       "nested subdirectory",
			url:        "github.com/indaco/sley/contrib/extensions/changelog-generator",
			wantSubdir: "contrib/extensions/changelog-generator",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseRepoURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && result.Subdir != tt.wantSubdir {
				t.Errorf("ParseRepoURL() Subdir = %q, want %q", result.Subdir, tt.wantSubdir)
			}
		})
	}
}
