package clix

import (
	"testing"

	"github.com/indaco/sley/internal/workspace"
)

func TestExecutionMode_String(t *testing.T) {
	tests := []struct {
		name string
		mode ExecutionMode
		want string
	}{
		{
			name: "single module mode",
			mode: SingleModuleMode,
			want: "SingleModule",
		},
		{
			name: "multi module mode",
			mode: MultiModuleMode,
			want: "MultiModule",
		},
		{
			name: "unknown mode",
			mode: ExecutionMode(99),
			want: "Unknown(99)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("ExecutionMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutionContext_IsSingleModule(t *testing.T) {
	tests := []struct {
		name string
		ctx  *ExecutionContext
		want bool
	}{
		{
			name: "single module mode",
			ctx: &ExecutionContext{
				Mode: SingleModuleMode,
			},
			want: true,
		},
		{
			name: "multi module mode",
			ctx: &ExecutionContext{
				Mode: MultiModuleMode,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ctx.IsSingleModule()
			if got != tt.want {
				t.Errorf("ExecutionContext.IsSingleModule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutionContext_IsMultiModule(t *testing.T) {
	tests := []struct {
		name string
		ctx  *ExecutionContext
		want bool
	}{
		{
			name: "single module mode",
			ctx: &ExecutionContext{
				Mode: SingleModuleMode,
			},
			want: false,
		},
		{
			name: "multi module mode",
			ctx: &ExecutionContext{
				Mode: MultiModuleMode,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ctx.IsMultiModule()
			if got != tt.want {
				t.Errorf("ExecutionContext.IsMultiModule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutionContext_EmptyModules(t *testing.T) {
	execCtx := &ExecutionContext{
		Mode:    MultiModuleMode,
		Modules: []*workspace.Module{},
	}

	if !execCtx.IsMultiModule() {
		t.Error("should be multi-module mode")
	}

	if len(execCtx.Modules) != 0 {
		t.Errorf("expected 0 modules, got %d", len(execCtx.Modules))
	}
}

func TestExecutionContext_SingleModuleWithModules(t *testing.T) {
	// This is an edge case - single module mode should not have Modules set
	execCtx := &ExecutionContext{
		Mode:    SingleModuleMode,
		Path:    "/test/.version",
		Modules: []*workspace.Module{{Name: "test"}},
	}

	if !execCtx.IsSingleModule() {
		t.Error("should be single-module mode")
	}

	if execCtx.Path != "/test/.version" {
		t.Errorf("Path = %q, want %q", execCtx.Path, "/test/.version")
	}

	// Even though Modules is set, we're in single-module mode
	// This verifies that Mode takes precedence
}

func TestExecutionContext_MultiModuleWithPath(t *testing.T) {
	// This is an edge case - multi-module mode should not have Path set
	execCtx := &ExecutionContext{
		Mode: MultiModuleMode,
		Path: "/test/.version",
		Modules: []*workspace.Module{
			{Name: "module-a"},
			{Name: "module-b"},
		},
	}

	if !execCtx.IsMultiModule() {
		t.Error("should be multi-module mode")
	}

	if len(execCtx.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(execCtx.Modules))
	}

	// Even though Path is set, we're in multi-module mode
	// This verifies that Mode takes precedence
}

func TestWithDefaultAll(t *testing.T) {
	opts := &executionOptions{}

	// Apply the option
	option := WithDefaultAll()
	option(opts)

	if !opts.defaultToAll {
		t.Error("expected defaultToAll to be true")
	}
}
