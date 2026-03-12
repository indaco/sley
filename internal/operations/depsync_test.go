package operations

import (
	"errors"
	"testing"

	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/semver"
)

// mockDependencyChecker implements dependencycheck.DependencyChecker for testing.
type mockDependencyChecker struct {
	enabled  bool
	config   *dependencycheck.Config
	syncErr  error
	syncCall string // captures the version string passed to SyncVersions
}

func (m *mockDependencyChecker) Name() string                       { return "mock-dep-check" }
func (m *mockDependencyChecker) Description() string                { return "mock" }
func (m *mockDependencyChecker) Version() string                    { return "v0.0.1" }
func (m *mockDependencyChecker) IsEnabled() bool                    { return m.enabled }
func (m *mockDependencyChecker) GetConfig() *dependencycheck.Config { return m.config }
func (m *mockDependencyChecker) CheckConsistency(currentVersion string) ([]dependencycheck.Inconsistency, error) {
	return nil, nil
}
func (m *mockDependencyChecker) SyncVersions(newVersion string) error {
	m.syncCall = newVersion
	return m.syncErr
}

func newTestRegistry(dc dependencycheck.DependencyChecker) *plugins.PluginRegistry {
	r := plugins.NewPluginRegistry()
	if dc != nil {
		_ = r.RegisterDependencyChecker(dc)
	}
	return r
}

func TestSyncDependencies_NilChecker(t *testing.T) {
	t.Parallel()
	registry := plugins.NewPluginRegistry() // no dependency checker registered
	ver := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	err := SyncDependencies(registry, ver)
	if err != nil {
		t.Fatalf("expected nil error when no checker registered, got %v", err)
	}
}

func TestSyncDependencies_DisabledChecker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		enabled bool
		auto    bool
	}{
		{"disabled plugin", false, true},
		{"enabled but no auto-sync", true, false},
		{"both disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dc := &mockDependencyChecker{
				enabled: tt.enabled,
				config: &dependencycheck.Config{
					Enabled:  tt.enabled,
					AutoSync: tt.auto,
					Files:    []dependencycheck.FileConfig{{Path: "package.json"}},
				},
			}
			registry := newTestRegistry(dc)
			ver := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

			err := SyncDependencies(registry, ver)
			if err != nil {
				t.Fatalf("expected nil error for disabled checker, got %v", err)
			}
			if dc.syncCall != "" {
				t.Error("SyncVersions should not have been called")
			}
		})
	}
}

func TestSyncDependencies_EmptyFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		files []dependencycheck.FileConfig
	}{
		{"nil files", nil},
		{"empty files", []dependencycheck.FileConfig{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dc := &mockDependencyChecker{
				enabled: true,
				config: &dependencycheck.Config{
					Enabled:  true,
					AutoSync: true,
					Files:    tt.files,
				},
			}
			registry := newTestRegistry(dc)
			ver := semver.SemVersion{Major: 2, Minor: 0, Patch: 0}

			err := SyncDependencies(registry, ver)
			if err != nil {
				t.Fatalf("expected nil error for empty files, got %v", err)
			}
			if dc.syncCall != "" {
				t.Error("SyncVersions should not have been called for empty files")
			}
		})
	}
}

func TestSyncDependencies_SyncError(t *testing.T) {
	t.Parallel()

	syncErr := errors.New("simulated sync failure")
	dc := &mockDependencyChecker{
		enabled: true,
		syncErr: syncErr,
		config: &dependencycheck.Config{
			Enabled:  true,
			AutoSync: true,
			Files:    []dependencycheck.FileConfig{{Path: "package.json", Format: "json"}},
		},
	}
	registry := newTestRegistry(dc)
	ver := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	err := SyncDependencies(registry, ver)
	if err == nil {
		t.Fatal("expected error when SyncVersions fails, got nil")
	}
	if !errors.Is(err, syncErr) {
		t.Errorf("expected wrapped sync error, got %v", err)
	}
}

func TestSyncDependencies_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		files       []dependencycheck.FileConfig
		bumpedPaths []string
		version     semver.SemVersion
		wantSync    string
	}{
		{
			name:     "single file",
			files:    []dependencycheck.FileConfig{{Path: "package.json", Format: "json"}},
			version:  semver.SemVersion{Major: 1, Minor: 2, Patch: 3},
			wantSync: "1.2.3",
		},
		{
			name: "multiple files",
			files: []dependencycheck.FileConfig{
				{Path: "package.json", Format: "json"},
				{Path: "pyproject.toml", Format: "toml"},
			},
			version:  semver.SemVersion{Major: 3, Minor: 0, Patch: 0},
			wantSync: "3.0.0",
		},
		{
			name:        "with bumped paths filtered",
			files:       []dependencycheck.FileConfig{{Path: "libs/a/.version"}, {Path: "libs/b/.version"}},
			bumpedPaths: []string{"libs/a/.version"},
			version:     semver.SemVersion{Major: 2, Minor: 1, Patch: 0},
			wantSync:    "2.1.0",
		},
		{
			name:        "all paths bumped",
			files:       []dependencycheck.FileConfig{{Path: "libs/a/.version"}},
			bumpedPaths: []string{"libs/a/.version"},
			version:     semver.SemVersion{Major: 1, Minor: 0, Patch: 0},
			wantSync:    "1.0.0",
		},
		{
			name:     "version with pre-release",
			files:    []dependencycheck.FileConfig{{Path: "package.json"}},
			version:  semver.SemVersion{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.1"},
			wantSync: "1.0.0-alpha.1",
		},
		{
			name:     "version with build metadata",
			files:    []dependencycheck.FileConfig{{Path: "package.json"}},
			version:  semver.SemVersion{Major: 1, Minor: 0, Patch: 0, Build: "build.42"},
			wantSync: "1.0.0+build.42",
		},
		{
			name:     "zero version",
			files:    []dependencycheck.FileConfig{{Path: "package.json"}},
			version:  semver.SemVersion{Major: 0, Minor: 0, Patch: 0},
			wantSync: "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dc := &mockDependencyChecker{
				enabled: true,
				config: &dependencycheck.Config{
					Enabled:  true,
					AutoSync: true,
					Files:    tt.files,
				},
			}
			registry := newTestRegistry(dc)

			err := SyncDependencies(registry, tt.version, tt.bumpedPaths...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if dc.syncCall != tt.wantSync {
				t.Errorf("SyncVersions called with %q, want %q", dc.syncCall, tt.wantSync)
			}
		})
	}
}

func TestDeriveDependencyName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{"version file uses parent dir", "libs/mylib/.version", "mylib"},
		{"nested version file", "packages/core/sub/.version", "sub"},
		{"root version file", ".version", "."},
		{"package.json uses filename", "libs/mylib/package.json", "package.json"},
		{"pyproject.toml uses filename", "pyproject.toml", "pyproject.toml"},
		{"Cargo.toml uses filename", "crates/foo/Cargo.toml", "Cargo.toml"},
		{"absolute path version file", "/home/user/project/.version", "project"},
		{"absolute path other file", "/home/user/project/package.json", "package.json"},
		{"single component", "package.json", "package.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DeriveDependencyName(tt.path)
			if got != tt.want {
				t.Errorf("DeriveDependencyName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
