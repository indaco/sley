package discovery

import (
	"context"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/parser"
	"github.com/indaco/sley/internal/semver"
)

// Service provides version source discovery functionality.
type Service struct {
	fs     core.FileSystem
	cfg    *config.Config
	parser *parser.Reader
}

// NewService creates a new discovery Service.
func NewService(fs core.FileSystem, cfg *config.Config) *Service {
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &Service{
		fs:     fs,
		cfg:    cfg,
		parser: parser.NewReader(fs),
	}
}

// Discover scans the given root directory and returns discovery results.
func (s *Service) Discover(ctx context.Context, root string) (*Result, error) {
	result := &Result{
		Mode:           NoModules,
		Modules:        make([]Module, 0),
		Manifests:      make([]ManifestSource, 0),
		SyncCandidates: make([]SyncCandidate, 0),
		Mismatches:     make([]Mismatch, 0),
	}

	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Discover .version files (modules)
	modules, err := s.discoverModules(ctx, root)
	if err != nil {
		return nil, err
	}
	result.Modules = modules

	// Set detection mode
	switch len(modules) {
	case 0:
		result.Mode = NoModules
	case 1:
		result.Mode = SingleModule
	default:
		result.Mode = MultiModule
	}

	// Discover manifest files
	manifests, err := s.discoverManifests(ctx, root)
	if err != nil {
		return nil, err
	}
	result.Manifests = manifests

	// Generate sync candidates from manifests
	result.SyncCandidates = s.generateSyncCandidates(manifests)

	// Detect version mismatches
	result.Mismatches = s.detectMismatches(result)

	return result, nil
}

// discoverModules finds all .version files in the directory tree.
func (s *Service) discoverModules(ctx context.Context, root string) ([]Module, error) {
	var modules []Module
	discovery := s.cfg.GetDiscoveryConfig()

	// Check if discovery is disabled
	if discovery.Enabled != nil && !*discovery.Enabled {
		return modules, nil
	}

	// First check for .version in root
	rootVersion := filepath.Join(root, ".version")
	if module, err := s.loadModule(ctx, rootVersion, root); err == nil {
		modules = append(modules, *module)
	}

	// Get max depth
	maxDepth := core.MaxDiscoveryDepth
	if discovery.MaxDepth != nil {
		maxDepth = *discovery.MaxDepth
	}

	// Check if recursive discovery is enabled
	recursive := discovery.Recursive == nil || *discovery.Recursive
	if !recursive {
		return modules, nil
	}

	// Scan subdirectories
	excludes := s.cfg.GetExcludePatterns()
	err := s.walkDirectory(ctx, root, 0, maxDepth, excludes, func(path string) error {
		// Skip if we already found this (the root .version)
		if path == rootVersion {
			return nil
		}

		if module, err := s.loadModule(ctx, path, root); err == nil {
			modules = append(modules, *module)
		}
		return nil
	})

	return modules, err
}

// loadModule creates a Module from a .version file path.
func (s *Service) loadModule(ctx context.Context, versionPath, root string) (*Module, error) {
	// Check if file exists
	if _, err := s.fs.Stat(ctx, versionPath); err != nil {
		return nil, err
	}

	// Read version
	data, err := s.fs.ReadFile(ctx, versionPath)
	if err != nil {
		return nil, err
	}
	version := strings.TrimSpace(string(data))

	// Get relative path
	relPath, err := filepath.Rel(root, versionPath)
	if err != nil {
		relPath = versionPath
	}

	// Determine module name
	dir := filepath.Dir(versionPath)
	name := filepath.Base(dir)
	if dir == root || dir == "." {
		name = "root"
	}

	return &Module{
		Name:    name,
		Path:    versionPath,
		RelPath: relPath,
		Version: version,
		Dir:     dir,
	}, nil
}

// walkDirectory walks the directory tree looking for .version files.
func (s *Service) walkDirectory(ctx context.Context, dir string, depth, maxDepth int, excludes []string, fn func(string) error) error {
	if depth > maxDepth {
		return nil
	}

	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return err
	}

	entries, err := s.fs.ReadDir(ctx, dir)
	if err != nil {
		// Skip directories we can't read
		return nil
	}

	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(dir, name)

		// Skip excluded patterns
		if s.shouldExclude(name, path, excludes) {
			continue
		}

		if entry.IsDir() {
			if err := s.walkDirectory(ctx, path, depth+1, maxDepth, excludes, fn); err != nil {
				return err
			}
		} else if name == ".version" {
			if err := fn(path); err != nil {
				return err
			}
		}
	}

	return nil
}

// shouldExclude checks if a path should be excluded from scanning.
func (s *Service) shouldExclude(name, path string, excludes []string) bool {
	// Skip hidden directories (except .version file)
	if strings.HasPrefix(name, ".") && name != ".version" {
		return true
	}

	// Skip common non-project directories
	skipDirs := []string{"node_modules", "vendor", ".git", "__pycache__", "target", "dist", "build"}
	if slices.Contains(skipDirs, name) {
		return true
	}

	// Check configured excludes
	for _, pattern := range excludes {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}

	return false
}

// discoverManifests finds manifest files in the root directory.
func (s *Service) discoverManifests(ctx context.Context, root string) ([]ManifestSource, error) {
	var manifests []ManifestSource

	for _, known := range DefaultKnownManifests() {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		path := filepath.Join(root, known.Filename)

		// Check if file exists
		if _, err := s.fs.Stat(ctx, path); err != nil {
			continue
		}

		// Try to read the version
		version, err := s.parser.ReadVersion(ctx, parser.FileConfig{
			Path:   path,
			Format: known.Format,
			Field:  known.Field,
		})
		if err != nil {
			continue
		}

		// Validate it looks like a semver
		if !isValidSemver(version) {
			continue
		}

		manifests = append(manifests, ManifestSource{
			Path:        path,
			RelPath:     known.Filename,
			Filename:    known.Filename,
			Version:     version,
			Format:      known.Format,
			Field:       known.Field,
			Description: known.Description,
		})
	}

	return manifests, nil
}

// generateSyncCandidates creates SyncCandidates from discovered manifests.
func (s *Service) generateSyncCandidates(manifests []ManifestSource) []SyncCandidate {
	candidates := make([]SyncCandidate, 0, len(manifests))

	for _, m := range manifests {
		candidates = append(candidates, SyncCandidate{
			Path:        m.RelPath,
			Format:      m.Format,
			Field:       m.Field,
			Version:     m.Version,
			Description: m.Description,
		})
	}

	return candidates
}

// detectMismatches finds version mismatches between sources.
func (s *Service) detectMismatches(result *Result) []Mismatch {
	return DetectMismatches(result)
}

// isValidSemver performs a basic check if the string looks like a semver version.
func isValidSemver(version string) bool {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Use the semver package for validation
	_, err := semver.ParseVersion(version)
	return err == nil
}

// DiscoverModulesOnly is a convenience method that only discovers .version files.
func (s *Service) DiscoverModulesOnly(ctx context.Context, root string) ([]Module, error) {
	return s.discoverModules(ctx, root)
}

// DiscoverManifestsOnly is a convenience method that only discovers manifest files.
func (s *Service) DiscoverManifestsOnly(ctx context.Context, root string) ([]ManifestSource, error) {
	return s.discoverManifests(ctx, root)
}

// DiscoverAt is a convenience function that creates a Service and runs discovery.
func DiscoverAt(ctx context.Context, fsys core.FileSystem, cfg *config.Config, root string) (*Result, error) {
	svc := NewService(fsys, cfg)
	return svc.Discover(ctx, root)
}

// FileFilter is a function type for filtering discovered files.
type FileFilter func(path string, info fs.FileInfo) bool

// WithFilter returns a filtered list of modules.
func (r *Result) WithFilter(filter FileFilter) []Module {
	if filter == nil {
		return r.Modules
	}

	var filtered []Module
	for _, m := range r.Modules {
		if filter(m.Path, nil) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
