package tagmanager

import "github.com/indaco/sley/internal/core"

// MockGitCommitOperations is a mock implementation of core.GitCommitOperations for testing.
type MockGitCommitOperations struct {
	StageFilesFn       func(files ...string) error
	CommitFn           func(message string) error
	GetModifiedFilesFn func() ([]string, error)
}

// Verify MockGitCommitOperations implements core.GitCommitOperations.
var _ core.GitCommitOperations = (*MockGitCommitOperations)(nil)

// StageFiles implements core.GitCommitOperations.
func (m *MockGitCommitOperations) StageFiles(files ...string) error {
	if m.StageFilesFn != nil {
		return m.StageFilesFn(files...)
	}
	return nil
}

// Commit implements core.GitCommitOperations.
func (m *MockGitCommitOperations) Commit(message string) error {
	if m.CommitFn != nil {
		return m.CommitFn(message)
	}
	return nil
}

// GetModifiedFiles implements core.GitCommitOperations.
func (m *MockGitCommitOperations) GetModifiedFiles() ([]string, error) {
	if m.GetModifiedFilesFn != nil {
		return m.GetModifiedFilesFn()
	}
	return []string{}, nil
}
