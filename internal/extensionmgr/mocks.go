package extensionmgr

import (
	"errors"
	"os"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/extensions"
)

// MockYAMLMarshaler is a mock implementation of YAMLMarshaler for testing
type MockYAMLMarshaler struct {
	MarshalFunc func(v any) ([]byte, error)
}

func (m *MockYAMLMarshaler) Marshal(v any) ([]byte, error) {
	if m.MarshalFunc != nil {
		return m.MarshalFunc(v)
	}
	return nil, errors.New("marshal not implemented")
}

// MockConfigUpdater is a mock implementation of ConfigUpdater for testing
type MockConfigUpdater struct {
	AddExtensionFunc func(path string, extension config.ExtensionConfig) error
}

func (m *MockConfigUpdater) AddExtension(path string, extension config.ExtensionConfig) error {
	if m.AddExtensionFunc != nil {
		return m.AddExtensionFunc(path, extension)
	}
	return nil
}

// MockManifestLoader is a mock implementation of ManifestLoader for testing
type MockManifestLoader struct {
	LoadFunc func(path string) (*extensions.ExtensionManifest, error)
}

func (m *MockManifestLoader) Load(path string) (*extensions.ExtensionManifest, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc(path)
	}
	return nil, errors.New("load not implemented")
}

// MockHomeDirectory is a mock implementation of HomeDirectory for testing
type MockHomeDirectory struct {
	GetFunc func() (string, error)
}

func (m *MockHomeDirectory) Get() (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc()
	}
	return "", errors.New("get not implemented")
}

// MockFileCopier is a mock implementation of core.FileCopier for testing
type MockFileCopier struct {
	CopyDirFunc  func(src, dst string) error
	CopyFileFunc func(src, dst string, perm os.FileMode) error
}

func (m *MockFileCopier) CopyDir(src, dst string) error {
	if m.CopyDirFunc != nil {
		return m.CopyDirFunc(src, dst)
	}
	return nil
}

func (m *MockFileCopier) CopyFile(src, dst string, perm os.FileMode) error {
	if m.CopyFileFunc != nil {
		return m.CopyFileFunc(src, dst, perm)
	}
	return nil
}
