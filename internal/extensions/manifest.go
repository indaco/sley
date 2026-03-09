package extensions

import (
	"fmt"
	"strings"
)

const (
	// CurrentSchemaVersion is the latest manifest schema version supported by this build.
	CurrentSchemaVersion = 1

	// DefaultSchemaVersion is used when the schema_version field is omitted (backward compat).
	DefaultSchemaVersion = 1
)

// SchemaVersionError indicates the manifest requires a newer schema version than supported.
type SchemaVersionError struct {
	Path         string
	Found        int
	MaxSupported int
}

func (e *SchemaVersionError) Error() string {
	return fmt.Sprintf("manifest at %s requires schema version %d, but this build only supports up to version %d",
		e.Path, e.Found, e.MaxSupported)
}

// Suggestion returns guidance on resolving the version mismatch
func (e *SchemaVersionError) Suggestion() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Manifest schema version %d is not supported (max: %d).\n\n", e.Found, e.MaxSupported)
	sb.WriteString("Please upgrade sley to a newer version that supports this manifest schema:\n\n")
	sb.WriteString("  go install github.com/indaco/sley@latest\n\n")
	sb.WriteString("Documentation: https://sley.indaco.dev/extensions/index.html\n")

	return sb.String()
}

// ManifestNotFoundError indicates that an extension.yaml file is missing
type ManifestNotFoundError struct {
	Path string
	Dir  string
}

func (e *ManifestNotFoundError) Error() string {
	return fmt.Sprintf("extension manifest not found: %s", e.Path)
}

// Suggestion returns a helpful message with a manifest template
func (e *ManifestNotFoundError) Suggestion() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Extension manifest not found at: %s\n\n", e.Path)
	sb.WriteString("A valid extension.yaml file is required with these fields:\n\n")
	sb.WriteString("  schema_version: 1\n")
	sb.WriteString("  name: my-extension\n")
	sb.WriteString("  version: 1.0.0\n")
	sb.WriteString("  description: Brief description of what this extension does\n")
	sb.WriteString("  author: Your Name\n")
	sb.WriteString("  repository: https://github.com/user/repo\n")
	sb.WriteString("  entry: script.sh\n")
	sb.WriteString("  hooks: [post-bump]  # optional\n\n")
	sb.WriteString("Documentation: https://sley.dev/extensions/manifest\n")

	return sb.String()
}

// ManifestParseError indicates that a manifest file has invalid YAML or structure
type ManifestParseError struct {
	Path string
	Err  error
}

func (e *ManifestParseError) Error() string {
	return fmt.Sprintf("failed to parse manifest at %s: %v", e.Path, e.Err)
}

// Unwrap returns the underlying error
func (e *ManifestParseError) Unwrap() error {
	return e.Err
}

// ManifestValidationError indicates that required fields are missing
type ManifestValidationError struct {
	Path          string
	MissingFields []string
}

func (e *ManifestValidationError) Error() string {
	return fmt.Sprintf("invalid manifest at %s: missing required fields: %s",
		e.Path, strings.Join(e.MissingFields, ", "))
}

// Suggestion returns guidance on fixing validation errors
func (e *ManifestValidationError) Suggestion() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Manifest validation failed: %s\n\n", e.Path)
	sb.WriteString("Missing required fields:\n")
	for _, field := range e.MissingFields {
		fmt.Fprintf(&sb, "  - %s\n", field)
	}
	sb.WriteString("\nAll extension manifests must include:\n")
	sb.WriteString("  - name: Unique extension identifier\n")
	sb.WriteString("  - version: Extension version (e.g., 1.0.0)\n")
	sb.WriteString("  - description: What the extension does\n")
	sb.WriteString("  - author: Extension author name\n")
	sb.WriteString("  - repository: Source code URL\n")
	sb.WriteString("  - entry: Path to executable script\n\n")
	sb.WriteString("Documentation: https://sley.dev/extensions/manifest\n")

	return sb.String()
}

// ExtensionManifest defines the metadata and entry point for a sley extension.
// This structure is expected to be defined in a extension's `extension.yaml` file.
//
// All fields except Hooks are required:
// - Name: A unique extension identifier (e.g. "changelog-generator")
// - Version: The extension's version (e.g. "0.1.0")
// - Description: A brief explanation of what the extension does
// - Author: Name or handle of the extension author
// - Repository: URL of the extension's source repository
// - Entry: Path to the executable script or binary (relative to extension directory)
// - Hooks: List of hook points this extension supports (optional)
type ExtensionManifest struct {
	SchemaVersion int      `yaml:"schema_version,omitempty"`
	Name          string   `yaml:"name"`
	Version       string   `yaml:"version"`
	Description   string   `yaml:"description"`
	Author        string   `yaml:"author"`
	Repository    string   `yaml:"repository"`
	Entry         string   `yaml:"entry"`
	Hooks         []string `yaml:"hooks,omitempty"`
}

// ValidateManifest ensures all required fields are present and the manifest version is supported.
// Returns an error listing all missing fields if validation fails.
func (m *ExtensionManifest) ValidateManifest() error {
	// Default omitted schema_version to DefaultSchemaVersion
	if m.SchemaVersion == 0 {
		m.SchemaVersion = DefaultSchemaVersion
	}

	// Reject negative or otherwise invalid schema versions
	if m.SchemaVersion < 1 {
		return &ManifestValidationError{
			MissingFields: []string{"schema_version (must be >= 1)"},
		}
	}

	// Reject schema versions newer than what this build supports
	if m.SchemaVersion > CurrentSchemaVersion {
		return &SchemaVersionError{
			Found:        m.SchemaVersion,
			MaxSupported: CurrentSchemaVersion,
		}
	}

	var missingFields []string

	if m.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if m.Version == "" {
		missingFields = append(missingFields, "version")
	}
	if m.Description == "" {
		missingFields = append(missingFields, "description")
	}
	if m.Author == "" {
		missingFields = append(missingFields, "author")
	}
	if m.Repository == "" {
		missingFields = append(missingFields, "repository")
	}
	if m.Entry == "" {
		missingFields = append(missingFields, "entry")
	}

	if len(missingFields) > 0 {
		return &ManifestValidationError{
			MissingFields: missingFields,
		}
	}

	return nil
}
