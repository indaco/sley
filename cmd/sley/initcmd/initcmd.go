package initcmd

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
	"github.com/urfave/cli/v3"
)

// Run returns the "init" command.
func Run() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Initialize a .version file (auto-detects Git tag or starts from 0.1.0)",
		UsageText: "sley init",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runInitCmd(cmd)
		},
	}
}

// runInitCmd initializes a .version file if not present.
func runInitCmd(cmd *cli.Command) error {
	path := cmd.String("path")

	created, err := semver.InitializeVersionFileWithFeedback(path)
	if err != nil {
		return err
	}

	version, err := semver.ReadVersion(path)
	if err != nil {
		return fmt.Errorf("failed to read version file at %s: %w", path, err)
	}

	if created {
		printer.PrintSuccess(fmt.Sprintf("Initialized %s with version %s", path, version.String()))
	} else {
		printer.PrintInfo(fmt.Sprintf("Version file already exists at %s", path))
	}
	return nil
}
