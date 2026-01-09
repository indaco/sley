# Changelog

All notable changes to this project will be documented in this file.

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). The changelog is generated and managed by [sley](https://github.com/indaco/sley).

## v0.5.0 - 2025-04-10

[compare changes](https://github.com/indaco/sley/compare/v0.4.0...v0.5.0)

### 🚀 Enhancements

- **cli:** Add `next` subcommand to `bump` ([46fa0d6](https://github.com/indaco/sley/commit/46fa0d6))
- **cli:** Support `--label` and `--meta` flags in `bump next` command ([7652165](https://github.com/indaco/sley/commit/7652165))

### 📖 Documentation

- **README:** Add smart bump logic for automatic versioning detection ([7025344](https://github.com/indaco/sley/commit/7025344))

### 🤖 CI

- Add golangci-lint configuration file ([#20](https://github.com/indaco/sley/pull/20))

### ❤️ Contributors

- Indaco ([@indaco](https://github.com/indaco))

## v0.4.0 - 2025-04-10

[compare changes](https://github.com/indaco/sley/compare/v0.3.0...v0.4.0)

### 🚀 Enhancements

- **set:** Support optional build metadata via `--meta` flag ([40215d5](https://github.com/indaco/sley/commit/40215d5))
- **bump:** Support optional `--pre` and `--meta` flags ([b1cb37b](https://github.com/indaco/sley/commit/b1cb37b))
- **bump:** Add `--preserve-meta` flag to preserve existing metadata ([a8ee225](https://github.com/indaco/sley/commit/a8ee225))
- **bump:** Add `release` command subcommand ([ea7609b](https://github.com/indaco/sley/commit/ea7609b))

### 💅 Refactors

- Restructure version bump commands under 'bump' subcommand ([4db6308](https://github.com/indaco/sley/commit/4db6308))

### 📖 Documentation

- **README:** Version badge with shields.io ([135e282](https://github.com/indaco/sley/commit/135e282))
- **README:** Fix link to the releases page ([cdf146c](https://github.com/indaco/sley/commit/cdf146c))
- **README:** Update Features and Why ([d815dbf](https://github.com/indaco/sley/commit/d815dbf))

### ✅ Tests

- **actions:** Refactor to table tests ([ca59a12](https://github.com/indaco/sley/commit/ca59a12))

### ❤️ Contributors

- Indaco ([@indaco](https://github.com/indaco))

## v0.3.0 - 2025-04-09

[compare changes](https://github.com/indaco/sley/compare/v0.2.0...v0.3.0)

### 🚀 Enhancements

- Add `set` command to manually set the version number ([9fefea8](https://github.com/indaco/sley/commit/9fefea8))
- Add `validate` command to validate the .version file ([5453f8f](https://github.com/indaco/sley/commit/5453f8f))
- Add `--no-auto-init` flag and reorganize CLI commands ([e11be92](https://github.com/indaco/sley/commit/e11be92))

### 🩹 Fixes

- Normalize file path to ensure correct file usage ([c0aa237](https://github.com/indaco/sley/commit/c0aa237))

### 📖 Documentation

- Update README enhance clarity and sections ([0615696](https://github.com/indaco/sley/commit/0615696))

### 🏡 Chore

- Fix lint errcheck ([3cd31f1](https://github.com/indaco/sley/commit/3cd31f1))

### ❤️ Contributors

- Indaco ([@indaco](https://github.com/indaco))

## v0.2.0 - 2025-04-09

[compare changes](https://github.com/indaco/sley/compare/v0.1.2...v0.2.0)

### 🚀 Enhancements

- Add `init` command to initialize version file with feedback ([13bfce7](https://github.com/indaco/sley/commit/13bfce7))

### 📖 Documentation

- **README:** Add `init` command ([fc14188](https://github.com/indaco/sley/commit/fc14188))

### 🏡 Chore

- Provide feedback when auto-initialize ([6cdc5cc](https://github.com/indaco/sley/commit/6cdc5cc))

### ❤️ Contributors

- Indaco ([@indaco](https://github.com/indaco))

## v0.1.2 - 2025-04-08

[compare changes](https://github.com/indaco/sley/compare/v0.1.1...v0.1.2)

### 🩹 Fixes

- Unintended version file created at the default path when --path flag ([9778a5e](https://github.com/indaco/sley/commit/9778a5e))

### 💅 Refactors

- Rename setupCLI to newCLI ([862cecb](https://github.com/indaco/sley/commit/862cecb))
- Simplify newCLI by removing unnecessary error return ([2df6b54](https://github.com/indaco/sley/commit/2df6b54))

### 📖 Documentation

- Update version badge in README ([69ef08d](https://github.com/indaco/sley/commit/69ef08d))

### 📦 Build

- Update install process to include modernize and lint ([314b827](https://github.com/indaco/sley/commit/314b827))
- **deps:** Bump github.com/urfave/cli/v3 from 3.0.0-beta1 to 3.1.1 ([ebb1b2b](https://github.com/indaco/sley/commit/ebb1b2b))

### 🏡 Chore

- Format go.mod ([bb0c6da](https://github.com/indaco/sley/commit/bb0c6da))

### ✅ Tests

- Move coverage report from coveralls to codecov ([00c94e3](https://github.com/indaco/sley/commit/00c94e3))

### 🤖 CI

- Update release name to version only ([c0bb30d](https://github.com/indaco/sley/commit/c0bb30d))
- Unify release and release notes workflows ([ce2523e](https://github.com/indaco/sley/commit/ce2523e))

### ❤️ Contributors

- Indaco ([@indaco](https://github.com/indaco))

## v0.1.1 - 2025-03-24

[compare changes](https://github.com/indaco/sley/compare/v0.1.0...v0.1.1)

### 🩹 Fixes

- Handle LoadConfig error in runCLI function ([9d75e16](https://github.com/indaco/sley/commit/9d75e16))

### 💅 Refactors

- Rename WriteVersion to SaveVersion and update usage ([71a3bda](https://github.com/indaco/sley/commit/71a3bda))

### 📖 Documentation

- **README:** Update headline ([f5ba5cc](https://github.com/indaco/sley/commit/f5ba5cc))

### 🤖 CI

- Remove release note parsing step in favor of softprops body_path ([21b71e4](https://github.com/indaco/sley/commit/21b71e4))

### ❤️ Contributors

- Indaco ([@indaco](https://github.com/indaco))

## v0.1.0 - 2025-03-24

### 🏡 Chore

- Initial release

### ❤️ Contributors

- Indaco ([@indaco](https://github.com/indaco))
