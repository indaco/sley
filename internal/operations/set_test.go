package operations

import (
	"context"
	"testing"
	"time"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/workspace"
)

func TestNewSetOperation(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	version := "2.0.0"
	op := NewSetOperation(fs, version)

	if op == nil {
		t.Fatal("NewSetOperation returned nil")
	}
	if op.fs == nil {
		t.Error("fs is nil")
	}
	if op.version != version {
		t.Errorf("version = %v, want %v", op.version, version)
	}
}

func TestSetOperation_Execute(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		initial  string
		newVer   string
		expected string
	}{
		{
			name:     "set simple version",
			initial:  "1.0.0\n",
			newVer:   "2.0.0",
			expected: "2.0.0\n",
		},
		{
			name:     "set version with pre-release",
			initial:  "1.0.0\n",
			newVer:   "2.0.0-alpha.1",
			expected: "2.0.0-alpha.1\n",
		},
		{
			name:     "set version with metadata",
			initial:  "1.0.0\n",
			newVer:   "2.0.0+build.123",
			expected: "2.0.0+build.123\n",
		},
		{
			name:     "set version with pre-release and metadata",
			initial:  "1.0.0\n",
			newVer:   "2.0.0-beta.2+build.456",
			expected: "2.0.0-beta.2+build.456\n",
		},
		{
			name:     "downgrade version",
			initial:  "5.0.0\n",
			newVer:   "1.0.0",
			expected: "1.0.0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := core.NewMockFileSystem()
			fs.SetFile("/test/.version", []byte(tt.initial))

			op := NewSetOperation(fs, tt.newVer)
			mod := &workspace.Module{
				Name: "test",
				Path: "/test/.version",
			}

			ctx := context.Background()
			err := op.Execute(ctx, mod)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			data, ok := fs.GetFile("/test/.version")
			if !ok {
				t.Fatal("version file not found")
			}

			if string(data) != tt.expected {
				t.Errorf("version = %q, want %q", string(data), tt.expected)
			}

			// Verify module's CurrentVersion is updated
			expectedVersion := tt.newVer
			if mod.CurrentVersion != expectedVersion {
				t.Errorf("module CurrentVersion = %q, want %q", mod.CurrentVersion, expectedVersion)
			}
		})
	}
}

func TestSetOperation_Execute_InvalidVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "empty version",
			version: "",
		},
		{
			name:    "invalid format",
			version: "not-a-version",
		},
		{
			name:    "missing patch",
			version: "1.2",
		},
		{
			name:    "invalid characters",
			version: "1.2.x",
		},
		{
			name:    "negative numbers",
			version: "-1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := core.NewMockFileSystem()
			fs.SetFile("/test/.version", []byte("1.0.0\n"))

			op := NewSetOperation(fs, tt.version)
			mod := &workspace.Module{
				Name: "test",
				Path: "/test/.version",
			}

			ctx := context.Background()
			err := op.Execute(ctx, mod)
			if err == nil {
				t.Fatal("expected error for invalid version, got nil")
			}
		})
	}
}

func TestSetOperation_Execute_ContextCancellation(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.0.0\n"))

	op := NewSetOperation(fs, "2.0.0")
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

func TestSetOperation_Execute_ContextTimeout(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.0.0\n"))

	op := NewSetOperation(fs, "2.0.0")
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	<-ctx.Done() // Wait for the context to actually expire

	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected context timeout error, got nil")
	}
}

func TestSetOperation_Execute_WriteError(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.WriteErr = context.DeadlineExceeded

	op := NewSetOperation(fs, "2.0.0")
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected write error, got nil")
	}
}

func TestSetOperation_Name(t *testing.T) {
	t.Parallel()
	tests := []struct {
		version  string
		expected string
	}{
		{
			version:  "1.0.0",
			expected: "set 1.0.0",
		},
		{
			version:  "2.5.3-alpha.1",
			expected: "set 2.5.3-alpha.1",
		},
		{
			version:  "3.0.0+build.123",
			expected: "set 3.0.0+build.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()
			fs := core.NewMockFileSystem()
			op := NewSetOperation(fs, tt.version)

			name := op.Name()
			if name != tt.expected {
				t.Errorf("Name() = %q, want %q", name, tt.expected)
			}
		})
	}
}

func TestSetOperation_Execute_NoFileExists(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	// Don't create any file initially

	op := NewSetOperation(fs, "1.0.0")
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// File should be created
	data, ok := fs.GetFile("/test/.version")
	if !ok {
		t.Fatal("version file not created")
	}

	expected := "1.0.0\n"
	if string(data) != expected {
		t.Errorf("version = %q, want %q", string(data), expected)
	}

	if mod.CurrentVersion != "1.0.0" {
		t.Errorf("module CurrentVersion = %q, want %q", mod.CurrentVersion, "1.0.0")
	}
}
