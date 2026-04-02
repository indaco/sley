package initialize

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	goyaml "github.com/goccy/go-yaml"
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

// MonorepoInfo holds information about a detected monorepo workspace.
type MonorepoInfo struct {
	Type       string   // "go-work", "pnpm", "npm", "cargo"
	MarkerFile string   // path to the detected marker file
	Modules    []string // discovered module directory paths (relative)
}

// DetectMonorepo checks for monorepo workspace markers in the current directory.
// Returns nil if no markers are found. Detection order: go.work, pnpm-workspace.yaml,
// package.json (workspaces), Cargo.toml ([workspace]).
func DetectMonorepo() (*MonorepoInfo, error) {
	// 1. go.work
	if fileExists("go.work") {
		return detectGoWork()
	}

	// 2. pnpm-workspace.yaml
	if fileExists("pnpm-workspace.yaml") {
		return detectPnpmWorkspace()
	}

	// 3. package.json with workspaces
	if fileExists("package.json") {
		info, err := detectNpmWorkspaces()
		if err != nil {
			return nil, err
		}
		if info != nil {
			return info, nil
		}
	}

	// 4. Cargo.toml with [workspace]
	if fileExists("Cargo.toml") {
		info, err := detectCargoWorkspace()
		if err != nil {
			return nil, err
		}
		if info != nil {
			return info, nil
		}
	}

	return nil, nil
}

// detectGoWork parses go.work to extract module paths from use directives.
func detectGoWork() (*MonorepoInfo, error) {
	f, err := os.Open("go.work")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var modules []string
	scanner := bufio.NewScanner(f)
	inBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Handle block syntax: use ( ... )
		if strings.HasPrefix(line, "use (") || line == "use (" {
			inBlock = true
			continue
		}
		if inBlock {
			if line == ")" {
				inBlock = false
				continue
			}
			mod := cleanGoWorkPath(line)
			if mod != "" {
				modules = append(modules, mod)
			}
			continue
		}

		// Handle single-line: use ./cobra
		if strings.HasPrefix(line, "use ") {
			mod := cleanGoWorkPath(strings.TrimPrefix(line, "use "))
			if mod != "" {
				modules = append(modules, mod)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &MonorepoInfo{
		Type:       "go-work",
		MarkerFile: "go.work",
		Modules:    modules,
	}, nil
}

// cleanGoWorkPath strips ./ prefix and skips external paths starting with "..".
func cleanGoWorkPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	path = strings.TrimPrefix(path, "./")
	if strings.HasPrefix(path, "..") {
		return ""
	}
	return path
}

// detectPnpmWorkspace parses pnpm-workspace.yaml to read packages field and expand globs.
func detectPnpmWorkspace() (*MonorepoInfo, error) {
	data, err := os.ReadFile("pnpm-workspace.yaml")
	if err != nil {
		return nil, err
	}

	var ws struct {
		Packages []string `yaml:"packages"`
	}
	if err := goyaml.Unmarshal(data, &ws); err != nil {
		return nil, err
	}

	modules := expandGlobPatterns(ws.Packages)

	return &MonorepoInfo{
		Type:       "pnpm",
		MarkerFile: "pnpm-workspace.yaml",
		Modules:    modules,
	}, nil
}

// detectNpmWorkspaces parses package.json for a workspaces field.
func detectNpmWorkspaces() (*MonorepoInfo, error) {
	data, err := os.ReadFile("package.json")
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Workspaces []string `json:"workspaces"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, nil // not a JSON parse error we care about, just skip
	}
	if len(pkg.Workspaces) == 0 {
		return nil, nil
	}

	modules := expandGlobPatterns(pkg.Workspaces)

	return &MonorepoInfo{
		Type:       "npm",
		MarkerFile: "package.json",
		Modules:    modules,
	}, nil
}

// detectCargoWorkspace checks Cargo.toml for a [workspace] section and parses members.
func detectCargoWorkspace() (*MonorepoInfo, error) {
	f, err := os.Open("Cargo.toml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inWorkspace := false
	inMembers := false
	var modules []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Detect [workspace] section
		if line == "[workspace]" {
			inWorkspace = true
			continue
		}

		// New section starts — stop if we were in [workspace]
		if strings.HasPrefix(line, "[") && line != "[workspace]" {
			if inWorkspace {
				break
			}
			continue
		}

		if !inWorkspace {
			continue
		}

		// Look for members = [...]
		if strings.HasPrefix(line, "members") {
			inMembers = true
			// Handle inline: members = ["a", "b"]
			if idx := strings.Index(line, "["); idx >= 0 {
				rest := line[idx:]
				modules = append(modules, parseTomlStringArray(rest)...)
				if strings.Contains(rest, "]") {
					inMembers = false
				}
			}
			continue
		}

		if inMembers {
			modules = append(modules, parseTomlStringArray(line)...)
			if strings.Contains(line, "]") {
				inMembers = false
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if !inWorkspace {
		return nil, nil
	}

	// Expand glob patterns in members
	expanded := expandGlobPatterns(modules)

	return &MonorepoInfo{
		Type:       "cargo",
		MarkerFile: "Cargo.toml",
		Modules:    expanded,
	}, nil
}

// parseTomlStringArray extracts quoted strings from a TOML-style array line.
// For example: `["crate-a", "crate-b"]` returns ["crate-a", "crate-b"].
func parseTomlStringArray(line string) []string {
	var results []string
	for {
		start := strings.IndexByte(line, '"')
		if start < 0 {
			break
		}
		end := strings.IndexByte(line[start+1:], '"')
		if end < 0 {
			break
		}
		results = append(results, line[start+1:start+1+end])
		line = line[start+1+end+1:]
	}
	return results
}

// expandGlobPatterns takes a list of glob patterns and expands them to actual directories.
func expandGlobPatterns(patterns []string) []string {
	var dirs []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// Strip trailing slash or /* for matching
		cleanPattern := strings.TrimSuffix(pattern, "/")

		matches, err := filepath.Glob(cleanPattern)
		if err != nil {
			// If the pattern itself is not a valid glob, try it as a literal directory
			if info, statErr := os.Stat(cleanPattern); statErr == nil && info.IsDir() {
				if !seen[cleanPattern] {
					dirs = append(dirs, cleanPattern)
					seen[cleanPattern] = true
				}
			}
			continue
		}

		for _, m := range matches {
			info, err := os.Stat(m)
			if err != nil || !info.IsDir() {
				continue
			}
			if !seen[m] {
				dirs = append(dirs, m)
				seen[m] = true
			}
		}
	}

	return dirs
}
