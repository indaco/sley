package hooks

import (
	"context"
	"os"
	"os/exec"
)

type CommandHook struct {
	Name    string
	Command string
}

func (h CommandHook) Run() error {
	cmd := exec.CommandContext(context.Background(), "sh", "-c", h.Command) //nolint:gosec // G204: intentional - user-configured hook commands
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (h CommandHook) HookName() string {
	return h.Name
}
