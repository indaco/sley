package tui

import (
	"testing"
)

func TestChoice_String(t *testing.T) {

	tests := []struct {
		name     string
		choice   Choice
		expected string
	}{
		{"all", ChoiceAll, "all"},
		{"select", ChoiceSelect, "select"},
		{"cancel", ChoiceCancel, "cancel"},
		{"unknown", Choice(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := tt.choice.String()
			if got != tt.expected {
				t.Errorf("Choice.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseChoice(t *testing.T) {

	tests := []struct {
		name     string
		input    string
		expected Choice
	}{
		{"all", "all", ChoiceAll},
		{"select", "select", ChoiceSelect},
		{"cancel", "cancel", ChoiceCancel},
		{"unknown defaults to cancel", "unknown", ChoiceCancel},
		{"empty defaults to cancel", "", ChoiceCancel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := ParseChoice(tt.input)
			if got != tt.expected {
				t.Errorf("ParseChoice(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAllModules(t *testing.T) {

	selection := AllModules()

	if !selection.All {
		t.Error("AllModules() should set All to true")
	}

	if len(selection.Modules) != 0 {
		t.Errorf("AllModules() should have empty Modules, got %v", selection.Modules)
	}

	if selection.Canceled {
		t.Error("AllModules() should not be canceled")
	}
}

func TestSelectedModules(t *testing.T) {

	names := []string{"module-a", "module-b", "module-c"}
	selection := SelectedModules(names)

	if selection.All {
		t.Error("SelectedModules() should not set All to true")
	}

	if len(selection.Modules) != 3 {
		t.Errorf("SelectedModules() should have 3 modules, got %d", len(selection.Modules))
	}

	for i, name := range names {
		if selection.Modules[i] != name {
			t.Errorf("SelectedModules()[%d] = %v, want %v", i, selection.Modules[i], name)
		}
	}

	if selection.Canceled {
		t.Error("SelectedModules() should not be canceled")
	}
}

func TestCanceledSelection(t *testing.T) {

	selection := CanceledSelection()

	if !selection.Canceled {
		t.Error("CanceledSelection() should set Canceled to true")
	}

	if selection.All {
		t.Error("CanceledSelection() should not set All to true")
	}

	if len(selection.Modules) != 0 {
		t.Errorf("CanceledSelection() should have empty Modules, got %v", selection.Modules)
	}
}

func TestSelection_ZeroValue(t *testing.T) {

	var selection Selection

	if selection.All {
		t.Error("zero-value Selection should have All = false")
	}

	if selection.Canceled {
		t.Error("zero-value Selection should have Canceled = false")
	}

	if selection.Modules != nil {
		t.Errorf("zero-value Selection should have nil Modules, got %v", selection.Modules)
	}
}

func TestSelectedModules_EmptySlice(t *testing.T) {

	selection := SelectedModules([]string{})

	if selection.All {
		t.Error("SelectedModules([]) should not set All to true")
	}

	if selection.Canceled {
		t.Error("SelectedModules([]) should not be canceled")
	}

	if selection.Modules == nil {
		t.Error("SelectedModules([]) should have non-nil Modules")
	}

	if len(selection.Modules) != 0 {
		t.Errorf("SelectedModules([]) should have 0 modules, got %d", len(selection.Modules))
	}
}

func TestSelectedModules_NilSlice(t *testing.T) {

	selection := SelectedModules(nil)

	if selection.All {
		t.Error("SelectedModules(nil) should not set All to true")
	}

	if selection.Canceled {
		t.Error("SelectedModules(nil) should not be canceled")
	}

	if selection.Modules != nil {
		t.Errorf("SelectedModules(nil) should have nil Modules, got %v", selection.Modules)
	}
}
