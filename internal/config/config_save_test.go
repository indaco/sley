package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
)

/* ------------------------------------------------------------------------- */
/* SAVE CONFIG                                                               */
/* ------------------------------------------------------------------------- */

func TestSaveConfigFn(t *testing.T) {
	t.Run("basic save scenarios", func(t *testing.T) {
		defer func() {
			marshalFn = yaml.Marshal
			openFileFn = os.OpenFile
		}()

		tests := []struct {
			name               string
			cfg                *Config
			wantErr            bool
			overwriteMarshalFn bool
			mockMarshalErr     error
			overwriteOpenFile  bool
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
				name:               "marshal failure",
				cfg:                &Config{Path: "fail.version"},
				wantErr:            true,
				overwriteMarshalFn: true,
				mockMarshalErr:     fmt.Errorf("mock marshal failure"),
			},
			{
				name:              "write fails due to file permission",
				cfg:               &Config{Path: "fail-write.version"},
				wantErr:           true,
				overwriteOpenFile: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmp := t.TempDir()
				runInTempDir(t, filepath.Join(tmp, "dummy"), func() {
					if tt.overwriteMarshalFn {
						marshalFn = func(any) ([]byte, error) {
							return nil, tt.mockMarshalErr
						}
					}

					if tt.overwriteOpenFile {
						openFileFn = func(name string, flag int, perm os.FileMode) (*os.File, error) {
							// Simulate permission denied by opening read-only
							path := filepath.Join(t.TempDir(), "readonly.yaml")
							f, err := os.Create(path)
							if err != nil {
								t.Fatal(err)
							}
							f.Close()
							_ = os.Chmod(path, 0400)
							return os.OpenFile(path, os.O_WRONLY, 0400)
						}
					}

					err := SaveConfigFn(tt.cfg)
					if (err != nil) != tt.wantErr {
						t.Fatalf("SaveConfigFn() error = %v, wantErr = %v", err, tt.wantErr)
					}

					if !tt.wantErr {
						if _, err := os.Stat(".sley.yaml"); err != nil {
							t.Errorf(".sley.yaml was not created: %v", err)
						}
					}
				})
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

		runInTempDir(t, filepath.Join(badDir, "dummy"), func() {
			err := SaveConfigFn(&Config{Path: "blocked.version"})
			if err == nil {
				t.Error("expected error due to write permission, got nil")
			}
		})
	})
}

func TestSaveConfigFn_WriteFileFn_Error(t *testing.T) {
	origWriteFn := writeFileFn
	defer func() {
		writeFileFn = origWriteFn
	}()

	tmp := t.TempDir()
	runInTempDir(t, filepath.Join(tmp, "dummy"), func() {
		writeFileFn = func(f *os.File, data []byte) (int, error) {
			fmt.Println(">>> writeFileFn invoked")
			return 0, fmt.Errorf("simulated write failure")
		}

		cfg := &Config{Path: "whatever"}
		err := SaveConfigFn(cfg)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		want := "failed to write config data: simulated write failure"
		if err.Error() != want {
			t.Errorf("unexpected error. got: %q, want: %q", err.Error(), want)
		}
	})
}
