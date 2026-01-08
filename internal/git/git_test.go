package git

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultCloneOrUpdate(t *testing.T) {
	t.Run("ExistingRepo", func(t *testing.T) {
		tempDir := setupTestRepo(t)

		// Mock UpdateRepo to verify it's being called
		originalUpdateRepo := UpdateRepo
		defer func() { UpdateRepo = originalUpdateRepo }()

		UpdateRepo = func(repoPath string) error {
			if repoPath != tempDir {
				t.Errorf("UpdateRepo called with wrong path: got %s, want %s", repoPath, tempDir)
			}
			return nil
		}

		err := DefaultCloneOrUpdate("https://github.com/octocat/Hello-World.git", tempDir)
		if err != nil {
			t.Fatalf("DefaultCloneOrUpdate failed: %v", err)
		}
	})

	t.Run("NonExistingRepo", func(t *testing.T) {
		tempDir := t.TempDir()
		destRepo := filepath.Join(tempDir, "new_repo")

		// Mock CloneRepoFunc to verify it's being called
		originalCloneRepo := CloneRepoFunc
		defer func() { CloneRepoFunc = originalCloneRepo }()

		CloneRepoFunc = func(repoURL, repoPath string) error {
			if repoPath != destRepo {
				t.Errorf("CloneRepoFunc called with wrong path: got %s, want %s", repoPath, destRepo)
			}
			return nil
		}

		err := DefaultCloneOrUpdate("https://github.com/octocat/Hello-World.git", destRepo)
		if err != nil {
			t.Fatalf("DefaultCloneOrUpdate failed: %v", err)
		}
	})
}

func TestIsValidGitRepo_ValidRepo(t *testing.T) {
	tempDir := setupTestRepo(t)
	defer os.RemoveAll(tempDir)

	if !IsValidGitRepo(tempDir) {
		t.Errorf("Expected IsValidGitRepo to return true for a valid repo, but got false")
	}
}

func TestIsValidGitRepo_InvalidRepo(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "tempo_non_git_test")
	_ = os.RemoveAll(tempDir) // Ensure cleanup before creating

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if IsValidGitRepo(tempDir) {
		t.Errorf("Expected IsValidGitRepo to return false for a non-Git directory, but got true")
	}
}

func TestCloneRepo(t *testing.T) {
	sourceRepo := setupTestRepo(t)

	tempDir := t.TempDir()
	destRepo := filepath.Join(tempDir, "cloned_repo")

	err := CloneRepo(sourceRepo, destRepo)
	if err != nil {
		t.Fatalf("CloneRepo failed: %v", err)
	}

	// Verify the cloned repo exists and is a valid git repository
	gitDir := filepath.Join(destRepo, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Fatalf("Cloned repo does not contain .git directory")
	}
}

func TestUpdateRepo(t *testing.T) {
	sourceRepo := setupTestRepo(t)

	tempDir := t.TempDir()
	destRepo := filepath.Join(tempDir, "cloned_repo")

	err := CloneRepo(sourceRepo, destRepo)
	if err != nil {
		t.Fatalf("CloneRepo failed: %v", err)
	}

	err = UpdateRepo(destRepo)
	if err != nil {
		t.Fatalf("UpdateRepo failed: %v", err)
	}
}

func TestForceReclone(t *testing.T) {
	sourceRepo := setupTestRepo(t)

	tempDir := t.TempDir()
	destRepo := filepath.Join(tempDir, "cloned_repo")

	// First clone
	err := CloneRepo(sourceRepo, destRepo)
	if err != nil {
		t.Fatalf("CloneRepo failed: %v", err)
	}

	// Force re-clone
	err = ForceReclone(sourceRepo, destRepo)
	if err != nil {
		t.Fatalf("ForceReclone failed: %v", err)
	}

	// Ensure the repo was recreated
	gitDir := filepath.Join(destRepo, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Fatalf("Recloned repo does not contain .git directory")
	}
}

func TestFailRemoveExistingRepo(t *testing.T) {
	repoPath := setupReadOnlyDir(t)

	// Attempt to force re-clone (expected to fail)
	err := ForceReclone("https://github.com/octocat/Hello-World.git", repoPath)

	if err == nil {
		t.Fatal("Expected failure when removing repo, but got nil")
	}

	if errors.Is(err, os.ErrPermission) {
		t.Log("Received expected permission error, test passed.")
		return
	}

	expectedError := "Failed to remove existing repository"
	if !strings.Contains(err.Error(), expectedError) && !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("Expected error containing: %q or 'permission denied', got: %q", expectedError, err.Error())
	}
}

func TestFailCloneRepo(t *testing.T) {
	tempDir := t.TempDir()
	destRepo := filepath.Join(tempDir, "cloned_repo")

	// Attempt to clone from an invalid repo URL
	err := CloneRepo("https://invalid.repo.url/nonexistent.git", destRepo)
	if err == nil {
		t.Fatal("Expected failure due to invalid repo URL, but got nil")
	}

	// Ensure error correctly wraps Git failure
	expectedErrors := []string{
		"Failed to clone repository",
		"command failed: exit status 128",
		"command \"git\" failed",
		"fatal",
		"Could not resolve host",
	}

	matched := false
	for _, expected := range expectedErrors {
		if strings.Contains(err.Error(), expected) {
			matched = true
			break
		}
	}

	if !matched {
		t.Fatalf("Expected error containing one of %q, got: %q", expectedErrors, err.Error())
	}
}

/* ------------------------------------------------------------------------- */
/* Helper Functions                                                          */
/* ------------------------------------------------------------------------- */

func setupTestRepo(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Initialize Git repository
	cmd := exec.CommandContext(context.Background(), "git", "init", tempDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Ensure Git has user config
	cmd = exec.CommandContext(context.Background(), "git", "-C", tempDir, "config", "--local", "user.email", "test@example.com")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git user.email: %v", err)
	}

	cmd = exec.CommandContext(context.Background(), "git", "-C", tempDir, "config", "--local", "user.name", "Test User")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to set git user.name: %v", err)
	}

	// Create a dummy file
	testFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add file to Git
	cmd = exec.CommandContext(context.Background(), "git", "-C", tempDir, "add", "testfile.txt")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file to git: %v", err)
	}

	// Commit changes
	cmd = exec.CommandContext(context.Background(), "git", "-C", tempDir, "commit", "-m", "Initial commit")
	if err := cmd.Run(); err != nil {
		// If commit fails, print git status for debugging
		statusCmd := exec.CommandContext(context.Background(), "git", "-C", tempDir, "status")
		statusOutput, _ := statusCmd.CombinedOutput()
		t.Fatalf("Failed to commit in source repo: %v\nGit Status:\n%s", err, statusOutput)
	}

	return tempDir
}

func setupReadOnlyDir(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	repoPath := filepath.Join(tempDir, "repo")
	if err := os.Mkdir(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a read-only lock file inside the repo
	lockFile := filepath.Join(repoPath, "lock")
	if err := os.WriteFile(lockFile, []byte("lock"), 0444); err != nil {
		t.Fatalf("Failed to create lock file: %v", err)
	}

	// Set directory as read-only after file creation
	if err := os.Chmod(repoPath, 0555); err != nil {
		t.Fatalf("Failed to set repo as read-only: %v", err)
	}

	// Cleanup function to remove the lock file explicitly before TempDir cleanup
	t.Cleanup(func() {
		_ = os.Chmod(repoPath, 0755) // Restore write permissions
		_ = os.Remove(lockFile)      // Remove the lock file first
		_ = os.RemoveAll(repoPath)   // Now safely remove the repo
	})

	return repoPath
}
