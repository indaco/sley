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
			fmt.Printf("  sley extension remove --name %s\n", extension.Name)
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

// replaceYAMLSection replaces a top-level YAML key and its indented block in
// content with the given replacement text. It returns the updated content and
// true if the key was found and replaced, or the original content and false
// if the key was not present.
func replaceYAMLSection(content, key, replacement string) (string, bool) {
	lines := strings.Split(content, "\n")
	prefix := key + ":"

	startIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Match only top-level keys (no leading whitespace) that start with
		// the key name followed by a colon.
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' &&
			(trimmed == prefix || strings.HasPrefix(trimmed, prefix+" ") || strings.HasPrefix(trimmed, prefix+"\n")) {
			startIdx = i
			break
		}
	}

	if startIdx == -1 {
		return content, false
	}

	// Determine the end of the section: all subsequent lines that are indented
	// or blank belong to this section. A non-indented, non-blank line marks
	// the start of the next section.
	endIdx := startIdx + 1
	for endIdx < len(lines) {
		line := lines[endIdx]
		if line == "" || strings.TrimSpace(line) == "" {
			// Blank lines might be part of the section or a separator.
			// Look ahead to see if the next non-blank line is still indented.
			ahead := endIdx + 1
			for ahead < len(lines) && strings.TrimSpace(lines[ahead]) == "" {
				ahead++
			}
			if ahead < len(lines) && len(lines[ahead]) > 0 && (lines[ahead][0] == ' ' || lines[ahead][0] == '\t') {
				// Still inside the section.
				endIdx = ahead + 1
				continue
			}
			// Blank line(s) followed by a top-level key or EOF: section ends here.
			break
		}
		if line[0] != ' ' && line[0] != '\t' {
			// Next top-level key: section ends before this line.
			break
		}
		endIdx++
	}

	// Build the result: lines before the section + replacement + lines after.
	var result strings.Builder
	for i := 0; i < startIdx; i++ {
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

	return result.String(), true
}
