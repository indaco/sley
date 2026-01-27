package extensionmgr

import (
	"os"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/extensions"
)

// DefaultYAMLMarshaler implements YAMLMarshaler with proper indentation
type DefaultYAMLMarshaler struct{}

// Marshal marshals config with proper indentation (2 spaces for both maps and sequences)
func (m *DefaultYAMLMarshaler) Marshal(v any) ([]byte, error) {
	return yaml.MarshalWithOptions(v, yaml.Indent(2), yaml.IndentSequence(true))
}

// OSHomeDirectory implements HomeDirectory using os.UserHomeDir
type OSHomeDirectory struct{}

// Get returns the user's home directory
func (h *OSHomeDirectory) Get() (string, error) {
	return os.UserHomeDir()
}

// DefaultManifestLoader implements ManifestLoader using extensions.LoadExtensionManifestFn
type DefaultManifestLoader struct{}

// Load loads an extension manifest from the given path
func (l *DefaultManifestLoader) Load(path string) (*extensions.ExtensionManifest, error) {
	return extensions.LoadExtensionManifestFn(path)
}
