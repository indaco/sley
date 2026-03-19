package dependencycheck

import (
	"errors"
	"testing"
)

func TestDependencyCheckerPlugin_Name(t *testing.T) {
	t.Parallel()

	dc := NewDependencyChecker(nil)
	if got := dc.Name(); got != "dependency-check" {
		t.Errorf("Name() = %q, want %q", got, "dependency-check")
	}
}

func TestDependencyCheckerPlugin_Description(t *testing.T) {
	t.Parallel()

	dc := NewDependencyChecker(nil)
	if got := dc.Description(); got == "" {
		t.Error("Description() should not be empty")
	}
}

func TestDependencyCheckerPlugin_Version(t *testing.T) {
	t.Parallel()

	dc := NewDependencyChecker(nil)
	if got := dc.Version(); got != "v0.1.0" {
		t.Errorf("Version() = %q, want %q", got, "v0.1.0")
	}
}

func TestDependencyCheckerPlugin_IsEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name:   "nil config defaults to disabled",
			config: nil,
			want:   false,
		},
		{
			name:   "explicitly disabled",
			config: &Config{Enabled: false},
			want:   false,
		},
		{
			name:   "explicitly enabled",
			config: &Config{Enabled: true},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dc := NewDependencyChecker(tt.config)
			if got := dc.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDependencyCheckerPlugin_GetConfig(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Enabled:  true,
		AutoSync: true,
		Files: []FileConfig{
			{Path: "package.json", Field: "version", Format: "json"},
		},
	}
	dc := NewDependencyChecker(cfg)

	got := dc.GetConfig()
	if got != cfg {
		t.Error("GetConfig() should return the same config instance")
	}
	if got.Enabled != true {
		t.Error("Config.Enabled should be true")
	}
	if got.AutoSync != true {
		t.Error("Config.AutoSync should be true")
	}
	if len(got.Files) != 1 {
		t.Errorf("Config.Files length = %d, want 1", len(got.Files))
	}
}

func TestDependencyCheckerPlugin_CheckConsistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		config        *Config
		currentVer    string
		mockReaders   map[string]func(string, string) (string, error)
		wantInconsLen int
		wantErr       bool
	}{
		{
			name: "all files consistent",
			config: &Config{
				Enabled: true,
				Files: []FileConfig{
					{Path: "package.json", Field: "version", Format: "json"},
					{Path: "Chart.yaml", Field: "version", Format: "yaml"},
				},
			},
			currentVer: "1.2.3",
			mockReaders: map[string]func(string, string) (string, error){
				"json": func(path, field string) (string, error) { return "1.2.3", nil },
				"yaml": func(path, field string) (string, error) { return "1.2.3", nil },
			},
			wantInconsLen: 0,
			wantErr:       false,
		},
		{
			name: "version mismatch in one file",
			config: &Config{
				Enabled: true,
				Files: []FileConfig{
					{Path: "package.json", Field: "version", Format: "json"},
					{Path: "Chart.yaml", Field: "version", Format: "yaml"},
				},
			},
			currentVer: "1.2.3",
			mockReaders: map[string]func(string, string) (string, error){
				"json": func(path, field string) (string, error) { return "1.2.2", nil },
				"yaml": func(path, field string) (string, error) { return "1.2.3", nil },
			},
			wantInconsLen: 1,
			wantErr:       false,
		},
		{
			name: "handles v prefix normalization",
			config: &Config{
				Enabled: true,
				Files: []FileConfig{
					{Path: "version.txt", Field: "", Format: "json"},
				},
			},
			currentVer: "v1.2.3",
			mockReaders: map[string]func(string, string) (string, error){
				"json": func(path, field string) (string, error) { return "1.2.3", nil },
			},
			wantInconsLen: 0,
			wantErr:       false,
		},
		{
			name: "read error",
			config: &Config{
				Enabled: true,
				Files: []FileConfig{
					{Path: "missing.json", Field: "version", Format: "json"},
				},
			},
			currentVer: "1.2.3",
			mockReaders: map[string]func(string, string) (string, error){
				"json": func(path, field string) (string, error) {
					return "", errors.New("file not found")
				},
			},
			wantInconsLen: 0,
			wantErr:       true,
		},
		{
			name: "disabled plugin returns nil",
			config: &Config{
				Enabled: false,
				Files: []FileConfig{
					{Path: "package.json", Field: "version", Format: "json"},
				},
			},
			currentVer:    "1.2.3",
			mockReaders:   nil,
			wantInconsLen: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dc := NewDependencyChecker(tt.config)

			// Inject mocks per-instance
			if tt.mockReaders != nil {
				if mockJSON, ok := tt.mockReaders["json"]; ok {
					dc.readJSONVersionFn = mockJSON
				}
				if mockYAML, ok := tt.mockReaders["yaml"]; ok {
					dc.readYAMLVersionFn = mockYAML
				}
				if mockTOML, ok := tt.mockReaders["toml"]; ok {
					dc.readTOMLVersionFn = mockTOML
				}
			}

			inconsistencies, err := dc.CheckConsistency(tt.currentVer)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckConsistency() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(inconsistencies) != tt.wantInconsLen {
				t.Errorf("CheckConsistency() inconsistencies length = %d, want %d", len(inconsistencies), tt.wantInconsLen)
			}
		})
	}
}

func TestDependencyCheckerPlugin_SyncVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *Config
		newVersion  string
		mockWriters map[string]func(string, string, string) error
		wantErr     bool
	}{
		{
			name: "sync all files successfully",
			config: &Config{
				Enabled:  true,
				AutoSync: true,
				Files: []FileConfig{
					{Path: "package.json", Field: "version", Format: "json"},
					{Path: "Chart.yaml", Field: "version", Format: "yaml"},
				},
			},
			newVersion: "1.2.4",
			mockWriters: map[string]func(string, string, string) error{
				"json": func(path, field, version string) error { return nil },
				"yaml": func(path, field, version string) error { return nil },
			},
			wantErr: false,
		},
		{
			name: "write error in one file",
			config: &Config{
				Enabled:  true,
				AutoSync: true,
				Files: []FileConfig{
					{Path: "package.json", Field: "version", Format: "json"},
				},
			},
			newVersion: "1.2.4",
			mockWriters: map[string]func(string, string, string) error{
				"json": func(path, field, version string) error {
					return errors.New("write failed")
				},
			},
			wantErr: true,
		},
		{
			name: "disabled plugin does nothing",
			config: &Config{
				Enabled: false,
				Files: []FileConfig{
					{Path: "package.json", Field: "version", Format: "json"},
				},
			},
			newVersion:  "1.2.4",
			mockWriters: nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dc := NewDependencyChecker(tt.config)

			// Inject mocks per-instance
			if tt.mockWriters != nil {
				if mockJSON, ok := tt.mockWriters["json"]; ok {
					dc.writeJSONVersionFn = mockJSON
				}
				if mockYAML, ok := tt.mockWriters["yaml"]; ok {
					dc.writeYAMLVersionFn = mockYAML
				}
				if mockTOML, ok := tt.mockWriters["toml"]; ok {
					dc.writeTOMLVersionFn = mockTOML
				}
			}

			err := dc.SyncVersions(tt.newVersion)

			if (err != nil) != tt.wantErr {
				t.Errorf("SyncVersions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInconsistency_String(t *testing.T) {
	t.Parallel()

	inc := Inconsistency{
		Path:     "package.json",
		Expected: "1.2.3",
		Found:    "1.2.2",
		Format:   "json",
	}

	got := inc.String()
	want := "package.json: expected 1.2.3, found 1.2.2 (format: json)"
	if got != want {
		t.Errorf("Inconsistency.String() = %q, want %q", got, want)
	}
}

func TestNormalizeVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"1.2.3", "1.2.3"},
		{"v1.2.3", "1.2.3"},
		{"v2.0.0-alpha", "2.0.0-alpha"},
		{"3.0.0+build", "3.0.0+build"},
		{"v3.0.0+build", "3.0.0+build"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got := normalizeVersion(tt.input)
			if got != tt.want {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("DefaultConfig() Enabled should be false")
	}
	if cfg.AutoSync {
		t.Error("DefaultConfig() AutoSync should be false")
	}
	if len(cfg.Files) != 0 {
		t.Errorf("DefaultConfig() Files length = %d, want 0", len(cfg.Files))
	}
}

func TestReadVersionFromFile_AllFormats(t *testing.T) {
	t.Parallel()

	dc := NewDependencyChecker(&Config{Enabled: true})

	// Override with mocks per-instance
	dc.readJSONVersionFn = func(path, field string) (string, error) { return "1.0.0", nil }
	dc.readYAMLVersionFn = func(path, field string) (string, error) { return "1.0.0", nil }
	dc.readTOMLVersionFn = func(path, field string) (string, error) { return "1.0.0", nil }
	dc.readRawVersionFn = func(path string) (string, error) { return "1.0.0", nil }
	dc.readRegexVersionFn = func(path, pattern string) (string, error) { return "1.0.0", nil }

	tests := []struct {
		name    string
		file    FileConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "json format",
			file:    FileConfig{Path: "package.json", Field: "version", Format: "json"},
			wantErr: false,
		},
		{
			name:    "yaml format",
			file:    FileConfig{Path: "Chart.yaml", Field: "version", Format: "yaml"},
			wantErr: false,
		},
		{
			name:    "toml format",
			file:    FileConfig{Path: "pyproject.toml", Field: "tool.poetry.version", Format: "toml"},
			wantErr: false,
		},
		{
			name:    "raw format",
			file:    FileConfig{Path: "VERSION", Format: "raw"},
			wantErr: false,
		},
		{
			name:    "regex format with pattern",
			file:    FileConfig{Path: "version.go", Pattern: `const Version = "(.*?)"`, Format: "regex"},
			wantErr: false,
		},
		{
			name:    "regex format without pattern",
			file:    FileConfig{Path: "version.go", Format: "regex"},
			wantErr: true,
			errMsg:  "regex format requires a pattern",
		},
		{
			name:    "unsupported format",
			file:    FileConfig{Path: "file.txt", Format: "unknown"},
			wantErr: true,
			errMsg:  "unsupported format: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := dc.readVersionFromFile(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("readVersionFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("readVersionFromFile() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestWriteVersionToFile_AllFormats(t *testing.T) {
	t.Parallel()

	dc := NewDependencyChecker(&Config{Enabled: true})

	// Override with mocks per-instance
	dc.writeJSONVersionFn = func(path, field, version string) error { return nil }
	dc.writeYAMLVersionFn = func(path, field, version string) error { return nil }
	dc.writeTOMLVersionFn = func(path, field, version string) error { return nil }
	dc.writeRawVersionFn = func(path, version string) error { return nil }
	dc.writeRegexVersionFn = func(path, pattern, version string) error { return nil }

	tests := []struct {
		name    string
		file    FileConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "json format",
			file:    FileConfig{Path: "package.json", Field: "version", Format: "json"},
			wantErr: false,
		},
		{
			name:    "yaml format",
			file:    FileConfig{Path: "Chart.yaml", Field: "version", Format: "yaml"},
			wantErr: false,
		},
		{
			name:    "toml format",
			file:    FileConfig{Path: "pyproject.toml", Field: "tool.poetry.version", Format: "toml"},
			wantErr: false,
		},
		{
			name:    "raw format",
			file:    FileConfig{Path: "VERSION", Format: "raw"},
			wantErr: false,
		},
		{
			name:    "regex format with pattern",
			file:    FileConfig{Path: "version.go", Pattern: `const Version = "(.*?)"`, Format: "regex"},
			wantErr: false,
		},
		{
			name:    "regex format without pattern",
			file:    FileConfig{Path: "version.go", Format: "regex"},
			wantErr: true,
			errMsg:  "regex format requires a pattern",
		},
		{
			name:    "unsupported format",
			file:    FileConfig{Path: "file.txt", Format: "unknown"},
			wantErr: true,
			errMsg:  "unsupported format: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := dc.writeVersionToFile(tt.file, "1.0.0")
			if (err != nil) != tt.wantErr {
				t.Errorf("writeVersionToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("writeVersionToFile() error = %q, want %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestCheckConsistency_AllFormats(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Enabled: true,
		Files: []FileConfig{
			{Path: "package.json", Field: "version", Format: "json"},
			{Path: "Chart.yaml", Field: "version", Format: "yaml"},
			{Path: "pyproject.toml", Field: "tool.poetry.version", Format: "toml"},
			{Path: "VERSION", Format: "raw"},
			{Path: "version.go", Pattern: `Version = "(.*?)"`, Format: "regex"},
		},
	}

	dc := NewDependencyChecker(cfg)

	// Override with mocks per-instance
	dc.readJSONVersionFn = func(path, field string) (string, error) { return "1.2.3", nil }
	dc.readYAMLVersionFn = func(path, field string) (string, error) { return "1.2.3", nil }
	dc.readTOMLVersionFn = func(path, field string) (string, error) { return "1.2.3", nil }
	dc.readRawVersionFn = func(path string) (string, error) { return "1.2.3", nil }
	dc.readRegexVersionFn = func(path, pattern string) (string, error) { return "1.2.3", nil }

	inconsistencies, err := dc.CheckConsistency("1.2.3")

	if err != nil {
		t.Errorf("CheckConsistency() error = %v", err)
	}
	if len(inconsistencies) != 0 {
		t.Errorf("CheckConsistency() found %d inconsistencies, want 0", len(inconsistencies))
	}
}

func TestSyncVersions_AllFormats(t *testing.T) {
	t.Parallel()

	// Track which writers were called
	called := make(map[string]bool)

	cfg := &Config{
		Enabled:  true,
		AutoSync: true,
		Files: []FileConfig{
			{Path: "package.json", Field: "version", Format: "json"},
			{Path: "Chart.yaml", Field: "version", Format: "yaml"},
			{Path: "pyproject.toml", Field: "tool.poetry.version", Format: "toml"},
			{Path: "VERSION", Format: "raw"},
			{Path: "version.go", Pattern: `Version = "(.*?)"`, Format: "regex"},
		},
	}

	dc := NewDependencyChecker(cfg)

	// Override with mocks per-instance
	dc.writeJSONVersionFn = func(path, field, version string) error { called["json"] = true; return nil }
	dc.writeYAMLVersionFn = func(path, field, version string) error { called["yaml"] = true; return nil }
	dc.writeTOMLVersionFn = func(path, field, version string) error { called["toml"] = true; return nil }
	dc.writeRawVersionFn = func(path, version string) error { called["raw"] = true; return nil }
	dc.writeRegexVersionFn = func(path, pattern, version string) error { called["regex"] = true; return nil }

	err := dc.SyncVersions("2.0.0")

	if err != nil {
		t.Errorf("SyncVersions() error = %v", err)
	}

	// Verify all formats were written
	for _, format := range []string{"json", "yaml", "toml", "raw", "regex"} {
		if !called[format] {
			t.Errorf("SyncVersions() did not write %s format", format)
		}
	}
}

func TestSyncVersions_PluginDisabled(t *testing.T) {
	t.Parallel()

	writeCalled := false

	cfg := &Config{
		Enabled:  false, // Plugin disabled
		AutoSync: true,
		Files: []FileConfig{
			{Path: "package.json", Field: "version", Format: "json"},
		},
	}

	dc := NewDependencyChecker(cfg)
	dc.writeJSONVersionFn = func(path, field, version string) error {
		writeCalled = true
		return nil
	}

	err := dc.SyncVersions("2.0.0")

	if err != nil {
		t.Errorf("SyncVersions() error = %v", err)
	}
	if writeCalled {
		t.Error("SyncVersions() should not write when plugin is disabled")
	}
}

func TestGetConfig_AutoSync(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Enabled:  true,
		AutoSync: false,
	}
	dc := NewDependencyChecker(cfg)

	got := dc.GetConfig()
	if got.AutoSync {
		t.Error("AutoSync should be false")
	}
}

func TestCheckConsistency_NilConfig(t *testing.T) {
	t.Parallel()

	dc := NewDependencyChecker(nil)
	inconsistencies, err := dc.CheckConsistency("1.0.0")

	if err != nil {
		t.Errorf("CheckConsistency() with nil config error = %v", err)
	}
	if len(inconsistencies) != 0 {
		t.Errorf("CheckConsistency() with nil config should return empty inconsistencies")
	}
}

func TestSyncVersions_NilConfig(t *testing.T) {
	t.Parallel()

	dc := NewDependencyChecker(nil)
	err := dc.SyncVersions("1.0.0")

	if err != nil {
		t.Errorf("SyncVersions() with nil config error = %v", err)
	}
}
