package extensionmgr

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/indaco/sley/internal/printer"
)

// GitErrorInfo contains user-friendly error information
type GitErrorInfo struct {
	Category    string
	Message     string
	Suggestions []string
}

// gitErrorPattern represents a pattern to match in git error output
type gitErrorPattern struct {
	pattern     *regexp.Regexp
	category    string
	message     string
	suggestions []string
}

// gitErrorPatterns contains common git error patterns with helpful suggestions.
// Order matters: more specific patterns should come before general ones.
var gitErrorPatterns = []gitErrorPattern{
	// Repository not found
	{
		pattern:  regexp.MustCompile(`(?i)fatal: repository .* not found`),
		category: "repository_not_found",
		message:  "Repository not found",
		suggestions: []string{
			"Verify the repository URL is correct",
			"Check if the repository is private and requires authentication",
			"Ensure the repository exists on the hosting platform",
		},
	},
	// SSL/Certificate errors (must come before generic "unable to access")
	{
		pattern:  regexp.MustCompile(`(?i)SSL certificate problem|certificate verify failed`),
		category: "ssl_error",
		message:  "SSL certificate verification failed",
		suggestions: []string{
			"Check your system's SSL certificates are up to date",
			"Verify the git server's SSL certificate is valid",
			"Check if a proxy is interfering with SSL connections",
		},
	},
	// Generic network access error
	{
		pattern:  regexp.MustCompile(`(?i)fatal: unable to access`),
		category: "network_error",
		message:  "Unable to access repository",
		suggestions: []string{
			"Check your network connection",
			"Verify you can access the git hosting platform",
			"Check if a proxy or firewall is blocking access",
		},
	},
	{
		pattern:  regexp.MustCompile(`(?i)fatal: [Rr]emote branch .* not found`),
		category: "ref_not_found",
		message:  "Branch, tag, or commit not found",
		suggestions: []string{
			"Verify the branch/tag name is correct",
			"Check available branches: git ls-remote <url>",
			"Use the repository's default branch or specify a valid ref",
		},
	},
	{
		pattern:  regexp.MustCompile(`(?i)fatal: couldn't find remote ref`),
		category: "ref_not_found",
		message:  "Branch, tag, or commit not found",
		suggestions: []string{
			"Verify the branch/tag/commit reference is correct",
			"Check available refs: git ls-remote <url>",
			"Ensure the ref exists in the remote repository",
		},
	},
	{
		pattern:  regexp.MustCompile(`(?i)Authentication failed|fatal: could not read (Username|Password)`),
		category: "auth_required",
		message:  "Authentication required",
		suggestions: []string{
			"Configure git credentials for this repository",
			"Use SSH URL if you have keys configured",
			"Check if you have access to this private repository",
		},
	},
	{
		pattern:  regexp.MustCompile(`(?i)Permission denied|fatal: The remote end hung up unexpectedly`),
		category: "permission_denied",
		message:  "Permission denied",
		suggestions: []string{
			"Verify you have access to this repository",
			"Check your git credentials or SSH keys",
			"Ensure you have read access to the repository",
		},
	},
	{
		pattern:  regexp.MustCompile(`(?i)Connection timed out|fatal: unable to connect`),
		category: "network_timeout",
		message:  "Connection timed out",
		suggestions: []string{
			"Check your network connection",
			"Verify the git server is accessible",
			"Try again in a few moments",
		},
	},
	{
		pattern:  regexp.MustCompile(`(?i)Could not resolve host|Temporary failure in name resolution`),
		category: "dns_error",
		message:  "Could not resolve hostname",
		suggestions: []string{
			"Check your network connection",
			"Verify the hostname is correct",
			"Check your DNS settings",
		},
	},
}

// parseGitError analyzes git error output and returns helpful context.
// It matches the output against known error patterns and returns structured
// information with actionable suggestions for the user.
func parseGitError(gitOutput string) *GitErrorInfo {
	// Try to match against known patterns
	for _, pattern := range gitErrorPatterns {
		if pattern.pattern.MatchString(gitOutput) {
			return &GitErrorInfo{
				Category:    pattern.category,
				Message:     pattern.message,
				Suggestions: pattern.suggestions,
			}
		}
	}

	// No specific pattern matched, return generic error info
	return nil
}

// FormatGitError creates a user-friendly error message from git output.
// It parses the git error output to identify common issues and provides
// context-aware suggestions to help users resolve the problem.
//
// If a known error pattern is detected, returns a formatted error with
// helpful suggestions. Otherwise, returns the original error with git output.
func FormatGitError(err error, gitOutput string, repoURL *RepoURL) error {
	if err == nil {
		return nil
	}

	// Parse git output for known error patterns
	errorInfo := parseGitError(gitOutput)

	// If we identified a specific error, format it nicely
	if errorInfo != nil {
		return &GitCloneError{
			RepoURL:     repoURL,
			ErrorInfo:   errorInfo,
			GitOutput:   gitOutput,
			OriginalErr: err,
		}
	}

	// Unknown error - return with git output for debugging
	return fmt.Errorf("git clone failed: %w\noutput: %s", err, gitOutput)
}

// GitCloneError represents a structured git clone error with helpful context
type GitCloneError struct {
	RepoURL     *RepoURL
	ErrorInfo   *GitErrorInfo
	GitOutput   string
	OriginalErr error
}

// Error implements the error interface
func (e *GitCloneError) Error() string {
	var sb strings.Builder

	// Build repository identification
	repoID := e.RepoURL.String()
	if e.RepoURL.Subdir != "" {
		repoID = fmt.Sprintf("%s (subdirectory: %s)", repoID, e.RepoURL.Subdir)
	}

	// Main error message
	fmt.Fprintf(&sb, "Failed to clone repository: %s\n\n", repoID)

	// Error category and message
	fmt.Fprintf(&sb, "Error: %s\n", e.ErrorInfo.Message)

	// Suggestions
	if len(e.ErrorInfo.Suggestions) > 0 {
		sb.WriteString("\nSuggestions:\n")
		for _, suggestion := range e.ErrorInfo.Suggestions {
			fmt.Fprintf(&sb, "  - %s\n", suggestion)
		}
	}

	return sb.String()
}

// Unwrap returns the original error for error unwrapping
func (e *GitCloneError) Unwrap() error {
	return e.OriginalErr
}

// PrintGitError prints a formatted git error to the console using the printer package.
// This provides a styled, user-friendly error message in the terminal.
func PrintGitError(err error) {
	var gitErr *GitCloneError
	if errors.As(err, &gitErr) {
		// Print using printer package for consistent styling
		printer.PrintError(fmt.Sprintf("Failed to clone repository: %s", gitErr.RepoURL.String()))
		fmt.Println()
		printer.PrintError(fmt.Sprintf("Error: %s", gitErr.ErrorInfo.Message))

		if len(gitErr.ErrorInfo.Suggestions) > 0 {
			fmt.Println()
			printer.PrintInfo("Suggestions:")
			for _, suggestion := range gitErr.ErrorInfo.Suggestions {
				fmt.Printf("  - %s\n", suggestion)
			}
		}
	} else {
		// Fallback for non-GitCloneError errors
		printer.PrintError(fmt.Sprintf("Error: %v", err))
	}
}
