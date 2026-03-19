package git

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

var fakeRefCommands = map[string]string{}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cmdStr := command + " " + strings.Join(args, " ")
	cmd := exec.Command(os.Args[0], "-test.run=TestRefHelperProcess", "--", cmdStr) //nolint:gosec // standard test re-exec pattern

	cmd.Env = append(os.Environ(),
		"GO_TEST_HELPER_PROCESS=1",
		"MOCK_KEY="+cmdStr,
		"MOCK_VAL="+fakeRefCommands[cmdStr],
	)

	return cmd
}

func TestRefHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_HELPER_PROCESS") != "1" {
		return
	}

	key := os.Getenv("MOCK_KEY")
	val := os.Getenv("MOCK_VAL")

	_ = key

	if val == "ERROR" {
		_, _ = os.Stderr.WriteString("mock failure")
		os.Exit(1)
	}

	_, _ = os.Stdout.WriteString(val)
	os.Exit(0)
}

func TestSafeFallbackSince(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		mocks    map[string]string
		expected string
	}{
		{
			name: "enough commits returns HEAD~n",
			n:    10,
			mocks: map[string]string{
				"git rev-list --count HEAD": "25",
			},
			expected: "HEAD~10",
		},
		{
			name: "fewer commits returns root hash",
			n:    10,
			mocks: map[string]string{
				"git rev-list --count HEAD":         "2",
				"git rev-list --max-parents=0 HEAD": "abc123def456",
			},
			expected: "abc123def456",
		},
		{
			name: "exactly n+1 commits returns HEAD~n",
			n:    10,
			mocks: map[string]string{
				"git rev-list --count HEAD": "11",
			},
			expected: "HEAD~10",
		},
		{
			name: "exactly n commits returns root hash",
			n:    10,
			mocks: map[string]string{
				"git rev-list --count HEAD":         "10",
				"git rev-list --max-parents=0 HEAD": "rootabc",
			},
			expected: "rootabc",
		},
		{
			name: "rev-list count fails returns HEAD~n",
			n:    10,
			mocks: map[string]string{
				"git rev-list --count HEAD": "ERROR",
			},
			expected: "HEAD~10",
		},
		{
			name: "root commit lookup fails returns HEAD~n",
			n:    10,
			mocks: map[string]string{
				"git rev-list --count HEAD":         "2",
				"git rev-list --max-parents=0 HEAD": "ERROR",
			},
			expected: "HEAD~10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeRefCommands = tt.mocks

			result := SafeFallbackSince(fakeExecCommand, tt.n)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
