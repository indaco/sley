package changeloggenerator

import (
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	cfg := DefaultConfig()
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g == nil {
		t.Fatal("expected non-nil generator")
	}
	if g.config != cfg {
		t.Error("expected config to match")
	}
}

func TestNewGenerator_InvalidFormat(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Format = "invalid-format"

	_, err := NewGenerator(cfg)
	if err == nil {
		t.Error("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "unknown changelog format") {
		t.Errorf("error = %v, expected to contain 'unknown changelog format'", err)
	}
}

func TestFormatCommitEntry(t *testing.T) {
	remote := &RemoteInfo{Provider: "github", Host: "github.com", Owner: "owner", Repo: "repo"}

	tests := []struct {
		name     string
		commit   *GroupedCommit
		remote   *RemoteInfo
		contains []string
	}{
		{
			name: "Basic commit",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{ShortHash: "abc123"},
					Description: "add feature",
				},
			},
			remote:   remote,
			contains: []string{"- add feature", "abc123"},
		},
		{
			name: "Commit with scope",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{ShortHash: "def456"},
					Description: "update config",
					Scope:       "cli",
				},
			},
			remote:   remote,
			contains: []string{"**cli:**", "update config"},
		},
		{
			name: "Commit with PR number",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{ShortHash: "ghi789"},
					Description: "fix bug",
					PRNumber:    "42",
				},
			},
			remote:   remote,
			contains: []string{"fix bug", "ghi789", "#42", "pull/42"},
		},
		{
			name: "Commit without remote",
			commit: &GroupedCommit{
				ParsedCommit: &ParsedCommit{
					CommitInfo:  CommitInfo{ShortHash: "jkl012"},
					Description: "simple change",
				},
			},
			remote:   nil,
			contains: []string{"- simple change"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCommitEntry(tt.commit, tt.remote)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("formatCommitEntry() = %q, expected to contain %q", got, want)
				}
			}
		})
	}
}
