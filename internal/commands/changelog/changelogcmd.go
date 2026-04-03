package changelog

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/plugins/changeloggenerator"
	"github.com/indaco/sley/internal/printer"
	"github.com/urfave/cli/v3"
)

// Run returns the "changelog" command.
func Run(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "changelog",
		Usage: "Manage changelog files",
		Commands: []*cli.Command{
			mergeCmd(cfg),
		},
	}
}

// mergeCmd returns the "merge" subcommand.
func mergeCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:      "merge",
		Usage:     "Merge versioned changelog files into unified CHANGELOG.md",
		UsageText: "sley changelog merge [--changes-dir .changes] [--output CHANGELOG.md] [--header-template path]",
		Description: `Merge all versioned changelog files from .changes directory into a unified CHANGELOG.md.

This command combines all versioned changelog files (.changes/v*.md) into a single
CHANGELOG.md file, sorted by version (newest first). It prepends a default header
or uses a custom header template if specified.

Examples:
  sley changelog merge
  sley changelog merge --changes-dir .changes --output CHANGELOG.md
  sley changelog merge --header-template .changes/header.md`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "changes-dir",
				Usage: "Directory containing versioned changelog files",
				Value: ".changes",
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "Output path for unified changelog",
				Value: "CHANGELOG.md",
			},
			&cli.StringFlag{
				Name:  "header-template",
				Usage: "Path to custom header template file",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runMergeCmd(cmd, cfg)
		},
	}
}

// runMergeCmd executes the merge operation.
func runMergeCmd(cmd *cli.Command, cfg *config.Config) error {
	// Check if changelog-generator plugin is enabled
	ty := printer.Typography()
	if !isChangelogGeneratorEnabled(cfg) {
		fmt.Println(ty.Compose(
			printer.Warning("Warning: The changelog-generator plugin is not enabled."),
			printer.Info("To enable it, add the following to your .sley.yaml:"),
			ty.CodeBlock("plugins:\n  changelog-generator:\n    enabled: true\n    mode: \"versioned\"", "yaml"),
			printer.Info("Proceeding with merge using default settings..."),
		))
	} else {
		// Warn if merge-after is set to something other than manual
		mergeAfter := cfg.Plugins.ChangelogGenerator.GetMergeAfter()
		if mergeAfter != "manual" {
			fmt.Println(ty.Compose(
				printer.Warning(fmt.Sprintf("Warning: 'merge-after' is set to '%s' in your configuration.", mergeAfter)),
				printer.Info("Versioned changelog files are already being merged automatically."),
				printer.Info("This manual merge command may result in duplicate entries or unexpected behavior."),
			))
		}
	}

	// Build generator config from flags, falling back to .sley.yaml settings
	genCfg := buildGeneratorConfig(cmd, cfg)

	// Create generator instance
	gen, err := changeloggenerator.NewGenerator(genCfg, changeloggenerator.NewGitOps())
	if err != nil {
		return fmt.Errorf("failed to create changelog generator: %w", err)
	}

	// Execute merge
	if err := gen.MergeVersionedFiles(); err != nil {
		return fmt.Errorf("failed to merge changelog files: %w", err)
	}

	printer.PrintFaint(fmt.Sprintf("Merged changelog files from %s into %s",
		printer.Info(genCfg.ChangesDir), printer.Info(genCfg.ChangelogPath)))

	return nil
}

// isChangelogGeneratorEnabled checks if the changelog-generator plugin is enabled.
func isChangelogGeneratorEnabled(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	if cfg.Plugins == nil {
		return false
	}
	if cfg.Plugins.ChangelogGenerator == nil {
		return false
	}
	return cfg.Plugins.ChangelogGenerator.Enabled
}

// buildGeneratorConfig creates a generator config from CLI flags and existing config.
func buildGeneratorConfig(cmd *cli.Command, cfg *config.Config) *changeloggenerator.Config {
	// Start with defaults
	genCfg := changeloggenerator.DefaultConfig()

	// Override from .sley.yaml if changelog-generator plugin is configured
	if cfg != nil && cfg.Plugins != nil && cfg.Plugins.ChangelogGenerator != nil {
		genCfg = changeloggenerator.FromConfigStruct(cfg.Plugins.ChangelogGenerator)
	}

	// Override from command flags (flags take precedence)
	if cmd.IsSet("changes-dir") {
		genCfg.ChangesDir = cmd.String("changes-dir")
	}
	if cmd.IsSet("output") {
		genCfg.ChangelogPath = cmd.String("output")
	}
	if cmd.IsSet("header-template") {
		genCfg.HeaderTemplate = cmd.String("header-template")
	}

	return genCfg
}
