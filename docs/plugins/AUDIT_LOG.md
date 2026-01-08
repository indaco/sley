# Audit Log Plugin

The audit log plugin records all version changes to a persistent log file with metadata such as author, timestamp, commit SHA, and branch. This provides a complete audit trail of version history for compliance, debugging, and rollback scenarios.

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
- Sorted entries (newest first) for easy browsing
- Non-blocking: failures don't prevent version bumps
- Selective metadata inclusion via configuration

## How It Works

1. After a successful version bump, the plugin is called
2. Git metadata is collected based on configuration
3. A new entry is created with version change details
4. The log file is read (or created if missing)
5. The new entry is appended and entries are sorted
6. The updated log is written back to disk

If the audit log write fails, a warning is displayed but the version bump succeeds. This ensures that logging issues never block releases.

## Configuration

Enable and configure in `.sley.yaml`:

```yaml
plugins:
  audit-log:
    enabled: true # Enable the plugin (required)
    path: .version-history.json # Path to audit log file (default: .version-history.json)
    format: json # Output format: json or yaml (default: json)
    include-author: true # Include git author (default: true)
    include-timestamp: true # Include ISO 8601 timestamp (default: true)
    include-commit-sha: true # Include current commit SHA (default: true)
    include-branch: true # Include current branch name (default: true)
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
    },
    {
      "timestamp": "2026-01-03T10:30:00Z",
      "previous_version": "1.2.2",
      "new_version": "1.2.3",
      "bump_type": "patch",
      "author": "Jane Smith <jane@example.com>",
      "commit_sha": "def9876543210abc",
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
  - timestamp: "2026-01-03T10:30:00Z"
    previous_version: "1.2.2"
    new_version: "1.2.3"
    bump_type: patch
    author: Jane Smith <jane@example.com>
    commit_sha: def9876543210abc
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

Once enabled, the plugin works automatically with all bump commands.

### Basic Usage

```bash
# Enable audit log
cat > .sley.yaml << EOF
plugins:
  audit-log:
    enabled: true
EOF

# Bump version
sley bump patch
# Output: Version bumped from 1.2.3 to 1.2.4

# Check audit log
cat .version-history.json
```

### With Custom Path

```yaml
plugins:
  audit-log:
    enabled: true
    path: .changes/version-audit.json
```

### With YAML Format

```yaml
plugins:
  audit-log:
    enabled: true
    format: yaml
    path: .version-audit.yaml
```

### Minimal Metadata

```yaml
plugins:
  audit-log:
    enabled: true
    include-author: false
    include-commit-sha: false
    include-branch: false
    # Only timestamp and version changes will be recorded
```

## Use Cases

### Compliance and Auditing

Track all version changes for compliance requirements:

```yaml
plugins:
  audit-log:
    enabled: true
    path: audit/version-history.json
    include-author: true
    include-timestamp: true
    include-commit-sha: true
```

### Debugging Version History

Investigate when and why versions changed:

```bash
# Find all major version bumps
jq '.entries[] | select(.bump_type == "major")' .version-history.json

# Find bumps by specific author
jq '.entries[] | select(.author | contains("john@example.com"))' .version-history.json

# Find bumps on specific branch
jq '.entries[] | select(.branch == "release/v2")' .version-history.json
```

### Rollback Information

Identify exact commit for version rollback:

```bash
# Find commit SHA for a specific version
jq '.entries[] | select(.new_version == "1.2.3") | .commit_sha' .version-history.json
```

### CI/CD Tracking

Monitor automated version bumps in CI:

```yaml
plugins:
  audit-log:
    enabled: true
    path: .version-history.json
    include-branch: true # Track which pipeline branch
    include-timestamp: true
```

## Integration with Other Plugins

### With Tag Manager

```yaml
plugins:
  tag-manager:
    enabled: true
    prefix: "v"
    push: true
  audit-log:
    enabled: true
```

Workflow:

1. Version bump succeeds
2. Audit log records the change
3. Tag manager creates and pushes tag

### With Changelog Generator

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: versioned
  audit-log:
    enabled: true
```

Workflow:

1. Version bump succeeds
2. Changelog generator creates changelog entry
3. Audit log records the change with metadata

### With Dependency Check

```yaml
plugins:
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: package.json
        field: version
        format: json
  audit-log:
    enabled: true
```

Workflow:

1. Version bump succeeds
2. Dependency check syncs package.json
3. Audit log records the change

## Error Handling

The audit log plugin is designed to be **non-blocking**:

### Git Errors

If git commands fail (e.g., not in a git repository):

```bash
sley bump patch
# Warning: failed to enrich audit log entry: git command failed
# Version bumped from 1.2.3 to 1.2.4
# (Audit log entry created but without git metadata)
```

### File Write Errors

If the log file can't be written:

```bash
sley bump patch
# Warning: failed to write audit log: permission denied
# Version bumped from 1.2.3 to 1.2.4
# (Version bump succeeds despite logging failure)
```

### Missing Git Configuration

If git user.name or user.email are not configured:

```bash
sley bump patch
# Warning: failed to enrich audit log entry: git config user.name not set
# Version bumped from 1.2.3 to 1.2.4
# (Entry created without author information)
```

## Querying the Audit Log

### Using jq (JSON)

```bash
# Count total bumps
jq '.entries | length' .version-history.json

# Get latest 5 bumps
jq '.entries[:5]' .version-history.json

# Find all minor bumps
jq '.entries[] | select(.bump_type == "minor")' .version-history.json

# Group by branch
jq 'group_by(.branch)' .version-history.json

# Find bumps in date range
jq '.entries[] | select(.timestamp >= "2026-01-01" and .timestamp <= "2026-12-31")' .version-history.json
```

### Using yq (YAML)

```bash
# Get latest bump
yq '.entries[0]' .version-audit.yaml

# Find bumps by author
yq '.entries[] | select(.author | contains("john"))' .version-audit.yaml

# Count patch bumps
yq '.entries[] | select(.bump_type == "patch") | length' .version-audit.yaml
```

## Best Practices

1. **Enable early**: Start tracking from the beginning of your project
2. **Choose format wisely**: JSON for programmatic access, YAML for readability
3. **Store in version control**: Commit `.version-history.json` to track team changes
4. **Regular backups**: Back up audit logs for long-term compliance
5. **Combine with changelog**: Use both for complete version history
6. **Query efficiently**: Use jq/yq for analysis rather than manual inspection
7. **Don't edit manually**: Let the plugin manage the file to ensure consistency

## Common Configurations

### Production Environment

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

### Development Environment

```yaml
plugins:
  audit-log:
    enabled: true
    path: .version-history.json
    format: json
    include-author: true
    include-timestamp: true
```

### Minimal Setup

```yaml
plugins:
  audit-log:
    enabled: true
    # All other options use defaults
```

## Troubleshooting

### Audit Log Not Created

1. Verify plugin is enabled:

   ```yaml
   plugins:
     audit-log:
       enabled: true
   ```

2. Check file permissions in the target directory

3. Verify you're running a bump command (show/get don't create entries)

### Missing Metadata Fields

If certain fields are missing from entries:

1. Check configuration options are set to `true`
2. Verify git is installed and accessible
3. Check git configuration:
   ```bash
   git config user.name
   git config user.email
   ```

### Log File Corruption

If the log file becomes corrupted:

1. Back up the current file
2. Create a new empty log:
   ```json
   {
     "entries": []
   }
   ```
3. Future bumps will append to the new log

### Large Log Files

For projects with many version bumps:

1. Consider archiving old entries periodically
2. Use log rotation tools
3. Query specific date ranges instead of loading entire file

## Migration Guide

### From Manual Tracking

If you've been tracking versions manually:

1. Enable the plugin:

   ```yaml
   plugins:
     audit-log:
       enabled: true
   ```

2. Create initial log file with historical data (optional):

   ```json
   {
     "entries": [
       {
         "timestamp": "2025-12-01T10:00:00Z",
         "previous_version": "1.0.0",
         "new_version": "1.1.0",
         "bump_type": "minor"
       }
     ]
   }
   ```

3. Future bumps will be automatically tracked

### Changing Formats

To switch from JSON to YAML:

1. Update configuration:

   ```yaml
   plugins:
     audit-log:
       enabled: true
       format: yaml
       path: .version-history.yaml
   ```

2. Convert existing log:
   ```bash
   yq eval -P .version-history.json > .version-history.yaml
   rm .version-history.json
   ```

## See Also

- [Tag Manager](./TAG_MANAGER.md) - Create git tags for versions
- [Changelog Generator](./CHANGELOG_GENERATOR.md) - Generate changelogs from commits
- [Dependency Check](./DEPENDENCY_CHECK.md) - Sync version across files
- [Release Gate](./RELEASE_GATE.md) - Validate releases before bumping
