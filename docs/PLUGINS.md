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

## When to Use Plugins vs Extensions

### Use Plugins When

- **Performance matters**: Plugins execute in <1ms with native Go performance
- **Feature is widely applicable**: Common versioning needs across many projects
- **Deep integration needed**: Requires tight coupling with bump logic or validation
- **Built-in reliability required**: No external dependencies or installation steps
- **Examples**: Git tagging, conventional commit parsing, version validation, file syncing

### Use Extensions When

- **Custom to your workflow**: Organization-specific automation or processes
- **Requires external tools**: Need to call AWS CLI, curl, custom scripts, etc.
- **Prototyping new features**: Testing ideas before proposing as built-in plugins
- **Language-specific needs**: Python/Node.js/Ruby tooling integration
- **Examples**: Custom notification systems, deployment triggers, proprietary tool integration

## Plugin vs Extension Comparison

| Feature           | Plugins                              | Extensions                      |
| ----------------- | ------------------------------------ | ------------------------------- |
| **Compilation**   | Built-in, compiled with CLI          | External scripts                |
| **Performance**   | Native Go, <1ms                      | Shell/Python/Node, ~50-100ms    |
| **Installation**  | None required                        | `sley extension install`        |
| **Configuration** | `.sley.yaml` plugins section         | `.sley.yaml` extensions section |
| **Use Case**      | Core version logic, validation, sync | Hook-based automation           |

> [!NOTE]
> Most users will only need plugins. Extensions are for advanced customization and organization-specific workflows.

## Common Workflow Patterns

See [Full Configuration](./plugins/examples/full-config.yaml) for complete examples of all plugins working together.

### Auto-Bump + Changelog

```bash
sley bump auto
# commit-parser analyzes commits -> determines bump type
# changelog-generator creates versioned changelog entry
```

### Auto-Bump + Tag + Push

```bash
sley bump auto
# commit-parser analyzes commits -> bump type
# tag-manager validates tag doesn't exist
# Version updated -> tag created and pushed
```

### Full CI/CD Pipeline

```bash
sley bump auto
# Pre-bump: version-validator, dependency-check, tag-manager validation
# Bump: commit-parser determines type, version updated
# Post-bump: files synced, changelog generated, tag created
```

## See Also

- [Extension System](EXTENSIONS.md) - External hook-based scripts
- [Monorepo Support](MONOREPO.md) - Multi-module workflows

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
