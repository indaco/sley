// Package cmdrunner provides context-aware external command execution.
//
// This package wraps os/exec to provide command execution with proper
// context support for cancellation and timeouts. All commands are
// executed with structured error handling via apperrors.CommandError.
//
// # Functions
//
// RunCommandContext executes a command with stdout/stderr connected to
// the terminal, suitable for interactive commands or those with visible
// output:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	err := cmdrunner.RunCommandContext(ctx, "/path/to/dir", "git", "status")
//
// RunCommandOutputContext executes a command and captures its combined
// output, suitable for commands whose output needs to be processed:
//
//	output, err := cmdrunner.RunCommandOutputContext(ctx, ".", "git", "rev-parse", "HEAD")
//
// # Timeout Handling
//
// The package defines default timeout constants:
//   - DefaultTimeout: 30 seconds for general commands
//   - DefaultOutputTimeout: 5 seconds for output-capturing commands
//
// When a command times out, the returned error will have Timeout set to
// true in the apperrors.CommandError struct.
package cmdrunner
