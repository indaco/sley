package dependencycheck

import (
	"errors"
	"os"
	"testing"
)

// Note: getNestedValue and setNestedValue tests have been moved to internal/parser package
// where these functions now reside.

func TestReadWriteRawVersion(t *testing.T) {
	// Save original and restore
	originalRead := readFileFn
	originalWrite := writeFileFn
	defer func() {
		readFileFn = originalRead
		writeFileFn = originalWrite
	}()

	t.Run("read raw version", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte("1.2.3\n"), nil
		}

		version, err := readRawVersion("version.txt")
		if err != nil {
			t.Fatalf("readRawVersion() error = %v", err)
		}
		if version != "1.2.3" {
			t.Errorf("readRawVersion() = %q, want %q", version, "1.2.3")
		}
	})

	t.Run("read raw version - file error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return nil, errors.New("file not found")
		}

		_, err := readRawVersion("missing.txt")
		if err == nil {
			t.Error("readRawVersion() should return error for missing file")
		}
	})

	t.Run("write raw version", func(t *testing.T) {
		var written []byte
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			written = data
			return nil
		}

		err := writeRawVersion("version.txt", "1.2.4")
		if err != nil {
			t.Fatalf("writeRawVersion() error = %v", err)
		}
		if string(written) != "1.2.4\n" {
			t.Errorf("writeRawVersion() wrote %q, want %q", written, "1.2.4\n")
		}
	})

	t.Run("write raw version - adds newline", func(t *testing.T) {
		var written []byte
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			written = data
			return nil
		}

		err := writeRawVersion("version.txt", "1.2.4\n")
		if err != nil {
			t.Fatalf("writeRawVersion() error = %v", err)
		}
		// Should not double the newline
		if string(written) != "1.2.4\n" {
			t.Errorf("writeRawVersion() wrote %q, want %q", written, "1.2.4\n")
		}
	})

	t.Run("write raw version - write error", func(t *testing.T) {
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			return errors.New("write failed")
		}

		err := writeRawVersion("version.txt", "1.2.4")
		if err == nil {
			t.Error("writeRawVersion() should return error on write failure")
		}
	})
}

func TestReadWriteRegexVersion(t *testing.T) {
	// Save original and restore
	originalRead := readFileFn
	originalWrite := writeFileFn
	defer func() {
		readFileFn = originalRead
		writeFileFn = originalWrite
	}()

	t.Run("read regex version", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`version = "1.2.3"`), nil
		}

		version, err := readRegexVersion("file.txt", `version = "(.*?)"`)
		if err != nil {
			t.Fatalf("readRegexVersion() error = %v", err)
		}
		if version != "1.2.3" {
			t.Errorf("readRegexVersion() = %q, want %q", version, "1.2.3")
		}
	})

	t.Run("read regex version - no match", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`no version here`), nil
		}

		_, err := readRegexVersion("file.txt", `version = "(.*?)"`)
		if err == nil {
			t.Error("readRegexVersion() should return error when no match found")
		}
	})

	t.Run("read regex version - invalid pattern", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`version = "1.2.3"`), nil
		}

		_, err := readRegexVersion("file.txt", `[invalid(`)
		if err == nil {
			t.Error("readRegexVersion() should return error for invalid regex")
		}
	})

	t.Run("write regex version", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`version = "1.2.3"`), nil
		}

		var written []byte
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			written = data
			return nil
		}

		err := writeRegexVersion("file.txt", `version = "(.*?)"`, "1.2.4")
		if err != nil {
			t.Fatalf("writeRegexVersion() error = %v", err)
		}
		if string(written) != `version = "1.2.4"` {
			t.Errorf("writeRegexVersion() wrote %q, want %q", written, `version = "1.2.4"`)
		}
	})

	t.Run("write regex version - pattern not found", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`no version here`), nil
		}

		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			return nil
		}

		err := writeRegexVersion("file.txt", `version = "(.*?)"`, "1.2.4")
		if err == nil {
			t.Error("writeRegexVersion() should return error when pattern not found")
		}
	})

	t.Run("write regex version - invalid pattern", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`version = "1.2.3"`), nil
		}

		err := writeRegexVersion("file.txt", `[invalid(`, "1.2.4")
		if err == nil {
			t.Error("writeRegexVersion() should return error for invalid regex")
		}
	})
}

func TestReadJSONVersion(t *testing.T) {
	originalRead := readFileFn
	defer func() { readFileFn = originalRead }()

	tests := []struct {
		name      string
		content   string
		field     string
		want      string
		wantError bool
	}{
		{
			name:    "simple field",
			content: `{"version": "1.2.3"}`,
			field:   "version",
			want:    "1.2.3",
		},
		{
			name:    "nested field",
			content: `{"metadata": {"version": "2.0.0"}}`,
			field:   "metadata.version",
			want:    "2.0.0",
		},
		{
			name:      "invalid JSON",
			content:   `{invalid}`,
			field:     "version",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readFileFn = func(path string) ([]byte, error) {
				return []byte(tt.content), nil
			}

			version, err := readJSONVersion("file.json", tt.field)
			if tt.wantError {
				if err == nil {
					t.Error("readJSONVersion() should return error")
				}
				return
			}
			if err != nil {
				t.Fatalf("readJSONVersion() error = %v", err)
			}
			if version != tt.want {
				t.Errorf("readJSONVersion() = %q, want %q", version, tt.want)
			}
		})
	}
}

func TestWriteJSONVersion(t *testing.T) {
	originalRead := readFileFn
	originalWrite := writeFileFn
	defer func() {
		readFileFn = originalRead
		writeFileFn = originalWrite
	}()

	t.Run("simple field with trailing newline", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`{"version": "1.2.3"}`), nil
		}

		var written []byte
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			written = data
			return nil
		}

		err := writeJSONVersion("package.json", "version", "1.2.4")
		if err != nil {
			t.Fatalf("writeJSONVersion() error = %v", err)
		}
		if len(written) == 0 || written[len(written)-1] != '\n' {
			t.Error("writeJSONVersion() should write data with trailing newline")
		}
	})

	t.Run("preserves field order", func(t *testing.T) {
		input := `{
  "name": "my-package",
  "version": "1.0.0",
  "description": "A test package",
  "main": "index.js"
}
`
		readFileFn = func(path string) ([]byte, error) {
			return []byte(input), nil
		}

		var written []byte
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			written = data
			return nil
		}

		err := writeJSONVersion("package.json", "version", "1.0.1")
		if err != nil {
			t.Fatalf("writeJSONVersion() error = %v", err)
		}

		writtenStr := string(written)
		verifyJSONFieldOrder(t, writtenStr, []string{"name", "version", "description", "main"})
		verifyVersionUpdated(t, writtenStr, "1.0.1")
	})

	t.Run("nested field preserves structure", func(t *testing.T) {
		input := `{
  "name": "my-package",
  "metadata": {
    "author": "test",
    "version": "2.0.0",
    "license": "MIT"
  },
  "main": "index.js"
}
`
		readFileFn = func(path string) ([]byte, error) {
			return []byte(input), nil
		}

		var written []byte
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			written = data
			return nil
		}

		err := writeJSONVersion("package.json", "metadata.version", "2.0.1")
		if err != nil {
			t.Fatalf("writeJSONVersion() error = %v", err)
		}

		writtenStr := string(written)
		verifyVersionUpdated(t, writtenStr, "2.0.1")
		for _, field := range []string{"author", "license"} {
			if !containsSubstring(writtenStr, `"`+field+`"`) {
				t.Errorf("%s field missing after update", field)
			}
		}
	})
}

func verifyJSONFieldOrder(t *testing.T, content string, fields []string) {
	t.Helper()
	indices := make([]int, len(fields))
	for i, field := range fields {
		indices[i] = indexOfSubstring(content, `"`+field+`"`)
		if indices[i] == -1 {
			t.Fatalf("Missing expected field %q in output: %s", field, content)
		}
	}
	for i := 1; i < len(indices); i++ {
		if indices[i-1] >= indices[i] {
			t.Errorf("Field order not preserved: %s should come before %s", fields[i-1], fields[i])
		}
	}
}

func verifyVersionUpdated(t *testing.T, content, version string) {
	t.Helper()
	if !containsSubstring(content, `"version":"`+version+`"`) &&
		!containsSubstring(content, `"version": "`+version+`"`) {
		t.Errorf("Version not updated to %s. Output: %s", version, content)
	}
}

// Helper functions for string matching in tests
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func containsSubstring(s, substr string) bool {
	return indexOfSubstring(s, substr) != -1
}

func TestReadWriteYAMLVersion(t *testing.T) {
	// Save original and restore
	originalRead := readFileFn
	originalWrite := writeFileFn
	defer func() {
		readFileFn = originalRead
		writeFileFn = originalWrite
	}()

	t.Run("read YAML simple field", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte("version: 1.2.3\n"), nil
		}

		version, err := readYAMLVersion("Chart.yaml", "version")
		if err != nil {
			t.Fatalf("readYAMLVersion() error = %v", err)
		}
		if version != "1.2.3" {
			t.Errorf("readYAMLVersion() = %q, want %q", version, "1.2.3")
		}
	})

	t.Run("read YAML nested field", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte("metadata:\n  version: 2.0.0\n"), nil
		}

		version, err := readYAMLVersion("file.yaml", "metadata.version")
		if err != nil {
			t.Fatalf("readYAMLVersion() error = %v", err)
		}
		if version != "2.0.0" {
			t.Errorf("readYAMLVersion() = %q, want %q", version, "2.0.0")
		}
	})

	t.Run("write YAML simple field", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte("version: 1.2.3\n"), nil
		}

		var written []byte
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			written = data
			return nil
		}

		err := writeYAMLVersion("Chart.yaml", "version", "1.2.4")
		if err != nil {
			t.Fatalf("writeYAMLVersion() error = %v", err)
		}

		if len(written) == 0 {
			t.Error("writeYAMLVersion() wrote empty data")
		}
	})
}

func TestReadWriteTOMLVersion(t *testing.T) {
	// Save original and restore
	originalRead := readFileFn
	originalWrite := writeFileFn
	defer func() {
		readFileFn = originalRead
		writeFileFn = originalWrite
	}()

	t.Run("read TOML simple field", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`version = "1.2.3"`), nil
		}

		version, err := readTOMLVersion("pyproject.toml", "version")
		if err != nil {
			t.Fatalf("readTOMLVersion() error = %v", err)
		}
		if version != "1.2.3" {
			t.Errorf("readTOMLVersion() = %q, want %q", version, "1.2.3")
		}
	})

	t.Run("read TOML nested field", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte("[tool.poetry]\nversion = \"2.0.0\"\n"), nil
		}

		version, err := readTOMLVersion("pyproject.toml", "tool.poetry.version")
		if err != nil {
			t.Fatalf("readTOMLVersion() error = %v", err)
		}
		if version != "2.0.0" {
			t.Errorf("readTOMLVersion() = %q, want %q", version, "2.0.0")
		}
	})

	t.Run("write TOML simple field", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`version = "1.2.3"`), nil
		}

		var written []byte
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			written = data
			return nil
		}

		err := writeTOMLVersion("pyproject.toml", "version", "1.2.4")
		if err != nil {
			t.Fatalf("writeTOMLVersion() error = %v", err)
		}

		if len(written) == 0 {
			t.Error("writeTOMLVersion() wrote empty data")
		}
	})
}

// Additional error path tests for improved coverage

func TestReadJSONVersion_Errors(t *testing.T) {
	originalRead := readFileFn
	defer func() { readFileFn = originalRead }()

	t.Run("file read error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return nil, errors.New("file not found")
		}

		_, err := readJSONVersion("missing.json", "version")
		if err == nil {
			t.Error("readJSONVersion() should return error for file read failure")
		}
	})

	t.Run("non-string version field", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`{"version": 123}`), nil
		}

		_, err := readJSONVersion("file.json", "version")
		if err == nil {
			t.Error("readJSONVersion() should return error for non-string version")
		}
	})

	t.Run("field not found", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`{"name": "test"}`), nil
		}

		_, err := readJSONVersion("file.json", "version")
		if err == nil {
			t.Error("readJSONVersion() should return error for missing field")
		}
	})
}

func TestWriteJSONVersion_Errors(t *testing.T) {
	originalRead := readFileFn
	originalWrite := writeFileFn
	defer func() {
		readFileFn = originalRead
		writeFileFn = originalWrite
	}()

	t.Run("file read error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return nil, errors.New("file not found")
		}

		err := writeJSONVersion("missing.json", "version", "1.0.0")
		if err == nil {
			t.Error("writeJSONVersion() should return error for file read failure")
		}
	})

	t.Run("write error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`{"version": "1.0.0"}`), nil
		}
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			return errors.New("write failed")
		}

		err := writeJSONVersion("file.json", "version", "1.0.1")
		if err == nil {
			t.Error("writeJSONVersion() should return error for write failure")
		}
	})
}

func TestReadYAMLVersion_Errors(t *testing.T) {
	originalRead := readFileFn
	defer func() { readFileFn = originalRead }()

	t.Run("file read error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return nil, errors.New("file not found")
		}

		_, err := readYAMLVersion("missing.yaml", "version")
		if err == nil {
			t.Error("readYAMLVersion() should return error for file read failure")
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte("invalid: yaml: content:"), nil
		}

		_, err := readYAMLVersion("file.yaml", "version")
		if err == nil {
			t.Error("readYAMLVersion() should return error for invalid YAML")
		}
	})

	t.Run("non-string version field", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte("version: 123\n"), nil
		}

		_, err := readYAMLVersion("file.yaml", "version")
		if err == nil {
			t.Error("readYAMLVersion() should return error for non-string version")
		}
	})

	t.Run("field not found", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte("name: test\n"), nil
		}

		_, err := readYAMLVersion("file.yaml", "version")
		if err == nil {
			t.Error("readYAMLVersion() should return error for missing field")
		}
	})
}

func TestWriteYAMLVersion_Errors(t *testing.T) {
	originalRead := readFileFn
	originalWrite := writeFileFn
	defer func() {
		readFileFn = originalRead
		writeFileFn = originalWrite
	}()

	t.Run("file read error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return nil, errors.New("file not found")
		}

		err := writeYAMLVersion("missing.yaml", "version", "1.0.0")
		if err == nil {
			t.Error("writeYAMLVersion() should return error for file read failure")
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte("invalid: yaml: content:"), nil
		}

		err := writeYAMLVersion("file.yaml", "version", "1.0.0")
		if err == nil {
			t.Error("writeYAMLVersion() should return error for invalid YAML")
		}
	})

	t.Run("write error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte("version: 1.0.0\n"), nil
		}
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			return errors.New("write failed")
		}

		err := writeYAMLVersion("file.yaml", "version", "1.0.1")
		if err == nil {
			t.Error("writeYAMLVersion() should return error for write failure")
		}
	})
}

func TestReadTOMLVersion_Errors(t *testing.T) {
	originalRead := readFileFn
	defer func() { readFileFn = originalRead }()

	t.Run("file read error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return nil, errors.New("file not found")
		}

		_, err := readTOMLVersion("missing.toml", "version")
		if err == nil {
			t.Error("readTOMLVersion() should return error for file read failure")
		}
	})

	t.Run("invalid TOML", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`[invalid toml`), nil
		}

		_, err := readTOMLVersion("file.toml", "version")
		if err == nil {
			t.Error("readTOMLVersion() should return error for invalid TOML")
		}
	})

	t.Run("non-string version field", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`version = 123`), nil
		}

		_, err := readTOMLVersion("file.toml", "version")
		if err == nil {
			t.Error("readTOMLVersion() should return error for non-string version")
		}
	})

	t.Run("field not found", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`name = "test"`), nil
		}

		_, err := readTOMLVersion("file.toml", "version")
		if err == nil {
			t.Error("readTOMLVersion() should return error for missing field")
		}
	})
}

func TestWriteTOMLVersion_Errors(t *testing.T) {
	originalRead := readFileFn
	originalWrite := writeFileFn
	defer func() {
		readFileFn = originalRead
		writeFileFn = originalWrite
	}()

	t.Run("file read error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return nil, errors.New("file not found")
		}

		err := writeTOMLVersion("missing.toml", "version", "1.0.0")
		if err == nil {
			t.Error("writeTOMLVersion() should return error for file read failure")
		}
	})

	t.Run("invalid TOML", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`[invalid toml`), nil
		}

		err := writeTOMLVersion("file.toml", "version", "1.0.0")
		if err == nil {
			t.Error("writeTOMLVersion() should return error for invalid TOML")
		}
	})

	t.Run("write error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`version = "1.0.0"`), nil
		}
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			return errors.New("write failed")
		}

		err := writeTOMLVersion("file.toml", "version", "1.0.1")
		if err == nil {
			t.Error("writeTOMLVersion() should return error for write failure")
		}
	})
}

func TestReadRegexVersion_FileError(t *testing.T) {
	originalRead := readFileFn
	defer func() { readFileFn = originalRead }()

	readFileFn = func(path string) ([]byte, error) {
		return nil, errors.New("file not found")
	}

	_, err := readRegexVersion("missing.go", `Version = "(.*?)"`)
	if err == nil {
		t.Error("readRegexVersion() should return error for file read failure")
	}
}

func TestWriteRegexVersion_FileError(t *testing.T) {
	originalRead := readFileFn
	originalWrite := writeFileFn
	defer func() {
		readFileFn = originalRead
		writeFileFn = originalWrite
	}()

	t.Run("file read error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return nil, errors.New("file not found")
		}

		err := writeRegexVersion("missing.go", `Version = "(.*?)"`, "1.0.0")
		if err == nil {
			t.Error("writeRegexVersion() should return error for file read failure")
		}
	})

	t.Run("write error", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`Version = "1.0.0"`), nil
		}
		writeFileFn = func(path string, data []byte, perm os.FileMode) error {
			return errors.New("write failed")
		}

		err := writeRegexVersion("file.go", `Version = "(.*?)"`, "1.0.1")
		if err == nil {
			t.Error("writeRegexVersion() should return error for write failure")
		}
	})
}
