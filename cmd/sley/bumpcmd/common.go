package bumpcmd

import (
	"context"

	"github.com/indaco/sley/internal/clix"
	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/operations"
	"github.com/indaco/sley/internal/plugins"
	"github.com/indaco/sley/internal/semver"
	"github.com/urfave/cli/v3"
)

// versionCalculator is a function that calculates the new version from the previous version.
// It receives the previous version and returns the new version.
type versionCalculator func(prev semver.SemVersion) semver.SemVersion

// bumpParams holds all parameters needed for a bump operation.
type bumpParams struct {
	pre          string
	meta         string
	preserveMeta bool
	skipHooks    bool
	bumpType     string
	versionCalc  versionCalculator
}

// executeSingleModuleBump is the unified execution pipeline for single-module bump operations.
// It handles all common logic: validation, hooks, update, and post-actions.
func executeSingleModuleBump(
	ctx context.Context,
	cmd *cli.Command,
	cfg *config.Config,
	registry *plugins.PluginRegistry,
	execCtx *clix.ExecutionContext,
	params bumpParams,
) error {
	// Validate command context
	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	// Read current version
	previousVersion, err := semver.ReadVersion(execCtx.Path)
	if err != nil {
		return err
	}

	// Calculate new version using the provided calculator
	newVersion := params.versionCalc(previousVersion)

	// Apply pre-release and build metadata
	newVersion.Build = calculateNewBuild(params.meta, params.preserveMeta, previousVersion.Build)

	// Run pre-bump extension hooks first - extensions may set up state that plugins need to validate
	if err := runPreBumpExtensionHooks(ctx, cfg, execCtx.Path, newVersion.String(), previousVersion.String(), params.bumpType, params.skipHooks); err != nil {
		return err
	}

	// Re-read the current version in case an extension modified the .version file
	// (e.g., an extension that fetches the latest release from GitHub and updates .version)
	if !params.skipHooks {
		previousVersion, err = semver.ReadVersion(execCtx.Path)
		if err != nil {
			return err
		}
		// Recalculate the new version based on the potentially updated previous version
		newVersion = params.versionCalc(previousVersion)
		newVersion.Build = calculateNewBuild(params.meta, params.preserveMeta, previousVersion.Build)
	}

	// Execute all pre-bump validations after extensions have run
	if err := executePreBumpValidations(registry, newVersion, previousVersion, params.bumpType); err != nil {
		return err
	}

	// Update the version file
	if err := semver.UpdateVersion(execCtx.Path, params.bumpType, params.pre, params.meta, params.preserveMeta); err != nil {
		return err
	}

	// Execute all post-bump actions
	if err := executePostBumpActions(registry, newVersion, previousVersion, params.bumpType, execCtx.Path); err != nil {
		return err
	}

	// Run post-bump extension hooks
	if err := runPostBumpExtensionHooks(ctx, cfg, execCtx.Path, previousVersion.String(), params.bumpType, params.skipHooks); err != nil {
		return err
	}

	// Create tag after successful bump
	return createTagAfterBump(registry, newVersion, params.bumpType)
}

// executePreBumpValidations runs all validation checks before performing a bump.
// Returns error if any validation fails.
func executePreBumpValidations(registry *plugins.PluginRegistry, newVersion, previousVersion semver.SemVersion, bumpType string) error {
	// Validate release gates before bumping
	if err := validateReleaseGate(registry, newVersion, previousVersion, bumpType); err != nil {
		return err
	}

	// Validate version policy before bumping
	if err := validateVersionPolicy(registry, newVersion, previousVersion, bumpType); err != nil {
		return err
	}

	// Validate dependency consistency before bumping
	if err := validateDependencyConsistency(registry, newVersion); err != nil {
		return err
	}

	// Validate tag availability before bumping
	return validateTagAvailable(registry, newVersion)
}

// executePostBumpActions runs all post-bump operations like syncing dependencies,
// generating changelog, and recording audit logs.
// The bumpedPath parameter is the path to the .version file that was just bumped,
// used to avoid showing it twice in the dependency sync output.
func executePostBumpActions(registry *plugins.PluginRegistry, newVersion, previousVersion semver.SemVersion, bumpType, bumpedPath string) error {
	// Sync dependency files after updating .version
	if err := syncDependencies(registry, newVersion, bumpedPath); err != nil {
		return err
	}

	// Generate changelog entry
	if err := generateChangelogAfterBump(registry, newVersion, previousVersion, bumpType); err != nil {
		return err
	}

	// Record audit log entry
	return recordAuditLogEntry(registry, newVersion, previousVersion, bumpType)
}

// makePatchCalculator returns a version calculator for patch bumps.
func makePatchCalculator(pre, meta string, preserveMeta bool) versionCalculator {
	return func(prev semver.SemVersion) semver.SemVersion {
		next := prev
		next.Patch++
		next.PreRelease = pre
		next.Build = calculateNewBuild(meta, preserveMeta, prev.Build)
		return next
	}
}

// makeMinorCalculator returns a version calculator for minor bumps.
func makeMinorCalculator(pre, meta string, preserveMeta bool) versionCalculator {
	return func(prev semver.SemVersion) semver.SemVersion {
		next := prev
		next.Minor++
		next.Patch = 0
		next.PreRelease = pre
		next.Build = calculateNewBuild(meta, preserveMeta, prev.Build)
		return next
	}
}

// makeMajorCalculator returns a version calculator for major bumps.
func makeMajorCalculator(pre, meta string, preserveMeta bool) versionCalculator {
	return func(prev semver.SemVersion) semver.SemVersion {
		next := prev
		next.Major++
		next.Minor = 0
		next.Patch = 0
		next.PreRelease = pre
		next.Build = calculateNewBuild(meta, preserveMeta, prev.Build)
		return next
	}
}

// extractBumpParams extracts common bump parameters from CLI command.
func extractBumpParams(cmd *cli.Command, bumpType string) bumpParams {
	return bumpParams{
		pre:          cmd.String("pre"),
		meta:         cmd.String("meta"),
		preserveMeta: cmd.Bool("preserve-meta"),
		skipHooks:    cmd.Bool("skip-hooks"),
		bumpType:     bumpType,
	}
}

// executeStandardBump handles the standard bump workflow for patch/minor/major commands.
// It manages the multi-module vs single-module branching and delegates to the appropriate handler.
func executeStandardBump(
	ctx context.Context,
	cmd *cli.Command,
	cfg *config.Config,
	registry *plugins.PluginRegistry,
	params bumpParams,
	multiModuleOp operations.BumpType,
) error {
	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	if !execCtx.IsSingleModule() {
		// Multi-module mode delegates to runMultiModuleBump
		return runMultiModuleBump(ctx, cmd, execCtx, registry, multiModuleOp, params.pre, params.meta, params.preserveMeta)
	}

	// Single-module mode uses the unified executor
	return executeSingleModuleBump(ctx, cmd, cfg, registry, execCtx, params)
}
