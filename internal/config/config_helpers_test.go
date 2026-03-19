package config

import (
	"path/filepath"
	"testing"
)

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */

// runInTempDir runs a function in a temporary directory using t.Chdir,
// which automatically saves and restores the working directory and is
// safe for use with t.Parallel().
func runInTempDir(t *testing.T, tmpPath string, fn func()) {
	t.Helper()
	targetDir := filepath.Dir(tmpPath)
	t.Chdir(targetDir)
	fn()
}

func checkError(t *testing.T, err error, wantErr bool) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Fatalf("expected err=%v, got err=%v", wantErr, err)
	}
}

func checkConfigNil(t *testing.T, cfg *Config, wantNil bool) {
	t.Helper()
	if wantNil && cfg != nil {
		t.Errorf("expected nil config, got %+v", cfg)
	}
	if !wantNil && cfg == nil {
		t.Fatal("expected non-nil config, got nil")
	}
}

func checkConfigPath(t *testing.T, cfg *Config, wantNil bool, wantPath string) {
	t.Helper()
	if !wantNil && cfg.Path != wantPath {
		t.Errorf("expected path %q, got %q", wantPath, cfg.Path)
	}
}

func requireNonNilWorkspace(t *testing.T, cfg *Config) {
	t.Helper()
	if cfg.Workspace == nil {
		t.Fatal("expected Workspace to be non-nil")
	}
}

func requireNonNilDiscovery(t *testing.T, cfg *Config) {
	t.Helper()
	requireNonNilWorkspace(t, cfg)
	if cfg.Workspace.Discovery == nil {
		t.Fatal("expected Discovery to be non-nil")
	}
}

func assertBoolPtr(t *testing.T, name string, ptr *bool, expected bool) {
	t.Helper()
	if ptr == nil {
		t.Errorf("expected %s to be non-nil", name)
		return
	}
	if *ptr != expected {
		t.Errorf("expected %s to be %v, got %v", name, expected, *ptr)
	}
}

func assertIntPtr(t *testing.T, name string, ptr *int, expected int) {
	t.Helper()
	if ptr == nil {
		t.Errorf("expected %s to be non-nil", name)
		return
	}
	if *ptr != expected {
		t.Errorf("expected %s to be %d, got %d", name, expected, *ptr)
	}
}

func assertDiscoveryEnabled(t *testing.T, disc *DiscoveryConfig, expected bool) {
	t.Helper()
	assertBoolPtr(t, "Enabled", disc.Enabled, expected)
}

func assertDiscoveryRecursive(t *testing.T, disc *DiscoveryConfig, expected bool) {
	t.Helper()
	assertBoolPtr(t, "Recursive", disc.Recursive, expected)
}

func assertDiscoveryMaxDepth(t *testing.T, disc *DiscoveryConfig, expected int) {
	t.Helper()
	assertIntPtr(t, "ModuleMaxDepth", disc.ModuleMaxDepth, expected)
}
