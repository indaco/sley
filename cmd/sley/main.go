package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/plugins"
)

func main() {
	if err := runCLI(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runCLI(args []string) error {
	cfg, err := config.LoadConfigFn()
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	// Normalize or fallback
	cfg.Path = config.NormalizeVersionPath(cfg.Path)
	if cfg.Path == "" {
		cfg.Path = ".version"
	}

	// Create plugin registry and register builtin plugins
	registry := plugins.NewPluginRegistry()
	plugins.RegisterBuiltinPlugins(cfg, registry)

	if err := hooks.LoadPreReleaseHooksFromConfigFn(cfg); err != nil {
		return fmt.Errorf("failed to load pre-release hooks: %w", err)
	}

	app := newCLI(cfg, registry)
	return app.Run(context.Background(), args)
}
