package core

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MockFileSystem is a thread-safe in-memory file system for testing.
type MockFileSystem struct {
	mu    sync.RWMutex
	files map[string][]byte
	perms map[string]fs.FileMode
	dirs  map[string]bool

	// Error injection
	ReadErr   error
	WriteErr  error
	StatErr   error
	MkdirErr  error
	RemoveErr error
}

// NewMockFileSystem creates a new MockFileSystem.
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
		perms: make(map[string]fs.FileMode),
		dirs:  make(map[string]bool),
	}
}

func (m *MockFileSystem) ReadFile(ctx context.Context, path string) ([]byte, error) {
	// Check if context is cancelled
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.ReadErr != nil {
		return nil, m.ReadErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return data, nil
}

func (m *MockFileSystem) WriteFile(ctx context.Context, path string, data []byte, perm fs.FileMode) error {
	// Check if context is cancelled
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.WriteErr != nil {
		return m.WriteErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[path] = data
	m.perms[path] = perm
	return nil
}

func (m *MockFileSystem) Stat(ctx context.Context, path string) (fs.FileInfo, error) {
	// Check if context is cancelled
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if m.StatErr != nil {
		return nil, m.StatErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	if data, ok := m.files[path]; ok {
		return &mockFileInfo{name: path, size: int64(len(data)), isDir: false}, nil
	}
	if m.dirs[path] {
		return &mockFileInfo{name: path, size: 0, isDir: true}, nil
	}
	return nil, fs.ErrNotExist
}

func (m *MockFileSystem) MkdirAll(ctx context.Context, path string, perm fs.FileMode) error {
	// Check if context is cancelled
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.MkdirErr != nil {
		return m.MkdirErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dirs[path] = true
	return nil
}

func (m *MockFileSystem) Remove(ctx context.Context, path string) error {
	// Check if context is cancelled
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.RemoveErr != nil {
		return m.RemoveErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.files, path)
	delete(m.dirs, path)
	return nil
}

func (m *MockFileSystem) RemoveAll(ctx context.Context, path string) error {
	return m.Remove(ctx, path)
}

func (m *MockFileSystem) ReadDir(ctx context.Context, path string) ([]fs.DirEntry, error) {
	// Check if context is cancelled
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Build list of entries in this directory
	var entries []fs.DirEntry

	// Find all files and dirs under this path
	for filePath := range m.files {
		if filepath.Dir(filePath) == path {
			name := filepath.Base(filePath)
			entries = append(entries, &mockDirEntry{
				name:  name,
				isDir: false,
			})
		}
	}

	for dirPath := range m.dirs {
		if filepath.Dir(dirPath) == path {
			name := filepath.Base(dirPath)
			entries = append(entries, &mockDirEntry{
				name:  name,
				isDir: true,
			})
		}
	}

	return entries, nil
}

// SetFile sets a file's content (for test setup).
func (m *MockFileSystem) SetFile(path string, content []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[path] = content
}

// GetFile returns a file's content (for test assertions).
func (m *MockFileSystem) GetFile(path string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.files[path]
	return data, ok
}

type mockFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() fs.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() any           { return nil }

type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string      { return m.name }
func (m *mockDirEntry) IsDir() bool       { return m.isDir }
func (m *mockDirEntry) Type() fs.FileMode { return 0644 }
func (m *mockDirEntry) Info() (fs.FileInfo, error) {
	return &mockFileInfo{name: m.name, isDir: m.isDir}, nil
}

// MockCommandExecutor is a mock command executor for testing.
type MockCommandExecutor struct {
	mu sync.Mutex

	// Responses maps "command args..." to output
	Responses map[string]string

	// Errors maps "command args..." to error
	Errors map[string]error

	// Calls records all command invocations
	Calls []CommandCall

	// DefaultOutput is returned if no specific response is set
	DefaultOutput string

	// DefaultError is returned if no specific error is set
	DefaultError error
}

// CommandCall records a command invocation.
type CommandCall struct {
	Dir     string
	Command string
	Args    []string
}

// NewMockCommandExecutor creates a new MockCommandExecutor.
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		Responses: make(map[string]string),
		Errors:    make(map[string]error),
	}
}

func (m *MockCommandExecutor) Run(ctx context.Context, dir string, command string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, CommandCall{Dir: dir, Command: command, Args: args})

	var key strings.Builder
	key.WriteString(command)
	for _, arg := range args {
		key.WriteString(" " + arg)
	}

	if err, ok := m.Errors[key.String()]; ok {
		return err
	}
	return m.DefaultError
}

func (m *MockCommandExecutor) Output(ctx context.Context, dir string, command string, args ...string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, CommandCall{Dir: dir, Command: command, Args: args})

	key := command
	for _, arg := range args {
		key += " " + arg
	}

	if err, ok := m.Errors[key]; ok {
		return "", err
	}

	if output, ok := m.Responses[key]; ok {
		return output, nil
	}

	if m.DefaultError != nil {
		return "", m.DefaultError
	}

	return m.DefaultOutput, nil
}

// SetResponse sets the response for a command.
func (m *MockCommandExecutor) SetResponse(command string, output string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Responses[command] = output
}

// SetError sets an error for a command.
func (m *MockCommandExecutor) SetError(command string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors[command] = err
}

// MockGitClient is a mock git client for testing.
type MockGitClient struct {
	mu sync.Mutex

	TagOutput    string
	TagError     error
	CloneError   error
	PullError    error
	IsValidRepos map[string]bool
}

// NewMockGitClient creates a new MockGitClient.
func NewMockGitClient() *MockGitClient {
	return &MockGitClient{
		IsValidRepos: make(map[string]bool),
	}
}

func (m *MockGitClient) DescribeTags(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.TagError != nil {
		return "", m.TagError
	}
	return m.TagOutput, nil
}

func (m *MockGitClient) Clone(ctx context.Context, url string, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.CloneError != nil {
		return m.CloneError
	}
	m.IsValidRepos[path] = true
	return nil
}

func (m *MockGitClient) Pull(ctx context.Context, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.PullError
}

func (m *MockGitClient) IsValidRepo(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.IsValidRepos[path]
}

// MockGitTagOperations is a mock git tag operations for testing.
type MockGitTagOperations struct {
	mu sync.Mutex

	// Error responses for each operation
	CreateAnnotatedTagErr   error
	CreateLightweightTagErr error
	TagExistsErr            error
	GetLatestTagErr         error
	PushTagErr              error

	// Response values
	TagExistsResult  bool
	GetLatestTagName string

	// Call tracking
	CreatedTags []string
	PushedTags  []string
}

// NewMockGitTagOperations creates a new MockGitTagOperations.
func NewMockGitTagOperations() *MockGitTagOperations {
	return &MockGitTagOperations{
		CreatedTags: make([]string, 0),
		PushedTags:  make([]string, 0),
	}
}

func (m *MockGitTagOperations) CreateAnnotatedTag(name, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.CreateAnnotatedTagErr != nil {
		return m.CreateAnnotatedTagErr
	}
	m.CreatedTags = append(m.CreatedTags, name)
	return nil
}

func (m *MockGitTagOperations) CreateLightweightTag(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.CreateLightweightTagErr != nil {
		return m.CreateLightweightTagErr
	}
	m.CreatedTags = append(m.CreatedTags, name)
	return nil
}

func (m *MockGitTagOperations) TagExists(name string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.TagExistsErr != nil {
		return false, m.TagExistsErr
	}
	return m.TagExistsResult, nil
}

func (m *MockGitTagOperations) GetLatestTag() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.GetLatestTagErr != nil {
		return "", m.GetLatestTagErr
	}
	return m.GetLatestTagName, nil
}

func (m *MockGitTagOperations) PushTag(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.PushTagErr != nil {
		return m.PushTagErr
	}
	m.PushedTags = append(m.PushedTags, name)
	return nil
}

// MockGitCommitReader is a mock git commit reader for testing.
type MockGitCommitReader struct {
	mu sync.Mutex

	// GetCommitsErr is the error to return from GetCommits.
	GetCommitsErr error

	// Commits is the list of commits to return.
	Commits []string

	// Calls tracks calls to GetCommits.
	Calls []struct{ Since, Until string }
}

// NewMockGitCommitReader creates a new MockGitCommitReader.
func NewMockGitCommitReader() *MockGitCommitReader {
	return &MockGitCommitReader{
		Commits: make([]string, 0),
		Calls:   make([]struct{ Since, Until string }, 0),
	}
}

func (m *MockGitCommitReader) GetCommits(since, until string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, struct{ Since, Until string }{since, until})
	if m.GetCommitsErr != nil {
		return nil, m.GetCommitsErr
	}
	return m.Commits, nil
}

// MockGitBranchReader is a mock git branch reader for testing.
type MockGitBranchReader struct {
	mu sync.Mutex

	// GetCurrentBranchErr is the error to return.
	GetCurrentBranchErr error

	// BranchName is the branch name to return.
	BranchName string
}

// NewMockGitBranchReader creates a new MockGitBranchReader.
func NewMockGitBranchReader() *MockGitBranchReader {
	return &MockGitBranchReader{
		BranchName: "main",
	}
}

func (m *MockGitBranchReader) GetCurrentBranch(_ context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.GetCurrentBranchErr != nil {
		return "", m.GetCurrentBranchErr
	}
	return m.BranchName, nil
}

// MockMarshaler is a mock marshaler for testing.
type MockMarshaler struct {
	mu sync.Mutex

	// MarshalErr is the error to return from Marshal.
	MarshalErr error

	// MarshalOutput is the output to return from Marshal.
	MarshalOutput []byte

	// UnmarshalErr is the error to return from Unmarshal.
	UnmarshalErr error

	// Calls tracks marshal calls.
	MarshalCalls []any
}

// NewMockMarshaler creates a new MockMarshaler.
func NewMockMarshaler() *MockMarshaler {
	return &MockMarshaler{
		MarshalCalls: make([]any, 0),
	}
}

func (m *MockMarshaler) Marshal(v any) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MarshalCalls = append(m.MarshalCalls, v)
	if m.MarshalErr != nil {
		return nil, m.MarshalErr
	}
	return m.MarshalOutput, nil
}

func (m *MockMarshaler) Unmarshal(data []byte, v any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.UnmarshalErr
}

// MockUserDirProvider is a mock user directory provider for testing.
type MockUserDirProvider struct {
	mu sync.Mutex

	// HomeDirErr is the error to return from HomeDir.
	HomeDirErr error

	// HomeDirPath is the path to return from HomeDir.
	HomeDirPath string
}

// NewMockUserDirProvider creates a new MockUserDirProvider.
func NewMockUserDirProvider() *MockUserDirProvider {
	return &MockUserDirProvider{
		HomeDirPath: "/home/testuser",
	}
}

func (m *MockUserDirProvider) HomeDir() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.HomeDirErr != nil {
		return "", m.HomeDirErr
	}
	return m.HomeDirPath, nil
}

// MockFileCopier is a mock file copier for testing.
type MockFileCopier struct {
	mu sync.Mutex

	// CopyDirErr is the error to return from CopyDir.
	CopyDirErr error

	// CopyFileErr is the error to return from CopyFile.
	CopyFileErr error

	// CopyDirCalls tracks CopyDir calls.
	CopyDirCalls []struct{ Src, Dst string }

	// CopyFileCalls tracks CopyFile calls.
	CopyFileCalls []struct {
		Src, Dst string
		Perm     FileMode
	}
}

// NewMockFileCopier creates a new MockFileCopier.
func NewMockFileCopier() *MockFileCopier {
	return &MockFileCopier{
		CopyDirCalls: make([]struct{ Src, Dst string }, 0),
		CopyFileCalls: make([]struct {
			Src, Dst string
			Perm     FileMode
		}, 0),
	}
}

func (m *MockFileCopier) CopyDir(src, dst string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CopyDirCalls = append(m.CopyDirCalls, struct{ Src, Dst string }{src, dst})
	return m.CopyDirErr
}

func (m *MockFileCopier) CopyFile(src, dst string, perm FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CopyFileCalls = append(m.CopyFileCalls, struct {
		Src, Dst string
		Perm     FileMode
	}{src, dst, perm})
	return m.CopyFileErr
}
