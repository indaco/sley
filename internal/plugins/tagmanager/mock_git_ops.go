package tagmanager

import "github.com/indaco/sley/internal/core"

// MockGitTagOperations is a mock implementation of core.GitTagOperations for testing.
type MockGitTagOperations struct {
	CreateAnnotatedTagFn   func(name, message string) error
	CreateLightweightTagFn func(name string) error
	CreateSignedTagFn      func(name, message, keyID string) error
	TagExistsFn            func(name string) (bool, error)
	GetLatestTagFn         func() (string, error)
	PushTagFn              func(name string) error
	ListTagsFn             func(pattern string) ([]string, error)
	DeleteTagFn            func(name string) error
	DeleteRemoteTagFn      func(name string) error
}

// Verify MockGitTagOperations implements core.GitTagOperations.
var _ core.GitTagOperations = (*MockGitTagOperations)(nil)

// CreateAnnotatedTag implements core.GitTagOperations.
func (m *MockGitTagOperations) CreateAnnotatedTag(name, message string) error {
	if m.CreateAnnotatedTagFn != nil {
		return m.CreateAnnotatedTagFn(name, message)
	}
	return nil
}

// CreateLightweightTag implements core.GitTagOperations.
func (m *MockGitTagOperations) CreateLightweightTag(name string) error {
	if m.CreateLightweightTagFn != nil {
		return m.CreateLightweightTagFn(name)
	}
	return nil
}

// CreateSignedTag implements core.GitTagOperations.
func (m *MockGitTagOperations) CreateSignedTag(name, message, keyID string) error {
	if m.CreateSignedTagFn != nil {
		return m.CreateSignedTagFn(name, message, keyID)
	}
	return nil
}

// TagExists implements core.GitTagOperations.
func (m *MockGitTagOperations) TagExists(name string) (bool, error) {
	if m.TagExistsFn != nil {
		return m.TagExistsFn(name)
	}
	return false, nil
}

// GetLatestTag implements core.GitTagOperations.
func (m *MockGitTagOperations) GetLatestTag() (string, error) {
	if m.GetLatestTagFn != nil {
		return m.GetLatestTagFn()
	}
	return "", nil
}

// PushTag implements core.GitTagOperations.
func (m *MockGitTagOperations) PushTag(name string) error {
	if m.PushTagFn != nil {
		return m.PushTagFn(name)
	}
	return nil
}

// ListTags implements core.GitTagOperations.
func (m *MockGitTagOperations) ListTags(pattern string) ([]string, error) {
	if m.ListTagsFn != nil {
		return m.ListTagsFn(pattern)
	}
	return []string{}, nil
}

// DeleteTag implements core.GitTagOperations.
func (m *MockGitTagOperations) DeleteTag(name string) error {
	if m.DeleteTagFn != nil {
		return m.DeleteTagFn(name)
	}
	return nil
}

// DeleteRemoteTag implements core.GitTagOperations.
func (m *MockGitTagOperations) DeleteRemoteTag(name string) error {
	if m.DeleteRemoteTagFn != nil {
		return m.DeleteRemoteTagFn(name)
	}
	return nil
}
