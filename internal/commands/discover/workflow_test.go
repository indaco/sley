package discover

import (
	"context"
	"strings"
	"testing"

	"charm.land/huh/v2"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/discovery"
	"github.com/indaco/sley/internal/testutils"
)

// MockPrompter is a test double for Prompter.
type MockPrompter struct {
	ConfirmResult     bool
	ConfirmErr        error
	MultiSelectResult []string
	MultiSelectErr    error
	SelectResult      string
	SelectErr         error

	ConfirmCalls     int
	MultiSelectCalls int
	SelectCalls      int
}

func (m *MockPrompter) Confirm(title, description string) (bool, error) {
	m.ConfirmCalls++
	return m.ConfirmResult, m.ConfirmErr
}

func (m *MockPrompter) MultiSelect(title, description string, options []huh.Option[string], defaults []string) ([]string, error) {
	m.MultiSelectCalls++
	return m.MultiSelectResult, m.MultiSelectErr
}

func (m *MockPrompter) Select(title, description string, options []huh.Option[string]) (string, error) {
	m.SelectCalls++
	return m.SelectResult, m.SelectErr
}

func TestNewWorkflow(t *testing.T) {

	mock := &MockPrompter{}
	result := &discovery.Result{}

	w := NewWorkflow(mock, result, "/test")

	if w.prompter != mock {
		t.Error("prompter mismatch")
	}
	if w.result != result {
		t.Error("result mismatch")
	}
	if w.rootDir != "/test" {
		t.Errorf("rootDir = %q, want %q", w.rootDir, "/test")
	}
}

func TestWorkflow_Run_NonInteractive(t *testing.T) {

	// When not interactive, Run should return false, nil

	mock := &MockPrompter{}
	result := &discovery.Result{}
	w := NewWorkflow(mock, result, "/test")

	// In non-interactive mode (test environment), should return early
	completed, err := w.Run(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// In test environment, tui.IsInteractive() returns false
	if completed {
		t.Log("Note: completed=true means test ran in interactive mode")
	}
}

func TestNewWorkflowWithConfig(t *testing.T) {
	mock := &MockPrompter{}
	result := &discovery.Result{}
	cfg := &config.Config{
		Workspace: &config.WorkspaceConfig{Versioning: "independent"},
	}

	w := NewWorkflowWithConfig(mock, result, "/test", cfg)

	if w.prompter != mock {
		t.Error("prompter mismatch")
	}
	if w.result != result {
		t.Error("result mismatch")
	}
	if w.rootDir != "/test" {
		t.Errorf("rootDir = %q, want %q", w.rootDir, "/test")
	}
	if w.cfg != cfg {
		t.Error("cfg mismatch")
	}
}

func TestRunMismatchWorkflow_Independent(t *testing.T) {
	result := &discovery.Result{
		Mismatches: []discovery.Mismatch{
			{Source: "sub/.version", ExpectedVersion: "1.0.0", ActualVersion: "2.0.0"},
		},
	}
	cfg := &config.Config{
		Workspace: &config.WorkspaceConfig{Versioning: "independent"},
	}

	mock := &MockPrompter{}
	w := NewWorkflowWithConfig(mock, result, "/test", cfg)

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runMismatchWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if completed {
			t.Error("expected completed=false for mismatch workflow")
		}
	})
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "independent versioning") {
		t.Errorf("expected 'independent versioning' in output, got: %q", output)
	}
	if !strings.Contains(output, "Version summary") {
		t.Errorf("expected 'Version summary' in output, got: %q", output)
	}
}

func TestRunMismatchWorkflow_Coordinated(t *testing.T) {
	result := &discovery.Result{
		Mismatches: []discovery.Mismatch{
			{Source: "package.json", ExpectedVersion: "1.0.0", ActualVersion: "2.0.0"},
		},
	}

	mock := &MockPrompter{}
	w := NewWorkflow(mock, result, "/test")

	output, err := testutils.CaptureStdout(func() {
		completed, runErr := w.runMismatchWorkflow(context.Background())
		if runErr != nil {
			t.Errorf("unexpected error: %v", runErr)
		}
		if completed {
			t.Error("expected completed=false for mismatch workflow")
		}
	})
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "mismatch") {
		t.Errorf("expected 'mismatch' in output, got: %q", output)
	}
}

// contains is a helper to check if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
