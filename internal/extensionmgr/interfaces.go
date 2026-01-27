package extensionmgr

import (
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/extensions"
)

// YAMLMarshaler handles YAML marshaling with custom options
type YAMLMarshaler interface {
	Marshal(v any) ([]byte, error)
}

// ConfigUpdater handles extension configuration updates
type ConfigUpdater interface {
	AddExtension(path string, extension config.ExtensionConfig) error
	RemoveExtension(path string, extensionName string) error
}

// ManifestLoader handles loading extension manifests
type ManifestLoader interface {
	Load(path string) (*extensions.ExtensionManifest, error)
}

// HomeDirectory provides access to the user's home directory
type HomeDirectory interface {
	Get() (string, error)
}

// ExtensionRegistrar handles registering local extensions
type ExtensionRegistrar interface {
	Register(localPath, configPath, extensionDirectory string) error
}
