package bumpcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins/changelogparser"
	"github.com/indaco/sley/internal/plugins/commitparser"
	"github.com/indaco/sley/internal/plugins/commitparser/gitlog"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
	"github.com/urfave/cli/v3"
)

var (
	tryInferBumpTypeFromCommitParserPluginFn    = tryInferBumpTypeFromCommitParserPlugin
	tryInferBumpTypeFromChangelogParserPluginFn = tryInferBumpTypeFromChangelogParserPlugin
)

// autoCmd returns the "auto" subcommand.
func autoCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:    "auto",
		Aliases: []string{"next"},
		Usage:   "Smart bump logic (e.g. promote pre-release or bump patch)",
		UsageText: `sley bump auto [--label patch|minor|major] [--meta data] [--preserve-meta] [--since ref] [--until ref] [--no-infer] [--all] [--module name]

By default, sley tries to infer the bump type from recent commit messages using the built-in commit-parser plugin.
You can override this behavior with the --label flag, disable it explicitly with --no-infer, or disable the plugin via the config file (.sley.yaml).`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "label",
				Usage: "Optional bump label override (patch, minor, major)",
			},
			&cli.StringFlag{
				Name:  "meta",
				Usage: "Set build metadata (e.g. 'ci.123')",
			},
			&cli.BoolFlag{
				Name:  "preserve-meta",
				Usage: "Preserve existing build metadata instead of clearing it",
			},
			&cli.StringFlag{
				Name:  "since",
				Usage: "Start commit/tag for bump inference (default: last tag or HEAD~10)",
			},
			&cli.StringFlag{
				Name:  "until",
				Usage: "End commit/tag for bump inference (default: HEAD)",
			},
			&cli.BoolFlag{
				Name:  "no-infer",
				Usage: "Disable bump inference from commit messages (overrides config)",
			},
			&cli.BoolFlag{
				Name:  "hook-only",
				Usage: "Only run pre-release hooks, do not modify the version",
			},
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpAuto(ctx, cfg, cmd)
		},
	}
}

// runBumpAuto performs smart bumping (e.g. promote, patch, infer).
func runBumpAuto(ctx context.Context, cfg *config.Config, cmd *cli.Command) error {
	label := cmd.String("label")
	meta := cmd.String("meta")
	since := cmd.String("since")
	until := cmd.String("until")
	isPreserveMeta := cmd.Bool("preserve-meta")
	isNoInferFlag := cmd.Bool("no-infer")
	isSkipHooks := cmd.Bool("skip-hooks")

	disableInfer := isNoInferFlag || (cfg != nil && cfg.Plugins != nil && !cfg.Plugins.CommitParser)

	// Run pre-release hooks first (before any version operations)
	if err := hooks.RunPreReleaseHooksFn(isSkipHooks); err != nil {
		return err
	}

	// Get execution context to determine single vs multi-module mode
	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	// Handle single-module mode
	if execCtx.IsSingleModule() {
		return runSingleModuleAuto(cmd, execCtx.Path, label, meta, since, until, isPreserveMeta, disableInfer)
	}

	// Handle multi-module mode
	// For auto bump, we need to determine the bump type first
	bumpType := determineBumpType(label, disableInfer, since, until)
	return runMultiModuleBump(ctx, cmd, execCtx, bumpType, "", meta, isPreserveMeta)
}

// determineBumpType determines the bump type for multi-module auto bump.
func determineBumpType(label string, disableInfer bool, since, until string) operations.BumpType {
	switch label {
	case "patch":
		return operations.BumpPatch
	case "minor":
		return operations.BumpMinor
	case "major":
		return operations.BumpMajor
	case "":
		if !disableInfer {
			// Try changelog parser first if it should take precedence
			inferred := tryInferBumpTypeFromChangelogParserPluginFn()
			if inferred == "" {
				// Fall back to commit parser
				inferred = tryInferBumpTypeFromCommitParserPluginFn(since, until)
			}

			if inferred != "" {
				fmt.Fprintf(os.Stderr, "Inferred bump type: %s\n", inferred)
				switch inferred {
				case "minor":
					return operations.BumpMinor
				case "major":
					return operations.BumpMajor
				default:
					return operations.BumpPatch
				}
			}
		}
		// Default to auto which will handle pre-release promotion or patch bump
		return operations.BumpAuto
	default:
		// Invalid label, will be caught during execution
		return operations.BumpAuto
	}
}

// runSingleModuleAuto handles the single-module auto bump operation.
func runSingleModuleAuto(cmd *cli.Command, path, label, meta, since, until string, isPreserveMeta, disableInfer bool) error {
	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	current, err := semver.ReadVersion(path)
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	next, err := getNextVersion(current, label, disableInfer, since, until, isPreserveMeta)
	if err != nil {
		return err
	}

	next = setBuildMetadata(current, next, meta, isPreserveMeta)

	// Validate release gates before bumping
	if err := validateReleaseGate(next, current, "auto"); err != nil {
		return err
	}

	// Validate version policy before bumping
	if err := validateVersionPolicy(next, current, "auto"); err != nil {
		return err
	}

	// Validate dependency consistency before bumping
	if err := validateDependencyConsistency(next); err != nil {
		return err
	}

	// Validate tag availability before bumping
	if err := validateTagAvailable(next); err != nil {
		return err
	}

	if err := semver.SaveVersion(path, next); err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	// Generate changelog entry
	if err := generateChangelogAfterBump(next, current, "auto"); err != nil {
		return err
	}

	// Record audit log entry
	if err := recordAuditLogEntry(next, current, "auto"); err != nil {
		return err
	}

	printer.PrintSuccess(fmt.Sprintf("Bumped version from %s to %s", current.String(), next.String()))
	return nil
}

// getNextVersion determines the next semantic version based on the provided label,
// commit inference, or default bump logic. It returns an error if bumping fails
// or if an invalid label is specified.
func getNextVersion(
	current semver.SemVersion,
	label string,
	disableInfer bool,
	since, until string,
	preserveMeta bool,
) (semver.SemVersion, error) {
	var next semver.SemVersion
	var err error

	switch label {
	case "patch", "minor", "major":
		next, err = semver.BumpByLabelFunc(current, label)
		if err != nil {
			return semver.SemVersion{}, fmt.Errorf("failed to bump version with label: %w", err)
		}
	case "":
		if !disableInfer {
			// Try changelog parser first if it should take precedence
			inferred := tryInferBumpTypeFromChangelogParserPluginFn()
			if inferred == "" {
				// Fall back to commit parser
				inferred = tryInferBumpTypeFromCommitParserPluginFn(since, until)
			}

			if inferred != "" {
				fmt.Fprintf(os.Stderr, "Inferred bump type: %s\n", inferred)

				if current.PreRelease != "" {
					return promotePreRelease(current, preserveMeta), nil
				}
				next, err = semver.BumpByLabelFunc(current, inferred)
				if err != nil {
					return semver.SemVersion{}, fmt.Errorf("failed to bump inferred version: %w", err)
				}
				return next, nil
			}
		}

		next, err = semver.BumpNextFunc(current)
		if err != nil {
			return semver.SemVersion{}, fmt.Errorf("failed to determine next version: %w", err)
		}
	default:
		return semver.SemVersion{}, cli.Exit("invalid --label: must be 'patch', 'minor', or 'major'", 1)
	}

	return next, nil
}

// setBuildMetadata updates the build metadata of the next version based on
// the provided meta string and the preserve flag.
func setBuildMetadata(current, next semver.SemVersion, meta string, preserve bool) semver.SemVersion {
	switch {
	case meta != "":
		next.Build = meta
	case preserve:
		next.Build = current.Build
	default:
		next.Build = ""
	}
	return next
}

// promotePreRelease strips pre-release and optionally preserves metadata.
func promotePreRelease(current semver.SemVersion, preserveMeta bool) semver.SemVersion {
	next := current
	next.PreRelease = ""
	if preserveMeta {
		next.Build = current.Build
	} else {
		next.Build = ""
	}
	return next
}

// tryInferBumpTypeFromCommitParserPlugin tries to infer bump type from commit messages.
func tryInferBumpTypeFromCommitParserPlugin(since, until string) string {
	parser := commitparser.GetCommitParserFn()
	if parser == nil {
		return ""
	}

	commits, err := gitlog.GetCommitsFn(since, until)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read commits: %v\n", err)
		return ""
	}

	label, err := parser.Parse(commits)
	if err != nil {
		fmt.Fprintf(os.Stderr, "commit parser failed: %v\n", err)
		return ""
	}

	return label
}

// tryInferBumpTypeFromChangelogParserPlugin tries to infer bump type from CHANGELOG.md.
func tryInferBumpTypeFromChangelogParserPlugin() string {
	parser := changelogparser.GetChangelogParserFn()
	if parser == nil {
		return ""
	}

	// Check if changelog parser is enabled
	plugin, ok := parser.(*changelogparser.ChangelogParserPlugin)
	if !ok || !plugin.IsEnabled() {
		return ""
	}

	// Only use changelog parser if it should take precedence
	if !plugin.ShouldTakePrecedence() {
		return ""
	}

	label, err := parser.InferBumpType()
	if err != nil {
		// Don't print error if changelog is not found or empty - fall back to commits
		return ""
	}

	fmt.Fprintf(os.Stderr, "Inferred from changelog: %s\n", label)
	return label
}
