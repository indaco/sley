package core

import (
	"context"
	"errors"
	"testing"
)

func TestMockCommandExecutor(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	ctx := context.Background()

	t.Run("run with default success", func(t *testing.T) {
		err := mockExec.Run(ctx, ".", "echo", "hello")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mockExec.Calls) != 1 {
			t.Errorf("expected 1 call, got %d", len(mockExec.Calls))
		}
		if mockExec.Calls[0].Command != "echo" {
			t.Errorf("expected 'echo' command, got %q", mockExec.Calls[0].Command)
		}
	})

	t.Run("output with set response", func(t *testing.T) {
		mockExec.SetResponse("git describe --tags", "v1.2.3\n")
		output, err := mockExec.Output(ctx, ".", "git", "describe", "--tags")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output != "v1.2.3\n" {
			t.Errorf("expected 'v1.2.3\\n', got %q", output)
		}
	})

	t.Run("error injection", func(t *testing.T) {
		mockExec.SetError("make test", errors.New("test failed"))
		_, err := mockExec.Output(ctx, ".", "make", "test")
		if err == nil || err.Error() != "test failed" {
			t.Errorf("expected 'test failed' error, got %v", err)
		}
	})

	t.Run("default error", func(t *testing.T) {
		mockExec := NewMockCommandExecutor()
		mockExec.DefaultError = errors.New("default error")

		err := mockExec.Run(ctx, ".", "unknown", "command")
		if err == nil || err.Error() != "default error" {
			t.Errorf("expected 'default error', got %v", err)
		}
	})
}

func TestMockCommandExecutor_DefaultOutput(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	ctx := context.Background()

	mockExec.DefaultOutput = "default response"
	output, err := mockExec.Output(ctx, ".", "unknown", "command")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "default response" {
		t.Errorf("expected 'default response', got %q", output)
	}
}

func TestMockCommandExecutor_Run_CommandKey(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	ctx := context.Background()

	expectedErr := errors.New("specific error")
	mockExec.SetError("git status", expectedErr)

	err := mockExec.Run(ctx, ".", "git", "status")
	if err != expectedErr {
		t.Errorf("expected specific error, got %v", err)
	}

	// Verify call was recorded
	if len(mockExec.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mockExec.Calls))
	}
	if mockExec.Calls[0].Command != "git" {
		t.Errorf("command = %q, want %q", mockExec.Calls[0].Command, "git")
	}
	if len(mockExec.Calls[0].Args) != 1 || mockExec.Calls[0].Args[0] != "status" {
		t.Errorf("args = %v, want [status]", mockExec.Calls[0].Args)
	}
}
