package version

import (
	"testing"
)

// TestGetVersion checks if the GetVersion function correctly retrieves the embedded version.
func TestGetVersion(t *testing.T) {
	t.Parallel()
	expectedVersion := "0.13.0"

	got := GetVersion()
	if got != expectedVersion {
		t.Errorf("GetVersion() = %q; want %q", got, expectedVersion)
	}
}
