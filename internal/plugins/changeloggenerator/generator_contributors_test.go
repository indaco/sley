package changeloggenerator

import (
	"strings"
	"testing"
)

func TestWriteContributorEntry(t *testing.T) {
	g, err := NewGenerator(DefaultConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	remote := &RemoteInfo{Provider: "github", Host: "github.com", Owner: "owner", Repo: "repo"}

	tests := []struct {
		name     string
		contrib  Contributor
		remote   *RemoteInfo
		contains []string
	}{
		{
			name:     "With remote",
			contrib:  Contributor{Name: "Alice", Username: "alice", Host: "github.com"},
			remote:   remote,
			contains: []string{"@alice", "github.com/alice"},
		},
		{
			name:     "Without remote",
			contrib:  Contributor{Name: "Bob", Username: "bob"},
			remote:   nil,
			contains: []string{"- @bob"},
		},
		{
			name:     "Contributor with different host",
			contrib:  Contributor{Name: "Charlie", Username: "charlie", Host: "gitlab.com"},
			remote:   remote,
			contains: []string{"@charlie", "gitlab.com/charlie"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			g.writeContributorEntry(&sb, tt.contrib, tt.remote)
			got := sb.String()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("writeContributorEntry() = %q, expected to contain %q", got, want)
				}
			}
		})
	}
}

func TestWriteContributorEntry_CustomFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		contrib  Contributor
		remote   *RemoteInfo
		expected string
	}{
		{
			name:     "Custom format with name and username",
			format:   "- {{.Name}} (@{{.Username}})",
			contrib:  Contributor{Name: "Alice Smith", Username: "alice", Host: "github.com"},
			remote:   &RemoteInfo{Host: "github.com", Owner: "test", Repo: "repo"},
			expected: "- Alice Smith (@alice)\n",
		},
		{
			name:     "Custom format username only",
			format:   "- @{{.Username}}",
			contrib:  Contributor{Name: "Bob Jones", Username: "bob", Host: "github.com"},
			remote:   &RemoteInfo{Host: "github.com", Owner: "test", Repo: "repo"},
			expected: "- @bob\n",
		},
		{
			name:     "Custom format with email",
			format:   "- {{.Username}} <{{.Email}}>",
			contrib:  Contributor{Name: "Charlie", Username: "charlie", Email: "charlie@example.com", Host: "github.com"},
			remote:   &RemoteInfo{Host: "github.com", Owner: "test", Repo: "repo"},
			expected: "- charlie <charlie@example.com>\n",
		},
		{
			name:     "Default format when empty",
			format:   "",
			contrib:  Contributor{Name: "Dave", Username: "dave", Host: "github.com"},
			remote:   &RemoteInfo{Host: "github.com", Owner: "test", Repo: "repo"},
			expected: "- [@dave](https://github.com/dave)\n",
		},
		{
			name:     "Fallback on invalid template",
			format:   "- {{.Invalid",
			contrib:  Contributor{Name: "Eve", Username: "eve", Host: "github.com"},
			remote:   &RemoteInfo{Host: "github.com", Owner: "test", Repo: "repo"},
			expected: "- [@eve](https://github.com/eve)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Contributors.Format = tt.format
			g, err := NewGenerator(cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var sb strings.Builder
			g.writeContributorEntry(&sb, tt.contrib, tt.remote)
			got := sb.String()

			if got != tt.expected {
				t.Errorf("writeContributorEntry() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWriteContributorEntry_NoHost(t *testing.T) {
	cfg := DefaultConfig()
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contrib := Contributor{Name: "NoHost User", Username: "nohost", Host: ""}

	var sb strings.Builder
	g.writeContributorEntry(&sb, contrib, nil)
	got := sb.String()

	expected := "- @nohost\n"
	if got != expected {
		t.Errorf("writeContributorEntry() = %q, want %q", got, expected)
	}
}

func TestWriteContributorEntry_TemplateExecutionError(t *testing.T) {
	cfg := DefaultConfig()
	// Invalid template that parses but fails on execution
	cfg.Contributors = &ContributorsConfig{
		Enabled: true,
		Format:  "- {{.NonExistentMethod}}",
	}
	g, err := NewGenerator(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	contrib := Contributor{
		Name:     "Test User",
		Username: "testuser",
		Host:     "github.com",
	}

	var sb strings.Builder
	g.writeContributorEntry(&sb, contrib, &RemoteInfo{Host: "github.com"})
	result := sb.String()

	// Should fallback to default format
	if !strings.Contains(result, "@testuser") {
		t.Error("expected fallback format with username")
	}
}

func TestWriteNewContributorsSection(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:             true,
		ShowNewContributors: true,
	}
	g, _ := NewGenerator(cfg)

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "testowner",
		Repo:     "testrepo",
	}

	newContributors := []NewContributor{
		{
			Contributor: Contributor{
				Name:     "New Dev",
				Username: "newdev",
				Host:     "github.com",
			},
			FirstCommit: CommitInfo{ShortHash: "abc123"},
			PRNumber:    "42",
		},
	}

	var sb strings.Builder
	g.writeNewContributorsSection(&sb, newContributors, remote)
	result := sb.String()

	if !strings.Contains(result, "### New Contributors") {
		t.Error("expected New Contributors header")
	}
	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in output")
	}
	if !strings.Contains(result, "#42") {
		t.Error("expected PR number in output")
	}
}

func TestWriteNewContributorsSection_WithIcon(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:             true,
		ShowNewContributors: true,
		NewContributorsIcon: "🎉",
	}
	g, _ := NewGenerator(cfg)

	newContributors := []NewContributor{
		{
			Contributor: Contributor{
				Name:     "New Dev",
				Username: "newdev",
				Host:     "github.com",
			},
			FirstCommit: CommitInfo{ShortHash: "abc123"},
			PRNumber:    "42",
		},
	}

	var sb strings.Builder
	g.writeNewContributorsSection(&sb, newContributors, nil)
	result := sb.String()

	if !strings.Contains(result, "### 🎉 New Contributors") {
		t.Error("expected New Contributors header with icon")
	}
}

func TestWriteNewContributorEntry_WithRemote(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:             true,
		ShowNewContributors: true,
	}
	g, _ := NewGenerator(cfg)

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, remote)
	result := sb.String()

	if !strings.Contains(result, "newdev") {
		t.Error("expected username in output")
	}
	if !strings.Contains(result, "first contribution") {
		t.Error("expected 'first contribution' text")
	}
}

func TestWriteNewContributorEntry_WithoutPR(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:             true,
		ShowNewContributors: true,
	}
	g, _ := NewGenerator(cfg)

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "", // No PR number
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, remote)
	result := sb.String()

	if !strings.Contains(result, "newdev") {
		t.Error("expected username in output")
	}
	// Should contain commit hash as a link when no PR number
	if !strings.Contains(result, "abc123") {
		t.Error("expected commit hash in output when no PR number")
	}
	// Verify commit hash is linked
	expectedCommitLink := "[abc123](https://github.com/owner/repo/commit/abc123)"
	if !strings.Contains(result, expectedCommitLink) {
		t.Errorf("expected commit hash link %q in output, got: %s", expectedCommitLink, result)
	}
}

func TestWriteNewContributorEntry_WithoutRemote(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:             true,
		ShowNewContributors: true,
	}
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, nil) // nil remote
	result := sb.String()

	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in output")
	}
	if !strings.Contains(result, "#42") {
		t.Error("expected PR number in output")
	}
}

func TestWriteNewContributorEntry_CustomFormat(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:               true,
		ShowNewContributors:   true,
		NewContributorsFormat: "* {{.Username}} joined in #{{.PRNumber}}",
	}
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, nil)
	result := sb.String()

	if !strings.Contains(result, "newdev joined in #42") {
		t.Errorf("expected custom format output, got: %s", result)
	}
}

func TestWriteNewContributorEntry_TemplateParseError(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:               true,
		ShowNewContributors:   true,
		NewContributorsFormat: "- {{.Invalid", // Invalid template syntax
	}
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, &RemoteInfo{Host: "github.com", Owner: "owner", Repo: "repo"})
	result := sb.String()

	// Should fallback
	if !strings.Contains(result, "@newdev") {
		t.Error("expected fallback format with username")
	}
}

func TestWriteNewContributorEntry_TemplateExecutionError(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Contributors = &ContributorsConfig{
		Enabled:               true,
		ShowNewContributors:   true,
		NewContributorsFormat: "- {{.NonExistent.Field}}", // Will fail on execution
	}
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorEntry(&sb, &nc, &RemoteInfo{Host: "github.com", Owner: "owner", Repo: "repo"})
	result := sb.String()

	// Should fallback
	if !strings.Contains(result, "@newdev") {
		t.Error("expected fallback format with username")
	}
}

func TestWriteNewContributorFallback_WithPR(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorFallback(&sb, &nc, remote)
	result := sb.String()

	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in fallback output")
	}
	if !strings.Contains(result, "#42") {
		t.Error("expected PR number in fallback output")
	}
}

func TestWriteNewContributorFallback_WithoutPR(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
			Host:     "github.com",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "",
	}

	remote := &RemoteInfo{
		Host:  "github.com",
		Owner: "owner",
		Repo:  "repo",
	}

	var sb strings.Builder
	g.writeNewContributorFallback(&sb, &nc, remote)
	result := sb.String()

	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in fallback output")
	}
	if !strings.Contains(result, "first contribution") {
		t.Error("expected 'first contribution' in fallback output")
	}
	// Verify commit hash is linked in fallback
	expectedCommitLink := "[abc123](https://github.com/owner/repo/commit/abc123)"
	if !strings.Contains(result, expectedCommitLink) {
		t.Errorf("expected commit hash link %q in fallback output, got: %s", expectedCommitLink, result)
	}
}

func TestWriteNewContributorFallback_WithoutRemote(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "42",
	}

	var sb strings.Builder
	g.writeNewContributorFallback(&sb, &nc, nil)
	result := sb.String()

	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in fallback output")
	}
	if !strings.Contains(result, "#42") {
		t.Error("expected PR number in fallback output")
	}
}

func TestWriteNewContributorFallback_NoRemoteNoPR(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	nc := NewContributor{
		Contributor: Contributor{
			Name:     "New Dev",
			Username: "newdev",
		},
		FirstCommit: CommitInfo{ShortHash: "abc123"},
		PRNumber:    "",
	}

	var sb strings.Builder
	g.writeNewContributorFallback(&sb, &nc, nil)
	result := sb.String()

	if !strings.Contains(result, "@newdev") {
		t.Error("expected username in fallback output")
	}
	if !strings.Contains(result, "first contribution") {
		t.Error("expected 'first contribution' in fallback output")
	}
}

func TestGetDefaultNewContributorFormat_WithRemote(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	remote := &RemoteInfo{
		Provider: "github",
		Host:     "github.com",
		Owner:    "owner",
		Repo:     "repo",
	}

	format := g.getDefaultNewContributorFormat(remote)

	if !strings.Contains(format, "{{.Username}}") {
		t.Error("expected username placeholder in format")
	}
	if !strings.Contains(format, "{{.PRNumber}}") {
		t.Error("expected PR number placeholder in format")
	}
	if !strings.Contains(format, "owner/repo") {
		t.Error("expected owner/repo in format for PR links")
	}
}

func TestGetDefaultNewContributorFormat_WithoutRemote(t *testing.T) {
	cfg := DefaultConfig()
	g, _ := NewGenerator(cfg)

	format := g.getDefaultNewContributorFormat(nil)

	if !strings.Contains(format, "{{.Username}}") {
		t.Error("expected username placeholder in format")
	}
	if !strings.Contains(format, "{{.PRNumber}}") {
		t.Error("expected PR number placeholder in format")
	}
	// Should not contain full URL format
	if strings.Contains(format, "https://{{.Host}}") {
		t.Error("expected simpler format without full URLs when no remote")
	}
}
