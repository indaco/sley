package hooks

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/indaco/sley/internal/testutils"
)

/* ------------------------------------------------------------------------- */
/* MOCK IMPLEMENTATIONS FOR TESTING                                          */
/* ------------------------------------------------------------------------- */

// mockHookProvider implements HookProvider for testing.
type mockHookProvider struct {
	hooks []PreReleaseHook
}

func (m *mockHookProvider) GetHooks() []PreReleaseHook {
	return m.hooks
}

// mockPrinter implements OutputPrinter for testing (captures output).
type mockPrinter struct {
	output []string
}

func (m *mockPrinter) Printf(format string, args ...any) {
	// Capture but don't actually print
}

func (m *mockPrinter) PrintSuccess(msg string) {
	m.output = append(m.output, "SUCCESS:"+msg)
}

func (m *mockPrinter) PrintFailure(msg string) {
	m.output = append(m.output, "FAILURE:"+msg)
}

/* ------------------------------------------------------------------------- */
/* TESTS                                                                     */
/* ------------------------------------------------------------------------- */

func TestRunPreReleaseHooks(t *testing.T) {
	tests := []struct {
		name       string
		skip       bool
		hooks      []PreReleaseHook
		wantErr    bool
		errMessage string
	}{
		{
			name:    "skip true",
			skip:    true,
			hooks:   []PreReleaseHook{testutils.MockHook{Name: "hook1"}},
			wantErr: false,
		},
		{
			name:    "no hooks",
			skip:    false,
			hooks:   nil,
			wantErr: false,
		},
		{
			name: "all hooks succeed",
			skip: false,
			hooks: []PreReleaseHook{
				testutils.MockHook{Name: "hook1"},
				testutils.MockHook{Name: "hook2"},
			},
			wantErr: false,
		},
		{
			name: "hook fails",
			skip: false,
			hooks: []PreReleaseHook{
				testutils.MockHook{Name: "hook1"},
				testutils.MockHook{Name: "bad-hook", ShouldErr: true},
				testutils.MockHook{Name: "hook2"}, // Should not be called after failure
			},
			wantErr:    true,
			errMessage: `pre-release hook "bad-hook" failed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetPreReleaseHooks()
			t.Cleanup(func() { ResetPreReleaseHooks() })
			for _, h := range tt.hooks {
				RegisterPreReleaseHook(h)
			}

			ctx := context.Background()
			err := runPreReleaseHooks(ctx, tt.skip)

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got error=%v", tt.wantErr, err != nil)
			}

			if tt.wantErr && !errors.Is(err, tt.hooks[1].Run(ctx)) {
				// error wrapping check
				if err == nil || !containsError(err.Error(), tt.errMessage) {
					t.Errorf("expected error message to contain %q, got: %v", tt.errMessage, err)
				}
			}
		})
	}
}

func TestPreReleaseHookRunner_WithMocks(t *testing.T) {
	tests := []struct {
		name       string
		skip       bool
		hooks      []PreReleaseHook
		wantErr    bool
		errMessage string
	}{
		{
			name:    "skip true",
			skip:    true,
			hooks:   []PreReleaseHook{testutils.MockHook{Name: "hook1"}},
			wantErr: false,
		},
		{
			name:    "no hooks",
			skip:    false,
			hooks:   nil,
			wantErr: false,
		},
		{
			name: "all hooks succeed",
			skip: false,
			hooks: []PreReleaseHook{
				testutils.MockHook{Name: "hook1"},
				testutils.MockHook{Name: "hook2"},
			},
			wantErr: false,
		},
		{
			name: "hook fails",
			skip: false,
			hooks: []PreReleaseHook{
				testutils.MockHook{Name: "hook1"},
				testutils.MockHook{Name: "bad-hook", ShouldErr: true},
			},
			wantErr:    true,
			errMessage: `pre-release hook "bad-hook" failed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use mock provider and printer (interface-based DI)
			provider := &mockHookProvider{hooks: tt.hooks}
			printer := &mockPrinter{}
			runner := NewPreReleaseHookRunner(provider, printer)

			ctx := context.Background()
			err := runner.Run(ctx, tt.skip)

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error=%v, got error=%v", tt.wantErr, err != nil)
			}

			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errMessage) {
					t.Errorf("expected error message to contain %q, got: %v", tt.errMessage, err)
				}
			}
		})
	}
}

func TestNewPreReleaseHookRunner_Defaults(t *testing.T) {
	runner := NewPreReleaseHookRunner(nil, nil)
	if runner == nil {
		t.Fatal("NewPreReleaseHookRunner returned nil")
	}
	if runner.provider == nil {
		t.Error("provider should not be nil")
	}
	if runner.printer == nil {
		t.Error("printer should not be nil")
	}
}

func containsError(got, want string) bool {
	return got != "" && want != "" && (got == want || (len(got) > len(want) && got[:len(want)] == want) || (len(want) > len(got) && want[:len(got)] == got))
}
