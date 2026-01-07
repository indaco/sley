package plugins

import (
	"fmt"
	"sync"

	"github.com/indaco/sley/internal/plugins/auditlog"
	"github.com/indaco/sley/internal/plugins/changeloggenerator"
	"github.com/indaco/sley/internal/plugins/changelogparser"
	"github.com/indaco/sley/internal/plugins/commitparser"
	"github.com/indaco/sley/internal/plugins/dependencycheck"
	"github.com/indaco/sley/internal/plugins/releasegate"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/plugins/versionvalidator"
)

// PluginRegistry is a thread-safe registry for all plugin instances.
// It replaces global package-level variables with a centralized, injectable registry.
type PluginRegistry struct {
	mu                 sync.RWMutex
	commitParser       commitparser.CommitParser
	tagManager         tagmanager.TagManager
	versionValidator   versionvalidator.VersionValidator
	dependencyChecker  dependencycheck.DependencyChecker
	changelogParser    changelogparser.ChangelogInferrer
	changelogGenerator changeloggenerator.ChangelogGenerator
	releaseGate        releasegate.ReleaseGate
	auditLog           auditlog.AuditLog
}

// NewPluginRegistry creates a new empty plugin registry.
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{}
}

// RegisterCommitParser registers a commit parser plugin.
func (r *PluginRegistry) RegisterCommitParser(p commitparser.CommitParser) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.commitParser != nil {
		return fmt.Errorf("commit parser %q is already registered, ignoring %q",
			r.commitParser.Name(), p.Name())
	}
	r.commitParser = p
	return nil
}

// GetCommitParser retrieves the registered commit parser, or nil if not registered.
func (r *PluginRegistry) GetCommitParser() commitparser.CommitParser {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.commitParser
}

// RegisterTagManager registers a tag manager plugin.
func (r *PluginRegistry) RegisterTagManager(tm tagmanager.TagManager) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.tagManager != nil {
		return fmt.Errorf("tag manager %q is already registered, ignoring %q",
			r.tagManager.Name(), tm.Name())
	}
	r.tagManager = tm
	return nil
}

// GetTagManager retrieves the registered tag manager, or nil if not registered.
func (r *PluginRegistry) GetTagManager() tagmanager.TagManager {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tagManager
}

// RegisterVersionValidator registers a version validator plugin.
func (r *PluginRegistry) RegisterVersionValidator(vv versionvalidator.VersionValidator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.versionValidator != nil {
		return fmt.Errorf("version validator %q is already registered, ignoring %q",
			r.versionValidator.Name(), vv.Name())
	}
	r.versionValidator = vv
	return nil
}

// GetVersionValidator retrieves the registered version validator, or nil if not registered.
func (r *PluginRegistry) GetVersionValidator() versionvalidator.VersionValidator {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.versionValidator
}

// RegisterDependencyChecker registers a dependency checker plugin.
func (r *PluginRegistry) RegisterDependencyChecker(dc dependencycheck.DependencyChecker) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.dependencyChecker != nil {
		return fmt.Errorf("dependency checker %q is already registered, ignoring %q",
			r.dependencyChecker.Name(), dc.Name())
	}
	r.dependencyChecker = dc
	return nil
}

// GetDependencyChecker retrieves the registered dependency checker, or nil if not registered.
func (r *PluginRegistry) GetDependencyChecker() dependencycheck.DependencyChecker {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.dependencyChecker
}

// RegisterChangelogParser registers a changelog parser plugin.
func (r *PluginRegistry) RegisterChangelogParser(cp changelogparser.ChangelogInferrer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.changelogParser != nil {
		return fmt.Errorf("changelog parser %q is already registered, ignoring %q",
			r.changelogParser.Name(), cp.Name())
	}
	r.changelogParser = cp
	return nil
}

// GetChangelogParser retrieves the registered changelog parser, or nil if not registered.
func (r *PluginRegistry) GetChangelogParser() changelogparser.ChangelogInferrer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.changelogParser
}

// RegisterChangelogGenerator registers a changelog generator plugin.
func (r *PluginRegistry) RegisterChangelogGenerator(cg changeloggenerator.ChangelogGenerator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.changelogGenerator != nil {
		return fmt.Errorf("changelog generator %q is already registered, ignoring %q",
			r.changelogGenerator.Name(), cg.Name())
	}
	r.changelogGenerator = cg
	return nil
}

// GetChangelogGenerator retrieves the registered changelog generator, or nil if not registered.
func (r *PluginRegistry) GetChangelogGenerator() changeloggenerator.ChangelogGenerator {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.changelogGenerator
}

// RegisterReleaseGate registers a release gate plugin.
func (r *PluginRegistry) RegisterReleaseGate(rg releasegate.ReleaseGate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.releaseGate != nil {
		return fmt.Errorf("release gate %q is already registered, ignoring %q",
			r.releaseGate.Name(), rg.Name())
	}
	r.releaseGate = rg
	return nil
}

// GetReleaseGate retrieves the registered release gate, or nil if not registered.
func (r *PluginRegistry) GetReleaseGate() releasegate.ReleaseGate {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.releaseGate
}

// RegisterAuditLog registers an audit log plugin.
func (r *PluginRegistry) RegisterAuditLog(al auditlog.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.auditLog != nil {
		return fmt.Errorf("audit log %q is already registered, ignoring %q",
			r.auditLog.Name(), al.Name())
	}
	r.auditLog = al
	return nil
}

// GetAuditLog retrieves the registered audit log, or nil if not registered.
func (r *PluginRegistry) GetAuditLog() auditlog.AuditLog {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.auditLog
}

// Reset clears all registered plugins. Useful for testing.
func (r *PluginRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commitParser = nil
	r.tagManager = nil
	r.versionValidator = nil
	r.dependencyChecker = nil
	r.changelogParser = nil
	r.changelogGenerator = nil
	r.releaseGate = nil
	r.auditLog = nil
}
