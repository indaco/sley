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

	if err := AddExtensionToConfig(configPath, extension); err != nil {
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
	err = AddExtensionToConfig(configPath, extension)
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

	err := AddExtensionToConfig(invalidPath, extension)
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

	err := AddExtensionToConfig(configPath, config.ExtensionConfig{
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

	// Backup the original yaml.Marshal
	originalMarshal := marshalFunc
	defer func() { marshalFunc = originalMarshal }()

	// Force yaml.Marshal to fail
	marshalFunc = func(v any) ([]byte, error) {
		return nil, errors.New("forced marshal failure")
	}

	err := AddExtensionToConfig(configPath, config.ExtensionConfig{
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

	err := AddExtensionToConfig(configPath, config.ExtensionConfig{
		Name:    "test",
		Path:    "some/path",
		Enabled: true,
	})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected write error, got: %v", err)
	}
}
