package gitlog

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

var fakeGitCommands = map[string]string{}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cmdStr := command + " " + strings.Join(args, " ")
	// println("[fakeExecCommand] registering mock:", cmdStr)
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", cmdStr) //nolint:gosec // G702: standard test re-exec pattern

	cmd.Env = append(os.Environ(),
		"GO_TEST_HELPER_PROCESS=1",
		"MOCK_KEY="+cmdStr,
		"MOCK_VAL="+fakeGitCommands[cmdStr],
	)

	return cmd
}

// Simulated process that prints predefined output.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_HELPER_PROCESS") != "1" {
		return
	}

	key := os.Getenv("MOCK_KEY")
	val := os.Getenv("MOCK_VAL")

	os.Stderr.WriteString("[mock] looking for: " + key + "\n")

	if val == "ERROR" {
		_, _ = os.Stderr.WriteString("mock log failure")
		os.Exit(1)
	}

	// Always write MOCK_VAL, even if empty
	_, _ = os.Stdout.WriteString(val)
	os.Exit(0)
}

func stubExecCommand() func() {
	orig := execCommand
	execCommand = fakeExecCommand
	return func() {
		execCommand = orig
	}
}

func TestGetCommits(t *testing.T) {
	restore := stubExecCommand()
	defer restore()

	tests := []struct {
		name            string
		since           string
		until           string
		mockGitCommands map[string]string
		expectedCommits []string
		expectErr       bool
	}{
		{
			name:  "With since and until",
			since: "v1.2.0",
			until: "HEAD",
			mockGitCommands: map[string]string{
				"git log --pretty=format:%s v1.2.0..HEAD": "feat: login\nfix: auth bug",
			},
			expectedCommits: []string{"feat: login", "fix: auth bug"},
		},
		{
			name:  "With default until",
			since: "v1.2.0",
			until: "",
			mockGitCommands: map[string]string{
				"git log --pretty=format:%s v1.2.0..HEAD": "feat: new api",
			},
			expectedCommits: []string{"feat: new api"},
		},
		{
			name:  "Empty commit log",
			since: "v1.2.0",
			until: "HEAD",
			mockGitCommands: map[string]string{
				"git log --pretty=format:%s v1.2.0..HEAD": "",
			},
			expectedCommits: []string{},
		},
		{
			name:  "Fallback to HEAD~10 when no tag found",
			since: "",
			until: "HEAD",
			mockGitCommands: map[string]string{
				"git describe --tags --abbrev=0":           "", // simulate error
				"git log --pretty=format:%s HEAD~10..HEAD": "fix: update",
			},
			expectedCommits: []string{"fix: update"},
		},
		{
			name:  "Since is empty, getLastTag returns valid tag",
			since: "",
			until: "HEAD",
			mockGitCommands: map[string]string{
				"git describe --tags --abbrev=0":          "v2.0.0",
				"git log --pretty=format:%s v2.0.0..HEAD": "feat: something",
			},
			expectedCommits: []string{"feat: something"},
			expectErr:       false,
		},
		{
			name:  "Git log returns error",
			since: "v1.0.0",
			until: "HEAD",
			mockGitCommands: map[string]string{
				"git log --pretty=format:%s v1.0.0..HEAD": "ERROR",
			},
			expectErr: true,
		},
		{
			name:  "GetLastTag returns error",
			since: "",
			until: "HEAD",
			mockGitCommands: map[string]string{
				"git describe --tags --abbrev=0":           "ERROR",
				"git log --pretty=format:%s HEAD~10..HEAD": "fix: fallback",
			},
			expectedCommits: []string{"fix: fallback"},
			expectErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeGitCommands = tt.mockGitCommands

			commits, err := GetCommitsFn(tt.since, tt.until)

			if (err != nil) != tt.expectErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(commits) != len(tt.expectedCommits) {
				t.Fatalf("expected %d commits, got %d", len(tt.expectedCommits), len(commits))
			}
			for i := range commits {
				if commits[i] != tt.expectedCommits[i] {
					t.Errorf("commit %d: expected %q, got %q", i, tt.expectedCommits[i], commits[i])
				}
			}
		})
	}
}
