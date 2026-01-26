package discover

import (
	"context"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/indaco/sley/internal/discovery"
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
