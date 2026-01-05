package operations

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/workspace"
)

// ValidateOperation validates the version file of a module.
type ValidateOperation struct {
	fs core.FileSystem
}

// NewValidateOperation creates a new validate operation.
func NewValidateOperation(fs core.FileSystem) *ValidateOperation {
	return &ValidateOperation{
		fs: fs,
	}
}

// Execute validates the version file in the module.
// The current version is stored in the module's CurrentVersion field on success.
func (op *ValidateOperation) Execute(ctx context.Context, mod *workspace.Module) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create version manager
	vm := semver.NewVersionManager(op.fs, nil)

	// Read and validate version
	ver, err := vm.Read(mod.Path)
	if err != nil {
		return fmt.Errorf("invalid version file at %s: %w", mod.Path, err)
	}

	// Update module's current version for display
	mod.CurrentVersion = ver.String()

	return nil
}

// Name returns the name of this operation.
func (op *ValidateOperation) Name() string {
	return "validate"
}
