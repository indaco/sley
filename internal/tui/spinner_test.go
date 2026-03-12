package tui

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewSpinner_Defaults(t *testing.T) {

	s := NewSpinner()

	if s.title != "Working..." {
		t.Errorf("expected default title 'Working...', got %q", s.title)
	}
	if s.spinnerType != SpinnerDots {
		t.Errorf("expected default type SpinnerDots, got %v", s.spinnerType)
	}
	if s.accessible {
		t.Error("expected accessible to be false by default")
	}
}

func TestNewSpinner_WithOptions(t *testing.T) {

	s := NewSpinner(
		WithSpinnerTitle("Custom title"),
		WithSpinnerType(SpinnerLine),
		WithAccessibleMode(true),
	)

	if s.title != "Custom title" {
		t.Errorf("expected title 'Custom title', got %q", s.title)
	}
	if s.spinnerType != SpinnerLine {
		t.Errorf("expected SpinnerLine, got %v", s.spinnerType)
	}
	if !s.accessible {
		t.Error("expected accessible to be true")
	}
}

func TestSpinnerTypeConstants(t *testing.T) {

	// Verify all spinner types are mapped

	types := []SpinnerType{SpinnerLine, SpinnerDots, SpinnerMiniDot, SpinnerPulse}

	for _, st := range types {
		if _, ok := spinnerTypeMap[st]; !ok {
			t.Errorf("spinner type %v not found in map", st)
		}
	}
}

func TestNewMultiModuleSpinner(t *testing.T) {

	s := NewMultiModuleSpinner(5, true)

	if s.total != 5 {
		t.Errorf("expected total 5, got %d", s.total)
	}
	if !s.accessible {
		t.Error("expected accessible to be true")
	}
	if s.title != "Processing modules..." {
		t.Errorf("expected default title, got %q", s.title)
	}
}

func TestMultiModuleSpinner_FormatTitle(t *testing.T) {

	s := NewMultiModuleSpinner(10, false)

	tests := []struct {
		name       string
		current    int
		moduleName string
		expected   string
	}{
		{
			name:       "with module name",
			current:    3,
			moduleName: "api",
			expected:   "[3/10] Processing modules...: api",
		},
		{
			name:       "without module name",
			current:    5,
			moduleName: "",
			expected:   "[5/10] Processing modules...",
		},
		{
			name:       "first module",
			current:    1,
			moduleName: "core",
			expected:   "[1/10] Processing modules...: core",
		},
		{
			name:       "last module",
			current:    10,
			moduleName: "cli",
			expected:   "[10/10] Processing modules...: cli",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result := s.formatTitle(tt.current, tt.moduleName)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMultiModuleSpinner_RunEach_ContextCancellation(t *testing.T) {

	s := NewMultiModuleSpinner(3, true) // accessible mode for no animation

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	modules := []string{"a", "b", "c"}
	err := s.RunEach(ctx, modules, func(_ context.Context, _ int) error {
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestMultiModuleSpinner_RunEach_ActionError(t *testing.T) {

	s := NewMultiModuleSpinner(3, true) // accessible mode for faster tests

	expectedErr := errors.New("action failed")
	modules := []string{"a", "b", "c"}

	err := s.RunEach(context.Background(), modules, func(_ context.Context, idx int) error {
		if idx == 1 {
			return expectedErr
		}
		return nil
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestMultiModuleSpinner_RunEach_Success(t *testing.T) {

	s := NewMultiModuleSpinner(3, true) // accessible mode

	modules := []string{"a", "b", "c"}
	executed := make([]int, 0, 3)

	err := s.RunEach(context.Background(), modules, func(_ context.Context, idx int) error {
		executed = append(executed, idx)
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(executed) != 3 {
		t.Errorf("expected 3 executions, got %d", len(executed))
	}

	for i, v := range executed {
		if v != i {
			t.Errorf("expected execution order [0,1,2], got %v", executed)
			break
		}
	}
}

func TestSpinner_Run_Accessible(t *testing.T) {

	s := NewSpinner(
		WithSpinnerTitle("Test"),
		WithAccessibleMode(true), // No animations
	)

	executed := false
	err := s.Run(func() {
		executed = true
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("action was not executed")
	}
}

func TestSpinner_RunWithErr_Accessible(t *testing.T) {

	s := NewSpinner(
		WithSpinnerTitle("Context test"),
		WithAccessibleMode(true),
	)

	executed := false
	err := s.RunWithErr(context.Background(), func(_ context.Context) error {
		executed = true
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("action was not executed")
	}
}

func TestSpinner_RunWithErr_Error(t *testing.T) {

	s := NewSpinner(WithAccessibleMode(true))

	expectedErr := errors.New("test error")
	err := s.RunWithErr(context.Background(), func(_ context.Context) error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestSpinner_RunWithErr_Cancellation(t *testing.T) {

	s := NewSpinner(WithAccessibleMode(true))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// This action would block forever without context cancellation
	err := s.RunWithErr(ctx, func(innerCtx context.Context) error {
		<-innerCtx.Done()
		return innerCtx.Err()
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestQuickSpinner_Accessible(t *testing.T) {

	// QuickSpinner uses non-accessible mode by default
	// We test with an inline spinner for controlled behavior

	s := NewSpinner(
		WithSpinnerTitle("Quick test"),
		WithAccessibleMode(true),
	)

	executed := false
	err := s.Run(func() {
		executed = true
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("action was not executed")
	}
}

func TestSpinnerAllTypes(t *testing.T) {

	types := []SpinnerType{
		SpinnerLine,
		SpinnerDots,
		SpinnerMiniDot,
		SpinnerPulse,
	}

	for _, spinnerType := range types {
		t.Run(spinnerType.String(), func(t *testing.T) {

			s := NewSpinner(
				WithSpinnerType(spinnerType),
				WithAccessibleMode(true),
			)

			executed := false
			err := s.Run(func() {
				executed = true
			})

			if err != nil {
				t.Errorf("unexpected error for type %v: %v", spinnerType, err)
			}
			if !executed {
				t.Errorf("action not executed for type %v", spinnerType)
			}
		})
	}
}

// String returns a string representation of the spinner type for testing.
func (t SpinnerType) String() string {
	switch t {
	case SpinnerLine:
		return "Line"
	case SpinnerDots:
		return "Dots"
	case SpinnerMiniDot:
		return "MiniDot"
	case SpinnerPulse:
		return "Pulse"
	default:
		return "Unknown"
	}
}

func TestQuickSpinnerWithErr(t *testing.T) {

	// Test the convenience function with accessible mode

	s := NewSpinner(
		WithSpinnerTitle("Quick error test"),
		WithAccessibleMode(true),
	)

	executed := false
	err := s.RunWithErr(context.Background(), func(_ context.Context) error {
		executed = true
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("action was not executed")
	}
}
