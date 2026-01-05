package operations

import (
	"context"
	"testing"
	"time"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/workspace"
)

func TestNewValidateOperation(t *testing.T) {
	fs := core.NewMockFileSystem()
	op := NewValidateOperation(fs)

	if op == nil {
		t.Fatal("NewValidateOperation returned nil")
	}
	if op.fs == nil {
		t.Error("fs is nil")
	}
}

func TestValidateOperation_Execute(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "simple version",
			version:  "1.2.3\n",
			expected: "1.2.3",
		},
		{
			name:     "version with pre-release",
			version:  "2.0.0-alpha.1\n",
			expected: "2.0.0-alpha.1",
		},
		{
			name:     "version with metadata",
			version:  "3.1.0+build.123\n",
			expected: "3.1.0+build.123",
		},
		{
			name:     "version with pre-release and metadata",
			version:  "4.0.0-beta.2+build.456\n",
			expected: "4.0.0-beta.2+build.456",
		},
		{
			name:     "version with extra whitespace",
			version:  "  1.0.0  \n",
			expected: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test/.version", []byte(tt.version))

			op := NewValidateOperation(fs)
			mod := &workspace.Module{
				Name: "test",
				Path: "/test/.version",
			}

			ctx := context.Background()
			err := op.Execute(ctx, mod)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			if mod.CurrentVersion != tt.expected {
				t.Errorf("module CurrentVersion = %q, want %q", mod.CurrentVersion, tt.expected)
			}
		})
	}
}

func TestValidateOperation_Execute_FileNotFound(t *testing.T) {
	fs := core.NewMockFileSystem()
	// Don't create any file

	op := NewValidateOperation(fs)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestValidateOperation_Execute_InvalidVersion(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("not-a-version\n"))

	op := NewValidateOperation(fs)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected error for invalid version, got nil")
	}
}

func TestValidateOperation_Execute_ContextCancellation(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.0.0\n"))

	op := NewValidateOperation(fs)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestValidateOperation_Execute_ContextTimeout(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.0.0\n"))

	op := NewValidateOperation(fs)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond)

	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected context timeout error, got nil")
	}
}

func TestValidateOperation_Execute_ReadError(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.ReadErr = context.DeadlineExceeded

	op := NewValidateOperation(fs)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected read error, got nil")
	}
}

func TestValidateOperation_Name(t *testing.T) {
	fs := core.NewMockFileSystem()
	op := NewValidateOperation(fs)

	name := op.Name()
	expected := "validate"
	if name != expected {
		t.Errorf("Name() = %q, want %q", name, expected)
	}
}

func TestValidateOperation_Execute_EmptyFile(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte(""))

	op := NewValidateOperation(fs)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
}

func TestValidateOperation_Execute_PreservesModulePath(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewValidateOperation(fs)
	mod := &workspace.Module{
		Name:    "test-module",
		Path:    "/test/.version",
		RelPath: "test/.version",
		Dir:     "/test",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify that module fields are preserved except CurrentVersion
	if mod.Name != "test-module" {
		t.Errorf("module Name changed, got %q", mod.Name)
	}
	if mod.Path != "/test/.version" {
		t.Errorf("module Path changed, got %q", mod.Path)
	}
	if mod.RelPath != "test/.version" {
		t.Errorf("module RelPath changed, got %q", mod.RelPath)
	}
	if mod.Dir != "/test" {
		t.Errorf("module Dir changed, got %q", mod.Dir)
	}
	if mod.CurrentVersion != "1.2.3" {
		t.Errorf("module CurrentVersion = %q, want %q", mod.CurrentVersion, "1.2.3")
	}
}

func TestValidateOperation_Execute_InvalidBuildMetadata(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.0.0+inv@lid-meta\n"))

	op := NewValidateOperation(fs)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected error for invalid build metadata, got nil")
	}
}

func TestValidateOperation_Execute_InvalidPrerelease(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.0.0-inv@lid\n"))

	op := NewValidateOperation(fs)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected error for invalid prerelease, got nil")
	}
}
