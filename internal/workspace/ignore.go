package workspace

import (
	"path/filepath"
	"slices"
	"strings"
)

// IgnoreFile represents a parsed .sleyignore file.
type IgnoreFile struct {
	patterns []string
}

// NewIgnoreFile creates an IgnoreFile from the given content.
func NewIgnoreFile(content string) *IgnoreFile {
	return &IgnoreFile{
		patterns: parseIgnoreContent(content),
	}
}

// Matches checks if the given path should be ignored.
// It supports:
//   - Exact matching: "node_modules" matches "node_modules"
//   - Glob patterns: "*.tmp" matches "foo.tmp", "test/*.log" matches "test/app.log"
//   - Directory patterns: "build/" matches "build" directory
func (i *IgnoreFile) Matches(path string) bool {
	// Normalize path separators
	path = filepath.ToSlash(path)

	for _, pattern := range i.patterns {
		if matchIgnorePattern(pattern, path) {
			return true
		}
	}
	return false
}

// Patterns returns all patterns in the ignore file.
func (i *IgnoreFile) Patterns() []string {
	return append([]string(nil), i.patterns...)
}

// parseIgnoreContent parses .sleyignore content and returns patterns.
func parseIgnoreContent(content string) []string {
	var patterns []string
	lines := strings.SplitSeq(content, "\n")

	for line := range lines {
		// Trim whitespace
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		patterns = append(patterns, line)
	}

	return patterns
}

// matchIgnorePattern performs gitignore-style pattern matching.
func matchIgnorePattern(pattern, path string) bool {
	// Normalize pattern
	pattern = filepath.ToSlash(pattern)

	// Exact match
	if pattern == path {
		return true
	}

	// Directory pattern (ends with /)
	if matchDirectoryPattern(pattern, path) {
		return true
	}

	// Check if pattern contains wildcard
	if !strings.Contains(pattern, "*") {
		// No wildcard, check if path contains pattern as component
		pathComponents := strings.Split(path, "/")
		return slices.Contains(pathComponents, pattern)
	}

	// Glob pattern matching
	return matchGlobPattern(pattern, path)
}

// matchDirectoryPattern checks if pattern is a directory pattern and matches.
func matchDirectoryPattern(pattern, path string) bool {
	before, ok := strings.CutSuffix(pattern, "/")
	if !ok {
		return false
	}
	// Match directory itself or anything under it
	return path == before || strings.HasPrefix(path, before+"/")
}

// matchGlobPattern handles glob pattern matching.
func matchGlobPattern(pattern, path string) bool {
	// If pattern contains /, it's a path pattern
	if strings.Contains(pattern, "/") {
		return matchPathGlob(pattern, path)
	}
	// Simple glob pattern (no /), match against any path component
	return matchComponentGlob(pattern, path)
}

// matchPathGlob matches a glob pattern that contains path separators.
func matchPathGlob(pattern, path string) bool {
	matched, err := filepath.Match(pattern, path)
	if err == nil && matched {
		return true
	}
	// Also try matching against just the filename part
	_, filename := filepath.Split(path)
	matched, err = filepath.Match(pattern, filename)
	return err == nil && matched
}

// matchComponentGlob matches a simple glob pattern against path components.
func matchComponentGlob(pattern, path string) bool {
	pathComponents := strings.SplitSeq(path, "/")
	for component := range pathComponents {
		matched, err := filepath.Match(pattern, component)
		if err == nil && matched {
			return true
		}
	}
	// Also try matching the full path
	matched, err := filepath.Match(pattern, path)
	return err == nil && matched
}
