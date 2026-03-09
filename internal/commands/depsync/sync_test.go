package depsync

import (
	"testing"

	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/semver"
)

func TestDeriveDependencyName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "version file returns parent dir",
			path:     "subdir/.version",
			expected: "subdir",
		},
		{
			name:     "nested version file returns parent dir",
			path:     "a/b/c/.version",
			expected: "c",
		},
		{
			name:     "package.json returns filename",
			path:     "frontend/package.json",
			expected: "package.json",
		},
		{
			name:     "plain filename",
			path:     "go.mod",
			expected: "go.mod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveDependencyName(tt.path)
			if got != tt.expected {
				t.Errorf("deriveDependencyName(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestSyncDependencies_NilChecker(t *testing.T) {
	registry := plugins.NewPluginRegistry()
	ver := semver.SemVersion{Major: 1, Minor: 0, Patch: 0}

	err := SyncDependencies(registry, ver)
	if err != nil {
		t.Errorf("expected nil error when no dependency checker, got %v", err)
	}
}
