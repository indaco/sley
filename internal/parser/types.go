package parser

// Format represents the supported file formats for version parsing.
type Format string

const (
	// FormatJSON is for JSON files (package.json, etc.).
	FormatJSON Format = "json"

	// FormatYAML is for YAML files (Chart.yaml, etc.).
	FormatYAML Format = "yaml"

	// FormatTOML is for TOML files (Cargo.toml, pyproject.toml, etc.).
	FormatTOML Format = "toml"

	// FormatRaw is for plain text files where the entire content is the version.
	FormatRaw Format = "raw"

	// FormatRegex is for files requiring regex extraction.
	FormatRegex Format = "regex"
)

// String returns the string representation of the format.
func (f Format) String() string {
	return string(f)
}

// IsValid returns true if the format is a known valid format.
func (f Format) IsValid() bool {
	switch f {
	case FormatJSON, FormatYAML, FormatTOML, FormatRaw, FormatRegex:
		return true
	default:
		return false
	}
}

// ParseFormat converts a string to a Format, returning FormatRaw as fallback.
func ParseFormat(s string) Format {
	f := Format(s)
	if f.IsValid() {
		return f
	}
	return FormatRaw
}

// FileConfig describes how to read a version from a specific file.
type FileConfig struct {
	// Path is the file path (absolute or relative).
	Path string

	// Format specifies the file format.
	Format Format

	// Field is the dot-notation path to the version field (for JSON/YAML/TOML).
	// Example: "version", "package.version", "tool.poetry.version"
	Field string

	// Pattern is the regex pattern for regex format.
	// Must contain a capturing group for the version.
	Pattern string
}

// Result represents the result of reading a version from a file.
type Result struct {
	// Version is the extracted version string.
	Version string

	// Path is the file path that was read.
	Path string

	// Format is the format that was used.
	Format Format

	// Field is the field path that was used (for structured formats).
	Field string
}
