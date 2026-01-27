package extensionmgr

import (
	"errors"
	"strings"
	"testing"
)

/* ------------------------------------------------------------------------- */
/* TABLE-DRIVEN TESTS FOR GIT ERROR PARSING                                */
/* ------------------------------------------------------------------------- */

// TestParseGitError tests parsing of various git error outputs
func TestParseGitError(t *testing.T) {
	tests := []struct {
		name            string
		gitOutput       string
		wantCategory    string
		wantMessage     string
		wantSuggestions int // number of suggestions expected
	}{
		{
			name:            "repository not found",
			gitOutput:       "fatal: repository 'https://github.com/user/nonexistent.git/' not found",
			wantCategory:    "repository_not_found",
			wantMessage:     "Repository not found",
			wantSuggestions: 3,
		},
		{
			name:            "repository not found (case insensitive)",
			gitOutput:       "Fatal: Repository 'https://github.com/user/repo.git/' NOT FOUND",
			wantCategory:    "repository_not_found",
			wantMessage:     "Repository not found",
			wantSuggestions: 3,
		},
		{
			name:            "unable to access",
			gitOutput:       "fatal: unable to access 'https://github.com/user/repo.git/': Could not resolve host: github.com",
			wantCategory:    "network_error",
			wantMessage:     "Unable to access repository",
			wantSuggestions: 3,
		},
		{
			name:            "remote branch not found",
			gitOutput:       "fatal: Remote branch v9999.0.0 not found in upstream origin",
			wantCategory:    "ref_not_found",
			wantMessage:     "Branch, tag, or commit not found",
			wantSuggestions: 3,
		},
		{
			name:            "couldn't find remote ref",
			gitOutput:       "fatal: couldn't find remote ref develop",
			wantCategory:    "ref_not_found",
			wantMessage:     "Branch, tag, or commit not found",
			wantSuggestions: 3,
		},
		{
			name:            "authentication failed",
			gitOutput:       "Authentication failed for 'https://github.com/private/repo.git/'",
			wantCategory:    "auth_required",
			wantMessage:     "Authentication required",
			wantSuggestions: 3,
		},
		{
			name:            "could not read username",
			gitOutput:       "fatal: could not read Username for 'https://github.com': terminal prompts disabled",
			wantCategory:    "auth_required",
			wantMessage:     "Authentication required",
			wantSuggestions: 3,
		},
		{
			name:            "permission denied",
			gitOutput:       "Permission denied (publickey).\r\nfatal: Could not read from remote repository.",
			wantCategory:    "permission_denied",
			wantMessage:     "Permission denied",
			wantSuggestions: 3,
		},
		{
			name:            "remote hung up",
			gitOutput:       "fatal: The remote end hung up unexpectedly",
			wantCategory:    "permission_denied",
			wantMessage:     "Permission denied",
			wantSuggestions: 3,
		},
		{
			name:            "connection timed out",
			gitOutput:       "Connection timed out after 60000 milliseconds",
			wantCategory:    "network_timeout",
			wantMessage:     "Connection timed out",
			wantSuggestions: 3,
		},
		{
			name:            "could not resolve host",
			gitOutput:       "Could not resolve host: invalid-host.example.com",
			wantCategory:    "dns_error",
			wantMessage:     "Could not resolve hostname",
			wantSuggestions: 3,
		},
		{
			name:            "temporary failure in name resolution",
			gitOutput:       "Temporary failure in name resolution",
			wantCategory:    "dns_error",
			wantMessage:     "Could not resolve hostname",
			wantSuggestions: 3,
		},
		{
			name:            "SSL certificate problem",
			gitOutput:       "SSL certificate problem: unable to get local issuer certificate",
			wantCategory:    "ssl_error",
			wantMessage:     "SSL certificate verification failed",
			wantSuggestions: 3,
		},
		{
			name:            "certificate verify failed",
			gitOutput:       "fatal: unable to access 'https://git.example.com/repo.git/': certificate verify failed",
			wantCategory:    "ssl_error",
			wantMessage:     "SSL certificate verification failed",
			wantSuggestions: 3,
		},
		{
			name:            "unknown error returns nil",
			gitOutput:       "some completely unknown error message",
			wantCategory:    "",
			wantMessage:     "",
			wantSuggestions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGitError(tt.gitOutput)

			// Handle nil case for unknown errors
			if tt.wantCategory == "" {
				if got != nil {
					t.Errorf("parseGitError() expected nil for unknown error, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatalf("parseGitError() returned nil, want error info")
			}

			if got.Category != tt.wantCategory {
				t.Errorf("parseGitError() Category = %v, want %v", got.Category, tt.wantCategory)
			}

			if got.Message != tt.wantMessage {
				t.Errorf("parseGitError() Message = %v, want %v", got.Message, tt.wantMessage)
			}

			if len(got.Suggestions) != tt.wantSuggestions {
				t.Errorf("parseGitError() got %d suggestions, want %d", len(got.Suggestions), tt.wantSuggestions)
			}

			// Verify all suggestions are non-empty
			for i, suggestion := range got.Suggestions {
				if suggestion == "" {
					t.Errorf("parseGitError() suggestion[%d] is empty", i)
				}
			}
		})
	}
}

// TestFormatGitError tests formatting of git errors with context
func TestFormatGitError(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		gitOutput       string
		repoURL         *RepoURL
		wantErrorType   bool // true if we expect GitCloneError
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:      "repository not found creates structured error",
			err:       errors.New("exit status 128"),
			gitOutput: "fatal: repository 'https://github.com/user/nonexistent.git/' not found",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "user",
				Repo:  "nonexistent",
			},
			wantErrorType: true,
			wantContains: []string{
				"Failed to clone repository",
				"user/nonexistent",
				"Repository not found",
				"Suggestions",
			},
		},
		{
			name:      "repository not found with ref",
			err:       errors.New("exit status 128"),
			gitOutput: "fatal: repository 'https://github.com/user/repo.git/' not found",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "user",
				Repo:  "repo",
				Ref:   "v1.0.0",
			},
			wantErrorType: true,
			wantContains: []string{
				"user/repo@v1.0.0",
				"Repository not found",
			},
		},
		{
			name:      "repository with subdirectory",
			err:       errors.New("exit status 128"),
			gitOutput: "fatal: repository 'https://github.com/org/repo.git/' not found",
			repoURL: &RepoURL{
				Host:   "github.com",
				Owner:  "org",
				Repo:   "repo",
				Subdir: "contrib/extensions",
			},
			wantErrorType: true,
			wantContains: []string{
				"org/repo",
				"subdirectory: contrib/extensions",
			},
		},
		{
			name:      "ref not found error",
			err:       errors.New("exit status 128"),
			gitOutput: "fatal: Remote branch nonexistent not found in upstream origin",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "user",
				Repo:  "repo",
			},
			wantErrorType: true,
			wantContains: []string{
				"Branch, tag, or commit not found",
				"git ls-remote",
			},
		},
		{
			name:      "authentication required",
			err:       errors.New("exit status 128"),
			gitOutput: "Authentication failed for 'https://github.com/private/repo.git/'",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "private",
				Repo:  "repo",
			},
			wantErrorType: true,
			wantContains: []string{
				"Authentication required",
				"Configure git credentials",
			},
		},
		{
			name:      "unknown error falls back to generic",
			err:       errors.New("unknown error"),
			gitOutput: "some weird error that we don't recognize",
			repoURL: &RepoURL{
				Host:  "github.com",
				Owner: "user",
				Repo:  "repo",
			},
			wantErrorType: false,
			wantContains: []string{
				"git clone failed",
				"some weird error",
			},
		},
		{
			name:          "nil error returns nil",
			err:           nil,
			gitOutput:     "",
			repoURL:       nil,
			wantErrorType: false,
			wantContains:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatGitError(tt.err, tt.gitOutput, tt.repoURL)

			// Handle nil case
			if tt.err == nil {
				if got != nil {
					t.Errorf("FormatGitError() expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Fatalf("FormatGitError() returned nil, want error")
			}

			// Check error type
			var gitErr *GitCloneError
			isGitCloneError := errors.As(got, &gitErr)
			if isGitCloneError != tt.wantErrorType {
				t.Errorf("FormatGitError() GitCloneError type = %v, want %v", isGitCloneError, tt.wantErrorType)
			}

			// Check error message contains expected strings
			errMsg := got.Error()
			for _, want := range tt.wantContains {
				if !strings.Contains(errMsg, want) {
					t.Errorf("FormatGitError() error missing %q\nGot: %s", want, errMsg)
				}
			}

			// Check error message doesn't contain unwanted strings
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(errMsg, notWant) {
					t.Errorf("FormatGitError() error should not contain %q\nGot: %s", notWant, errMsg)
				}
			}
		})
	}
}

// TestGitCloneError_Unwrap tests error unwrapping
func TestGitCloneError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	gitErr := &GitCloneError{
		RepoURL: &RepoURL{
			Host:  "github.com",
			Owner: "user",
			Repo:  "repo",
		},
		ErrorInfo: &GitErrorInfo{
			Category: "test",
			Message:  "test error",
		},
		OriginalErr: originalErr,
	}

	unwrapped := gitErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// Test with errors.Is
	if !errors.Is(gitErr, originalErr) {
		t.Error("errors.Is() should return true for unwrapped error")
	}
}

// TestGitCloneError_Error tests error message formatting
func TestGitCloneError_Error(t *testing.T) {
	tests := []struct {
		name     string
		gitErr   *GitCloneError
		contains []string
	}{
		{
			name: "basic error with suggestions",
			gitErr: &GitCloneError{
				RepoURL: &RepoURL{
					Host:  "github.com",
					Owner: "user",
					Repo:  "repo",
				},
				ErrorInfo: &GitErrorInfo{
					Category: "test",
					Message:  "Test error message",
					Suggestions: []string{
						"First suggestion",
						"Second suggestion",
					},
				},
				OriginalErr: errors.New("original"),
			},
			contains: []string{
				"Failed to clone repository: user/repo",
				"Error: Test error message",
				"Suggestions:",
				"• First suggestion",
				"• Second suggestion",
			},
		},
		{
			name: "error with subdirectory",
			gitErr: &GitCloneError{
				RepoURL: &RepoURL{
					Host:   "github.com",
					Owner:  "org",
					Repo:   "repo",
					Subdir: "path/to/ext",
				},
				ErrorInfo: &GitErrorInfo{
					Category:    "test",
					Message:     "Test error",
					Suggestions: []string{"Suggestion"},
				},
				OriginalErr: errors.New("original"),
			},
			contains: []string{
				"org/repo",
				"subdirectory: path/to/ext",
			},
		},
		{
			name: "error without suggestions",
			gitErr: &GitCloneError{
				RepoURL: &RepoURL{
					Host:  "github.com",
					Owner: "user",
					Repo:  "repo",
				},
				ErrorInfo: &GitErrorInfo{
					Category:    "test",
					Message:     "Test error",
					Suggestions: []string{},
				},
				OriginalErr: errors.New("original"),
			},
			contains: []string{
				"Failed to clone repository: user/repo",
				"Error: Test error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gitErr.Error()

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("Error() missing %q\nGot: %s", want, got)
				}
			}
		})
	}
}
