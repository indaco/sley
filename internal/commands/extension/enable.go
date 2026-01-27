package extension

import (
	"context"
	"fmt"
	"strings"

	"github.com/indaco/sley/internal/extensionmgr"
	"github.com/indaco/sley/internal/printer"
	"github.com/urfave/cli/v3"
)

// enableCmd returns the "enable" subcommand.
func enableCmd() *cli.Command {
	return &cli.Command{
		Name:  "enable",
		Usage: "Enable a registered extension",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Name of the extension to enable",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExtensionEnable(cmd)
		},
	}
}

// runExtensionEnable enables a registered extension by setting enabled: true
// in the configuration file. It uses a surgical YAML replacement so that
// comments and formatting are preserved.
func runExtensionEnable(cmd *cli.Command) error {
	extensionName := cmd.String("name")
	if extensionName == "" {
		return fmt.Errorf("please provide an extension name to enable")
	}

	updater := extensionmgr.NewDefaultConfigUpdater(&extensionmgr.DefaultYAMLMarshaler{})
	if err := updater.SetExtensionEnabled(configFilePath, extensionName, true); err != nil {
		if strings.Contains(err.Error(), "not found in configuration") {
			printer.PrintWarning(fmt.Sprintf("extension %q not found", extensionName))
			return nil
		}
		printer.PrintError(fmt.Sprintf("failed to enable extension: %v", err))
		return nil
	}

	printer.PrintSuccess(fmt.Sprintf("Extension %q enabled.", extensionName))
	return nil
}
