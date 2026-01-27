//go:build !short

package extensionmgr

import (
	"os"
	"path/filepath"
	"testing"
)

/* ------------------------------------------------------------------------- */
/* INTEGRATION TESTS FOR REMOTE INSTALLATION                                */
/* ------------------------------------------------------------------------- */

// Note: These tests make actual network calls and clone real repositories.
// They are skipped when running with -short flag.
// They may be flaky depending on network conditions and should be used
// primarily for manual verification.

// TestRemoteInstallation_UnsupportedHost tests installation from unsupported git hosts
// This test doesn't make network calls, it just tests URL validation
func TestRemoteInstallation_UnsupportedHost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	// Create minimal config
	initialConfig := `extensions: []`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Test with unsupported host (bitbucket)
	testURL := "bitbucket.org/user/repo"

	err := InstallFromURL(testURL, configPath, tmpDir)
	if err == nil {
		t.Error("expected error for unsupported host")
	}

	// Verify error message mentions unsupported host
	if !contains(err.Error(), "unsupported") {
		t.Errorf("expected error about unsupported host, got: %v", err)
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
