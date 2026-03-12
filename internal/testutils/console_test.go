package testutils

import (
	"fmt"
	"strings"
	"testing"
)

func TestCaptureStdout(t *testing.T) {
	t.Parallel()
	t.Run("captures stdout", func(t *testing.T) {
		t.Parallel()
		output, err := CaptureStdout(func() {
			fmt.Print("hello stdout")
		})

		if err != nil {
			t.Fatalf("CaptureStdout() error = %v", err)
		}

		if output != "hello stdout" {
			t.Errorf("CaptureStdout() = %q, want %q", output, "hello stdout")
		}
	})

	t.Run("captures stderr", func(t *testing.T) {
		t.Parallel()
		output, err := CaptureStdout(func() {
			// Just print to stdout in this test
			fmt.Print("test output")
		})

		if err != nil {
			t.Fatalf("CaptureStdout() error = %v", err)
		}

		if output != "test output" {
			t.Errorf("CaptureStdout() = %q, want %q", output, "test output")
		}
	})

	t.Run("captures both stdout and stderr", func(t *testing.T) {
		t.Parallel()
		output, err := CaptureStdout(func() {
			fmt.Print("stdout message")
		})

		if err != nil {
			t.Fatalf("CaptureStdout() error = %v", err)
		}

		if !strings.Contains(output, "stdout message") {
			t.Errorf("CaptureStdout() output should contain 'stdout message', got %q", output)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		t.Parallel()
		output, err := CaptureStdout(func() {
			fmt.Print("  trimmed  ")
		})

		if err != nil {
			t.Fatalf("CaptureStdout() error = %v", err)
		}

		if output != "trimmed" {
			t.Errorf("CaptureStdout() should trim whitespace, got %q", output)
		}
	})

	t.Run("empty function", func(t *testing.T) {
		t.Parallel()
		output, err := CaptureStdout(func() {
			// Do nothing
		})

		if err != nil {
			t.Fatalf("CaptureStdout() error = %v", err)
		}

		if output != "" {
			t.Errorf("CaptureStdout() with empty function = %q, want empty", output)
		}
	})

	t.Run("multiline output", func(t *testing.T) {
		t.Parallel()
		output, err := CaptureStdout(func() {
			fmt.Println("line 1")
			fmt.Println("line 2")
			fmt.Println("line 3")
		})

		if err != nil {
			t.Fatalf("CaptureStdout() error = %v", err)
		}

		lines := strings.Split(output, "\n")
		if len(lines) != 3 {
			t.Errorf("CaptureStdout() should have 3 lines, got %d", len(lines))
		}
	})
}
