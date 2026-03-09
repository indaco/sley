package tagmanager

import (
	"context"

	"github.com/indaco/sley/internal/core"
)

// MockGitTagOperations is a mock implementation of core.GitTagOperations for testing.
type MockGitTagOperations struct {
	CreateAnnotatedTagFn   func(ctx context.Context, name, message string) error
	CreateLightweightTagFn func(ctx context.Context, name string) error
	CreateSignedTagFn      func(ctx context.Context, name, message, keyID string) error
	TagExistsFn            func(ctx context.Context, name string) (bool, error)
	GetLatestTagFn         func(ctx context.Context) (string, error)
	PushTagFn              func(ctx context.Context, name string) error
	ListTagsFn             func(ctx context.Context, pattern string) ([]string, error)
	DeleteTagFn            func(ctx context.Context, name string) error
	DeleteRemoteTagFn      func(ctx context.Context, name string) error
}

// Verify MockGitTagOperations implements core.GitTagOperations.
var _ core.GitTagOperations = (*MockGitTagOperations)(nil)

// CreateAnnotatedTag implements core.GitTagOperations.
func (m *MockGitTagOperations) CreateAnnotatedTag(ctx context.Context, name, message string) error {
	if m.CreateAnnotatedTagFn != nil {
		return m.CreateAnnotatedTagFn(ctx, name, message)
	}
	return nil
}

// CreateLightweightTag implements core.GitTagOperations.
func (m *MockGitTagOperations) CreateLightweightTag(ctx context.Context, name string) error {
	if m.CreateLightweightTagFn != nil {
		return m.CreateLightweightTagFn(ctx, name)
	}
	return nil
}

// CreateSignedTag implements core.GitTagOperations.
func (m *MockGitTagOperations) CreateSignedTag(ctx context.Context, name, message, keyID string) error {
	if m.CreateSignedTagFn != nil {
		return m.CreateSignedTagFn(ctx, name, message, keyID)
	}
	return nil
}

// TagExists implements core.GitTagOperations.
func (m *MockGitTagOperations) TagExists(ctx context.Context, name string) (bool, error) {
	if m.TagExistsFn != nil {
		return m.TagExistsFn(ctx, name)
	}
	return false, nil
}

// GetLatestTag implements core.GitTagOperations.
func (m *MockGitTagOperations) GetLatestTag(ctx context.Context) (string, error) {
	if m.GetLatestTagFn != nil {
		return m.GetLatestTagFn(ctx)
	}
	return "", nil
}

// PushTag implements core.GitTagOperations.
func (m *MockGitTagOperations) PushTag(ctx context.Context, name string) error {
	if m.PushTagFn != nil {
		return m.PushTagFn(ctx, name)
	}
	return nil
}

// ListTags implements core.GitTagOperations.
func (m *MockGitTagOperations) ListTags(ctx context.Context, pattern string) ([]string, error) {
	if m.ListTagsFn != nil {
		return m.ListTagsFn(ctx, pattern)
	}
	return []string{}, nil
}

// DeleteTag implements core.GitTagOperations.
func (m *MockGitTagOperations) DeleteTag(ctx context.Context, name string) error {
	if m.DeleteTagFn != nil {
		return m.DeleteTagFn(ctx, name)
	}
	return nil
}

// DeleteRemoteTag implements core.GitTagOperations.
func (m *MockGitTagOperations) DeleteRemoteTag(ctx context.Context, name string) error {
	if m.DeleteRemoteTagFn != nil {
		return m.DeleteRemoteTagFn(ctx, name)
	}
	return nil
}
