package initcmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
	"github.com/urfave/cli/v3"
)

// Run returns the "init" command.
func Run() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Initialize .version file and .sley.yaml configuration",
		UsageText: "sley init [options]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Use default settings without prompts",
			},
			&cli.StringFlag{
				Name:  "template",
				Usage: "Use a pre-configured template (basic, git, automation, strict, full)",
			},
			&cli.StringFlag{
				Name:  "enable",
				Usage: "Comma-separated list of plugins to enable (e.g., commit-parser,tag-manager)",
			},
			&cli.BoolFlag{
				Name:  "workspace",
				Usage: "Initialize as monorepo/workspace with module discovery",
			},
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Overwrite existing .sley.yaml if it exists",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runInitCmd(cmd)
		},
	}
}

// runInitCmd initializes a .version file and .sley.yaml configuration.
func runInitCmd(cmd *cli.Command) error {
	path := cmd.String("path")
	yesFlag := cmd.Bool("yes")
	templateFlag := cmd.String("template")
	enableFlag := cmd.String("enable")
	workspaceFlag := cmd.Bool("workspace")
	forceFlag := cmd.Bool("force")

	// Handle workspace mode differently
	if workspaceFlag {
		return runWorkspaceInit(path, yesFlag, templateFlag, enableFlag, forceFlag)
	}

	// Step 1: Initialize .version file if needed
	versionCreated, err := initializeVersionFile(path)
	if err != nil {
		return err
	}

	// Step 2: Detect project context
	projectCtx := DetectProjectContext()

	// Step 3: Determine which plugins to enable
	selectedPlugins, err := determinePlugins(projectCtx, yesFlag, templateFlag, enableFlag)
	if err != nil {
		return err
	}

	// If no plugins selected (user canceled), skip config creation
	if len(selectedPlugins) == 0 {
		if versionCreated {
			printVersionOnlySuccess(path)
		}
		return nil
	}

	// Step 4: Create .sley.yaml configuration
	configCreated, err := createConfigFile(selectedPlugins, forceFlag)
	if err != nil {
		return err
	}

	// Step 5: Print success messages and next steps
	printSuccessSummary(path, versionCreated, configCreated, selectedPlugins, projectCtx)

	return nil
}

// initializeVersionFile creates the .version file if it doesn't exist.
// Returns true if created, false if already existed.
func initializeVersionFile(path string) (bool, error) {
	created, err := semver.InitializeVersionFileWithFeedback(path)
	if err != nil {
		return false, err
	}

	// Verify we can read it
	_, err = semver.ReadVersion(path)
	if err != nil {
		return created, fmt.Errorf("failed to read version file at %s: %w", path, err)
	}

	return created, nil
}

// determinePlugins decides which plugins to enable based on flags and user input.
func determinePlugins(ctx *ProjectContext, yesFlag bool, templateFlag, enableFlag string) ([]string, error) {
	// Priority 1: --enable flag (most specific)
	if enableFlag != "" {
		return parseEnableFlag(enableFlag), nil
	}

	// Priority 2: --template flag
	if templateFlag != "" {
		template, err := GetTemplate(templateFlag)
		if err != nil {
			return nil, err
		}
		return template.Plugins, nil
	}

	// Priority 3: --yes flag (use defaults)
	if yesFlag {
		return DefaultPluginNames(), nil
	}

	// Priority 4: Interactive prompt
	if !isTerminalInteractive() {
		// Non-interactive terminal, use defaults
		return DefaultPluginNames(), nil
	}

	detectionSummary := ctx.FormatDetectionSummary()
	selectedPlugins, err := PromptPluginSelection(detectionSummary)
	if err != nil {
		// User canceled or error occurred
		return []string{}, nil
	}

	return selectedPlugins, nil
}

// parseEnableFlag parses the comma-separated --enable flag value.
func parseEnableFlag(enableFlag string) []string {
	plugins := []string{}
	for name := range strings.SplitSeq(enableFlag, ",") {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			plugins = append(plugins, trimmed)
		}
	}
	return plugins
}

// createConfigFile generates and writes the .sley.yaml configuration file.
// Returns true if created, false if skipped due to existing file.
func createConfigFile(selectedPlugins []string, forceFlag bool) (bool, error) {
	configPath := ".sley.yaml"

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !forceFlag {
		// Config exists and force not specified
		if !isTerminalInteractive() {
			// Non-interactive: skip creation
			return false, nil
		}

		// Interactive: ask for confirmation
		confirmed, err := ConfirmOverwrite()
		if err != nil {
			return false, err
		}
		if !confirmed {
			return false, nil
		}
	}

	// Generate config with comments
	configData, err := GenerateConfigWithComments(selectedPlugins)
	if err != nil {
		return false, fmt.Errorf("failed to generate config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, configData, 0600); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}

	return true, nil
}

// printVersionOnlySuccess prints a message when only .version was created (no config).
func printVersionOnlySuccess(path string) {
	version, err := semver.ReadVersion(path)
	if err != nil {
		printer.PrintSuccess(fmt.Sprintf("Initialized %s", path))
		return
	}
	printer.PrintSuccess(fmt.Sprintf("Initialized %s with version %s", path, version.String()))
}

// printSuccessSummary prints the final success message with next steps.
func printSuccessSummary(path string, versionCreated, configCreated bool, plugins []string, ctx *ProjectContext) {
	version, err := semver.ReadVersion(path)

	if versionCreated {
		if err == nil {
			printer.PrintSuccess(fmt.Sprintf("Created %s with version %s", path, version.String()))
		} else {
			printer.PrintSuccess(fmt.Sprintf("Created %s", path))
		}
	}

	if configCreated {
		pluginCount := len(plugins)
		printer.PrintSuccess(fmt.Sprintf("Created .sley.yaml with %d plugin%s enabled", pluginCount, pluralize(pluginCount)))
	}

	// Print next steps
	if configCreated || versionCreated {
		fmt.Println()
		printer.PrintInfo("Next steps:")
		if configCreated {
			fmt.Println("  - Review .sley.yaml and adjust settings")
		}
		if ctx.IsGitRepo {
			fmt.Println("  - Run 'sley bump patch' to increment version")
		}
		if len(plugins) > 0 {
			fmt.Println("  - Run 'sley doctor' to verify setup (if available)")
		}
	}
}

// isTerminalInteractive checks if stdin is connected to a terminal.
func isTerminalInteractive() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// pluralize returns "s" if count != 1.
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
