package apperrors

import (
	"errors"
	"fmt"
	"testing"
)

func TestVersionFileNotFoundError(t *testing.T) {
	err := &VersionFileNotFoundError{Path: "/path/to/.version"}

	if err.Error() != "version file not found at /path/to/.version" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	// Test errors.As
	var vfErr *VersionFileNotFoundError
	if !errors.As(err, &vfErr) {
		t.Error("expected errors.As to match VersionFileNotFoundError")
	}
}

func TestInvalidVersionError(t *testing.T) {
	tests := []struct {
		version  string
		reason   string
		expected string
	}{
		{"abc", "not a number", `invalid version format "abc": not a number`},
		{"1.2.x", "", "invalid version format: 1.2.x"},
	}

	for _, tt := range tests {
		err := &InvalidVersionError{Version: tt.version, Reason: tt.reason}
		if err.Error() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, err.Error())
		}
	}
}

func TestInvalidBumpTypeError(t *testing.T) {
	err := &InvalidBumpTypeError{BumpType: "huge"}
	expected := "invalid bump type: huge (expected: patch, minor, or major)"

	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestConfigError(t *testing.T) {
	inner := errors.New("file not found")
	err := &ConfigError{Operation: "load", Err: inner}

	if err.Error() != "config load failed: file not found" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !errors.Is(err, inner) {
		t.Error("expected errors.Is to match inner error")
	}

	if err.Unwrap() != inner {
		t.Error("expected Unwrap to return inner error")
	}
}

func TestCommandError(t *testing.T) {
	inner := fmt.Errorf("exit status 1")

	tests := []struct {
		command  string
		timeout  bool
		expected string
	}{
		{"git", false, `command "git" failed: exit status 1`},
		{"make", true, `command "make" timed out: exit status 1`},
	}

	for _, tt := range tests {
		err := &CommandError{Command: tt.command, Err: inner, Timeout: tt.timeout}
		if err.Error() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, err.Error())
		}

		if !errors.Is(err, inner) {
			t.Error("expected errors.Is to match inner error")
		}
	}
}

func TestPathValidationError(t *testing.T) {
	err := &PathValidationError{Path: "../secret", Reason: "path traversal detected"}
	expected := `invalid path "../secret": path traversal detected`

	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestHookError(t *testing.T) {
	inner := errors.New("make: *** No rule to make target")
	err := &HookError{HookName: "pre-build", Err: inner}

	if err.Error() != `hook "pre-build" failed: make: *** No rule to make target` {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !errors.Is(err, inner) {
		t.Error("expected errors.Is to match inner error")
	}

	if err.Unwrap() != inner {
		t.Error("expected Unwrap to return inner error")
	}
}

func TestGitError(t *testing.T) {
	inner := errors.New("fatal: not a git repository")
	err := &GitError{Operation: "status", Err: inner}

	// Test Error()
	expected := "git status failed: fatal: not a git repository"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}

	// Test Unwrap()
	if err.Unwrap() != inner {
		t.Error("expected Unwrap to return inner error")
	}

	// Test Is() matches ErrGitOperation
	if !errors.Is(err, ErrGitOperation) {
		t.Error("expected errors.Is to match ErrGitOperation")
	}

	// Test Is() also matches inner error via Unwrap
	if !errors.Is(err, inner) {
		t.Error("expected errors.Is to match inner error")
	}
}

func TestFileError(t *testing.T) {
	inner := errors.New("permission denied")
	err := &FileError{Op: "read", Path: "/etc/shadow", Err: inner}

	// Test Error()
	expected := `read "/etc/shadow": permission denied`
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}

	// Test Unwrap()
	if err.Unwrap() != inner {
		t.Error("expected Unwrap to return inner error")
	}

	// Test errors.Is matches inner error
	if !errors.Is(err, inner) {
		t.Error("expected errors.Is to match inner error")
	}
}

func TestExtensionError(t *testing.T) {
	inner := errors.New("script failed")

	t.Run("with name", func(t *testing.T) {
		err := &ExtensionError{Name: "my-ext", Op: "execute", Err: inner}

		expected := `extension "my-ext" execute: script failed`
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}

		if err.Unwrap() != inner {
			t.Error("expected Unwrap to return inner error")
		}

		if !errors.Is(err, ErrExtension) {
			t.Error("expected errors.Is to match ErrExtension")
		}
	})

	t.Run("without name", func(t *testing.T) {
		err := &ExtensionError{Name: "", Op: "load", Err: inner}

		expected := "extension load: script failed"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})
}

func TestWrapGit(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		result := WrapGit("clone", nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("non-nil error wraps correctly", func(t *testing.T) {
		inner := errors.New("remote not found")
		result := WrapGit("fetch", inner)

		var gitErr *GitError
		if !errors.As(result, &gitErr) {
			t.Fatal("expected result to be *GitError")
		}

		if gitErr.Operation != "fetch" {
			t.Errorf("expected operation %q, got %q", "fetch", gitErr.Operation)
		}

		if gitErr.Err != inner {
			t.Error("expected inner error to be preserved")
		}
	})
}

func TestWrapFile(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		result := WrapFile("write", "/tmp/test", nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("non-nil error wraps correctly", func(t *testing.T) {
		inner := errors.New("disk full")
		result := WrapFile("write", "/tmp/test", inner)

		var fileErr *FileError
		if !errors.As(result, &fileErr) {
			t.Fatal("expected result to be *FileError")
		}

		if fileErr.Op != "write" {
			t.Errorf("expected op %q, got %q", "write", fileErr.Op)
		}

		if fileErr.Path != "/tmp/test" {
			t.Errorf("expected path %q, got %q", "/tmp/test", fileErr.Path)
		}

		if fileErr.Err != inner {
			t.Error("expected inner error to be preserved")
		}
	})
}

func TestWrapExtension(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		result := WrapExtension("my-ext", "run", nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("non-nil error wraps correctly", func(t *testing.T) {
		inner := errors.New("timeout")
		result := WrapExtension("my-ext", "run", inner)

		var extErr *ExtensionError
		if !errors.As(result, &extErr) {
			t.Fatal("expected result to be *ExtensionError")
		}

		if extErr.Name != "my-ext" {
			t.Errorf("expected name %q, got %q", "my-ext", extErr.Name)
		}

		if extErr.Op != "run" {
			t.Errorf("expected op %q, got %q", "run", extErr.Op)
		}

		if extErr.Err != inner {
			t.Error("expected inner error to be preserved")
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{ErrNotFound, "not found"},
		{ErrInvalidInput, "invalid input"},
		{ErrPermissionDenied, "permission denied"},
		{ErrAlreadyExists, "already exists"},
		{ErrTimeout, "operation timed out"},
		{ErrCanceled, "operation canceled"},
		{ErrGitOperation, "git operation failed"},
		{ErrExtension, "extension error"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.err.Error())
			}

			// Verify sentinel errors work with errors.Is
			wrapped := fmt.Errorf("context: %w", tt.err)
			if !errors.Is(wrapped, tt.err) {
				t.Error("expected errors.Is to match wrapped sentinel error")
			}
		})
	}
}
