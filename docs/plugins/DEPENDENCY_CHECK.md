# Dependency Check Plugin

The dependency check plugin validates and synchronizes version numbers across multiple files in your repository. This ensures consistency between your `.version` file and other version declarations in package manifests, build files, and source code.

## Plugin Metadata

| Field       | Value                                |
| ----------- | ------------------------------------ |
| Name        | `dependency-check`                   |
| Version     | v0.1.0                               |
| Type        | `dependency-check`                   |
| Description | Syncs version across dependent files |

## Status

Built-in, **disabled by default**

## Features

- Validates version consistency across multiple file formats
- Automatically syncs versions during bumps (optional)
- Supports JSON, YAML, TOML, raw text, and regex patterns
- Handles nested fields with dot notation
- Normalizes version formats (e.g., `1.2.3` matches `v1.2.3`)

## Configuration

Enable and configure in `.sley.yaml`. See [dependency-check.yaml](./examples/dependency-check.yaml) for complete examples.

```yaml
plugins:
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: package.json
        field: version
        format: json
      - path: Chart.yaml
        field: version
        format: yaml
```

### Configuration Options

| Option      | Type  | Default | Description                            |
| ----------- | ----- | ------- | -------------------------------------- |
| `enabled`   | bool  | false   | Enable/disable the plugin              |
| `auto-sync` | bool  | false   | Automatically sync versions after bump |
| `files`     | array | []      | List of files to check/sync            |

### File Configuration

| Field     | Type   | Required | Description                                           |
| --------- | ------ | -------- | ----------------------------------------------------- |
| `path`    | string | yes      | File path relative to repository root                 |
| `format`  | string | yes      | File format: `json`, `yaml`, `toml`, `raw`, `regex`   |
| `field`   | string | no       | Dot-notation path to version field (JSON/YAML/TOML)   |
| `pattern` | string | no       | Regex pattern with capturing group (for regex format) |

## Supported Formats

| Format  | Field Required | Pattern Required | Example Use Case                   |
| ------- | -------------- | ---------------- | ---------------------------------- |
| `json`  | yes            | no               | `package.json`, `composer.json`    |
| `yaml`  | yes            | no               | `Chart.yaml`, `pubspec.yaml`       |
| `toml`  | yes            | no               | `Cargo.toml`, `pyproject.toml`     |
| `raw`   | no             | no               | `VERSION` file (entire content)    |
| `regex` | no             | yes              | Source code constants, build files |

### Format Examples

```yaml
files:
  # JSON with nested field
  - path: package.json
    field: version
    format: json

  # YAML
  - path: Chart.yaml
    field: version
    format: yaml

  # TOML with nested section
  - path: pyproject.toml
    field: tool.poetry.version
    format: toml

  # Raw file (entire content is version)
  - path: VERSION
    format: raw

  # Regex pattern (must have one capturing group)
  - path: src/version.go
    format: regex
    pattern: 'const Version = "(.*?)"'
```

## Behavior

### Validation (Before Bump)

The plugin validates that all configured files can be updated. If inconsistencies or errors are detected, the bump is aborted.

### Auto-Sync (After Bump)

With `auto-sync: true`, files are automatically updated after the `.version` file is bumped:

```bash
sley bump patch
# Version bumped from 1.2.3 to 1.2.4
# Synced version to 3 dependency file(s)
```

## Version Normalization

The plugin normalizes version strings for comparison:

- `1.2.3` matches `v1.2.3`
- `2.0.0-alpha` matches `v2.0.0-alpha`

## Integration with Other Plugins

```yaml
plugins:
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: package.json
        field: version
        format: json
  tag-manager:
    enabled: true
    prefix: "v"
```

Flow: dependency-check validates -> version updated -> files synced -> tag created

## Error Handling

| Error Type       | Behavior                                       |
| ---------------- | ---------------------------------------------- |
| File not found   | Bump aborted with error                        |
| Invalid format   | Bump aborted with parse error                  |
| Pattern mismatch | Bump aborted (regex must have capturing group) |

## Best Practices

1. **Start with validation only** - Set `auto-sync: false` initially
2. **Test regex patterns** - Validate patterns before adding
3. **Commit atomically** - Commit all changed files together
4. **CI validation** - Add `sley show` to CI to catch inconsistencies

## Troubleshooting

| Issue              | Solution                                              |
| ------------------ | ----------------------------------------------------- |
| Plugin not running | Verify `enabled: true` in configuration               |
| Files not syncing  | Check `auto-sync: true` is set                        |
| Regex not matching | Ensure pattern has exactly one capturing group `(.*)` |

## Limitations

- Regex patterns must have exactly one capturing group
- File modifications use basic formatting (JSON: 2-space indent)
- Binary files are not supported

## See Also

- [Example Configuration](./examples/dependency-check.yaml) - Multi-format file sync examples
- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Version Validator](./VERSION_VALIDATOR.md) - Validate versions before syncing
- [Tag Manager](./TAG_MANAGER.md) - Create git tags after syncing
