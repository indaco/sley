package hooks

import (
	"fmt"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/printer"
)

// LoadPreReleaseHooksFromConfig loads pre-release hooks from the configuration.
func LoadPreReleaseHooksFromConfig(cfg *config.Config) error {
	if cfg == nil || cfg.PreReleaseHooks == nil {
		return nil
	}

	for _, h := range cfg.PreReleaseHooks {
		for name, def := range h {
			if def.Command != "" {
				RegisterPreReleaseHook(CommandHook{
					Name:    name,
					Command: def.Command,
				})
			} else {
				printer.PrintWarning(fmt.Sprintf("Skipping pre-release hook %q: no command defined", name))
			}
		}
	}

	return nil
}
