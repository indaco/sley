package hooks

import (
	"strings"
	"testing"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/testutils"
)

func TestLoadPreReleaseHooksFromConfig(t *testing.T) {
	ResetPreReleaseHooks()
	t.Cleanup(func() { ResetPreReleaseHooks() })

	cfg := &config.Config{
		PreReleaseHooks: []map[string]config.PreReleaseHookConfig{
			{
				"run-tests": {Command: "go test ./..."},
			},
			{
				"check-license": {Command: "./scripts/check_license.sh"},
			},
		},
	}

	err := LoadPreReleaseHooksFromConfig(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	hooks := GetPreReleaseHooks()
	if len(hooks) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(hooks))
	}

	if hooks[0].HookName() != "run-tests" {
		t.Errorf("expected first hook 'run-tests', got %q", hooks[0].HookName())
	}
	if hooks[1].HookName() != "check-license" {
		t.Errorf("expected second hook 'check-license', got %q", hooks[1].HookName())
	}
}

func TestLoadPreReleaseHooksFromConfig_NilConfig(t *testing.T) {
	ResetPreReleaseHooks()
	t.Cleanup(func() { ResetPreReleaseHooks() })

	err := LoadPreReleaseHooksFromConfig(nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	hooks := GetPreReleaseHooks()
	if len(hooks) != 0 {
		t.Errorf("expected 0 hooks, got %d", len(hooks))
	}
}

func TestLoadPreReleaseHooksFromConfig_NilPreReleaseHooks(t *testing.T) {
	ResetPreReleaseHooks()
	t.Cleanup(func() { ResetPreReleaseHooks() })

	cfg := &config.Config{
		PreReleaseHooks: nil,
	}

	err := LoadPreReleaseHooksFromConfig(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	hooks := GetPreReleaseHooks()
	if len(hooks) != 0 {
		t.Errorf("expected 0 hooks, got %d", len(hooks))
	}
}

func TestLoadPreReleaseHooksFromConfig_SkipMissingCommand(t *testing.T) {
	ResetPreReleaseHooks()
	t.Cleanup(func() { ResetPreReleaseHooks() })

	cfg := &config.Config{
		PreReleaseHooks: []map[string]config.PreReleaseHookConfig{
			{
				"no-command": {}, // Missing Command field
			},
		},
	}

	output, err := testutils.CaptureStdout(func() {
		err := LoadPreReleaseHooksFromConfig(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "⚠️  Skipping pre-release hook \"no-command\": no command defined") {
		t.Errorf("expected warning output about missing command, got: %q", output)
	}

	hooks := GetPreReleaseHooks()
	if len(hooks) != 0 {
		t.Errorf("expected 0 hooks registered, got %d", len(hooks))
	}
}
