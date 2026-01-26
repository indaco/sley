package dependencycheck

import (
	"context"
	"os"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/parser"
)

// Function variables for testability.
// These allow tests to inject mock implementations without modifying the core logic.
var (
	// File I/O functions
	readFileFn  = os.ReadFile
	writeFileFn = os.WriteFile

	// Format-specific read functions
	readJSONVersionFn  = readJSONVersion
	readYAMLVersionFn  = readYAMLVersion
	readTOMLVersionFn  = readTOMLVersion
	readRawVersionFn   = readRawVersion
	readRegexVersionFn = readRegexVersion

	// Format-specific write functions
	writeJSONVersionFn  = writeJSONVersion
	writeYAMLVersionFn  = writeYAMLVersion
	writeTOMLVersionFn  = writeTOMLVersion
	writeRawVersionFn   = writeRawVersion
	writeRegexVersionFn = writeRegexVersion
)

// osFileSystemAdapter wraps os functions to implement core.FileSystem.
// This allows us to use the parser package with the default os functions.
type osFileSystemAdapter struct{}

func (a *osFileSystemAdapter) ReadFile(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return readFileFn(path)
}

func (a *osFileSystemAdapter) WriteFile(ctx context.Context, path string, data []byte, perm os.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return writeFileFn(path, data, perm)
}

func (a *osFileSystemAdapter) Stat(ctx context.Context, path string) (os.FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.Stat(path)
}

func (a *osFileSystemAdapter) MkdirAll(ctx context.Context, path string, perm os.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.MkdirAll(path, perm)
}

func (a *osFileSystemAdapter) Remove(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.Remove(path)
}

func (a *osFileSystemAdapter) RemoveAll(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func (a *osFileSystemAdapter) ReadDir(ctx context.Context, path string) ([]os.DirEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.ReadDir(path)
}

// Ensure osFileSystemAdapter implements core.FileSystem.
var _ core.FileSystem = (*osFileSystemAdapter)(nil)

// getParserReader returns a parser.Reader using the OS filesystem adapter.
func getParserReader() *parser.Reader {
	return parser.NewReader(&osFileSystemAdapter{})
}

// getParserWriter returns a parser.Writer using the OS filesystem adapter.
func getParserWriter() *parser.Writer {
	return parser.NewWriter(&osFileSystemAdapter{})
}

// readJSONVersion reads a version from a JSON file using dot notation for nested fields.
func readJSONVersion(path, field string) (string, error) {
	return getParserReader().ReadVersion(context.Background(), parser.FileConfig{
		Path:   path,
		Format: parser.FormatJSON,
		Field:  field,
	})
}

// writeJSONVersion writes a version to a JSON file using dot notation for nested fields.
func writeJSONVersion(path, field, version string) error {
	return getParserWriter().Write(context.Background(), parser.FileConfig{
		Path:   path,
		Format: parser.FormatJSON,
		Field:  field,
	}, version)
}

// readYAMLVersion reads a version from a YAML file using dot notation for nested fields.
func readYAMLVersion(path, field string) (string, error) {
	return getParserReader().ReadVersion(context.Background(), parser.FileConfig{
		Path:   path,
		Format: parser.FormatYAML,
		Field:  field,
	})
}

// writeYAMLVersion writes a version to a YAML file using dot notation for nested fields.
func writeYAMLVersion(path, field, version string) error {
	return getParserWriter().Write(context.Background(), parser.FileConfig{
		Path:   path,
		Format: parser.FormatYAML,
		Field:  field,
	}, version)
}

// readTOMLVersion reads a version from a TOML file using dot notation for nested fields.
func readTOMLVersion(path, field string) (string, error) {
	return getParserReader().ReadVersion(context.Background(), parser.FileConfig{
		Path:   path,
		Format: parser.FormatTOML,
		Field:  field,
	})
}

// writeTOMLVersion writes a version to a TOML file using dot notation for nested fields.
func writeTOMLVersion(path, field, version string) error {
	return getParserWriter().Write(context.Background(), parser.FileConfig{
		Path:   path,
		Format: parser.FormatTOML,
		Field:  field,
	}, version)
}

// readRawVersion reads the entire file contents as the version (trimmed).
func readRawVersion(path string) (string, error) {
	return getParserReader().ReadVersion(context.Background(), parser.FileConfig{
		Path:   path,
		Format: parser.FormatRaw,
	})
}

// writeRawVersion writes the version as the entire file contents.
func writeRawVersion(path, version string) error {
	return getParserWriter().Write(context.Background(), parser.FileConfig{
		Path:   path,
		Format: parser.FormatRaw,
	}, version)
}

// readRegexVersion extracts the version using a regex pattern with a capturing group.
func readRegexVersion(path, pattern string) (string, error) {
	return getParserReader().ReadVersion(context.Background(), parser.FileConfig{
		Path:    path,
		Format:  parser.FormatRegex,
		Pattern: pattern,
	})
}

// writeRegexVersion replaces the version in a file using a regex pattern.
func writeRegexVersion(path, pattern, version string) error {
	return getParserWriter().Write(context.Background(), parser.FileConfig{
		Path:    path,
		Format:  parser.FormatRegex,
		Pattern: pattern,
	}, version)
}
