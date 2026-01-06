package operations

import (
	"context"
	"testing"
	"time"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/workspace"
)

func TestNewPreOperation(t *testing.T) {
	fs := core.NewMockFileSystem()
	op := NewPreOperation(fs, "alpha", true)

	if op == nil {
		t.Fatal("NewPreOperation returned nil")
	}
	if op.fs == nil {
		t.Error("fs is nil")
	}
	if op.label != "alpha" {
		t.Errorf("label = %v, want %v", op.label, "alpha")
	}
	if !op.increment {
		t.Error("increment should be true")
	}
}

func TestPreOperation_Execute_SetLabel(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		label    string
		expected string
	}{
		{
			name:     "set label on stable version",
			initial:  "1.2.3\n",
			label:    "alpha",
			expected: "1.2.4-alpha\n",
		},
		{
			name:     "set label on version with existing pre-release",
			initial:  "1.2.3-beta.1\n",
			label:    "alpha",
			expected: "1.2.3-alpha\n",
		},
		{
			name:     "set rc label",
			initial:  "2.0.0\n",
			label:    "rc",
			expected: "2.0.1-rc\n",
		},
		{
			name:     "set beta.1 label",
			initial:  "0.1.0\n",
			label:    "beta.1",
			expected: "0.1.1-beta.1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test/.version", []byte(tt.initial))

			op := NewPreOperation(fs, tt.label, false)
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
		})
	}
}

func TestPreOperation_Execute_IncrementLabel(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		label    string
		expected string
	}{
		{
			name:     "increment existing alpha.1",
			initial:  "1.2.3-alpha.1\n",
			label:    "alpha",
			expected: "1.2.3-alpha.2\n",
		},
		{
			name:     "increment existing rc.5",
			initial:  "2.0.0-rc.5\n",
			label:    "rc",
			expected: "2.0.0-rc.6\n",
		},
		{
			name:     "increment base label without number",
			initial:  "1.0.0-beta\n",
			label:    "beta",
			expected: "1.0.0-beta.1\n",
		},
		{
			name:     "increment with different base creates new label",
			initial:  "1.2.3-alpha.1\n",
			label:    "beta",
			expected: "1.2.3-beta.1\n",
		},
		{
			name:     "increment from stable version",
			initial:  "1.2.3\n",
			label:    "alpha",
			expected: "1.2.3-alpha.1\n",
		},
		{
			name:     "increment with dash separator",
			initial:  "1.0.0-rc-1\n",
			label:    "rc",
			expected: "1.0.0-rc-2\n",
		},
		{
			name:     "increment with no separator",
			initial:  "1.0.0-rc1\n",
			label:    "rc",
			expected: "1.0.0-rc2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test/.version", []byte(tt.initial))

			op := NewPreOperation(fs, tt.label, true)
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
		})
	}
}

func TestPreOperation_Execute_UpdatesModuleVersion(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewPreOperation(fs, "alpha", false)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if mod.CurrentVersion != "1.2.4-alpha" {
		t.Errorf("module CurrentVersion = %q, want %q", mod.CurrentVersion, "1.2.4-alpha")
	}
}

func TestPreOperation_Execute_ContextCancellation(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewPreOperation(fs, "alpha", false)
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

func TestPreOperation_Execute_ContextTimeout(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewPreOperation(fs, "alpha", false)
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

func TestPreOperation_Execute_ReadError(t *testing.T) {
	fs := core.NewMockFileSystem()
	// Don't set any file, so read will fail

	op := NewPreOperation(fs, "alpha", false)
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

func TestPreOperation_Execute_SaveError(t *testing.T) {
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	// Inject write error - this will be checked when Save is called
	fs.WriteErr = context.DeadlineExceeded

	op := NewPreOperation(fs, "alpha", false)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()

	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected save error, got nil")
	}
}

func TestPreOperation_Name(t *testing.T) {
	tests := []struct {
		name      string
		increment bool
		expected  string
	}{
		{
			name:      "set pre-release name",
			increment: false,
			expected:  "set pre-release",
		},
		{
			name:      "increment pre-release name",
			increment: true,
			expected:  "increment pre-release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			op := NewPreOperation(fs, "alpha", tt.increment)

			name := op.Name()
			if name != tt.expected {
				t.Errorf("Name() = %q, want %q", name, tt.expected)
			}
		})
	}
}

func TestPreOperation_Execute_WithBuildMetadata(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		label    string
		expected string
	}{
		{
			name:     "set preserves build metadata",
			initial:  "1.2.3+build.123\n",
			label:    "alpha",
			expected: "1.2.4-alpha+build.123\n", // Build metadata is preserved from read
		},
		{
			name:     "increment preserves build metadata",
			initial:  "1.2.3-alpha.1+build.123\n",
			label:    "alpha",
			expected: "1.2.3-alpha.2+build.123\n", // Build metadata is preserved from read
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := core.NewMockFileSystem()
			fs.SetFile("/test/.version", []byte(tt.initial))

			increment := tt.name == "increment preserves build metadata"
			op := NewPreOperation(fs, tt.label, increment)
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
		})
	}
}
