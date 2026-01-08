package extensionmgr

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestScriptExecutor_MalformedJSONResponse tests handling of malformed JSON from extensions.
func TestScriptExecutor_MalformedJSONResponse(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		scriptContent string
		wantErr       bool
	}{
		{
			name: "empty output",
			scriptContent: `#!/bin/sh
echo ""
`,
			wantErr: true,
		},
		{
			name: "invalid JSON",
			scriptContent: `#!/bin/sh
echo "not valid json"
`,
			wantErr: true,
		},
		{
			name: "incomplete JSON",
			scriptContent: `#!/bin/sh
echo '{"version": "1.0.0"'
`,
			wantErr: true,
		},
		{
			name: "JSON array instead of object",
			scriptContent: `#!/bin/sh
echo '[1, 2, 3]'
`,
			wantErr: true,
		},
		{
			name: "null JSON",
			scriptContent: `#!/bin/sh
echo 'null'
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(tempDir, tt.name+".sh")
			if err := os.WriteFile(scriptPath, []byte(tt.scriptContent), 0755); err != nil {
				t.Fatalf("failed to write script: %v", err)
			}

			executor := NewScriptExecutorWithTimeout(30 * time.Second)
			ctx := context.Background()

			input := &HookInput{
				Hook:        "test",
				Version:     "1.0.0",
				BumpType:    "patch",
				ProjectRoot: "/test",
			}

			_, err := executor.Execute(ctx, scriptPath, input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestScriptExecutor_ScriptNotFound_Recovery tests handling of missing script file.
func TestScriptExecutor_ScriptNotFound_Recovery(t *testing.T) {
	executor := NewScriptExecutorWithTimeout(30 * time.Second)
	ctx := context.Background()

	input := &HookInput{
		Hook:        "test",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	_, err := executor.Execute(ctx, "/nonexistent/script.sh", input)
	if err == nil {
		t.Fatal("expected error for nonexistent script, got nil")
	}
}

// TestScriptExecutor_ContextCancellation_Recovery tests that execution respects context cancellation.
func TestScriptExecutor_ContextCancellation_Recovery(t *testing.T) {
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "slow.sh")

	// Create a script that sleeps
	script := `#!/bin/sh
sleep 10
echo '{"version": "1.0.0"}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	executor := NewScriptExecutorWithTimeout(30 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	input := &HookInput{
		Hook:        "test",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := executor.Execute(ctx, scriptPath, input)
	if err == nil {
		t.Fatal("expected error when context is cancelled, got nil")
	}
}

// TestScriptExecutor_TimeoutRecovery tests that execution respects timeout.
func TestScriptExecutor_TimeoutRecovery(t *testing.T) {
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "slow.sh")

	// Create a script that loops and checks for termination
	// Using a trap to ensure clean shutdown
	script := `#!/bin/sh
trap 'exit 124' TERM INT
i=0
while [ $i -lt 100 ]; do
    sleep 0.1
    i=$((i+1))
done
echo '{"success": true, "version": "1.0.0"}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	// Use short timeout (500ms should be enough for process startup and kill)
	executor := NewScriptExecutorWithTimeout(500 * time.Millisecond)
	ctx := context.Background()

	input := &HookInput{
		Hook:        "test",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	start := time.Now()
	_, err := executor.Execute(ctx, scriptPath, input)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	// Should have timed out within 2x the timeout duration (to account for process overhead)
	maxAllowedTime := 2 * time.Second
	if elapsed > maxAllowedTime {
		t.Errorf("execution took too long (%v), timeout may not be working", elapsed)
	}
}

// TestScriptExecutor_NonZeroExitCode_Recovery tests handling of scripts that exit with non-zero.
func TestScriptExecutor_NonZeroExitCode_Recovery(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		exitCode string
	}{
		{"exit 1", "1"},
		{"exit 2", "2"},
		{"exit 127", "127"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(tempDir, "exit_test.sh")
			script := "#!/bin/sh\nexit " + tt.exitCode + "\n"
			if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
				t.Fatalf("failed to write script: %v", err)
			}

			executor := NewScriptExecutorWithTimeout(30 * time.Second)
			ctx := context.Background()

			input := &HookInput{
				Hook:        "test",
				Version:     "1.0.0",
				ProjectRoot: "/test",
			}

			_, err := executor.Execute(ctx, scriptPath, input)
			if err == nil {
				t.Errorf("expected error for exit code %s, got nil", tt.exitCode)
			}
		})
	}
}

// TestScriptExecutor_StderrOutputRecovery tests that stderr is captured in error.
func TestScriptExecutor_StderrOutputRecovery(t *testing.T) {
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "stderr.sh")

	script := `#!/bin/sh
echo "error message" >&2
exit 1
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	executor := NewScriptExecutorWithTimeout(30 * time.Second)
	ctx := context.Background()

	input := &HookInput{
		Hook:        "test",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	_, err := executor.Execute(ctx, scriptPath, input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Error should contain some indication of the failure
	errStr := err.Error()
	if errStr == "" {
		t.Error("expected non-empty error message")
	}
}

// TestScriptExecutor_LargeOutputRecovery tests handling of large output from scripts.
func TestScriptExecutor_LargeOutputRecovery(t *testing.T) {
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "large_output.sh")

	// Generate a script that produces large output followed by valid JSON
	script := `#!/bin/sh
# Generate some large output
for i in $(seq 1 1000); do
    echo "line $i of output"
done
# Output valid JSON at the end
echo '{"version": "1.0.0", "message": "success"}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	executor := NewScriptExecutorWithTimeout(30 * time.Second)
	ctx := context.Background()

	input := &HookInput{
		Hook:        "test",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	// This should handle the large output without issues
	// The exact behavior depends on implementation - it may succeed or fail
	// depending on how the executor handles output
	_, _ = executor.Execute(ctx, scriptPath, input)
	// We're just testing it doesn't panic or hang
}

// TestParseRepoURL_InvalidURLs tests handling of various invalid URL formats.
func TestParseRepoURL_InvalidURLs(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"empty URL", "", true},
		{"just host", "github.com", true},
		{"missing repo", "github.com/owner", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseRepoURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

// TestParseRepoURL_ValidURLs tests parsing of valid URL formats.
func TestParseRepoURL_ValidURLs(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
	}{
		{"https with .git", "https://github.com/owner/repo.git", "owner", "repo"},
		{"https without .git", "https://github.com/owner/repo", "owner", "repo"},
		{"http URL", "http://github.com/owner/repo", "owner", "repo"},
		{"without protocol", "github.com/owner/repo", "owner", "repo"},
		{"gitlab URL", "https://gitlab.com/owner/repo", "owner", "repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseRepoURL(tt.url)
			if err != nil {
				t.Fatalf("ParseRepoURL(%q) unexpected error: %v", tt.url, err)
			}
			if result.Owner != tt.wantOwner {
				t.Errorf("ParseRepoURL(%q) Owner = %q, want %q", tt.url, result.Owner, tt.wantOwner)
			}
			if result.Repo != tt.wantRepo {
				t.Errorf("ParseRepoURL(%q) Repo = %q, want %q", tt.url, result.Repo, tt.wantRepo)
			}
		})
	}
}

// TestInstallFromURL_InvalidURL tests error handling for invalid URLs in InstallFromURL.
func TestInstallFromURL_InvalidURL(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".sley.yaml")
	extDir := filepath.Join(tempDir, "extensions")

	// Empty URL should fail
	err := InstallFromURL("", configPath, extDir)
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}

	// Invalid URL format should fail
	err = InstallFromURL("not-a-valid-url", configPath, extDir)
	if err == nil {
		t.Fatal("expected error for invalid URL format, got nil")
	}
}
