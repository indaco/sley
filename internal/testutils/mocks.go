package testutils

import (
	"context"
	"fmt"
)

// MockCommitParser implements the plugins.CommitParser interface for testing.
type MockCommitParser struct {
	Label string
	Err   error
}

func (m MockCommitParser) Name() string {
	return "mock"
}
func (m MockCommitParser) Parse(_ []string) (string, error) {
	return m.Label, m.Err
}

func WithMock(setup func(), testFunc func()) {
	setup()
	testFunc()
}

// MockHook implements PreReleaseHook for testing
type MockHook struct {
	Name      string
	ShouldErr bool
}

func (m MockHook) HookName() string {
	return m.Name
}

func (m MockHook) Run(ctx context.Context) error {
	if m.ShouldErr {
		return fmt.Errorf("%s failed", m.Name)
	}
	return nil
}
