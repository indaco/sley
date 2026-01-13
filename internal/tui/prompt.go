package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/indaco/sley/internal/workspace"
)

// ErrCanceled is returned when the user cancels the operation.
var ErrCanceled = fmt.Errorf("operation canceled by user")

// CustomKeyMap returns a KeyMap with Esc added as a quit key.
func CustomKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	// Add "esc" to the quit binding (alongside ctrl+c)
	km.Quit = key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("esc", "cancel"),
	)
	return km
}

// ModulePrompt implements the Prompter interface using charmbracelet/huh.
// It provides an interactive TUI for module selection in multi-module workspaces.
type ModulePrompt struct {
	modules []*workspace.Module
}

// NewModulePrompt creates a new ModulePrompt with the given modules.
func NewModulePrompt(modules []*workspace.Module) *ModulePrompt {
	return &ModulePrompt{modules: modules}
}

// Ensure ModulePrompt implements Prompter.
var _ Prompter = (*ModulePrompt)(nil)

// PromptModuleSelection shows an interactive module selection UI.
// First, it presents a choice: apply to all, select specific, or cancel.
// If "select specific" is chosen, it shows a multi-select checkbox interface.
func (p *ModulePrompt) PromptModuleSelection(modules []*workspace.Module) (Selection, error) {
	if len(modules) == 0 {
		return Selection{}, fmt.Errorf("no modules provided for selection")
	}

	// Override modules if provided (for testing)
	if len(modules) > 0 {
		p.modules = modules
	}

	// Show initial choice prompt
	choice, err := p.showInitialPrompt()
	if err != nil {
		return Selection{}, err
	}

	switch choice {
	case ChoiceAll:
		return AllModules(), nil

	case ChoiceSelect:
		return p.showMultiSelect()

	case ChoiceCancel:
		return CanceledSelection(), ErrCanceled

	default:
		return CanceledSelection(), ErrCanceled
	}
}

// showInitialPrompt presents the first choice to the user.
func (p *ModulePrompt) showInitialPrompt() (Choice, error) {
	var choice string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Found %d module%s with .version files", len(p.modules), Pluralize(len(p.modules)))).
				Description(p.formatModuleList()).
				Options(
					huh.NewOption("Apply to all modules", "all"),
					huh.NewOption("Select specific modules...", "select"),
					huh.NewOption("Cancel", "cancel"),
				).
				Value(&choice),
		),
	).WithKeyMap(CustomKeyMap())

	if err := form.Run(); err != nil {
		// User pressed Escape or Ctrl+C
		if errors.Is(err, huh.ErrUserAborted) {
			return ChoiceCancel, nil
		}
		return ChoiceCancel, fmt.Errorf("prompt failed: %w", err)
	}

	return ParseChoice(choice), nil
}

// showMultiSelect presents a checkbox interface for selecting specific modules.
func (p *ModulePrompt) showMultiSelect() (Selection, error) {
	var selected []string

	// Build options from modules using DisplayNameWithPath for disambiguation
	options := make([]huh.Option[string], len(p.modules))
	for i, mod := range p.modules {
		options[i] = huh.NewOption(mod.DisplayNameWithPath(), mod.Name)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select modules to operate on (Esc to cancel):").
				Description("Space: toggle | Enter: confirm").
				Options(options...).
				Value(&selected),
		),
	).WithKeyMap(CustomKeyMap())

	if err := form.Run(); err != nil {
		// User pressed Escape or Ctrl+C
		if errors.Is(err, huh.ErrUserAborted) {
			return CanceledSelection(), ErrCanceled
		}
		return Selection{}, fmt.Errorf("multi-select failed: %w", err)
	}

	if len(selected) == 0 {
		return CanceledSelection(), ErrCanceled
	}

	return SelectedModules(selected), nil
}

// ConfirmOperation asks for yes/no confirmation.
func (p *ModulePrompt) ConfirmOperation(message string) (bool, error) {
	return Confirm(message, "")
}

// Confirm asks for yes/no confirmation with a title and optional description.
// If description is empty, only the title is shown.
func Confirm(title, description string) (bool, error) {
	var confirmed bool

	confirm := huh.NewConfirm().
		Title(title).
		Value(&confirmed)

	if description != "" {
		confirm = confirm.Description(description)
	}

	form := huh.NewForm(
		huh.NewGroup(confirm),
	).WithKeyMap(CustomKeyMap())

	if err := form.Run(); err != nil {
		// User pressed Escape or Ctrl+C - treat as declined
		if errors.Is(err, huh.ErrUserAborted) {
			return false, nil
		}
		return false, fmt.Errorf("confirmation failed: %w", err)
	}

	return confirmed, nil
}

// Select shows a single-selection prompt and returns the selected value.
// Returns the zero value of T if the user cancels.
func Select[T comparable](title, description string, options []huh.Option[T]) (T, error) {
	var selected T

	selectField := huh.NewSelect[T]().
		Title(title).
		Options(options...).
		Value(&selected)

	if description != "" {
		selectField = selectField.Description(description)
	}

	form := huh.NewForm(
		huh.NewGroup(selectField),
	).WithKeyMap(CustomKeyMap())

	if err := form.Run(); err != nil {
		// User pressed Escape or Ctrl+C - return zero value
		if errors.Is(err, huh.ErrUserAborted) {
			var zero T
			return zero, nil
		}
		var zero T
		return zero, fmt.Errorf("selection failed: %w", err)
	}

	return selected, nil
}

// MultiSelect shows a multi-selection prompt and returns the selected values.
// The defaults parameter specifies which options are pre-selected.
func MultiSelect[T comparable](title, description string, options []huh.Option[T], defaults []T) ([]T, error) {
	selected := defaults

	multiSelectField := huh.NewMultiSelect[T]().
		Title(title).
		Options(options...).
		Value(&selected)

	if description != "" {
		multiSelectField = multiSelectField.Description(description)
	}

	form := huh.NewForm(
		huh.NewGroup(multiSelectField),
	).WithKeyMap(CustomKeyMap())

	if err := form.Run(); err != nil {
		// User pressed Escape or Ctrl+C
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrCanceled
		}
		return nil, fmt.Errorf("multi-select failed: %w", err)
	}

	return selected, nil
}

// formatModuleList returns a formatted list of modules for display.
func (p *ModulePrompt) formatModuleList() string {
	if len(p.modules) == 0 {
		return ""
	}

	// Show first few modules as a preview
	const maxPreview = 5
	preview := p.modules
	if len(p.modules) > maxPreview {
		preview = p.modules[:maxPreview]
	}

	var result strings.Builder
	for _, mod := range preview {
		// Use DisplayNameWithPath for disambiguation when modules have same name
		fmt.Fprintf(&result, "\n  - %s", mod.DisplayNameWithPath())
	}

	if len(p.modules) > maxPreview {
		fmt.Fprintf(&result, "\n  ... and %d more", len(p.modules)-maxPreview)
	}

	return result.String()
}

// Pluralize returns "s" if count != 1, empty string otherwise.
func Pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
