package testutils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFile(t *testing.T) {
	t.Parallel(
	// Create a temp file with content
	)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "hello world\n"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test reading the file
	got := ReadFile(t, testFile)
	if got != content {
		t.Errorf("ReadFile() = %q, want %q", got, content)
	}
}

func TestWriteFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "write-test.txt")
	content := "test content"

	WriteFile(t, testFile, content, 0644)

	// Verify the file was written
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if string(data) != content {
		t.Errorf("WriteFile() wrote %q, want %q", string(data), content)
	}
}

func TestWriteFile_WithExecutablePermissions(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "script.sh")
	content := "#!/bin/bash\necho hello"

	WriteFile(t, testFile, content, 0755)

	// Verify the file exists
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat written file: %v", err)
	}

	// Check executable bit (at least for owner)
	if info.Mode().Perm()&0100 == 0 {
		t.Error("WriteFile() should have set executable permission")
	}
}

func TestReadTempVersionFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Write a version file
	if err := os.WriteFile(versionPath, []byte("1.2.3\n"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	got := ReadTempVersionFile(t, tmpDir)
	if got != "1.2.3" {
		t.Errorf("ReadTempVersionFile() = %q, want %q", got, "1.2.3")
	}
}

func TestReadTempVersionFile_TrimsWhitespace(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Write a version file with extra whitespace
	if err := os.WriteFile(versionPath, []byte("  2.0.0-alpha  \n\n"), 0644); err != nil {
		t.Fatalf("failed to create version file: %v", err)
	}

	got := ReadTempVersionFile(t, tmpDir)
	if got != "2.0.0-alpha" {
		t.Errorf("ReadTempVersionFile() = %q, want %q", got, "2.0.0-alpha")
	}
}

func TestWriteTempVersionFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	path := WriteTempVersionFile(t, tmpDir, "3.0.0")

	// Verify path is correct
	expectedPath := filepath.Join(tmpDir, ".version")
	if path != expectedPath {
		t.Errorf("WriteTempVersionFile() path = %q, want %q", path, expectedPath)
	}

	// Verify content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read version file: %v", err)
	}

	if string(data) != "3.0.0" {
		t.Errorf("WriteTempVersionFile() wrote %q, want %q", string(data), "3.0.0")
	}
}

func TestWriteTempConfig(t *testing.T) {
	t.Parallel()
	configContent := `
version:
  path: .version
plugins:
  commit-parser: true
`

	path := WriteTempConfig(t, configContent)

	// Verify path ends with .sley.yaml
	if !strings.HasSuffix(path, ".sley.yaml") {
		t.Errorf("WriteTempConfig() path should end with .sley.yaml, got %q", path)
	}

	// Verify content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if string(data) != configContent {
		t.Errorf("WriteTempConfig() wrote %q, want %q", string(data), configContent)
	}
}
