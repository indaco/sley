package core

import (
	"context"
	"errors"
	"testing"
)

func TestMockGitClient(t *testing.T) {
	mockGit := NewMockGitClient()
	ctx := context.Background()

	t.Run("describe tags", func(t *testing.T) {
		mockGit.TagOutput = "v1.0.0"
		tag, err := mockGit.DescribeTags(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tag != "v1.0.0" {
			t.Errorf("expected 'v1.0.0', got %q", tag)
		}
	})

	t.Run("describe tags error", func(t *testing.T) {
		mockGit.TagError = errors.New("no tags")
		_, err := mockGit.DescribeTags(ctx)
		if err == nil || err.Error() != "no tags" {
			t.Errorf("expected 'no tags' error, got %v", err)
		}
		mockGit.TagError = nil
	})

	t.Run("clone", func(t *testing.T) {
		err := mockGit.Clone(ctx, "https://example.com/repo.git", "/tmp/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !mockGit.IsValidRepo("/tmp/repo") {
			t.Error("expected cloned repo to be marked as valid")
		}
	})

	t.Run("is valid repo", func(t *testing.T) {
		mockGit.IsValidRepos["/existing/repo"] = true

		if !mockGit.IsValidRepo("/existing/repo") {
			t.Error("expected /existing/repo to be valid")
		}
		if mockGit.IsValidRepo("/nonexistent") {
			t.Error("expected /nonexistent to be invalid")
		}
	})
}

func TestMockGitClient_Pull(t *testing.T) {
	mockGit := NewMockGitClient()
	ctx := context.Background()

	t.Run("pull success", func(t *testing.T) {
		err := mockGit.Pull(ctx, "/repo")
		if err != nil {
			t.Errorf("Pull() unexpected error: %v", err)
		}
	})

	t.Run("pull error", func(t *testing.T) {
		mockGit.PullError = errors.New("pull failed")
		err := mockGit.Pull(ctx, "/repo")
		if err == nil || err.Error() != "pull failed" {
			t.Errorf("expected pull error, got %v", err)
		}
	})
}

func TestMockGitClient_CloneError(t *testing.T) {
	mockGit := NewMockGitClient()
	ctx := context.Background()

	mockGit.CloneError = errors.New("clone failed")
	err := mockGit.Clone(ctx, "https://example.com/repo.git", "/tmp/repo")
	if err == nil || err.Error() != "clone failed" {
		t.Errorf("expected clone error, got %v", err)
	}
}

func TestMockGitTagOperations_CreateAnnotatedTag(t *testing.T) {
	mock := NewMockGitTagOperations()

	err := mock.CreateAnnotatedTag("v1.0.0", "Release 1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.CreatedTags) != 1 || mock.CreatedTags[0] != "v1.0.0" {
		t.Errorf("expected tag v1.0.0 in CreatedTags, got %v", mock.CreatedTags)
	}

	// Test with error
	mock.CreateAnnotatedTagErr = errors.New("tag error")
	err = mock.CreateAnnotatedTag("v1.0.1", "msg")
	if err == nil || err.Error() != "tag error" {
		t.Errorf("expected 'tag error', got %v", err)
	}
}

func TestMockGitTagOperations_CreateLightweightTag(t *testing.T) {
	mock := NewMockGitTagOperations()

	err := mock.CreateLightweightTag("v2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.CreatedTags) != 1 || mock.CreatedTags[0] != "v2.0.0" {
		t.Errorf("expected tag v2.0.0, got %v", mock.CreatedTags)
	}

	// Test with error
	mock.CreateLightweightTagErr = errors.New("lightweight tag error")
	err = mock.CreateLightweightTag("v2.0.1")
	if err == nil || err.Error() != "lightweight tag error" {
		t.Errorf("expected 'lightweight tag error', got %v", err)
	}
}

func TestMockGitTagOperations_TagExists(t *testing.T) {
	mock := NewMockGitTagOperations()

	mock.TagExistsResult = true
	exists, err := mock.TagExists("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected tag to exist")
	}

	mock.TagExistsResult = false
	exists, err = mock.TagExists("v9.9.9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected tag to not exist")
	}

	// Test with error
	mock.TagExistsErr = errors.New("tag exists error")
	_, err = mock.TagExists("v1.0.0")
	if err == nil || err.Error() != "tag exists error" {
		t.Errorf("expected 'tag exists error', got %v", err)
	}
}

func TestMockGitTagOperations_GetLatestTag(t *testing.T) {
	mock := NewMockGitTagOperations()

	mock.GetLatestTagName = "v3.0.0"
	tag, err := mock.GetLatestTag()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v3.0.0" {
		t.Errorf("expected v3.0.0, got %s", tag)
	}

	// Test with error
	mock.GetLatestTagErr = errors.New("no tags")
	_, err = mock.GetLatestTag()
	if err == nil || err.Error() != "no tags" {
		t.Errorf("expected 'no tags', got %v", err)
	}
}

func TestMockGitTagOperations_PushTag(t *testing.T) {
	mock := NewMockGitTagOperations()

	err := mock.PushTag("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.PushedTags) != 1 || mock.PushedTags[0] != "v1.0.0" {
		t.Errorf("expected pushed tag v1.0.0, got %v", mock.PushedTags)
	}

	// Test with error
	mock.PushTagErr = errors.New("push error")
	err = mock.PushTag("v1.0.1")
	if err == nil || err.Error() != "push error" {
		t.Errorf("expected 'push error', got %v", err)
	}
}

func TestMockGitCommitReader(t *testing.T) {
	mock := NewMockGitCommitReader()

	t.Run("get commits", func(t *testing.T) {
		mock.Commits = []string{"abc123", "def456"}
		commits, err := mock.GetCommits("v1.0.0", "HEAD")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(commits) != 2 {
			t.Errorf("expected 2 commits, got %d", len(commits))
		}
	})

	t.Run("get commits with error", func(t *testing.T) {
		mock.GetCommitsErr = errors.New("git error")
		_, err := mock.GetCommits("v1.0.0", "HEAD")
		if err == nil || err.Error() != "git error" {
			t.Errorf("expected 'git error', got %v", err)
		}
		mock.GetCommitsErr = nil
	})
}

func TestMockGitBranchReader(t *testing.T) {
	mock := NewMockGitBranchReader()

	t.Run("get current branch", func(t *testing.T) {
		mock.BranchName = "main"
		branch, err := mock.GetCurrentBranch()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "main" {
			t.Errorf("expected 'main', got %s", branch)
		}
	})

	t.Run("get current branch with error", func(t *testing.T) {
		mock.GetCurrentBranchErr = errors.New("branch error")
		_, err := mock.GetCurrentBranch()
		if err == nil || err.Error() != "branch error" {
			t.Errorf("expected 'branch error', got %v", err)
		}
		mock.GetCurrentBranchErr = nil
	})
}
