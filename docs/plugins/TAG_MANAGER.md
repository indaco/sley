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

Once enabled, the plugin works automatically:

```bash
sley bump patch
# Version bumped from 1.2.3 to 1.2.4
# Created tag: v1.2.4

# With push: true
sley bump minor
# Version bumped from 1.2.4 to 1.3.0
# Created tag: v1.3.0
# Pushed tag: v1.3.0
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
