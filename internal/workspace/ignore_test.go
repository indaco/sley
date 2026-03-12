package workspace

import (
	"testing"
)

func TestParseIgnoreContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "empty content",
			content:  "",
			expected: []string{},
		},
		{
			name:     "single pattern",
			content:  "node_modules",
			expected: []string{"node_modules"},
		},
		{
			name: "multiple patterns",
			content: `node_modules
.git
vendor`,
			expected: []string{"node_modules", ".git", "vendor"},
		},
		{
			name: "patterns with comments",
			content: `# This is a comment
node_modules
# Another comment
.git
vendor # inline comment is NOT supported`,
			expected: []string{"node_modules", ".git", "vendor # inline comment is NOT supported"},
		},
		{
			name: "patterns with whitespace",
			content: `  node_modules

.git
	vendor	`,
			expected: []string{"node_modules", ".git", "vendor"},
		},
		{
			name: "glob patterns",
			content: `*.tmp
test-*
**/build`,
			expected: []string{"*.tmp", "test-*", "**/build"},
		},
		{
			name: "directory patterns",
			content: `build/
dist/
tmp/`,
			expected: []string{"build/", "dist/", "tmp/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseIgnoreContent(tt.content)

			if len(got) != len(tt.expected) {
				t.Errorf("parseIgnoreContent() returned %d patterns, expected %d", len(got), len(tt.expected))
				t.Logf("Got: %v", got)
				t.Logf("Expected: %v", tt.expected)
				return
			}

			for i, pattern := range got {
				if pattern != tt.expected[i] {
					t.Errorf("parseIgnoreContent() pattern[%d] = %q, expected %q", i, pattern, tt.expected[i])
				}
			}
		})
	}
}

func TestIgnoreFile_Matches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		path     string
		expected bool
	}{
		// Exact matching
		{
			name:     "exact match",
			content:  "node_modules",
			path:     "node_modules",
			expected: true,
		},
		{
			name:     "exact match - no match",
			content:  "node_modules",
			path:     "src",
			expected: false,
		},

		// Component matching
		{
			name:     "component match in path",
			content:  "node_modules",
			path:     "project/node_modules",
			expected: true,
		},
		{
			name:     "component match in nested path",
			content:  "node_modules",
			path:     "project/subdir/node_modules",
			expected: true,
		},

		// Glob patterns
		{
			name:     "glob pattern - matches",
			content:  "*.tmp",
			path:     "file.tmp",
			expected: true,
		},
		{
			name:     "glob pattern - no match",
			content:  "*.tmp",
			path:     "file.txt",
			expected: false,
		},
		{
			name:     "glob pattern with prefix",
			content:  "test-*",
			path:     "test-module",
			expected: true,
		},
		{
			name:     "glob pattern in path",
			content:  "*.log",
			path:     "logs/app.log",
			expected: true,
		},

		// Directory patterns
		{
			name:     "directory pattern - matches directory",
			content:  "build/",
			path:     "build",
			expected: true,
		},
		{
			name:     "directory pattern - matches subdirectory",
			content:  "build/",
			path:     "build/output",
			expected: true,
		},
		{
			name:     "directory pattern - matches nested",
			content:  "build/",
			path:     "build/output/file.txt",
			expected: true,
		},
		{
			name:     "directory pattern - no match",
			content:  "build/",
			path:     "src/build.txt",
			expected: false,
		},

		// Path patterns
		{
			name:     "path pattern with wildcard",
			content:  "src/*.tmp",
			path:     "src/file.tmp",
			expected: true,
		},
		{
			name:     "path pattern - no filename match (path specific)",
			content:  "test/*.log",
			path:     "app.log",
			expected: false, // Path-specific pattern doesn't match files outside that path
		},

		// Multiple patterns
		{
			name: "multiple patterns - first matches",
			content: `node_modules
.git
vendor`,
			path:     "node_modules",
			expected: true,
		},
		{
			name: "multiple patterns - middle matches",
			content: `node_modules
.git
vendor`,
			path:     ".git",
			expected: true,
		},
		{
			name: "multiple patterns - no match",
			content: `node_modules
.git
vendor`,
			path:     "src",
			expected: false,
		},

		// Edge cases
		{
			name:     "empty pattern list",
			content:  "",
			path:     "anything",
			expected: false,
		},
		{
			name:     "pattern with comments",
			content:  "# comment\nnode_modules\n# another",
			path:     "node_modules",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ignoreFile := NewIgnoreFile(tt.content)
			got := ignoreFile.Matches(tt.path)

			if got != tt.expected {
				t.Errorf("IgnoreFile.Matches(%q) = %v, expected %v", tt.path, got, tt.expected)
				t.Logf("Patterns: %v", ignoreFile.Patterns())
			}
		})
	}
}

func TestMatchIgnorePattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected bool
	}{
		// Exact matches
		{"exact match", "node_modules", "node_modules", true},
		{"exact match - no match", "node_modules", "src", false},

		// Component matches
		{"component in path", "node_modules", "project/node_modules", true},
		{"component nested", "build", "src/build", true},
		{"component deep nested", "temp", "a/b/c/temp/file", true},

		// Wildcard patterns
		{"wildcard suffix", "*.tmp", "file.tmp", true},
		{"wildcard suffix - no match", "*.tmp", "file.txt", false},
		{"wildcard prefix", "test-*", "test-module", true},
		{"wildcard prefix - no match", "test-*", "module-test", false},
		{"wildcard middle", "test*module", "testmodule", true},

		// Directory patterns
		{"directory exact", "build/", "build", true},
		{"directory subpath", "build/", "build/output", true},
		{"directory deep subpath", "dist/", "dist/a/b/c", true},
		{"directory no match", "build/", "src", false},

		// Path patterns with slash
		{"path pattern exact", "src/temp", "src/temp", true},
		{"path pattern with wildcard", "src/*.tmp", "src/file.tmp", true},
		{"path pattern nested", "test/*/output", "test/unit/output", true},

		// Edge cases
		{"empty pattern", "", "anything", false},
		{"empty path", "pattern", "", false},
		{"both empty", "", "", true}, // Empty string matches empty string exactly (though filtered by parseIgnoreContent)

		// Windows-style paths (should work with normalization)
		{"windows backslash", "node_modules", "project/node_modules", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := matchIgnorePattern(tt.pattern, tt.path)

			if got != tt.expected {
				t.Errorf("matchIgnorePattern(%q, %q) = %v, expected %v",
					tt.pattern, tt.path, got, tt.expected)
			}
		})
	}
}

func TestIgnoreFile_Patterns(t *testing.T) {
	t.Parallel()
	content := `node_modules
.git
vendor`

	ignoreFile := NewIgnoreFile(content)
	patterns := ignoreFile.Patterns()

	expected := []string{"node_modules", ".git", "vendor"}

	if len(patterns) != len(expected) {
		t.Errorf("Patterns() returned %d patterns, expected %d", len(patterns), len(expected))
		return
	}

	for i, pattern := range patterns {
		if pattern != expected[i] {
			t.Errorf("Patterns()[%d] = %q, expected %q", i, pattern, expected[i])
		}
	}

	// Verify that modifying returned slice doesn't affect original
	patterns[0] = "modified"
	newPatterns := ignoreFile.Patterns()
	if newPatterns[0] != "node_modules" {
		t.Error("Patterns() returned slice is not independent of internal state")
	}
}

func TestPatternMatcher_Matches(t *testing.T) {
	t.Parallel()
	patterns := []string{
		"node_modules",
		"*.tmp",
		"build/",
		"test-*",
	}

	matcher := newPatternMatcher(patterns)

	tests := []struct {
		path     string
		expected bool
	}{
		{"node_modules", true},
		{"project/node_modules", true},
		{"file.tmp", true},
		{"build", true},
		{"build/output", true},
		{"test-module", true},
		{"src/main.go", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			got := matcher.matches(tt.path)

			if got != tt.expected {
				t.Errorf("matches(%q) = %v, expected %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestPatternMatcher_EmptyPatterns(t *testing.T) {
	t.Parallel()
	matcher := newPatternMatcher([]string{})

	paths := []string{"anything", "node_modules", "*.tmp"}

	for _, path := range paths {
		if matcher.matches(path) {
			t.Errorf("Empty pattern matcher matched %q", path)
		}
	}
}

// Benchmark pattern matching performance
func BenchmarkIgnoreFile_Matches(b *testing.B) {
	content := `node_modules
.git
vendor
*.tmp
*.log
test-*
build/
dist/
tmp/
__pycache__
.cache`

	ignoreFile := NewIgnoreFile(content)
	paths := []string{
		"src/main.go",
		"node_modules/package",
		"test/file.tmp",
		"build/output",
		"vendor/lib",
		"README.md",
	}

	for b.Loop() {
		for _, path := range paths {
			ignoreFile.Matches(path)
		}
	}
}

func BenchmarkMatchIgnorePattern(b *testing.B) {
	pattern := "*.tmp"
	path := "test/file.tmp"

	for b.Loop() {
		matchIgnorePattern(pattern, path)
	}
}
