package tui

import (
	"os"
	"testing"
)

func TestIsInteractive_CI(t *testing.T) {

	tests := []struct {
		name   string
		envVar string
		value  string
	}{
		{"CI", "CI", "true"},
		{"CONTINUOUS_INTEGRATION", "CONTINUOUS_INTEGRATION", "1"},
		{"GITHUB_ACTIONS", "GITHUB_ACTIONS", "true"},
		{"GITLAB_CI", "GITLAB_CI", "true"},
		{"CIRCLECI", "CIRCLECI", "true"},
		{"TRAVIS", "TRAVIS", "true"},
		{"JENKINS_HOME", "JENKINS_HOME", "/var/jenkins"},
		{"BUILDKITE", "BUILDKITE", "true"},
		{"BITBUCKET_BUILD_NUMBER", "BITBUCKET_BUILD_NUMBER", "123"},
		{"DRONE", "DRONE", "true"},
		{"SEMAPHORE", "SEMAPHORE", "true"},
		{"APPVEYOR", "APPVEYOR", "True"},
		{"CODEBUILD_BUILD_ID", "CODEBUILD_BUILD_ID", "build-123"},
		{"TF_BUILD", "TF_BUILD", "True"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Set CI environment variable

			oldVal := os.Getenv(tt.envVar)
			os.Setenv(tt.envVar, tt.value)
			defer func() {
				if oldVal == "" {
					os.Unsetenv(tt.envVar)
				} else {
					os.Setenv(tt.envVar, oldVal)
				}
			}()

			// IsInteractive should return false in CI
			if IsInteractive() {
				t.Errorf("IsInteractive() should return false when %s is set", tt.envVar)
			}
		})
	}
}

func TestIsInteractive_NoCIEnv(t *testing.T) {

	// This test is environment-dependent
	// In a real terminal: should return true
	// In CI or with redirected stdout: should return false
	// We can't reliably test the TTY check, so we just ensure it doesn't panic

	_ = IsInteractive()
}

func TestIsTTY(t *testing.T) {

	// This test is environment-dependent
	// Just ensure it doesn't panic

	_ = IsTTY()
}
