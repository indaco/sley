package semver

import (
	"context"
	"errors"
	"testing"

	"github.com/indaco/sley/internal/core"
)

func TestVersionManager_Read(t *testing.T) {
	mockFS := core.NewMockFileSystem()
	ctx := context.Background()
	mockFS.SetFile("/test/.version", []byte("1.2.3-alpha.1+build.123\n"))

	mgr := NewVersionManager(mockFS, nil)

	v, err := mgr.Read(ctx, "/test/.version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Errorf("expected 1.2.3, got %d.%d.%d", v.Major, v.Minor, v.Patch)
	}
	if v.PreRelease != "alpha.1" {
		t.Errorf("expected pre-release 'alpha.1', got %q", v.PreRelease)
	}
	if v.Build != "build.123" {
		t.Errorf("expected build 'build.123', got %q", v.Build)
	}
}

func TestVersionManager_Read_FileNotFound(t *testing.T) {
	mockFS := core.NewMockFileSystem()
	ctx := context.Background()
	mgr := NewVersionManager(mockFS, nil)

	_, err := mgr.Read(ctx, "/nonexistent/.version")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVersionManager_Read_InvalidVersion(t *testing.T) {
	mockFS := core.NewMockFileSystem()
	ctx := context.Background()
	mockFS.SetFile("/test/.version", []byte("not-a-version\n"))

	mgr := NewVersionManager(mockFS, nil)

	_, err := mgr.Read(ctx, "/test/.version")
	if err == nil {
		t.Fatal("expected error for invalid version, got nil")
	}
}

func TestVersionManager_Save(t *testing.T) {
	mockFS := core.NewMockFileSystem()
	ctx := context.Background()
	mgr := NewVersionManager(mockFS, nil)

	v := SemVersion{Major: 2, Minor: 0, Patch: 0, PreRelease: "beta.1"}

	err := mgr.Save(ctx, "/test/.version", v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, ok := mockFS.GetFile("/test/.version")
	if !ok {
		t.Fatal("file not found in mock filesystem")
	}

	expected := "2.0.0-beta.1\n"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestVersionManager_Initialize_WithGitTag(t *testing.T) {
	mockFS := core.NewMockFileSystem()
	mockGit := &MockGitTagReader{Tag: "v1.5.0\n"}

	mgr := NewVersionManager(mockFS, mockGit)
	ctx := context.Background()

	err := mgr.Initialize(ctx, "/test/.version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, ok := mockFS.GetFile("/test/.version")
	if !ok {
		t.Fatal("version file not created")
	}

	expected := "1.5.0\n"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestVersionManager_Initialize_WithoutGitTag(t *testing.T) {
	mockFS := core.NewMockFileSystem()
	mockGit := &MockGitTagReader{Err: errors.New("no tags")}

	mgr := NewVersionManager(mockFS, mockGit)
	ctx := context.Background()

	err := mgr.Initialize(ctx, "/test/.version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, ok := mockFS.GetFile("/test/.version")
	if !ok {
		t.Fatal("version file not created")
	}

	expected := "0.1.0\n"
	if string(data) != expected {
		t.Errorf("expected default %q, got %q", expected, string(data))
	}
}

func TestVersionManager_Initialize_FileExists(t *testing.T) {
	mockFS := core.NewMockFileSystem()
	mockFS.SetFile("/test/.version", []byte("9.9.9\n"))

	mgr := NewVersionManager(mockFS, nil)
	ctx := context.Background()

	err := mgr.Initialize(ctx, "/test/.version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should be unchanged
	data, _ := mockFS.GetFile("/test/.version")
	if string(data) != "9.9.9\n" {
		t.Errorf("file should be unchanged, got %q", string(data))
	}
}

func TestVersionManager_InitializeWithFeedback(t *testing.T) {
	tests := []struct {
		name        string
		fileExists  bool
		wantCreated bool
	}{
		{"file exists", true, false},
		{"file does not exist", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := core.NewMockFileSystem()
			if tt.fileExists {
				mockFS.SetFile("/test/.version", []byte("1.0.0\n"))
			}

			mgr := NewVersionManager(mockFS, &MockGitTagReader{Err: errors.New("no tags")})
			ctx := context.Background()

			created, err := mgr.InitializeWithFeedback(ctx, "/test/.version")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if created != tt.wantCreated {
				t.Errorf("created = %v, want %v", created, tt.wantCreated)
			}
		})
	}
}

func TestVersionManager_Update(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		bumpType string
		pre      string
		meta     string
		preserve bool
		expected string
	}{
		{"patch bump", "1.2.3\n", "patch", "", "", false, "1.2.4\n"},
		{"minor bump", "1.2.3\n", "minor", "", "", false, "1.3.0\n"},
		{"major bump", "1.2.3\n", "major", "", "", false, "2.0.0\n"},
		{"with pre-release", "1.2.3\n", "patch", "alpha.1", "", false, "1.2.4-alpha.1\n"},
		{"with metadata", "1.2.3\n", "patch", "", "build.1", false, "1.2.4+build.1\n"},
		{"preserve metadata", "1.2.3+old\n", "patch", "", "", true, "1.2.4+old\n"},
		{"override metadata", "1.2.3+old\n", "patch", "", "new", false, "1.2.4+new\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockFS := core.NewMockFileSystem()
			mockFS.SetFile("/test/.version", []byte(tt.initial))

			mgr := NewVersionManager(mockFS, nil)

			err := mgr.Update(ctx, "/test/.version", tt.bumpType, tt.pre, tt.meta, tt.preserve)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			data, _ := mockFS.GetFile("/test/.version")
			if string(data) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(data))
			}
		})
	}
}

func TestVersionManager_Update_InvalidBumpType(t *testing.T) {
	ctx := context.Background()
	mockFS := core.NewMockFileSystem()
	mockFS.SetFile("/test/.version", []byte("1.0.0\n"))

	mgr := NewVersionManager(mockFS, nil)

	err := mgr.Update(ctx, "/test/.version", "invalid", "", "", false)
	if err == nil {
		t.Fatal("expected error for invalid bump type, got nil")
	}
}

func TestSetDefaultManager(t *testing.T) {
	mockFS := core.NewMockFileSystem()
	mockFS.SetFile("/test/.version", []byte("5.5.5\n"))

	customMgr := NewVersionManager(mockFS, nil)
	restore := SetDefaultManager(customMgr)
	defer restore()

	// Now the legacy functions should use our mock
	v, err := ReadVersion("/test/.version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v.Major != 5 || v.Minor != 5 || v.Patch != 5 {
		t.Errorf("expected 5.5.5, got %d.%d.%d", v.Major, v.Minor, v.Patch)
	}
}

func TestDefaultVersionManager(t *testing.T) {
	mgr := DefaultVersionManager()
	if mgr == nil {
		t.Fatal("DefaultVersionManager returned nil")
	}
	if mgr.fs == nil {
		t.Error("fs is nil")
	}
	if mgr.git == nil {
		t.Error("git is nil")
	}
}
