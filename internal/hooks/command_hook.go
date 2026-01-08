package hooks

import (
	"context"
	"os"
	"os/exec"
	"time"
)

type CommandHook struct {
	Name    string
	Command string
}

func (h CommandHook) Run(ctx context.Context) error {
	// Add default timeout if context has no deadline
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", h.Command) //nolint:gosec // G204: intentional - user-configured hook commands
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (h CommandHook) HookName() string {
	return h.Name
}
