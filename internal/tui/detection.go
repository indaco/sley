package tui

import (
	"os"

	"golang.org/x/term"
)

// IsInteractive determines if the current environment supports interactive prompts.
// It returns false in the following cases:
//   - stdout is not a terminal (redirected to file, pipe, etc.)
//   - running in a CI/CD environment (detected via environment variables)
//
// This function is used to automatically skip TUI prompts in non-interactive contexts.
func IsInteractive() bool {
	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) { //nolint:gosec // G115: fd is a small value, no overflow risk
		return false
	}

	// Check for common CI environment variables
	ciEnvs := []string{
		"CI",                     // Generic CI indicator
		"CONTINUOUS_INTEGRATION", // Generic CI indicator
		"GITHUB_ACTIONS",         // GitHub Actions
		"GITLAB_CI",              // GitLab CI
		"CIRCLECI",               // CircleCI
		"TRAVIS",                 // Travis CI
		"JENKINS_HOME",           // Jenkins
		"BUILDKITE",              // Buildkite
		"BITBUCKET_BUILD_NUMBER", // Bitbucket Pipelines
		"DRONE",                  // Drone CI
		"SEMAPHORE",              // Semaphore CI
		"APPVEYOR",               // AppVeyor
		"CODEBUILD_BUILD_ID",     // AWS CodeBuild
		"TF_BUILD",               // Azure Pipelines
	}

	for _, env := range ciEnvs {
		if os.Getenv(env) != "" {
			return false
		}
	}

	return true
}

// IsTTY checks if stdout is a terminal.
// This is a lower-level check than IsInteractive.
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd())) //nolint:gosec // G115: fd is a small value, no overflow risk
}
