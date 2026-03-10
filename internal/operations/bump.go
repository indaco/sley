// Package operations provides reusable operations for module manipulation.
package operations

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/workspace"
)

// BumpType represents the type of version bump to perform.
type BumpType string

const (
	BumpPatch   BumpType = "patch"
	BumpMinor   BumpType = "minor"
	BumpMajor   BumpType = "major"
	BumpRelease BumpType = "release"
	BumpAuto    BumpType = "auto"
	BumpPre     BumpType = "pre"
)

// BumpOperation performs a version bump on a module.
type BumpOperation struct {
	fs               core.FileSystem
	bumper           semver.VersionBumper
	bumpType         BumpType
	preRelease       string
	metadata         string
	preserveMetadata bool
}

// NewBumpOperation creates a new bump operation.
func NewBumpOperation(fs core.FileSystem, bumper semver.VersionBumper, bumpType BumpType, preRelease, metadata string, preserveMetadata bool) *BumpOperation {
	return &BumpOperation{
		fs:               fs,
		bumper:           bumper,
		bumpType:         bumpType,
		preRelease:       preRelease,
		metadata:         metadata,
		preserveMetadata: preserveMetadata,
	}
}

// Execute performs the bump operation on the module.
func (op *BumpOperation) Execute(ctx context.Context, mod *workspace.Module) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create version manager
	vm := semver.NewVersionManager(op.fs, nil)

	// Read current version
	currentVer, err := vm.Read(ctx, mod.Path)
	if err != nil {
		return fmt.Errorf("failed to read version from %s: %w", mod.Path, err)
	}

	// Calculate the new version
	newVer, err := op.calculateNewVersion(currentVer)
	if err != nil {
		return err
	}

	// BumpPre handles its own metadata, others use common logic
	if op.bumpType != BumpPre {
		op.applyPreReleaseAndMetadata(&newVer, currentVer)
	}

	// Write the new version
	if err := vm.Save(ctx, mod.Path, newVer); err != nil {
		return fmt.Errorf("failed to write version to %s: %w", mod.Path, err)
	}

	// Update module's current version for display
	mod.CurrentVersion = newVer.String()

	return nil
}

// calculateNewVersion computes the new version based on bump type.
func (op *BumpOperation) calculateNewVersion(currentVer semver.SemVersion) (semver.SemVersion, error) {
	switch op.bumpType {
	case BumpPatch:
		return op.bumpPatch(currentVer), nil
	case BumpMinor:
		return op.bumpMinor(currentVer), nil
	case BumpMajor:
		return op.bumpMajor(currentVer), nil
	case BumpRelease:
		return op.bumpRelease(currentVer), nil
	case BumpAuto:
		return op.bumpAuto(currentVer)
	case BumpPre:
		return op.bumpPre(currentVer)
	default:
		return semver.SemVersion{}, fmt.Errorf("unknown bump type: %s", op.bumpType)
	}
}

func (op *BumpOperation) bumpPatch(current semver.SemVersion) semver.SemVersion {
	return semver.SemVersion{
		Major: current.Major,
		Minor: current.Minor,
		Patch: current.Patch + 1,
	}
}

func (op *BumpOperation) bumpMinor(current semver.SemVersion) semver.SemVersion {
	return semver.SemVersion{
		Major: current.Major,
		Minor: current.Minor + 1,
		Patch: 0,
	}
}

func (op *BumpOperation) bumpMajor(current semver.SemVersion) semver.SemVersion {
	return semver.SemVersion{
		Major: current.Major + 1,
		Minor: 0,
		Patch: 0,
	}
}

func (op *BumpOperation) bumpRelease(current semver.SemVersion) semver.SemVersion {
	return semver.SemVersion{
		Major: current.Major,
		Minor: current.Minor,
		Patch: current.Patch,
	}
}

func (op *BumpOperation) bumpAuto(current semver.SemVersion) (semver.SemVersion, error) {
	newVer, err := op.bumper.BumpNext(current)
	if err != nil {
		return semver.SemVersion{}, fmt.Errorf("auto bump failed: %w", err)
	}
	return newVer, nil
}

func (op *BumpOperation) bumpPre(current semver.SemVersion) (semver.SemVersion, error) {
	newVer := semver.SemVersion{
		Major: current.Major,
		Minor: current.Minor,
		Patch: current.Patch,
	}

	// Determine the pre-release value
	preRelease, err := op.calculatePreRelease(current)
	if err != nil {
		return semver.SemVersion{}, err
	}
	newVer.PreRelease = preRelease

	// Apply metadata for pre-release bump
	op.applyMetadata(&newVer, current)

	return newVer, nil
}

// calculatePreRelease determines the pre-release string for BumpPre.
func (op *BumpOperation) calculatePreRelease(current semver.SemVersion) (string, error) {
	if op.preRelease != "" {
		return semver.IncrementPreRelease(current.PreRelease, op.preRelease), nil
	}
	if current.PreRelease != "" {
		base := semver.ExtractPreReleaseBase(current.PreRelease)
		return semver.IncrementPreRelease(current.PreRelease, base), nil
	}
	return "", fmt.Errorf("current version has no pre-release; use --label to specify one")
}

// applyPreReleaseAndMetadata applies pre-release and metadata to the version.
func (op *BumpOperation) applyPreReleaseAndMetadata(newVer *semver.SemVersion, currentVer semver.SemVersion) {
	if op.preRelease != "" {
		newVer.PreRelease = op.preRelease
	}
	op.applyMetadata(newVer, currentVer)
}

// applyMetadata applies build metadata to the version.
func (op *BumpOperation) applyMetadata(newVer *semver.SemVersion, currentVer semver.SemVersion) {
	if op.metadata != "" {
		newVer.Build = op.metadata
	} else if op.preserveMetadata && currentVer.Build != "" {
		newVer.Build = currentVer.Build
	}
}

// Name returns the name of this operation.
func (op *BumpOperation) Name() string {
	return fmt.Sprintf("bump %s", op.bumpType)
}

// BumpResult holds the result of a version bump calculation.
type BumpResult struct {
	PreviousVersion semver.SemVersion
	NewVersion      semver.SemVersion
}

// Preview calculates the new version without writing it.
// This allows callers to inspect the result for hooks and validation
// before committing the write via Write().
func (op *BumpOperation) Preview(ctx context.Context, path string) (BumpResult, error) {
	select {
	case <-ctx.Done():
		return BumpResult{}, ctx.Err()
	default:
	}

	vm := semver.NewVersionManager(op.fs, nil)
	currentVer, err := vm.Read(ctx, path)
	if err != nil {
		return BumpResult{}, fmt.Errorf("failed to read version from %s: %w", path, err)
	}

	newVer, err := op.calculateNewVersion(currentVer)
	if err != nil {
		return BumpResult{}, err
	}

	if op.bumpType != BumpPre {
		op.applyPreReleaseAndMetadata(&newVer, currentVer)
	}

	return BumpResult{
		PreviousVersion: currentVer,
		NewVersion:      newVer,
	}, nil
}

// Write saves the given version to the specified path.
// Used after Preview() when the caller has completed hooks and validation.
func (op *BumpOperation) Write(ctx context.Context, path string, version semver.SemVersion) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	vm := semver.NewVersionManager(op.fs, nil)
	return vm.Save(ctx, path, version)
}
