package extensionmgr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/pathutil"
)

// HookInput represents the JSON input passed to an extension script
type HookInput struct {
	Hook            string  `json:"hook"`
	Version         string  `json:"version"`
	PreviousVersion string  `json:"previous_version,omitempty"`
	BumpType        string  `json:"bump_type,omitempty"`
	Prerelease      *string `json:"prerelease,omitempty"`
	Metadata        *string `json:"metadata,omitempty"`
	ProjectRoot     string  `json:"project_root"`
	ModuleDir       string  `json:"module_dir,omitempty"`  // Directory containing the .version file (for monorepo support)
	ModuleName      string  `json:"module_name,omitempty"` // Module identifier (for monorepo support)
}

// HookOutput represents the JSON output expected from an extension script
type HookOutput struct {
	Success bool           `json:"success"`
	Message string         `json:"message,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

// Executor defines the interface for executing extension scripts
type Executor interface {
	Execute(ctx context.Context, scriptPath string, input *HookInput) (*HookOutput, error)
}

// ScriptExecutor implements the Executor interface for running shell scripts
type ScriptExecutor struct {
	Timeout time.Duration
}

// DefaultTimeout is the default execution timeout for extension scripts.
// Uses core.TimeoutDefault for consistency across the codebase.
const DefaultTimeout = core.TimeoutDefault

// MaxOutputSize limits the amount of data read from script stdout/stderr
const MaxOutputSize = 1024 * 1024 // 1MB

// NewScriptExecutor creates a new ScriptExecutor with the default timeout
func NewScriptExecutor() *ScriptExecutor {
	return &ScriptExecutor{
		Timeout: DefaultTimeout,
	}
}

// NewScriptExecutorWithTimeout creates a new ScriptExecutor with a custom timeout
func NewScriptExecutorWithTimeout(timeout time.Duration) *ScriptExecutor {
	return &ScriptExecutor{
		Timeout: timeout,
	}
}

// Execute runs an extension script with the provided input and returns the output
func (e *ScriptExecutor) Execute(ctx context.Context, scriptPath string, input *HookInput) (*HookOutput, error) {
	// Clean and validate script path to prevent path traversal attacks
	cleanPath := filepath.Clean(scriptPath)

	// Resolve to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve script path %s: %w", scriptPath, err)
	}

	// Check if script exists and is executable
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("script not found at %s: %w", absPath, err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("script path is a directory: %s", absPath)
	}

	// Check if file is executable (Unix-like systems)
	if info.Mode()&core.PermExecutable == 0 {
		return nil, fmt.Errorf("script is not executable: %s", absPath)
	}

	// Serialize input to JSON
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input: %w", err)
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()

	// Prepare command
	cmd := exec.CommandContext(execCtx, absPath)
	cmd.Stdin = bytes.NewReader(inputJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the script
	if err := cmd.Run(); err != nil {
		// Check if it was a timeout
		if execCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("script execution timeout after %v: %s", e.Timeout, stderr.String())
		}

		// Return execution error with stderr
		return nil, fmt.Errorf("script execution failed: %w\nstderr: %s", err, stderr.String())
	}

	// Check output size
	if stdout.Len() > MaxOutputSize {
		return nil, fmt.Errorf("script output exceeds maximum size of %d bytes", MaxOutputSize)
	}

	// Parse output JSON
	var output HookOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, fmt.Errorf("failed to parse script output as JSON: %w\noutput: %s", err, stdout.String())
	}

	// Check if script reported success
	if !output.Success {
		return &output, fmt.Errorf("script reported failure: %s", output.Message)
	}

	return &output, nil
}

// ExecuteExtensionHook is a convenience function to execute an extension hook
// It resolves the full script path relative to the extension directory and validates
// that the script remains within the extension directory to prevent path traversal attacks
func ExecuteExtensionHook(ctx context.Context, extensionPath, entry string, input *HookInput) (*HookOutput, error) {
	executor := NewScriptExecutor()

	// Validate that the entry point doesn't attempt path traversal
	// This prevents malicious extensions from accessing files outside their directory
	scriptPath := filepath.Join(extensionPath, entry)

	// Validate the script path is within the extension directory
	if _, err := pathutil.ValidatePath(scriptPath, extensionPath); err != nil {
		return nil, fmt.Errorf("invalid script path: %w", err)
	}

	return executor.Execute(ctx, scriptPath, input)
}
