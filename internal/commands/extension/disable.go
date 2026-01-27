package extension

import (
	"context"
	"fmt"
	"strings"

	"github.com/indaco/sley/internal/extensionmgr"
	"github.com/indaco/sley/internal/printer"
	"github.com/urfave/cli/v3"
)

// disableCmd returns the "disable" subcommand.
func disableCmd() *cli.Command {
	return &cli.Command{
		Name:  "disable",
		Usage: "Disable a registered extension",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Name of the extension to disable",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExtensionDisable(cmd)
		},
	}
}

// runExtensionDisable disables a registered extension by setting enabled: false
// in the configuration file. It uses a surgical YAML replacement so that
// comments and formatting are preserved.
func runExtensionDisable(cmd *cli.Command) error {
	extensionName := cmd.String("name")
	if extensionName == "" {
		return fmt.Errorf("please provide an extension name to disable")
	}

	updater := extensionmgr.NewDefaultConfigUpdater(&extensionmgr.DefaultYAMLMarshaler{})
	if err := updater.SetExtensionEnabled(configFilePath, extensionName, false); err != nil {
		if strings.Contains(err.Error(), "not found in configuration") {
			printer.PrintWarning(fmt.Sprintf("extension %q not found", extensionName))
			return nil
		}
		printer.PrintError(fmt.Sprintf("failed to disable extension: %v", err))
		return nil
	}

	printer.PrintSuccess(fmt.Sprintf("Extension %q disabled.", extensionName))
	return nil
}
