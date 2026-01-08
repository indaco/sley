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

1. Before a version bump, validates the new version against all configured rules
2. If any rule fails, the bump is rejected with a descriptive error message
3. Rules are evaluated in order; first failure stops the validation

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
      - type: "require-pre-release-for-0x"
        enabled: true
      - type: "branch-constraint"
        branch: "release/*"
        allowed: ["patch"]
```

## Available Rule Types

| Rule Type                    | Description                                         | Parameters                          |
| ---------------------------- | --------------------------------------------------- | ----------------------------------- |
| `pre-release-format`         | Validates pre-release label matches a regex pattern | `pattern`: regex                    |
| `major-version-max`          | Limits maximum major version number                 | `value`: int                        |
| `minor-version-max`          | Limits maximum minor version number                 | `value`: int                        |
| `patch-version-max`          | Limits maximum patch version number                 | `value`: int                        |
| `require-pre-release-for-0x` | Requires pre-release label for 0.x versions         | `enabled`: bool                     |
| `branch-constraint`          | Restricts bump types on specific branches           | `branch`: glob, `allowed`: []string |
| `no-major-bump`              | Disallows major version bumps                       | `enabled`: bool                     |
| `no-minor-bump`              | Disallows minor version bumps                       | `enabled`: bool                     |
| `no-patch-bump`              | Disallows patch version bumps                       | `enabled`: bool                     |

## Rule Examples

### Pre-release Format Validation

Only allow standard pre-release labels:

```yaml
rules:
  - type: "pre-release-format"
    pattern: "^(alpha|beta|rc)(\\.[0-9]+)?$"
```

```bash
sley bump minor --pre alpha.1   # OK
sley bump minor --pre beta      # OK
sley bump minor --pre rc.2      # OK
sley bump minor --pre preview   # Error: pre-release label "preview" does not match required pattern
sley bump minor --pre dev       # Error: pre-release label "dev" does not match required pattern
```

### Version Number Limits

Prevent version numbers from growing too large:

```yaml
rules:
  - type: "major-version-max"
    value: 10
  - type: "minor-version-max"
    value: 99
  - type: "patch-version-max"
    value: 999
```

```bash
# At version 10.0.0:
sley bump major
# Error: major version 11 exceeds maximum allowed value 10

# At version 1.99.0:
sley bump minor
# Error: minor version 100 exceeds maximum allowed value 99
```

### 0.x Pre-release Requirement

Ensure unstable versions always have pre-release labels:

```yaml
rules:
  - type: "require-pre-release-for-0x"
    enabled: true
```

```bash
# At version 0.1.0-alpha:
sley bump minor              # Error: version 0.x.x requires a pre-release label
sley bump minor --pre beta   # OK: 0.2.0-beta

# At version 1.0.0:
sley bump minor              # OK: 1.1.0 (rule only applies to 0.x)
```

### Branch-Based Constraints

Restrict bump types on specific branches:

```yaml
rules:
  - type: "branch-constraint"
    branch: "release/*"
    allowed: ["patch"]
  - type: "branch-constraint"
    branch: "main"
    allowed: ["minor", "patch"]
  - type: "branch-constraint"
    branch: "develop"
    allowed: ["major", "minor", "patch"]
```

```bash
# On branch release/1.0:
sley bump minor   # Error: bump type "minor" is not allowed on branch "release/1.0"
sley bump patch   # OK

# On branch main:
sley bump major   # Error: bump type "major" is not allowed on branch "main"
sley bump minor   # OK

# On branch develop:
sley bump major   # OK (all types allowed)
```

Branch patterns support glob syntax:

- `release/*` matches `release/1.0`, `release/2.0`, etc.
- `feature/*` matches `feature/auth`, `feature/api`, etc.
- `main` matches exactly `main`

### Disabling Bump Types

Temporarily or permanently disallow certain bump types:

```yaml
rules:
  - type: "no-major-bump"
    enabled: true # Freeze major version
```

```bash
sley bump major   # Error: major bumps are not allowed by policy
sley bump minor   # OK
sley bump patch   # OK
```

Useful for:

- Maintenance mode (only patch bumps)
- Feature freeze (no major bumps)
- Temporary restrictions during releases

## Multiple Rules

Rules are evaluated in order. All must pass for the bump to succeed:

```yaml
plugins:
  version-validator:
    enabled: true
    rules:
      - type: "major-version-max"
        value: 10
      - type: "pre-release-format"
        pattern: "^(alpha|beta|rc)(\\.[0-9]+)?$"
      - type: "require-pre-release-for-0x"
        enabled: true
      - type: "branch-constraint"
        branch: "main"
        allowed: ["minor", "patch"]
```

## Common Configurations

### Strict Release Workflow

```yaml
plugins:
  version-validator:
    enabled: true
    rules:
      # Only standard pre-release labels
      - type: "pre-release-format"
        pattern: "^(alpha|beta|rc)(\\.[0-9]+)?$"

      # Limit major version growth
      - type: "major-version-max"
        value: 10

      # Release branches: patch only
      - type: "branch-constraint"
        branch: "release/*"
        allowed: ["patch"]

      # Main branch: no major bumps
      - type: "branch-constraint"
        branch: "main"
        allowed: ["minor", "patch"]
```

### Early Development (0.x)

```yaml
plugins:
  version-validator:
    enabled: true
    rules:
      # All 0.x versions must have pre-release label
      - type: "require-pre-release-for-0x"
        enabled: true

      # Prepare for 1.0 stability
      - type: "major-version-max"
        value: 1
```

### Maintenance Mode

```yaml
plugins:
  version-validator:
    enabled: true
    rules:
      # Only bug fixes allowed
      - type: "no-major-bump"
        enabled: true
      - type: "no-minor-bump"
        enabled: true
```

## Integration with Other Plugins

The version validator runs **before** other plugins, ensuring invalid versions are never processed:

```yaml
plugins:
  version-validator:
    enabled: true
    rules:
      - type: "major-version-max"
        value: 10
  dependency-check:
    enabled: true
    auto-sync: true
  tag-manager:
    enabled: true
    prefix: "v"
```

Execution order:

```bash
sley bump major
# 1. version-validator: Checks if major version is within limit
# 2. dependency-check: Validates file consistency
# 3. tag-manager: Validates tag doesn't exist
# 4. Version file updated
# 5. dependency-check: Syncs files
# 6. tag-manager: Creates tag
```

If version validation fails, no subsequent operations occur.

## Error Messages

The plugin provides clear, actionable error messages:

```bash
# Pre-release format
Error: pre-release label "preview" does not match required pattern "^(alpha|beta|rc)(\\.[0-9]+)?$"

# Version max
Error: major version 11 exceeds maximum allowed value 10

# Branch constraint
Error: bump type "major" is not allowed on branch "main" (allowed: minor, patch)

# Bump type disabled
Error: major bumps are not allowed by policy
```

## Best Practices

1. **Start permissive**: Begin with fewer rules and add more as needed
2. **Document policies**: Explain why rules exist in team documentation
3. **Branch strategy alignment**: Match rules to your Git workflow
4. **Gradual enforcement**: Use `no-*-bump` rules temporarily during freezes
5. **Test rules**: Verify rules work as expected in a test environment

## Troubleshooting

### Rule Not Applying

1. Verify plugin is enabled:

   ```yaml
   plugins:
     version-validator:
       enabled: true
   ```

2. Check rule syntax:
   ```yaml
   rules:
     - type: "major-version-max" # Correct
       value: 10
   ```

### Branch Constraint Not Matching

1. Check current branch:

   ```bash
   git branch --show-current
   ```

2. Verify glob pattern matches:
   - `release/*` matches `release/1.0` but not `release`
   - `main` matches exactly `main`

### Regex Pattern Issues

Test your regex pattern:

```yaml
# Escape backslashes in YAML
pattern: "^(alpha|beta|rc)(\\.[0-9]+)?$"

# This matches: alpha, beta, rc, alpha.1, beta.2, rc.10
# This does NOT match: dev, preview, alpha1 (missing dot)
```

## See Also

- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Tag Manager](./TAG_MANAGER.md) - Create git tags after validation
- [Dependency Check](./DEPENDENCY_CHECK.md) - Sync versions across files
