package operations

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/workspace"
)

// PreOperation sets or increments pre-release labels on a module.
type PreOperation struct {
	fs        core.FileSystem
	label     string
	increment bool
}

// NewPreOperation creates a new pre-release operation.
func NewPreOperation(fs core.FileSystem, label string, increment bool) *PreOperation {
	return &PreOperation{
		fs:        fs,
		label:     label,
		increment: increment,
	}
}

// Execute performs the pre-release operation on the module.
func (op *PreOperation) Execute(ctx context.Context, mod *workspace.Module) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create version manager
	vm := semver.NewVersionManager(op.fs, nil)

	// Read current version
	currentVer, err := vm.Read(mod.Path)
	if err != nil {
		return fmt.Errorf("failed to read version from %s: %w", mod.Path, err)
	}

	// Store old version for display
	oldVersion := currentVer.String()

	// Create new version
	newVer := currentVer

	if op.increment {
		newVer.PreRelease = semver.IncrementPreRelease(currentVer.PreRelease, op.label)
	} else {
		// If there's no existing pre-release, bump patch first
		if currentVer.PreRelease == "" {
			newVer.Patch++
		}
		newVer.PreRelease = op.label
	}

	// Write the new version
	if err := vm.Save(mod.Path, newVer); err != nil {
		return fmt.Errorf("failed to write version to %s: %w", mod.Path, err)
	}

	// Update module's current version for display (set both old and new for output)
	mod.CurrentVersion = newVer.String()
	// Store old version in a custom field if workspace module supports it
	// For now, we rely on the workspace executor to capture the old version before Execute

	_ = oldVersion // For potential logging

	return nil
}

// Name returns the name of this operation.
func (op *PreOperation) Name() string {
	if op.increment {
		return "increment pre-release"
	}
	return "set pre-release"
}
