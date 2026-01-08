package tagmanager

import (
	"os/exec"
	"testing"
)

// createTestGitTagOps creates an OSGitTagOperations with a custom exec.Command for testing.
func createTestGitTagOps(mockExec func(name string, args ...string) *exec.Cmd) *OSGitTagOperations {
	return &OSGitTagOperations{
		execCommand: mockExec,
	}
}

func TestOSGitTagOperations_CreateAnnotatedTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			// Verify correct arguments
			if name != "git" {
				t.Errorf("expected git command, got %s", name)
			}
			if len(args) < 4 || args[0] != "tag" || args[1] != "-a" || args[2] != "v1.0.0" {
				t.Errorf("unexpected args: %v", args)
			}
			return exec.Command("true")
		})

		err := ops.CreateAnnotatedTag("v1.0.0", "Release 1.0.0")
		if err != nil {
			t.Errorf("CreateAnnotatedTag() error = %v", err)
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'tag already exists' >&2 && exit 1")
		})

		err := ops.CreateAnnotatedTag("v1.0.0", "Release 1.0.0")
		if err == nil {
			t.Error("CreateAnnotatedTag() expected error")
		}
		if err.Error() == "" {
			t.Error("expected error message")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		})

		err := ops.CreateAnnotatedTag("v1.0.0", "Release 1.0.0")
		if err == nil {
			t.Error("CreateAnnotatedTag() expected error")
		}
	})
}

func TestOSGitTagOperations_CreateLightweightTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			if name != "git" || len(args) < 2 || args[0] != "tag" {
				t.Errorf("unexpected command: %s %v", name, args)
			}
			return exec.Command("true")
		})

		err := ops.CreateLightweightTag("v1.0.0")
		if err != nil {
			t.Errorf("CreateLightweightTag() error = %v", err)
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'error' >&2 && exit 1")
		})

		err := ops.CreateLightweightTag("v1.0.0")
		if err == nil {
			t.Error("CreateLightweightTag() expected error")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		})

		err := ops.CreateLightweightTag("v1.0.0")
		if err == nil {
			t.Error("CreateLightweightTag() expected error")
		}
	})
}

func TestOSGitTagOperations_TagExists(t *testing.T) {
	t.Run("tag exists", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "v1.0.0")
		})

		exists, err := ops.TagExists("v1.0.0")
		if err != nil {
			t.Errorf("TagExists() error = %v", err)
		}
		if !exists {
			t.Error("TagExists() expected true")
		}
	})

	t.Run("tag does not exist", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "")
		})

		exists, err := ops.TagExists("v1.0.0")
		if err != nil {
			t.Errorf("TagExists() error = %v", err)
		}
		if exists {
			t.Error("TagExists() expected false")
		}
	})

	t.Run("error", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		})

		_, err := ops.TagExists("v1.0.0")
		if err == nil {
			t.Error("TagExists() expected error")
		}
	})
}

func TestOSGitTagOperations_GetLatestTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "v1.2.3")
		})

		tag, err := ops.GetLatestTag()
		if err != nil {
			t.Errorf("GetLatestTag() error = %v", err)
		}
		if tag != "v1.2.3" {
			t.Errorf("GetLatestTag() = %q, want %q", tag, "v1.2.3")
		}
	})

	t.Run("empty output", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "")
		})

		_, err := ops.GetLatestTag()
		if err == nil {
			t.Error("GetLatestTag() expected error for empty output")
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'fatal: No names found' >&2 && exit 128")
		})

		_, err := ops.GetLatestTag()
		if err == nil {
			t.Error("GetLatestTag() expected error")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		})

		_, err := ops.GetLatestTag()
		if err == nil {
			t.Error("GetLatestTag() expected error")
		}
	})
}

func TestOSGitTagOperations_PushTag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			if name != "git" || len(args) < 3 || args[0] != "push" || args[1] != "origin" {
				t.Errorf("unexpected command: %s %v", name, args)
			}
			return exec.Command("true")
		})

		err := ops.PushTag("v1.0.0")
		if err != nil {
			t.Errorf("PushTag() error = %v", err)
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'remote rejected' >&2 && exit 1")
		})

		err := ops.PushTag("v1.0.0")
		if err == nil {
			t.Error("PushTag() expected error")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		ops := createTestGitTagOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		})

		err := ops.PushTag("v1.0.0")
		if err == nil {
			t.Error("PushTag() expected error")
		}
	})
}

func TestListTags(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	t.Run("list all tags", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("printf", "v1.0.0\nv1.1.0\nv2.0.0")
		}

		tags, err := ListTags("")
		if err != nil {
			t.Errorf("ListTags() error = %v", err)
		}
		if len(tags) != 3 {
			t.Errorf("ListTags() returned %d tags, want 3", len(tags))
		}
	})

	t.Run("list with pattern", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			// Verify pattern is passed
			if len(args) < 3 || args[2] != "v1.*" {
				t.Errorf("expected pattern v1.*, got args: %v", args)
			}
			return exec.Command("printf", "v1.0.0\nv1.1.0")
		}

		tags, err := ListTags("v1.*")
		if err != nil {
			t.Errorf("ListTags() error = %v", err)
		}
		if len(tags) != 2 {
			t.Errorf("ListTags() returned %d tags, want 2", len(tags))
		}
	})

	t.Run("empty result", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "")
		}

		tags, err := ListTags("nonexistent*")
		if err != nil {
			t.Errorf("ListTags() error = %v", err)
		}
		if len(tags) != 0 {
			t.Errorf("ListTags() returned %d tags, want 0", len(tags))
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'git error' >&2 && exit 1")
		}

		_, err := ListTags("")
		if err == nil {
			t.Error("ListTags() expected error")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		}

		_, err := ListTags("")
		if err == nil {
			t.Error("ListTags() expected error")
		}
	})
}

func TestDeleteTag(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	t.Run("success", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			if name != "git" || len(args) < 3 || args[0] != "tag" || args[1] != "-d" {
				t.Errorf("unexpected command: %s %v", name, args)
			}
			return exec.Command("true")
		}

		err := DeleteTag("v1.0.0")
		if err != nil {
			t.Errorf("DeleteTag() error = %v", err)
		}
	})

	t.Run("error with stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'tag not found' >&2 && exit 1")
		}

		err := DeleteTag("v1.0.0")
		if err == nil {
			t.Error("DeleteTag() expected error")
		}
	})

	t.Run("error without stderr", func(t *testing.T) {
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		}

		err := DeleteTag("v1.0.0")
		if err == nil {
			t.Error("DeleteTag() expected error")
		}
	})
}

func TestNewOSGitTagOperations(t *testing.T) {
	ops := NewOSGitTagOperations()
	if ops == nil {
		t.Fatal("NewOSGitTagOperations() returned nil")
	}
	if ops.execCommand == nil {
		t.Error("execCommand should not be nil")
	}
}
