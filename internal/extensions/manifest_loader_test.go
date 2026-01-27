package extensions

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeExtensionYAML(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "extension.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write extension.yaml: %v", err)
	}
	return path
}

func TestLoadExtensionManifest_Valid(t *testing.T) {
	dir := t.TempDir()
	content := `
name: test
version: 0.1.0
description: test plugin
author: me
repository: https://example.com/repo
entry: actions.json
`
	writeExtensionYAML(t, dir, content)

	m, err := LoadExtensionManifestFn(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.Name != "test" {
		t.Errorf("expected name 'test', got %q", m.Name)
	}
}

func TestLoadExtensionManifest_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadExtensionManifestFn(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Should return ManifestNotFoundError
	var notFoundErr *ManifestNotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected ManifestNotFoundError, got %T: %v", err, err)
	}

	// Error should contain helpful message
	if !strings.Contains(err.Error(), "extension manifest not found") {
		t.Errorf("expected error to contain 'extension manifest not found', got: %v", err)
	}
}

func TestLoadExtensionManifest_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	content := ": this is not valid yaml"
	writeExtensionYAML(t, dir, content)

	_, err := LoadExtensionManifestFn(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Should return ManifestParseError
	var parseErr *ManifestParseError
	if !errors.As(err, &parseErr) {
		t.Errorf("expected ManifestParseError, got %T: %v", err, err)
	}

	// Error should contain parse error message
	if !strings.Contains(err.Error(), "failed to parse manifest") {
		t.Errorf("expected error to contain 'failed to parse manifest', got: %v", err)
	}
}

func TestLoadExtensionManifest_InvalidManifest(t *testing.T) {
	dir := t.TempDir()
	content := `
name: ""
version: ""
description: ""
author: ""
repository: ""
entry: ""
`
	writeExtensionYAML(t, dir, content)

	_, err := LoadExtensionManifestFn(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Should return ManifestValidationError
	var valErr *ManifestValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected ManifestValidationError, got %T: %v", err, err)
	} else if len(valErr.MissingFields) != 6 {
		// Should have all 6 missing fields
		t.Errorf("expected 6 missing fields, got %d: %v", len(valErr.MissingFields), valErr.MissingFields)
	}

	// Error should contain missing fields message
	if !strings.Contains(err.Error(), "missing required fields") {
		t.Errorf("expected error to contain 'missing required fields', got: %v", err)
	}
}
