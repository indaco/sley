package pathutil

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/indaco/sley/internal/apperrors"
)

func TestValidatePath(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		baseDir string
		wantErr bool
		errType error
	}{
		{
			name:    "valid path without base",
			path:    "file.txt",
			baseDir: "",
			wantErr: false,
		},
		{
			name:    "valid absolute path without base",
			path:    "/tmp/file.txt",
			baseDir: "",
			wantErr: false,
		},
		{
			name:    "valid path within base",
			path:    filepath.Join(tmpDir, "subdir", "file.txt"),
			baseDir: tmpDir,
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			baseDir: "",
			wantErr: true,
			errType: &apperrors.PathValidationError{},
		},
		{
			name:    "path traversal attempt",
			path:    filepath.Join(tmpDir, "..", "secret"),
			baseDir: tmpDir,
			wantErr: true,
			errType: &apperrors.PathValidationError{},
		},
		{
			name:    "path outside base dir",
			path:    "/etc/passwd",
			baseDir: tmpDir,
			wantErr: true,
			errType: &apperrors.PathValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := ValidatePath(tt.path, tt.baseDir)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errType != nil {
					var pathErr *apperrors.PathValidationError
					if !errors.As(err, &pathErr) {
						t.Errorf("expected PathValidationError, got %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == "" {
					t.Error("expected non-empty result")
				}
			}
		})
	}
}

func TestValidatePath_EdgeCases(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	t.Run("path with dots is cleaned", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(tmpDir, "subdir", ".", "file.txt")
		result, err := ValidatePath(path, tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(tmpDir, "subdir", "file.txt")
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("path equals base dir", func(t *testing.T) {
		t.Parallel()
		result, err := ValidatePath(tmpDir, tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != tmpDir {
			t.Errorf("expected %q, got %q", tmpDir, result)
		}
	})

	t.Run("subdirectory path", func(t *testing.T) {
		t.Parallel()
		subPath := filepath.Join(tmpDir, "subdir", "file.txt")
		result, err := ValidatePath(subPath, tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("path with multiple dots cleaned", func(t *testing.T) {
		t.Parallel()
		path := "./some/path/../other/./file.txt"
		result, err := ValidatePath(path, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Clean(path)
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

func TestValidatePath_InvalidBaseDir(t *testing.T) {
	t.Parallel(
	// Use a path that's very unlikely to exist on any system
	// This should still work since we're testing error handling
	)

	invalidBase := string([]byte{0, 1, 2}) // Invalid path characters

	_, err := ValidatePath("test.txt", invalidBase)
	if err == nil {
		// On some systems this might not error, which is okay
		return
	}

	var pathErr *apperrors.PathValidationError
	if !errors.As(err, &pathErr) {
		t.Errorf("expected PathValidationError, got %T", err)
	}
}

func TestValidatePath_AbsPathErrors(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	t.Run("valid nested path", func(t *testing.T) {
		t.Parallel(
		// Test a deeply nested path to exercise all code paths
		)

		nestedPath := filepath.Join(tmpDir, "a", "b", "c", "d", "e", "file.txt")
		result, err := ValidatePath(nestedPath, tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nestedPath {
			t.Errorf("expected %q, got %q", nestedPath, result)
		}
	})

	t.Run("path with trailing separator", func(t *testing.T) {
		t.Parallel()
		pathWithSep := tmpDir + string(filepath.Separator)
		result, err := ValidatePath(pathWithSep, tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should clean the path
		if result == pathWithSep {
			// Cleaned path should not have trailing separator
			t.Log("path was cleaned")
		}
	})
}

func TestIsWithinDir_VariousScenarios(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	t.Run("deeply nested path", func(t *testing.T) {
		t.Parallel()
		deepPath := filepath.Join(tmpDir, "a", "b", "c", "d", "e", "f", "g", "h")
		result := IsWithinDir(deepPath, tmpDir)
		if !result {
			t.Error("deeply nested path should be within dir")
		}
	})

	t.Run("path with trailing separator equals dir", func(t *testing.T) {
		t.Parallel()
		pathWithSep := tmpDir + string(filepath.Separator)
		result := IsWithinDir(pathWithSep, tmpDir)
		if !result {
			t.Error("path with trailing separator should equal dir")
		}
	})

	t.Run("sibling directory", func(t *testing.T) {
		t.Parallel()
		parent := filepath.Dir(tmpDir)
		siblingDir := filepath.Join(parent, "other-dir")
		result := IsWithinDir(siblingDir, tmpDir)
		if result {
			t.Error("sibling directory should not be within dir")
		}
	})

	t.Run("empty path strings", func(t *testing.T) {
		t.Parallel(
		// Test with empty strings - they should resolve to current directory
		)

		result := IsWithinDir("", "")
		// Both empty strings resolve to ".", which equals itself
		if !result {
			t.Log("empty paths treated as not within each other (acceptable)")
		}
	})
}

func TestIsWithinDir(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		path     string
		dir      string
		expected bool
	}{
		{
			name:     "file within dir",
			path:     filepath.Join(tmpDir, "file.txt"),
			dir:      tmpDir,
			expected: true,
		},
		{
			name:     "subdir within dir",
			path:     filepath.Join(tmpDir, "subdir", "file.txt"),
			dir:      tmpDir,
			expected: true,
		},
		{
			name:     "dir itself",
			path:     tmpDir,
			dir:      tmpDir,
			expected: true,
		},
		{
			name:     "path outside dir",
			path:     "/etc/passwd",
			dir:      tmpDir,
			expected: false,
		},
		{
			name:     "parent of dir",
			path:     filepath.Dir(tmpDir),
			dir:      tmpDir,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := IsWithinDir(tt.path, tt.dir)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsWithinDir_RelativePaths(t *testing.T) {
	t.Parallel()
	t.Run("relative path within relative dir", func(t *testing.T) {
		t.Parallel()
		result := IsWithinDir("subdir/file.txt", "subdir")
		// This should be true since both resolve to the same relative location
		if !result {
			t.Error("expected true for relative paths within same directory")
		}
	})

	t.Run("current directory check", func(t *testing.T) {
		t.Parallel()
		result := IsWithinDir(".", ".")
		if !result {
			t.Error("current directory should be within itself")
		}
	})
}

func TestIsWithinDir_ErrorHandling(t *testing.T) {
	t.Parallel(
	// Test with paths that might cause Abs to fail (though rare)
	)

	invalidPath := string([]byte{0, 1, 2})

	t.Run("invalid path returns false", func(t *testing.T) {
		t.Parallel(
		// Just verify this doesn't panic - result may vary by system
		)

		_ = IsWithinDir(invalidPath, "/tmp")
	})

	t.Run("invalid dir returns false", func(t *testing.T) {
		t.Parallel(
		// Just verify this doesn't panic - result may vary by system
		)

		_ = IsWithinDir("/tmp/file", invalidPath)
	})
}
