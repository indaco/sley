# Version Validator Plugin

The version validator plugin enforces versioning policies and constraints beyond basic SemVer syntax validation. It validates version bumps against configurable rules before the version file is updated.

## Plugin Metadata

| Field       | Value                                        |
| ----------- | -------------------------------------------- |
| Name        | `version-validator`                          |
| Version     | v0.1.0                                       |
| Type        | `version-validator`                          |
| Description | Enforces versioning policies and constraints |

## Status

Built-in, **disabled by default**

## Features

- Pre-bump validation with fail-fast behavior
- Multiple configurable rule types
- Branch-based constraints for release workflows
- Version number limits (major, minor, patch)
- Pre-release label format enforcement
- Bump type restrictions

## How It Works

1. Before bump: validates the new version against all configured rules
2. If any rule fails, the bump is rejected with a descriptive error
3. Rules are evaluated in order; first failure stops validation

## Configuration

Enable and configure in `.sley.yaml`:

```yaml
plugins:
  version-validator:
    enabled: true
    rules:
      - type: "pre-release-format"
        pattern: "^(alpha|beta|rc)(\\.[0-9]+)?$"
      - type: "major-version-max"
        value: 10
      - type: "branch-constraint"
        branch: "main"
        allowed: ["minor", "patch"]
```

## Available Rule Types

| Rule Type                    | Description                                 | Parameters                          |
| ---------------------------- | ------------------------------------------- | ----------------------------------- |
| `pre-release-format`         | Validates pre-release label matches regex   | `pattern`: regex                    |
| `major-version-max`          | Limits maximum major version number         | `value`: int                        |
| `minor-version-max`          | Limits maximum minor version number         | `value`: int                        |
| `patch-version-max`          | Limits maximum patch version number         | `value`: int                        |
| `require-pre-release-for-0x` | Requires pre-release label for 0.x versions | `enabled`: bool                     |
| `branch-constraint`          | Restricts bump types on specific branches   | `branch`: glob, `allowed`: []string |
| `no-major-bump`              | Disallows major version bumps               | `enabled`: bool                     |
| `no-minor-bump`              | Disallows minor version bumps               | `enabled`: bool                     |
| `no-patch-bump`              | Disallows patch version bumps               | `enabled`: bool                     |

## Rule Examples

### Pre-release Format

```yaml
rules:
  - type: "pre-release-format"
    pattern: "^(alpha|beta|rc)(\\.[0-9]+)?$"
```

```bash
sley bump minor --pre alpha.1   # OK
sley bump minor --pre preview   # Error: does not match pattern
```

### Version Number Limits

```yaml
rules:
  - type: "major-version-max"
    value: 10
```

```bash
# At version 10.0.0:
sley bump major  # Error: major version 11 exceeds maximum 10
```

### Branch Constraints

```yaml
rules:
  - type: "branch-constraint"
    branch: "release/*"
    allowed: ["patch"]
  - type: "branch-constraint"
    branch: "main"
    allowed: ["minor", "patch"]
```

```bash
# On release/1.0:
sley bump minor  # Error: not allowed on this branch
sley bump patch  # OK
```

Branch patterns support glob syntax: `release/*`, `feature/*`, `main`

### Disabling Bump Types

```yaml
rules:
  - type: "no-major-bump"
    enabled: true # Maintenance mode: only minor/patch
```

## Integration with Other Plugins

The version validator runs **before** other plugins:

```yaml
plugins:
  version-validator:
    enabled: true
    rules:
      - type: "major-version-max"
        value: 10
  dependency-check:
    enabled: true
  tag-manager:
    enabled: true
```

Flow: version-validator checks -> dependency-check validates -> version updated -> files synced -> tag created

## Error Messages

```bash
# Pre-release format
Error: pre-release label "preview" does not match required pattern

# Version max
Error: major version 11 exceeds maximum allowed value 10

# Branch constraint
Error: bump type "major" is not allowed on branch "main"

# Bump type disabled
Error: major bumps are not allowed by policy
```

## Best Practices

1. **Start permissive** - Begin with fewer rules, add more as needed
2. **Document policies** - Explain why rules exist
3. **Branch strategy alignment** - Match rules to your Git workflow
4. **Gradual enforcement** - Use `no-*-bump` rules temporarily during freezes

## Troubleshooting

| Issue                      | Solution                                           |
| -------------------------- | -------------------------------------------------- |
| Rule not applying          | Verify `enabled: true` and rule syntax             |
| Branch constraint mismatch | Check `git branch --show-current` and glob pattern |
| Regex pattern issues       | Escape backslashes in YAML: `\\.[0-9]+`            |

## See Also

- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Tag Manager](./TAG_MANAGER.md) - Create git tags after validation
- [Dependency Check](./DEPENDENCY_CHECK.md) - Sync versions across files
