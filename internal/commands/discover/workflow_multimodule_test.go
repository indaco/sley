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

func TestWorkflow_runMultiModuleSetup_WorkspaceChoice(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mock := &MockPrompter{
		SelectResult: string(WorkspaceChoiceWorkspace),
	}

	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
			{Name: "backend", RelPath: "backend/.version", Version: "1.0.0"},
			{Name: "frontend", RelPath: "frontend/.version", Version: "1.0.0"},
		},
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Field: "version"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runMultiModuleSetup(context.Background())
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

	// Should show module count
	if !strings.Contains(output, "3 modules") {
		t.Errorf("expected module count in output, got: %q", output)
	}

	// Should have called Select
	if mock.SelectCalls != 1 {
		t.Errorf("expected 1 Select call, got %d", mock.SelectCalls)
	}

	// Verify .sley.yaml was created with workspace config
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(configData)
	if !strings.Contains(content, "workspace:") {
		t.Errorf("expected workspace config in file, got: %q", content)
	}
	if !strings.Contains(content, "discovery:") {
		t.Errorf("expected discovery config in file, got: %q", content)
	}
}

func TestWorkflow_runMultiModuleSetup_SingleRootChoice(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mock := &MockPrompter{
		SelectResult:      string(WorkspaceChoiceSingleRoot),
		MultiSelectResult: []string{"package.json"},
	}

	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
			{Name: "backend", RelPath: "backend/.version", Version: "1.0.0"},
		},
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Field: "version", Description: "Node.js"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	_, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runMultiModuleSetup(context.Background())
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

	// Should have called Select and then MultiSelect for dependency-check
	if mock.SelectCalls != 1 {
		t.Errorf("expected 1 Select call, got %d", mock.SelectCalls)
	}
	if mock.MultiSelectCalls != 1 {
		t.Errorf("expected 1 MultiSelect call, got %d", mock.MultiSelectCalls)
	}

	// Verify .sley.yaml was created with dependency-check
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(configData)
	if !strings.Contains(content, "dependency-check:") {
		t.Errorf("expected dependency-check config in file, got: %q", content)
	}
	// Should NOT have workspace config for single-root
	if strings.Contains(content, "workspace:") {
		t.Errorf("should not have workspace config for single-root, got: %q", content)
	}
}

func TestWorkflow_runMultiModuleSetup_CoordinatedChoice(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mock := &MockPrompter{
		SelectResult: string(WorkspaceChoiceCoordinated),
	}

	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "2.0.0"},
			{Name: "gateway", RelPath: "gateway/.version", Version: "1.0.0"},
			{Name: "api", RelPath: "services/api/.version", Version: "1.0.0"},
			{Name: "frontend", RelPath: "frontend/.version", Version: "1.0.0"},
		},
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "frontend/package.json", Format: parser.FormatJSON, Field: "version", Description: "Node.js"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runMultiModuleSetup(context.Background())
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

	// Should show module count
	if !strings.Contains(output, "4 modules") {
		t.Errorf("expected module count in output, got: %q", output)
	}

	// Should have called Select
	if mock.SelectCalls != 1 {
		t.Errorf("expected 1 Select call, got %d", mock.SelectCalls)
	}

	// Verify .sley.yaml was created with dependency-check config (not workspace)
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(configData)

	// Should have dependency-check with all .version files as sync targets
	if !strings.Contains(content, "dependency-check:") {
		t.Errorf("expected dependency-check config in file, got: %q", content)
	}
	if !strings.Contains(content, "gateway/.version") {
		t.Errorf("expected gateway/.version in config, got: %q", content)
	}
	if !strings.Contains(content, "services/api/.version") {
		t.Errorf("expected services/api/.version in config, got: %q", content)
	}
	if !strings.Contains(content, "frontend/.version") {
		t.Errorf("expected frontend/.version in config, got: %q", content)
	}
	if !strings.Contains(content, "frontend/package.json") {
		t.Errorf("expected frontend/package.json in config, got: %q", content)
	}

	// Should NOT have workspace config for coordinated versioning
	if strings.Contains(content, "workspace:") {
		t.Errorf("should not have workspace config for coordinated versioning, got: %q", content)
	}
}

func TestWorkflow_runMultiModuleSetup_Canceled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mock := &MockPrompter{
		SelectResult: "", // User canceled
	}

	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runMultiModuleSetup(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if completed {
			t.Error("expected completed to be false when canceled")
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "canceled") {
		t.Errorf("expected canceled message in output, got: %q", output)
	}
}

func TestWorkflow_runMultiModuleSetup_SelectError(t *testing.T) {
	tmpDir := t.TempDir()

	expectedErr := errors.New("select error")
	mock := &MockPrompter{
		SelectErr: expectedErr,
	}

	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	_, _ = testutils.CaptureStdout(func() {
		_, runErr := w.runMultiModuleSetup(context.Background())
		if !errors.Is(runErr, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, runErr)
		}
	})
}

func TestWorkflow_createConfigWithWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
			{Name: "backend", RelPath: "backend/.version", Version: "1.0.0"},
		},
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Field: "version"},
		},
	}

	w := NewWorkflow(nil, result, tmpDir)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.createConfigWithWorkspace(context.Background())
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

	// Verify output mentions workspace
	if !strings.Contains(output, "workspace") {
		t.Errorf("expected workspace in output, got: %q", output)
	}

	// Verify .sley.yaml contains workspace config
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(configData)

	// Check workspace discovery config
	if !strings.Contains(content, "workspace:") {
		t.Errorf("expected workspace in config, got: %q", content)
	}
	if !strings.Contains(content, "discovery:") {
		t.Errorf("expected discovery in config, got: %q", content)
	}
	if !strings.Contains(content, "enabled: true") {
		t.Errorf("expected enabled: true in config, got: %q", content)
	}
	if !strings.Contains(content, "recursive: true") {
		t.Errorf("expected recursive: true in config, got: %q", content)
	}
	if !strings.Contains(content, "module_max_depth: 10") {
		t.Errorf("expected module_max_depth: 10 in config, got: %q", content)
	}
}

func TestWorkflow_createConfigWithCoordinatedVersioning(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "2.0.0"},
			{Name: "gateway", RelPath: "gateway/.version", Version: "1.0.0"},
			{Name: "api", RelPath: "services/api/.version", Version: "1.5.0"},
		},
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Field: "version", Description: "Node.js"},
			{Path: "Cargo.toml", Format: parser.FormatTOML, Field: "package.version", Description: "Rust"},
		},
	}

	w := NewWorkflow(nil, result, tmpDir)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.createConfigWithCoordinatedVersioning(context.Background())
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

	// Verify output mentions sync files
	if !strings.Contains(output, "sync") {
		t.Errorf("expected sync in output, got: %q", output)
	}

	// Verify .sley.yaml was created
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(configData)

	// Check dependency-check config
	if !strings.Contains(content, "dependency-check:") {
		t.Errorf("expected dependency-check in config, got: %q", content)
	}
	if !strings.Contains(content, "auto-sync: true") {
		t.Errorf("expected auto-sync: true in config, got: %q", content)
	}

	// Check submodule .version files are included (format: raw)
	if !strings.Contains(content, "gateway/.version") {
		t.Errorf("expected gateway/.version in config, got: %q", content)
	}
	if !strings.Contains(content, "services/api/.version") {
		t.Errorf("expected services/api/.version in config, got: %q", content)
	}
	if !strings.Contains(content, "format: raw") {
		t.Errorf("expected format: raw for .version files, got: %q", content)
	}

	// Check manifest files are included
	if !strings.Contains(content, "package.json") {
		t.Errorf("expected package.json in config, got: %q", content)
	}
	if !strings.Contains(content, "Cargo.toml") {
		t.Errorf("expected Cargo.toml in config, got: %q", content)
	}

	// Should NOT have workspace config
	if strings.Contains(content, "workspace:") {
		t.Errorf("should not have workspace config, got: %q", content)
	}
}

func TestWorkflow_createConfigWithCoordinatedVersioning_NoSubmodules(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Only root module (edge case)
	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
		},
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Field: "version", Description: "Node.js"},
		},
	}

	w := NewWorkflow(nil, result, tmpDir)

	_, err := testutils.CaptureStdout(func() {
		completed, runErr := w.createConfigWithCoordinatedVersioning(context.Background())
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

	// Verify .sley.yaml was created with only manifest files
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(configData)

	// Should have package.json but no submodule .version files
	if !strings.Contains(content, "package.json") {
		t.Errorf("expected package.json in config, got: %q", content)
	}
}

func TestWorkflow_createConfigWithCoordinatedVersioning_NoManifests(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Modules but no manifest files
	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
			{Name: "backend", RelPath: "backend/.version", Version: "1.0.0"},
			{Name: "frontend", RelPath: "frontend/.version", Version: "1.0.0"},
		},
		SyncCandidates: []discovery.SyncCandidate{}, // No manifests
	}

	w := NewWorkflow(nil, result, tmpDir)

	_, err := testutils.CaptureStdout(func() {
		completed, runErr := w.createConfigWithCoordinatedVersioning(context.Background())
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

	// Verify .sley.yaml was created with submodule .version files
	configPath := filepath.Join(tmpDir, ".sley.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(configData)

	// Should have submodule .version files
	if !strings.Contains(content, "backend/.version") {
		t.Errorf("expected backend/.version in config, got: %q", content)
	}
	if !strings.Contains(content, "frontend/.version") {
		t.Errorf("expected frontend/.version in config, got: %q", content)
	}
}

func TestWorkflow_runInitWorkflow_MultiModule(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	mock := &MockPrompter{
		ConfirmResult: true, // Accept init
		SelectResult:  string(WorkspaceChoiceWorkspace),
	}

	result := &discovery.Result{
		Mode: discovery.MultiModule,
		Modules: []discovery.Module{
			{Name: "root", RelPath: ".version", Version: "1.0.0"},
			{Name: "backend", RelPath: "backend/.version", Version: "1.0.0"},
		},
		SyncCandidates: []discovery.SyncCandidate{
			{Path: "package.json", Format: parser.FormatJSON, Field: "version"},
		},
	}

	w := NewWorkflow(mock, result, tmpDir)

	_, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runInitWorkflow(context.Background())
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

	// Verify Confirm and Select were called
	if mock.ConfirmCalls != 1 {
		t.Errorf("expected 1 Confirm call, got %d", mock.ConfirmCalls)
	}
	if mock.SelectCalls != 1 {
		t.Errorf("expected 1 Select call, got %d", mock.SelectCalls)
	}
}
