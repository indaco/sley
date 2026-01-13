# Extension Quick Start Guide

Quick reference for using sley extensions.

> [!TIP]
> For most use cases, prefer built-in plugins over extensions. Extensions are best for custom integrations or learning how the extension system works. See [README.md](./README.md) for details.

## Installation

```bash
# Install a single extension
sley extension install ./contrib/extensions/docker-tag-sync

# List installed extensions
sley extension list

# Remove an extension
sley extension remove docker-tag-sync
```

## Configuration

Add to your `.sley.yaml`:

```yaml
extensions:
  - name: docker-tag-sync
    enabled: true
    hooks:
      - post-bump
    config:
      image: "myapp"
      push: true
```

## Available Extensions

| Extension        | Language | Hook      | Purpose                        |
| ---------------- | -------- | --------- | ------------------------------ |
| docker-tag-sync  | Bash     | post-bump | Tag Docker images with version |
| commit-validator | Python   | pre-bump  | Validate commit message format |

## Common Workflows

### Docker Image Tagging

```yaml
extensions:
  - name: docker-tag-sync
    enabled: true
    hooks:
      - post-bump
    config:
      image: "myapp"
      push: true
      registry: "ghcr.io/myorg"
```

### Strict Commit Validation

```yaml
extensions:
  - name: commit-validator
    enabled: true
    hooks:
      - pre-bump
    config:
      allowed_types: [feat, fix, docs, refactor]
      require_scope: true
```

### Combined with Plugins

```yaml
# Use plugins for core functionality
plugins:
  commit-parser: true
  tag-manager:
    enabled: true
    prefix: "v"
    push: true
  version-validator:
    enabled: true
    rules:
      - type: require-even-minor
        enabled: true
      - type: max-prerelease-iterations
        value: 10

# Use extensions for additional validation
extensions:
  - name: commit-validator
    enabled: true
    hooks: [pre-bump]
    config:
      require_scope: true

  - name: docker-tag-sync
    enabled: true
    hooks: [post-bump]
    config:
      image: "myapp"
```

Then:

```bash
sley bump auto
# 1. commit-validator: Validates commit format
# 2. version-validator: Validates version policies
# 3. commitparser: Analyzes commits -> determines bump type
# 4. Version bumped
# 5. tag-manager: Creates and pushes git tag
# 6. docker-tag-sync: Tags Docker image
```

## Documentation

- Overview: `contrib/extensions/README.md`
- Individual extension docs: `contrib/extensions/<name>/README.md`
- Extension authoring: `docs/EXTENSIONS.md`
- Built-in plugins: `docs/PLUGINS.md`

## Support

For detailed configuration and troubleshooting, see the README.md in each extension directory.
