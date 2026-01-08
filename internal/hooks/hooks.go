package hooks

import (
	"fmt"
	"sync"

	"github.com/indaco/sley/internal/console"
)

type PreReleaseHook interface {
	HookName() string
	Run() error
}

var (
	hooksMu              sync.RWMutex
	preReleaseHooks      []PreReleaseHook
	RunPreReleaseHooksFn = runPreReleaseHooks
)

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

func runPreReleaseHooks(skip bool) error {
	if skip {
		return nil
	}

	for _, hook := range GetPreReleaseHooks() {
		fmt.Printf("Running pre-release hook: %s... ", hook.HookName())
		if err := hook.Run(); err != nil {
			console.PrintFailure("FAIL")
			return fmt.Errorf("pre-release hook %q failed: %w", hook.HookName(), err)
		}
		console.PrintSuccess("OK")
	}

	return nil
}
