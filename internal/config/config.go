package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/core"
)

// ExtensionConfig holds configuration for external extensions.
type ExtensionConfig struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Enabled bool   `yaml:"enabled"`
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

var (
	LoadConfigFn = loadConfig
	SaveConfigFn = saveConfig
	marshalFn    = yaml.Marshal
	openFileFn   = os.OpenFile
	writeFileFn  = func(file *os.File, data []byte) (int, error) {
		return file.Write(data)
	}
)

func loadConfig() (*Config, error) {
	// Highest priority: ENV variable
	if envPath := os.Getenv("SLEY_PATH"); envPath != "" {
		return &Config{Path: envPath}, nil
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

func saveConfig(cfg *Config) error {
	const configFile = ".sley.yaml"
	file, err := openFileFn(configFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, ConfigFilePerm)
	if err != nil {
		return fmt.Errorf("failed to open config file %q: %w", configFile, err)
	}
	defer file.Close()

	data, err := marshalFn(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to %q: %w", configFile, err)
	}

	if _, err := writeFileFn(file, data); err != nil {
		return fmt.Errorf("failed to write config to %q: %w", configFile, err)
	}

	return nil
}
