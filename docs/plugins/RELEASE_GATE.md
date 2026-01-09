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

## How It Works

The plugin runs validation checks **before** any version bump. If any gate fails, the bump is aborted and the version file remains unchanged.

Validation order:

1. Clean worktree check (if enabled)
2. Branch constraints (blocked branches, then allowed branches)
3. WIP commit detection (if enabled)

## Configuration

Enable and configure in `.sley.yaml`. See [release-gate.yaml](./examples/release-gate.yaml) for complete examples.

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

### Configuration Options

| Option                   | Type     | Default | Description                                                     |
| ------------------------ | -------- | ------- | --------------------------------------------------------------- |
| `enabled`                | bool     | false   | Enable/disable the plugin                                       |
| `require-clean-worktree` | bool     | true    | Block bumps if git has uncommitted changes                      |
| `blocked-on-wip-commits` | bool     | true    | Block if recent commits contain WIP/fixup/squash                |
| `allowed-branches`       | []string | []      | Branches where bumps are allowed (empty = all branches allowed) |
| `blocked-branches`       | []string | []      | Branches where bumps are never allowed (takes precedence)       |

### WIP Commit Detection

The plugin detects these patterns in recent commit messages:

- `WIP` (case-insensitive)
- `fixup!`
- `squash!`
- `DO NOT MERGE` (case-insensitive)
- `DNM`

## Branch Pattern Matching

Branch patterns support glob-style wildcards:

| Pattern        | Matches                               |
| -------------- | ------------------------------------- |
| `main`         | Exact match only                      |
| `release/*`    | `release/v1.0`, `release/production`  |
| `*/production` | `team-a/production`, `api/production` |

## Error Messages

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

## Troubleshooting

| Issue               | Solution                                      |
| ------------------- | --------------------------------------------- |
| Uncommitted changes | `git add -A && git commit` or `git stash`     |
| WIP commit detected | `git rebase -i HEAD~N` to squash/reword WIP   |
| Branch not allowed  | Switch to allowed branch: `git checkout main` |

## Emergency Bypass

To temporarily bypass release gates, set `enabled: false` in `.sley.yaml`. Re-enable after the emergency release.

## See Also

- [Example Configuration](./examples/release-gate.yaml) - Complete release-gate setup
- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Version Validator](./VERSION_VALIDATOR.md) - Enforce version policies
- [Tag Manager](./TAG_MANAGER.md) - Manage git tags
