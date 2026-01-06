package core

import (
	"context"
	"errors"
	"io/fs"
	"testing"
)

func TestMockFileSystem(t *testing.T) {
	mockFS := NewMockFileSystem()
	ctx := context.Background()

	t.Run("write and read file", func(t *testing.T) {
		content := []byte("test content")
		err := mockFS.WriteFile(ctx, "/test/file.txt", content, 0644)
		if err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		data, err := mockFS.ReadFile(ctx, "/test/file.txt")
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}

		if string(data) != string(content) {
			t.Errorf("expected %q, got %q", string(content), string(data))
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		_, err := mockFS.ReadFile(ctx, "/nonexistent")
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("expected fs.ErrNotExist, got %v", err)
		}
	})

	t.Run("stat file", func(t *testing.T) {
		mockFS.SetFile("/stat/test.txt", []byte("hello"))
		info, err := mockFS.Stat(ctx, "/stat/test.txt")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if info.Size() != 5 {
			t.Errorf("expected size 5, got %d", info.Size())
		}
		if info.IsDir() {
			t.Error("expected file, got directory")
		}
	})

	t.Run("mkdir and stat directory", func(t *testing.T) {
		err := mockFS.MkdirAll(ctx, "/test/dir", 0755)
		if err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}

		info, err := mockFS.Stat(ctx, "/test/dir")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected directory, got file")
		}
	})

	t.Run("remove file", func(t *testing.T) {
		mockFS.SetFile("/remove/test.txt", []byte("to be removed"))
		err := mockFS.Remove(ctx, "/remove/test.txt")
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		_, err = mockFS.ReadFile(ctx, "/remove/test.txt")
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("expected fs.ErrNotExist after removal, got %v", err)
		}
	})

	t.Run("error injection", func(t *testing.T) {
		mockFS.ReadErr = errors.New("read error")
		_, err := mockFS.ReadFile(ctx, "/any/path")
		if err == nil || err.Error() != "read error" {
			t.Errorf("expected read error, got %v", err)
		}
		mockFS.ReadErr = nil
	})
}

func TestMockCommandExecutor(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	ctx := context.Background()

	t.Run("run with default success", func(t *testing.T) {
		err := mockExec.Run(ctx, ".", "echo", "hello")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(mockExec.Calls) != 1 {
			t.Errorf("expected 1 call, got %d", len(mockExec.Calls))
		}
		if mockExec.Calls[0].Command != "echo" {
			t.Errorf("expected 'echo' command, got %q", mockExec.Calls[0].Command)
		}
	})

	t.Run("output with set response", func(t *testing.T) {
		mockExec.SetResponse("git describe --tags", "v1.2.3\n")
		output, err := mockExec.Output(ctx, ".", "git", "describe", "--tags")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output != "v1.2.3\n" {
			t.Errorf("expected 'v1.2.3\\n', got %q", output)
		}
	})

	t.Run("error injection", func(t *testing.T) {
		mockExec.SetError("make test", errors.New("test failed"))
		_, err := mockExec.Output(ctx, ".", "make", "test")
		if err == nil || err.Error() != "test failed" {
			t.Errorf("expected 'test failed' error, got %v", err)
		}
	})

	t.Run("default error", func(t *testing.T) {
		mockExec := NewMockCommandExecutor()
		mockExec.DefaultError = errors.New("default error")

		err := mockExec.Run(ctx, ".", "unknown", "command")
		if err == nil || err.Error() != "default error" {
			t.Errorf("expected 'default error', got %v", err)
		}
	})
}

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

func TestOSFileSystem_WriteAndRead(t *testing.T) {
	osFS := NewOSFileSystem()
	ctx := context.Background()
	tmpDir := t.TempDir()

	path := tmpDir + "/test.txt"
	content := []byte("hello world")

	err := osFS.WriteFile(ctx, path, content, 0600)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := osFS.ReadFile(ctx, path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("expected %q, got %q", string(content), string(data))
	}
}

func TestOSFileSystem_Stat(t *testing.T) {
	osFS := NewOSFileSystem()
	ctx := context.Background()
	tmpDir := t.TempDir()

	info, err := osFS.Stat(ctx, tmpDir)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestOSFileSystem_MkdirAll(t *testing.T) {
	osFS := NewOSFileSystem()
	ctx := context.Background()
	tmpDir := t.TempDir()

	path := tmpDir + "/a/b/c"
	err := osFS.MkdirAll(ctx, path, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	info, err := osFS.Stat(ctx, path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestOSFileSystem_Remove(t *testing.T) {
	osFS := NewOSFileSystem()
	ctx := context.Background()
	tmpDir := t.TempDir()

	path := tmpDir + "/to-remove.txt"
	err := osFS.WriteFile(ctx, path, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = osFS.Remove(ctx, path)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	_, err = osFS.ReadFile(ctx, path)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expected fs.ErrNotExist after removal, got %v", err)
	}
}

func TestOSFileSystem_RemoveAll(t *testing.T) {
	osFS := NewOSFileSystem()
	ctx := context.Background()
	tmpDir := t.TempDir()

	path := tmpDir + "/remove-all-test"
	err := osFS.MkdirAll(ctx, path+"/subdir", 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	err = osFS.WriteFile(ctx, path+"/file.txt", []byte("test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = osFS.RemoveAll(ctx, path)
	if err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	_, err = osFS.Stat(ctx, path)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expected fs.ErrNotExist after RemoveAll, got %v", err)
	}
}

func TestOSFileSystem_ReadDir(t *testing.T) {
	osFS := NewOSFileSystem()
	ctx := context.Background()
	tmpDir := t.TempDir()

	dirPath := tmpDir + "/read-dir-test"
	err := osFS.MkdirAll(ctx, dirPath, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	err = osFS.WriteFile(ctx, dirPath+"/file1.txt", []byte("test1"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = osFS.WriteFile(ctx, dirPath+"/file2.txt", []byte("test2"), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	entries, err := osFS.ReadDir(ctx, dirPath)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestEnsureParentDir(t *testing.T) {
	mockFS := NewMockFileSystem()
	ctx := context.Background()

	t.Run("creates parent directory", func(t *testing.T) {
		filePath := "/a/b/c/file.txt"
		err := EnsureParentDir(ctx, mockFS, filePath, 0755)
		if err != nil {
			t.Fatalf("EnsureParentDir failed: %v", err)
		}

		// Verify parent directory was created
		info, err := mockFS.Stat(ctx, "/a/b/c")
		if err != nil {
			t.Fatalf("parent directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected parent to be a directory")
		}
	})

	t.Run("handles root path", func(t *testing.T) {
		err := EnsureParentDir(ctx, mockFS, "/file.txt", 0755)
		if err != nil {
			t.Fatalf("EnsureParentDir failed for root: %v", err)
		}
	})
}

func TestMockFileSystem_ErrorInjection(t *testing.T) {
	mockFS := NewMockFileSystem()
	ctx := context.Background()

	t.Run("write error", func(t *testing.T) {
		mockFS.WriteErr = errors.New("write error")
		err := mockFS.WriteFile(ctx, "/test", []byte("data"), 0644)
		if err == nil || err.Error() != "write error" {
			t.Errorf("expected write error, got %v", err)
		}
		mockFS.WriteErr = nil
	})

	t.Run("stat error", func(t *testing.T) {
		mockFS.StatErr = errors.New("stat error")
		_, err := mockFS.Stat(ctx, "/test")
		if err == nil || err.Error() != "stat error" {
			t.Errorf("expected stat error, got %v", err)
		}
		mockFS.StatErr = nil
	})

	t.Run("mkdir error", func(t *testing.T) {
		mockFS.MkdirErr = errors.New("mkdir error")
		err := mockFS.MkdirAll(ctx, "/test", 0755)
		if err == nil || err.Error() != "mkdir error" {
			t.Errorf("expected mkdir error, got %v", err)
		}
		mockFS.MkdirErr = nil
	})

	t.Run("remove error", func(t *testing.T) {
		mockFS.RemoveErr = errors.New("remove error")
		err := mockFS.Remove(ctx, "/test")
		if err == nil || err.Error() != "remove error" {
			t.Errorf("expected remove error, got %v", err)
		}
		mockFS.RemoveErr = nil
	})
}

func TestMockFileSystem_GetFile(t *testing.T) {
	mockFS := NewMockFileSystem()

	t.Run("get existing file", func(t *testing.T) {
		content := []byte("test content")
		mockFS.SetFile("/test/file.txt", content)

		data, ok := mockFS.GetFile("/test/file.txt")
		if !ok {
			t.Fatal("GetFile() should return true for existing file")
		}
		if string(data) != string(content) {
			t.Errorf("expected %q, got %q", string(content), string(data))
		}
	})

	t.Run("get non-existent file", func(t *testing.T) {
		_, ok := mockFS.GetFile("/nonexistent")
		if ok {
			t.Error("GetFile() should return false for non-existent file")
		}
	})
}

func TestMockFileSystem_RemoveAll(t *testing.T) {
	mockFS := NewMockFileSystem()
	ctx := context.Background()

	t.Run("remove directory", func(t *testing.T) {
		mockFS.dirs["/test-dir"] = true

		err := mockFS.RemoveAll(ctx, "/test-dir")
		if err != nil {
			t.Fatalf("RemoveAll failed: %v", err)
		}

		_, err = mockFS.Stat(ctx, "/test-dir")
		if !errors.Is(err, fs.ErrNotExist) {
			t.Error("directory should be removed")
		}
	})

	t.Run("remove file", func(t *testing.T) {
		mockFS.SetFile("/test-file.txt", []byte("content"))

		err := mockFS.RemoveAll(ctx, "/test-file.txt")
		if err != nil {
			t.Fatalf("RemoveAll failed: %v", err)
		}

		_, ok := mockFS.GetFile("/test-file.txt")
		if ok {
			t.Error("file should be removed")
		}
	})

	t.Run("with remove error", func(t *testing.T) {
		mockFS.RemoveErr = errors.New("remove failed")
		err := mockFS.RemoveAll(ctx, "/anything")
		if err == nil || err.Error() != "remove failed" {
			t.Errorf("expected remove error, got %v", err)
		}
		mockFS.RemoveErr = nil
	})
}

func TestMockFileSystem_ReadDir(t *testing.T) {
	mockFS := NewMockFileSystem()
	ctx := context.Background()

	// Create some files and directories
	mockFS.SetFile("/test/file1.txt", []byte("content1"))
	mockFS.SetFile("/test/file2.txt", []byte("content2"))
	mockFS.dirs["/test/subdir"] = true
	mockFS.SetFile("/other/file.txt", []byte("other"))

	entries, err := mockFS.ReadDir(ctx, "/test")
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// Check entries
	hasFile1 := false
	hasFile2 := false
	hasSubdir := false

	for _, entry := range entries {
		switch entry.Name() {
		case "file1.txt":
			hasFile1 = true
			if entry.IsDir() {
				t.Error("file1.txt should not be a directory")
			}
		case "file2.txt":
			hasFile2 = true
		case "subdir":
			hasSubdir = true
			if !entry.IsDir() {
				t.Error("subdir should be a directory")
			}
		}
	}

	if !hasFile1 || !hasFile2 || !hasSubdir {
		t.Error("missing expected entries")
	}
}

func TestMockFileInfo_Methods(t *testing.T) {
	info := &mockFileInfo{
		name:  "test.txt",
		size:  100,
		isDir: false,
	}

	if info.Name() != "test.txt" {
		t.Errorf("Name() = %q, want %q", info.Name(), "test.txt")
	}

	if info.Size() != 100 {
		t.Errorf("Size() = %d, want %d", info.Size(), 100)
	}

	if info.Mode() != 0644 {
		t.Errorf("Mode() = %v, want 0644", info.Mode())
	}

	if info.ModTime().IsZero() {
		t.Error("ModTime() should not be zero")
	}

	if info.IsDir() {
		t.Error("IsDir() should return false")
	}

	if info.Sys() != nil {
		t.Error("Sys() should return nil")
	}
}

func TestMockDirEntry_Methods(t *testing.T) {
	entry := &mockDirEntry{
		name:  "test-dir",
		isDir: true,
	}

	if entry.Name() != "test-dir" {
		t.Errorf("Name() = %q, want %q", entry.Name(), "test-dir")
	}

	if !entry.IsDir() {
		t.Error("IsDir() should return true")
	}

	if entry.Type() != 0644 {
		t.Errorf("Type() = %v, want 0644", entry.Type())
	}

	info, err := entry.Info()
	if err != nil {
		t.Fatalf("Info() failed: %v", err)
	}
	if info == nil {
		t.Fatal("Info() returned nil")
	}
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

func TestMockCommandExecutor_DefaultOutput(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	ctx := context.Background()

	mockExec.DefaultOutput = "default response"
	output, err := mockExec.Output(ctx, ".", "unknown", "command")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "default response" {
		t.Errorf("expected 'default response', got %q", output)
	}
}

func TestMockCommandExecutor_Run_CommandKey(t *testing.T) {
	mockExec := NewMockCommandExecutor()
	ctx := context.Background()

	expectedErr := errors.New("specific error")
	mockExec.SetError("git status", expectedErr)

	err := mockExec.Run(ctx, ".", "git", "status")
	if err != expectedErr {
		t.Errorf("expected specific error, got %v", err)
	}

	// Verify call was recorded
	if len(mockExec.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mockExec.Calls))
	}
	if mockExec.Calls[0].Command != "git" {
		t.Errorf("command = %q, want %q", mockExec.Calls[0].Command, "git")
	}
	if len(mockExec.Calls[0].Args) != 1 || mockExec.Calls[0].Args[0] != "status" {
		t.Errorf("args = %v, want [status]", mockExec.Calls[0].Args)
	}
}
