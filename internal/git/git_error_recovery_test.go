package git

import (
	"context"
	"testing"
	"time"
)

// TestCloneRepo_ContextCancellation tests that CloneRepo respects context cancellation.
func TestCloneRepo_ContextCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tempDir := t.TempDir()
	err := CloneRepo(ctx, "https://github.com/octocat/Hello-World.git", tempDir)

	if err == nil {
		t.Fatal("expected error when context is cancelled, got nil")
	}
}

// TestCloneRepo_ContextTimeout tests that CloneRepo respects context deadline.
func TestCloneRepo_ContextTimeout(t *testing.T) {
	t.Parallel(
	// Use an extremely short timeout to trigger deadline exceeded
	)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for the context to actually expire instead of sleeping
	<-ctx.Done()

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
	t.Parallel()
	sourceRepo := setupTestRepo(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := UpdateRepo(ctx, sourceRepo)

	if err == nil {
		t.Fatal("expected error when context is cancelled, got nil")
	}
}

// TestCloneOrUpdate_ContextCancellation tests context cancellation for clone or update.
func TestCloneOrUpdate_ContextCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tempDir := t.TempDir()
	err := CloneOrUpdate(ctx, "https://github.com/octocat/Hello-World.git", tempDir)

	if err == nil {
		t.Fatal("expected error when context is cancelled, got nil")
	}
}

// TestForceReclone_ContextCancellation tests that ForceReclone respects context cancellation.
func TestForceReclone_ContextCancellation(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
	ctx := context.Background()
	err := UpdateRepo(ctx, "/nonexistent/path/to/repo")

	if err == nil {
		t.Fatal("expected error for non-existent repo, got nil")
	}
}

// TestCloneOrUpdate_UpdateError tests error propagation when update fails.
func TestCloneOrUpdate_UpdateError(t *testing.T) {
	t.Parallel(
	// Create a repo with no remote configured - git pull will fail
	)

	sourceRepo := setupTestRepo(t)

	ctx := context.Background()
	err := CloneOrUpdate(ctx, "https://github.com/test/repo.git", sourceRepo)

	if err == nil {
		t.Fatal("expected error when git pull fails on repo with no remote, got nil")
	}
}

// TestCloneOrUpdate_CloneError tests error propagation when clone fails.
func TestCloneOrUpdate_CloneError(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	destPath := tempDir + "/new"

	ctx := context.Background()
	err := CloneOrUpdate(ctx, "https://invalid.repo.url/nonexistent.git", destPath)

	if err == nil {
		t.Fatal("expected error when clone fails with invalid URL, got nil")
	}
}
