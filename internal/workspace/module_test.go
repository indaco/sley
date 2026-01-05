package workspace

import (
	"testing"
)

func TestModule_DisplayName(t *testing.T) {
	tests := []struct {
		name     string
		module   *Module
		expected string
	}{
		{
			name: "with version",
			module: &Module{
				Name:           "module-a",
				CurrentVersion: "1.0.0",
			},
			expected: "module-a (1.0.0)",
		},
		{
			name: "without version",
			module: &Module{
				Name:           "module-b",
				CurrentVersion: "",
			},
			expected: "module-b",
		},
		{
			name: "with complex version",
			module: &Module{
				Name:           "api",
				CurrentVersion: "2.5.1-beta.1+build.123",
			},
			expected: "api (2.5.1-beta.1+build.123)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.module.DisplayName()
			if got != tt.expected {
				t.Errorf("DisplayName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestModule_String(t *testing.T) {
	module := &Module{
		Name:           "test-module",
		CurrentVersion: "3.2.1",
	}

	// String() should call DisplayName()
	if module.String() != module.DisplayName() {
		t.Errorf("String() = %q, DisplayName() = %q, should be equal",
			module.String(), module.DisplayName())
	}
}

func TestModule_DisplayNameWithPath(t *testing.T) {
	tests := []struct {
		name     string
		module   *Module
		expected string
	}{
		{
			name: "with version and path",
			module: &Module{
				Name:           "version",
				CurrentVersion: "1.0.0",
				Dir:            "backend/gateway/internal/version",
			},
			expected: "version (1.0.0) - backend/gateway/internal/version",
		},
		{
			name: "without version but with path",
			module: &Module{
				Name:           "version",
				CurrentVersion: "",
				Dir:            "cli/internal/version",
			},
			expected: "version - cli/internal/version",
		},
		{
			name: "with version but root dir",
			module: &Module{
				Name:           "myapp",
				CurrentVersion: "2.0.0",
				Dir:            ".",
			},
			expected: "myapp (2.0.0)",
		},
		{
			name: "with version but empty dir",
			module: &Module{
				Name:           "myapp",
				CurrentVersion: "2.0.0",
				Dir:            "",
			},
			expected: "myapp (2.0.0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.module.DisplayNameWithPath()
			if got != tt.expected {
				t.Errorf("DisplayNameWithPath() = %q, want %q", got, tt.expected)
			}
		})
	}
}
