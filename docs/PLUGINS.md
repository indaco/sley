# Plugin System

## Overview

Plugins are **built-in** features that extend sley's core functionality. Unlike extensions (which are external scripts), plugins are compiled into the binary and provide deep integration with version bump logic.

## Available Plugins

| Plugin                                                  | Description                                            | Default  |
| ------------------------------------------------------- | ------------------------------------------------------ | -------- |
| [commit-parser](./plugins/COMMIT_PARSER.md)             | Analyzes conventional commits to determine bump type   | Enabled  |
| [tag-manager](./plugins/TAG_MANAGER.md)                 | Automatically creates git tags synchronized with bumps | Disabled |
| [version-validator](./plugins/VERSION_VALIDATOR.md)     | Enforces versioning policies and constraints           | Disabled |
| [dependency-check](./plugins/DEPENDENCY_CHECK.md)       | Validates and syncs versions across multiple files     | Disabled |
| [changelog-parser](./plugins/CHANGELOG_PARSER.md)       | Infers bump type from CHANGELOG.md entries             | Disabled |
| [changelog-generator](./plugins/CHANGELOG_GENERATOR.md) | Generates changelog from conventional commits          | Disabled |
| [release-gate](./plugins/RELEASE_GATE.md)               | Pre-bump validation (clean worktree, branch, WIP)      | Disabled |
| [audit-log](./plugins/AUDIT_LOG.md)                     | Records version changes with metadata to a log file    | Disabled |

> **Note**: The "Default" column shows which plugins are active when running sley without a `.sley.yaml` configuration file. When using `sley init --yes`, a recommended starting configuration is created with both `commit-parser` and `tag-manager` enabled.

## Quick Start

Enable plugins in your `.sley.yaml`:

```yaml
plugins:
  # Analyze commits for automatic bump type detection
  commit-parser: true

  # Automatically create git tags after bumps
  tag-manager:
    enabled: true
    prefix: "v"
    annotate: true
    push: false

  # Enforce versioning policies
  version-validator:
    enabled: true
    rules:
      - type: "major-version-max"
        value: 10
      - type: "branch-constraint"
        branch: "release/*"
        allowed: ["patch"]

  # Sync versions across multiple files
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: "package.json"
        field: "version"
        format: "json"
```

## Plugin Execution Order

During a version bump, plugins execute in a specific order:

```
sley bump patch
  |
  +-- 1. release-gate: Validates pre-conditions (clean worktree, branch, WIP)
  |
  +-- 2. version-validator: Validates version policy
  |
  +-- 3. dependency-check: Validates file consistency
  |
  +-- 4. tag-manager: Validates tag doesn't exist
  |
  +-- 5. Version file updated
  |
  +-- 6. dependency-check: Syncs version to configured files
  |
  +-- 7. changelog-generator: Creates changelog entry
  |
  +-- 8. audit-log: Records version change to log file
  |
  +-- 9. tag-manager: Creates git tag
```

If any pre-bump validation step fails (1-4), the bump is aborted and no changes are made.
Post-bump actions (6-9) are non-blocking - failures are logged but don't fail the bump.

## Plugin vs Extension Comparison

| Feature           | Plugins                              | Extensions                      |
| ----------------- | ------------------------------------ | ------------------------------- |
| **Compilation**   | Built-in, compiled with CLI          | External scripts                |
| **Performance**   | Native Go, <1ms                      | Shell/Python/Node, ~50-100ms    |
| **Installation**  | None required                        | `sley extension install`        |
| **Configuration** | `.sley.yaml` plugins section         | `.sley.yaml` extensions section |
| **Use Case**      | Core version logic, validation, sync | Hook-based automation           |

## Plugins + Extensions: Powerful Combinations

Plugins and extensions work together to create automated version management workflows.

### Pattern 1: Validation + Auto-Bump + Changelog

```yaml
# .sley.yaml
plugins:
  commit-parser: true # Analyze commits for bump type
  changelog-generator:
    enabled: true
    mode: "versioned"
    format: "grouped" # or "keepachangelog" for Keep a Changelog spec
    repository:
      auto-detect: true
```

Workflow:

```bash
sley bump auto
# 1. commit-parser plugin: Analyzes commits -> determines "minor" bump
# 2. Version bumped: 1.2.3 -> 1.3.0
# 3. changelog-generator: Creates .changes/v1.3.0.md
```

### Pattern 2: Auto-Bump + Tag + Push

```yaml
plugins:
  commit-parser: true
  tag-manager:
    enabled: true
    prefix: "v"
    annotate: true
    push: true
```

Workflow:

```bash
sley bump auto
# 1. commit-parser analyzes: feat commits -> minor bump
# 2. tag-manager validates: v1.3.0 doesn't exist
# 3. Version: 1.2.3 -> 1.3.0
# 4. tag-manager creates tag: v1.3.0
# 5. tag-manager pushes tag to remote
```

### Pattern 3: Full CI/CD Automation

```yaml
plugins:
  commit-parser: true
  version-validator:
    enabled: true
    rules:
      - type: "branch-constraint"
        branch: "release/*"
        allowed: ["patch"]
      - type: "major-version-max"
        value: 10
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: "package.json"
        field: "version"
        format: "json"
  changelog-generator:
    enabled: true
    mode: "both"
    format: "keepachangelog" # Keep a Changelog specification format
    repository:
      auto-detect: true
  tag-manager:
    enabled: true
    prefix: "v"
    push: true
```

CI Workflow:

```bash
sley bump auto
# Pre-bump validation:
#   1. version-validator: Checks branch constraints and version limits
#   2. dependency-check: Validates file consistency
#   3. tag-manager: Validates tag doesn't exist
#
# Bump operation:
#   4. commit-parser determines: feat commits -> minor
#   5. Version: 1.2.3 -> 1.3.0
#
# Post-bump actions:
#   6. dependency-check syncs package.json
#   7. changelog-generator creates .changes/v1.3.0.md and updates CHANGELOG.md
#   8. tag-manager creates and pushes tag v1.3.0
```

## See Also

- [Extension System](./EXTENSIONS.md) - External hook-based scripts
- [Monorepo Support](./MONOREPO.md) - Multi-module workflows

### Individual Plugin Documentation

- [Commit Parser](./plugins/COMMIT_PARSER.md) - Conventional commit analysis
- [Tag Manager](./plugins/TAG_MANAGER.md) - Git tag automation
- [Version Validator](./plugins/VERSION_VALIDATOR.md) - Policy enforcement
- [Dependency Check](./plugins/DEPENDENCY_CHECK.md) - Cross-file version sync
- [Changelog Parser](./plugins/CHANGELOG_PARSER.md) - CHANGELOG.md analysis
- [Changelog Generator](./plugins/CHANGELOG_GENERATOR.md) - Changelog generation from commits
- [Release Gate](./plugins/RELEASE_GATE.md) - Pre-bump validation and quality gates
- [Audit Log](./plugins/AUDIT_LOG.md) - Version change history tracking

### Example Configurations

- [Full Configuration](./plugins/examples/full-config.yaml) - All plugins working together
- [Changelog Generator](./plugins/examples/changelog-generator.yaml) - Changelog generation from commits
- [Changelog Parser](./plugins/examples/changelog-parser.yaml) - Changelog-based versioning
- [Dependency Check](./plugins/examples/dependency-check.yaml) - Multi-format file sync
- [Release Gate](./plugins/examples/release-gate.yaml) - Quality gates and branch constraints
- [Audit Log](./plugins/examples/audit-log.yaml) - Version history tracking
