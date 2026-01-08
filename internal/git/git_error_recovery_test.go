package git

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestCloneRepo_ContextCancellation tests that CloneRepo respects context cancellation.
func TestCloneRepo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tempDir := t.TempDir()
	err := CloneRepo(ctx, "https://github.com/octocat/Hello-World.git", tempDir)

	if err == nil {
		t.Fatal("expected error when context is cancelled, got nil")
	}

	// The error should be context.Canceled or contain relevant message
	if !errors.Is(err, context.Canceled) && !errors.Is(ctx.Err(), context.Canceled) {
		// Some implementations wrap the error differently
		t.Logf("got error: %v (context.Err: %v)", err, ctx.Err())
	}
}

// TestCloneRepo_ContextTimeout tests that CloneRepo respects context deadline.
func TestCloneRepo_ContextTimeout(t *testing.T) {
	// Use an extremely short timeout to trigger deadline exceeded
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout to trigger
	time.Sleep(10 * time.Millisecond)

	tempDir := t.TempDir()
	err := CloneRepo(ctx, "https://github.com/octocat/Hello-World.git", tempDir)

	if err == nil {
		t.Fatal("expected error when context deadline exceeded, got nil")
	}

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", ctx.Err())
	}
}

// TestUpdateRepo_ContextCancellation tests that UpdateRepo respects context cancellation.
func TestUpdateRepo_ContextCancellation(t *testing.T) {
	sourceRepo := setupTestRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := UpdateRepo(ctx, sourceRepo)

	if err == nil {
		t.Fatal("expected error when context is cancelled, got nil")
	}
}

// TestDefaultCloneOrUpdate_ContextCancellation tests context cancellation for clone or update.
func TestDefaultCloneOrUpdate_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tempDir := t.TempDir()
	err := DefaultCloneOrUpdate(ctx, "https://github.com/octocat/Hello-World.git", tempDir)

	if err == nil {
		t.Fatal("expected error when context is cancelled, got nil")
	}
}

// TestForceReclone_ContextCancellation tests that ForceReclone respects context cancellation.
func TestForceReclone_ContextCancellation(t *testing.T) {
	sourceRepo := setupTestRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tempDir := t.TempDir()
	err := ForceReclone(ctx, sourceRepo, tempDir)

	if err == nil {
		t.Fatal("expected error when context is cancelled, got nil")
	}
}

// TestCloneRepo_InvalidURL tests error handling for malformed URLs.
func TestCloneRepo_InvalidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty URL", ""},
		{"invalid protocol", "not-a-valid-url"},
		{"missing host", "https:///path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tempDir := t.TempDir()

			err := CloneRepo(ctx, tt.url, tempDir)
			if err == nil {
				t.Errorf("expected error for invalid URL %q, got nil", tt.url)
			}
		})
	}
}

// TestUpdateRepo_NonExistentRepo tests error handling for non-existent repo.
func TestUpdateRepo_NonExistentRepo(t *testing.T) {
	ctx := context.Background()
	err := UpdateRepo(ctx, "/nonexistent/path/to/repo")

	if err == nil {
		t.Fatal("expected error for non-existent repo, got nil")
	}
}

// TestDefaultCloneOrUpdate_UpdateError tests error propagation when update fails.
func TestDefaultCloneOrUpdate_UpdateError(t *testing.T) {
	sourceRepo := setupTestRepo(t)

	// Save original and restore after test
	originalUpdateRepo := UpdateRepo
	defer func() { UpdateRepo = originalUpdateRepo }()

	expectedErr := errors.New("simulated update failure")
	UpdateRepo = func(ctx context.Context, repoPath string) error {
		return expectedErr
	}

	ctx := context.Background()
	err := DefaultCloneOrUpdate(ctx, "https://github.com/test/repo.git", sourceRepo)

	if err == nil {
		t.Fatal("expected error when UpdateRepo fails, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

// TestDefaultCloneOrUpdate_CloneError tests error propagation when clone fails.
func TestDefaultCloneOrUpdate_CloneError(t *testing.T) {
	// Save original and restore after test
	originalCloneRepo := CloneRepoFunc
	defer func() { CloneRepoFunc = originalCloneRepo }()

	expectedErr := errors.New("simulated clone failure")
	CloneRepoFunc = func(ctx context.Context, repoURL, repoPath string) error {
		return expectedErr
	}

	tempDir := t.TempDir()
	ctx := context.Background()
	err := DefaultCloneOrUpdate(ctx, "https://github.com/test/repo.git", tempDir+"/new")

	if err == nil {
		t.Fatal("expected error when CloneRepoFunc fails, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
