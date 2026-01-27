package extensionmgr

import (
	"fmt"
	"os"
	"strings"

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
// It avoids duplicates and preserves existing comments and formatting by only
// replacing the extensions section rather than rewriting the entire file.
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
			fmt.Printf("  sley extension uninstall --name %s\n", extension.Name)
			return fmt.Errorf("extension %q already registered in configuration", extension.Name)
		}
	}

	cfg.Extensions = append(cfg.Extensions, extension)

	// Marshal only the extensions section to preserve comments elsewhere.
	type extSection struct {
		Extensions []config.ExtensionConfig `yaml:"extensions"`
	}
	sectionBytes, err := u.marshaler.Marshal(extSection{Extensions: cfg.Extensions})
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Surgically replace just the extensions section in the original content.
	replacement := strings.TrimRight(string(sectionBytes), "\n")
	result, replaced := replaceYAMLSection(string(data), "extensions", replacement)
	if !replaced {
		// Key not found; append the new section at the end.
		original := strings.TrimRight(string(data), "\n")
		result = original + "\n" + replacement + "\n"
	}

	if err := os.WriteFile(path, []byte(result), config.ConfigFilePerm); err != nil {
		return fmt.Errorf("failed to write config %q: %w", path, err)
	}
	return nil
}

// RemoveExtension removes the named extension from the YAML config at the
// given path. It preserves existing comments and formatting by only replacing
// the extensions section rather than rewriting the entire file.
func (u *DefaultConfigUpdater) RemoveExtension(path string, extensionName string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config %q: %w", path, err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config %q: %w", path, err)
	}

	// Find and remove the extension
	found := false
	filtered := make([]config.ExtensionConfig, 0, len(cfg.Extensions))
	for _, ext := range cfg.Extensions {
		if ext.Name == extensionName {
			found = true
			continue
		}
		filtered = append(filtered, ext)
	}

	if !found {
		return fmt.Errorf("extension %q not found in configuration", extensionName)
	}

	cfg.Extensions = filtered

	// Marshal only the extensions section to preserve comments elsewhere.
	type extSection struct {
		Extensions []config.ExtensionConfig `yaml:"extensions"`
	}
	sectionBytes, err := u.marshaler.Marshal(extSection{Extensions: cfg.Extensions})
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Surgically replace just the extensions section in the original content.
	replacement := strings.TrimRight(string(sectionBytes), "\n")
	result, replaced := replaceYAMLSection(string(data), "extensions", replacement)
	if !replaced {
		return fmt.Errorf("extensions section not found in config file %q", path)
	}

	if err := os.WriteFile(path, []byte(result), config.ConfigFilePerm); err != nil {
		return fmt.Errorf("failed to write config %q: %w", path, err)
	}
	return nil
}

// SetExtensionEnabled sets the enabled field for the named extension in the
// YAML config at the given path. It preserves existing comments and formatting
// by only replacing the extensions section rather than rewriting the entire file.
func (u *DefaultConfigUpdater) SetExtensionEnabled(path string, extensionName string, enabled bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config %q: %w", path, err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config %q: %w", path, err)
	}

	// Find the extension and update its enabled field
	found := false
	for i := range cfg.Extensions {
		if cfg.Extensions[i].Name == extensionName {
			cfg.Extensions[i].Enabled = enabled
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("extension %q not found in configuration", extensionName)
	}

	// Marshal only the extensions section to preserve comments elsewhere.
	type extSection struct {
		Extensions []config.ExtensionConfig `yaml:"extensions"`
	}
	sectionBytes, err := u.marshaler.Marshal(extSection{Extensions: cfg.Extensions})
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Surgically replace just the extensions section in the original content.
	replacement := strings.TrimRight(string(sectionBytes), "\n")
	result, replaced := replaceYAMLSection(string(data), "extensions", replacement)
	if !replaced {
		return fmt.Errorf("extensions section not found in config file %q", path)
	}

	if err := os.WriteFile(path, []byte(result), config.ConfigFilePerm); err != nil {
		return fmt.Errorf("failed to write config %q: %w", path, err)
	}
	return nil
}

// replaceYAMLSection replaces a top-level YAML key and its indented block in
// content with the given replacement text. It returns the updated content and
// true if the key was found and replaced, or the original content and false
// if the key was not present.
func replaceYAMLSection(content, key, replacement string) (string, bool) {
	lines := strings.Split(content, "\n")

	startIdx := findTopLevelKeyIndex(lines, key)
	if startIdx == -1 {
		return content, false
	}

	endIdx := findSectionEnd(lines, startIdx)

	return buildReplacedContent(lines, startIdx, endIdx, replacement), true
}

// findTopLevelKeyIndex returns the line index of the first top-level YAML key
// matching key, or -1 if not found. A top-level key has no leading whitespace
// and is followed by a colon.
func findTopLevelKeyIndex(lines []string, key string) int {
	prefix := key + ":"
	for i, line := range lines {
		if !isTopLevelLine(line) {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == prefix || strings.HasPrefix(trimmed, prefix+" ") || strings.HasPrefix(trimmed, prefix+"\n") {
			return i
		}
	}
	return -1
}

// isTopLevelLine reports whether a line is non-empty and has no leading
// whitespace, indicating a top-level YAML key.
func isTopLevelLine(line string) bool {
	return len(line) > 0 && line[0] != ' ' && line[0] != '\t'
}

// isBlankLine reports whether a line is empty or contains only whitespace.
func isBlankLine(line string) bool {
	return strings.TrimSpace(line) == ""
}

// isIndentedLine reports whether a non-empty line starts with whitespace.
func isIndentedLine(line string) bool {
	return len(line) > 0 && (line[0] == ' ' || line[0] == '\t')
}

// findSectionEnd returns the index of the first line after the section that
// starts at startIdx. Indented and blank lines belong to the section. A blank
// line followed by a non-indented line (or EOF) marks the section boundary.
func findSectionEnd(lines []string, startIdx int) int {
	endIdx := startIdx + 1
	for endIdx < len(lines) {
		line := lines[endIdx]
		if isBlankLine(line) {
			if ahead := skipBlankLines(lines, endIdx+1); isIndentedLine(safeLineAt(lines, ahead)) {
				endIdx = ahead + 1
				continue
			}
			break
		}
		if !isIndentedLine(line) {
			break
		}
		endIdx++
	}
	return endIdx
}

// skipBlankLines advances from index start past any consecutive blank lines
// and returns the index of the first non-blank line (or len(lines) if none).
func skipBlankLines(lines []string, start int) int {
	i := start
	for i < len(lines) && isBlankLine(lines[i]) {
		i++
	}
	return i
}

// safeLineAt returns lines[i] if i is within bounds, or an empty string
// otherwise. This avoids index-out-of-range checks at call sites.
func safeLineAt(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return ""
}

// buildReplacedContent assembles the final content by concatenating lines
// before startIdx, the replacement text, and lines from endIdx onward.
func buildReplacedContent(lines []string, startIdx, endIdx int, replacement string) string {
	var result strings.Builder
	for i := range startIdx {
		result.WriteString(lines[i])
		result.WriteString("\n")
	}
	result.WriteString(replacement)
	result.WriteString("\n")
	for i := endIdx; i < len(lines); i++ {
		result.WriteString(lines[i])
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
}
