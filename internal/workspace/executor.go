package workspace

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Operation represents a single operation that can be executed on a module.
// Operations are the unit of work in the executor system.
type Operation interface {
	// Execute runs the operation on the given module.
	// It should return an error if the operation fails.
	Execute(ctx context.Context, mod *Module) error

	// Name returns a human-readable name for this operation.
	// Used for logging and display purposes.
	Name() string
}

// ExecutionResult represents the result of executing an operation on a module.
type ExecutionResult struct {
	// Module is the module that was operated on.
	Module *Module

	// OldVersion is the version before the operation (if applicable).
	OldVersion string

	// NewVersion is the version after the operation (if applicable).
	NewVersion string

	// Success indicates whether the operation succeeded.
	Success bool

	// Error contains the error if the operation failed.
	Error error

	// Duration is how long the operation took.
	Duration time.Duration
}

// ExecutorOption configures an Executor.
type ExecutorOption func(*Executor)

// WithParallel enables parallel execution of operations.
func WithParallel(parallel bool) ExecutorOption {
	return func(e *Executor) {
		e.parallel = parallel
	}
}

// WithFailFast causes execution to stop on the first error.
func WithFailFast(failFast bool) ExecutorOption {
	return func(e *Executor) {
		e.failFast = failFast
	}
}

// Executor executes operations on multiple modules.
// It supports both sequential and parallel execution with error handling strategies.
type Executor struct {
	parallel bool
	failFast bool
}

// NewExecutor creates a new Executor with the given options.
func NewExecutor(opts ...ExecutorOption) *Executor {
	e := &Executor{
		parallel: false,
		failFast: false,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Run executes the given operation on all modules.
// It returns a slice of ExecutionResults, one for each module.
// The context can be used for cancellation.
func (e *Executor) Run(ctx context.Context, modules []*Module, op Operation) ([]ExecutionResult, error) {
	if len(modules) == 0 {
		return nil, fmt.Errorf("no modules provided")
	}

	if op == nil {
		return nil, fmt.Errorf("operation is nil")
	}

	if e.parallel {
		return e.runParallel(ctx, modules, op)
	}

	return e.runSequential(ctx, modules, op)
}

// runSequential executes the operation on each module sequentially.
func (e *Executor) runSequential(ctx context.Context, modules []*Module, op Operation) ([]ExecutionResult, error) {
	results := make([]ExecutionResult, 0, len(modules))

	for _, mod := range modules {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		result := e.executeOperation(ctx, mod, op)
		results = append(results, result)

		// If fail-fast is enabled and this operation failed, stop
		if e.failFast && !result.Success {
			return results, fmt.Errorf("operation failed on module %s: %w", mod.Name, result.Error)
		}
	}

	return results, nil
}

// runParallel executes the operation on all modules in parallel.
func (e *Executor) runParallel(ctx context.Context, modules []*Module, op Operation) ([]ExecutionResult, error) {
	results := make([]ExecutionResult, len(modules))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error
	var failed atomic.Bool

	// Create a context that we can cancel if fail-fast is triggered
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, mod := range modules {
		wg.Add(1)

		go func(idx int, module *Module) {
			defer wg.Done()

			// Check if we should skip due to fail-fast (lock-free)
			if e.failFast && failed.Load() {
				return
			}

			result := e.executeOperation(execCtx, module, op)

			results[idx] = result
			if e.failFast && !result.Success && failed.CompareAndSwap(false, true) {
				mu.Lock()
				firstError = result.Error
				mu.Unlock()
				cancel() // Cancel all other operations
			}
		}(i, mod)
	}

	wg.Wait()

	if e.failFast && firstError != nil {
		return results, fmt.Errorf("operation failed: %w", firstError)
	}

	return results, nil
}

// executeOperation runs the operation on a single module and captures the result.
func (e *Executor) executeOperation(ctx context.Context, mod *Module, op Operation) ExecutionResult {
	start := time.Now()

	// Store the old version before the operation
	oldVersion := mod.CurrentVersion

	// Execute the operation
	err := op.Execute(ctx, mod)

	// Calculate duration
	duration := time.Since(start)

	// Build the result
	result := ExecutionResult{
		Module:     mod,
		OldVersion: oldVersion,
		NewVersion: mod.CurrentVersion, // Operation may have updated this
		Success:    err == nil,
		Error:      err,
		Duration:   duration,
	}

	return result
}

// HasErrors returns true if any of the results contain errors.
func HasErrors(results []ExecutionResult) bool {
	for _, result := range results {
		if !result.Success {
			return true
		}
	}
	return false
}

// SuccessCount returns the number of successful results.
func SuccessCount(results []ExecutionResult) int {
	count := 0
	for _, result := range results {
		if result.Success {
			count++
		}
	}
	return count
}

// ErrorCount returns the number of failed results.
func ErrorCount(results []ExecutionResult) int {
	count := 0
	for _, result := range results {
		if !result.Success {
			count++
		}
	}
	return count
}

// TotalDuration returns the sum of all result durations.
func TotalDuration(results []ExecutionResult) time.Duration {
	var total time.Duration
	for _, result := range results {
		total += result.Duration
	}
	return total
}
