package dependencycheck

import (
	"errors"
	"os"
	"testing"
)

func TestGetNestedValue(t *testing.T) {
	tests := []struct {
		name    string
		obj     map[string]any
		field   string
		want    any
		wantErr bool
	}{
		{
			name:    "simple field",
			obj:     map[string]any{"version": "1.2.3"},
			field:   "version",
			want:    "1.2.3",
			wantErr: false,
		},
		{
			name: "nested field",
			obj: map[string]any{
				"tool": map[string]any{
					"poetry": map[string]any{
						"version": "2.0.0",
					},
				},
			},
			field:   "tool.poetry.version",
			want:    "2.0.0",
			wantErr: false,
		},
		{
			name: "deeply nested field",
			obj: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": map[string]any{
							"d": "value",
						},
					},
				},
			},
			field:   "a.b.c.d",
			want:    "value",
			wantErr: false,
		},
		{
			name:    "field not found",
			obj:     map[string]any{"other": "value"},
			field:   "version",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty field path",
			obj:     map[string]any{"version": "1.2.3"},
			field:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name: "intermediate field is not object",
			obj: map[string]any{
				"tool": "string",
			},
			field:   "tool.poetry.version",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNestedValue(tt.obj, tt.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNestedValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("getNestedValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetNestedValue(t *testing.T) {
	tests := []struct {
		name    string
		obj     map[string]any
		field   string
		value   any
		wantErr bool
		check   func(map[string]any) bool
	}{
		{
			name:    "simple field",
			obj:     map[string]any{"version": "1.2.3"},
			field:   "version",
			value:   "1.2.4",
			wantErr: false,
			check: func(obj map[string]any) bool {
				return obj["version"] == "1.2.4"
			},
		},
		{
			name: "nested field - existing path",
			obj: map[string]any{
				"tool": map[string]any{
					"poetry": map[string]any{
						"version": "2.0.0",
					},
				},
			},
			field:   "tool.poetry.version",
			value:   "2.0.1",
			wantErr: false,
			check: func(obj map[string]any) bool {
				tool := obj["tool"].(map[string]any)
				poetry := tool["poetry"].(map[string]any)
				return poetry["version"] == "2.0.1"
			},
		},
		{
			name:    "nested field - create intermediate maps",
			obj:     map[string]any{},
			field:   "a.b.c",
			value:   "value",
			wantErr: false,
			check: func(obj map[string]any) bool {
				a := obj["a"].(map[string]any)
				b := a["b"].(map[string]any)
				return b["c"] == "value"
			},
		},
		{
			name:    "empty field path",
			obj:     map[string]any{},
			field:   "",
			value:   "value",
			wantErr: true,
			check:   nil,
		},
		{
			name: "intermediate field is not object",
			obj: map[string]any{
				"tool": "string",
			},
			field:   "tool.poetry.version",
			value:   "2.0.0",
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setNestedValue(tt.obj, tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("setNestedValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				if !tt.check(tt.obj) {
					t.Error("setNestedValue() did not set value correctly")
				}
			}
		})
	}
}

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

func TestReadWriteJSONVersion(t *testing.T) {
	// Save original and restore
	originalRead := readFileFn
	originalWrite := writeFileFn
	defer func() {
		readFileFn = originalRead
		writeFileFn = originalWrite
	}()

	t.Run("read JSON simple field", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`{"version": "1.2.3"}`), nil
		}

		version, err := readJSONVersion("package.json", "version")
		if err != nil {
			t.Fatalf("readJSONVersion() error = %v", err)
		}
		if version != "1.2.3" {
			t.Errorf("readJSONVersion() = %q, want %q", version, "1.2.3")
		}
	})

	t.Run("read JSON nested field", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`{"metadata": {"version": "2.0.0"}}`), nil
		}

		version, err := readJSONVersion("file.json", "metadata.version")
		if err != nil {
			t.Fatalf("readJSONVersion() error = %v", err)
		}
		if version != "2.0.0" {
			t.Errorf("readJSONVersion() = %q, want %q", version, "2.0.0")
		}
	})

	t.Run("read JSON invalid JSON", func(t *testing.T) {
		readFileFn = func(path string) ([]byte, error) {
			return []byte(`{invalid}`), nil
		}

		_, err := readJSONVersion("file.json", "version")
		if err == nil {
			t.Error("readJSONVersion() should return error for invalid JSON")
		}
	})

	t.Run("write JSON simple field", func(t *testing.T) {
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

		// Should have trailing newline
		if len(written) == 0 {
			t.Error("writeJSONVersion() wrote empty data")
		}
		if written[len(written)-1] != '\n' {
			t.Error("writeJSONVersion() should add trailing newline")
		}
	})

	t.Run("write JSON preserves field order", func(t *testing.T) {
		// package.json with specific field order: name, version, description, main
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

		// Verify the field order is preserved (name before version before description)
		writtenStr := string(written)

		nameIdx := indexOfSubstring(writtenStr, `"name"`)
		versionIdx := indexOfSubstring(writtenStr, `"version"`)
		descIdx := indexOfSubstring(writtenStr, `"description"`)
		mainIdx := indexOfSubstring(writtenStr, `"main"`)

		if nameIdx == -1 || versionIdx == -1 || descIdx == -1 || mainIdx == -1 {
			t.Fatalf("Missing expected fields in output: %s", writtenStr)
		}

		if nameIdx >= versionIdx || versionIdx >= descIdx || descIdx >= mainIdx {
			t.Errorf("Field order not preserved. Expected name < version < description < main, got indices: name=%d, version=%d, description=%d, main=%d",
				nameIdx, versionIdx, descIdx, mainIdx)
		}

		// Verify the version was actually updated
		if !containsSubstring(writtenStr, `"version":"1.0.1"`) && !containsSubstring(writtenStr, `"version": "1.0.1"`) {
			t.Errorf("Version not updated correctly. Output: %s", writtenStr)
		}
	})

	t.Run("write JSON nested field preserves structure", func(t *testing.T) {
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

		// Verify the nested version was updated
		if !containsSubstring(writtenStr, `"version":"2.0.1"`) && !containsSubstring(writtenStr, `"version": "2.0.1"`) {
			t.Errorf("Nested version not updated correctly. Output: %s", writtenStr)
		}

		// Verify other fields are still present
		if !containsSubstring(writtenStr, `"author"`) {
			t.Error("author field missing after update")
		}
		if !containsSubstring(writtenStr, `"license"`) {
			t.Error("license field missing after update")
		}
	})
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
