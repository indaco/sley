package extensionmgr

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewScriptExecutor(t *testing.T) {
	executor := NewScriptExecutor()
	if executor == nil {
		t.Fatal("expected non-nil executor")
	}
	if executor.Timeout != DefaultTimeout {
		t.Errorf("expected timeout %v, got %v", DefaultTimeout, executor.Timeout)
	}
}

func TestNewScriptExecutorWithTimeout(t *testing.T) {
	customTimeout := 5 * time.Second
	executor := NewScriptExecutorWithTimeout(customTimeout)
	if executor == nil {
		t.Fatal("expected non-nil executor")
	}
	if executor.Timeout != customTimeout {
		t.Errorf("expected timeout %v, got %v", customTimeout, executor.Timeout)
	}
}

func TestScriptExecutor_Execute_Success(t *testing.T) {
	// Create a temporary test script
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")

	script := `#!/bin/sh
read input
echo '{"success": true, "message": "Test successful", "data": {"hook": "test"}}'
`

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewScriptExecutor()
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.2.3",
		BumpType:    "patch",
		ProjectRoot: "/test/project",
	}

	ctx := context.Background()
	output, err := executor.Execute(ctx, scriptPath, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("expected success=true")
	}
	if output.Message != "Test successful" {
		t.Errorf("expected message 'Test successful', got '%s'", output.Message)
	}
}

func TestScriptExecutor_Execute_Failure(t *testing.T) {
	// Create a temporary test script that reports failure
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")

	script := `#!/bin/sh
read input
echo '{"success": false, "message": "Test failure"}'
`

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewScriptExecutor()
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.2.3",
		ProjectRoot: "/test/project",
	}

	ctx := context.Background()
	output, err := executor.Execute(ctx, scriptPath, input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if output == nil {
		t.Fatal("expected output even on failure")
	}
	if output.Success {
		t.Error("expected success=false")
	}
}

func TestScriptExecutor_Execute_NonExistentScript(t *testing.T) {
	executor := NewScriptExecutor()
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.2.3",
		ProjectRoot: "/test/project",
	}

	ctx := context.Background()
	_, err := executor.Execute(ctx, "/nonexistent/script.sh", input)
	if err == nil {
		t.Fatal("expected error for nonexistent script")
	}
}

func TestScriptExecutor_Execute_NotExecutable(t *testing.T) {
	// Create a temporary test script without execute permissions
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")

	script := `#!/bin/sh
echo '{"success": true}'
`

	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewScriptExecutor()
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.2.3",
		ProjectRoot: "/test/project",
	}

	ctx := context.Background()
	_, err := executor.Execute(ctx, scriptPath, input)
	if err == nil {
		t.Fatal("expected error for non-executable script")
	}
}

func TestScriptExecutor_Execute_Timeout(t *testing.T) {
	// Create a temporary test script that sleeps
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")

	script := `#!/bin/sh
sleep 5
echo '{"success": true}'
`

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	// Use a very short timeout
	executor := NewScriptExecutorWithTimeout(100 * time.Millisecond)
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.2.3",
		ProjectRoot: "/test/project",
	}

	ctx := context.Background()
	_, err := executor.Execute(ctx, scriptPath, input)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestScriptExecutor_Execute_InvalidJSON(t *testing.T) {
	// Create a temporary test script that outputs invalid JSON
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")

	script := `#!/bin/sh
read input
echo 'not valid json'
`

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewScriptExecutor()
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.2.3",
		ProjectRoot: "/test/project",
	}

	ctx := context.Background()
	_, err := executor.Execute(ctx, scriptPath, input)
	if err == nil {
		t.Fatal("expected error for invalid JSON output")
	}
}

func TestScriptExecutor_Execute_InputParsing(t *testing.T) {
	// Create a temporary test script that echoes back the input
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")

	script := `#!/bin/sh
read input
# Parse input and verify it's valid JSON
version=$(echo "$input" | grep -o '"version":"[^"]*"' | cut -d'"' -f4)
echo "{\"success\": true, \"message\": \"Received version: $version\"}"
`

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewScriptExecutor()
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "2.0.0",
		BumpType:    "major",
		ProjectRoot: "/test/project",
	}

	ctx := context.Background()
	output, err := executor.Execute(ctx, scriptPath, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("expected success=true")
	}
}

func TestExecuteExtensionHook(t *testing.T) {
	// Create a temporary extension directory
	tmpDir := t.TempDir()
	scriptName := "hook.sh"
	scriptPath := filepath.Join(tmpDir, scriptName)

	script := `#!/bin/sh
read input
echo '{"success": true, "message": "Extension executed"}'
`

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	input := HookInput{
		Hook:        "post-bump",
		Version:     "1.2.3",
		ProjectRoot: "/test/project",
	}

	ctx := context.Background()
	output, err := ExecuteExtensionHook(ctx, tmpDir, scriptName, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !output.Success {
		t.Error("expected success=true")
	}
	if output.Message != "Extension executed" {
		t.Errorf("expected message 'Extension executed', got '%s'", output.Message)
	}
}

func TestHookInput_JSON_Serialization(t *testing.T) {
	prerelease := "alpha"
	metadata := "build123"

	input := HookInput{
		Hook:            "pre-bump",
		Version:         "1.2.3",
		PreviousVersion: "1.2.2",
		BumpType:        "patch",
		Prerelease:      &prerelease,
		Metadata:        &metadata,
		ProjectRoot:     "/test/project",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}

	var decoded HookInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal input: %v", err)
	}

	if decoded.Hook != input.Hook {
		t.Errorf("hook mismatch: expected %s, got %s", input.Hook, decoded.Hook)
	}
	if decoded.Version != input.Version {
		t.Errorf("version mismatch: expected %s, got %s", input.Version, decoded.Version)
	}
	if *decoded.Prerelease != *input.Prerelease {
		t.Errorf("prerelease mismatch: expected %s, got %s", *input.Prerelease, *decoded.Prerelease)
	}
}

func TestHookOutput_JSON_Deserialization(t *testing.T) {
	jsonOutput := `{"success": true, "message": "Test", "data": {"key": "value"}}`

	var output HookOutput
	if err := json.Unmarshal([]byte(jsonOutput), &output); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if !output.Success {
		t.Error("expected success=true")
	}
	if output.Message != "Test" {
		t.Errorf("expected message 'Test', got '%s'", output.Message)
	}
	if output.Data["key"] != "value" {
		t.Errorf("expected data.key='value', got '%v'", output.Data["key"])
	}
}

/* ------------------------------------------------------------------------- */
/* TABLE-DRIVEN TESTS                                                        */
/* ------------------------------------------------------------------------- */

func TestScriptExecutor_Execute_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		scriptContent string
		permissions   os.FileMode
		input         HookInput
		timeout       time.Duration
		wantErr       bool
		wantSuccess   bool
		wantMessage   string
	}{
		{
			name: "successful execution with message",
			scriptContent: `#!/bin/sh
read input
echo '{"success": true, "message": "Hook executed successfully"}'
`,
			permissions: 0755,
			input: HookInput{
				Hook:        "pre-bump",
				Version:     "1.2.3",
				BumpType:    "patch",
				ProjectRoot: "/test/project",
			},
			timeout:     DefaultTimeout,
			wantErr:     false,
			wantSuccess: true,
			wantMessage: "Hook executed successfully",
		},
		{
			name: "successful execution with data",
			scriptContent: `#!/bin/sh
read input
echo '{"success": true, "message": "Data returned", "data": {"version": "1.0.0"}}'
`,
			permissions: 0755,
			input: HookInput{
				Hook:        "post-bump",
				Version:     "1.0.0",
				ProjectRoot: "/test",
			},
			timeout:     DefaultTimeout,
			wantErr:     false,
			wantSuccess: true,
			wantMessage: "Data returned",
		},
		{
			name: "script reports failure",
			scriptContent: `#!/bin/sh
read input
echo '{"success": false, "message": "Validation failed"}'
`,
			permissions: 0755,
			input: HookInput{
				Hook:        "validate",
				Version:     "2.0.0",
				ProjectRoot: "/test",
			},
			timeout:     DefaultTimeout,
			wantErr:     true,
			wantSuccess: false,
			wantMessage: "Validation failed",
		},
		{
			name: "script exits with error",
			scriptContent: `#!/bin/sh
read input
echo "error on stderr" >&2
exit 1
`,
			permissions: 0755,
			input: HookInput{
				Hook:    "pre-bump",
				Version: "1.0.0",
			},
			timeout: DefaultTimeout,
			wantErr: true,
		},
		{
			name: "invalid JSON output",
			scriptContent: `#!/bin/sh
read input
echo 'not valid json at all'
`,
			permissions: 0755,
			input: HookInput{
				Hook:    "pre-bump",
				Version: "1.0.0",
			},
			timeout: DefaultTimeout,
			wantErr: true,
		},
		{
			name: "script with prerelease and metadata",
			scriptContent: `#!/bin/sh
read input
echo '{"success": true, "message": "Release processed"}'
`,
			permissions: 0755,
			input: HookInput{
				Hook:            "post-bump",
				Version:         "2.0.0",
				PreviousVersion: "1.9.9",
				BumpType:        "major",
				Prerelease:      stringPtr("rc.1"),
				Metadata:        stringPtr("build.123"),
				ProjectRoot:     "/project",
			},
			timeout:     DefaultTimeout,
			wantErr:     false,
			wantSuccess: true,
			wantMessage: "Release processed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			scriptPath := filepath.Join(tmpDir, "test-script.sh")

			if err := os.WriteFile(scriptPath, []byte(tt.scriptContent), tt.permissions); err != nil {
				t.Fatalf("failed to create test script: %v", err)
			}

			executor := NewScriptExecutorWithTimeout(tt.timeout)
			ctx := context.Background()

			output, err := executor.Execute(ctx, scriptPath, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if output.Success != tt.wantSuccess {
					t.Errorf("Expected success=%v, got %v", tt.wantSuccess, output.Success)
				}
				if tt.wantMessage != "" && output.Message != tt.wantMessage {
					t.Errorf("Expected message=%q, got %q", tt.wantMessage, output.Message)
				}
			}
		})
	}
}

func TestScriptExecutor_Execute_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setupScript func(t *testing.T) string
		input       HookInput
		timeout     time.Duration
		wantErrText string
	}{
		{
			name: "nonexistent script path",
			setupScript: func(t *testing.T) string {
				return "/nonexistent/path/script.sh"
			},
			input: HookInput{
				Hook:    "pre-bump",
				Version: "1.0.0",
			},
			timeout:     DefaultTimeout,
			wantErrText: "script not found",
		},
		{
			name: "script path is directory",
			setupScript: func(t *testing.T) string {
				tmpDir := t.TempDir()
				scriptDir := filepath.Join(tmpDir, "script-dir")
				if err := os.MkdirAll(scriptDir, 0755); err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				return scriptDir
			},
			input: HookInput{
				Hook:    "pre-bump",
				Version: "1.0.0",
			},
			timeout:     DefaultTimeout,
			wantErrText: "script path is a directory",
		},
		{
			name: "script not executable",
			setupScript: func(t *testing.T) string {
				tmpDir := t.TempDir()
				scriptPath := filepath.Join(tmpDir, "non-exec.sh")
				script := `#!/bin/sh
echo '{"success": true}'
`
				if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
					t.Fatalf("failed to create script: %v", err)
				}
				return scriptPath
			},
			input: HookInput{
				Hook:    "pre-bump",
				Version: "1.0.0",
			},
			timeout:     DefaultTimeout,
			wantErrText: "script is not executable",
		},
		{
			name: "script timeout",
			setupScript: func(t *testing.T) string {
				tmpDir := t.TempDir()
				scriptPath := filepath.Join(tmpDir, "slow.sh")
				script := `#!/bin/sh
sleep 10
echo '{"success": true}'
`
				if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
					t.Fatalf("failed to create script: %v", err)
				}
				return scriptPath
			},
			input: HookInput{
				Hook:    "pre-bump",
				Version: "1.0.0",
			},
			timeout:     100 * time.Millisecond,
			wantErrText: "timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := tt.setupScript(t)
			executor := NewScriptExecutorWithTimeout(tt.timeout)
			ctx := context.Background()

			_, err := executor.Execute(ctx, scriptPath, tt.input)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.wantErrText != "" && !contains(err.Error(), tt.wantErrText) {
				t.Errorf("Expected error containing %q, got %q", tt.wantErrText, err.Error())
			}
		})
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestScriptExecutor_Execute_DirectoryAsScript tests that executor rejects directories
func TestScriptExecutor_Execute_DirectoryAsScript(t *testing.T) {
	tmpDir := t.TempDir()
	scriptDir := filepath.Join(tmpDir, "not-a-script")
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	executor := NewScriptExecutor()
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	ctx := context.Background()
	_, err := executor.Execute(ctx, scriptDir, input)
	if err == nil {
		t.Fatal("expected error when script path is a directory")
	}
	if !contains(err.Error(), "script path is a directory") {
		t.Errorf("expected error about directory, got: %v", err)
	}
}

// TestScriptExecutor_Execute_OutputTooLarge tests the max output size limit
func TestScriptExecutor_Execute_OutputTooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "large-output.sh")

	// Generate output larger than MaxOutputSize (1MB)
	script := `#!/bin/sh
read input
# Generate ~2MB of output
head -c 2000000 /dev/zero | base64
echo '{"success": true}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewScriptExecutor()
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	ctx := context.Background()
	_, err := executor.Execute(ctx, scriptPath, input)
	if err == nil {
		t.Fatal("expected error for output exceeding max size")
	}
	if !contains(err.Error(), "exceeds maximum size") {
		t.Errorf("expected error about max size, got: %v", err)
	}
}

// TestScriptExecutor_Execute_ContextCancellation tests context cancellation
func TestScriptExecutor_Execute_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "slow.sh")

	script := `#!/bin/sh
sleep 5
echo '{"success": true}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewScriptExecutor()
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := executor.Execute(ctx, scriptPath, input)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// TestScriptExecutor_Execute_EmptyOutput tests script with no JSON output
func TestScriptExecutor_Execute_EmptyOutput(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "empty.sh")

	// Script outputs nothing and exits successfully
	script := `#!/bin/sh
read input
# Output nothing to stdout, exit cleanly
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewScriptExecutor()
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	ctx := context.Background()
	_, err := executor.Execute(ctx, scriptPath, input)
	if err == nil {
		t.Fatal("expected error for empty output")
	}
	// The error should be about JSON parsing since script exits cleanly but has no output
	if !contains(err.Error(), "failed to parse") && !contains(err.Error(), "unexpected end of JSON input") {
		t.Errorf("expected JSON parse error, got: %v", err)
	}
}

// TestScriptExecutor_Execute_RelativePath tests relative script paths are resolved
func TestScriptExecutor_Execute_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "hook.sh")

	script := `#!/bin/sh
read input
echo '{"success": true, "message": "Relative path works"}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create test script: %v", err)
	}

	executor := NewScriptExecutor()
	input := HookInput{
		Hook:        "pre-bump",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	// Change to tmpDir and use relative path
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()
	output, err := executor.Execute(ctx, "./hook.sh", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true")
	}
}

// TestHookInput_NilPointers tests HookInput with nil optional fields
func TestHookInput_NilPointers(t *testing.T) {
	input := HookInput{
		Hook:            "validate",
		Version:         "1.0.0",
		PreviousVersion: "",
		BumpType:        "",
		Prerelease:      nil,
		Metadata:        nil,
		ProjectRoot:     "/test",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal input: %v", err)
	}

	var decoded HookInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal input: %v", err)
	}

	if decoded.Prerelease != nil {
		t.Error("expected nil prerelease")
	}
	if decoded.Metadata != nil {
		t.Error("expected nil metadata")
	}
}

// TestHookOutput_EmptyData tests HookOutput with nil data field
func TestHookOutput_EmptyData(t *testing.T) {
	output := HookOutput{
		Success: true,
		Message: "No data",
		Data:    nil,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal output: %v", err)
	}

	var decoded HookOutput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if !decoded.Success {
		t.Error("expected success=true")
	}
	if decoded.Message != "No data" {
		t.Errorf("expected message 'No data', got '%s'", decoded.Message)
	}
}

/* ------------------------------------------------------------------------- */
/* PATH TRAVERSAL SECURITY TESTS                                             */
/* ------------------------------------------------------------------------- */

// TestExecuteExtensionHook_PathTraversal tests that path traversal attempts are blocked
func TestExecuteExtensionHook_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a secret file outside the extension directory
	secretDir := filepath.Join(tmpDir, "secret")
	if err := os.MkdirAll(secretDir, 0755); err != nil {
		t.Fatalf("failed to create secret directory: %v", err)
	}
	secretFile := filepath.Join(secretDir, "sensitive.sh")
	script := `#!/bin/sh
read input
echo '{"success": true, "message": "Accessed secret file"}'
`
	if err := os.WriteFile(secretFile, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create secret file: %v", err)
	}

	// Create extension directory
	extDir := filepath.Join(tmpDir, "extension")
	if err := os.MkdirAll(extDir, 0755); err != nil {
		t.Fatalf("failed to create extension directory: %v", err)
	}

	input := HookInput{
		Hook:        "post-bump",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	ctx := context.Background()

	tests := []struct {
		name        string
		entryPoint  string
		wantErrText string
	}{
		{
			name:        "path traversal with ..",
			entryPoint:  "../secret/sensitive.sh",
			wantErrText: "invalid script path",
		},
		{
			name:        "path traversal with multiple ..",
			entryPoint:  "../../etc/passwd",
			wantErrText: "invalid script path",
		},
		{
			name:        "path traversal attempt to parent",
			entryPoint:  "../sensitive.sh",
			wantErrText: "invalid script path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ExecuteExtensionHook(ctx, extDir, tt.entryPoint, input)

			if err == nil {
				t.Fatal("expected error for path traversal attempt, got nil")
			}

			if !contains(err.Error(), tt.wantErrText) {
				t.Errorf("expected error containing %q, got %q", tt.wantErrText, err.Error())
			}
		})
	}
}

// TestExecuteExtensionHook_ValidPaths tests that valid paths within extension dir work
func TestExecuteExtensionHook_ValidPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension directory with nested structure
	extDir := filepath.Join(tmpDir, "extension")
	hooksDir := filepath.Join(extDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("failed to create hooks directory: %v", err)
	}

	// Create script in root of extension
	rootScript := filepath.Join(extDir, "hook.sh")
	script := `#!/bin/sh
read input
echo '{"success": true, "message": "Valid path"}'
`
	if err := os.WriteFile(rootScript, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create root script: %v", err)
	}

	// Create script in subdirectory
	nestedScript := filepath.Join(hooksDir, "nested.sh")
	if err := os.WriteFile(nestedScript, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create nested script: %v", err)
	}

	input := HookInput{
		Hook:        "post-bump",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	ctx := context.Background()

	tests := []struct {
		name       string
		entryPoint string
		wantErr    bool
	}{
		{
			name:       "script in extension root",
			entryPoint: "hook.sh",
			wantErr:    false,
		},
		{
			name:       "script in subdirectory",
			entryPoint: "hooks/nested.sh",
			wantErr:    false,
		},
		{
			name:       "script with ./prefix",
			entryPoint: "./hook.sh",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ExecuteExtensionHook(ctx, extDir, tt.entryPoint, input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !output.Success {
					t.Error("expected success=true")
				}
			}
		})
	}
}

// TestExecuteExtensionHook_PathCleaning tests that paths are properly cleaned
func TestExecuteExtensionHook_PathCleaning(t *testing.T) {
	tmpDir := t.TempDir()
	extDir := filepath.Join(tmpDir, "extension")
	if err := os.MkdirAll(extDir, 0755); err != nil {
		t.Fatalf("failed to create extension directory: %v", err)
	}

	scriptPath := filepath.Join(extDir, "hook.sh")
	script := `#!/bin/sh
read input
echo '{"success": true, "message": "Path cleaned"}'
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	input := HookInput{
		Hook:        "validate",
		Version:     "1.0.0",
		ProjectRoot: "/test",
	}

	ctx := context.Background()

	// Path with redundant ./ should be cleaned and work
	output, err := ExecuteExtensionHook(ctx, extDir, "./././hook.sh", input)
	if err != nil {
		t.Fatalf("unexpected error for cleaned path: %v", err)
	}
	if !output.Success {
		t.Error("expected success=true for cleaned path")
	}
}
