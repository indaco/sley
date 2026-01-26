package discover

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/parser"
	"github.com/indaco/sley/internal/testutils"
)

func TestWorkflow_runMismatchWorkflow(t *testing.T) {
	mock := &MockPrompter{}
	result := &discovery.Result{
		Mode: discovery.SingleModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
		Mismatches: []discovery.Mismatch{
			{Source: "package.json", ExpectedVersion: "1.0.0", ActualVersion: "2.0.0"},
			{Source: "Cargo.toml", ExpectedVersion: "1.0.0", ActualVersion: "3.0.0"},
		},
	}
	w := NewWorkflow(mock, result, "/test")

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runMismatchWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if completed {
			t.Error("expected completed to be false")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify warning message is printed
	if !strings.Contains(output, "mismatch") {
		t.Errorf("expected mismatch warning in output, got: %q", output)
	}
}

func TestWorkflow_suggestAdditionalSyncFiles(t *testing.T) {
	mock := &MockPrompter{}
	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
		},
	}
	w := NewWorkflow(mock, result, "/test")

	completed, err := w.suggestAdditionalSyncFiles(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// This function is informational only, returns false, nil
	if completed {
		t.Error("expected completed to be false")
	}
}

func TestWorkflow_printInitSuccess(t *testing.T) {
	w := &Workflow{}

	t.Run("with plugins only", func(t *testing.T) {
		plugins := []string{"commit-parser", "tag-manager"}

		output, err := testutils.CaptureStdout(func() {
			w.printInitSuccess(plugins, nil)
		})
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		if !strings.Contains(output, "2 plugin(s) enabled") {
			t.Errorf("expected plugin count in output, got: %q", output)
		}
		if !strings.Contains(output, "commit-parser") {
			t.Errorf("expected commit-parser in output, got: %q", output)
		}
		if !strings.Contains(output, "tag-manager") {
			t.Errorf("expected tag-manager in output, got: %q", output)
		}
	})

	t.Run("with sync candidates", func(t *testing.T) {
		plugins := []string{"commit-parser", "dependency-check"}
		candidates := []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
			{Path: "Cargo.toml", Format: parser.FormatTOML},
		}

		output, err := testutils.CaptureStdout(func() {
			w.printInitSuccess(plugins, candidates)
		})
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		if !strings.Contains(output, "Configured sync files") {
			t.Errorf("expected sync files section in output, got: %q", output)
		}
		if !strings.Contains(output, "package.json") {
			t.Errorf("expected package.json in output, got: %q", output)
		}
	})

	t.Run("shows next steps", func(t *testing.T) {
		plugins := []string{"commit-parser"}

		output, err := testutils.CaptureStdout(func() {
			w.printInitSuccess(plugins, nil)
		})
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		if !strings.Contains(output, "Next steps") {
			t.Errorf("expected Next steps in output, got: %q", output)
		}
		if !strings.Contains(output, "sley bump") {
			t.Errorf("expected 'sley bump' suggestion in output, got: %q", output)
		}
	})
}

func TestWorkflow_ensureVersionFile(t *testing.T) {
	t.Run("creates version file when missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		w := &Workflow{rootDir: tmpDir}

		err := w.ensureVersionFile(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify .version file was created
		versionPath := filepath.Join(tmpDir, ".version")
		if _, err := os.Stat(versionPath); os.IsNotExist(err) {
			t.Error("expected .version file to be created")
		}
	})

	t.Run("does nothing when file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Chdir(tmpDir)

		// Create existing version file
		versionPath := filepath.Join(tmpDir, ".version")
		if err := os.WriteFile(versionPath, []byte("9.9.9\n"), 0644); err != nil {
			t.Fatal(err)
		}

		w := &Workflow{rootDir: tmpDir}

		err := w.ensureVersionFile(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify version wasn't changed
		data, _ := os.ReadFile(versionPath)
		if strings.TrimSpace(string(data)) != "9.9.9" {
			t.Errorf("expected version to remain 9.9.9, got: %s", string(data))
		}
	})
}

func TestWorkflow_createConfigWithDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	w := &Workflow{rootDir: tmpDir}

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.createConfigWithDefaults(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if !completed {
			t.Error("expected completed to be true")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify .version was created
	versionPath := filepath.Join(tmpDir, ".version")
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("expected .version file to be created")
	}

	// Verify .sley.yaml was created
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected .sley.yaml file to be created")
	}

	// Verify output contains success message
	if !strings.Contains(output, "plugin") {
		t.Errorf("expected plugin info in output, got: %q", output)
	}
}

func TestWorkflow_createConfigWithDependencyCheck(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	candidates := []discovery.SyncCandidate{
		{Path: "package.json", Format: parser.FormatJSON, Field: "version", Description: "Node.js"},
	}

	w := &Workflow{rootDir: tmpDir}

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.createConfigWithDependencyCheck(context.Background(), candidates)
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if !completed {
			t.Error("expected completed to be true")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify .sley.yaml was created with dependency-check
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(configData)
	if !strings.Contains(content, "dependency-check") {
		t.Errorf("expected dependency-check in config, got: %q", content)
	}

	// Verify output mentions sync files
	if !strings.Contains(output, "sync") || !strings.Contains(output, "package.json") {
		t.Errorf("expected sync file info in output, got: %q", output)
	}
}

func TestWorkflow_runInitWorkflow_NoSyncCandidates(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mock := &MockPrompter{
		ConfirmResult: false, // User declines init
	}

	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{},
		Modules:        []discovery.Module{},
	}

	w := NewWorkflow(mock, result, tmpDir)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runInitWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		// Should return false since no useful suggestions and user declined
		_ = completed
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Should mention no config found
	if !strings.Contains(output, "No .sley.yaml configuration found") {
		t.Errorf("expected no config message, got: %q", output)
	}
}

func TestWorkflow_runInitWorkflow_UserDeclines(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mock := &MockPrompter{
		ConfirmResult: false, // User declines init
	}

	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
		},
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runInitWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if completed {
			t.Error("expected completed to be false when user declines")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Should mention sley init
	if !strings.Contains(output, "sley init") {
		t.Errorf("expected sley init suggestion, got: %q", output)
	}
}

func TestWorkflow_runInitWorkflow_UserAccepts(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mock := &MockPrompter{
		ConfirmResult:     true,
		MultiSelectResult: []string{"package.json"},
	}

	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Field: "version", Description: "Node.js"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	_, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runInitWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if !completed {
			t.Error("expected completed to be true when user accepts")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify .sley.yaml was created
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("expected .sley.yaml to be created")
	}
}

func TestWorkflow_runInitWorkflow_ConfirmError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	expectedErr := errors.New("confirm error")
	mock := &MockPrompter{
		ConfirmErr: expectedErr,
	}

	result := &discovery.Result{
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	_, _ = testutils.CaptureStdout(func() {
		_, runErr := w.runInitWorkflow(context.Background())
		if !errors.Is(runErr, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, runErr)
		}
	})
}

func TestWorkflow_runExistingConfigWorkflow_WithMismatches(t *testing.T) {
	mock := &MockPrompter{}

	result := &discovery.Result{
		Mismatches: []discovery.Mismatch{
			{Source: "package.json", ExpectedVersion: "1.0.0", ActualVersion: "2.0.0"},
		},
	}

	w := NewWorkflow(mock, result, "/test")

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runExistingConfigWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if completed {
			t.Error("expected completed to be false (informational only)")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Should show mismatch warning
	if !strings.Contains(output, "mismatch") {
		t.Errorf("expected mismatch warning, got: %q", output)
	}
}

func TestWorkflow_runExistingConfigWorkflow_NoMismatches(t *testing.T) {
	mock := &MockPrompter{}

	result := &discovery.Result{
		Mismatches:     []discovery.Mismatch{},
		SyncCandidates: []discovery.SyncCandidate{},
	}

	w := NewWorkflow(mock, result, "/test")

	completed, err := w.runExistingConfigWorkflow(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if completed {
		t.Error("expected completed to be false")
	}
}
