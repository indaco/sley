package hooks

import (
	"context"
	"testing"
	"time"
)

func TestCommandHook_Run_Success(t *testing.T) {
	h := CommandHook{
		Name:    "echo-test",
		Command: "echo 'hello world'",
	}

	ctx := context.Background()
	if err := h.Run(ctx); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestCommandHook_Run_Failure(t *testing.T) {
	h := CommandHook{
		Name:    "fail-test",
		Command: "exit 1",
	}

	ctx := context.Background()
	if err := h.Run(ctx); err == nil {
		t.Fatalf("expected failure, got nil")
	}
}

func TestCommandHook_HookName(t *testing.T) {
	h := CommandHook{Name: "hook-name"}
	if got := h.HookName(); got != "hook-name" {
		t.Errorf("expected 'hook-name', got %q", got)
	}
}

func TestCommandHook_Run_WithTimeout(t *testing.T) {
	h := CommandHook{
		Name:    "timeout-test",
		Command: "sleep 2",
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := h.Run(ctx)
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", ctx.Err())
	}
}

func TestCommandHook_Run_WithDefaultTimeout(t *testing.T) {
	h := CommandHook{
		Name:    "default-timeout-test",
		Command: "echo 'test'",
	}

	// Pass context without deadline - should get 30s default timeout
	ctx := context.Background()
	if err := h.Run(ctx); err != nil {
		t.Fatalf("expected success with default timeout, got error: %v", err)
	}
}
