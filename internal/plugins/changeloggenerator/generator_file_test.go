package changeloggenerator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helper functions to reduce cyclomatic complexity

func setupTestChangesDir(t *testing.T) (tmpDir, changesDir string) {
	t.Helper()
	tmpDir = t.TempDir()
	changesDir = filepath.Join(tmpDir, ".changes")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("failed to create changes dir: %v", err)
	}
	return tmpDir, changesDir
}

func createVersionFiles(t *testing.T, changesDir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(changesDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", name, err)
		}
	}
}

func createTestGenerator(t *testing.T, changesDir, changelogPath string) *Generator {
	t.Helper()
	cfg := DefaultConfig()
	cfg.ChangesDir = changesDir
	cfg.ChangelogPath = changelogPath
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}
	return g
}

func readChangelogContent(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read changelog: %v", err)
	}
	return content
}

func mergeAndVerifyIdempotent(t *testing.T, g *Generator, changelogPath string, firstContent []byte, iteration int) {
	t.Helper()
	if err := g.MergeVersionedFiles(); err != nil {
		t.Fatalf("merge #%d failed: %v", iteration, err)
	}
	content := readChangelogContent(t, changelogPath)
	if string(content) != string(firstContent) {
		t.Errorf("merge #%d produced different content", iteration)
	}
}

func TestWriteVersionedFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.ChangesDir = filepath.Join(tmpDir, ".changes")
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := "## v1.0.0\n\nTest content"
	err = g.WriteVersionedFile("v1.0.0", content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check file exists
	expectedPath := filepath.Join(cfg.ChangesDir, "v1.0.0.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s", expectedPath)
	}

	// Check content - file should be normalized with single trailing newline
	data, readErr := os.ReadFile(expectedPath)
	if readErr != nil {
		t.Fatalf("failed to read file: %v", readErr)
	}
	expectedContent := "## v1.0.0\n\nTest content\n"
	if string(data) != expectedContent {
		t.Errorf("file content = %q, want %q", string(data), expectedContent)
	}
}

func TestWriteVersionedFile_Error(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ChangesDir = "/nonexistent/readonly/path"
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = g.WriteVersionedFile("v1.0.0", "content")
	if err == nil {
		t.Error("expected error for non-writable path")
	}
}

func TestWriteUnifiedChangelog_New(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.ChangelogPath = filepath.Join(tmpDir, "CHANGELOG.md")
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	newContent := "## v1.0.0\n\nNew content"
	err = g.WriteUnifiedChangelog(newContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(cfg.ChangelogPath); os.IsNotExist(err) {
		t.Error("expected CHANGELOG.md to be created")
	}

	// Check content includes header
	data, readErr := os.ReadFile(cfg.ChangelogPath)
	if readErr != nil {
		t.Fatalf("failed to read file: %v", readErr)
	}
	content := string(data)
	if !strings.Contains(content, "# Changelog") {
		t.Error("expected changelog header")
	}
	if !strings.Contains(content, "v1.0.0") {
		t.Error("expected version content")
	}
}

func TestWriteUnifiedChangelog_Existing(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.ChangelogPath = filepath.Join(tmpDir, "CHANGELOG.md")
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Create existing changelog
	existingContent := `# Changelog

## v0.9.0

Previous content
`
	if err := os.WriteFile(cfg.ChangelogPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to create existing changelog: %v", err)
	}

	// Write new content
	newContent := "## v1.0.0\n\nNew content\n\n"
	err = g.WriteUnifiedChangelog(newContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check content
	data, readErr := os.ReadFile(cfg.ChangelogPath)
	if readErr != nil {
		t.Fatalf("failed to read file: %v", readErr)
	}
	content := string(data)

	// New content should come before old
	v1Index := strings.Index(content, "v1.0.0")
	v09Index := strings.Index(content, "v0.9.0")
	if v1Index > v09Index {
		t.Error("expected new version to appear before old version")
	}
}

func TestWriteUnifiedChangelog_Error(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ChangelogPath = "/nonexistent/readonly/CHANGELOG.md"
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = g.WriteUnifiedChangelog("content")
	if err == nil {
		t.Error("expected error for non-writable path")
	}
}

func TestMergeVersionedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("failed to create changes dir: %v", err)
	}

	// Create version files
	files := map[string]string{
		"v0.1.0.md": "## v0.1.0\n\nFirst version\n",
		"v0.2.0.md": "## v0.2.0\n\nSecond version\n",
	}
	for name, content := range files {
		path := filepath.Join(changesDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	cfg := DefaultConfig()
	cfg.ChangesDir = changesDir
	cfg.ChangelogPath = filepath.Join(tmpDir, "CHANGELOG.md")
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = g.MergeVersionedFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check merged file
	data, readErr := os.ReadFile(cfg.ChangelogPath)
	if readErr != nil {
		t.Fatalf("failed to read changelog: %v", readErr)
	}
	content := string(data)

	if !strings.Contains(content, "v0.1.0") {
		t.Error("expected v0.1.0 in merged changelog")
	}
	if !strings.Contains(content, "v0.2.0") {
		t.Error("expected v0.2.0 in merged changelog")
	}
}

func TestMergeVersionedFiles_SemanticOrder(t *testing.T) {
	tmpDir, changesDir := setupTestChangesDir(t)

	createVersionFiles(t, changesDir, map[string]string{
		"v0.1.0.md":  "## v0.1.0\n\nFirst version\n",
		"v0.1.2.md":  "## v0.1.2\n\nPatch release\n",
		"v0.2.0.md":  "## v0.2.0\n\nSecond minor\n",
		"v0.9.1.md":  "## v0.9.1\n\nNinth minor patch\n",
		"v0.10.0.md": "## v0.10.0\n\nTenth minor\n",
	})

	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")
	g := createTestGenerator(t, changesDir, changelogPath)

	if err := g.MergeVersionedFiles(); err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	content := string(readChangelogContent(t, changelogPath))

	// v0.10.0 must appear before v0.9.1 (not between v0.1.2 and v0.2.0)
	pos010 := strings.Index(content, "## v0.10.0")
	pos091 := strings.Index(content, "## v0.9.1")
	pos020 := strings.Index(content, "## v0.2.0")
	pos012 := strings.Index(content, "## v0.1.2")
	pos010v := strings.Index(content, "## v0.1.0")

	if pos010 == -1 || pos091 == -1 || pos020 == -1 || pos012 == -1 || pos010v == -1 {
		t.Fatalf("not all versions found in merged content:\n%s", content)
	}

	// Correct descending order: v0.10.0 > v0.9.1 > v0.2.0 > v0.1.2 > v0.1.0
	if pos010 > pos091 {
		t.Errorf("v0.10.0 (pos %d) should appear before v0.9.1 (pos %d)", pos010, pos091)
	}
	if pos091 > pos020 {
		t.Errorf("v0.9.1 (pos %d) should appear before v0.2.0 (pos %d)", pos091, pos020)
	}
	if pos020 > pos012 {
		t.Errorf("v0.2.0 (pos %d) should appear before v0.1.2 (pos %d)", pos020, pos012)
	}
	if pos012 > pos010v {
		t.Errorf("v0.1.2 (pos %d) should appear before v0.1.0 (pos %d)", pos012, pos010v)
	}
}

func TestMergeVersionedFiles_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("failed to create changes dir: %v", err)
	}

	cfg := DefaultConfig()
	cfg.ChangesDir = changesDir
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not error with empty directory
	err = g.MergeVersionedFiles()
	if err != nil {
		t.Errorf("unexpected error for empty dir: %v", err)
	}
}

func TestMergeVersionedFiles_NonexistentDir(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ChangesDir = "/nonexistent/path"
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should error with non-existent directory
	err = g.MergeVersionedFiles()
	if err == nil {
		t.Error("expected error for non-existent dir")
	}
}

func TestMergeVersionedFiles_Idempotent(t *testing.T) {
	tmpDir, changesDir := setupTestChangesDir(t)
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	createVersionFiles(t, changesDir, map[string]string{
		"v0.1.0.md": "## v0.1.0\n\n### Added\n- Initial release\n",
		"v0.2.0.md": "## v0.2.0\n\n### Fixed\n- Bug fix\n\n### Added\n- New feature\n",
		"v1.0.0.md": "## v1.0.0\n\n### Breaking Changes\n- Major update\n",
	})

	g := createTestGenerator(t, changesDir, changelogPath)

	// First merge
	if err := g.MergeVersionedFiles(); err != nil {
		t.Fatalf("first merge failed: %v", err)
	}
	firstContent := readChangelogContent(t, changelogPath)

	// Run merge multiple times and verify idempotency
	for i := 2; i <= 5; i++ {
		mergeAndVerifyIdempotent(t, g, changelogPath, firstContent, i)
	}

	// Verify version ordering (newest first)
	verifyVersionOrder(t, string(firstContent))
}

func verifyVersionOrder(t *testing.T, content string) {
	t.Helper()
	v1Pos := strings.Index(content, "v1.0.0")
	v02Pos := strings.Index(content, "v0.2.0")
	v01Pos := strings.Index(content, "v0.1.0")

	if v1Pos == -1 || v02Pos == -1 || v01Pos == -1 {
		t.Error("not all versions found in merged content")
		return
	}

	if v1Pos > v02Pos || v02Pos > v01Pos {
		t.Error("versions not in correct order (newest first)")
	}
}

func TestMergeVersionedFiles_IdempotentWithExistingChangelog(t *testing.T) {
	tmpDir, changesDir := setupTestChangesDir(t)
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Create an existing changelog with different content
	existingChangelog := "# Changelog\n\nOld header content.\n\n## v0.0.1\n\n- Old version that should be replaced\n"
	if err := os.WriteFile(changelogPath, []byte(existingChangelog), 0644); err != nil {
		t.Fatalf("failed to write existing changelog: %v", err)
	}

	createVersionFiles(t, changesDir, map[string]string{
		"v0.1.0.md": "## v0.1.0\n\n### Added\n- Initial release\n",
		"v0.2.0.md": "## v0.2.0\n\n### Fixed\n- Bug fix\n",
	})

	g := createTestGenerator(t, changesDir, changelogPath)

	// First merge - should overwrite existing content
	if err := g.MergeVersionedFiles(); err != nil {
		t.Fatalf("first merge failed: %v", err)
	}
	firstContent := readChangelogContent(t, changelogPath)

	// Verify old content is gone
	verifyOldContentRemoved(t, string(firstContent))

	// Run merge twice more and verify idempotency
	mergeAndVerifyIdempotent(t, g, changelogPath, firstContent, 2)
	mergeAndVerifyIdempotent(t, g, changelogPath, firstContent, 3)
}

func verifyOldContentRemoved(t *testing.T, content string) {
	t.Helper()
	if strings.Contains(content, "v0.0.1") {
		t.Error("old version v0.0.1 should not be in merged changelog")
	}
	if strings.Contains(content, "Old header") {
		t.Error("old header should not be in merged changelog")
	}
}

func TestCollectVersionFiles(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("failed to create changes dir: %v", err)
	}

	// Create version files and other files
	files := map[string]string{
		"v1.0.0.md": "version 1",
		"v0.1.0.md": "version 0.1",
		"README.md": "not a version",
		"other.txt": "not markdown",
		"notes.md":  "not starting with v",
	}
	for name, content := range files {
		path := filepath.Join(changesDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	// Create a subdirectory (should be skipped)
	subdir := filepath.Join(changesDir, "v2.0.0")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	collected, err := collectVersionFiles(changesDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have v1.0.0.md and v0.1.0.md
	if len(collected) != 2 {
		t.Errorf("expected 2 version files, got %d", len(collected))
	}

	// Check files are the right ones
	hasV1 := false
	hasV01 := false
	for _, f := range collected {
		if strings.Contains(f, "v1.0.0.md") {
			hasV1 = true
		}
		if strings.Contains(f, "v0.1.0.md") {
			hasV01 = true
		}
	}
	if !hasV1 || !hasV01 {
		t.Error("expected both v1.0.0.md and v0.1.0.md in collected files")
	}
}

func TestBuildMergedContent(t *testing.T) {
	tmpDir := t.TempDir()
	changesDir := filepath.Join(tmpDir, ".changes")
	if err := os.MkdirAll(changesDir, 0755); err != nil {
		t.Fatalf("failed to create changes dir: %v", err)
	}

	// Create version files
	v1Content := "## v1.0.0\n\nFirst release"
	v2Content := "## v2.0.0\n\nSecond release"
	if err := os.WriteFile(filepath.Join(changesDir, "v1.0.0.md"), []byte(v1Content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changesDir, "v2.0.0.md"), []byte(v2Content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	cfg := DefaultConfig()
	cfg.ChangesDir = changesDir
	g, _ := NewGenerator(cfg)

	files := []string{
		filepath.Join(changesDir, "v2.0.0.md"),
		filepath.Join(changesDir, "v1.0.0.md"),
	}

	content := g.buildMergedContent(files)

	if !strings.Contains(content, "# Changelog") {
		t.Error("expected header in merged content")
	}
	if !strings.Contains(content, "v1.0.0") {
		t.Error("expected v1.0.0 in merged content")
	}
	if !strings.Contains(content, "v2.0.0") {
		t.Error("expected v2.0.0 in merged content")
	}
}

func TestGetDefaultHeader(t *testing.T) {
	cfg := DefaultConfig()
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	header := g.getDefaultHeader()

	if !strings.Contains(header, "Changelog") {
		t.Error("expected 'Changelog' in header")
	}
	if !strings.Contains(header, "Semantic Versioning") {
		t.Error("expected 'Semantic Versioning' in header")
	}
}

func TestGetDefaultHeader_CustomTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "header.md")
	customHeader := "# Custom Header\n\nCustom description"
	if err := os.WriteFile(templatePath, []byte(customHeader), 0644); err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	cfg := DefaultConfig()
	cfg.HeaderTemplate = templatePath
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	header := g.getDefaultHeader()

	if header != strings.TrimSpace(customHeader) {
		t.Errorf("header = %q, want %q", header, strings.TrimSpace(customHeader))
	}
}

func TestInsertAfterHeader(t *testing.T) {
	cfg := DefaultConfig()
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	existing := `# Changelog

Some description

## v0.9.0

Old content
`
	newContent := "## v1.0.0\n\nNew content\n\n"

	result := g.insertAfterHeader(existing, newContent)

	// New content should be before v0.9.0
	v1Index := strings.Index(result, "v1.0.0")
	v09Index := strings.Index(result, "v0.9.0")
	if v1Index > v09Index {
		t.Error("expected new version to appear before old version")
	}
}

func TestInsertAfterHeader_NoVersionFound(t *testing.T) {
	cfg := DefaultConfig()
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Existing content with no version headers
	existing := `# Changelog

Some description about this project.
`
	newContent := "## v1.0.0\n\nNew content\n\n"

	result := g.insertAfterHeader(existing, newContent)

	// New content should be appended
	if !strings.Contains(result, "v1.0.0") {
		t.Error("expected new version in result")
	}
}

func TestSortVersionFiles(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "basic ordering",
			input: []string{
				"/tmp/.changes/v0.1.0.md",
				"/tmp/.changes/v1.0.0.md",
				"/tmp/.changes/v0.9.0.md",
			},
			expected: []string{
				"/tmp/.changes/v1.0.0.md",
				"/tmp/.changes/v0.9.0.md",
				"/tmp/.changes/v0.1.0.md",
			},
		},
		{
			name: "double digit minor version sorted semantically not lexicographically",
			input: []string{
				"/tmp/.changes/v0.1.0.md",
				"/tmp/.changes/v0.1.2.md",
				"/tmp/.changes/v0.2.0.md",
				"/tmp/.changes/v0.10.0.md",
				"/tmp/.changes/v0.9.1.md",
			},
			expected: []string{
				"/tmp/.changes/v0.10.0.md",
				"/tmp/.changes/v0.9.1.md",
				"/tmp/.changes/v0.2.0.md",
				"/tmp/.changes/v0.1.2.md",
				"/tmp/.changes/v0.1.0.md",
			},
		},
		{
			name: "mixed major and minor versions",
			input: []string{
				"/tmp/.changes/v0.1.0.md",
				"/tmp/.changes/v2.0.0.md",
				"/tmp/.changes/v0.10.0.md",
				"/tmp/.changes/v1.0.0.md",
				"/tmp/.changes/v1.10.0.md",
				"/tmp/.changes/v1.9.0.md",
			},
			expected: []string{
				"/tmp/.changes/v2.0.0.md",
				"/tmp/.changes/v1.10.0.md",
				"/tmp/.changes/v1.9.0.md",
				"/tmp/.changes/v1.0.0.md",
				"/tmp/.changes/v0.10.0.md",
				"/tmp/.changes/v0.1.0.md",
			},
		},
		{
			name: "single element",
			input: []string{
				"/tmp/.changes/v1.0.0.md",
			},
			expected: []string{
				"/tmp/.changes/v1.0.0.md",
			},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Copy to avoid mutating test data
			files := make([]string, len(tt.input))
			copy(files, tt.input)

			sortVersionFiles(files)

			if len(files) != len(tt.expected) {
				t.Fatalf("expected %d files, got %d", len(tt.expected), len(files))
			}
			for i, want := range tt.expected {
				if files[i] != want {
					t.Errorf("position %d: expected %s, got %s", i, want, files[i])
				}
			}
		})
	}
}

func TestSortVersionFiles_SemanticOrder_0_10_0(t *testing.T) {
	// Regression test: v0.10.0 must sort AFTER v0.9.1 and not between v0.1.2 and v0.2.0.
	// This was the original bug: lexicographic comparison treated "10" < "2" because
	// "1" < "2" character-by-character.
	files := []string{
		"/tmp/.changes/v0.1.2.md",
		"/tmp/.changes/v0.2.0.md",
		"/tmp/.changes/v0.10.0.md",
		"/tmp/.changes/v0.9.1.md",
	}

	sortVersionFiles(files)

	// Newest first (descending semantic order)
	expected := []string{
		"/tmp/.changes/v0.10.0.md",
		"/tmp/.changes/v0.9.1.md",
		"/tmp/.changes/v0.2.0.md",
		"/tmp/.changes/v0.1.2.md",
	}

	for i, want := range expected {
		if files[i] != want {
			t.Errorf("position %d: expected %s, got %s", i, want, files[i])
		}
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		path    string
		wantStr string
	}{
		{"/tmp/.changes/v1.2.3.md", "1.2.3"},
		{"/tmp/.changes/v0.10.0.md", "0.10.0"},
		{"v0.1.0.md", "0.1.0"},
		{"/some/path/v2.0.0-rc.1.md", "2.0.0-rc.1"},
		// Unparseable returns zero value
		{"/tmp/.changes/README.md", "0.0.0"},
		{"/tmp/.changes/not-version.md", "0.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			v := extractVersion(tt.path)
			got := v.String()
			if got != tt.wantStr {
				t.Errorf("extractVersion(%q) = %q, want %q", tt.path, got, tt.wantStr)
			}
		})
	}
}
