package tagmanager

import (
	"context"

	"github.com/indaco/sley/internal/core"
)

// MockGitCommitOperations is a mock implementation of core.GitCommitOperations for testing.
type MockGitCommitOperations struct {
	StageFilesFn       func(ctx context.Context, files ...string) error
	CommitFn           func(ctx context.Context, message string) error
	GetModifiedFilesFn func(ctx context.Context) ([]string, error)
}

// Verify MockGitCommitOperations implements core.GitCommitOperations.
var _ core.GitCommitOperations = (*MockGitCommitOperations)(nil)

// StageFiles implements core.GitCommitOperations.
func (m *MockGitCommitOperations) StageFiles(ctx context.Context, files ...string) error {
	if m.StageFilesFn != nil {
		return m.StageFilesFn(ctx, files...)
	}
	return nil
}

// Commit implements core.GitCommitOperations.
func (m *MockGitCommitOperations) Commit(ctx context.Context, message string) error {
	if m.CommitFn != nil {
		return m.CommitFn(ctx, message)
	}
	return nil
}

// GetModifiedFiles implements core.GitCommitOperations.
func (m *MockGitCommitOperations) GetModifiedFiles(ctx context.Context) ([]string, error) {
	if m.GetModifiedFilesFn != nil {
		return m.GetModifiedFilesFn(ctx)
	}
	return []string{}, nil
}
