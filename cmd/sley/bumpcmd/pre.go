package bumpcmd

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/hooks"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/semver"
	"github.com/urfave/cli/v3"
)

// preCmd returns the "pre" subcommand for incrementing pre-release versions.
func preCmd(cfg *config.Config, registry *plugins.PluginRegistry) *cli.Command {
	return &cli.Command{
		Name:      "pre",
		Usage:     "Increment pre-release version (e.g., rc.1 -> rc.2)",
		UsageText: "sley bump pre [--label name] [--meta data] [--preserve-meta] [--skip-hooks]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "label",
				Aliases: []string{"l"},
				Usage:   "Pre-release label (e.g., alpha, beta, rc). If omitted, increments existing pre-release",
			},
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpPre(ctx, cmd, cfg, registry)
		},
	}
}

// runBumpPre executes the pre-release bump logic.
func runBumpPre(ctx context.Context, cmd *cli.Command, cfg *config.Config, registry *plugins.PluginRegistry) error {
	label := cmd.String("label")
	meta := cmd.String("meta")
	isPreserveMeta := cmd.Bool("preserve-meta")
	isSkipHooks := cmd.Bool("skip-hooks")

	if err := hooks.RunPreReleaseHooksFn(isSkipHooks); err != nil {
		return err
	}

	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	if !execCtx.IsSingleModule() {
		return fmt.Errorf("pre-release bump not yet supported for multi-module mode")
	}

	return runSingleModulePreBump(ctx, cmd, cfg, registry, execCtx, label, meta, isPreserveMeta, isSkipHooks)
}

// runSingleModulePreBump handles pre-release bump for single-module mode.
func runSingleModulePreBump(ctx context.Context, cmd *cli.Command, cfg *config.Config, registry *plugins.PluginRegistry, execCtx *clix.ExecutionContext, label, meta string, isPreserveMeta, isSkipHooks bool) error {
	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	previousVersion, err := semver.ReadVersion(execCtx.Path)
	if err != nil {
		return err
	}

	// Calculate new version
	newVersion := previousVersion
	if label != "" {
		newVersion.PreRelease = semver.IncrementPreRelease(previousVersion.PreRelease, label)
	} else {
		if previousVersion.PreRelease == "" {
			return fmt.Errorf("current version has no pre-release; use --label to specify one")
		}
		base := extractPreReleaseBase(previousVersion.PreRelease)
		newVersion.PreRelease = semver.IncrementPreRelease(previousVersion.PreRelease, base)
	}
	newVersion.Build = calculateNewBuild(meta, isPreserveMeta, previousVersion.Build)

	// Execute all pre-bump validations
	if err := executePreBumpValidations(registry, newVersion, previousVersion, "pre"); err != nil {
		return err
	}

	if err := runPreBumpExtensionHooks(ctx, cfg, execCtx.Path, newVersion.String(), previousVersion.String(), "pre", isSkipHooks); err != nil {
		return err
	}

	if err := semver.UpdatePreRelease(execCtx.Path, label, meta, isPreserveMeta); err != nil {
		return err
	}

	// Execute all post-bump actions
	if err := executePostBumpActions(registry, newVersion, previousVersion, "pre"); err != nil {
		return err
	}

	if err := runPostBumpExtensionHooks(ctx, cfg, execCtx.Path, previousVersion.String(), "pre", isSkipHooks); err != nil {
		return err
	}

	// Create tag after successful bump
	return createTagAfterBump(registry, newVersion, "pre")
}

// extractPreReleaseBase extracts the base label from a pre-release string.
// e.g., "rc.1" -> "rc", "beta.2" -> "beta", "alpha" -> "alpha", "rc1" -> "rc"
func extractPreReleaseBase(pre string) string {
	// First, check for dot followed by a number
	for i := len(pre) - 1; i >= 0; i-- {
		if pre[i] == '.' {
			// Check if everything after the dot is numeric
			suffix := pre[i+1:]
			isNumeric := true
			for _, c := range suffix {
				if c < '0' || c > '9' {
					isNumeric = false
					break
				}
			}
			if isNumeric && len(suffix) > 0 {
				return pre[:i]
			}
		}
	}

	// Check for trailing digits without dot (e.g., "rc1" -> "rc")
	lastNonDigit := -1
	for i := len(pre) - 1; i >= 0; i-- {
		if pre[i] < '0' || pre[i] > '9' {
			lastNonDigit = i
			break
		}
	}
	if lastNonDigit >= 0 && lastNonDigit < len(pre)-1 {
		return pre[:lastNonDigit+1]
	}

	return pre
}
