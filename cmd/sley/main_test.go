package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMain_ShowVersion(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	if err := os.WriteFile(versionPath, []byte("1.2.3\n"), 0600); err != nil {
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
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	// Redirect stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCLI([]string{"sley", "show"})

	_ = w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	got := strings.TrimSpace(buf.String())
	if got != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %q", got)
	}
}

func TestRunMain_SetupCLIError(t *testing.T) {
	tmp := t.TempDir()

	noWrite := filepath.Join(tmp, "bad")
	if err := os.Mkdir(noWrite, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(noWrite, 0755); err != nil {
			t.Fatalf("failed to reset permissions: %v", err)
		}
	})

	versionPath := filepath.Join("bad", ".version")
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
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	err = runCLI([]string{"sley", "bump", "patch"})
	if err == nil {
		t.Fatal("expected error from setupCLI, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunMain_LoadConfigError(t *testing.T) {
	tmp := t.TempDir()

	// Create an unreadable .sley.yaml file
	configPath := filepath.Join(tmp, ".sley.yaml")
	if err := os.WriteFile(configPath, []byte("path: ./broken.version"), 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(configPath, 0644) // restore permissions so temp dir can be cleaned
	})

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	err = runCLI([]string{"sley", "bump", "patch"})
	if err == nil {
		t.Fatal("expected error from LoadConfig, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("unexpected error: %v", err)
	}
}
