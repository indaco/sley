<h1 align="center">
  <img src="assets/logo.svg" alt="sley logo" width="120" height="120">
  <br>
  <code>sley</code>
</h1>
<h2 align="center" style="font-size: 1.5rem;">
  A semantic version orchestrator for your projects.
</h2>

<p align="center">
  <a href="https://github.com/indaco/sley/actions/workflows/ci.yml" target="_blank">
    <img src="https://github.com/indaco/sley/actions/workflows/ci.yml/badge.svg" alt="CI" />
  </a>
  <a href="https://codecov.io/gh/indaco/sley" target="_blank">
    <img src="https://codecov.io/gh/indaco/sley/branch/main/graph/badge.svg" alt="Code coverage" />
  </a>
  <a href="https://goreportcard.com/report/github.com/indaco/sley" target="_blank">
    <img src="https://goreportcard.com/badge/github.com/indaco/sley" alt="Go Report Card" />
  </a>
  <a href="https://github.com/indaco/sley/security" target="_blank">
    <img src="https://img.shields.io/badge/security-govulncheck-green" alt="Security Scan" />
  </a>
  <a href="https://github.com/indaco/sley/releases" target="_blank">
    <img src="https://img.shields.io/github/v/tag/indaco/sley?label=version&sort=semver&color=4c1" alt="version">
  </a>
  <a href="https://pkg.go.dev/github.com/indaco/sley" target="_blank">
    <img src="https://pkg.go.dev/badge/github.com/indaco/sley.svg" alt="Go Reference" />
  </a>
  <a href="LICENSE" target="_blank">
    <img src="https://img.shields.io/badge/license-mit-blue?style=flat-square" alt="License" />
  </a>
  <a href="https://www.jetify.com/devbox" target="_blank">
    <img src="https://www.jetify.com/img/devbox/shield_moon.svg" alt="Built with Devbox" />
  </a>
</p>

<p align="center">
  <b><a href="https://sley.indaco.dev">Documentation</a></b> |
  <b><a href="https://sley.indaco.dev/guide/quick-start.html">Quick Start</a></b> |
  <b><a href="https://sley.indaco.dev/reference/cli.html">CLI Reference</a></b>
</p>

## Overview

sley manages [SemVer 2.0.0](https://semver.org/) versions using a simple `.version` file. It's language-agnostic, works with any stack, and integrates with Git for tagging and changelog management.

```
.version → 1.2.3
```

## Features

- **Simple** — One `.version` file, one source of truth
- **Language-agnostic** — Works with Go, Node, Python, Rust, or any stack
- **Plugin system** — Extend with built-in or custom plugins
- **Git integration** — Auto-tag releases, generate changelogs
- **Monorepo support** — Manage multiple modules independently
- **CI/CD ready** — Designed for automation pipelines

## Installation

### Homebrew (macOS/Linux)

```bash
brew install indaco/tap/sley
```

### asdf

```bash
asdf plugin add sley https://github.com/indaco/asdf-sley.git
asdf install sley latest
asdf set --home sley latest
```

### Go

```bash
go install github.com/indaco/sley/cmd/sley@latest
```

### Pre-built binaries

Download from [Releases](https://github.com/indaco/sley/releases).

## Quick Start

```bash
# Initialize in your project
sley init

# Bump the version
sley bump patch    # 1.0.0 → 1.0.1
sley bump minor    # 1.0.1 → 1.1.0
sley bump major    # 1.1.0 → 2.0.0

# Auto-detect from conventional commits
sley bump auto

# Show current version
sley show
```

## Plugins

sley includes built-in plugins for common workflows:

| Plugin                                                                          | Purpose                                   |
| ------------------------------------------------------------------------------- | ----------------------------------------- |
| [commit-parser](https://sley.indaco.dev/plugins/commit-parser.html)             | Infer bump type from conventional commits |
| [tag-manager](https://sley.indaco.dev/plugins/tag-manager.html)                 | Auto-create Git tags                      |
| [changelog-generator](https://sley.indaco.dev/plugins/changelog-generator.html) | Generate changelog from commits           |
| [version-validator](https://sley.indaco.dev/plugins/version-validator.html)     | Enforce versioning policies               |
| [dependency-check](https://sley.indaco.dev/plugins/dependency-check.html)       | Sync versions across files                |
| [release-gate](https://sley.indaco.dev/plugins/release-gate.html)               | Pre-bump validation checks                |

See all plugins in the [documentation](https://sley.indaco.dev/plugins/).

## Configuration

Configure sley with a `.sley.yaml` file:

```yaml
plugins:
  commit-parser: true
  tag-manager:
    enabled: true
    prefix: "v"
    sign: true
  changelog-generator:
    enabled: true
```

See the [configuration reference](https://sley.indaco.dev/reference/sley-yaml-configuration.html) for all options.

## Documentation

Full documentation is available at **[sley.indaco.dev](https://sley.indaco.dev)**:

- [What is sley?](https://sley.indaco.dev/guide/what-is-sley.html)
- [Installation](https://sley.indaco.dev/guide/installation.html)
- [Usage Guide](https://sley.indaco.dev/guide/usage.html)
- [Plugins](https://sley.indaco.dev/plugins/)
- [Extensions](https://sley.indaco.dev/extensions/)
- [CI/CD Integration](https://sley.indaco.dev/guide/ci-cd.html)

## Contributing

Contributions are welcome. See [CONTRIBUTING](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
