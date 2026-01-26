package parser

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/core"
	"github.com/pelletier/go-toml/v2"
	"github.com/tidwall/sjson"
)

// Writer provides version writing capabilities for multiple file formats.
type Writer struct {
	fs core.FileSystem
}

// NewWriter creates a new Writer with the given filesystem.
func NewWriter(fs core.FileSystem) *Writer {
	return &Writer{fs: fs}
}

// Write writes a version to a file based on the provided configuration.
func (w *Writer) Write(ctx context.Context, cfg FileConfig, version string) error {
	if cfg.Path == "" {
		return fmt.Errorf("file path is required")
	}

	if !cfg.Format.IsValid() {
		return fmt.Errorf("invalid format: %s", cfg.Format)
	}

	switch cfg.Format {
	case FormatJSON:
		return w.writeJSON(ctx, cfg.Path, cfg.Field, version)
	case FormatYAML:
		return w.writeYAML(ctx, cfg.Path, cfg.Field, version)
	case FormatTOML:
		return w.writeTOML(ctx, cfg.Path, cfg.Field, version)
	case FormatRaw:
		return w.writeRaw(ctx, cfg.Path, version)
	case FormatRegex:
		return w.writeRegex(ctx, cfg.Path, cfg.Pattern, version)
	default:
		return fmt.Errorf("unsupported format: %s", cfg.Format)
	}
}

// writeJSON writes a version to a JSON file using sjson to preserve formatting.
func (w *Writer) writeJSON(ctx context.Context, path, field, version string) error {
	if field == "" {
		return fmt.Errorf("field is required for JSON format")
	}

	data, err := w.fs.ReadFile(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", path, err)
	}

	// Use sjson to update only the specified field, preserving structure and field order
	updated, err := sjson.SetBytes(data, field, version)
	if err != nil {
		return fmt.Errorf("failed to set version in %q: %w", path, err)
	}

	// Ensure trailing newline
	if len(updated) > 0 && updated[len(updated)-1] != '\n' {
		updated = append(updated, '\n')
	}

	if err := w.fs.WriteFile(ctx, path, updated, core.PermOwnerRW); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}

// writeYAML writes a version to a YAML file.
func (w *Writer) writeYAML(ctx context.Context, path, field, version string) error {
	if field == "" {
		return fmt.Errorf("field is required for YAML format")
	}

	data, err := w.fs.ReadFile(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", path, err)
	}

	var obj map[string]any
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("failed to parse YAML in %q: %w", path, err)
	}

	if err := setNestedValue(obj, field, version); err != nil {
		return fmt.Errorf("in file %q: %w", path, err)
	}

	updated, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML for %q: %w", path, err)
	}

	if err := w.fs.WriteFile(ctx, path, updated, core.PermOwnerRW); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}

// writeTOML writes a version to a TOML file.
func (w *Writer) writeTOML(ctx context.Context, path, field, version string) error {
	if field == "" {
		return fmt.Errorf("field is required for TOML format")
	}

	data, err := w.fs.ReadFile(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", path, err)
	}

	var obj map[string]any
	if err := toml.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("failed to parse TOML in %q: %w", path, err)
	}

	if err := setNestedValue(obj, field, version); err != nil {
		return fmt.Errorf("in file %q: %w", path, err)
	}

	updated, err := toml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal TOML for %q: %w", path, err)
	}

	if err := w.fs.WriteFile(ctx, path, updated, core.PermOwnerRW); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}

// writeRaw writes the version as the entire file contents.
func (w *Writer) writeRaw(ctx context.Context, path, version string) error {
	// Ensure version has a trailing newline
	content := version
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := w.fs.WriteFile(ctx, path, []byte(content), core.PermOwnerRW); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}

// writeRegex replaces the version in a file using a regex pattern.
func (w *Writer) writeRegex(ctx context.Context, path, pattern, version string) error {
	if pattern == "" {
		return fmt.Errorf("pattern is required for regex format")
	}

	data, err := w.fs.ReadFile(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", path, err)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
	}

	// Find the first match to ensure pattern is valid
	if !re.Match(data) {
		return fmt.Errorf("pattern %q does not match contents of %q", pattern, path)
	}

	// Replace using ReplaceAllFunc to preserve surrounding text
	updated := re.ReplaceAllFunc(data, func(match []byte) []byte {
		// Find submatch to get the structure
		submatches := re.FindSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		// Replace the first capturing group
		return []byte(strings.Replace(string(match), string(submatches[1]), version, 1))
	})

	if err := w.fs.WriteFile(ctx, path, updated, core.PermOwnerRW); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}

// setNestedValue sets a value in a nested map using dot notation.
// Example: "tool.poetry.version" sets obj["tool"]["poetry"]["version"] = value
func setNestedValue(obj map[string]any, field string, value any) error {
	if field == "" {
		return fmt.Errorf("field path cannot be empty")
	}

	parts := strings.Split(field, ".")
	current := obj

	// Navigate to the parent of the target field
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]

		next, exists := current[part]
		if !exists {
			// Create intermediate maps if they don't exist
			newMap := make(map[string]any)
			current[part] = newMap
			current = newMap
			continue
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			return fmt.Errorf("field %q is not an object at path %q", strings.Join(parts[:i+1], "."), part)
		}

		current = nextMap
	}

	// Set the final value
	current[parts[len(parts)-1]] = value
	return nil
}

// Exists checks if a file exists at the given path.
func (w *Writer) Exists(ctx context.Context, path string) bool {
	_, err := w.fs.Stat(ctx, path)
	return err == nil
}

// ReadWriter combines Reader and Writer functionality.
type ReadWriter struct {
	*Reader
	*Writer
}

// NewReadWriter creates a new ReadWriter with the given filesystem.
func NewReadWriter(fs core.FileSystem) *ReadWriter {
	return &ReadWriter{
		Reader: NewReader(fs),
		Writer: NewWriter(fs),
	}
}

// FieldForFormat returns the typical field path for common file types.
func FieldForFormat(filename string) string {
	// Map common file names to their version field paths
	fields := map[string]string{
		"package.json":   "version",
		"composer.json":  "version",
		"Cargo.toml":     "package.version",
		"pyproject.toml": "project.version",
		"Chart.yaml":     "version",
		"pubspec.yaml":   "version",
	}

	// Check both the full path and just the filename
	if field, ok := fields[filename]; ok {
		return field
	}

	// Try to extract just the filename
	parts := strings.Split(filename, "/")
	basename := parts[len(parts)-1]
	if field, ok := fields[basename]; ok {
		return field
	}

	return "version"
}

// FormatForFile detects the format based on file extension or name.
func FormatForFile(filename string) Format {
	lower := strings.ToLower(filename)

	switch {
	case strings.HasSuffix(lower, ".json"):
		return FormatJSON
	case strings.HasSuffix(lower, ".yaml"), strings.HasSuffix(lower, ".yml"):
		return FormatYAML
	case strings.HasSuffix(lower, ".toml"):
		return FormatTOML
	case strings.HasSuffix(lower, ".txt"), lower == "version", lower == ".version":
		return FormatRaw
	default:
		// Check specific file names
		switch {
		case strings.HasSuffix(lower, "/package.json"), lower == "package.json":
			return FormatJSON
		case strings.HasSuffix(lower, "/cargo.toml"), lower == "cargo.toml":
			return FormatTOML
		case strings.HasSuffix(lower, "/pyproject.toml"), lower == "pyproject.toml":
			return FormatTOML
		case strings.HasSuffix(lower, "/chart.yaml"), lower == "chart.yaml":
			return FormatYAML
		default:
			return FormatRaw
		}
	}
}
