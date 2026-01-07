package auditlog

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/goccy/go-yaml"
)

// AuditLog defines the interface for audit logging.
type AuditLog interface {
	Name() string
	Description() string
	Version() string

	// RecordEntry logs a version bump with metadata.
	RecordEntry(entry *Entry) error

	// IsEnabled returns whether the plugin is enabled.
	IsEnabled() bool

	// GetConfig returns the plugin configuration.
	GetConfig() *Config
}

// AuditLogPlugin implements the AuditLog interface.
type AuditLogPlugin struct {
	config      *Config
	gitOps      GitOperations
	fileOps     FileOperations
	timeFunc    func() time.Time
	marshalFn   func(any) ([]byte, error)
	unmarshalFn func([]byte, any) error
}

// Entry represents a single audit log entry.
type Entry struct {
	Timestamp       string `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
	PreviousVersion string `json:"previous_version" yaml:"previous_version"`
	NewVersion      string `json:"new_version" yaml:"new_version"`
	BumpType        string `json:"bump_type" yaml:"bump_type"`
	Author          string `json:"author,omitempty" yaml:"author,omitempty"`
	CommitSHA       string `json:"commit_sha,omitempty" yaml:"commit_sha,omitempty"`
	Branch          string `json:"branch,omitempty" yaml:"branch,omitempty"`
}

// AuditLogFile represents the structure of the audit log file.
type AuditLogFile struct {
	Entries []Entry `json:"entries" yaml:"entries"`
}

// GitOperations defines the interface for git operations.
type GitOperations interface {
	GetAuthor() (string, error)
	GetCommitSHA() (string, error)
	GetBranch() (string, error)
}

// FileOperations defines the interface for file operations.
type FileOperations interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	FileExists(path string) bool
}

// Ensure AuditLogPlugin implements AuditLog.
var _ AuditLog = (*AuditLogPlugin)(nil)

// NewAuditLog creates a new audit log plugin.
func NewAuditLog(cfg *Config) *AuditLogPlugin {
	return NewAuditLogWithOps(cfg, &DefaultGitOps{}, &DefaultFileOps{})
}

// NewAuditLogWithOps creates a new audit log plugin with custom operations (for testing).
func NewAuditLogWithOps(cfg *Config, gitOps GitOperations, fileOps FileOperations) *AuditLogPlugin {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	plugin := &AuditLogPlugin{
		config:   cfg,
		gitOps:   gitOps,
		fileOps:  fileOps,
		timeFunc: time.Now,
	}

	// Set marshal/unmarshal functions based on format
	if cfg.GetFormat() == "yaml" {
		plugin.marshalFn = yaml.Marshal
		plugin.unmarshalFn = yaml.Unmarshal
	} else {
		plugin.marshalFn = func(v any) ([]byte, error) {
			return json.MarshalIndent(v, "", "  ")
		}
		plugin.unmarshalFn = json.Unmarshal
	}

	return plugin
}

// Name returns the plugin name.
func (p *AuditLogPlugin) Name() string { return "audit-log" }

// Description returns the plugin description.
func (p *AuditLogPlugin) Description() string {
	return "Records version changes to an audit log"
}

// Version returns the plugin version.
func (p *AuditLogPlugin) Version() string { return "v0.1.0" }

// IsEnabled returns whether the plugin is enabled.
func (p *AuditLogPlugin) IsEnabled() bool {
	return p.config.Enabled
}

// GetConfig returns the plugin configuration.
func (p *AuditLogPlugin) GetConfig() *Config {
	return p.config
}

// RecordEntry logs a version bump with metadata.
func (p *AuditLogPlugin) RecordEntry(entry *Entry) error {
	if !p.config.Enabled {
		return nil
	}

	// Enrich entry with metadata based on config
	if err := p.enrichEntry(entry); err != nil {
		// Log warning but don't fail the version bump
		fmt.Fprintf(os.Stderr, "Warning: failed to enrich audit log entry: %v\n", err)
	}

	// Read existing log
	logFile, err := p.readLogFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to read audit log: %v\n", err)
		return nil // Don't fail the version bump
	}

	// Add new entry
	logFile.Entries = append(logFile.Entries, *entry)

	// Sort entries by timestamp (newest first)
	p.sortEntries(logFile.Entries)

	// Write updated log
	if err := p.writeLogFile(logFile); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write audit log: %v\n", err)
		return nil // Don't fail the version bump
	}

	return nil
}

// enrichEntry adds metadata to the entry based on configuration.
func (p *AuditLogPlugin) enrichEntry(entry *Entry) error {
	if p.config.IncludeTimestamp {
		entry.Timestamp = p.timeFunc().UTC().Format(time.RFC3339)
	}

	if p.config.IncludeAuthor {
		author, err := p.gitOps.GetAuthor()
		if err == nil {
			entry.Author = author
		}
	}

	if p.config.IncludeCommitSHA {
		sha, err := p.gitOps.GetCommitSHA()
		if err == nil {
			entry.CommitSHA = sha
		}
	}

	if p.config.IncludeBranch {
		branch, err := p.gitOps.GetBranch()
		if err == nil {
			entry.Branch = branch
		}
	}

	return nil
}

// readLogFile reads and parses the audit log file.
func (p *AuditLogPlugin) readLogFile() (*AuditLogFile, error) {
	path := p.config.GetPath()

	// If file doesn't exist, return empty log
	if !p.fileOps.FileExists(path) {
		return &AuditLogFile{Entries: []Entry{}}, nil
	}

	data, err := p.fileOps.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read audit log %q: %w", path, err)
	}

	var logFile AuditLogFile
	if err := p.unmarshalFn(data, &logFile); err != nil {
		return nil, fmt.Errorf("failed to parse audit log %q: %w", path, err)
	}

	return &logFile, nil
}

// writeLogFile writes the audit log to disk.
func (p *AuditLogPlugin) writeLogFile(logFile *AuditLogFile) error {
	data, err := p.marshalFn(logFile)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}

	path := p.config.GetPath()
	if err := p.fileOps.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write audit log %q: %w", path, err)
	}

	return nil
}

// sortEntries sorts entries by timestamp, newest first.
func (p *AuditLogPlugin) sortEntries(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool {
		// Parse timestamps
		ti, errI := time.Parse(time.RFC3339, entries[i].Timestamp)
		tj, errJ := time.Parse(time.RFC3339, entries[j].Timestamp)

		// If parsing fails, maintain current order
		if errI != nil || errJ != nil {
			return false
		}

		// Sort newest first
		return ti.After(tj)
	})
}
