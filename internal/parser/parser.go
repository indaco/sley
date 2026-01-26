package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/indaco/sley/internal/core"
	"github.com/pelletier/go-toml/v2"
)

// Reader provides version reading capabilities for multiple file formats.
type Reader struct {
	fs core.FileSystem
}

// NewReader creates a new Reader with the given filesystem.
func NewReader(fs core.FileSystem) *Reader {
	return &Reader{fs: fs}
}

// Read reads a version from a file based on the provided configuration.
func (r *Reader) Read(ctx context.Context, cfg FileConfig) (*Result, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("file path is required")
	}

	if !cfg.Format.IsValid() {
		return nil, fmt.Errorf("invalid format: %s", cfg.Format)
	}

	data, err := r.fs.ReadFile(ctx, cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", cfg.Path, err)
	}

	var version string
	switch cfg.Format {
	case FormatJSON:
		version, err = r.readJSON(data, cfg.Path, cfg.Field)
	case FormatYAML:
		version, err = r.readYAML(data, cfg.Path, cfg.Field)
	case FormatTOML:
		version, err = r.readTOML(data, cfg.Path, cfg.Field)
	case FormatRaw:
		version, err = r.readRaw(data)
	case FormatRegex:
		version, err = r.readRegex(data, cfg.Path, cfg.Pattern)
	default:
		return nil, fmt.Errorf("unsupported format: %s", cfg.Format)
	}

	if err != nil {
		return nil, err
	}

	return &Result{
		Version: version,
		Path:    cfg.Path,
		Format:  cfg.Format,
		Field:   cfg.Field,
	}, nil
}

// ReadVersion is a convenience method that reads and returns just the version string.
func (r *Reader) ReadVersion(ctx context.Context, cfg FileConfig) (string, error) {
	result, err := r.Read(ctx, cfg)
	if err != nil {
		return "", err
	}
	return result.Version, nil
}

// readJSON extracts a version from JSON data using dot notation for the field path.
func (r *Reader) readJSON(data []byte, path, field string) (string, error) {
	if field == "" {
		return "", fmt.Errorf("field is required for JSON format")
	}

	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", fmt.Errorf("failed to parse JSON in %q: %w", path, err)
	}

	value, err := getNestedValue(obj, field)
	if err != nil {
		return "", fmt.Errorf("in file %q: %w", path, err)
	}

	version, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("field %q in %q is not a string", field, path)
	}

	return version, nil
}

// readYAML extracts a version from YAML data using dot notation for the field path.
func (r *Reader) readYAML(data []byte, path, field string) (string, error) {
	if field == "" {
		return "", fmt.Errorf("field is required for YAML format")
	}

	var obj map[string]any
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return "", fmt.Errorf("failed to parse YAML in %q: %w", path, err)
	}

	value, err := getNestedValue(obj, field)
	if err != nil {
		return "", fmt.Errorf("in file %q: %w", path, err)
	}

	version, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("field %q in %q is not a string", field, path)
	}

	return version, nil
}

// readTOML extracts a version from TOML data using dot notation for the field path.
func (r *Reader) readTOML(data []byte, path, field string) (string, error) {
	if field == "" {
		return "", fmt.Errorf("field is required for TOML format")
	}

	var obj map[string]any
	if err := toml.Unmarshal(data, &obj); err != nil {
		return "", fmt.Errorf("failed to parse TOML in %q: %w", path, err)
	}

	value, err := getNestedValue(obj, field)
	if err != nil {
		return "", fmt.Errorf("in file %q: %w", path, err)
	}

	version, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("field %q in %q is not a string", field, path)
	}

	return version, nil
}

// readRaw reads the entire file contents as the version (trimmed).
func (r *Reader) readRaw(data []byte) (string, error) {
	return strings.TrimSpace(string(data)), nil
}

// readRegex extracts a version using a regex pattern with a capturing group.
func (r *Reader) readRegex(data []byte, path, pattern string) (string, error) {
	if pattern == "" {
		return "", fmt.Errorf("pattern is required for regex format")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
	}

	matches := re.FindSubmatch(data)
	if len(matches) < 2 {
		return "", fmt.Errorf("no version match found in %q (pattern %q must have capturing group)", path, pattern)
	}

	return string(matches[1]), nil
}

// getNestedValue retrieves a value from a nested map using dot notation.
// Example: "tool.poetry.version" accesses obj["tool"]["poetry"]["version"]
func getNestedValue(obj map[string]any, field string) (any, error) {
	if field == "" {
		return nil, fmt.Errorf("field path cannot be empty")
	}

	parts := strings.Split(field, ".")
	current := any(obj)

	for i, part := range parts {
		currentMap, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("field %q is not an object at path %q", strings.Join(parts[:i], "."), part)
		}

		value, exists := currentMap[part]
		if !exists {
			return nil, fmt.Errorf("field %q not found", field)
		}

		current = value
	}

	return current, nil
}
