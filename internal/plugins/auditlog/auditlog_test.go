package auditlog

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
)

// MockGitOps implements GitOperations for testing.
type MockGitOps struct {
	AuthorFunc    func() (string, error)
	CommitSHAFunc func() (string, error)
	BranchFunc    func() (string, error)
}

func (m *MockGitOps) GetAuthor() (string, error) {
	if m.AuthorFunc != nil {
		return m.AuthorFunc()
	}
	return "Test User <test@example.com>", nil
}

func (m *MockGitOps) GetCommitSHA() (string, error) {
	if m.CommitSHAFunc != nil {
		return m.CommitSHAFunc()
	}
	return "abc1234567890def", nil
}

func (m *MockGitOps) GetBranch() (string, error) {
	if m.BranchFunc != nil {
		return m.BranchFunc()
	}
	return "main", nil
}

// MockFileOps implements FileOperations for testing.
type MockFileOps struct {
	data   map[string][]byte
	exists map[string]bool
}

func NewMockFileOps() *MockFileOps {
	return &MockFileOps{
		data:   make(map[string][]byte),
		exists: make(map[string]bool),
	}
}

func (m *MockFileOps) ReadFile(path string) ([]byte, error) {
	if data, ok := m.data[path]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileOps) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.data[path] = data
	m.exists[path] = true
	return nil
}

func (m *MockFileOps) FileExists(path string) bool {
	return m.exists[path]
}

func TestNewAuditLog(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		config         *Config
		expectedPath   string
		expectedFormat string
	}{
		{
			name:           "nil config uses defaults",
			config:         nil,
			expectedPath:   ".version-history.json",
			expectedFormat: "json",
		},
		{
			name: "custom config",
			config: &Config{
				Enabled: true,
				Path:    "custom.json",
				Format:  "json",
			},
			expectedPath:   "custom.json",
			expectedFormat: "json",
		},
		{
			name: "yaml format",
			config: &Config{
				Enabled: true,
				Path:    "history.yaml",
				Format:  "yaml",
			},
			expectedPath:   "history.yaml",
			expectedFormat: "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plugin := NewAuditLog(tt.config)
			if plugin == nil {
				t.Fatal("expected plugin to be non-nil")
			}
			if plugin.GetConfig().GetPath() != tt.expectedPath {
				t.Errorf("expected path %q, got %q", tt.expectedPath, plugin.GetConfig().GetPath())
			}
			if plugin.GetConfig().GetFormat() != tt.expectedFormat {
				t.Errorf("expected format %q, got %q", tt.expectedFormat, plugin.GetConfig().GetFormat())
			}
		})
	}
}

func TestAuditLogPlugin_Metadata(t *testing.T) {
	t.Parallel()
	plugin := NewAuditLog(DefaultConfig())

	if plugin.Name() != "audit-log" {
		t.Errorf("expected name 'audit-log', got %q", plugin.Name())
	}

	if plugin.Version() != "v0.1.0" {
		t.Errorf("expected version 'v0.1.0', got %q", plugin.Version())
	}

	if plugin.Description() == "" {
		t.Error("expected non-empty description")
	}
}

func TestAuditLogPlugin_IsEnabled(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := DefaultConfig()
			cfg.Enabled = tt.enabled
			plugin := NewAuditLog(cfg)

			if plugin.IsEnabled() != tt.enabled {
				t.Errorf("expected IsEnabled() = %v, got %v", tt.enabled, plugin.IsEnabled())
			}
		})
	}
}

func TestAuditLogPlugin_RecordEntry_Disabled(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.Enabled = false

	plugin := NewAuditLogWithOps(cfg, &MockGitOps{}, NewMockFileOps())

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Errorf("expected no error when disabled, got %v", err)
	}
}

func TestAuditLogPlugin_RecordEntry_JSON(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeAuthor:    true,
		IncludeTimestamp: true,
		IncludeCommitSHA: true,
		IncludeBranch:    true,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOps()
	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	// Fix time for consistent testing
	fixedTime := time.Date(2026, 1, 4, 12, 0, 0, 0, time.UTC)
	plugin.timeFunc = func() time.Time { return fixedTime }

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was written
	data, ok := mockFile.data[cfg.Path]
	if !ok {
		t.Fatal("expected file to be written")
	}

	var logFile AuditLogFile
	if err := json.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(logFile.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(logFile.Entries))
	}

	actualEntry := logFile.Entries[0]
	if actualEntry.PreviousVersion != "1.0.0" {
		t.Errorf("expected previous version '1.0.0', got %q", actualEntry.PreviousVersion)
	}
	if actualEntry.NewVersion != "1.0.1" {
		t.Errorf("expected new version '1.0.1', got %q", actualEntry.NewVersion)
	}
	if actualEntry.BumpType != "patch" {
		t.Errorf("expected bump type 'patch', got %q", actualEntry.BumpType)
	}
	if actualEntry.Author != "Test User <test@example.com>" {
		t.Errorf("expected author 'Test User <test@example.com>', got %q", actualEntry.Author)
	}
	if actualEntry.CommitSHA != "abc1234567890def" {
		t.Errorf("expected commit SHA 'abc1234567890def', got %q", actualEntry.CommitSHA)
	}
	if actualEntry.Branch != "main" {
		t.Errorf("expected branch 'main', got %q", actualEntry.Branch)
	}
	if actualEntry.Timestamp != fixedTime.UTC().Format(time.RFC3339) {
		t.Errorf("expected timestamp %q, got %q", fixedTime.UTC().Format(time.RFC3339), actualEntry.Timestamp)
	}
}

func TestAuditLogPlugin_RecordEntry_YAML(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.yaml",
		Format:           "yaml",
		IncludeAuthor:    true,
		IncludeTimestamp: true,
		IncludeCommitSHA: true,
		IncludeBranch:    true,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOps()
	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	entry := &Entry{
		PreviousVersion: "2.0.0",
		NewVersion:      "3.0.0",
		BumpType:        "major",
	}

	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was written
	data, ok := mockFile.data[cfg.Path]
	if !ok {
		t.Fatal("expected file to be written")
	}

	var logFile AuditLogFile
	if err := yaml.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(logFile.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(logFile.Entries))
	}

	actualEntry := logFile.Entries[0]
	if actualEntry.PreviousVersion != "2.0.0" {
		t.Errorf("expected previous version '2.0.0', got %q", actualEntry.PreviousVersion)
	}
	if actualEntry.NewVersion != "3.0.0" {
		t.Errorf("expected new version '3.0.0', got %q", actualEntry.NewVersion)
	}
	if actualEntry.BumpType != "major" {
		t.Errorf("expected bump type 'major', got %q", actualEntry.BumpType)
	}
}

func TestAuditLogPlugin_RecordEntry_MultipleEntries(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeTimestamp: true,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOps()
	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	// Create entries at different times
	times := []time.Time{
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 3, 12, 0, 0, 0, time.UTC),
	}

	entries := []*Entry{
		{PreviousVersion: "1.0.0", NewVersion: "1.0.1", BumpType: "patch"},
		{PreviousVersion: "1.0.1", NewVersion: "1.1.0", BumpType: "minor"},
		{PreviousVersion: "1.1.0", NewVersion: "2.0.0", BumpType: "major"},
	}

	for i, entry := range entries {
		plugin.timeFunc = func() time.Time { return times[i] }
		if err := plugin.RecordEntry(entry); err != nil {
			t.Fatalf("unexpected error on entry %d: %v", i, err)
		}
	}

	// Verify entries are sorted newest first
	data := mockFile.data[cfg.Path]
	var logFile AuditLogFile
	if err := json.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(logFile.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(logFile.Entries))
	}

	// Newest first
	if logFile.Entries[0].NewVersion != "2.0.0" {
		t.Errorf("expected first entry to be newest (2.0.0), got %q", logFile.Entries[0].NewVersion)
	}
	if logFile.Entries[2].NewVersion != "1.0.1" {
		t.Errorf("expected last entry to be oldest (1.0.1), got %q", logFile.Entries[2].NewVersion)
	}
}

func TestAuditLogPlugin_RecordEntry_GitError(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeAuthor:    true,
		IncludeCommitSHA: true,
		IncludeBranch:    true,
	}

	mockGit := &MockGitOps{
		AuthorFunc: func() (string, error) {
			return "", errors.New("git error")
		},
		CommitSHAFunc: func() (string, error) {
			return "", errors.New("git error")
		},
		BranchFunc: func() (string, error) {
			return "", errors.New("git error")
		},
	}
	mockFile := NewMockFileOps()
	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	// Should not fail even if git operations fail
	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Errorf("expected no error when git operations fail, got %v", err)
	}

	// Verify entry was still written with basic info
	data := mockFile.data[cfg.Path]
	var logFile AuditLogFile
	if err := json.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(logFile.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(logFile.Entries))
	}

	// Git fields should be empty
	actualEntry := logFile.Entries[0]
	if actualEntry.Author != "" {
		t.Errorf("expected empty author when git fails, got %q", actualEntry.Author)
	}
	if actualEntry.CommitSHA != "" {
		t.Errorf("expected empty commit SHA when git fails, got %q", actualEntry.CommitSHA)
	}
	if actualEntry.Branch != "" {
		t.Errorf("expected empty branch when git fails, got %q", actualEntry.Branch)
	}
}

func TestAuditLogPlugin_RecordEntry_SelectiveMetadata(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		includeAuthor bool
		includeSHA    bool
		includeBranch bool
		includeTime   bool
	}{
		{"all disabled", false, false, false, false},
		{"only author", true, false, false, false},
		{"only sha", false, true, false, false},
		{"only branch", false, false, true, false},
		{"only timestamp", false, false, false, true},
		{"author and sha", true, true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			entry := recordEntryWithConfig(t, tt.includeAuthor, tt.includeSHA, tt.includeBranch, tt.includeTime)
			verifyMetadataInclusion(t, entry, tt.includeAuthor, tt.includeSHA, tt.includeBranch, tt.includeTime)
		})
	}
}

func recordEntryWithConfig(t *testing.T, includeAuthor, includeSHA, includeBranch, includeTime bool) Entry {
	t.Helper()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeAuthor:    includeAuthor,
		IncludeCommitSHA: includeSHA,
		IncludeBranch:    includeBranch,
		IncludeTimestamp: includeTime,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOps()
	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	if err := plugin.RecordEntry(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := mockFile.data[cfg.Path]
	var logFile AuditLogFile
	if err := json.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	return logFile.Entries[0]
}

func verifyMetadataInclusion(t *testing.T, entry Entry, includeAuthor, includeSHA, includeBranch, includeTime bool) {
	t.Helper()
	verifyField(t, "author", entry.Author, includeAuthor)
	verifyField(t, "commit SHA", entry.CommitSHA, includeSHA)
	verifyField(t, "branch", entry.Branch, includeBranch)
	verifyField(t, "timestamp", entry.Timestamp, includeTime)
}

func verifyField(t *testing.T, fieldName, fieldValue string, shouldInclude bool) {
	t.Helper()
	if shouldInclude && fieldValue == "" {
		t.Errorf("expected %s to be included", fieldName)
	}
	if !shouldInclude && fieldValue != "" {
		t.Errorf("expected %s to be excluded, got %q", fieldName, fieldValue)
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("expected default enabled to be false")
	}
	if cfg.GetPath() != ".version-history.json" {
		t.Errorf("expected default path '.version-history.json', got %q", cfg.GetPath())
	}
	if cfg.GetFormat() != "json" {
		t.Errorf("expected default format 'json', got %q", cfg.GetFormat())
	}
	if !cfg.IncludeAuthor {
		t.Error("expected default include-author to be true")
	}
	if !cfg.IncludeTimestamp {
		t.Error("expected default include-timestamp to be true")
	}
	if !cfg.IncludeCommitSHA {
		t.Error("expected default include-commit-sha to be true")
	}
	if !cfg.IncludeBranch {
		t.Error("expected default include-branch to be true")
	}
}

func TestConfig_GetPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"empty path uses default", "", ".version-history.json"},
		{"custom path", "custom.json", "custom.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{Path: tt.path}
			if cfg.GetPath() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, cfg.GetPath())
			}
		})
	}
}

func TestConfig_GetFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{"empty format uses default", "", "json"},
		{"json format", "json", "json"},
		{"yaml format", "yaml", "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{Format: tt.format}
			if cfg.GetFormat() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, cfg.GetFormat())
			}
		})
	}
}

// MockFileOpsWithErrors is a mock that can simulate file operation errors.
type MockFileOpsWithErrors struct {
	data       map[string][]byte
	exists     map[string]bool
	readError  error
	writeError error
}

func NewMockFileOpsWithErrors() *MockFileOpsWithErrors {
	return &MockFileOpsWithErrors{
		data:   make(map[string][]byte),
		exists: make(map[string]bool),
	}
}

func (m *MockFileOpsWithErrors) ReadFile(path string) ([]byte, error) {
	if m.readError != nil {
		return nil, m.readError
	}
	if data, ok := m.data[path]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileOpsWithErrors) WriteFile(path string, data []byte, perm os.FileMode) error {
	if m.writeError != nil {
		return m.writeError
	}
	m.data[path] = data
	m.exists[path] = true
	return nil
}

func (m *MockFileOpsWithErrors) FileExists(path string) bool {
	return m.exists[path]
}

func TestAuditLogPlugin_RecordEntry_FileReadError(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeTimestamp: true,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOpsWithErrors()
	mockFile.exists[".version-history.json"] = true
	mockFile.readError = errors.New("disk read error")

	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	// Should not return error (non-blocking)
	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Errorf("expected no error (non-blocking), got %v", err)
	}
}

func TestAuditLogPlugin_RecordEntry_FileWriteError(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeTimestamp: true,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOpsWithErrors()
	mockFile.writeError = errors.New("disk write error")

	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	// Should not return error (non-blocking)
	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Errorf("expected no error (non-blocking), got %v", err)
	}
}

func TestAuditLogPlugin_RecordEntry_InvalidJSONInFile(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeTimestamp: true,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOpsWithErrors()
	mockFile.data[".version-history.json"] = []byte("invalid json{{{")
	mockFile.exists[".version-history.json"] = true

	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	// Should not return error (non-blocking)
	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Errorf("expected no error (non-blocking), got %v", err)
	}
}

func TestAuditLogPlugin_RecordEntry_ExistingEntries(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeTimestamp: true,
	}

	// Pre-populate with existing entries
	existingLog := AuditLogFile{
		Entries: []Entry{
			{
				Timestamp:       "2026-01-01T10:00:00Z",
				PreviousVersion: "0.9.0",
				NewVersion:      "1.0.0",
				BumpType:        "major",
			},
		},
	}
	existingData, _ := json.Marshal(existingLog)

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOpsWithErrors()
	mockFile.data[".version-history.json"] = existingData
	mockFile.exists[".version-history.json"] = true

	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)
	plugin.timeFunc = func() time.Time {
		return time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	}

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify both entries exist
	data := mockFile.data[cfg.Path]
	var logFile AuditLogFile
	if err := json.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(logFile.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(logFile.Entries))
	}

	// Newest should be first
	if logFile.Entries[0].NewVersion != "1.0.1" {
		t.Errorf("expected newest entry first (1.0.1), got %q", logFile.Entries[0].NewVersion)
	}
}

func TestAuditLogPlugin_SortEntries_InvalidTimestamps(t *testing.T) {
	t.Parallel()
	plugin := NewAuditLog(nil)

	entries := []Entry{
		{Timestamp: "invalid-timestamp", NewVersion: "1.0.0"},
		{Timestamp: "also-invalid", NewVersion: "2.0.0"},
		{Timestamp: "2026-01-01T10:00:00Z", NewVersion: "3.0.0"},
	}

	plugin.sortEntries(entries)

	// With invalid timestamps, entries should maintain relative order (only valid one gets sorted)
	// The valid timestamp entry should be at the end since it's the only one that parses
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestAuditLogPlugin_SortEntries_AllValid(t *testing.T) {
	t.Parallel()
	plugin := NewAuditLog(nil)

	entries := []Entry{
		{Timestamp: "2026-01-01T10:00:00Z", NewVersion: "1.0.0"},
		{Timestamp: "2026-01-03T10:00:00Z", NewVersion: "3.0.0"},
		{Timestamp: "2026-01-02T10:00:00Z", NewVersion: "2.0.0"},
	}

	plugin.sortEntries(entries)

	// Should be sorted newest first
	if entries[0].NewVersion != "3.0.0" {
		t.Errorf("expected 3.0.0 first, got %q", entries[0].NewVersion)
	}
	if entries[1].NewVersion != "2.0.0" {
		t.Errorf("expected 2.0.0 second, got %q", entries[1].NewVersion)
	}
	if entries[2].NewVersion != "1.0.0" {
		t.Errorf("expected 1.0.0 last, got %q", entries[2].NewVersion)
	}
}

func TestAuditLogPlugin_NewAuditLogWithOps_NilConfig(t *testing.T) {
	t.Parallel()
	mockGit := &MockGitOps{}
	mockFile := NewMockFileOps()

	plugin := NewAuditLogWithOps(nil, mockGit, mockFile)

	if plugin == nil {
		t.Fatal("expected plugin to be non-nil")
	}

	if plugin.GetConfig() == nil {
		t.Fatal("expected config to be non-nil")
	}

	// Should use default config
	if plugin.GetConfig().GetPath() != ".version-history.json" {
		t.Errorf("expected default path, got %q", plugin.GetConfig().GetPath())
	}
}

func TestDefaultFileOps_FileExists(t *testing.T) {
	t.Parallel()
	fileOps := &DefaultFileOps{}

	// Test non-existent file
	if fileOps.FileExists("/non/existent/path/file.json") {
		t.Error("expected FileExists to return false for non-existent file")
	}
}

func TestDefaultFileOps_ReadFile_NonExistent(t *testing.T) {
	t.Parallel()
	fileOps := &DefaultFileOps{}

	_, err := fileOps.ReadFile("/non/existent/path/file.json")
	if err == nil {
		t.Error("expected error when reading non-existent file")
	}
}

func TestDefaultFileOps_WriteAndReadFile(t *testing.T) {
	t.Parallel()
	fileOps := &DefaultFileOps{}
	tmpFile := t.TempDir() + "/test-audit.json"

	testData := []byte(`{"entries":[]}`)

	// Write file
	err := fileOps.WriteFile(tmpFile, testData, 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Verify file exists
	if !fileOps.FileExists(tmpFile) {
		t.Error("expected file to exist after write")
	}

	// Read file
	data, err := fileOps.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("expected %q, got %q", string(testData), string(data))
	}
}

func TestDefaultGitOps_Integration(t *testing.T) {
	t.Parallel(
	// Skip if not in a git repo
	)

	gitOps := &DefaultGitOps{}

	// These will work if we're in a git repo, otherwise they'll fail
	// We test that the methods don't panic and return appropriate errors
	_, err := gitOps.GetBranch()
	if err != nil {
		t.Skipf("skipping git integration test: %v", err)
	}

	// If we got here, we're in a git repo
	author, err := gitOps.GetAuthor()
	if err != nil {
		t.Logf("GetAuthor failed (may be expected if git user not configured): %v", err)
	} else if author == "" {
		t.Error("expected non-empty author")
	}

	sha, err := gitOps.GetCommitSHA()
	if err != nil {
		t.Logf("GetCommitSHA failed (may be expected if no commits): %v", err)
	} else if sha == "" {
		t.Error("expected non-empty SHA")
	}

	branch, err := gitOps.GetBranch()
	if err != nil {
		t.Errorf("GetBranch failed: %v", err)
	} else if branch == "" {
		t.Error("expected non-empty branch")
	}
}

func TestAuditLogPlugin_YAMLFormat(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.yaml",
		Format:           "yaml",
		IncludeTimestamp: true,
		IncludeAuthor:    true,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOps()

	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)
	plugin.timeFunc = func() time.Time {
		return time.Date(2026, 1, 4, 12, 0, 0, 0, time.UTC)
	}

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify YAML was written
	data, ok := mockFile.data[cfg.Path]
	if !ok {
		t.Fatal("expected file to be written")
	}

	var logFile AuditLogFile
	if err := yaml.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	if len(logFile.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(logFile.Entries))
	}
}

func TestAuditLogPlugin_YAMLWithExistingEntries(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.yaml",
		Format:           "yaml",
		IncludeTimestamp: true,
	}

	// Pre-populate with existing entries
	existingLog := AuditLogFile{
		Entries: []Entry{
			{
				Timestamp:       "2026-01-01T10:00:00Z",
				PreviousVersion: "0.9.0",
				NewVersion:      "1.0.0",
				BumpType:        "major",
			},
		},
	}
	existingData, _ := yaml.Marshal(existingLog)

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOpsWithErrors()
	mockFile.data[".version-history.yaml"] = existingData
	mockFile.exists[".version-history.yaml"] = true

	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)
	plugin.timeFunc = func() time.Time {
		return time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	}

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify both entries exist
	data := mockFile.data[cfg.Path]
	var logFile AuditLogFile
	if err := yaml.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(logFile.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(logFile.Entries))
	}
}
