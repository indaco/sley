package core

import (
	"errors"
	"testing"
)

func TestMockMarshaler(t *testing.T) {
	mock := NewMockMarshaler()

	t.Run("marshal", func(t *testing.T) {
		mock.MarshalOutput = []byte(`{"key":"value"}`)
		data, err := mock.Marshal(map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != `{"key":"value"}` {
			t.Errorf("expected {\"key\":\"value\"}, got %s", string(data))
		}
	})

	t.Run("marshal with error", func(t *testing.T) {
		mock.MarshalErr = errors.New("marshal error")
		_, err := mock.Marshal(nil)
		if err == nil || err.Error() != "marshal error" {
			t.Errorf("expected 'marshal error', got %v", err)
		}
		mock.MarshalErr = nil
	})

	t.Run("unmarshal", func(t *testing.T) {
		mock.UnmarshalErr = nil
		err := mock.Unmarshal([]byte(`{}`), &struct{}{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unmarshal with error", func(t *testing.T) {
		mock.UnmarshalErr = errors.New("unmarshal error")
		err := mock.Unmarshal([]byte(`{}`), &struct{}{})
		if err == nil || err.Error() != "unmarshal error" {
			t.Errorf("expected 'unmarshal error', got %v", err)
		}
	})
}

func TestMockUserDirProvider(t *testing.T) {
	mock := NewMockUserDirProvider()

	t.Run("home dir", func(t *testing.T) {
		mock.HomeDirPath = "/home/testuser"
		home, err := mock.HomeDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if home != "/home/testuser" {
			t.Errorf("expected '/home/testuser', got %s", home)
		}
	})

	t.Run("home dir with error", func(t *testing.T) {
		mock.HomeDirErr = errors.New("home error")
		_, err := mock.HomeDir()
		if err == nil || err.Error() != "home error" {
			t.Errorf("expected 'home error', got %v", err)
		}
	})
}

func TestMockFileCopier(t *testing.T) {
	mock := NewMockFileCopier()

	t.Run("copy dir", func(t *testing.T) {
		err := mock.CopyDir("/src", "/dst")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(mock.CopyDirCalls) != 1 {
			t.Errorf("expected 1 copied dir, got %d", len(mock.CopyDirCalls))
		}
		if mock.CopyDirCalls[0].Src != "/src" || mock.CopyDirCalls[0].Dst != "/dst" {
			t.Errorf("unexpected copied dir: %+v", mock.CopyDirCalls[0])
		}
	})

	t.Run("copy dir with error", func(t *testing.T) {
		mock.CopyDirErr = errors.New("copy dir error")
		err := mock.CopyDir("/src", "/dst")
		if err == nil || err.Error() != "copy dir error" {
			t.Errorf("expected 'copy dir error', got %v", err)
		}
		mock.CopyDirErr = nil
	})

	t.Run("copy file", func(t *testing.T) {
		mock.CopyFileCalls = nil
		err := mock.CopyFile("/src/file.txt", "/dst/file.txt", 0644)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(mock.CopyFileCalls) != 1 {
			t.Errorf("expected 1 copied file, got %d", len(mock.CopyFileCalls))
		}
	})

	t.Run("copy file with error", func(t *testing.T) {
		mock.CopyFileErr = errors.New("copy file error")
		err := mock.CopyFile("/src/file.txt", "/dst/file.txt", 0644)
		if err == nil || err.Error() != "copy file error" {
			t.Errorf("expected 'copy file error', got %v", err)
		}
	})
}
