// Package modulescmd provides commands for module discovery and management.
package modules

import (
	"github.com/urfave/cli/v3"
)

// Run returns the parent "modules" command with its subcommands.
func Run() *cli.Command {
	return &cli.Command{
		Name:    "modules",
		Aliases: []string{"mods"},
		Usage:   "Manage and discover modules in workspace",
		Commands: []*cli.Command{
			listCmd(),
			discoverCmd(),
		},
	}
}
