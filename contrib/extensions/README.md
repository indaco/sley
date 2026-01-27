# sley Extensions

This directory contains example extensions demonstrating how to build sley extensions in different languages.

> [!TIP]
> For complete documentation, see the [Extensions Guide](https://sley.indaco.dev/extensions/) on the sley documentation website.

## Available Extensions

| Extension                                     | Language | Hook      | Description                                  |
| --------------------------------------------- | -------- | --------- | -------------------------------------------- |
| [commit-validator](./commit-validator/)       | Python   | pre-bump  | Validates conventional commit format         |
| [docker-tag-sync](./docker-tag-sync/)         | Bash     | post-bump | Tags and pushes Docker images                |
| [github-version-sync](./github-version-sync/) | Bash     | pre-bump  | Syncs version from GitHub repository release |

Each extension includes a complete `README.md` with usage examples and configuration options.

## Quick Install

Install directly from this repository using subdirectory URLs:

```bash
# Install commit-validator (Python)
sley extension install --url github.com/indaco/sley/contrib/extensions/commit-validator

# Install docker-tag-sync (Bash)
sley extension install --url github.com/indaco/sley/contrib/extensions/docker-tag-sync

# Install github-version-sync (Bash)
sley extension install --url github.com/indaco/sley/contrib/extensions/github-version-sync
```

Or from a local clone:

```bash
sley extension install --path ./contrib/extensions/commit-validator
sley extension install --path ./contrib/extensions/docker-tag-sync
sley extension install --path ./contrib/extensions/github-version-sync
```

## Management

```bash
# List installed extensions
sley extension list

# Remove an extension
sley extension remove --name commit-validator
```

## Documentation

For complete documentation, configuration examples, and troubleshooting:

- **[Extension System Guide](https://sley.indaco.dev/extensions/)** - Complete extension system documentation
- **[Creating Extensions](https://sley.indaco.dev/extensions/#creating-an-extension)** - Build your own extensions
- **[Commit Validator](https://sley.indaco.dev/extensions/commit-validator.html)** - Commit validation extension docs
- **[Docker Tag Sync](https://sley.indaco.dev/extensions/docker-tag-sync.html)** - Docker tagging extension docs
- **[GitHub Version Sync](https://sley.indaco.dev/extensions/github-version-sync.html)** - GitHub version sync extension docs
- **[Plugin System](https://sley.indaco.dev/plugins/)** - Built-in plugins vs extensions comparison

## Using These Examples

Each extension in this directory serves as a working example and template:

- `extension.yaml` - Extension manifest with metadata
- Implementation script - Executable hook script (Python, Bash, etc.)
- `README.md` - Extension-specific documentation and configuration

Browse the source code to learn extension development patterns, or use them as starting points for your own extensions.

## Extensions vs Plugins

**When to use extensions:**

- Custom organization-specific workflows
- External tool integration (AWS CLI, curl, etc.)
- Prototyping new features
- Language-specific tooling

**When to use plugins:**

- Core versioning functionality
- Performance-critical operations
- Features with broad applicability

See [Plugins vs Extensions](https://sley.indaco.dev/extensions/#when-to-use-extensions-vs-plugins) for detailed comparison.

## Contributing

Want to contribute an extension?

1. Follow the [Extension Authoring Guide](https://sley.indaco.dev/extensions/#creating-an-extension)
2. Include comprehensive documentation (README.md)
3. Add usage examples
4. Minimize external dependencies
5. Test thoroughly
6. Submit a pull request

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for general contribution guidelines.

## License

All extensions in this directory are licensed under the same terms as sley. See [LICENSE](../../LICENSE) for details.
