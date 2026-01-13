package releasegate

import (
	"os/exec"
	"testing"
)

// createTestGitOps creates an OSGitOperations with a custom exec.Command for testing.
func createTestGitOps(mockExec func(name string, args ...string) *exec.Cmd) *OSGitOperations {
	return &OSGitOperations{
		execCommand: mockExec,
	}
}

func TestOSGitOperations_IsWorktreeClean(t *testing.T) {
	tests := []struct {
		name      string
		stdout    string
		wantClean bool
		wantErr   bool
	}{
		{
			name:      "clean worktree",
			stdout:    "",
			wantClean: true,
			wantErr:   false,
		},
		{
			name:      "dirty worktree with modified files",
			stdout:    " M file.txt",
			wantClean: false,
			wantErr:   false,
		},
		{
			name:      "dirty worktree with untracked files",
			stdout:    "?? newfile.txt",
			wantClean: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := createTestGitOps(func(name string, args ...string) *exec.Cmd {
				if tt.stdout == "" {
					return exec.Command("true")
				}
				return exec.Command("printf", tt.stdout)
			})

			clean, err := ops.IsWorktreeClean()

			if tt.wantErr {
				if err == nil {
					t.Error("IsWorktreeClean() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("IsWorktreeClean() unexpected error: %v", err)
				}
			}

			if clean != tt.wantClean {
				t.Errorf("IsWorktreeClean() = %v, want %v", clean, tt.wantClean)
			}
		})
	}
}

func TestOSGitOperations_IsWorktreeClean_Error(t *testing.T) {
	t.Run("git error with stderr", func(t *testing.T) {
		ops := createTestGitOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'not a git repo' >&2 && exit 1")
		})

		_, err := ops.IsWorktreeClean()
		if err == nil {
			t.Error("IsWorktreeClean() expected error")
		}
	})

	t.Run("git error without stderr", func(t *testing.T) {
		ops := createTestGitOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		})

		_, err := ops.IsWorktreeClean()
		if err == nil {
			t.Error("IsWorktreeClean() expected error")
		}
	})
}

func TestOSGitOperations_GetCurrentBranch(t *testing.T) {
	tests := []struct {
		name    string
		stdout  string
		want    string
		wantErr bool
	}{
		{
			name:    "main branch",
			stdout:  "main",
			want:    "main",
			wantErr: false,
		},
		{
			name:    "feature branch",
			stdout:  "feature/test",
			want:    "feature/test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := createTestGitOps(func(name string, args ...string) *exec.Cmd {
				return exec.Command("echo", tt.stdout)
			})

			branch, err := ops.GetCurrentBranch()

			if tt.wantErr {
				if err == nil {
					t.Error("GetCurrentBranch() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetCurrentBranch() unexpected error: %v", err)
				}
			}

			if branch != tt.want {
				t.Errorf("GetCurrentBranch() = %q, want %q", branch, tt.want)
			}
		})
	}
}

func TestOSGitOperations_GetCurrentBranch_Error(t *testing.T) {
	t.Run("git error with stderr", func(t *testing.T) {
		ops := createTestGitOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'not a git repo' >&2 && exit 1")
		})

		_, err := ops.GetCurrentBranch()
		if err == nil {
			t.Error("GetCurrentBranch() expected error")
		}
	})

	t.Run("git error without stderr", func(t *testing.T) {
		ops := createTestGitOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		})

		_, err := ops.GetCurrentBranch()
		if err == nil {
			t.Error("GetCurrentBranch() expected error")
		}
	})

	t.Run("empty output", func(t *testing.T) {
		ops := createTestGitOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("true")
		})

		_, err := ops.GetCurrentBranch()
		if err == nil {
			t.Error("GetCurrentBranch() expected error for empty output")
		}
	})
}

func TestOSGitOperations_GetRecentCommits(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		stdout  string
		want    []string
		wantErr bool
	}{
		{
			name:   "multiple commits",
			count:  5,
			stdout: "abc123 feat: add feature\ndef456 fix: resolve bug\nghi789 chore: update deps",
			want: []string{
				"abc123 feat: add feature",
				"def456 fix: resolve bug",
				"ghi789 chore: update deps",
			},
			wantErr: false,
		},
		{
			name:    "single commit",
			count:   1,
			stdout:  "abc123 Initial commit",
			want:    []string{"abc123 Initial commit"},
			wantErr: false,
		},
		{
			name:    "zero count defaults to 10",
			count:   0,
			stdout:  "abc123 feat: add feature",
			want:    []string{"abc123 feat: add feature"},
			wantErr: false,
		},
		{
			name:    "negative count defaults to 10",
			count:   -1,
			stdout:  "abc123 feat: add feature",
			want:    []string{"abc123 feat: add feature"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := createTestGitOps(func(name string, args ...string) *exec.Cmd {
				return exec.Command("printf", tt.stdout)
			})

			commits, err := ops.GetRecentCommits(tt.count)

			if tt.wantErr {
				if err == nil {
					t.Error("GetRecentCommits() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("GetRecentCommits() unexpected error: %v", err)
				}
			}

			if len(commits) != len(tt.want) {
				t.Errorf("GetRecentCommits() returned %d commits, want %d", len(commits), len(tt.want))
				return
			}

			for i, commit := range commits {
				if commit != tt.want[i] {
					t.Errorf("GetRecentCommits()[%d] = %q, want %q", i, commit, tt.want[i])
				}
			}
		})
	}
}

func TestOSGitOperations_GetRecentCommits_Empty(t *testing.T) {
	ops := createTestGitOps(func(name string, args ...string) *exec.Cmd {
		return exec.Command("true")
	})

	commits, err := ops.GetRecentCommits(5)
	if err != nil {
		t.Errorf("GetRecentCommits() unexpected error: %v", err)
	}
	if len(commits) != 0 {
		t.Errorf("GetRecentCommits() returned %d commits, want 0", len(commits))
	}
}

func TestOSGitOperations_GetRecentCommits_Error(t *testing.T) {
	t.Run("git error with stderr", func(t *testing.T) {
		ops := createTestGitOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("sh", "-c", "echo 'not a git repo' >&2 && exit 1")
		})

		_, err := ops.GetRecentCommits(5)
		if err == nil {
			t.Error("GetRecentCommits() expected error")
		}
	})

	t.Run("git error without stderr", func(t *testing.T) {
		ops := createTestGitOps(func(name string, args ...string) *exec.Cmd {
			return exec.Command("false")
		})

		_, err := ops.GetRecentCommits(5)
		if err == nil {
			t.Error("GetRecentCommits() expected error")
		}
	})
}

func TestNewOSGitOperations(t *testing.T) {
	ops := NewOSGitOperations()
	if ops == nil {
		t.Fatal("NewOSGitOperations() returned nil")
	}
	if ops.execCommand == nil {
		t.Error("execCommand should not be nil")
	}
}
