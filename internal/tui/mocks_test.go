package tui

import (
	"errors"
	"testing"

	"github.com/indaco/sley/internal/workspace"
)

func TestMockPrompter_PromptModuleSelection(t *testing.T) {

	modules := []*workspace.Module{
		{Name: "module-a", CurrentVersion: "1.0.0"},
		{Name: "module-b", CurrentVersion: "2.0.0"},
	}

	t.Run("returns pre-configured selection", func(t *testing.T) {

		mock := NewMockPrompter()
		expected := AllModules()
		mock.SelectionResult = expected

		got, err := mock.PromptModuleSelection(modules)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.All != expected.All {
			t.Errorf("Selection.All = %v, want %v", got.All, expected.All)
		}
	})

	t.Run("returns pre-configured error", func(t *testing.T) {

		mock := NewMockPrompter()
		expectedErr := errors.New("test error")
		mock.SelectionError = expectedErr

		_, err := mock.PromptModuleSelection(modules)
		if err != expectedErr {
			t.Errorf("error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("records call", func(t *testing.T) {

		mock := NewMockPrompter()
		mock.SelectionResult = AllModules()

		_, err := mock.PromptModuleSelection(modules)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mock.CallCount("PromptModuleSelection") != 1 {
			t.Errorf("CallCount = %d, want 1", mock.CallCount("PromptModuleSelection"))
		}

		lastCall := mock.LastCall("PromptModuleSelection")
		if lastCall == nil {
			t.Fatal("LastCall returned nil")
		}

		if len(lastCall.Modules) != len(modules) {
			t.Errorf("LastCall.Modules length = %d, want %d", len(lastCall.Modules), len(modules))
		}
	})
}

func TestMockPrompter_ConfirmOperation(t *testing.T) {

	t.Run("returns pre-configured result", func(t *testing.T) {

		mock := NewMockPrompter()
		mock.ConfirmResult = true

		got, err := mock.ConfirmOperation("test message")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !got {
			t.Error("ConfirmOperation() = false, want true")
		}
	})

	t.Run("returns pre-configured error", func(t *testing.T) {

		mock := NewMockPrompter()
		expectedErr := errors.New("test error")
		mock.ConfirmError = expectedErr

		_, err := mock.ConfirmOperation("test message")
		if err != expectedErr {
			t.Errorf("error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("records call", func(t *testing.T) {

		mock := NewMockPrompter()
		mock.ConfirmResult = true

		_, err := mock.ConfirmOperation("test message")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mock.CallCount("ConfirmOperation") != 1 {
			t.Errorf("CallCount = %d, want 1", mock.CallCount("ConfirmOperation"))
		}

		lastCall := mock.LastCall("ConfirmOperation")
		if lastCall == nil {
			t.Fatal("LastCall returned nil")
		}

		if lastCall.Message != "test message" {
			t.Errorf("LastCall.Message = %q, want %q", lastCall.Message, "test message")
		}
	})
}

func TestMockPrompter_Reset(t *testing.T) {

	mock := NewMockPrompter()
	mock.SelectionResult = AllModules()
	mock.SelectionError = errors.New("test error")
	mock.ConfirmResult = true
	mock.ConfirmError = errors.New("confirm error")

	modules := []*workspace.Module{{Name: "test"}}
	_, _ = mock.PromptModuleSelection(modules)
	_, _ = mock.ConfirmOperation("test")

	if len(mock.Calls) == 0 {
		t.Error("Expected calls to be recorded")
	}

	mock.Reset()

	if len(mock.Calls) != 0 {
		t.Errorf("Reset() did not clear Calls, got %d calls", len(mock.Calls))
	}

	if mock.SelectionResult.All {
		t.Error("Reset() did not clear SelectionResult")
	}

	if mock.SelectionError != nil {
		t.Error("Reset() did not clear SelectionError")
	}

	if mock.ConfirmResult {
		t.Error("Reset() did not clear ConfirmResult")
	}

	if mock.ConfirmError != nil {
		t.Error("Reset() did not clear ConfirmError")
	}
}

func TestMockPrompter_Chaining(t *testing.T) {

	mock := NewMockPrompter().
		WithSelectionResult(AllModules()).
		WithConfirmResult(true)

	if !mock.SelectionResult.All {
		t.Error("WithSelectionResult() did not set SelectionResult")
	}

	if !mock.ConfirmResult {
		t.Error("WithConfirmResult() did not set ConfirmResult")
	}
}

func TestMockPrompter_WithSelectionError(t *testing.T) {

	expectedErr := errors.New("selection error")
	mock := NewMockPrompter().WithSelectionError(expectedErr)

	if mock.SelectionError != expectedErr {
		t.Errorf("WithSelectionError() did not set error correctly, got %v, want %v", mock.SelectionError, expectedErr)
	}

	modules := []*workspace.Module{{Name: "test"}}
	_, err := mock.PromptModuleSelection(modules)
	if err != expectedErr {
		t.Errorf("PromptModuleSelection() returned error %v, want %v", err, expectedErr)
	}
}

func TestMockPrompter_WithConfirmError(t *testing.T) {

	expectedErr := errors.New("confirm error")
	mock := NewMockPrompter().WithConfirmError(expectedErr)

	if mock.ConfirmError != expectedErr {
		t.Errorf("WithConfirmError() did not set error correctly, got %v, want %v", mock.ConfirmError, expectedErr)
	}

	_, err := mock.ConfirmOperation("test message")
	if err != expectedErr {
		t.Errorf("ConfirmOperation() returned error %v, want %v", err, expectedErr)
	}
}

func TestMockPrompter_CallCount(t *testing.T) {

	mock := NewMockPrompter()
	mock.SelectionResult = AllModules()

	modules := []*workspace.Module{{Name: "test"}}

	// No calls yet
	if count := mock.CallCount("PromptModuleSelection"); count != 0 {
		t.Errorf("initial CallCount = %d, want 0", count)
	}

	// Make multiple calls
	_, _ = mock.PromptModuleSelection(modules)
	_, _ = mock.PromptModuleSelection(modules)
	_, _ = mock.PromptModuleSelection(modules)

	if count := mock.CallCount("PromptModuleSelection"); count != 3 {
		t.Errorf("CallCount = %d, want 3", count)
	}

	// Different method should not affect count
	_, _ = mock.ConfirmOperation("test")
	if count := mock.CallCount("PromptModuleSelection"); count != 3 {
		t.Errorf("CallCount after other call = %d, want 3", count)
	}
}

func TestMockPrompter_LastCall(t *testing.T) {

	mock := NewMockPrompter()
	mock.SelectionResult = AllModules()

	modules1 := []*workspace.Module{{Name: "module-a"}}
	modules2 := []*workspace.Module{{Name: "module-b"}}

	// No calls yet
	if lastCall := mock.LastCall("PromptModuleSelection"); lastCall != nil {
		t.Error("LastCall() should return nil when no calls made")
	}

	// Make calls
	_, _ = mock.PromptModuleSelection(modules1)
	_, _ = mock.PromptModuleSelection(modules2)

	lastCall := mock.LastCall("PromptModuleSelection")
	if lastCall == nil {
		t.Fatal("LastCall() returned nil")
	}

	// Should return the last call (modules2)
	if len(lastCall.Modules) != 1 || lastCall.Modules[0].Name != "module-b" {
		t.Error("LastCall() did not return the most recent call")
	}
}
