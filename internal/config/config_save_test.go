package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

/* ------------------------------------------------------------------------- */
/* MOCK IMPLEMENTATIONS FOR TESTING                                          */
/* ------------------------------------------------------------------------- */

// mockMarshaler implements core.Marshaler for testing.
type mockMarshaler struct {
	marshalErr    error
	marshalOutput []byte
}

func (m *mockMarshaler) Marshal(v any) ([]byte, error) {
	if m.marshalErr != nil {
		return nil, m.marshalErr
	}
	if m.marshalOutput != nil {
		return m.marshalOutput, nil
	}
	// Default: return valid YAML
	return []byte("path: test\n"), nil
}

// mockFileOpener implements FileOpener for testing.
type mockFileOpener struct {
	openFileErr  error
	openFileFunc func(name string, flag int, perm os.FileMode) (*os.File, error)
}

func (m *mockFileOpener) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	if m.openFileErr != nil {
		return nil, m.openFileErr
	}
	if m.openFileFunc != nil {
		return m.openFileFunc(name, flag, perm)
	}
	return os.OpenFile(name, flag, perm)
}

// mockFileWriter implements FileWriter for testing.
type mockFileWriter struct {
	writeFileErr error
}

func (m *mockFileWriter) WriteFile(file *os.File, data []byte) (int, error) {
	if m.writeFileErr != nil {
		return 0, m.writeFileErr
	}
	return file.Write(data)
}

/* ------------------------------------------------------------------------- */
/* SAVE CONFIG                                                               */
/* ------------------------------------------------------------------------- */

func TestConfigSaver_Save(t *testing.T) {
	t.Run("basic save scenarios", func(t *testing.T) {
		tests := []struct {
			name          string
			cfg           *Config
			wantErr       bool
			mockMarshaler *mockMarshaler
			mockOpener    *mockFileOpener
			mockWriter    *mockFileWriter
		}{
			{
				name:    "save minimal config",
				cfg:     &Config{Path: "my.version"},
				wantErr: false,
			},
			{
				name: "save config with plugins",
				cfg: &Config{
					Path: "custom.version",
					Extensions: []ExtensionConfig{
						{Name: "example", Path: "/plugin/path", Enabled: true},
					},
				},
				wantErr: false,
			},
			{
				name:    "marshal failure",
				cfg:     &Config{Path: "fail.version"},
				wantErr: true,
				mockMarshaler: &mockMarshaler{
					marshalErr: fmt.Errorf("mock marshal failure"),
				},
			},
			{
				name:    "open file failure",
				cfg:     &Config{Path: "fail-open.version"},
				wantErr: true,
				mockOpener: &mockFileOpener{
					openFileErr: fmt.Errorf("permission denied"),
				},
			},
			{
				name:    "write file failure",
				cfg:     &Config{Path: "fail-write.version"},
				wantErr: true,
				mockWriter: &mockFileWriter{
					writeFileErr: fmt.Errorf("simulated write failure"),
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmp := t.TempDir()
				configFile := filepath.Join(tmp, ".sley.yaml")

				// Determine which dependencies to use (nil means use default)
				var marshaler interface{ Marshal(any) ([]byte, error) }
				var opener FileOpener
				var writer FileWriter

				if tt.mockMarshaler != nil {
					marshaler = tt.mockMarshaler
				}
				if tt.mockOpener != nil {
					opener = tt.mockOpener
				}
				if tt.mockWriter != nil {
					writer = tt.mockWriter
				}

				// Create the ConfigSaver with mock dependencies
				saver := NewConfigSaver(marshaler, opener, writer)

				err := saver.SaveTo(tt.cfg, configFile)
				if (err != nil) != tt.wantErr {
					t.Fatalf("ConfigSaver.SaveTo() error = %v, wantErr = %v", err, tt.wantErr)
				}

				if !tt.wantErr {
					if _, err := os.Stat(configFile); err != nil {
						t.Errorf("config file was not created: %v", err)
					}
				}
			})
		}
	})

	t.Run("write fails due to directory permission", func(t *testing.T) {
		tmp := t.TempDir()
		badDir := filepath.Join(tmp, "readonly")
		if err := os.Mkdir(badDir, 0500); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Chmod(badDir, 0755); err != nil {
				t.Logf("cleanup warning: failed to chmod %q: %v", badDir, err)
			}
		}()

		saver := NewConfigSaver(nil, nil, nil)
		configFile := filepath.Join(badDir, ".sley.yaml")
		err := saver.SaveTo(&Config{Path: "blocked.version"}, configFile)
		if err == nil {
			t.Error("expected error due to write permission, got nil")
		}
	})
}

func TestConfigSaver_WriteError(t *testing.T) {
	tmp := t.TempDir()
	configFile := filepath.Join(tmp, ".sley.yaml")

	// Create saver with mock writer that returns an error
	mockWriter := &mockFileWriter{
		writeFileErr: fmt.Errorf("simulated write failure"),
	}
	saver := NewConfigSaver(nil, nil, mockWriter)

	cfg := &Config{Path: "whatever"}
	err := saver.SaveTo(cfg, configFile)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	want := fmt.Sprintf("failed to write config to %q: simulated write failure", configFile)
	if err.Error() != want {
		t.Errorf("unexpected error. got: %q, want: %q", err.Error(), want)
	}
}

// TestSaveConfigFn_BackwardCompatibility ensures the backward-compatible SaveConfigFn still works.
func TestSaveConfigFn_BackwardCompatibility(t *testing.T) {
	tmp := t.TempDir()
	runInTempDir(t, filepath.Join(tmp, "dummy"), func() {
		cfg := &Config{Path: "test.version"}
		err := SaveConfigFn(cfg)
		if err != nil {
			t.Fatalf("SaveConfigFn() error = %v", err)
		}

		if _, err := os.Stat(".sley.yaml"); err != nil {
			t.Errorf(".sley.yaml was not created: %v", err)
		}
	})
}

func TestNewConfigSaver_Defaults(t *testing.T) {
	// Test that NewConfigSaver uses defaults when nil is passed
	saver := NewConfigSaver(nil, nil, nil)
	if saver == nil {
		t.Fatal("NewConfigSaver returned nil")
	}
	if saver.marshaler == nil {
		t.Error("marshaler should not be nil")
	}
	if saver.fileOpener == nil {
		t.Error("fileOpener should not be nil")
	}
	if saver.fileWriter == nil {
		t.Error("fileWriter should not be nil")
	}
}
