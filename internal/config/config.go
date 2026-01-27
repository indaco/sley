package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/core"
)

// ExtensionConfig holds configuration for external extensions.
type ExtensionConfig struct {
	Name    string         `yaml:"name"`
	Path    string         `yaml:"path"`
	Enabled bool           `yaml:"enabled"`
	Config  map[string]any `yaml:"config,omitempty"`
}

// PreReleaseHookConfig holds configuration for pre-release hooks.
type PreReleaseHookConfig struct {
	Command string `yaml:"command,omitempty"`
}

// Config is the main configuration structure for sley.
type Config struct {
	Path            string                            `yaml:"path"`
	Plugins         *PluginConfig                     `yaml:"plugins,omitempty"`
	Extensions      []ExtensionConfig                 `yaml:"extensions,omitempty"`
	PreReleaseHooks []map[string]PreReleaseHookConfig `yaml:"pre-release-hooks,omitempty"`
	Workspace       *WorkspaceConfig                  `yaml:"workspace,omitempty"`
}

// FileOpener abstracts file opening operations for testability.
type FileOpener interface {
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
}

// FileWriter abstracts file writing operations for testability.
type FileWriter interface {
	WriteFile(file *os.File, data []byte) (int, error)
}

// ConfigSaver handles configuration saving with injected dependencies.
type ConfigSaver struct {
	marshaler  core.Marshaler
	fileOpener FileOpener
	fileWriter FileWriter
}

// osFileOpener is the production implementation of FileOpener.
type osFileOpener struct{}

func (o *osFileOpener) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

// osFileWriter is the production implementation of FileWriter.
type osFileWriter struct{}

func (w *osFileWriter) WriteFile(file *os.File, data []byte) (int, error) {
	return file.Write(data)
}

// yamlMarshaler is the production implementation of core.Marshaler using YAML.
type yamlMarshaler struct{}

func (m *yamlMarshaler) Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

// NewConfigSaver creates a ConfigSaver with the given dependencies.
// If any dependency is nil, the production default is used.
func NewConfigSaver(marshaler core.Marshaler, opener FileOpener, writer FileWriter) *ConfigSaver {
	if marshaler == nil {
		marshaler = &yamlMarshaler{}
	}
	if opener == nil {
		opener = &osFileOpener{}
	}
	if writer == nil {
		writer = &osFileWriter{}
	}
	return &ConfigSaver{
		marshaler:  marshaler,
		fileOpener: opener,
		fileWriter: writer,
	}
}

// Save saves the configuration to the default config file.
func (s *ConfigSaver) Save(cfg *Config) error {
	return s.SaveTo(cfg, ".sley.yaml")
}

// SaveTo saves the configuration to the specified file path.
func (s *ConfigSaver) SaveTo(cfg *Config, configFile string) error {
	file, err := s.fileOpener.OpenFile(configFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, ConfigFilePerm)
	if err != nil {
		return fmt.Errorf("failed to open config file %q: %w", configFile, err)
	}
	defer file.Close()

	data, err := s.marshaler.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to %q: %w", configFile, err)
	}

	if _, err := s.fileWriter.WriteFile(file, data); err != nil {
		return fmt.Errorf("failed to write config to %q: %w", configFile, err)
	}

	return nil
}

// defaultConfigSaver is the default ConfigSaver instance for backward compatibility.
var defaultConfigSaver = NewConfigSaver(nil, nil, nil)

// LoadConfigFn and SaveConfigFn are kept for backward compatibility during migration.
// They delegate to the interface-based implementations.
var (
	LoadConfigFn = loadConfig
	SaveConfigFn = func(cfg *Config) error {
		return defaultConfigSaver.Save(cfg)
	}
)

func loadConfig() (*Config, error) {
	// Highest priority: ENV variable
	if envPath := os.Getenv("SLEY_PATH"); envPath != "" {
		cleanPath := filepath.Clean(envPath)
		// Reject relative paths with traversal (use absolute paths instead)
		if strings.Contains(cleanPath, "..") {
			return nil, fmt.Errorf("invalid SLEY_PATH: path traversal not allowed, use absolute path instead")
		}
		return &Config{Path: cleanPath}, nil
	}

	// Second priority: YAML file
	data, err := os.ReadFile(".sley.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // fallback to default
		}
		return nil, err
	}

	var cfg Config
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.Strict())
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	if cfg.Path == "" {
		cfg.Path = ".version"
	}

	if cfg.Plugins == nil {
		cfg.Plugins = &PluginConfig{CommitParser: true}
	}

	return &cfg, nil
}

// NormalizeVersionPath ensures the path is a file, not just a directory.
func NormalizeVersionPath(path string) string {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return filepath.Join(path, ".version")
	}

	// If it doesn't exist or is already a file, return as-is
	return path
}

// ConfigFilePerm defines secure file permissions for config files (owner read/write only).
// References core.PermOwnerRW for consistency across the codebase.
const ConfigFilePerm = core.PermOwnerRW
