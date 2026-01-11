# Tag Manager Plugin

The tag manager plugin automatically creates and manages git tags synchronized with version bumps. It validates tag availability before bumping and creates tags after successful version updates.

## Plugin Metadata

| Field       | Value                                            |
| ----------- | ------------------------------------------------ |
| Name        | `tag-manager`                                    |
| Type        | `automation`                                     |
| Description | Manages git tags synchronized with version bumps |

## Status

Built-in, **disabled by default**

> **Note**: While disabled by default, tag-manager is included in the recommended configuration created by `sley init --yes`.

## Features

- Automatic git tag creation after version bumps
- Pre-bump validation to ensure tag doesn't already exist
- Configurable tag prefix (`v`, `release-`, or custom)
- Support for annotated and lightweight tags
- Optional automatic push to remote repository

## How It Works

1. Before bump: validates that the target tag doesn't already exist (fail-fast)
2. After bump: creates a git tag for the new version
3. Optionally pushes the tag to the remote repository

## Configuration

Enable and configure in `.sley.yaml`:

```yaml
plugins:
  tag-manager:
    enabled: true
    auto-create: true
    prefix: "v"
    annotate: true
    push: false
    tag-prereleases: true # Set to false to skip tagging pre-releases
```

### Configuration Options

| Option            | Type   | Default | Description                            |
| ----------------- | ------ | ------- | -------------------------------------- |
| `enabled`         | bool   | false   | Enable/disable the plugin              |
| `auto-create`     | bool   | true    | Automatically create tags after bumps  |
| `prefix`          | string | `"v"`   | Prefix for tag names                   |
| `annotate`        | bool   | true    | Create annotated tags (vs lightweight) |
| `push`            | bool   | false   | Push tags to remote after creation     |
| `tag-prereleases` | bool   | true    | Create tags for pre-release versions   |

#### Pre-release Tagging Behavior

The `tag-prereleases` option controls whether git tags are created for pre-release versions (e.g., `1.0.0-alpha.1`, `2.0.0-rc.1`):

- **`true` (default)**: Tags are created for all versions, including pre-releases. This is useful when you want to track all version changes in git history.

- **`false`**: Tags are only created for stable releases (versions without pre-release identifiers). Pre-release version bumps will update the `.version` file but skip tag creation. This is useful when:
  - You want to keep your tag list clean and only show stable releases
  - Pre-release versions are experimental and shouldn't be tagged
  - You're using a CI/CD workflow that only needs tags for production releases

**Example with `tag-prereleases: false`:**

```bash
# Pre-release bumps - no tags created
sley bump pre alpha
# => 1.0.0-alpha.1 (no tag)

sley bump pre alpha
# => 1.0.0-alpha.2 (no tag)

# Stable release - tag created
sley bump release
# => 1.0.0 (tag: v1.0.0)
```

**Example with `tag-prereleases: true` (default):**

```bash
# All bumps create tags
sley bump pre beta
# => 1.1.0-beta.1 (tag: v1.1.0-beta.1)

sley bump release
# => 1.1.0 (tag: v1.1.0)
```

## Tag Formats

| Version       | Prefix     | Tag Name         |
| ------------- | ---------- | ---------------- |
| 1.2.3         | `v`        | `v1.2.3`         |
| 1.2.3         | `release-` | `release-1.2.3`  |
| 1.2.3         | (empty)    | `1.2.3`          |
| 1.0.0-alpha.1 | `v`        | `v1.0.0-alpha.1` |

## Usage

Once enabled, the plugin works automatically:

```bash
sley bump patch
# => 1.2.4 (tag: v1.2.4)

# With push: true
sley bump minor
# => 1.3.0 (tag: v1.3.0, pushed)
```

## Tag Validation (Fail-Fast)

The plugin validates tag availability **before** bumping:

```bash
# If v1.3.0 already exists:
sley bump minor
# Error: tag v1.3.0 already exists
# Version file remains unchanged
```

## Annotated vs Lightweight Tags

**Annotated tags** (default, `annotate: true`):

- Include author, date, message, and optional GPG signature
- Recommended for releases

**Lightweight tags** (`annotate: false`):

- Simple pointers to a commit
- No additional metadata

## Integration with Other Plugins

```yaml
plugins:
  commit-parser: true
  tag-manager:
    enabled: true
    prefix: "v"
    push: true
```

Flow: commit-parser analyzes -> tag-manager validates -> version updated -> tag created and pushed

### Example: Production-Only Tagging

```yaml
plugins:
  tag-manager:
    enabled: true
    prefix: "v"
    annotate: true
    push: true
    tag-prereleases: false # Only tag stable releases
```

This configuration is useful for CI/CD pipelines where:

- Pre-release versions are used for testing/staging
- Only stable releases should be tagged and pushed to production
- Tag list should remain clean and only show production-ready versions

## Error Handling

| Error Type         | Behavior                                  |
| ------------------ | ----------------------------------------- |
| Tag already exists | Bump aborted, version file unchanged      |
| Git not available  | Error: executable not found               |
| Push failed        | Tag created locally, push error displayed |

## Best Practices

1. **Use annotated tags** - Better metadata for releases
2. **Consistent prefix** - Choose one and stick with it (`v` is most common)
3. **CI/CD push** - Enable `push: true` only in CI/CD pipelines
4. **Local development** - Keep `push: false` for local work
5. **Clean tag list** - Use `tag-prereleases: false` if you only want tags for stable releases
6. **Pre-release tracking** - Keep `tag-prereleases: true` (default) if you need to track all version changes in git

## Troubleshooting

| Issue            | Solution                                              |
| ---------------- | ----------------------------------------------------- |
| Tags not created | Verify `enabled: true` and you're in a git repository |
| Tags not pushing | Check `push: true` and remote configuration           |
| Wrong tag format | Verify `prefix` configuration                         |

## See Also

- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Changelog Generator](./CHANGELOG_GENERATOR.md) - Generate changelogs after tagging
- [Version Validator](./VERSION_VALIDATOR.md) - Validate versions before tagging
