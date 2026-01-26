package parser

import (
	"context"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/core"
)

func TestWriter_WriteJSON(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		field       string
		newVersion  string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "simple version",
			content:     `{"version": "1.0.0"}`,
			field:       "version",
			newVersion:  "2.0.0",
			wantVersion: "2.0.0",
		},
		{
			name:        "nested field",
			content:     `{"package": {"version": "1.0.0"}}`,
			field:       "package.version",
			newVersion:  "2.0.0",
			wantVersion: "2.0.0",
		},
		{
			name:        "preserves other fields",
			content:     `{"name": "test", "version": "1.0.0"}`,
			field:       "version",
			newVersion:  "2.0.0",
			wantVersion: "2.0.0",
		},
		{
			name:       "empty field",
			content:    `{"version": "1.0.0"}`,
			field:      "",
			newVersion: "2.0.0",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test.json", []byte(tt.content))

			writer := NewWriter(fs)
			err := writer.Write(context.Background(), FileConfig{
				Path:   "/test.json",
				Format: FormatJSON,
				Field:  tt.field,
			}, tt.newVersion)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the version was written
			reader := NewReader(fs)
			result, err := reader.Read(context.Background(), FileConfig{
				Path:   "/test.json",
				Format: FormatJSON,
				Field:  tt.field,
			})
			if err != nil {
				t.Fatalf("failed to read back: %v", err)
			}

			if result.Version != tt.wantVersion {
				t.Errorf("got version %q, want %q", result.Version, tt.wantVersion)
			}
		})
	}
}

func TestWriter_WriteYAML(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		field       string
		newVersion  string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "simple version",
			content:     "version: 1.0.0\n",
			field:       "version",
			newVersion:  "2.0.0",
			wantVersion: "2.0.0",
		},
		{
			name:        "nested field",
			content:     "app:\n  version: 1.0.0\n",
			field:       "app.version",
			newVersion:  "2.0.0",
			wantVersion: "2.0.0",
		},
		{
			name:       "empty field",
			content:    "version: 1.0.0\n",
			field:      "",
			newVersion: "2.0.0",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test.yaml", []byte(tt.content))

			writer := NewWriter(fs)
			err := writer.Write(context.Background(), FileConfig{
				Path:   "/test.yaml",
				Format: FormatYAML,
				Field:  tt.field,
			}, tt.newVersion)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the version was written
			reader := NewReader(fs)
			result, err := reader.Read(context.Background(), FileConfig{
				Path:   "/test.yaml",
				Format: FormatYAML,
				Field:  tt.field,
			})
			if err != nil {
				t.Fatalf("failed to read back: %v", err)
			}

			if result.Version != tt.wantVersion {
				t.Errorf("got version %q, want %q", result.Version, tt.wantVersion)
			}
		})
	}
}

func TestWriter_WriteTOML(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		field       string
		newVersion  string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "cargo style",
			content:     "[package]\nname = \"test\"\nversion = \"1.0.0\"\n",
			field:       "package.version",
			newVersion:  "2.0.0",
			wantVersion: "2.0.0",
		},
		{
			name:        "pyproject style",
			content:     "[project]\nname = \"test\"\nversion = \"1.0.0\"\n",
			field:       "project.version",
			newVersion:  "2.0.0",
			wantVersion: "2.0.0",
		},
		{
			name:       "empty field",
			content:    "[package]\nversion = \"1.0.0\"\n",
			field:      "",
			newVersion: "2.0.0",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test.toml", []byte(tt.content))

			writer := NewWriter(fs)
			err := writer.Write(context.Background(), FileConfig{
				Path:   "/test.toml",
				Format: FormatTOML,
				Field:  tt.field,
			}, tt.newVersion)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the version was written
			reader := NewReader(fs)
			result, err := reader.Read(context.Background(), FileConfig{
				Path:   "/test.toml",
				Format: FormatTOML,
				Field:  tt.field,
			})
			if err != nil {
				t.Fatalf("failed to read back: %v", err)
			}

			if result.Version != tt.wantVersion {
				t.Errorf("got version %q, want %q", result.Version, tt.wantVersion)
			}
		})
	}
}

func TestWriter_WriteRaw(t *testing.T) {
	tests := []struct {
		name       string
		newVersion string
		wantSuffix string
	}{
		{
			name:       "simple version",
			newVersion: "1.2.3",
			wantSuffix: "1.2.3\n",
		},
		{
			name:       "with prefix",
			newVersion: "v1.2.3",
			wantSuffix: "v1.2.3\n",
		},
		{
			name:       "already has newline",
			newVersion: "1.2.3\n",
			wantSuffix: "1.2.3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/VERSION", []byte("0.0.0"))

			writer := NewWriter(fs)
			err := writer.Write(context.Background(), FileConfig{
				Path:   "/VERSION",
				Format: FormatRaw,
			}, tt.newVersion)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Read back the raw content
			data, ok := fs.GetFile("/VERSION")
			if !ok {
				t.Fatal("file not found after write")
			}

			if string(data) != tt.wantSuffix {
				t.Errorf("got content %q, want %q", string(data), tt.wantSuffix)
			}
		})
	}
}

func TestWriter_WriteRegex(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		pattern     string
		newVersion  string
		wantContain string
		wantErr     bool
	}{
		{
			name:        "go version constant",
			content:     `const Version = "1.0.0"`,
			pattern:     `Version\s*=\s*"([^"]+)"`,
			newVersion:  "2.0.0",
			wantContain: `"2.0.0"`,
		},
		{
			name:        "python version",
			content:     `__version__ = '1.0.0'`,
			pattern:     `__version__\s*=\s*'([^']+)'`,
			newVersion:  "2.0.0",
			wantContain: `'2.0.0'`,
		},
		{
			name:       "no match",
			content:    `const Name = "test"`,
			pattern:    `Version\s*=\s*"([^"]+)"`,
			newVersion: "2.0.0",
			wantErr:    true,
		},
		{
			name:       "invalid pattern",
			content:    `Version = "1.0.0"`,
			pattern:    `[invalid`,
			newVersion: "2.0.0",
			wantErr:    true,
		},
		{
			name:       "empty pattern",
			content:    `Version = "1.0.0"`,
			pattern:    "",
			newVersion: "2.0.0",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test.go", []byte(tt.content))

			writer := NewWriter(fs)
			err := writer.Write(context.Background(), FileConfig{
				Path:    "/test.go",
				Format:  FormatRegex,
				Pattern: tt.pattern,
			}, tt.newVersion)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the content contains the new version
			data, ok := fs.GetFile("/test.go")
			if !ok {
				t.Fatal("file not found after write")
			}

			if !strings.Contains(string(data), tt.wantContain) {
				t.Errorf("content %q does not contain %q", string(data), tt.wantContain)
			}
		})
	}
}

func TestWriter_FileNotFound(t *testing.T) {
	fs := core.NewMockFileSystem()
	writer := NewWriter(fs)

	err := writer.Write(context.Background(), FileConfig{
		Path:   "/nonexistent.json",
		Format: FormatJSON,
		Field:  "version",
	}, "1.0.0")

	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestWriter_EmptyPath(t *testing.T) {
	fs := core.NewMockFileSystem()
	writer := NewWriter(fs)

	err := writer.Write(context.Background(), FileConfig{
		Path:   "",
		Format: FormatJSON,
		Field:  "version",
	}, "1.0.0")

	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

func TestWriter_InvalidFormat(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test", []byte("1.0.0"))
	writer := NewWriter(fs)

	err := writer.Write(context.Background(), FileConfig{
		Path:   "/test",
		Format: Format("invalid"),
	}, "2.0.0")

	if err == nil {
		t.Error("expected error for invalid format, got nil")
	}
}

func TestWriter_Exists(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/exists.json", []byte(`{}`))

	writer := NewWriter(fs)

	if !writer.Exists(context.Background(), "/exists.json") {
		t.Error("expected Exists to return true for existing file")
	}

	if writer.Exists(context.Background(), "/nonexistent.json") {
		t.Error("expected Exists to return false for nonexistent file")
	}
}

func TestReadWriter(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/package.json", []byte(`{"version": "1.0.0"}`))

	rw := NewReadWriter(fs)

	// Test read
	version, err := rw.ReadVersion(context.Background(), FileConfig{
		Path:   "/package.json",
		Format: FormatJSON,
		Field:  "version",
	})
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if version != "1.0.0" {
		t.Errorf("got version %q, want %q", version, "1.0.0")
	}

	// Test write
	err = rw.Write(context.Background(), FileConfig{
		Path:   "/package.json",
		Format: FormatJSON,
		Field:  "version",
	}, "2.0.0")
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Verify
	version, err = rw.ReadVersion(context.Background(), FileConfig{
		Path:   "/package.json",
		Format: FormatJSON,
		Field:  "version",
	})
	if err != nil {
		t.Fatalf("read after write failed: %v", err)
	}
	if version != "2.0.0" {
		t.Errorf("got version %q, want %q", version, "2.0.0")
	}
}

func TestWriter_ContextCancellation(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test.json", []byte(`{"version": "1.0.0"}`))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	writer := NewWriter(fs)
	err := writer.Write(ctx, FileConfig{
		Path:   "/test.json",
		Format: FormatJSON,
		Field:  "version",
	}, "2.0.0")

	if err == nil {
		t.Error("expected error for canceled context, got nil")
	}
}

func TestSetNestedValue(t *testing.T) {
	tests := []struct {
		name    string
		initial map[string]any
		field   string
		value   string
		wantErr bool
	}{
		{
			name:    "create nested path",
			initial: map[string]any{},
			field:   "a.b.c",
			value:   "test",
		},
		{
			name:    "update existing",
			initial: map[string]any{"version": "1.0.0"},
			field:   "version",
			value:   "2.0.0",
		},
		{
			name:    "empty field",
			initial: map[string]any{},
			field:   "",
			value:   "test",
			wantErr: true,
		},
		{
			name:    "conflict with non-object",
			initial: map[string]any{"a": "string"},
			field:   "a.b",
			value:   "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setNestedValue(tt.initial, tt.field, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the value was set
			got, err := getNestedValue(tt.initial, tt.field)
			if err != nil {
				t.Fatalf("failed to get value: %v", err)
			}

			if got != tt.value {
				t.Errorf("got %v, want %v", got, tt.value)
			}
		})
	}
}
