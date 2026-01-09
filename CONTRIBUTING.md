# Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request.

## Table of Contents

- [Reporting Issues](#reporting-issues)
- [Development Environment Setup](#development-environment-setup)
- [Setting Up Git Hooks](#setting-up-git-hooks)
- [Running Tasks](#running-tasks)
- [Code Style Guidelines](#code-style-guidelines)
- [Commit Message Format](#commit-message-format)
- [Submitting Pull Requests](#submitting-pull-requests)
- [Review Process](#review-process)
- [Getting Help](#getting-help)

## Reporting Issues

Before creating an issue:

- Search existing issues to avoid duplicates
- Check the [Troubleshooting](README.md#troubleshooting) section in the README

When reporting bugs, include:

- sley version (`sley --version`)
- Operating system and version
- Steps to reproduce the issue
- Expected vs. actual behavior
- Relevant configuration (`.sley.yaml`, `.version` file)
- Error messages or logs

For feature requests, describe the use case and problem you're trying to solve.

## Development Environment Setup

### Using Devbox (Recommended)

To set up a development environment for this repository, you can use [devbox](https://www.jetify.com/devbox) along with the provided `devbox.json` configuration file.

1. Install devbox by following [these instructions](https://www.jetify.com/devbox/docs/installing_devbox/).
2. Clone this repository to your local machine.

   ```bash
   git clone https://github.com/indaco/sley.git
   cd sley
   ```

3. Run `devbox install` to install all dependencies specified in `devbox.json`.
4. Enter the environment with `devbox shell --pure`.
5. Start developing, testing, and contributing!

### Manual Setup

If you prefer not to use Devbox, ensure you have the following tools installed:

- [Go](https://go.dev/dl/) (1.25 or later)
- [just](https://github.com/casey/just): Command runner for project tasks
- [golangci-lint](https://golangci-lint.run/): For linting Go code
- [modernize](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize): Run the modernizer analyzer to simplify code
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck): Reports known vulnerabilities that affect Go code

Optional tools:

- [goreportcard-cli](https://github.com/gojp/goreportcard): For local code quality reports (requires manual installation via git clone + make)
- [prek](https://github.com/j178/prek): For Git hooks management

## Setting Up Git Hooks

Git hooks are used to enforce code quality and streamline the workflow.

### Using Devbox

If using `devbox`, Git hooks are automatically installed when you run `devbox shell`. The hooks are managed by [prek](https://github.com/j178/prek).

### Manual Setup

For users not using `devbox`, first install [prek](https://prek.j178.dev/installation/) and then run the setup:

```bash
# Install prek (choose one)
brew install prek        # macOS/Linux with Homebrew
pip install prek         # via pip
cargo binstall prek      # via cargo

# Install git hooks
prek install -- --workspace
```

This installs the commit-msg hook that validates conventional commit messages.

## Running Tasks

This project uses [just](https://github.com/casey/just) for running tasks.

### View all available recipes

```bash
just help
```

### Available Recipes

| Recipe               | Description                                 |
| -------------------- | ------------------------------------------- |
| `just help`          | Print help message                          |
| `just all`           | Clean and build                             |
| `just clean`         | Clean the build directory and Go cache      |
| `just build`         | Build the binary with optimizations         |
| `just install`       | Install the binary using Go install         |
| `just test`          | Run all tests and print code coverage value |
| `just test-coverage` | Run all tests and generate coverage report  |
| `just test-force`    | Clean go tests cache and run all tests      |
| `just test-race`     | Run all tests with race detector            |
| `just lint`          | Run golangci-lint                           |
| `just modernize`     | Run go-modernize with auto-fix              |
| `just check`         | Run modernize, lint, and reportcard         |
| `just reportcard`    | Run goreportcard-cli                        |
| `just security-scan` | Run govulncheck                             |

## Code Style Guidelines

### Go Conventions

Follow standard Go conventions:

- **Error handling**: Use typed errors from `internal/apperrors/` (e.g., `apperrors.WrapGit()`, `apperrors.WrapFile()`) for structured error handling. Use `fmt.Errorf("context: %w", err)` for simpler cases.
- **Testability**: Use dependency injection interfaces from `internal/core/` (e.g., `FileSystem`, `CommandExecutor`, `GitClient`)
- **Context**: All async operations should accept `context.Context`
- **Documentation**: Add `doc.go` files for new internal packages

### Testing

- Write unit tests for new functionality
- Use mock implementations from `internal/core/` for unit testing
- Place test helpers in `internal/testutils` (excluded from coverage)
- Run tests: `just test`
- Check for race conditions: `just test-race`

### Before Submitting

Ensure your code passes all checks:

```bash
just check
```

This runs modernize, linting, and reportcard checks.

## Commit Message Format

This project follows [Conventional Commits](https://www.conventionalcommits.org/). The commit-msg hook validates your messages automatically.

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type       | Description                              |
| ---------- | ---------------------------------------- |
| `feat`     | New feature                              |
| `fix`      | Bug fix                                  |
| `docs`     | Documentation changes                    |
| `refactor` | Code refactoring without behavior change |
| `test`     | Adding or updating tests                 |
| `chore`    | Maintenance tasks, dependency updates    |
| `ci`       | CI/CD configuration changes              |
| `perf`     | Performance improvements                 |

### Examples

```bash
feat(changelog): add GitHub release format
fix(version): correct parsing of pre-release identifiers
docs(plugins): clarify tag-manager configuration
refactor(semver): simplify bump validation logic
test(workspace): add module discovery tests
```

### Breaking Changes

For breaking changes, add `!` after the type or include `BREAKING CHANGE:` in the footer:

```bash
feat(api)!: change plugin interface signature

BREAKING CHANGE: Plugin.Execute now requires context.Context as first parameter.
```

## Submitting Pull Requests

### Before You Start

1. Check if an issue exists for the feature/bug
2. Fork the repository and create a branch from `main`
3. Keep changes focused - one feature/fix per PR

### PR Process

1. Create a descriptive branch:

   ```bash
   git checkout -b feat/add-custom-format
   git checkout -b fix/version-parsing-bug
   ```

2. Make your changes following the code style guidelines

3. Write tests for your changes and ensure all tests pass:

   ```bash
   just test
   ```

4. Run quality checks:

   ```bash
   just check
   ```

5. Commit using conventional commits (see above)

6. Push and create a pull request:
   - Provide a clear description of what the PR does
   - Reference related issues (e.g., "Fixes #123")
   - Include examples for new features

### PR Requirements

All PRs must:

- Pass CI checks (tests, linting, security scan)
- Include tests for new functionality
- Update documentation if adding/changing features
- Follow conventional commit format

## Review Process

1. **Automated checks** run on all PRs (CI, tests, linting, govulncheck)
2. **Maintainer review** - feedback may include requests for changes or additional tests
3. **Approval and merge** after all checks pass and feedback is addressed

## Getting Help

- **Documentation**: Start with the [README](README.md) and [docs/](docs/) directory
- **Issues**: Search or create an issue for questions

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
