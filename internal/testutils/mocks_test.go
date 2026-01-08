package testutils

import (
	"context"
	"testing"
)

func TestMockCommitParser(t *testing.T) {
	t.Run("successful parse", func(t *testing.T) {
		mock := MockCommitParser{
			Label: "minor",
			Err:   nil,
		}

		if got := mock.Name(); got != "mock" {
			t.Errorf("Name() = %q, want %q", got, "mock")
		}

		label, err := mock.Parse([]string{"feat: new feature"})
		if err != nil {
			t.Errorf("Parse() unexpected error: %v", err)
		}
		if label != "minor" {
			t.Errorf("Parse() = %q, want %q", label, "minor")
		}
	})

	t.Run("parse with error", func(t *testing.T) {
		mock := MockCommitParser{
			Label: "",
			Err:   ErrMockParseFailed,
		}

		_, err := mock.Parse([]string{})
		if err != ErrMockParseFailed {
			t.Errorf("Parse() error = %v, want %v", err, ErrMockParseFailed)
		}
	})
}

// ErrMockParseFailed is a sentinel error for testing
var ErrMockParseFailed = testError{"mock parse failed"}

type testError struct {
	msg string
}

func (e testError) Error() string { return e.msg }

func TestMockHook(t *testing.T) {
	t.Run("successful hook", func(t *testing.T) {
		mock := MockHook{
			Name:      "pre-release",
			ShouldErr: false,
		}

		if got := mock.HookName(); got != "pre-release" {
			t.Errorf("HookName() = %q, want %q", got, "pre-release")
		}

		ctx := context.Background()
		if err := mock.Run(ctx); err != nil {
			t.Errorf("Run() unexpected error: %v", err)
		}
	})

	t.Run("failing hook", func(t *testing.T) {
		mock := MockHook{
			Name:      "validation",
			ShouldErr: true,
		}

		if got := mock.HookName(); got != "validation" {
			t.Errorf("HookName() = %q, want %q", got, "validation")
		}

		ctx := context.Background()
		err := mock.Run(ctx)
		if err == nil {
			t.Error("Run() expected error, got nil")
		}
		if err.Error() != "validation failed" {
			t.Errorf("Run() error = %q, want %q", err.Error(), "validation failed")
		}
	})
}

func TestWithMock(t *testing.T) {
	setupCalled := false
	testFuncCalled := false

	WithMock(
		func() {
			setupCalled = true
		},
		func() {
			testFuncCalled = true
		},
	)

	if !setupCalled {
		t.Error("WithMock should call setup function")
	}
	if !testFuncCalled {
		t.Error("WithMock should call test function")
	}
}
