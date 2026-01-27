package extensions

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestExtensionManifest_Validate(t *testing.T) {
	base := ExtensionManifest{
		Name:        "commit-parser",
		Version:     "0.1.0",
		Description: "Parses conventional commits",
		Author:      "indaco",
		Repository:  "https://github.com/indaco/sley-commit-parser",
		Entry:       "github.com/indaco/sley-commit/parser",
	}

	tests := []struct {
		field    string
		modify   func(m *ExtensionManifest)
		expected string
	}{
		{"missing name", func(m *ExtensionManifest) { m.Name = "" }, "name"},
		{"missing version", func(m *ExtensionManifest) { m.Version = "" }, "version"},
		{"missing description", func(m *ExtensionManifest) { m.Description = "" }, "description"},
		{"missing author", func(m *ExtensionManifest) { m.Author = "" }, "author"},
		{"missing repository", func(m *ExtensionManifest) { m.Repository = "" }, "repository"},
		{"missing entry", func(m *ExtensionManifest) { m.Entry = "" }, "entry"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			m := base
			tt.modify(&m)

			err := m.ValidateManifest()
			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}

			// Check if it's our custom error type
			var valErr *ManifestValidationError
			if errors.As(err, &valErr) {
				found := slices.Contains(valErr.MissingFields, tt.expected)
				if !found {
					t.Errorf("expected missing field %q, got %v", tt.expected, valErr.MissingFields)
				}
			} else if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("expected error to contain %q, got %v", tt.expected, err)
			}
		})
	}

	t.Run("valid manifest", func(t *testing.T) {
		err := base.ValidateManifest()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

/* ------------------------------------------------------------------------- */
/* ADDITIONAL TABLE-DRIVEN TESTS FOR MANIFEST VALIDATION                   */
/* ------------------------------------------------------------------------- */

// TestExtensionManifest_ValidateManifest_TableDriven tests comprehensive validation scenarios
func TestExtensionManifest_ValidateManifest_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		manifest    ExtensionManifest
		wantErr     bool
		wantErrText string
	}{
		{
			name: "complete valid manifest",
			manifest: ExtensionManifest{
				Name:        "test-extension",
				Version:     "1.0.0",
				Description: "A test extension",
				Author:      "Test Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
				Hooks:       []string{"pre-bump", "post-bump"},
			},
			wantErr: false,
		},
		{
			name: "valid manifest without hooks",
			manifest: ExtensionManifest{
				Name:        "simple-ext",
				Version:     "0.1.0",
				Description: "Simple extension",
				Author:      "Developer",
				Repository:  "https://gitlab.com/dev/simple",
				Entry:       "run.sh",
				Hooks:       nil,
			},
			wantErr: false,
		},
		{
			name: "valid manifest with empty hooks array",
			manifest: ExtensionManifest{
				Name:        "no-hooks",
				Version:     "2.0.0",
				Description: "Extension without hooks",
				Author:      "Author",
				Repository:  "https://github.com/author/nohooks",
				Entry:       "script.py",
				Hooks:       []string{},
			},
			wantErr: false,
		},
		{
			name: "missing name only",
			manifest: ExtensionManifest{
				Name:        "",
				Version:     "1.0.0",
				Description: "Missing name",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			wantErr:     true,
			wantErrText: "name",
		},
		{
			name: "missing version only",
			manifest: ExtensionManifest{
				Name:        "test",
				Version:     "",
				Description: "Missing version",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			wantErr:     true,
			wantErrText: "version",
		},
		{
			name: "missing description only",
			manifest: ExtensionManifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			wantErr:     true,
			wantErrText: "description",
		},
		{
			name: "missing author only",
			manifest: ExtensionManifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Author:      "",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			wantErr:     true,
			wantErrText: "author",
		},
		{
			name: "missing repository only",
			manifest: ExtensionManifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Author:      "Author",
				Repository:  "",
				Entry:       "hook.sh",
			},
			wantErr:     true,
			wantErrText: "repository",
		},
		{
			name: "missing entry only",
			manifest: ExtensionManifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "",
			},
			wantErr:     true,
			wantErrText: "entry",
		},
		{
			name: "all fields empty",
			manifest: ExtensionManifest{
				Name:        "",
				Version:     "",
				Description: "",
				Author:      "",
				Repository:  "",
				Entry:       "",
			},
			wantErr:     true,
			wantErrText: "name", // Should contain all missing fields
		},
		{
			name: "whitespace only fields",
			manifest: ExtensionManifest{
				Name:        "   ",
				Version:     "1.0.0",
				Description: "Test",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			wantErr: false, // Currently doesn't trim/validate whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.ValidateManifest()

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrText != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrText) {
					t.Errorf("expected error containing %q, got: %v", tt.wantErrText, err)
				}
			}
		})
	}
}

// TestExtensionManifest_Fields tests individual field properties
func TestExtensionManifest_Fields(t *testing.T) {
	tests := []struct {
		name     string
		manifest ExtensionManifest
		checkFn  func(t *testing.T, m ExtensionManifest)
	}{
		{
			name: "hooks with all valid types",
			manifest: ExtensionManifest{
				Name:        "multi-hook",
				Version:     "1.0.0",
				Description: "Multiple hooks",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
				Hooks:       []string{"pre-bump", "post-bump", "pre-release", "validate"},
			},
			checkFn: func(t *testing.T, m ExtensionManifest) {
				if len(m.Hooks) != 4 {
					t.Errorf("expected 4 hooks, got %d", len(m.Hooks))
				}
			},
		},
		{
			name: "single hook",
			manifest: ExtensionManifest{
				Name:        "single-hook",
				Version:     "1.0.0",
				Description: "Single hook",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
				Hooks:       []string{"pre-bump"},
			},
			checkFn: func(t *testing.T, m ExtensionManifest) {
				if len(m.Hooks) != 1 {
					t.Errorf("expected 1 hook, got %d", len(m.Hooks))
				}
				if m.Hooks[0] != "pre-bump" {
					t.Errorf("expected hook 'pre-bump', got %q", m.Hooks[0])
				}
			},
		},
		{
			name: "various version formats",
			manifest: ExtensionManifest{
				Name:        "version-test",
				Version:     "1.2.3-alpha.1+build.456",
				Description: "Version format test",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			checkFn: func(t *testing.T, m ExtensionManifest) {
				if m.Version != "1.2.3-alpha.1+build.456" {
					t.Errorf("version not preserved correctly: %s", m.Version)
				}
			},
		},
		{
			name: "different entry formats",
			manifest: ExtensionManifest{
				Name:        "entry-test",
				Version:     "1.0.0",
				Description: "Entry format test",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "./scripts/main.py",
			},
			checkFn: func(t *testing.T, m ExtensionManifest) {
				if m.Entry != "./scripts/main.py" {
					t.Errorf("entry path not preserved: %s", m.Entry)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.ValidateManifest()
			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
			tt.checkFn(t, tt.manifest)
		})
	}
}

/* ------------------------------------------------------------------------- */
/* TESTS FOR CUSTOM MANIFEST ERROR TYPES                                   */
/* ------------------------------------------------------------------------- */

// TestManifestNotFoundError tests the ManifestNotFoundError type
func TestManifestNotFoundError(t *testing.T) {
	err := &ManifestNotFoundError{
		Path: "/path/to/extension.yaml",
		Dir:  "/path/to",
	}

	// Test Error() method
	if !strings.Contains(err.Error(), "extension manifest not found") {
		t.Errorf("Error() should contain \"extension manifest not found\", got: %s", err.Error())
	}

	if !strings.Contains(err.Error(), "/path/to/extension.yaml") {
		t.Errorf("Error() should contain path, got: %s", err.Error())
	}

	// Test Suggestion() method
	suggestion := err.Suggestion()
	expectedParts := []string{
		"Extension manifest not found",
		"name:",
		"version:",
		"description:",
		"author:",
		"repository:",
		"entry:",
		"hooks:",
		"Documentation:",
	}

	for _, part := range expectedParts {
		if !strings.Contains(suggestion, part) {
			t.Errorf("Suggestion() should contain %q, got: %s", part, suggestion)
		}
	}
}

// TestManifestParseError tests the ManifestParseError type
func TestManifestParseError(t *testing.T) {
	originalErr := errors.New("yaml: line 5: mapping values are not allowed in this context")
	err := &ManifestParseError{
		Path: "/path/to/extension.yaml",
		Err:  originalErr,
	}

	// Test Error() method
	errMsg := err.Error()
	if !strings.Contains(errMsg, "failed to parse manifest") {
		t.Errorf("Error() should contain \"failed to parse manifest\", got: %s", errMsg)
	}

	if !strings.Contains(errMsg, "/path/to/extension.yaml") {
		t.Errorf("Error() should contain path, got: %s", errMsg)
	}

	// Test Unwrap() method
	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() should return original error, got: %v", unwrapped)
	}
}

// TestManifestValidationError tests the ManifestValidationError type
func TestManifestValidationError(t *testing.T) {
	err := &ManifestValidationError{
		Path:          "/path/to/extension.yaml",
		MissingFields: []string{"name", "version", "entry"},
	}

	// Test Error() method
	errMsg := err.Error()
	if !strings.Contains(errMsg, "invalid manifest") {
		t.Errorf("Error() should contain \"invalid manifest\", got: %s", errMsg)
	}

	if !strings.Contains(errMsg, "name") || !strings.Contains(errMsg, "version") || !strings.Contains(errMsg, "entry") {
		t.Errorf("Error() should contain all missing fields, got: %s", errMsg)
	}

	// Test Suggestion() method
	suggestion := err.Suggestion()
	expectedParts := []string{
		"Manifest validation failed",
		"Missing required fields:",
		"• name",
		"• version",
		"• entry",
		"Documentation:",
	}

	for _, part := range expectedParts {
		if !strings.Contains(suggestion, part) {
			t.Errorf("Suggestion() should contain %q, got: %s", part, suggestion)
		}
	}
}

// TestManifestValidation_MultipleErrors tests that all missing fields are reported
func TestManifestValidation_MultipleErrors(t *testing.T) {
	manifest := ExtensionManifest{
		Name:        "",
		Version:     "",
		Description: "Has description",
		Author:      "",
		Repository:  "https://github.com/test/repo",
		Entry:       "",
	}

	err := manifest.ValidateManifest()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var valErr *ManifestValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ManifestValidationError, got %T", err)
	}

	// Should have 4 missing fields: name, version, author, entry
	expectedMissing := map[string]bool{
		"name":    true,
		"version": true,
		"author":  true,
		"entry":   true,
	}

	if len(valErr.MissingFields) != len(expectedMissing) {
		t.Errorf("expected %d missing fields, got %d: %v",
			len(expectedMissing), len(valErr.MissingFields), valErr.MissingFields)
	}

	for _, field := range valErr.MissingFields {
		if !expectedMissing[field] {
			t.Errorf("unexpected missing field: %s", field)
		}
	}
}
