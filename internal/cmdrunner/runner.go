package cmdrunner

import (
	"context"
	"os"
	"os/exec"

	"github.com/indaco/sley/internal/apperrors"
	"github.com/indaco/sley/internal/core"
)

// Default timeouts for command execution.
// These reference the centralized timeout constants in core package.
const (
	DefaultTimeout       = core.TimeoutDefault
	DefaultOutputTimeout = core.TimeoutShort
)

// RunCommandContext executes a command with the given context.
// The context should be used to control cancellation and timeouts.
func RunCommandContext(ctx context.Context, dir string, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &apperrors.CommandError{Command: command, Err: err, Timeout: true}
		}
		return &apperrors.CommandError{Command: command, Err: err, Timeout: false}
	}

	return nil
}

// RunCommandOutputContext executes a command and returns its output.
// The context should be used to control cancellation and timeouts.
func RunCommandOutputContext(ctx context.Context, dir string, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return "", &apperrors.CommandError{Command: command, Err: err, Timeout: true}
	}

	if err != nil {
		return "", &apperrors.CommandError{Command: command, Err: err, Timeout: false}
	}

	return string(output), nil
}
