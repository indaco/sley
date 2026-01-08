# Release Gate Plugin

The release gate plugin enforces quality gates before allowing version bumps. It validates conditions like clean git state, branch constraints, and commit hygiene to ensure releases meet quality standards.

## Plugin Metadata

| Field       | Value                                     |
| ----------- | ----------------------------------------- |
| Name        | `release-gate`                            |
| Version     | v0.1.0                                    |
| Type        | `release-gate`                            |
| Description | Pre-bump validation for release readiness |

## Status

Built-in, **disabled by default**

## Features

- Require clean git working tree (no uncommitted changes)
- Block releases from specific branches
- Restrict releases to allowed branches only
- Detect WIP (work in progress) commits in recent history
- Prevent releases with fixup/squash commits
- Optional CI status checking (placeholder for future implementation)

## How It Works

The plugin runs validation checks **before** any version bump operation. If any gate fails, the bump is aborted and the version file remains unchanged. This fail-fast behavior ensures releases only happen when quality standards are met.

### Validation Order

1. Clean worktree check (if enabled)
2. Branch constraints (blocked branches, then allowed branches)
3. WIP commit detection (if enabled)
4. CI status check (if enabled, currently a placeholder)

## Configuration

Enable and configure in `.sley.yaml`:

```yaml
plugins:
  release-gate:
    enabled: true
    require-clean-worktree: true
    require-ci-pass: false
    blocked-on-wip-commits: true
    allowed-branches: []
    blocked-branches: []
```

### Configuration Options

| Option                   | Type     | Default | Description                                                     |
| ------------------------ | -------- | ------- | --------------------------------------------------------------- |
| `enabled`                | bool     | false   | Enable/disable the plugin                                       |
| `require-clean-worktree` | bool     | true    | Block bumps if git has uncommitted changes                      |
| `require-ci-pass`        | bool     | false   | Check CI status before allowing bumps (not yet implemented)     |
| `blocked-on-wip-commits` | bool     | true    | Block if recent commits contain WIP/fixup/squash                |
| `allowed-branches`       | []string | []      | Branches where bumps are allowed (empty = all branches allowed) |
| `blocked-branches`       | []string | []      | Branches where bumps are never allowed (takes precedence)       |

## Usage Examples

### Basic Setup - Clean Worktree Required

```yaml
plugins:
  release-gate:
    enabled: true
    require-clean-worktree: true
```

```bash
# With uncommitted changes
$ sley bump patch
Error: release-gate: uncommitted changes detected. Commit or stash changes before bumping

# After committing changes
$ git add -A && git commit -m "feat: add feature"
$ sley bump patch
Version bumped from 1.2.3 to 1.2.4
```

### Branch Restrictions - Production Releases Only

Restrict version bumps to specific branches:

```yaml
plugins:
  release-gate:
    enabled: true
    allowed-branches:
      - "main"
      - "release/*"
```

```bash
# On feature branch
$ git checkout -b feature/new-feature
$ sley bump minor
Error: release-gate: bumps not allowed from branch "feature/new-feature". Allowed branches: [main release/*]

# On allowed branch
$ git checkout main
$ sley bump minor
Version bumped from 1.2.3 to 1.3.0
```

### Block Development Branches

Prevent accidental releases from development branches:

```yaml
plugins:
  release-gate:
    enabled: true
    blocked-branches:
      - "dev"
      - "develop"
      - "experimental/*"
```

```bash
# On blocked branch
$ git checkout dev
$ sley bump patch
Error: release-gate: bumps not allowed from branch "dev" (blocked branches: [dev develop experimental/*])

# Blocked branches take precedence even if in allowed list
```

### WIP Commit Detection

Block releases if recent commits contain work-in-progress markers:

```yaml
plugins:
  release-gate:
    enabled: true
    blocked-on-wip-commits: true
```

The plugin detects the following patterns in recent commit messages:

- `WIP` (case-insensitive)
- `fixup!`
- `squash!`
- `DO NOT MERGE` (case-insensitive)
- `DNM`

```bash
# With WIP commit in history
$ git log --oneline -5
abc123 WIP: testing feature
def456 feat: add feature

$ sley bump patch
Error: release-gate: WIP commit detected in recent history: "WIP: testing feature". Complete your work before releasing

# After cleaning up commits
$ git rebase -i HEAD~2  # squash or reword WIP commits
$ sley bump patch
Version bumped from 1.2.3 to 1.2.4
```

### Combined Configuration

Production-ready setup with multiple gates:

```yaml
plugins:
  release-gate:
    enabled: true
    require-clean-worktree: true
    blocked-on-wip-commits: true
    allowed-branches:
      - "main"
      - "release/*"
    blocked-branches:
      - "experimental/*"
```

This configuration ensures:

- No uncommitted changes
- No WIP commits in recent history
- Releases only from `main` or `release/*` branches
- Never from `experimental/*` branches

## Branch Pattern Matching

Branch patterns support glob-style wildcards:

| Pattern        | Matches              | Examples                              |
| -------------- | -------------------- | ------------------------------------- |
| `main`         | Exact match          | `main` only                           |
| `release/*`    | Prefix with wildcard | `release/v1.0`, `release/production`  |
| `*/production` | Suffix with wildcard | `team-a/production`, `api/production` |
| `feature/*/*`  | Multiple wildcards   | `feature/team/login`                  |

## Error Messages

The plugin provides clear, actionable error messages:

| Condition          | Error Message                                                                        |
| ------------------ | ------------------------------------------------------------------------------------ |
| Dirty worktree     | `release-gate: uncommitted changes detected. Commit or stash changes before bumping` |
| WIP commit         | `release-gate: WIP commit detected in recent history: "...". Complete your work...`  |
| Branch not allowed | `release-gate: bumps not allowed from branch "...". Allowed branches: [...]`         |
| Branch blocked     | `release-gate: bumps not allowed from branch "..." (blocked branches: [...])`        |

## Integration with Other Plugins

The release gate runs **before** other validation plugins:

1. Release Gate (this plugin) - Quality gates
2. Version Validator - Version policy rules
3. Dependency Check - Consistency validation
4. Tag Manager - Tag availability

This order ensures quality standards are met before any version-specific validations.

## Best Practices

### Development Workflow

```yaml
# .sley.yaml for a typical development workflow
plugins:
  release-gate:
    enabled: true
    require-clean-worktree: true
    blocked-on-wip-commits: true
    allowed-branches:
      - "main"
      - "release/*"
```

### Monorepo Setup

For monorepos, the release gate applies to each module independently:

```bash
# All modules must pass release gates
$ sley bump minor --all

# Single module bump
$ sley bump patch --module api
```

### CI/CD Integration

Combine with tag manager for automated releases:

```yaml
plugins:
  release-gate:
    enabled: true
    require-clean-worktree: true
    blocked-on-wip-commits: true
    allowed-branches: ["main"]

  tag-manager:
    enabled: true
    auto-create: true
    push: true
```

## Troubleshooting

### Uncommitted Changes Error

```bash
$ git status --porcelain
M .version
M src/main.go

$ sley bump patch
Error: release-gate: uncommitted changes detected...
```

**Solution**: Commit or stash your changes:

```bash
$ git add -A && git commit -m "feat: changes"
# or
$ git stash
```

### WIP Commit Error

```bash
$ git log --oneline -5
abc123 fixup! previous commit
```

**Solution**: Clean up commits using interactive rebase:

```bash
$ git rebase -i HEAD~5
# Mark 'fixup!' commits for squashing
```

### Branch Not Allowed Error

```bash
$ git branch --show-current
feature/test

$ sley bump patch
Error: release-gate: bumps not allowed from branch "feature/test"...
```

**Solution**: Switch to an allowed branch:

```bash
$ git checkout main
$ sley bump patch
```

## Disabling for Emergency Releases

To temporarily bypass the release gate:

1. Edit `.sley.yaml` and set `enabled: false`
2. Perform the bump
3. Re-enable the plugin

**Note**: This is not recommended as it bypasses safety checks. Consider using `--skip-hooks` for extension hooks instead, which doesn't affect plugins.

## Future Enhancements

Planned features:

- **CI Status Integration**: Check GitHub Actions, GitLab CI, or other CI systems before allowing bumps
- **Custom Git Commands**: Run custom git checks via configuration
- **Time-based Gates**: Require minimum time since last release
- **Approval Requirements**: Require manual approval file or PR approval

## Related Plugins

- [Version Validator](VERSION_VALIDATOR.md) - Enforce version policies
- [Tag Manager](TAG_MANAGER.md) - Manage git tags
- [Dependency Check](DEPENDENCY_CHECK.md) - Sync dependency versions

## See Also

- [Plugin Architecture](../PLUGINS.md)
- [Configuration Guide](../CONFIGURATION.md)
- [Monorepo Workflows](../WORKSPACE.md)
