<h1 align="center">
  <code>verso</code>
</h1>
<h2 align="center" style="font-size: 1.5rem;">
  Version orchestrator for semantic versioning
</h2>

<p align="center">
  <a href="https://github.com/indaco/verso/actions/workflows/ci.yml" target="_blank">
    <img src="https://github.com/indaco/verso/actions/workflows/ci.yml/badge.svg" alt="CI" />
  </a>
  <a href="https://codecov.io/gh/indaco/verso">
    <img src="https://codecov.io/gh/indaco/verso/branch/main/graph/badge.svg" alt="Code coverage" />
  </a>
  <a href="https://goreportcard.com/report/github.com/indaco/verso" target="_blank">
    <img src="https://goreportcard.com/badge/github.com/indaco/verso" alt="Go Report Card" />
  </a>
  <a href="https://github.com/indaco/verso/releases/latest">
    <img src="https://img.shields.io/github/v/tag/indaco/verso?label=version&sort=semver&color=4c1" alt="version">
  </a>
  <a href="https://pkg.go.dev/github.com/indaco/verso" target="_blank">
    <img src="https://pkg.go.dev/badge/github.com/indaco/verso.svg" alt="Go Reference" />
  </a>
  <a href="https://github.com/indaco/verso/blob/main/LICENSE" target="_blank">
    <img src="https://img.shields.io/badge/license-mit-blue?style=flat-square" alt="License" />
  </a>
  <a href="https://www.jetify.com/devbox/docs/contributor-quickstart/" target="_blank">
    <img src="https://www.jetify.com/img/devbox/shield_moon.svg" alt="Built with Devbox" />
  </a>
</p>

---

A command-line tool for managing [SemVer 2.0.0](https://semver.org/) versions using a simple `.version` file. Works with any language or stack, integrates with CI/CD pipelines, and extends via built-in plugins for git tagging, changelog generation, and version validation.

---

## Table of Contents

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
- [License](#license)

## Features

- Lightweight `.version` file - SemVer 2.0.0 compliant
- `init`, `bump`, `set`, `show`, `validate` - intuitive version control
- Pre-release support with auto-increment (`alpha`, `beta.1`, `rc.2`, `--inc`)
- Built-in plugins - git tagging, changelog generation, version policy enforcement, commit parsing
- Extension system - hook external scripts into the version lifecycle
- Monorepo/multi-module support - manage multiple `.version` files at once
- Works standalone or in CI - `--strict` for strict mode
- Configurable via flags, env vars, or `.verso.yaml`

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

The `.version` file complements your existing tools. Pair it with `git tag` for releases, inject it into binaries at build time, or sync it across `package.json`, `Cargo.toml`, and other files using the `dependency-check` plugin.

## Installation

### Option 1: Install via `go install` (global)

```bash
go install github.com/indaco/verso/cmd/verso@latest
```

### Option 2: Install via `go install` (tool)

With Go 1.24 or greater installed, you can install `verso` locally in your project by running:

```bash
go get -tool github.com/indaco/verso/cmd/verso@latest
```

Once installed, use it with

```bash
go tool verso
```

### Option 3: Prebuilt binaries

Download the pre-compiled binaries from the [releases page](https://github.com/indaco/verso/releases) and place the binary in your system's PATH.

### Option 4: Clone and build manually

```bash
git clone https://github.com/indaco/verso.git
cd verso
just install
```

## CLI Commands & Options

```bash
NAME:
   verso - Version orchestrator for semantic versioning

USAGE:
   verso [global options] [command [command options]]

VERSION:
   v0.6.0-rc3

COMMANDS:
   show              Display current version
   set               Set the version manually
   bump              Bump semantic version (patch, minor, major)
   pre               Set pre-release label (e.g., alpha, beta.1)
   doctor, validate  Validate the .version file
   init              Initialize a .version file (auto-detects Git tag or starts from 0.1.0)
   extension         Manage extensions for verso
   modules, mods     Manage and discover modules in workspace
   help, h           Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --path string, -p string  Path to .version file (default: "internal/version/.version")
   --strict, --no-auto-init  Fail if .version file is missing (disable auto-initialization)
   --no-color                Disable colored output
   --help, -h                show help
   --version, -v             print the version
```

## Configuration

The CLI determines the `.version` path in the following order:

1. `--path` flag
2. `VERSO_PATH` environment variable
3. `.verso.yaml` file
4. Fallback: `.version` in the current directory

**Example: Use Environment Variable**

```bash
export VERSO_PATH=./my-folder/.version
verso patch
```

**Example: Use .verso.yaml**

```bash
# .verso.yaml
path: ./my-folder/.version
```

If both are missing, the CLI uses `.version` in the current directory.

## Auto-initialization

If the `.version` file does not exist when running the CLI:

1. It tries to read the latest Git tag via `git describe --tags`.
2. If the tag is a valid semantic version, it is used.
3. Otherwise, the file is initialized to 0.1.0.

This ensures your project always has a starting point.

Alternatively, run `verso init` explicitly:

```bash
verso init
# => Initialized .version with version 0.1.0
```

You can also specify a custom path:

```bash
verso init --path internal/version/.version
```

This behavior ensures your project always has a valid version starting point.

**To disable auto-initialization**, use the `--strict` flag.
This is useful in CI/CD environments or stricter workflows where you want the command to fail if the file is missing:

```bash
verso patch --strict
# => Error: .version file not found
```

## Usage

**Display current version**

```bash
# .version = 1.2.3
verso show
# => 1.2.3
```

```bash
# Fail if .version is missing (strict mode)
verso show --strict
# => Error: version file not found at .version
```

**Set version manually**

```bash
verso set 2.1.0
# => .version is now 2.1.0
```

You can also set a pre-release version:

```bash
verso set 2.1.0 --pre beta.1
# => .version is now 2.1.0-beta.1
```

You can also attach build metadata:

```bash
verso set 1.0.0 --meta ci.001
# => .version is now 1.0.0+ci.001
```

Or combine both:

```bash
verso set 1.0.0 --pre alpha --meta build.42
# => .version is now 1.0.0-alpha+build.42
```

**Bump version**

```bash
verso show
# => 1.2.3

verso bump patch
# => 1.2.4

verso bump minor
# => 1.3.0

verso bump major
# => 2.0.0

# .version = 1.3.0-alpha.1+build.123
verso bump release
# => 1.3.0
```

**Increment pre-release (`bump pre`)**

Increment only the pre-release portion without bumping the version number:

```bash
# .version = 1.0.0-rc.1
verso bump pre
# => 1.0.0-rc.2

# .version = 1.0.0-rc1
verso bump pre
# => 1.0.0-rc2

# Switch to a different pre-release label
# .version = 1.0.0-alpha.3
verso bump pre --label beta
# => 1.0.0-beta.1
```

You can also pass `--pre` and/or `--meta` flags to any bump:

```bash
verso bump patch --pre beta.1
# => 1.2.4-beta.1

verso bump minor --meta ci.123
# => 1.3.0+ci.123

verso bump major --pre rc.1 --meta build.7
# => 2.0.0-rc.1+build.7
```

> [!NOTE]
> By default, any existing build metadata (the part after `+`) is **cleared** when bumping the version.

To **preserve** existing metadata, pass the `--preserve-meta` flag:

```bash
# .version = 1.2.3+build.789
verso bump patch --preserve-meta
# => 1.2.4+build.789

# .version = 1.2.3+build.789
verso bump patch --meta new.build
# => 1.2.4+new.build (overrides existing metadata)
```

**Smart bump logic (`bump auto`)**

Automatically determine the next version:

```bash
# .version = 1.2.3-alpha.1
verso bump auto
# => 1.2.3

# .version = 1.2.3
verso bump auto
# => 1.2.4
```

Override bump with `--label`:

```bash
verso bump auto --label minor
# => 1.3.0

verso bump auto --label major --meta ci.9
# => 2.0.0+ci.9

verso bump auto --label patch --preserve-meta
# => bumps patch and keeps build metadata
```

Valid `--label` values: `patch`, `minor`, `major`.

**Manage pre-release versions**

```bash
# .version = 0.2.1
verso pre --label alpha
# => 0.2.2-alpha
```

If a pre-release is already present, it's replaced:

```bash
# .version = 0.2.2-beta.3
verso pre --label alpha
# => 0.2.2-alpha
```

**Auto-increment pre-release label**

```bash
# .version = 1.2.3
verso pre --label alpha --inc
# => 1.2.3-alpha.1
```

```bash
# .version = 1.2.3-alpha.1
verso pre --label alpha --inc
# => 1.2.3-alpha.2
```

**Validate .version file**

Check whether the `.version` file exists and contains a valid semantic version:

```bash
# .version = 1.2.3
verso validate
# => Valid version file at ./<path>/.version
```

If the file is missing or contains an invalid value, an error is returned:

```bash
# .version = invalid-content
verso validate
# => Error: invalid version format: ...
```

**Initialize .version file**

```bash
verso init
# => Initialized .version with version 0.1.0
```

## Plugin System

`verso` includes built-in plugins that provide deep integration with version bump logic. Unlike extensions (external scripts), plugins are compiled into the binary for native performance.

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
# .verso.yaml
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

`verso` supports extensions - external scripts that hook into the version lifecycle for automation tasks like updating changelogs, creating git tags, or enforcing version policies.

```bash
# Install an extension
verso extension install --path ./path/to/extension

# List installed extensions
verso extension list

# Remove an extension
verso extension remove my-extension
```

Ready-to-use extensions are available in [contrib/extensions/](contrib/extensions/).

For detailed documentation on hooks, JSON interface, and creating extensions, see [docs/EXTENSIONS.md](docs/EXTENSIONS.md).

## Monorepo / Multi-Module Support

`verso` supports managing multiple `.version` files across a monorepo. When multiple modules are detected, the CLI automatically enables multi-module mode.

```bash
# List discovered modules
verso modules list

# Show all module versions
verso show --all

# Bump all modules
verso bump patch --all

# Bump specific module
verso bump patch --module api

# Bump multiple modules
verso bump patch --modules api,web
```

For CI/CD, use `--non-interactive` or set `CI=true` to disable prompts.

For detailed documentation on module discovery, configuration, and patterns, see [docs/MONOREPO.md](docs/MONOREPO.md).

## Contributing

Contributions are welcome!

See the [Contributing Guide](/CONTRIBUTING.md) for setting up the development tools.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
