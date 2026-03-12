package operations

import (
	"context"
	"testing"
	"time"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/semver"
	"github.com/indaco/sley/internal/workspace"
)

func TestNewBumpOperation(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "alpha", "build123", true)

	if op == nil {
		t.Fatal("NewBumpOperation returned nil")
		return
	}
	if op.fs == nil {
		t.Error("fs is nil")
	}
	if op.bumper == nil {
		t.Error("bumper is nil")
	}
	if op.bumpType != BumpPatch {
		t.Errorf("bumpType = %v, want %v", op.bumpType, BumpPatch)
	}
	if op.preRelease != "alpha" {
		t.Errorf("preRelease = %v, want %v", op.preRelease, "alpha")
	}
	if op.metadata != "build123" {
		t.Errorf("metadata = %v, want %v", op.metadata, "build123")
	}
	if !op.preserveMetadata {
		t.Error("preserveMetadata should be true")
	}
}

func TestBumpOperation_Execute_Patch(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", "", false)
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

	expected := "1.2.4\n"
	if string(data) != expected {
		t.Errorf("version = %q, want %q", string(data), expected)
	}

	if mod.CurrentVersion != "1.2.4" {
		t.Errorf("module CurrentVersion = %q, want %q", mod.CurrentVersion, "1.2.4")
	}
}

func TestBumpOperation_Execute_Minor(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpMinor, "", "", false)
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

	expected := "1.3.0\n"
	if string(data) != expected {
		t.Errorf("version = %q, want %q", string(data), expected)
	}

	if mod.CurrentVersion != "1.3.0" {
		t.Errorf("module CurrentVersion = %q, want %q", mod.CurrentVersion, "1.3.0")
	}
}

func TestBumpOperation_Execute_Major(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpMajor, "", "", false)
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

	expected := "2.0.0\n"
	if string(data) != expected {
		t.Errorf("version = %q, want %q", string(data), expected)
	}

	if mod.CurrentVersion != "2.0.0" {
		t.Errorf("module CurrentVersion = %q, want %q", mod.CurrentVersion, "2.0.0")
	}
}

func TestBumpOperation_Execute_Release(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3-beta.1+build.123\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpRelease, "", "", false)
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

	expected := "1.2.3\n"
	if string(data) != expected {
		t.Errorf("version = %q, want %q", string(data), expected)
	}

	if mod.CurrentVersion != "1.2.3" {
		t.Errorf("module CurrentVersion = %q, want %q", mod.CurrentVersion, "1.2.3")
	}
}

func TestBumpOperation_Execute_Auto(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		initial  string
		expected string
	}{
		{
			name:     "auto bump pre-release",
			initial:  "1.2.3-beta.1\n",
			expected: "1.2.3\n",
		},
		{
			name:     "auto bump stable",
			initial:  "1.2.3\n",
			expected: "1.2.4\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := core.NewMockFileSystem()
			fs.SetFile("/test/.version", []byte(tt.initial))

			op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpAuto, "", "", false)
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

func TestBumpOperation_Execute_WithPreRelease(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "alpha.1", "", false)
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

	expected := "1.2.4-alpha.1\n"
	if string(data) != expected {
		t.Errorf("version = %q, want %q", string(data), expected)
	}
}

func TestBumpOperation_Execute_WithMetadata(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", "build.456", false)
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

	expected := "1.2.4+build.456\n"
	if string(data) != expected {
		t.Errorf("version = %q, want %q", string(data), expected)
	}
}

func TestBumpOperation_Execute_PreserveMetadata(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		initial          string
		preserveMetadata bool
		newMetadata      string
		expected         string
	}{
		{
			name:             "preserve existing metadata",
			initial:          "1.2.3+old.meta\n",
			preserveMetadata: true,
			newMetadata:      "",
			expected:         "1.2.4+old.meta\n",
		},
		{
			name:             "override with new metadata",
			initial:          "1.2.3+old.meta\n",
			preserveMetadata: false,
			newMetadata:      "new.meta",
			expected:         "1.2.4+new.meta\n",
		},
		{
			name:             "no metadata preserved when none exists",
			initial:          "1.2.3\n",
			preserveMetadata: true,
			newMetadata:      "",
			expected:         "1.2.4\n",
		},
		{
			name:             "new metadata overrides preserve flag",
			initial:          "1.2.3+old.meta\n",
			preserveMetadata: true,
			newMetadata:      "new.meta",
			expected:         "1.2.4+new.meta\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := core.NewMockFileSystem()
			fs.SetFile("/test/.version", []byte(tt.initial))

			op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", tt.newMetadata, tt.preserveMetadata)
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

func TestBumpOperation_Execute_ContextCancellation(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", "", false)
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

func TestBumpOperation_Execute_ContextTimeout(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", "", false)
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

func TestBumpOperation_Execute_ReadError(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	// Don't set any file, so read will fail

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", "", false)
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

// mockErrorBumper is a VersionBumper that always returns errors.
type mockErrorBumper struct {
	err error
}

func (m mockErrorBumper) BumpNext(_ semver.SemVersion) (semver.SemVersion, error) {
	return semver.SemVersion{}, m.err
}

func (m mockErrorBumper) BumpByLabel(_ semver.SemVersion, _ string) (semver.SemVersion, error) {
	return semver.SemVersion{}, m.err
}

func TestBumpOperation_Execute_AutoBumpError(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	// Use a mock bumper that returns an error
	bumper := mockErrorBumper{err: context.DeadlineExceeded}

	op := NewBumpOperation(fs, bumper, BumpAuto, "", "", false)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected auto bump error, got nil")
	}
}

func TestBumpOperation_Execute_UnknownBumpType(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpType("unknown"), "", "", false)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected unknown bump type error, got nil")
	}
}

func TestBumpOperation_Execute_SaveError(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	// Inject write error - this will be checked when Save is called
	fs.WriteErr = context.DeadlineExceeded

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", "", false)
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

func TestBumpOperation_Name(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		bumpType BumpType
		expected string
	}{
		{
			name:     "patch",
			bumpType: BumpPatch,
			expected: "bump patch",
		},
		{
			name:     "minor",
			bumpType: BumpMinor,
			expected: "bump minor",
		},
		{
			name:     "major",
			bumpType: BumpMajor,
			expected: "bump major",
		},
		{
			name:     "release",
			bumpType: BumpRelease,
			expected: "bump release",
		},
		{
			name:     "auto",
			bumpType: BumpAuto,
			expected: "bump auto",
		},
		{
			name:     "pre",
			bumpType: BumpPre,
			expected: "bump pre",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := core.NewMockFileSystem()
			op := NewBumpOperation(fs, semver.NewDefaultBumper(), tt.bumpType, "", "", false)

			name := op.Name()
			if name != tt.expected {
				t.Errorf("Name() = %q, want %q", name, tt.expected)
			}
		})
	}
}

func TestBumpOperation_Execute_Pre_IncrementExisting(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		initial  string
		label    string
		expected string
	}{
		{
			name:     "increment rc.1 to rc.2",
			initial:  "1.2.3-rc.1\n",
			label:    "",
			expected: "1.2.3-rc.2\n",
		},
		{
			name:     "increment rc.9 to rc.10",
			initial:  "1.2.3-rc.9\n",
			label:    "",
			expected: "1.2.3-rc.10\n",
		},
		{
			name:     "increment beta.1 to beta.2",
			initial:  "2.0.0-beta.1\n",
			label:    "",
			expected: "2.0.0-beta.2\n",
		},
		{
			name:     "increment alpha to alpha.1 with no number",
			initial:  "1.0.0-alpha\n",
			label:    "",
			expected: "1.0.0-alpha.1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := core.NewMockFileSystem()
			fs.SetFile("/test/.version", []byte(tt.initial))

			op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPre, tt.label, "", false)
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

func TestBumpOperation_Execute_Pre_WithLabel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		initial  string
		label    string
		expected string
	}{
		{
			name:     "add rc to stable version",
			initial:  "1.2.3\n",
			label:    "rc",
			expected: "1.2.3-rc.1\n",
		},
		{
			name:     "switch from alpha to beta",
			initial:  "1.2.3-alpha.3\n",
			label:    "beta",
			expected: "1.2.3-beta.1\n",
		},
		{
			name:     "switch from rc to alpha",
			initial:  "1.2.3-rc.1\n",
			label:    "alpha",
			expected: "1.2.3-alpha.1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := core.NewMockFileSystem()
			fs.SetFile("/test/.version", []byte(tt.initial))

			op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPre, tt.label, "", false)
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

func TestBumpOperation_Execute_Pre_NoExistingPreRelease(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	// BumpPre with no label and no existing pre-release should fail
	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPre, "", "", false)
	mod := &workspace.Module{
		Name: "test",
		Path: "/test/.version",
	}

	ctx := context.Background()
	err := op.Execute(ctx, mod)
	if err == nil {
		t.Fatal("expected error when no pre-release exists and no label provided")
	}
}

func TestBumpOperation_Execute_Pre_PreserveMetadata(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3-rc.1+build.99\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPre, "", "", true)
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

	expected := "1.2.3-rc.2+build.99\n"
	if string(data) != expected {
		t.Errorf("version = %q, want %q", string(data), expected)
	}
}

func TestBumpOperation_Execute_Pre_WithNewMetadata(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3-rc.1\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPre, "", "ci.456", false)
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

	expected := "1.2.3-rc.2+ci.456\n"
	if string(data) != expected {
		t.Errorf("version = %q, want %q", string(data), expected)
	}
}

func TestExtractPreReleaseBase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"rc.1", "rc"},
		{"rc.10", "rc"},
		{"beta.2", "beta"},
		{"alpha.1", "alpha"},
		{"rc1", "rc"},
		{"beta5", "beta"},
		{"rc-1", "rc-"},
		{"alpha", "alpha"},
		{"rc", "rc"},
		{"1", "1"},           // Pure number returns as-is (edge case)
		{"123", "123"},       // Pure number returns as-is (edge case)
		{"dev.1.2", "dev.1"}, // Multiple dots, extracts up to last numeric part
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := semver.ExtractPreReleaseBase(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractPreReleaseBase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBumpOperation_Preview_ReturnsVersionsWithoutWriting(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", "", false)

	ctx := context.Background()
	result, err := op.Preview(ctx, "/test/.version")
	if err != nil {
		t.Fatalf("Preview failed: %v", err)
	}

	if result.PreviousVersion.String() != "1.2.3" {
		t.Errorf("PreviousVersion = %q, want %q", result.PreviousVersion.String(), "1.2.3")
	}
	if result.NewVersion.String() != "1.2.4" {
		t.Errorf("NewVersion = %q, want %q", result.NewVersion.String(), "1.2.4")
	}

	// Verify file was NOT modified
	data, ok := fs.GetFile("/test/.version")
	if !ok {
		t.Fatal("version file not found")
	}
	if string(data) != "1.2.3\n" {
		t.Errorf("file was modified by Preview: got %q, want %q", string(data), "1.2.3\n")
	}
}

func TestBumpOperation_Preview_Minor(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpMinor, "", "", false)

	result, err := op.Preview(context.Background(), "/test/.version")
	if err != nil {
		t.Fatalf("Preview failed: %v", err)
	}

	if result.NewVersion.String() != "1.3.0" {
		t.Errorf("NewVersion = %q, want %q", result.NewVersion.String(), "1.3.0")
	}
}

func TestBumpOperation_Preview_ContextCancellation(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", "", false)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := op.Preview(ctx, "/test/.version")
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

func TestBumpOperation_Preview_ReadError(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", "", false)

	_, err := op.Preview(context.Background(), "/test/.version")
	if err == nil {
		t.Fatal("expected read error, got nil")
	}
}

func TestBumpOperation_Write(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", "", false)
	newVer := semver.SemVersion{Major: 1, Minor: 2, Patch: 4}

	err := op.Write(context.Background(), "/test/.version", newVer)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	data, ok := fs.GetFile("/test/.version")
	if !ok {
		t.Fatal("version file not found")
	}
	if string(data) != "1.2.4\n" {
		t.Errorf("version = %q, want %q", string(data), "1.2.4\n")
	}
}

func TestBumpOperation_Write_ContextCancellation(t *testing.T) {
	t.Parallel()
	fs := core.NewMockFileSystem()
	fs.SetFile("/test/.version", []byte("1.2.3\n"))

	op := NewBumpOperation(fs, semver.NewDefaultBumper(), BumpPatch, "", "", false)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := op.Write(ctx, "/test/.version", semver.SemVersion{Major: 1, Minor: 2, Patch: 4})
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

func TestBumpOperation_PreviewThenWrite_EqualsExecute(t *testing.T) {
	t.Parallel()
	fs1 := core.NewMockFileSystem()
	fs1.SetFile("/test/.version", []byte("1.2.3\n"))

	fs2 := core.NewMockFileSystem()
	fs2.SetFile("/test/.version", []byte("1.2.3\n"))

	ctx := context.Background()

	// Path 1: Preview + Write
	op1 := NewBumpOperation(fs1, semver.NewDefaultBumper(), BumpMinor, "alpha", "build.1", false)
	result, err := op1.Preview(ctx, "/test/.version")
	if err != nil {
		t.Fatalf("Preview failed: %v", err)
	}
	if err := op1.Write(ctx, "/test/.version", result.NewVersion); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Path 2: Execute
	op2 := NewBumpOperation(fs2, semver.NewDefaultBumper(), BumpMinor, "alpha", "build.1", false)
	mod := &workspace.Module{Name: "test", Path: "/test/.version"}
	if err := op2.Execute(ctx, mod); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	data1, _ := fs1.GetFile("/test/.version")
	data2, _ := fs2.GetFile("/test/.version")
	if string(data1) != string(data2) {
		t.Errorf("Preview+Write = %q, Execute = %q", string(data1), string(data2))
	}
}
