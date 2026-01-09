# Audit Log Plugin

The audit log plugin records all version changes to a persistent log file with metadata such as author, timestamp, commit SHA, and branch. This provides a complete audit trail for compliance, debugging, and rollback scenarios.

## Plugin Metadata

| Field       | Value                                   |
| ----------- | --------------------------------------- |
| Name        | `audit-log`                             |
| Version     | v0.1.0                                  |
| Type        | `audit-log`                             |
| Description | Records version history for audit trail |

## Status

Built-in, **disabled by default**

## Features

- Records every version bump with complete metadata
- Configurable output format (JSON or YAML)
- Captures git author, commit SHA, and branch information
- ISO 8601 timestamps for precise change tracking
- Sorted entries (newest first)
- Non-blocking: failures don't prevent version bumps

## How It Works

1. After a successful version bump, git metadata is collected
2. A new entry is created with version change details
3. The log file is updated (created if missing)
4. Entries are sorted newest first

If the audit log write fails, a warning is displayed but the version bump succeeds.

## Configuration

Enable and configure in `.sley.yaml`. See [audit-log.yaml](./examples/audit-log.yaml) for complete examples.

```yaml
plugins:
  audit-log:
    enabled: true
    path: .version-history.json
    format: json
    include-author: true
    include-timestamp: true
    include-commit-sha: true
    include-branch: true
```

### Configuration Options

| Option               | Type   | Default                 | Description                     |
| -------------------- | ------ | ----------------------- | ------------------------------- |
| `enabled`            | bool   | false                   | Enable/disable the plugin       |
| `path`               | string | `.version-history.json` | Path to the audit log file      |
| `format`             | string | `json`                  | Output format: `json` or `yaml` |
| `include-author`     | bool   | true                    | Include git user name and email |
| `include-timestamp`  | bool   | true                    | Include ISO 8601 timestamp      |
| `include-commit-sha` | bool   | true                    | Include current commit SHA      |
| `include-branch`     | bool   | true                    | Include current branch name     |

## Log File Format

### JSON Format (Default)

```json
{
  "entries": [
    {
      "timestamp": "2026-01-04T12:00:00Z",
      "previous_version": "1.2.3",
      "new_version": "1.2.4",
      "bump_type": "patch",
      "author": "John Doe <john@example.com>",
      "commit_sha": "abc1234567890def",
      "branch": "main"
    }
  ]
}
```

### YAML Format

```yaml
entries:
  - timestamp: "2026-01-04T12:00:00Z"
    previous_version: "1.2.3"
    new_version: "1.2.4"
    bump_type: patch
    author: John Doe <john@example.com>
    commit_sha: abc1234567890def
    branch: main
```

## Entry Fields

| Field              | Type   | Description                                           |
| ------------------ | ------ | ----------------------------------------------------- |
| `timestamp`        | string | ISO 8601 timestamp (UTC) of the version bump          |
| `previous_version` | string | Version before the bump                               |
| `new_version`      | string | Version after the bump                                |
| `bump_type`        | string | Type of bump: patch, minor, major, pre, release, auto |
| `author`           | string | Git user name and email (if include-author: true)     |
| `commit_sha`       | string | Current git commit SHA (if include-commit-sha: true)  |
| `branch`           | string | Current git branch (if include-branch: true)          |

## Usage

Once enabled, the plugin works automatically with all bump commands:

```bash
sley bump patch
# Output: Version bumped from 1.2.3 to 1.2.4
# Creates/updates .version-history.json
```

## Querying the Audit Log

```bash
# Using jq (JSON)
jq '.entries | length' .version-history.json              # Count total bumps
jq '.entries[:5]' .version-history.json                   # Get latest 5 bumps
jq '.entries[] | select(.bump_type == "major")' .version-history.json  # Find major bumps
jq '.entries[] | select(.new_version == "1.2.3") | .commit_sha' .version-history.json  # Find commit for version

# Using yq (YAML)
yq '.entries[0]' .version-audit.yaml                      # Get latest bump
```

## Integration with Other Plugins

```yaml
plugins:
  changelog-generator:
    enabled: true
  tag-manager:
    enabled: true
  audit-log:
    enabled: true
```

Flow: Version bump -> Changelog generated -> Tag created -> Audit log entry recorded.

## Error Handling

The audit log plugin is **non-blocking**:

| Error Type         | Behavior                                          |
| ------------------ | ------------------------------------------------- |
| Git errors         | Warning displayed, entry created without metadata |
| File write errors  | Warning displayed, version bump still succeeds    |
| Missing git config | Entry created without author information          |

## Best Practices

1. **Enable early** - Start tracking from project inception
2. **Choose format wisely** - JSON for programmatic access, YAML for readability
3. **Store in version control** - Commit the log file to track team changes
4. **Don't edit manually** - Let the plugin manage the file

## Troubleshooting

| Issue                   | Solution                                            |
| ----------------------- | --------------------------------------------------- |
| Audit log not created   | Verify `enabled: true` and file permissions         |
| Missing metadata fields | Check `include-*` options and git configuration     |
| Log file corruption     | Back up and create new empty log: `{"entries": []}` |

## See Also

- [Example Configuration](./examples/audit-log.yaml) - Complete audit-log setup
- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Tag Manager](./TAG_MANAGER.md) - Create git tags for versions
- [Changelog Generator](./CHANGELOG_GENERATOR.md) - Generate changelogs from commits
