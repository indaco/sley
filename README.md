<h1 align="center">
  <img src="assets/logo.svg" alt="sley logo" width="120" height="120">
  <br>
  <code>sley</code>
</h1>
<h2 align="center" style="font-size: 1.5rem;">
  Version orchestrator for semantic versioning
</h2>

<p align="center">
  <a href="https://github.com/indaco/sley/actions/workflows/ci.yml" target="_blank">
    <img src="https://github.com/indaco/sley/actions/workflows/ci.yml/badge.svg" alt="CI" />
  </a>
  <a href="https://codecov.io/gh/indaco/sley">
    <img src="https://codecov.io/gh/indaco/sley/branch/main/graph/badge.svg" alt="Code coverage" />
  </a>
  <a href="https://goreportcard.com/report/github.com/indaco/sley" target="_blank">
    <img src="https://goreportcard.com/badge/github.com/indaco/sley" alt="Go Report Card" />
  </a>
  <a href="https://github.com/indaco/sley/actions/workflows/ci.yml" target="_blank">
    <img src="https://img.shields.io/badge/security-govulncheck-green" alt="Security Scan" />
  </a>
  <a href="https://github.com/indaco/sley/releases/latest">
    <img src="https://img.shields.io/github/v/tag/indaco/sley?label=version&sort=semver&color=4c1" alt="version">
  </a>
  <a href="https://pkg.go.dev/github.com/indaco/sley" target="_blank">
    <img src="https://pkg.go.dev/badge/github.com/indaco/sley.svg" alt="Go Reference" />
  </a>
  <a href="https://github.com/indaco/sley/blob/main/LICENSE" target="_blank">
    <img src="https://img.shields.io/badge/license-mit-blue?style=flat-square" alt="License" />
  </a>
  <a href="https://www.jetify.com/devbox/docs/contributor-quickstart/" target="_blank">
    <img src="https://www.jetify.com/img/devbox/shield_moon.svg" alt="Built with Devbox" />
  </a>
</p>

A command-line tool for managing [SemVer 2.0.0](https://semver.org/) versions using a simple `.version` file. Works with any language or stack, integrates with CI/CD pipelines, and extends via built-in plugins for git tagging, changelog generation, and version validation.

> _sley - named for the weaving tool that arranges threads in precise order._

## Quick Start

```bash
# Initialize version file
sley init

# Show current version
sley show

# Bump patch version (1.2.3 -> 1.2.4)
sley bump patch
```

## Table of Contents

- [Quick Start](#quick-start)
- [Features](#features)
- [Why .version?](#why-version)
- [Installation](#installation)
- [CLI Commands & Options](#cli-commands--options)
- [Configuration](#configuration)
- [Auto-initialization](#auto-initialization)
- [Usage](#usage)
- [Plugin System](#plugin-system)
- [Extension System](#extension-system)
- [Monorepo / Multi-Module Support](#monorepo--multi-module-support)
- [Contributing](#contributing)
- [AI Assistance](#ai-assistance)
- [License](#license)

## Features

- Lightweight `.version` file - SemVer 2.0.0 compliant
- `init`, `bump`, `set`, `show`, `validate` - intuitive version control
- Pre-release support with auto-increment (`alpha`, `beta.1`, `rc.2`, `--inc`)
- Built-in plugins - git tagging, changelog generation, version policy enforcement, commit parsing
- Extension system - hook external scripts into the version lifecycle
- Monorepo/multi-module support - manage multiple `.version` files at once
- Works standalone or in CI - `--strict` for strict mode
- Configurable via flags, env vars, or `.sley.yaml`

## Why .version?

Most projects - especially CLIs, scripts, and internal tools - need a clean way to manage versioning outside of `go.mod` or `package.json`.

### What it is

- A **single source of truth** for your project version
- **Language-agnostic** - works with Go, Python, Node, Rust, or any stack
- **CI/CD friendly** - inject into Docker labels, GitHub Actions, release scripts
- **Human-readable** - just a plain text file containing `1.2.3`
- **Predictable** - no magic, no hidden state, version is what you set

### What it is NOT

- **Not a replacement for git tags** - use the `tag-manager` plugin to sync both
- **Not a package manager** - it doesn't publish or distribute anything
- **Not a changelog tool** - use the `changelog-generator` plugin for that
- **Not a build system** - it just manages the version string

The `.version` file complements your existing tools. Pair it with `git tag` for releases, inject it into binaries at build time, or sync it across `package.json`, `Cargo.toml`, and other files using the [`dependency-check` plugin](#plugin-system).

## Installation

### Option 1: Homebrew (macOS/Linux)

```bash
brew install indaco/tap/sley
```

### Option 2: Install via `go install` (global)

```bash
go install github.com/indaco/sley/cmd/sley@latest
```

### Option 3: Install via `go install` (tool)

With Go 1.24 or greater installed, you can install `sley` locally in your project by running:

```bash
go get -tool github.com/indaco/sley/cmd/sley@latest
```

Once installed, use it with

```bash
go tool sley
```

### Option 4: Prebuilt binaries

Download the pre-compiled binaries from the [releases page](https://github.com/indaco/sley/releases) and place the binary in your system's PATH.

### Option 5: Clone and build manually

```bash
git clone https://github.com/indaco/sley.git
cd sley
just install
```

## CLI Commands & Options

```bash
NAME:
   sley - Version orchestrator for semantic versioning

USAGE:
   sley [global options] [command [command options]]

VERSION:
   v0.6.0-rc5

COMMANDS:
   show              Display current version
   set               Set the version manually
   bump              Bump semantic version (patch, minor, major)
   changelog         Manage changelog files
   pre               Set pre-release label (e.g., alpha, beta.1)
   doctor, validate  Validate .version file(s) and configuration
   init              Initialize .version file and .sley.yaml configuration
   extension         Manage extensions for sley
   modules, mods     Manage and discover modules in workspace
   help, h           Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --path string, -p string  Path to .version file (default: ".version")
   --strict, --no-auto-init  Fail if .version file is missing (disable auto-initialization)
   --no-color                Disable colored output
   --help, -h                show help
   --version, -v             print the version
```

## Configuration

The CLI determines the `.version` path in the following order:

1. `--path` flag
2. `SLEY_PATH` environment variable
3. `.sley.yaml` file
4. Fallback: `.version` in the current directory

**Example: Use Environment Variable**

```bash
export SLEY_PATH=./my-folder/.version
sley patch
```

**Example: Use .sley.yaml**

```bash
# .sley.yaml
path: ./my-folder/.version
```

If both are missing, the CLI uses `.version` in the current directory.

## Auto-initialization

If the `.version` file does not exist when running the CLI:

1. It tries to read the latest Git tag via `git describe --tags`.
2. If the tag is a valid semantic version, it is used.
3. Otherwise, the file is initialized to 0.1.0.

This ensures your project always has a starting point.

### Using `sley init`

The recommended way to initialize a new project is with `sley init`:

```bash
# Interactive mode - select plugins and generate .sley.yaml
sley init

# Non-interactive with sensible defaults
sley init --yes

# Use a pre-configured template
sley init --template automation

# Enable specific plugins
sley init --enable commit-parser,tag-manager,changelog-generator

# Initialize as monorepo with workspace configuration
sley init --workspace --yes

# Migrate version from existing package.json, Cargo.toml, etc.
sley init --migrate --yes

# Custom path
sley init --path internal/version/.version
```

**Available flags:**

| Flag           | Description                                               |
| -------------- | --------------------------------------------------------- |
| `--yes`, `-y`  | Use defaults without prompts (commit-parser, tag-manager) |
| `--template`   | Use a pre-configured template (see below)                 |
| `--enable`     | Comma-separated list of plugins to enable                 |
| `--workspace`  | Initialize as monorepo with workspace configuration       |
| `--migrate`    | Detect version from existing files (package.json, etc.)   |
| `--force`      | Overwrite existing .sley.yaml                             |
| `--path`, `-p` | Custom path for .version file                             |

**Available templates:**

| Template     | Plugins Enabled                                             |
| ------------ | ----------------------------------------------------------- |
| `basic`      | commit-parser                                               |
| `git`        | commit-parser, tag-manager                                  |
| `automation` | commit-parser, tag-manager, changelog-generator             |
| `strict`     | commit-parser, tag-manager, version-validator, release-gate |
| `full`       | All plugins enabled                                         |

**To disable auto-initialization**, use the `--strict` flag.
This is useful in CI/CD environments or stricter workflows where you want the command to fail if the file is missing:

```bash
sley bump patch --strict
# => Error: .version file not found
```

## Usage

**Display current version**

```bash
# .version = 1.2.3
sley show
# => 1.2.3
```

```bash
# Fail if .version is missing (strict mode)
sley show --strict
# => Error: version file not found at .version
```

**Set version manually**

```bash
sley set 2.1.0
# => .version is now 2.1.0
```

You can also set a pre-release version:

```bash
sley set 2.1.0 --pre beta.1
# => .version is now 2.1.0-beta.1
```

You can also attach build metadata:

```bash
sley set 1.0.0 --meta ci.001
# => .version is now 1.0.0+ci.001
```

Or combine both:

```bash
sley set 1.0.0 --pre alpha --meta build.42
# => .version is now 1.0.0-alpha+build.42
```

**Bump version**

```bash
sley show
# => 1.2.3

sley bump patch
# => 1.2.4

sley bump minor
# => 1.3.0

sley bump major
# => 2.0.0

# .version = 1.3.0-alpha.1+build.123
sley bump release
# => 1.3.0
```

**Increment pre-release (`bump pre`)**

Increment only the pre-release portion without bumping the version number:

```bash
# .version = 1.0.0-rc.1
sley bump pre
# => 1.0.0-rc.2

# .version = 1.0.0-rc1
sley bump pre
# => 1.0.0-rc2

# Switch to a different pre-release label
# .version = 1.0.0-alpha.3
sley bump pre --label beta
# => 1.0.0-beta.1
```

You can also pass `--pre` and/or `--meta` flags to any bump:

```bash
sley bump patch --pre beta.1
# => 1.2.4-beta.1

sley bump minor --meta ci.123
# => 1.3.0+ci.123

sley bump major --pre rc.1 --meta build.7
# => 2.0.0-rc.1+build.7
```

> [!NOTE]
> By default, any existing build metadata (the part after `+`) is **cleared** when bumping the version.

To **preserve** existing metadata, pass the `--preserve-meta` flag:

```bash
# .version = 1.2.3+build.789
sley bump patch --preserve-meta
# => 1.2.4+build.789

# .version = 1.2.3+build.789
sley bump patch --meta new.build
# => 1.2.4+new.build (overrides existing metadata)
```

**Smart bump logic (`bump auto`)**

Automatically determine the next version:

```bash
# .version = 1.2.3-alpha.1
sley bump auto
# => 1.2.3

# .version = 1.2.3
sley bump auto
# => 1.2.4
```

Override bump with `--label`:

```bash
sley bump auto --label minor
# => 1.3.0

sley bump auto --label major --meta ci.9
# => 2.0.0+ci.9

sley bump auto --label patch --preserve-meta
# => bumps patch and keeps build metadata
```

Valid `--label` values: `patch`, `minor`, `major`.

**Manage pre-release versions**

```bash
# .version = 0.2.1
sley pre --label alpha
# => 0.2.2-alpha
```

If a pre-release is already present, it's replaced:

```bash
# .version = 0.2.2-beta.3
sley pre --label alpha
# => 0.2.2-alpha
```

**Auto-increment pre-release label**

```bash
# .version = 1.2.3
sley pre --label alpha --inc
# => 1.2.3-alpha.1
```

```bash
# .version = 1.2.3-alpha.1
sley pre --label alpha --inc
# => 1.2.3-alpha.2
```

**Validate .version file**

Check whether the `.version` file exists and contains a valid semantic version:

```bash
# .version = 1.2.3
sley validate
# => Valid version file at ./<path>/.version
```

If the file is missing or contains an invalid value, an error is returned:

```bash
# .version = invalid-content
sley validate
# => Error: invalid version format: ...
```

**Initialize .version file**

```bash
sley init
# => Interactive mode: select plugins, create .sley.yaml
```

Use `--yes` for non-interactive initialization with defaults:

```bash
sley init --yes
# => Created .version with version 0.1.0
# => Created .sley.yaml with default plugins (commit-parser, tag-manager)
```

Enable specific plugins:

```bash
sley init --enable commit-parser,changelog-generator,audit-log
# => Created .sley.yaml with 3 plugins enabled
```

Force overwrite existing configuration:

```bash
sley init --yes --force
# => Overwrites existing .sley.yaml
```

Migrate version from existing project files:

```bash
sley init --migrate --yes
# => Detected 2.0.0 from package.json, uses it for .version
```

### Interactive Mode

When running `sley init` without flags in an interactive terminal, you'll see:

```
Initializing sley...

Detected:
  - Git repository
  - package.json (Node.js project)

Select plugins to enable:
  [x] Commit Parser - Analyze conventional commits to determine bump type
  [x] Tag Manager - Auto-create git tags after version bumps
  [ ] Version Validator - Enforce versioning policies
  [ ] Dependency Check - Sync version to package.json and other files
  [ ] Changelog Parser - Infer bump type from CHANGELOG.md
  [ ] Changelog Generator - Generate changelogs from commits
  [ ] Release Gate - Pre-bump validation (clean worktree, CI status)
  [ ] Audit Log - Record version history with metadata

Created .version with version 0.1.0
Created .sley.yaml with 2 plugins enabled

Next steps:
  - Review .sley.yaml and adjust settings
  - Run 'sley bump patch' to increment version
  - Run 'sley doctor' to verify setup
```

The init command automatically detects your project type (Git, Node.js, Go, Rust, Python) and suggests relevant plugins.

## Plugin System

`sley` includes built-in plugins that provide deep integration with version bump logic. Unlike extensions (external scripts), plugins are compiled into the binary for native performance.

### Available Plugins

| Plugin                | Description                                            | Default  |
| --------------------- | ------------------------------------------------------ | -------- |
| `commit-parser`       | Analyzes conventional commits to determine bump type   | Enabled  |
| `tag-manager`         | Automatically creates git tags synchronized with bumps | Disabled |
| `version-validator`   | Enforces versioning policies and constraints           | Disabled |
| `dependency-check`    | Validates and syncs versions across multiple files     | Disabled |
| `changelog-parser`    | Infers bump type from CHANGELOG.md entries             | Disabled |
| `changelog-generator` | Generates changelog from conventional commits          | Disabled |
| `release-gate`        | Pre-bump validation (clean worktree, branch, WIP)      | Disabled |
| `audit-log`           | Records version changes with metadata to a log file    | Disabled |

### Quick Example

```yaml
# .sley.yaml
plugins:
  commit-parser: true
  tag-manager:
    enabled: true
    prefix: "v"
    annotate: true
    push: false
  version-validator:
    enabled: true
    rules:
      - type: "major-version-max"
        value: 10
      - type: "branch-constraint"
        branch: "release/*"
        allowed: ["patch"]
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: "package.json"
        field: "version"
        format: "json"
  changelog-generator:
    enabled: true
    mode: "versioned"
    format: "grouped" # or "keepachangelog" for Keep a Changelog spec
    repository:
      auto-detect: true
```

For detailed documentation on all plugins and their configuration, see [docs/PLUGINS.md](docs/PLUGINS.md).

## Extension System

`sley` supports extensions - external scripts that hook into the version lifecycle for automation tasks like updating changelogs, creating git tags, or enforcing version policies.

```bash
# Install an extension
sley extension install --path ./path/to/extension

# List installed extensions
sley extension list

# Remove an extension
sley extension remove my-extension
```

Ready-to-use extensions are available in [contrib/extensions/](contrib/extensions/).

For detailed documentation on hooks, JSON interface, and creating extensions, see [docs/EXTENSIONS.md](docs/EXTENSIONS.md).

## Monorepo / Multi-Module Support

`sley` supports managing multiple `.version` files across a monorepo. When multiple modules are detected, the CLI automatically enables multi-module mode.

```bash
# List discovered modules
sley modules list

# Show all module versions
sley show --all

# Bump all modules
sley bump patch --all

# Bump specific module
sley bump patch --module api

# Bump multiple modules
sley bump patch --modules api,web
```

For CI/CD, use `--non-interactive` or set `CI=true` to disable prompts.

For detailed documentation on module discovery, configuration, and patterns, see [docs/MONOREPO.md](docs/MONOREPO.md).

## Contributing

Contributions are welcome!

See the [Contributing Guide](/CONTRIBUTING.md) for setting up the development tools.

## AI Assistance

Built by humans, with some help from AI. Starting from v0.5.0, [Claude Code](https://claude.ai/code) assisted with test generation, documentation scaffolding, code review, and tedious refactoring (like renaming from "semver" to "sley"). The project logo was also AI-generated under the maintainer's direction. All output was reviewed, reworked and approved by the maintainer.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
