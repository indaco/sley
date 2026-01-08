# Tag Manager Plugin

The tag manager plugin automatically creates and manages git tags synchronized with version bumps. It validates tag availability before bumping and creates tags after successful version updates.

## Plugin Metadata

| Field       | Value                                            |
| ----------- | ------------------------------------------------ |
| Name        | `tag-manager`                                    |
| Version     | v0.1.0                                           |
| Type        | `tag-manager`                                    |
| Description | Manages git tags synchronized with version bumps |

## Status

Built-in, **disabled by default** (when running without configuration)

> **Note**: While tag-manager is disabled by default when running sley without a config file, it is included in the recommended starting configuration created by `sley init --yes`. This reflects the common workflow of versioning alongside git tagging.

## Features

- Automatic git tag creation after version bumps
- Pre-bump validation to ensure tag doesn't already exist
- Configurable tag prefix (`v`, `release-`, or custom)
- Support for annotated and lightweight tags
- Optional automatic push to remote repository
- Fail-fast behavior prevents version file updates when tags can't be created

## How It Works

1. Before a version bump, validates that the target tag doesn't already exist (fail-fast)
2. After a successful bump, creates a git tag for the new version
3. Optionally pushes the tag to the remote repository

## Configuration

Enable and configure in `.sley.yaml`:

```yaml
plugins:
  tag-manager:
    enabled: true # Enable the plugin (required)
    auto-create: true # Create tags automatically after bumps (default: true)
    prefix: "v" # Tag prefix (default: "v")
    annotate: true # Create annotated tags with message (default: true)
    push: false # Push tags to remote after creation (default: false)
```

### Configuration Options

| Option        | Type   | Default | Description                            |
| ------------- | ------ | ------- | -------------------------------------- |
| `enabled`     | bool   | false   | Enable/disable the plugin              |
| `auto-create` | bool   | true    | Automatically create tags after bumps  |
| `prefix`      | string | `"v"`   | Prefix for tag names                   |
| `annotate`    | bool   | true    | Create annotated tags (vs lightweight) |
| `push`        | bool   | false   | Push tags to remote after creation     |

## Tag Formats

| Version       | Prefix     | Tag Name         |
| ------------- | ---------- | ---------------- |
| 1.2.3         | `v`        | `v1.2.3`         |
| 1.2.3         | `release-` | `release-1.2.3`  |
| 1.2.3         | (empty)    | `1.2.3`          |
| 1.0.0-alpha.1 | `v`        | `v1.0.0-alpha.1` |

## Usage

Once enabled, the plugin works automatically with all bump commands.

### Basic Usage

```bash
# Bump patch version and create tag
sley bump patch
# Output: Version bumped from 1.2.3 to 1.2.4
# Output: Created tag: v1.2.4
```

### With Push Enabled

```bash
# With push: true in config
sley bump minor
# Output: Version bumped from 1.2.4 to 1.3.0
# Output: Created tag: v1.3.0
# Output: Pushed tag: v1.3.0
```

### With Pre-release

```bash
sley bump minor --pre alpha.1
# Output: Version bumped from 1.2.3 to 1.3.0-alpha.1
# Output: Created tag: v1.3.0-alpha.1
```

## Tag Validation (Fail-Fast)

The plugin validates tag availability **before** bumping:

```bash
# If v1.3.0 already exists:
sley bump minor
# Error: tag v1.3.0 already exists
# Version file remains unchanged
```

This fail-fast behavior prevents version file updates when the corresponding tag cannot be created.

## Annotated vs Lightweight Tags

### Annotated Tags (Default)

With `annotate: true`:

```bash
git tag -a v1.2.3 -m "Release 1.2.3 (patch bump)"
```

Annotated tags include:

- Author name and email
- Creation date
- Tag message
- GPG signature (if configured)

### Lightweight Tags

With `annotate: false`:

```bash
git tag v1.2.3
```

Lightweight tags are simple pointers to a commit without additional metadata.

**Recommendation**: Use annotated tags for releases. They provide better audit trails and are recommended by Git best practices.

## Common Configurations

### Release Workflow (CI/CD)

```yaml
plugins:
  tag-manager:
    enabled: true
    prefix: "v"
    annotate: true
    push: true # Automatically push to remote
```

### Development Workflow

```yaml
plugins:
  tag-manager:
    enabled: true
    prefix: "v"
    annotate: true
    push: false # Manual push control
```

### Custom Prefix

```yaml
plugins:
  tag-manager:
    enabled: true
    prefix: "release-" # Tags: release-1.2.3
    annotate: true
```

### No Prefix

```yaml
plugins:
  tag-manager:
    enabled: true
    prefix: "" # Tags: 1.2.3
    annotate: true
```

## Integration with Other Plugins

### With Version Validator

```yaml
plugins:
  version-validator:
    enabled: true
    rules:
      - type: "major-version-max"
        value: 10
  tag-manager:
    enabled: true
    prefix: "v"
```

Execution flow:

1. `version-validator` checks version policy
2. `tag-manager` validates tag doesn't exist
3. Version file updated
4. `tag-manager` creates tag

### With Commit Parser

```yaml
plugins:
  commit-parser: true
  tag-manager:
    enabled: true
    prefix: "v"
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

### With Dependency Check

```yaml
plugins:
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: "package.json"
        field: "version"
        format: "json"
  tag-manager:
    enabled: true
    prefix: "v"
```

Execution order:

1. `dependency-check` validates file consistency
2. Version file updated
3. `dependency-check` syncs package.json
4. `tag-manager` creates tag

## Error Handling

### Tag Already Exists

```bash
sley bump patch
# Error: tag v1.2.4 already exists
```

**Solution**: Delete the existing tag or bump to a different version.

### Git Not Available

```bash
sley bump patch
# Error: git command failed: exec: "git": executable file not found
```

**Solution**: Ensure git is installed and in PATH.

### Push Failed

```bash
sley bump patch
# Output: Created tag: v1.2.4
# Error: failed to push tag: remote rejected
```

**Solution**: Check remote permissions and network connectivity.

## Best Practices

1. **Use annotated tags**: They provide better metadata for releases
2. **Consistent prefix**: Choose a prefix and stick with it (`v` is most common)
3. **CI/CD push**: Enable `push: true` only in CI/CD pipelines
4. **Local development**: Keep `push: false` for local work
5. **Version validator first**: Combine with version-validator to prevent invalid tags

## Troubleshooting

### Tags Not Being Created

1. Verify plugin is enabled:

   ```yaml
   plugins:
     tag-manager:
       enabled: true
   ```

2. Check `auto-create` is true (default)

3. Verify you're in a git repository:
   ```bash
   git status
   ```

### Tags Not Pushing

1. Check `push: true` is set
2. Verify remote is configured:
   ```bash
   git remote -v
   ```
3. Check authentication/permissions

### Wrong Tag Format

Verify your prefix configuration:

```yaml
plugins:
  tag-manager:
    prefix: "v" # Results in v1.2.3
```

## See Also

- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Changelog Generator](./CHANGELOG_GENERATOR.md) - Generate changelogs after tagging
- [Version Validator](./VERSION_VALIDATOR.md) - Validate versions before tagging
