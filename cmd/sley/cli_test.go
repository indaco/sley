package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunCLI_InitializeVersionFileError tests the runCLI function from main.go
// which handles version file initialization errors.
func TestRunCLI_InitializeVersionFileError(t *testing.T) {
	tmp := t.TempDir()

	noWrite := filepath.Join(tmp, "nonwritable")
	if err := os.Mkdir(noWrite, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(noWrite, 0755)
	})

	versionPath := filepath.Join("nonwritable", ".version")
	yamlPath := filepath.Join(tmp, ".sley.yaml")
	if err := os.WriteFile(yamlPath, []byte("path: "+versionPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	err = runCLI([]string{"sley", "bump", "patch"})
	if err == nil {
		t.Fatal("expected error from InitializeVersionFile, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("unexpected error: %v", err)
	}
}
