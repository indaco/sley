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
)

// BumpOperation performs a version bump on a module.
type BumpOperation struct {
	fs               core.FileSystem
	bumpType         BumpType
	preRelease       string
	metadata         string
	preserveMetadata bool
}

// NewBumpOperation creates a new bump operation.
func NewBumpOperation(fs core.FileSystem, bumpType BumpType, preRelease, metadata string, preserveMetadata bool) *BumpOperation {
	return &BumpOperation{
		fs:               fs,
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

	// Perform the bump based on type
	var newVer semver.SemVersion
	switch op.bumpType {
	case BumpPatch:
		newVer = semver.SemVersion{
			Major: currentVer.Major,
			Minor: currentVer.Minor,
			Patch: currentVer.Patch + 1,
		}
	case BumpMinor:
		newVer = semver.SemVersion{
			Major: currentVer.Major,
			Minor: currentVer.Minor + 1,
			Patch: 0,
		}
	case BumpMajor:
		newVer = semver.SemVersion{
			Major: currentVer.Major + 1,
			Minor: 0,
			Patch: 0,
		}
	case BumpRelease:
		// Release removes pre-release and build metadata
		newVer = semver.SemVersion{
			Major: currentVer.Major,
			Minor: currentVer.Minor,
			Patch: currentVer.Patch,
		}
	case BumpAuto:
		// Auto bump uses heuristic-based logic
		autoVer, autoErr := semver.BumpNextFunc(currentVer)
		if autoErr != nil {
			return fmt.Errorf("auto bump failed: %w", autoErr)
		}
		newVer = autoVer
	default:
		return fmt.Errorf("unknown bump type: %s", op.bumpType)
	}

	// Apply pre-release label if provided
	if op.preRelease != "" {
		newVer.PreRelease = op.preRelease
	}

	// Apply metadata
	if op.metadata != "" {
		newVer.Build = op.metadata
	} else if op.preserveMetadata && currentVer.Build != "" {
		newVer.Build = currentVer.Build
	}

	// Write the new version
	if err := vm.Save(ctx, mod.Path, newVer); err != nil {
		return fmt.Errorf("failed to write version to %s: %w", mod.Path, err)
	}

	// Update module's current version for display
	mod.CurrentVersion = newVer.String()

	return nil
}

// Name returns the name of this operation.
func (op *BumpOperation) Name() string {
	return fmt.Sprintf("bump %s", op.bumpType)
}
