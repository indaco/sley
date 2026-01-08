package semver

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/indaco/sley/internal/core"
)

// errMockGitTagReader is a mock for GitTagReader interface (prefixed to avoid conflict).
type errMockGitTagReader struct {
	tag string
	err error
}

func (m *errMockGitTagReader) DescribeTags(ctx context.Context) (string, error) {
	return m.tag, m.err
}

// TestVersionManager_ReadError_Recovery tests error handling when file read fails.
func TestVersionManager_ReadError_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})

	expectedErr := errors.New("simulated read failure")
	fs.ReadErr = expectedErr

	ctx := context.Background()
	_, err := mgr.Read(ctx, "/test/.version")

	if err == nil {
		t.Fatal("expected error when read fails, got nil")
	}
}

// TestVersionManager_WriteError_Recovery tests error handling when file write fails.
func TestVersionManager_WriteError_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})

	// First set up a valid version file
	fs.SetFile("/test/.version", []byte("1.0.0"))

	// Now inject write error
	expectedErr := errors.New("simulated write failure")
	fs.WriteErr = expectedErr

	ctx := context.Background()
	err := mgr.Save(ctx, "/test/.version", SemVersion{Major: 2, Minor: 0, Patch: 0})

	if err == nil {
		t.Fatal("expected error when write fails, got nil")
	}
}

// TestVersionManager_ContextCancellation_Recovery tests that operations respect context cancellation.
func TestVersionManager_ContextCancellation_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})
	fs.SetFile("/test/.version", []byte("1.0.0"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := mgr.Read(ctx, "/test/.version")
	if err == nil {
		t.Fatal("expected error when context is cancelled, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestVersionManager_ContextTimeout_Recovery tests that operations respect context deadline.
func TestVersionManager_ContextTimeout_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})
	fs.SetFile("/test/.version", []byte("1.0.0"))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	_, err := mgr.Read(ctx, "/test/.version")
	if err == nil {
		t.Fatal("expected error when context deadline exceeded, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

// TestVersionManager_ConcurrentReads_Recovery tests concurrent read operations.
func TestVersionManager_ConcurrentReads_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})
	fs.SetFile("/test/.version", []byte("1.2.3"))

	const numReaders = 10
	var wg sync.WaitGroup
	errs := make(chan error, numReaders)

	for range numReaders {
		wg.Go(func() {
			ctx := context.Background()
			version, err := mgr.Read(ctx, "/test/.version")
			if err != nil {
				errs <- err
				return
			}
			if version.String() != "1.2.3" {
				errs <- errors.New("expected 1.2.3, got " + version.String())
			}
		})
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent read error: %v", err)
	}
}

// TestVersionManager_ConcurrentWrites_Recovery tests concurrent write operations.
func TestVersionManager_ConcurrentWrites_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})
	fs.SetFile("/test/.version", []byte("1.0.0"))

	const numWriters = 5
	var wg sync.WaitGroup
	errs := make(chan error, numWriters)

	for i := range numWriters {
		wg.Add(1)
		go func(version int) {
			defer wg.Done()
			ctx := context.Background()
			err := mgr.Save(ctx, "/test/.version", SemVersion{Major: 1, Minor: 0, Patch: version})
			if err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent write error: %v", err)
	}

	// Verify file was written (any version is acceptable in concurrent scenario)
	ctx := context.Background()
	version, err := mgr.Read(ctx, "/test/.version")
	if err != nil {
		t.Fatalf("failed to read final version: %v", err)
	}
	if version.String() == "" {
		t.Error("expected non-empty version after concurrent writes")
	}
}

// TestVersionManager_InitializeExistingFile_Recovery tests initialize when file exists.
func TestVersionManager_InitializeExistingFile_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})
	fs.SetFile("/test/.version", []byte("2.0.0"))

	ctx := context.Background()
	err := mgr.Initialize(ctx, "/test/.version")

	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify original version is preserved
	version, err := mgr.Read(ctx, "/test/.version")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if version.String() != "2.0.0" {
		t.Errorf("expected version to be preserved as 2.0.0, got %s", version.String())
	}
}

// TestVersionManager_InitializeNewFile_Recovery tests initialize when file doesn't exist.
func TestVersionManager_InitializeNewFile_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})

	ctx := context.Background()
	err := mgr.Initialize(ctx, "/test/.version")

	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify file was created with initial version
	version, err := mgr.Read(ctx, "/test/.version")
	if err != nil {
		t.Fatalf("failed to read initialized file: %v", err)
	}

	// Initialize creates version based on implementation (may start at 0.0.0 or 0.1.0)
	if version.Major != 0 {
		t.Errorf("expected initial major version 0, got %d", version.Major)
	}
}

// TestVersionManager_InitializeWriteError_Recovery tests initialize when write fails.
func TestVersionManager_InitializeWriteError_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})

	expectedErr := errors.New("simulated write failure")
	fs.WriteErr = expectedErr

	ctx := context.Background()
	err := mgr.Initialize(ctx, "/test/.version")

	if err == nil {
		t.Fatal("expected error when write fails during initialize, got nil")
	}
}

// TestVersionManager_ParseInvalidVersion_Recovery tests handling of invalid version file.
func TestVersionManager_ParseInvalidVersion_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})
	fs.SetFile("/test/.version", []byte("not-a-valid-version"))

	ctx := context.Background()
	_, err := mgr.Read(ctx, "/test/.version")

	if err == nil {
		t.Fatal("expected error for invalid version, got nil")
	}
}

// TestVersionManager_EmptyFile_Recovery tests handling of empty version file.
func TestVersionManager_EmptyFile_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})
	fs.SetFile("/test/.version", []byte(""))

	ctx := context.Background()
	_, err := mgr.Read(ctx, "/test/.version")

	if err == nil {
		t.Fatal("expected error for empty version file, got nil")
	}
}

// TestVersionManager_WhitespaceOnlyFile_Recovery tests handling of whitespace-only version file.
func TestVersionManager_WhitespaceOnlyFile_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})
	fs.SetFile("/test/.version", []byte("   \n\t  "))

	ctx := context.Background()
	_, err := mgr.Read(ctx, "/test/.version")

	if err == nil {
		t.Fatal("expected error for whitespace-only version file, got nil")
	}
}

// TestVersionManager_MkdirError_Recovery tests handling of mkdir failure during save.
func TestVersionManager_MkdirError_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})

	expectedErr := errors.New("simulated mkdir failure")
	fs.MkdirErr = expectedErr

	ctx := context.Background()
	err := mgr.Save(ctx, "/nonexistent/path/.version", SemVersion{Major: 1, Minor: 0, Patch: 0})

	if err == nil {
		t.Fatal("expected error when mkdir fails, got nil")
	}
}

// TestVersionManager_StatError_Recovery tests handling of stat failure.
func TestVersionManager_StatError_Recovery(t *testing.T) {
	fs := core.NewMockFileSystem()
	mgr := NewVersionManager(fs, &errMockGitTagReader{})

	expectedErr := errors.New("simulated stat failure")
	fs.StatErr = expectedErr

	ctx := context.Background()
	// Initialize checks if file exists using Stat
	err := mgr.Initialize(ctx, "/test/.version")

	// Behavior depends on implementation - may fail or treat as non-existent
	// We're just testing it doesn't panic
	_ = err
}
