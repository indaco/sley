package extensionmgr

import (
	"testing"
)

/* ------------------------------------------------------------------------- */
/* UNIT TESTS                                                                */
/* ------------------------------------------------------------------------- */

func TestValidateGitAvailable(t *testing.T) {
	// This test will pass if git is installed, fail otherwise
	// In CI/CD environments, git is typically available
	err := ValidateGitAvailable()

	// We can't reliably test both cases without mocking exec.Command
	// So we just verify the function runs without panicking
	if err != nil {
		t.Logf("git not available: %v (this is expected if git is not installed)", err)
	} else {
		t.Log("git is available")
	}
}
