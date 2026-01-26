package discovery

import (
	"testing"

	"github.com/indaco/sley/internal/parser"
)

func TestDetectionMode_String(t *testing.T) {
	tests := []struct {
		mode DetectionMode
		want string
	}{
		{NoModules, "NoModules"},
		{SingleModule, "SingleModule"},
		{MultiModule, "MultiModule"},
		{DetectionMode(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("DetectionMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}

func TestResult_HasModules(t *testing.T) {
	tests := []struct {
		name    string
		modules []Module
		want    bool
	}{
		{
			name:    "no modules",
			modules: nil,
			want:    false,
		},
		{
			name:    "empty modules",
			modules: []Module{},
			want:    false,
		},
		{
			name:    "has modules",
			modules: []Module{{Name: "test"}},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{Modules: tt.modules}
			if got := r.HasModules(); got != tt.want {
				t.Errorf("HasModules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_HasManifests(t *testing.T) {
	tests := []struct {
		name      string
		manifests []ManifestSource
		want      bool
	}{
		{
			name:      "no manifests",
			manifests: nil,
			want:      false,
		},
		{
			name:      "empty manifests",
			manifests: []ManifestSource{},
			want:      false,
		},
		{
			name:      "has manifests",
			manifests: []ManifestSource{{Filename: "package.json"}},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{Manifests: tt.manifests}
			if got := r.HasManifests(); got != tt.want {
				t.Errorf("HasManifests() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_HasMismatches(t *testing.T) {
	tests := []struct {
		name       string
		mismatches []Mismatch
		want       bool
	}{
		{
			name:       "no mismatches",
			mismatches: nil,
			want:       false,
		},
		{
			name:       "has mismatches",
			mismatches: []Mismatch{{Source: "test.json"}},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{Mismatches: tt.mismatches}
			if got := r.HasMismatches(); got != tt.want {
				t.Errorf("HasMismatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_IsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		modules   []Module
		manifests []ManifestSource
		want      bool
	}{
		{
			name:      "both empty",
			modules:   nil,
			manifests: nil,
			want:      true,
		},
		{
			name:      "has modules",
			modules:   []Module{{Name: "test"}},
			manifests: nil,
			want:      false,
		},
		{
			name:      "has manifests",
			modules:   nil,
			manifests: []ManifestSource{{Filename: "test"}},
			want:      false,
		},
		{
			name:      "has both",
			modules:   []Module{{Name: "test"}},
			manifests: []ManifestSource{{Filename: "test"}},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{Modules: tt.modules, Manifests: tt.manifests}
			if got := r.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_PrimaryVersion(t *testing.T) {
	tests := []struct {
		name      string
		modules   []Module
		manifests []ManifestSource
		want      string
	}{
		{
			name:      "empty result",
			modules:   nil,
			manifests: nil,
			want:      "",
		},
		{
			name: "root version",
			modules: []Module{
				{Name: "root", RelPath: ".version", Version: "1.0.0"},
				{Name: "other", RelPath: "other/.version", Version: "2.0.0"},
			},
			want: "1.0.0",
		},
		{
			name: "first module",
			modules: []Module{
				{Name: "first", RelPath: "first/.version", Version: "1.0.0"},
				{Name: "second", RelPath: "second/.version", Version: "2.0.0"},
			},
			want: "1.0.0",
		},
		{
			name:      "first manifest",
			modules:   nil,
			manifests: []ManifestSource{{Filename: "package.json", Version: "3.0.0"}},
			want:      "3.0.0",
		},
		{
			name: "module takes precedence over manifest",
			modules: []Module{
				{Name: "test", RelPath: "test/.version", Version: "1.0.0"},
			},
			manifests: []ManifestSource{{Filename: "package.json", Version: "2.0.0"}},
			want:      "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{Modules: tt.modules, Manifests: tt.manifests}
			if got := r.PrimaryVersion(); got != tt.want {
				t.Errorf("PrimaryVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSyncCandidate_ToFileConfig(t *testing.T) {
	candidate := SyncCandidate{
		Path:    "package.json",
		Format:  parser.FormatJSON,
		Field:   "version",
		Pattern: "",
	}

	cfg := candidate.ToFileConfig()

	if cfg.Path != candidate.Path {
		t.Errorf("Path = %q, want %q", cfg.Path, candidate.Path)
	}
	if cfg.Format != candidate.Format {
		t.Errorf("Format = %v, want %v", cfg.Format, candidate.Format)
	}
	if cfg.Field != candidate.Field {
		t.Errorf("Field = %q, want %q", cfg.Field, candidate.Field)
	}
}

func TestDefaultKnownManifests(t *testing.T) {
	manifests := DefaultKnownManifests()

	// Should have common manifest types
	if len(manifests) == 0 {
		t.Error("expected non-empty manifest list")
	}

	// Check for expected files
	expectedFiles := []string{"package.json", "Cargo.toml", "pyproject.toml", "Chart.yaml"}
	for _, expected := range expectedFiles {
		found := false
		for _, m := range manifests {
			if m.Filename == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find %q in known manifests", expected)
		}
	}

	// Verify priorities are set
	for _, m := range manifests {
		if m.Priority == 0 && m.Filename != "" {
			t.Errorf("manifest %q should have a priority set", m.Filename)
		}
	}
}
