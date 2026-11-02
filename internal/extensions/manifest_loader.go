package extensions

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// LoadExtensionManifest loads and validates a extension.yaml file from the given directory.
// Returns context-aware errors for common failure scenarios.
func LoadExtensionManifest(dir string) (*ExtensionManifest, error) {
	manifestPath := filepath.Join(dir, "extension.yaml")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		// Distinguish file not found from other read errors
		if os.IsNotExist(err) {
			return nil, &ManifestNotFoundError{
				Path: manifestPath,
				Dir:  dir,
			}
		}
		return nil, fmt.Errorf("failed to read manifest at %q: %w", manifestPath, err)
	}

	var manifest ExtensionManifest
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.Strict())

	if err := decoder.Decode(&manifest); err != nil {
		return nil, &ManifestParseError{
			Path: manifestPath,
			Err:  err,
		}
	}

	if err := manifest.ValidateManifest(); err != nil {
		// If it's already our custom error type, add the path
		if versionErr, ok := errors.AsType[*SchemaVersionError](err); ok {
			versionErr.Path = manifestPath
			return nil, versionErr
		}
		if valErr, ok := errors.AsType[*ManifestValidationError](err); ok {
			valErr.Path = manifestPath
			return nil, valErr
		}
		return nil, fmt.Errorf("invalid manifest at %q: %w", manifestPath, err)
	}

	return &manifest, nil
}
