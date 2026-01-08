package extensionmgr

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCopyDir_Success(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// create nested dir and file
	subDir := filepath.Join(src, "nested")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := []byte("hello")
	if err := os.WriteFile(filepath.Join(subDir, "file.txt"), content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyDirFn(src, dst); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	copied := filepath.Join(dst, "nested", "file.txt")
	data, err := os.ReadFile(copied)
	if err != nil {
		t.Fatalf("file not copied: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected file content 'hello', got: %q", string(data))
	}
}

func TestCopyDir_FailsOnWalk(t *testing.T) {
	// Create a broken source path that doesn't exist
	err := copyDirFn("non-existent-src", t.TempDir())
	if err == nil {
		t.Fatal("expected error due to non-existent source directory, got nil")
	}
}

func TestCopyDir_SkipsExcludedFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	// Create a fake .git directory and a .DS_Store file in source
	_ = os.Mkdir(filepath.Join(src, ".git"), 0755)
	if err := os.WriteFile(filepath.Join(src, ".DS_Store"), []byte("mac metadata"), 0644); err != nil {
		t.Fatal(err)
	}

	// Also include a normal file to ensure it *is* copied
	helloPath := filepath.Join(src, "hello.txt")
	if err := os.WriteFile(helloPath, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyDirFn(src, dst); err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}

	// .git directory should NOT exist in destination
	if _, err := os.Stat(filepath.Join(dst, ".git")); !os.IsNotExist(err) {
		t.Errorf(".git directory should be skipped, but it exists")
	}

	// .DS_Store should also be skipped
	if _, err := os.Stat(filepath.Join(dst, ".DS_Store")); !os.IsNotExist(err) {
		t.Errorf(".DS_Store file should be skipped, but it exists")
	}

	// hello.txt SHOULD exist
	if _, err := os.Stat(filepath.Join(dst, "hello.txt")); err != nil {
		t.Errorf("hello.txt should be copied, but got error: %v", err)
	}
}

func TestCopyDir_FailsOnRel(t *testing.T) {
	// Create a custom OSFileCopier with mocked relFn
	copier := &OSFileCopier{
		walkFn: filepath.Walk,
		relFn: func(basepath, targpath string) (string, error) {
			return "", errors.New("mock Rel error")
		},
		openSrcFile: os.Open,
		openDstFile: os.OpenFile,
		copyFn:      io.Copy,
	}

	src := t.TempDir()
	file := filepath.Join(src, "file.txt")
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	dst := t.TempDir()
	err := copier.CopyDir(src, dst)

	if err == nil || !strings.Contains(err.Error(), "mock Rel error") {
		t.Fatalf("expected mock Rel error, got: %v", err)
	}
}

func TestCopyDir_FailsOnOpenSource(t *testing.T) {
	tmp := t.TempDir()
	srcFile := filepath.Join(tmp, "readonly.txt")
	dstDir := filepath.Join(tmp, "dst")

	_ = os.WriteFile(srcFile, []byte("test"), 0000)   // no read permissions
	t.Cleanup(func() { _ = os.Chmod(srcFile, 0644) }) // cleanup permissions

	srcDir := filepath.Join(tmp, "src")
	_ = os.MkdirAll(srcDir, 0755)
	_ = os.Rename(srcFile, filepath.Join(srcDir, "readonly.txt"))

	err := copyDirFn(srcDir, dstDir)
	if err == nil {
		t.Fatal("expected error opening unreadable source file, got nil")
	}
}

func TestCopyDir_FailsOnOpenTarget(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src")
	dstDir := filepath.Join(tmp, "dst")

	_ = os.MkdirAll(srcDir, 0755)
	_ = os.MkdirAll(dstDir, 0555) // no write permission

	srcFile := filepath.Join(srcDir, "file.txt")
	_ = os.WriteFile(srcFile, []byte("test"), 0644)

	t.Cleanup(func() { _ = os.Chmod(dstDir, 0755) })

	err := copyDirFn(srcDir, dstDir)
	if err == nil {
		t.Fatal("expected error opening unwritable target file, got nil")
	}
}

func TestCopyFile_FailsOnCopy(t *testing.T) {
	// Create a custom OSFileCopier with mocked copyFn
	copier := &OSFileCopier{
		walkFn:      filepath.Walk,
		relFn:       filepath.Rel,
		openSrcFile: os.Open,
		openDstFile: os.OpenFile,
		copyFn: func(dst io.Writer, src io.Reader) (int64, error) {
			return 0, errors.New("mock copy failure")
		},
	}

	srcPath := filepath.Join(t.TempDir(), "source.txt")
	dstPath := filepath.Join(t.TempDir(), "dest.txt")

	if err := os.WriteFile(srcPath, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	err := copier.CopyFile(srcPath, dstPath, 0644)
	if err == nil || !strings.Contains(err.Error(), "mock copy failure") {
		t.Fatalf("expected mock copy error, got: %v", err)
	}
}

// CopyFileTestHelper is used in tests to inject a fake reader (for io.Copy failure scenarios).
func CopyFileTestHelper(src io.Reader, dst string, perm os.FileMode) error {
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		return err
	}

	return nil
}

func TestShouldSkipEntry(t *testing.T) {
	tests := []struct {
		name      string
		fileName  string
		isDir     bool
		wantSkipF bool
		wantSkipD bool
	}{
		{"macOS metadata file", ".DS_Store", false, true, false},
		{"git directory", ".git", true, false, true},
		{"node_modules directory", "node_modules", true, false, true},
		{"regular file", "main.go", false, false, false},
		{"dotfile", ".env", false, false, false},
		{"unknown dir", "config", true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := fakeFileInfo{name: tt.fileName, dir: tt.isDir}
			skipF, skipD := shouldSkipEntry(info)
			if skipF != tt.wantSkipF || skipD != tt.wantSkipD {
				t.Errorf("shouldSkipEntry(%q, dir=%v) = (%v, %v), want (%v, %v)",
					tt.fileName, tt.isDir, skipF, skipD, tt.wantSkipF, tt.wantSkipD)
			}
		})
	}
}

// fakeFileInfo is a minimal os.FileInfo stub for testing
type fakeFileInfo struct {
	name string
	dir  bool
}

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() os.FileMode  { return 0 }
func (f fakeFileInfo) ModTime() time.Time { return time.Now() }
func (f fakeFileInfo) IsDir() bool        { return f.dir }
func (f fakeFileInfo) Sys() any           { return nil }
