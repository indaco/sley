package releasegate

// MockGitOperations is a mock implementation of GitOperations for testing.
type MockGitOperations struct {
	IsWorktreeCleanFn  func() (bool, error)
	GetCurrentBranchFn func() (string, error)
	GetRecentCommitsFn func(count int) ([]string, error)
}

// Verify MockGitOperations implements GitOperations.
var _ GitOperations = (*MockGitOperations)(nil)

// IsWorktreeClean implements GitOperations.
func (m *MockGitOperations) IsWorktreeClean() (bool, error) {
	if m.IsWorktreeCleanFn != nil {
		return m.IsWorktreeCleanFn()
	}
	return true, nil
}

// GetCurrentBranch implements GitOperations.
func (m *MockGitOperations) GetCurrentBranch() (string, error) {
	if m.GetCurrentBranchFn != nil {
		return m.GetCurrentBranchFn()
	}
	return "main", nil
}

// GetRecentCommits implements GitOperations.
func (m *MockGitOperations) GetRecentCommits(count int) ([]string, error) {
	if m.GetRecentCommitsFn != nil {
		return m.GetRecentCommitsFn(count)
	}
	return []string{}, nil
}
