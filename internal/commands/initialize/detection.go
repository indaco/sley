package initialize

import (
	"os"
	"path/filepath"
)

// ProjectContext holds information about the detected project environment.
type ProjectContext struct {
	IsGitRepo        bool
	HasPackageJSON   bool
	HasGoMod         bool
	HasCargoToml     bool
	HasPyprojectToml bool
}

// DetectProjectContext analyzes the current directory to detect project type and environment.
// This helps provide smart defaults and suggestions during initialization.
func DetectProjectContext() *ProjectContext {
	ctx := &ProjectContext{}

	// Detect Git repository
	ctx.IsGitRepo = isGitRepository()

	// Detect package managers and project files
	ctx.HasPackageJSON = fileExists("package.json")
	ctx.HasGoMod = fileExists("go.mod")
	ctx.HasCargoToml = fileExists("Cargo.toml")
	ctx.HasPyprojectToml = fileExists("pyproject.toml")

	return ctx
}

// SuggestedPlugins returns a list of plugin names that would be useful for this project.
func (ctx *ProjectContext) SuggestedPlugins() []string {
	suggestions := []string{}

	// Always suggest commit-parser and tag-manager for git repos
	if ctx.IsGitRepo {
		suggestions = append(suggestions, "commit-parser", "tag-manager")
	}

	// Suggest dependency-check for projects with lockfiles
	if ctx.HasPackageJSON || ctx.HasGoMod || ctx.HasCargoToml || ctx.HasPyprojectToml {
		suggestions = append(suggestions, "dependency-check")
	}

	return suggestions
}

// FormatDetectionSummary returns a human-readable summary of detected project features.
func (ctx *ProjectContext) FormatDetectionSummary() string {
	if !ctx.HasAnyDetection() {
		return ""
	}

	summary := "Detected:\n"

	if ctx.IsGitRepo {
		summary += "  - Git repository\n"
	}
	if ctx.HasPackageJSON {
		summary += "  - package.json (Node.js project)\n"
	}
	if ctx.HasGoMod {
		summary += "  - go.mod (Go project)\n"
	}
	if ctx.HasCargoToml {
		summary += "  - Cargo.toml (Rust project)\n"
	}
	if ctx.HasPyprojectToml {
		summary += "  - pyproject.toml (Python project)\n"
	}

	return summary
}

// HasAnyDetection returns true if any project features were detected.
func (ctx *ProjectContext) HasAnyDetection() bool {
	return ctx.IsGitRepo || ctx.HasPackageJSON || ctx.HasGoMod || ctx.HasCargoToml || ctx.HasPyprojectToml
}

// isGitRepository checks if the current directory is inside a git repository.
func isGitRepository() bool {
	// Check for .git directory in current or parent directories
	dir, err := os.Getwd()
	if err != nil {
		return false
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			// .git can be either a directory or a file (for worktrees/submodules)
			return info.IsDir() || info.Mode().IsRegular()
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return false
}

// fileExists checks if a file exists in the current directory.
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
