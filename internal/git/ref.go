package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// SafeFallbackSince returns "HEAD~n" if the repo has more than n commits,
// otherwise it returns the hash of the root (first) commit so that
// `git log <ref>..HEAD` works even in repos with very few commits.
func SafeFallbackSince(execCommand func(string, ...string) *exec.Cmd, n int) string {
	cmd := execCommand("git", "rev-list", "--count", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("HEAD~%d", n)
	}

	var count int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &count); err != nil || count == 0 {
		return fmt.Sprintf("HEAD~%d", n)
	}

	if count > n {
		return fmt.Sprintf("HEAD~%d", n)
	}

	// Fewer than n commits — use the root commit so the range covers everything.
	root := execCommand("git", "rev-list", "--max-parents=0", "HEAD")
	rootOut, err := root.Output()
	if err != nil {
		return fmt.Sprintf("HEAD~%d", n)
	}

	// rev-list --max-parents=0 can return multiple roots; take the first one.
	rootHash := strings.TrimSpace(strings.SplitN(string(rootOut), "\n", 2)[0])
	if rootHash == "" {
		return fmt.Sprintf("HEAD~%d", n)
	}
	return rootHash
}
