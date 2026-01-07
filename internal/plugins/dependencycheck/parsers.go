package dependencycheck

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"
)

// Function variables for testability.
var (
	readFileFn  = os.ReadFile
	writeFileFn = os.WriteFile

	readJSONVersionFn   = readJSONVersion
	writeJSONVersionFn  = writeJSONVersion
	readYAMLVersionFn   = readYAMLVersion
	writeYAMLVersionFn  = writeYAMLVersion
	readTOMLVersionFn   = readTOMLVersion
	writeTOMLVersionFn  = writeTOMLVersion
	readRawVersionFn    = readRawVersion
	writeRawVersionFn   = writeRawVersion
	readRegexVersionFn  = readRegexVersion
	writeRegexVersionFn = writeRegexVersion
)

// readJSONVersion reads a version from a JSON file using dot notation for nested fields.
func readJSONVersion(path, field string) (string, error) {
	data, err := readFileFn(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %w", path, err)
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

// writeJSONVersion writes a version to a JSON file using dot notation for nested fields.
func writeJSONVersion(path, field, version string) error {
	data, err := readFileFn(path)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", path, err)
	}

	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("failed to parse JSON in %q: %w", path, err)
	}

	if err := setNestedValue(obj, field, version); err != nil {
		return fmt.Errorf("in file %q: %w", path, err)
	}

	// Marshal with indentation for readability
	updated, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for %q: %w", path, err)
	}

	// Add trailing newline
	updated = append(updated, '\n')

	if err := writeFileFn(path, updated, 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}

// readYAMLVersion reads a version from a YAML file using dot notation for nested fields.
func readYAMLVersion(path, field string) (string, error) {
	data, err := readFileFn(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %w", path, err)
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

// writeYAMLVersion writes a version to a YAML file using dot notation for nested fields.
func writeYAMLVersion(path, field, version string) error {
	data, err := readFileFn(path)
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

	if err := writeFileFn(path, updated, 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}

// readTOMLVersion reads a version from a TOML file using dot notation for nested fields.
func readTOMLVersion(path, field string) (string, error) {
	data, err := readFileFn(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %w", path, err)
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

// writeTOMLVersion writes a version to a TOML file using dot notation for nested fields.
func writeTOMLVersion(path, field, version string) error {
	data, err := readFileFn(path)
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

	if err := writeFileFn(path, updated, 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}

// readRawVersion reads the entire file contents as the version (trimmed).
func readRawVersion(path string) (string, error) {
	data, err := readFileFn(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %w", path, err)
	}

	return strings.TrimSpace(string(data)), nil
}

// writeRawVersion writes the version as the entire file contents.
func writeRawVersion(path, version string) error {
	// Ensure version has a trailing newline
	content := version
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := writeFileFn(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
}

// readRegexVersion extracts the version using a regex pattern with a capturing group.
func readRegexVersion(path, pattern string) (string, error) {
	data, err := readFileFn(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %w", path, err)
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

// writeRegexVersion replaces the version in a file using a regex pattern.
func writeRegexVersion(path, pattern, version string) error {
	data, err := readFileFn(path)
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

	if err := writeFileFn(path, updated, 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}

	return nil
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
