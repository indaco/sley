# sley Extensions

This directory contains example extensions for sley. These extensions demonstrate how to build extensions in different languages and can be used as templates for your own.

For the complete extension authoring guide, see [docs/EXTENSIONS.md](../../docs/EXTENSIONS.md).

## Extensions vs Plugins

sley provides two extensibility mechanisms:

| Aspect            | Built-in Plugins     | Extensions                            |
| ----------------- | -------------------- | ------------------------------------- |
| **Language**      | Go (native)          | Any (Python, Bash, Go, JS, etc.)      |
| **Performance**   | Fastest              | Varies by language                    |
| **Dependencies**  | None                 | May require runtime (Python, Node.js) |
| **Configuration** | `plugins:` section   | `extensions:` section                 |
| **Use case**      | Production workflows | Custom integrations, examples         |

**Recommendation:** Use built-in plugins when available. Use extensions for:

- Custom integrations not covered by plugins
- Learning how the extension system works
- Prototyping before contributing a plugin

## Available Extensions

### 1. docker-tag-sync (Bash)

**Hook**: `post-bump`
Tags and pushes Docker images with the new version.

Features:

- Docker image tagging
- Optional push to registry
- Configurable image name

[View Documentation](./docker-tag-sync/README.md)

---

### 2. commit-validator (Python)

**Hook**: `pre-bump`
Validates commits follow conventional commit format.

Features:

- Validates commits since last tag
- Configurable allowed types
- Optional scope requirement
- Blocks bump on invalid commits

> [!NOTE]
> This complements the `commitparser` plugin. While `commitparser` parses commits to infer bump type (permissive), `commit-validator` enforces strict format compliance (blocks on invalid).

[View Documentation](./commit-validator/README.md)

---

## Language Examples

These extensions demonstrate different implementation languages:

| Extension        | Language | Runtime | Startup Time |
| ---------------- | -------- | ------- | ------------ |
| docker-tag-sync  | Bash     | sh      | <10ms        |
| commit-validator | Python 3 | python3 | ~50ms        |

## Installing Extensions

### From Local Path

```bash
sley extension install --path ./contrib/extensions/docker-tag-sync
```

### From URL

```bash
sley extension install --url https://github.com/user/my-extension
```

### Configuration

After installation, configure in `.sley.yaml`:

```yaml
extensions:
  - name: docker-tag-sync
    enabled: true
    hooks:
      - post-bump
    config:
      image: myapp
      push: true
```

### Managing Extensions

```bash
# List installed extensions
sley extension list

# Remove an extension
sley extension remove docker-tag-sync
```

## Plugin Integration

Extensions work seamlessly with built-in plugins for complete automation.

### Example: Strict Validation Workflow

```yaml
# .sley.yaml
plugins:
  commit-parser: true # Built-in commit analysis
  tag-manager:
    enabled: true
    prefix: "v"
  version-validator:
    enabled: true
    rules:
      - type: require-even-minor
        enabled: true
      - type: max-prerelease-iterations
        value: 10

extensions:
  - name: commit-validator # Strict format validation
    enabled: true
    hooks: [pre-bump]
    config:
      require_scope: true
```

```bash
sley bump auto
# 1. commit-validator: Blocks if commits are invalid
# 2. version-validator: Validates version policies
# 3. commitparser: Analyzes commits -> determines bump type
# 4. Version bumped
# 5. tag-manager: Creates git tag
```

See [docs/PLUGINS.md](../../docs/PLUGINS.md) for detailed plugin documentation.

## Creating Your Own Extension

See the [Extension System documentation](../../docs/EXTENSIONS.md) for:

- Directory structure and manifest format
- JSON input/output specification
- Hook points reference
- Code examples in multiple languages
- Best practices and troubleshooting

## Contributing

Want to contribute an extension?

1. Follow the structure in [docs/EXTENSIONS.md](../../docs/EXTENSIONS.md)
2. Include comprehensive documentation
3. Add tests to `test-extensions.sh`
4. Minimize external dependencies

## License

All extensions in this directory are licensed under the same terms as sley.
