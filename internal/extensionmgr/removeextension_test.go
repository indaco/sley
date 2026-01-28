package extensionmgr

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/config"
)

// ---------------------------------------------------------------------------
// RemoveExtension tests
// ---------------------------------------------------------------------------

func TestRemoveExtension_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	initial := []byte(`path: .version
extensions:
  - name: ext-one
    path: .sley-extensions/ext-one
    enabled: true
  - name: ext-two
    path: .sley-extensions/ext-two
    enabled: true
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	if err := updater.RemoveExtension(configPath, "ext-one"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var parsed config.Config
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal updated config: %v", err)
	}

	if len(parsed.Extensions) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(parsed.Extensions))
	}
	if parsed.Extensions[0].Name != "ext-two" {
		t.Errorf("expected remaining extension to be ext-two, got %q", parsed.Extensions[0].Name)
	}
}

func TestRemoveExtension_LastExtension_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	initial := []byte(`path: .version
extensions:
  - name: only-ext
    path: .sley-extensions/only-ext
    enabled: true
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	if err := updater.RemoveExtension(configPath, "only-ext"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	// When all extensions are removed the section should be "extensions: []"
	if !strings.Contains(content, "extensions: []") {
		t.Errorf("expected extensions section to become empty list, got:\n%s", content)
	}

	var parsed config.Config
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal updated config: %v", err)
	}
	if len(parsed.Extensions) != 0 {
		t.Fatalf("expected 0 extensions, got %d", len(parsed.Extensions))
	}
}

func TestRemoveExtension_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	initial := []byte("path: .version\nextensions: []\n")
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	err := updater.RemoveExtension(configPath, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent extension, got nil")
	}
	if !strings.Contains(err.Error(), "not found in configuration") {
		t.Errorf("expected error to contain 'not found in configuration', got: %v", err)
	}
}

func TestRemoveExtension_ReadFileError(t *testing.T) {
	invalidPath := filepath.Join(t.TempDir(), "nonexistent.yaml")

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	err := updater.RemoveExtension(invalidPath, "test")
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected file not found error, got: %v", err)
	}
}

func TestRemoveExtension_UnmarshalError(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".sley.yaml")

	badYAML := []byte(": invalid yaml")
	if err := os.WriteFile(configPath, badYAML, 0644); err != nil {
		t.Fatal(err)
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	err := updater.RemoveExtension(configPath, "test")
	if err == nil || !strings.Contains(err.Error(), "unexpected key name") {
		t.Fatalf("expected YAML unmarshal error, got: %v", err)
	}
}

func TestRemoveExtension_MarshalError(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".sley.yaml")

	initial := []byte(`extensions:
  - name: test-ext
    path: .sley-extensions/test-ext
    enabled: true
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	mockMarshaler := &MockYAMLMarshaler{
		MarshalFunc: func(v any) ([]byte, error) {
			return nil, errors.New("forced marshal failure")
		},
	}
	updater := NewDefaultConfigUpdater(mockMarshaler)

	err := updater.RemoveExtension(configPath, "test-ext")
	if err == nil || !strings.Contains(err.Error(), "forced marshal failure") {
		t.Fatalf("expected marshal error, got: %v", err)
	}
}

func TestRemoveExtension_WriteFileError(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".sley.yaml")

	initial := []byte(`extensions:
  - name: test-ext
    path: .sley-extensions/test-ext
    enabled: true
`)
	if err := os.WriteFile(configPath, initial, 0444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(configPath, 0644)
	})

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	err := updater.RemoveExtension(configPath, "test-ext")
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected write error, got: %v", err)
	}
}

func TestRemoveExtension_PreservesComments(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	initial := []byte(`# sley configuration file
# Generated by sley init
path: .version # path to the version file

# Plugin settings
plugins:
  commit-parser: true

# Installed extensions
extensions:
  - name: ext-keep
    path: .sley-extensions/ext-keep
    enabled: true
  - name: ext-remove
    path: .sley-extensions/ext-remove
    enabled: true

# Pre-release hooks
pre-release-hooks:
  - changelog:
      command: git-chglog -o CHANGELOG.md
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	if err := updater.RemoveExtension(configPath, "ext-remove"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	// Verify all comments are preserved
	expectedComments := []string{
		"# sley configuration file",
		"# Generated by sley init",
		"# path to the version file",
		"# Plugin settings",
		"# Installed extensions",
		"# Pre-release hooks",
	}
	for _, comment := range expectedComments {
		if !strings.Contains(content, comment) {
			t.Errorf("expected comment %q to be preserved, but it is missing.\nActual content:\n%s", comment, content)
		}
	}

	// Verify the extension was removed correctly by parsing the YAML
	var parsed config.Config
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal updated config: %v", err)
	}

	if len(parsed.Extensions) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(parsed.Extensions))
	}

	got := parsed.Extensions[0]
	if got.Name != "ext-keep" || got.Path != ".sley-extensions/ext-keep" || !got.Enabled {
		t.Errorf("unexpected remaining extension entry: %+v", got)
	}

	// Verify the removed extension is gone
	if strings.Contains(content, "ext-remove") {
		t.Errorf("expected ext-remove to be removed, but it is still present.\nActual content:\n%s", content)
	}

	// Verify other sections remain intact
	if parsed.Path != ".version" {
		t.Errorf("expected path to be .version, got %q", parsed.Path)
	}
	if len(parsed.PreReleaseHooks) != 1 {
		t.Errorf("expected 1 pre-release hook, got %d", len(parsed.PreReleaseHooks))
	}
}

func TestRemoveExtension_PreservesComments_EmptyResult(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	initial := []byte(`# sley configuration file
path: .version # path to the version file

# Installed extensions
extensions:
  - name: only-ext
    path: .sley-extensions/only-ext
    enabled: true

# Pre-release hooks
pre-release-hooks:
  - changelog:
      command: git-chglog -o CHANGELOG.md
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	if err := updater.RemoveExtension(configPath, "only-ext"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	// Verify comments are preserved
	expectedComments := []string{
		"# sley configuration file",
		"# path to the version file",
		"# Installed extensions",
		"# Pre-release hooks",
	}
	for _, comment := range expectedComments {
		if !strings.Contains(content, comment) {
			t.Errorf("expected comment %q to be preserved, but it is missing.\nActual content:\n%s", comment, content)
		}
	}

	// Verify extensions section is now empty
	if !strings.Contains(content, "extensions: []") {
		t.Errorf("expected extensions section to be empty list.\nActual content:\n%s", content)
	}

	// Verify other sections remain intact
	var parsed config.Config
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal updated config: %v", err)
	}
	if parsed.Path != ".version" {
		t.Errorf("expected path to be .version, got %q", parsed.Path)
	}
	if len(parsed.PreReleaseHooks) != 1 {
		t.Errorf("expected 1 pre-release hook, got %d", len(parsed.PreReleaseHooks))
	}
}
