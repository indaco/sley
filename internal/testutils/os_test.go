package testutils

import (
	"os"
	"testing"
)

func TestIsWindows(t *testing.T) {
	t.Parallel(
	// Save original OS env
	)

	origOS := os.Getenv("OS")
	defer func() {
		if origOS == "" {
			os.Unsetenv("OS")
		} else {
			os.Setenv("OS", origOS)
		}
	}()

	tests := []struct {
		name   string
		osEnv  string
		expect bool
	}{
		{
			name:   "Windows_NT",
			osEnv:  "Windows_NT",
			expect: true,
		},
		{
			name:   "windows lowercase",
			osEnv:  "windows",
			expect: true,
		},
		{
			name:   "WINDOWS uppercase",
			osEnv:  "WINDOWS",
			expect: true,
		},
		{
			name:   "empty",
			osEnv:  "",
			expect: false,
		},
		{
			name:   "linux",
			osEnv:  "linux",
			expect: false,
		},
		{
			name:   "darwin",
			osEnv:  "darwin",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			os.Setenv("OS", tt.osEnv)

			got := IsWindows()
			if got != tt.expect {
				t.Errorf("IsWindows() with OS=%q = %v, want %v", tt.osEnv, got, tt.expect)
			}
		})
	}
}
