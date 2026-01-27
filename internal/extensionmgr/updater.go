package extensionmgr

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/printer"
)

// DefaultConfigUpdater implements ConfigUpdater for updating extension configurations
type DefaultConfigUpdater struct {
	marshaler YAMLMarshaler
}

// NewDefaultConfigUpdater creates a new DefaultConfigUpdater with the given marshaler
func NewDefaultConfigUpdater(marshaler YAMLMarshaler) *DefaultConfigUpdater {
	return &DefaultConfigUpdater{marshaler: marshaler}
}

// AddExtension appends an extension entry to the YAML config at the given path.
// It avoids duplicates and preserves existing fields.
func (u *DefaultConfigUpdater) AddExtension(path string, extension config.ExtensionConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config %q: %w", path, err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config %q: %w", path, err)
	}

	// Check for duplicates and inform user
	for _, ext := range cfg.Extensions {
		if ext.Name == extension.Name {
			printer.PrintInfo(fmt.Sprintf("Extension %q is already installed at: %s", extension.Name, ext.Path))
			printer.PrintInfo("To reinstall, remove it first:")
			fmt.Printf("  sley extension remove --name %s\n", extension.Name)
			return fmt.Errorf("extension %q already registered in configuration", extension.Name)
		}
	}

	cfg.Extensions = append(cfg.Extensions, extension)

	out, err := u.marshaler.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, out, config.ConfigFilePerm); err != nil {
		return fmt.Errorf("failed to write config %q: %w", path, err)
	}
	return nil
}
