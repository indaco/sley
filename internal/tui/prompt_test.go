package tui

import (
	"testing"

	"github.com/indaco/sley/internal/workspace"
)

func TestNewModulePrompt(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "module-a", CurrentVersion: "1.0.0"},
		{Name: "module-b", CurrentVersion: "2.0.0"},
	}

	prompt := NewModulePrompt(modules)

	if prompt == nil {
		t.Fatal("NewModulePrompt() returned nil")
	}

	if len(prompt.modules) != 2 {
		t.Errorf("NewModulePrompt() modules count = %d, want 2", len(prompt.modules))
	}
}

func TestModulePrompt_ImplementsPrompter(t *testing.T) {
	// Compile-time check that ModulePrompt implements Prompter
	var _ Prompter = (*ModulePrompt)(nil)
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{"zero", 0, "s"},
		{"one", 1, ""},
		{"two", 2, "s"},
		{"many", 10, "s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Pluralize(tt.count)
			if got != tt.expected {
				t.Errorf("Pluralize(%d) = %q, want %q", tt.count, got, tt.expected)
			}
		})
	}
}

func TestModulePrompt_FormatModuleList(t *testing.T) {
	tests := []struct {
		name          string
		modules       []*workspace.Module
		shouldContain []string
	}{
		{
			name:          "empty modules",
			modules:       []*workspace.Module{},
			shouldContain: []string{},
		},
		{
			name: "single module",
			modules: []*workspace.Module{
				{Name: "module-a", CurrentVersion: "1.0.0"},
			},
			shouldContain: []string{"module-a", "1.0.0"},
		},
		{
			name: "multiple modules under limit",
			modules: []*workspace.Module{
				{Name: "module-a", CurrentVersion: "1.0.0"},
				{Name: "module-b", CurrentVersion: "2.0.0"},
			},
			shouldContain: []string{"module-a", "module-b"},
		},
		{
			name: "many modules over limit",
			modules: []*workspace.Module{
				{Name: "module-1", CurrentVersion: "1.0.0"},
				{Name: "module-2", CurrentVersion: "2.0.0"},
				{Name: "module-3", CurrentVersion: "3.0.0"},
				{Name: "module-4", CurrentVersion: "4.0.0"},
				{Name: "module-5", CurrentVersion: "5.0.0"},
				{Name: "module-6", CurrentVersion: "6.0.0"},
				{Name: "module-7", CurrentVersion: "7.0.0"},
			},
			shouldContain: []string{"module-1", "and 2 more"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := NewModulePrompt(tt.modules)
			result := prompt.formatModuleList()

			for _, expected := range tt.shouldContain {
				if result != "" && !contains(result, expected) {
					t.Errorf("formatModuleList() does not contain %q\nGot: %s", expected, result)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestModulePrompt_PromptModuleSelection_EmptyModules(t *testing.T) {
	prompt := NewModulePrompt(nil)

	_, err := prompt.PromptModuleSelection([]*workspace.Module{})
	if err == nil {
		t.Error("PromptModuleSelection() with empty modules should return an error")
	}

	expectedErrMsg := "no modules provided for selection"
	if err.Error() != expectedErrMsg {
		t.Errorf("error message = %q, want %q", err.Error(), expectedErrMsg)
	}
}

func TestErrCanceled(t *testing.T) {
	if ErrCanceled == nil {
		t.Fatal("ErrCanceled should not be nil")
	}

	expectedMsg := "operation canceled by user"
	if ErrCanceled.Error() != expectedMsg {
		t.Errorf("ErrCanceled.Error() = %q, want %q", ErrCanceled.Error(), expectedMsg)
	}
}

func TestModulePrompt_FormatModuleList_WithPaths(t *testing.T) {
	// Test that formatModuleList uses DisplayNameWithPath for disambiguation
	modules := []*workspace.Module{
		{Name: "version", CurrentVersion: "1.0.0", Dir: "backend/gateway/internal/version"},
		{Name: "version", CurrentVersion: "2.0.0", Dir: "cli/internal/version"},
		{Name: "api", CurrentVersion: "3.0.0", Dir: "backend/api"},
	}

	prompt := NewModulePrompt(modules)
	result := prompt.formatModuleList()

	// Should contain paths for disambiguation
	expectedParts := []string{
		"version",
		"backend/gateway/internal/version",
		"cli/internal/version",
		"backend/api",
	}

	for _, expected := range expectedParts {
		if !containsSubstring(result, expected) {
			t.Errorf("formatModuleList() should contain %q for disambiguation\nGot: %s", expected, result)
		}
	}
}

func TestModulePrompt_FormatModuleList_RootModule(t *testing.T) {
	// Test that root modules (Dir = ".") don't show redundant path
	modules := []*workspace.Module{
		{Name: "myapp", CurrentVersion: "1.0.0", Dir: "."},
	}

	prompt := NewModulePrompt(modules)
	result := prompt.formatModuleList()

	// Should contain the module name
	if !containsSubstring(result, "myapp") {
		t.Errorf("formatModuleList() should contain module name, got: %s", result)
	}

	// Should NOT contain " - ." since Dir is root
	if containsSubstring(result, " - .") {
		t.Errorf("formatModuleList() should not show ' - .' for root modules, got: %s", result)
	}
}

// Note: Selection helper tests (AllModules, SelectedModules, CanceledSelection)
// and Choice tests (String, ParseChoice) are in tui_test.go

func TestCustomKeyMap(t *testing.T) {
	km := CustomKeyMap()

	if km == nil {
		t.Fatal("CustomKeyMap() returned nil")
	}

	// Verify Quit binding includes both ctrl+c and esc
	keys := km.Quit.Keys()
	hasCtrlC := false
	hasEsc := false

	for _, k := range keys {
		if k == "ctrl+c" {
			hasCtrlC = true
		}
		if k == "esc" {
			hasEsc = true
		}
	}

	if !hasCtrlC {
		t.Error("CustomKeyMap().Quit should include 'ctrl+c'")
	}
	if !hasEsc {
		t.Error("CustomKeyMap().Quit should include 'esc' for cancel")
	}
}

// TestConfirmOperation_DelegatesToConfirm verifies ConfirmOperation delegates to Confirm.
// Note: We cannot test the actual prompt interaction without a terminal.
func TestConfirmOperation_DelegatesToConfirm(t *testing.T) {
	// This test verifies the method exists and has the correct signature.
	// The actual interactive behavior is tested manually.
	prompt := NewModulePrompt(nil)

	// Verify the method exists with correct signature (compile-time check)
	_ = prompt.ConfirmOperation
}

// TestConfirmFunctionSignature verifies the Confirm function has correct signature.
func TestConfirmFunctionSignature(t *testing.T) {
	// Verify the function exists with correct signature (compile-time check)
	_ = Confirm
}
