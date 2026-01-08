package hooks

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/indaco/sley/internal/console"
)

// PreReleaseHook is the interface for pre-release hooks.
type PreReleaseHook interface {
	HookName() string
	Run(ctx context.Context) error
}

// HookProvider provides access to registered hooks.
type HookProvider interface {
	GetHooks() []PreReleaseHook
}

// OutputPrinter abstracts output printing for testability.
type OutputPrinter interface {
	Printf(format string, args ...any)
	PrintSuccess(msg string)
	PrintFailure(msg string)
}

// PreReleaseHookRunner runs pre-release hooks with injected dependencies.
type PreReleaseHookRunner struct {
	provider HookProvider
	printer  OutputPrinter
}

// defaultHookProvider is the default implementation using global state.
type defaultHookProvider struct{}

func (p *defaultHookProvider) GetHooks() []PreReleaseHook {
	return GetPreReleaseHooks()
}

// defaultPrinter is the production implementation of OutputPrinter.
type defaultPrinter struct {
	out io.Writer
}

func (p *defaultPrinter) Printf(format string, args ...any) {
	fmt.Fprintf(p.out, format, args...)
}

func (p *defaultPrinter) PrintSuccess(msg string) {
	console.PrintSuccess(msg)
}

func (p *defaultPrinter) PrintFailure(msg string) {
	console.PrintFailure(msg)
}

// NewPreReleaseHookRunner creates a PreReleaseHookRunner with the given dependencies.
// If any dependency is nil, the production default is used.
func NewPreReleaseHookRunner(provider HookProvider, printer OutputPrinter) *PreReleaseHookRunner {
	if provider == nil {
		provider = &defaultHookProvider{}
	}
	if printer == nil {
		printer = &defaultPrinter{out: os.Stdout}
	}
	return &PreReleaseHookRunner{
		provider: provider,
		printer:  printer,
	}
}

// Run executes all registered pre-release hooks.
func (r *PreReleaseHookRunner) Run(ctx context.Context, skip bool) error {
	if skip {
		return nil
	}

	for _, hook := range r.provider.GetHooks() {
		r.printer.Printf("Running pre-release hook: %s... ", hook.HookName())
		if err := hook.Run(ctx); err != nil {
			r.printer.PrintFailure("FAIL")
			return fmt.Errorf("pre-release hook %q failed: %w", hook.HookName(), err)
		}
		r.printer.PrintSuccess("OK")
	}

	return nil
}

// Global hook registry (kept for backward compatibility)
var (
	hooksMu         sync.RWMutex
	preReleaseHooks []PreReleaseHook
)

// defaultRunner is the default PreReleaseHookRunner for backward compatibility.
var defaultRunner = NewPreReleaseHookRunner(nil, nil)

// RunPreReleaseHooksFn is kept for backward compatibility during migration.
var RunPreReleaseHooksFn = func(ctx context.Context, skip bool) error {
	return defaultRunner.Run(ctx, skip)
}

func RegisterPreReleaseHook(h PreReleaseHook) {
	hooksMu.Lock()
	defer hooksMu.Unlock()
	preReleaseHooks = append(preReleaseHooks, h)
}

func GetPreReleaseHooks() []PreReleaseHook {
	hooksMu.RLock()
	defer hooksMu.RUnlock()
	// Return a copy to prevent external modification
	hooks := make([]PreReleaseHook, len(preReleaseHooks))
	copy(hooks, preReleaseHooks)
	return hooks
}

func ResetPreReleaseHooks() {
	hooksMu.Lock()
	defer hooksMu.Unlock()
	preReleaseHooks = nil
}

// runPreReleaseHooks is kept for direct testing.
func runPreReleaseHooks(ctx context.Context, skip bool) error {
	return defaultRunner.Run(ctx, skip)
}
