package flags

import "github.com/urfave/cli/v3"

// MultiModuleFlags returns flags for multi-module operations.
// These flags are shared by show, set, and bump commands.
func MultiModuleFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "Operate on all discovered modules",
		},
		&cli.StringFlag{
			Name:    "module",
			Aliases: []string{"m"},
			Usage:   "Operate on specific module by name",
		},
		&cli.StringSliceFlag{
			Name:  "modules",
			Usage: "Operate on multiple modules (comma-separated)",
		},
		&cli.StringFlag{
			Name:  "pattern",
			Usage: "Operate on modules matching glob pattern (e.g., 'services/*')",
		},
		&cli.BoolFlag{
			Name:    "yes",
			Aliases: []string{"y"},
			Usage:   "Auto-select all modules without prompting (implies --all)",
		},
		&cli.BoolFlag{
			Name:  "non-interactive",
			Usage: "Disable interactive prompts (CI mode)",
		},
		&cli.BoolFlag{
			Name:  "parallel",
			Usage: "Execute operations in parallel across modules",
		},
		&cli.BoolFlag{
			Name:  "fail-fast",
			Usage: "Stop execution on first error",
			Value: true,
		},
		&cli.BoolFlag{
			Name:  "continue-on-error",
			Usage: "Continue execution even if some modules fail",
		},
		&cli.BoolFlag{
			Name:    "quiet",
			Aliases: []string{"q"},
			Usage:   "Suppress module-level output, show summary only",
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "Output format: text, json, table",
			Value: "text",
		},
	}
}
