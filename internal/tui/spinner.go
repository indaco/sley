package tui

import (
	"context"
	"fmt"

	"charm.land/huh/v2/spinner"
)

// SpinnerType represents different spinner animation styles.
type SpinnerType int

const (
	// SpinnerLine uses line-based animation.
	SpinnerLine SpinnerType = iota
	// SpinnerDots uses dots animation.
	SpinnerDots
	// SpinnerMiniDot uses mini dot animation.
	SpinnerMiniDot
	// SpinnerPulse uses pulse animation.
	SpinnerPulse
)

// spinnerTypeMap maps our SpinnerType to huh spinner types.
var spinnerTypeMap = map[SpinnerType]spinner.Type{
	SpinnerLine:    spinner.Line,
	SpinnerDots:    spinner.Dots,
	SpinnerMiniDot: spinner.MiniDot,
	SpinnerPulse:   spinner.Pulse,
}

// Spinner provides progress indication for long-running operations.
// It wraps charmbracelet/huh/spinner for a consistent interface.
type Spinner struct {
	title       string
	spinnerType SpinnerType
	accessible  bool
}

// SpinnerOption configures a Spinner.
type SpinnerOption func(*Spinner)

// WithSpinnerTitle sets the spinner's title message.
func WithSpinnerTitle(title string) SpinnerOption {
	return func(s *Spinner) {
		s.title = title
	}
}

// WithSpinnerType sets the spinner's animation style.
func WithSpinnerType(t SpinnerType) SpinnerOption {
	return func(s *Spinner) {
		s.spinnerType = t
	}
}

// WithAccessibleMode enables accessible mode (no animations).
func WithAccessibleMode(accessible bool) SpinnerOption {
	return func(s *Spinner) {
		s.accessible = accessible
	}
}

// NewSpinner creates a new Spinner with the given options.
func NewSpinner(opts ...SpinnerOption) *Spinner {
	s := &Spinner{
		title:       "Working...",
		spinnerType: SpinnerDots,
		accessible:  false,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Run executes the given action with a spinner displayed.
// The action is run in a separate goroutine while the spinner animates.
// Returns any error from the action.
func (s *Spinner) Run(action func()) error {
	sp := spinner.New().
		Title(s.title).
		Type(spinnerTypeMap[s.spinnerType]).
		WithAccessible(s.accessible).
		Action(action)

	return sp.Run()
}

// RunWithErr executes the given action with context and error support.
// Returns any error from the action or context cancellation.
func (s *Spinner) RunWithErr(ctx context.Context, action func(context.Context) error) error {
	sp := spinner.New().
		Title(s.title).
		Type(spinnerTypeMap[s.spinnerType]).
		WithAccessible(s.accessible).
		Context(ctx).
		ActionWithErr(action)

	return sp.Run()
}

// ProgressReporter provides a way to update progress during multi-module operations.
type ProgressReporter interface {
	// Start begins showing progress.
	Start(title string)

	// Update updates the current progress state.
	Update(current, total int, currentModule string)

	// Finish completes the progress display.
	Finish()
}

// MultiModuleSpinner shows progress for multi-module operations.
type MultiModuleSpinner struct {
	title      string
	total      int
	accessible bool
}

// NewMultiModuleSpinner creates a spinner for multi-module operations.
func NewMultiModuleSpinner(total int, accessible bool) *MultiModuleSpinner {
	return &MultiModuleSpinner{
		title:      "Processing modules...",
		total:      total,
		accessible: accessible,
	}
}

// formatTitle creates a progress title string.
func (m *MultiModuleSpinner) formatTitle(current int, moduleName string) string {
	if moduleName != "" {
		return fmt.Sprintf("[%d/%d] %s: %s", current, m.total, m.title, moduleName)
	}
	return fmt.Sprintf("[%d/%d] %s", current, m.total, m.title)
}

// RunEach executes an action for each item with progress indication.
// The action receives the current index and should return an error if it fails.
func (m *MultiModuleSpinner) RunEach(
	ctx context.Context,
	moduleNames []string,
	action func(ctx context.Context, idx int) error,
) error {
	for i, name := range moduleNames {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		currentIdx := i
		title := m.formatTitle(i+1, name)
		sp := spinner.New().
			Title(title).
			Type(spinner.Dots).
			WithAccessible(m.accessible).
			Context(ctx).
			ActionWithErr(func(innerCtx context.Context) error {
				return action(innerCtx, currentIdx)
			})

		if err := sp.Run(); err != nil {
			return err
		}
	}

	return nil
}

// QuickSpinner runs a simple spinner with the given title and action.
// This is a convenience function for simple use cases.
func QuickSpinner(title string, action func()) error {
	return NewSpinner(WithSpinnerTitle(title)).Run(action)
}

// QuickSpinnerWithErr runs a spinner with context and error support.
func QuickSpinnerWithErr(ctx context.Context, title string, action func(context.Context) error) error {
	return NewSpinner(WithSpinnerTitle(title)).RunWithErr(ctx, action)
}
