package parser

import (
	"context"
	"testing"

	"github.com/indaco/sley/internal/core"
)

func TestFormat_IsValid(t *testing.T) {
	tests := []struct {
		format Format
		want   bool
	}{
		{FormatJSON, true},
		{FormatYAML, true},
		{FormatTOML, true},
		{FormatRaw, true},
		{FormatRegex, true},
		{Format("invalid"), false},
		{Format(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			got := tt.format.IsValid()
			if got != tt.want {
				t.Errorf("Format(%q).IsValid() = %v, want %v", tt.format, got, tt.want)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		want  Format
	}{
		{"json", FormatJSON},
		{"yaml", FormatYAML},
		{"toml", FormatTOML},
		{"raw", FormatRaw},
		{"regex", FormatRegex},
		{"invalid", FormatRaw}, // Fallback
		{"", FormatRaw},        // Fallback
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseFormat(tt.input)
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestReader_ReadJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
		field   string
		want    string
		wantErr bool
	}{
		{
			name:    "simple version",
			content: `{"version": "1.2.3"}`,
			field:   "version",
			want:    "1.2.3",
		},
		{
			name:    "nested field",
			content: `{"package": {"version": "2.0.0"}}`,
			field:   "package.version",
			want:    "2.0.0",
		},
		{
			name:    "deeply nested",
			content: `{"tool": {"poetry": {"version": "3.0.0-alpha.1"}}}`,
			field:   "tool.poetry.version",
			want:    "3.0.0-alpha.1",
		},
		{
			name:    "field not found",
			content: `{"name": "test"}`,
			field:   "version",
			wantErr: true,
		},
		{
			name:    "non-string version",
			content: `{"version": 123}`,
			field:   "version",
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			content: `{invalid`,
			field:   "version",
			wantErr: true,
		},
		{
			name:    "empty field",
			content: `{"version": "1.0.0"}`,
			field:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test.json", []byte(tt.content))

			reader := NewReader(fs)
			result, err := reader.Read(context.Background(), FileConfig{
				Path:   "/test.json",
				Format: FormatJSON,
				Field:  tt.field,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Version != tt.want {
				t.Errorf("got version %q, want %q", result.Version, tt.want)
			}
		})
	}
}

func TestReader_ReadYAML(t *testing.T) {
	tests := []struct {
		name    string
		content string
		field   string
		want    string
		wantErr bool
	}{
		{
			name:    "simple version",
			content: "version: 1.2.3\n",
			field:   "version",
			want:    "1.2.3",
		},
		{
			name:    "nested field",
			content: "app:\n  version: 2.0.0\n",
			field:   "app.version",
			want:    "2.0.0",
		},
		{
			name:    "chart yaml",
			content: "apiVersion: v2\nname: myapp\nversion: 0.1.0\n",
			field:   "version",
			want:    "0.1.0",
		},
		{
			name:    "field not found",
			content: "name: test\n",
			field:   "version",
			wantErr: true,
		},
		{
			name:    "invalid YAML",
			content: "invalid: [unclosed",
			field:   "version",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test.yaml", []byte(tt.content))

			reader := NewReader(fs)
			result, err := reader.Read(context.Background(), FileConfig{
				Path:   "/test.yaml",
				Format: FormatYAML,
				Field:  tt.field,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Version != tt.want {
				t.Errorf("got version %q, want %q", result.Version, tt.want)
			}
		})
	}
}

func TestReader_ReadTOML(t *testing.T) {
	tests := []struct {
		name    string
		content string
		field   string
		want    string
		wantErr bool
	}{
		{
			name:    "cargo toml style",
			content: "[package]\nname = \"test\"\nversion = \"1.2.3\"\n",
			field:   "package.version",
			want:    "1.2.3",
		},
		{
			name:    "pyproject style",
			content: "[project]\nname = \"test\"\nversion = \"2.0.0\"\n",
			field:   "project.version",
			want:    "2.0.0",
		},
		{
			name:    "poetry style",
			content: "[tool.poetry]\nname = \"test\"\nversion = \"3.0.0\"\n",
			field:   "tool.poetry.version",
			want:    "3.0.0",
		},
		{
			name:    "field not found",
			content: "[package]\nname = \"test\"\n",
			field:   "package.version",
			wantErr: true,
		},
		{
			name:    "invalid TOML",
			content: "[invalid",
			field:   "version",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test.toml", []byte(tt.content))

			reader := NewReader(fs)
			result, err := reader.Read(context.Background(), FileConfig{
				Path:   "/test.toml",
				Format: FormatTOML,
				Field:  tt.field,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Version != tt.want {
				t.Errorf("got version %q, want %q", result.Version, tt.want)
			}
		})
	}
}

func TestReader_ReadRaw(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "simple version",
			content: "1.2.3",
			want:    "1.2.3",
		},
		{
			name:    "with newline",
			content: "1.2.3\n",
			want:    "1.2.3",
		},
		{
			name:    "with prefix",
			content: "v1.2.3\n",
			want:    "v1.2.3",
		},
		{
			name:    "with whitespace",
			content: "  1.2.3  \n",
			want:    "1.2.3",
		},
		{
			name:    "prerelease",
			content: "1.0.0-alpha.1\n",
			want:    "1.0.0-alpha.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/VERSION", []byte(tt.content))

			reader := NewReader(fs)
			result, err := reader.Read(context.Background(), FileConfig{
				Path:   "/VERSION",
				Format: FormatRaw,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Version != tt.want {
				t.Errorf("got version %q, want %q", result.Version, tt.want)
			}
		})
	}
}

func TestReader_ReadRegex(t *testing.T) {
	tests := []struct {
		name    string
		content string
		pattern string
		want    string
		wantErr bool
	}{
		{
			name:    "go version constant",
			content: `package version\n\nconst Version = "1.2.3"\n`,
			pattern: `Version\s*=\s*"([^"]+)"`,
			want:    "1.2.3",
		},
		{
			name:    "python version",
			content: `__version__ = '2.0.0'`,
			pattern: `__version__\s*=\s*'([^']+)'`,
			want:    "2.0.0",
		},
		{
			name:    "no match",
			content: `const Name = "test"`,
			pattern: `Version\s*=\s*"([^"]+)"`,
			wantErr: true,
		},
		{
			name:    "no capturing group",
			content: `Version = "1.0.0"`,
			pattern: `Version = "[^"]+"`,
			wantErr: true,
		},
		{
			name:    "invalid regex",
			content: `Version = "1.0.0"`,
			pattern: `[invalid`,
			wantErr: true,
		},
		{
			name:    "empty pattern",
			content: `Version = "1.0.0"`,
			pattern: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test.go", []byte(tt.content))

			reader := NewReader(fs)
			result, err := reader.Read(context.Background(), FileConfig{
				Path:    "/test.go",
				Format:  FormatRegex,
				Pattern: tt.pattern,
			})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Version != tt.want {
				t.Errorf("got version %q, want %q", result.Version, tt.want)
			}
		})
	}
}

func TestReader_FileNotFound(t *testing.T) {
	fs := core.NewMockFileSystem()
	reader := NewReader(fs)

	_, err := reader.Read(context.Background(), FileConfig{
		Path:   "/nonexistent.json",
		Format: FormatJSON,
		Field:  "version",
	})

	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestReader_EmptyPath(t *testing.T) {
	fs := core.NewMockFileSystem()
	reader := NewReader(fs)

	_, err := reader.Read(context.Background(), FileConfig{
		Path:   "",
		Format: FormatJSON,
		Field:  "version",
	})

	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

func TestReader_InvalidFormat(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test", []byte("1.0.0"))
	reader := NewReader(fs)

	_, err := reader.Read(context.Background(), FileConfig{
		Path:   "/test",
		Format: Format("invalid"),
	})

	if err == nil {
		t.Error("expected error for invalid format, got nil")
	}
}

func TestReader_ReadVersion(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/package.json", []byte(`{"version": "1.2.3"}`))

	reader := NewReader(fs)
	version, err := reader.ReadVersion(context.Background(), FileConfig{
		Path:   "/package.json",
		Format: FormatJSON,
		Field:  "version",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if version != "1.2.3" {
		t.Errorf("got version %q, want %q", version, "1.2.3")
	}
}

func TestReader_ContextCancellation(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test.json", []byte(`{"version": "1.0.0"}`))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	reader := NewReader(fs)
	_, err := reader.Read(ctx, FileConfig{
		Path:   "/test.json",
		Format: FormatJSON,
		Field:  "version",
	})

	if err == nil {
		t.Error("expected error for canceled context, got nil")
	}
}

func TestFormatForFile(t *testing.T) {
	tests := []struct {
		filename string
		want     Format
	}{
		{"package.json", FormatJSON},
		{"config.json", FormatJSON},
		{"Chart.yaml", FormatYAML},
		{"config.yml", FormatYAML},
		{"Cargo.toml", FormatTOML},
		{"pyproject.toml", FormatTOML},
		{"VERSION", FormatRaw},
		{".version", FormatRaw},
		{"version.txt", FormatRaw},
		{"unknown.xyz", FormatRaw},
		{"/path/to/package.json", FormatJSON},
		{"/path/to/Cargo.toml", FormatTOML},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := FormatForFile(tt.filename)
			if got != tt.want {
				t.Errorf("FormatForFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestFieldForFormat(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"package.json", "version"},
		{"Cargo.toml", "package.version"},
		{"pyproject.toml", "project.version"},
		{"Chart.yaml", "version"},
		{"unknown.json", "version"},
		{"/path/to/package.json", "version"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := FieldForFormat(tt.filename)
			if got != tt.want {
				t.Errorf("FieldForFormat(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
