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

func TestAddExtensionToConfig_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	initial := []byte("path: .version\nextensions: []\n")
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	extension := config.ExtensionConfig{
		Name:    "commit-parser",
		Path:    ".sley-extensions/commit-parser",
		Enabled: true,
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	if err := updater.AddExtension(configPath, extension); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	// Re-read and verify
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var parsed config.Config
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal updated config: %v", err)
	}

	if len(parsed.Extensions) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(parsed.Extensions))
	}

	got := parsed.Extensions[0]
	if got.Name != extension.Name || got.Path != extension.Path || !got.Enabled {
		t.Errorf("unexpected plugin entry: %+v", got)
	}
}

func TestAddExtensionToConfig_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	configPath := filepath.Join(tmpDir, ".sley.yaml")
	// Initial config with one plugin
	initial := []byte(`
path: .version
extensions:
  - name: test-extension
    path: .sley-extensions/test-extension
    enabled: true
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	extension := config.ExtensionConfig{
		Name:    "test-extension",
		Path:    ".sley-extensions/test-extension",
		Enabled: true,
	}

	// First registration (extension already exists in config, should error)
	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	err = updater.AddExtension(configPath, extension)
	if err == nil {
		t.Fatal("expected error for duplicate extension, got nil")
	}

	// Error should indicate extension already registered
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("expected error to contain 'already registered', got: %v", err)
	}

	// Ensure the config file still has only one plugin
	cfg, err := config.LoadConfigFn()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}
	if len(cfg.Extensions) != 1 {
		t.Fatalf("expected 1 extension in config, got: %d", len(cfg.Extensions))
	}
}

func TestAddExtensionToConfig_ReadFileError(t *testing.T) {
	invalidPath := filepath.Join(t.TempDir(), "nonexistent.yaml")

	extension := config.ExtensionConfig{
		Name:    "test",
		Path:    "some/path",
		Enabled: true,
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	err := updater.AddExtension(invalidPath, extension)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected file not found error, got: %v", err)
	}
}

func TestAddExtensionToConfig_UnmarshalError(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".sley.yaml")

	badYAML := []byte(": invalid yaml")
	if err := os.WriteFile(configPath, badYAML, 0644); err != nil {
		t.Fatal(err)
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	err := updater.AddExtension(configPath, config.ExtensionConfig{
		Name:    "test",
		Path:    "some/path",
		Enabled: true,
	})

	if err == nil || !strings.Contains(err.Error(), "unexpected key name") {
		t.Fatalf("expected YAML unmarshal error, got: %v", err)
	}
}

func TestAddExtensionToConfig_MarshalError(t *testing.T) {
	// Create a temporary file with a valid config
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".sley.yaml")
	initial := []byte(`path: .version`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	// Create updater with mock marshaler that fails
	mockMarshaler := &MockYAMLMarshaler{
		MarshalFunc: func(v any) ([]byte, error) {
			return nil, errors.New("forced marshal failure")
		},
	}
	updater := NewDefaultConfigUpdater(mockMarshaler)

	err := updater.AddExtension(configPath, config.ExtensionConfig{
		Name:    "fail-marshaling",
		Path:    ".sley-extensions/fail",
		Enabled: true,
	})

	if err == nil || !strings.Contains(err.Error(), "forced marshal failure") {
		t.Fatalf("expected marshal error, got: %v", err)
	}
}

func TestAddExtensionToConfig_WriteFileError(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".sley.yaml")

	initial := []byte("path: .version\nextensions: []\n")
	if err := os.WriteFile(configPath, initial, 0444); err != nil {
		t.Fatal(err)
	}
	// Ensure cleanup restores perms so t.TempDir can delete
	t.Cleanup(func() {
		_ = os.Chmod(configPath, 0644)
	})

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	err := updater.AddExtension(configPath, config.ExtensionConfig{
		Name:    "test",
		Path:    "some/path",
		Enabled: true,
	})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected write error, got: %v", err)
	}
}

func TestAddExtensionToConfig_ProperYAMLIndentation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	initial := []byte("path: .version\nextensions: []\n")
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	// Add first extension
	ext1 := config.ExtensionConfig{
		Name:    "extension-one",
		Path:    ".sley-extensions/extension-one",
		Enabled: true,
	}
	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	if err := updater.AddExtension(configPath, ext1); err != nil {
		t.Fatalf("failed to add first extension: %v", err)
	}

	// Add second extension
	ext2 := config.ExtensionConfig{
		Name:    "extension-two",
		Path:    ".sley-extensions/extension-two",
		Enabled: true,
	}
	if err := updater.AddExtension(configPath, ext2); err != nil {
		t.Fatalf("failed to add second extension: %v", err)
	}

	// Read the raw YAML content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	yamlContent := string(data)

	// Verify proper indentation: list items should be indented with 2 spaces
	expectedIndentation := []string{
		"extensions:",
		"  - name: extension-one",
		"    path: .sley-extensions/extension-one",
		"    enabled: true",
		"  - name: extension-two",
		"    path: .sley-extensions/extension-two",
		"    enabled: true",
	}

	for _, expected := range expectedIndentation {
		if !strings.Contains(yamlContent, expected) {
			t.Errorf("expected YAML to contain %q, but it doesn't.\nActual YAML:\n%s", expected, yamlContent)
		}
	}

	// Verify it doesn't have improper indentation (no indent for list items)
	improperPatterns := []string{
		"extensions:\n- name:", // List item directly after extensions: with no indent
	}

	for _, improper := range improperPatterns {
		if strings.Contains(yamlContent, improper) {
			t.Errorf("YAML should not contain improper indentation pattern %q.\nActual YAML:\n%s", improper, yamlContent)
		}
	}
}

func TestAddExtensionToConfig_PreservesComments(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	initial := []byte(`# sley configuration file
# Generated by sley init
path: .version # path to the version file

# Plugin settings
plugins:
  commit-parser: true

# Installed extensions
extensions: []

# Pre-release hooks
pre-release-hooks:
  - changelog:
      command: git-chglog -o CHANGELOG.md
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	extension := config.ExtensionConfig{
		Name:    "my-extension",
		Path:    ".sley-extensions/my-extension",
		Enabled: true,
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	if err := updater.AddExtension(configPath, extension); err != nil {
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

	// Verify the extension was added correctly by parsing the YAML
	var parsed config.Config
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal updated config: %v", err)
	}

	if len(parsed.Extensions) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(parsed.Extensions))
	}

	got := parsed.Extensions[0]
	if got.Name != "my-extension" || got.Path != ".sley-extensions/my-extension" || !got.Enabled {
		t.Errorf("unexpected extension entry: %+v", got)
	}

	// Verify other sections remain intact
	if parsed.Path != ".version" {
		t.Errorf("expected path to be .version, got %q", parsed.Path)
	}
	if len(parsed.PreReleaseHooks) != 1 {
		t.Errorf("expected 1 pre-release hook, got %d", len(parsed.PreReleaseHooks))
	}
}

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

// ---------------------------------------------------------------------------
// SetExtensionEnabled tests
// ---------------------------------------------------------------------------

func TestSetExtensionEnabled_Disable(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	initial := []byte(`path: .version
extensions:
  - name: my-ext
    path: .sley-extensions/my-ext
    enabled: true
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	if err := updater.SetExtensionEnabled(configPath, "my-ext", false); err != nil {
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
	if parsed.Extensions[0].Enabled {
		t.Errorf("expected extension to be disabled, but it is enabled")
	}

	// Verify the raw YAML contains enabled: false
	if !strings.Contains(string(data), "enabled: false") {
		t.Errorf("expected YAML to contain 'enabled: false', got:\n%s", string(data))
	}
}

func TestSetExtensionEnabled_Enable(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	initial := []byte(`path: .version
extensions:
  - name: my-ext
    path: .sley-extensions/my-ext
    enabled: false
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	if err := updater.SetExtensionEnabled(configPath, "my-ext", true); err != nil {
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
	if !parsed.Extensions[0].Enabled {
		t.Errorf("expected extension to be enabled, but it is disabled")
	}

	// Verify the raw YAML contains enabled: true
	if !strings.Contains(string(data), "enabled: true") {
		t.Errorf("expected YAML to contain 'enabled: true', got:\n%s", string(data))
	}
}

func TestSetExtensionEnabled_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".sley.yaml")

	initial := []byte("path: .version\nextensions: []\n")
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	err := updater.SetExtensionEnabled(configPath, "nonexistent", true)
	if err == nil {
		t.Fatal("expected error for nonexistent extension, got nil")
	}
	if !strings.Contains(err.Error(), "not found in configuration") {
		t.Errorf("expected error to contain 'not found in configuration', got: %v", err)
	}
}

func TestSetExtensionEnabled_PreservesComments(t *testing.T) {
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
  - name: ext-one
    path: .sley-extensions/ext-one
    enabled: true
  - name: ext-two
    path: .sley-extensions/ext-two
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
	if err := updater.SetExtensionEnabled(configPath, "ext-one", false); err != nil {
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

	// Verify the extension was toggled correctly by parsing the YAML
	var parsed config.Config
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal updated config: %v", err)
	}

	if len(parsed.Extensions) != 2 {
		t.Fatalf("expected 2 extensions, got %d", len(parsed.Extensions))
	}

	// ext-one should be disabled
	if parsed.Extensions[0].Name != "ext-one" || parsed.Extensions[0].Enabled {
		t.Errorf("expected ext-one to be disabled, got: %+v", parsed.Extensions[0])
	}

	// ext-two should remain enabled
	if parsed.Extensions[1].Name != "ext-two" || !parsed.Extensions[1].Enabled {
		t.Errorf("expected ext-two to remain enabled, got: %+v", parsed.Extensions[1])
	}

	// Verify other sections remain intact
	if parsed.Path != ".version" {
		t.Errorf("expected path to be .version, got %q", parsed.Path)
	}
	if len(parsed.PreReleaseHooks) != 1 {
		t.Errorf("expected 1 pre-release hook, got %d", len(parsed.PreReleaseHooks))
	}
}

func TestSetExtensionEnabled_ReadFileError(t *testing.T) {
	invalidPath := filepath.Join(t.TempDir(), "nonexistent.yaml")

	updater := NewDefaultConfigUpdater(&DefaultYAMLMarshaler{})
	err := updater.SetExtensionEnabled(invalidPath, "test", true)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected file not found error, got: %v", err)
	}
}

func TestSetExtensionEnabled_MarshalError(t *testing.T) {
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

	err := updater.SetExtensionEnabled(configPath, "test-ext", false)
	if err == nil || !strings.Contains(err.Error(), "forced marshal failure") {
		t.Fatalf("expected marshal error, got: %v", err)
	}
}

func TestSetExtensionEnabled_WriteFileError(t *testing.T) {
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
	err := updater.SetExtensionEnabled(configPath, "test-ext", false)
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected write error, got: %v", err)
	}
}

func TestReplaceYAMLSection(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		key         string
		replacement string
		wantResult  string
		wantFound   bool
	}{
		{
			name:        "replace empty section",
			content:     "path: .version\nextensions: []\n",
			key:         "extensions",
			replacement: "extensions:\n  - name: foo\n    path: bar\n    enabled: true",
			wantResult:  "path: .version\nextensions:\n  - name: foo\n    path: bar\n    enabled: true\n",
			wantFound:   true,
		},
		{
			name:        "replace section with existing entries",
			content:     "path: .version\nextensions:\n  - name: old\n    path: old/path\n    enabled: true\n",
			key:         "extensions",
			replacement: "extensions:\n  - name: old\n    path: old/path\n    enabled: true\n  - name: new\n    path: new/path\n    enabled: true",
			wantResult:  "path: .version\nextensions:\n  - name: old\n    path: old/path\n    enabled: true\n  - name: new\n    path: new/path\n    enabled: true\n",
			wantFound:   true,
		},
		{
			name:        "key not found",
			content:     "path: .version\nplugins:\n  commit-parser: true\n",
			key:         "extensions",
			replacement: "extensions:\n  - name: foo\n    path: bar\n    enabled: true",
			wantResult:  "path: .version\nplugins:\n  commit-parser: true\n",
			wantFound:   false,
		},
		{
			name:        "preserves surrounding content and comments",
			content:     "# Header comment\npath: .version # inline comment\n\n# Extensions section\nextensions: []\n\n# Hooks section\npre-release-hooks:\n  - changelog:\n      command: git-chglog\n",
			key:         "extensions",
			replacement: "extensions:\n  - name: test\n    path: test/path\n    enabled: true",
			wantResult:  "# Header comment\npath: .version # inline comment\n\n# Extensions section\nextensions:\n  - name: test\n    path: test/path\n    enabled: true\n\n# Hooks section\npre-release-hooks:\n  - changelog:\n      command: git-chglog\n",
			wantFound:   true,
		},
		{
			name:        "section at end of file without trailing newline",
			content:     "path: .version\nextensions: []",
			key:         "extensions",
			replacement: "extensions:\n  - name: foo\n    path: bar\n    enabled: true",
			wantResult:  "path: .version\nextensions:\n  - name: foo\n    path: bar\n    enabled: true\n",
			wantFound:   true,
		},
		{
			name:        "does not match indented key",
			content:     "path: .version\nplugins:\n  extensions: something\n",
			key:         "extensions",
			replacement: "extensions:\n  - name: foo",
			wantResult:  "path: .version\nplugins:\n  extensions: something\n",
			wantFound:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotFound := replaceYAMLSection(tt.content, tt.key, tt.replacement)
			if gotFound != tt.wantFound {
				t.Errorf("replaceYAMLSection() found = %v, want %v", gotFound, tt.wantFound)
			}
			if gotResult != tt.wantResult {
				t.Errorf("replaceYAMLSection() result mismatch.\ngot:\n%s\nwant:\n%s", gotResult, tt.wantResult)
			}
		})
	}
}
