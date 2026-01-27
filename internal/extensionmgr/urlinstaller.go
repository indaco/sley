package extensionmgr

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/printer"
)

// RepoURL represents a parsed repository URL
type RepoURL struct {
	Host   string // github.com, gitlab.com, etc.
	Owner  string // user or organization
	Repo   string // repository name
	Subdir string // optional subdirectory path within the repository
	Raw    string // original URL
}

// ParseRepoURL parses various repository URL formats into a RepoURL struct
func ParseRepoURL(urlStr string) (*RepoURL, error) {
	// Trim whitespace
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return nil, fmt.Errorf("empty URL")
	}

	// Handle URLs without protocol
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	host := parsed.Host
	if host == "" {
		return nil, fmt.Errorf("invalid URL: missing host")
	}

	// Extract owner and repo from path
	path := strings.TrimPrefix(parsed.Path, "/")
	path = strings.TrimSuffix(path, "/")

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid repository URL format: expected owner/repo")
	}

	// Remove .git suffix from repo name if present
	repo := strings.TrimSuffix(parts[1], ".git")

	// Extract subdirectory if present (anything beyond owner/repo)
	var subdir string
	if len(parts) > 2 {
		subdir = strings.Join(parts[2:], "/")
	}

	return &RepoURL{
		Host:   host,
		Owner:  parts[0],
		Repo:   repo,
		Subdir: subdir,
		Raw:    urlStr,
	}, nil
}

// IsGitHubURL checks if the URL is a GitHub repository
func (r *RepoURL) IsGitHubURL() bool {
	return r.Host == "github.com"
}

// IsGitLabURL checks if the URL is a GitLab repository
func (r *RepoURL) IsGitLabURL() bool {
	return r.Host == "gitlab.com"
}

// CloneURL returns the HTTPS clone URL for the repository
func (r *RepoURL) CloneURL() string {
	return fmt.Sprintf("https://%s/%s/%s.git", r.Host, r.Owner, r.Repo)
}

// String returns a human-readable representation
func (r *RepoURL) String() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Repo)
}

// CloneRepository clones a repository to a temporary directory
func CloneRepository(repoURL *RepoURL) (string, error) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "sley-ext-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clone the repository with timeout
	ctx, cancel := context.WithTimeout(context.Background(), core.TimeoutGit)
	defer cancel()

	cloneURL := repoURL.CloneURL()
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", cloneURL, tempDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up temp dir on failure
		_ = os.RemoveAll(tempDir)
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("git clone timeout after %v: %w\noutput: %s", core.TimeoutGit, err, string(output))
		}
		return "", fmt.Errorf("git clone failed: %w\noutput: %s", err, string(output))
	}

	return tempDir, nil
}

// InstallFromURL clones a repository and installs the extension
func InstallFromURL(urlStr, configPath, extensionDirectory string) error {
	// Parse the URL
	repoURL, err := ParseRepoURL(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Clone the repository (any git-accessible host is supported)
	printer.PrintInfo(fmt.Sprintf("Cloning %s...", repoURL.String()))
	tempDir, err := CloneRepository(repoURL)
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w", repoURL.String(), err)
	}

	// Clean up temp directory after installation
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			printer.PrintWarning(fmt.Sprintf("Failed to clean up temp directory %s: %v", tempDir, err))
		}
	}()

	// Navigate to subdirectory if specified
	extensionPath := tempDir
	if repoURL.Subdir != "" {
		extensionPath = fmt.Sprintf("%s/%s", tempDir, repoURL.Subdir)
		if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
			return fmt.Errorf("subdirectory %q not found in repository %s", repoURL.Subdir, repoURL.String())
		} else if err != nil {
			return fmt.Errorf("failed to access subdirectory %q: %w", repoURL.Subdir, err)
		}
	}

	// Install from the cloned directory (or subdirectory)
	if repoURL.Subdir != "" {
		printer.PrintInfo(fmt.Sprintf("Installing extension from %s (subdirectory: %s)...", repoURL.String(), repoURL.Subdir))
	} else {
		printer.PrintInfo(fmt.Sprintf("Installing extension from %s...", repoURL.String()))
	}
	return registerLocalExtension(extensionPath, configPath, extensionDirectory)
}

// IsURL checks if a string looks like a URL (has a host and path)
func IsURL(str string) bool {
	str = strings.TrimSpace(str)

	// Check for http/https prefix
	if strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://") {
		return true
	}

	// Check for github.com/owner/repo or gitlab.com/owner/repo pattern
	if strings.Contains(str, "github.com/") || strings.Contains(str, "gitlab.com/") {
		parts := strings.Split(str, "/")
		return len(parts) >= 3 // host/owner/repo minimum
	}

	return false
}

// ValidateGitAvailable checks if git is available in the system
func ValidateGitAvailable() error {
	// Short timeout for version check
	ctx, cancel := context.WithTimeout(context.Background(), core.TimeoutShort)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "--version")
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git version check timeout: %w (required for URL-based installation)", err)
		}
		return fmt.Errorf("git is not available: %w (required for URL-based installation)", err)
	}
	return nil
}
