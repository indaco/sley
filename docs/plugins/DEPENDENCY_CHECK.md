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
- Provides detailed inconsistency reports

## Configuration

Add the plugin configuration to your `.sley.yaml`:

```yaml
plugins:
  dependency-check:
    enabled: true
    auto-sync: true # Automatically update files during bumps
    files:
      - path: package.json
        field: version
        format: json
      - path: Chart.yaml
        field: version
        format: yaml
```

### Configuration Fields

- `enabled` (bool): Enable/disable the plugin
- `auto-sync` (bool): Automatically sync versions after bumps
- `files` (array): List of files to check/sync

### File Configuration

Each file entry supports:

- `path` (string, required): File path relative to repository root
- `format` (string, required): File format (`json`, `yaml`, `toml`, `raw`, `regex`)
- `field` (string, optional): Dot-notation path to version field (for JSON/YAML/TOML)
- `pattern` (string, optional): Regex pattern with capturing group (for `regex` format)

## Supported Formats

### JSON

Reads and writes JSON files with support for nested fields.

```yaml
files:
  - path: package.json
    field: version
    format: json

  # Nested field example
  - path: composer.json
    field: extra.version
    format: json
```

**Example package.json:**

```json
{
  "name": "my-app",
  "version": "1.2.3"
}
```

### YAML

Reads and writes YAML files with nested field support.

```yaml
files:
  - path: Chart.yaml
    field: version
    format: yaml

  # Nested field example
  - path: config.yaml
    field: app.version
    format: yaml
```

**Example Chart.yaml:**

```yaml
apiVersion: v2
name: my-chart
version: 1.2.3
```

### TOML

Reads and writes TOML files with nested section support.

```yaml
files:
  - path: pyproject.toml
    field: tool.poetry.version
    format: toml

  - path: Cargo.toml
    field: package.version
    format: toml
```

**Example pyproject.toml:**

```toml
[tool.poetry]
name = "my-package"
version = "1.2.3"
```

### Raw

Reads the entire file contents as the version string.

```yaml
files:
  - path: VERSION
    format: raw

  - path: version.txt
    format: raw
```

**Example VERSION:**

```
1.2.3
```

### Regex

Uses a regular expression pattern with a capturing group to extract and replace the version.

```yaml
files:
  - path: src/version.go
    format: regex
    pattern: 'const Version = "(.*?)"'

  - path: CMakeLists.txt
    format: regex
    pattern: 'project\(.*? VERSION (.*?)\)'
```

**Example version.go:**

```go
package version

const Version = "1.2.3"
```

The regex pattern must include exactly one capturing group `(.*?)` that matches the version string.

## Nested Field Syntax

For JSON, YAML, and TOML formats, use dot notation to access nested fields:

```yaml
files:
  - path: pyproject.toml
    field: tool.poetry.version
    format: toml
```

This accesses:

```toml
[tool.poetry]
version = "1.2.3"
```

Deeply nested fields are supported:

```yaml
field: metadata.project.info.version
```

## Behavior

### Validation (Before Bump)

When you run a bump command, the plugin validates that all configured files match the new version:

```bash
sley bump patch
```

If inconsistencies are detected:

```
Error: version inconsistencies detected:
  - package.json: expected 1.2.4, found 1.2.3 (format: json)
  - Chart.yaml: expected 1.2.4, found 1.2.2 (format: yaml)

Run with auto-sync enabled to fix automatically, or update files manually.
```

### Auto-Sync (After Bump)

When `auto-sync: true` is enabled, the plugin automatically updates all configured files after the `.version` file is bumped:

```bash
sley bump patch
```

Output:

```
Version bumped from 1.2.3 to 1.2.4
Synced version to 3 dependency file(s)
```

### Manual Validation

You can check consistency without bumping:

```bash
sley show
```

The plugin runs silently during show operations and only reports errors if inconsistencies exist.

## Version Normalization

The plugin normalizes version strings for comparison:

- `1.2.3` matches `v1.2.3`
- `2.0.0-alpha` matches `v2.0.0-alpha`

This prevents false positives when some files use the `v` prefix and others don't.

## Examples

### Node.js Project

```yaml
plugins:
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: package.json
        field: version
        format: json
      - path: package-lock.json
        field: version
        format: json
```

### Python Project

```yaml
plugins:
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: pyproject.toml
        field: tool.poetry.version
        format: toml
      - path: setup.py
        format: regex
        pattern: 'version="(.*?)"'
```

### Rust Project

```yaml
plugins:
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: Cargo.toml
        field: package.version
        format: toml
```

### Kubernetes Helm Chart

```yaml
plugins:
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: Chart.yaml
        field: version
        format: yaml
      - path: Chart.yaml
        field: appVersion
        format: yaml
```

### Multi-Language Monorepo

```yaml
plugins:
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      # Frontend
      - path: frontend/package.json
        field: version
        format: json

      # Backend
      - path: backend/Cargo.toml
        field: package.version
        format: toml

      # Infrastructure
      - path: infrastructure/Chart.yaml
        field: version
        format: yaml

      # Build metadata
      - path: VERSION
        format: raw

      # Source code constant
      - path: backend/src/version.rs
        format: regex
        pattern: 'pub const VERSION: &str = "(.*?)";'
```

## Workflow Integration

### With Version Validator

```yaml
plugins:
  version-validator:
    enabled: true
    rules:
      - type: pre-release-format
        pattern: '^(alpha|beta|rc)\.\d+$'

  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: package.json
        field: version
        format: json
```

Execution order during bump:

1. Version validator checks new version
2. Dependency check validates consistency
3. `.version` file is updated
4. Dependency files are synced (if auto-sync enabled)

### With Tag Manager

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
    auto-create: true
    prefix: v
```

Execution order during bump:

1. Dependency check validates consistency
2. `.version` file is updated
3. Dependency files are synced
4. Git tag is created

## Error Handling

### File Not Found

If a configured file doesn't exist, the plugin reports an error:

```
Error: dependency check failed: failed to read version from package.json: failed to read file: open package.json: no such file or directory
```

### Invalid Format

If a file cannot be parsed:

```
Error: dependency check failed: failed to read version from package.json: failed to parse JSON: invalid character '}' looking for beginning of object key string
```

### Pattern Mismatch

For regex format, if the pattern doesn't match:

```
Error: dependency check failed: failed to read version from version.go: no version match found (pattern must have capturing group)
```

## Best Practices

1. **Start with validation only**: Set `auto-sync: false` initially to verify the plugin detects all files correctly
2. **Test regex patterns**: Use a regex tester to validate your patterns before adding them
3. **Version file first**: Always update `.version` file first, then let auto-sync handle other files
4. **Commit atomically**: When using auto-sync, commit all changed files together
5. **CI validation**: Add a CI check that runs `sley show` to catch inconsistencies

## Troubleshooting

### Plugin Not Running

Check that the plugin is enabled in `.sley.yaml`:

```yaml
plugins:
  dependency-check:
    enabled: true
```

### Files Not Syncing

Verify `auto-sync: true` is set:

```yaml
plugins:
  dependency-check:
    enabled: true
    auto-sync: true
```

### Regex Pattern Not Matching

Test your pattern with a simpler version:

```yaml
# Instead of complex pattern
pattern: 'project\(.*? VERSION (.*?)\)'

# Try simpler pattern first
pattern: 'VERSION (.*?)'
```

Ensure your pattern has exactly one capturing group `(.*?)`.

## Limitations

- Regex patterns must have exactly one capturing group
- File modifications use basic formatting (JSON indented with 2 spaces)
- Large files may have performance impact
- Binary files are not supported

## See Also

- [Example Configuration](./examples/dependency-check.yaml) - Multi-format file sync examples
- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Version Validator](./VERSION_VALIDATOR.md) - Validate versions before syncing
- [Tag Manager](./TAG_MANAGER.md) - Create git tags after syncing
