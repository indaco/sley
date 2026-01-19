# Changelog

All notable changes to this project will be documented in this file.

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html). The changelog is generated and managed by [sley](https://github.com/indaco/sley).

## v0.8.0 - 2026-01-19

### 🚀 Enhancements

- **changelog-parser:** support multiple changelog formats ([b682ab8](https://github.com/indaco/sley/commit/b682ab8)) ([#154](https://github.com/indaco/sley/pull/154))

### 💅 Refactors

- **tui:** consolidate TUI utilities and prompt functions ([7dfdf16](https://github.com/indaco/sley/commit/7dfdf16)) ([#151](https://github.com/indaco/sley/pull/151))

### 📖 Documentation

- update README flow ([548fbb9](https://github.com/indaco/sley/commit/548fbb9)) ([#155](https://github.com/indaco/sley/pull/155))

### 🤖 CI

- ci/update-goreleaser-config-and-workflow (#152) ([a58129d](https://github.com/indaco/sley/commit/a58129d))

**Full Changelog:** [v0.7.0...v0.8.0](https://github.com/indaco/sley/compare/v0.7.0...v0.8.0)

### ❤️ Contributors

- [@indaco](https://github.com/indaco)

## v0.7.0 - 2026-01-13

### 🚀 Enhancements

- **cli:** add tag command with create/list/push/delete subcommands and ([154cd25](https://github.com/indaco/sley/commit/154cd25))
- **tagmanager:** change auto-create and tag-prereleases defaults to false ([73ad90a](https://github.com/indaco/sley/commit/73ad90a))
- **tagmanager:** implement ListTags, DeleteTag, DeleteRemoteTag on OSGitTagOperations ([5009d15](https://github.com/indaco/sley/commit/5009d15))
- **core:** extend GitTagOperations interface with list and delete methods ([cc01ccd](https://github.com/indaco/sley/commit/cc01ccd))
- **changelog:** add merge-after config and improve CLI feedback ([256c0ec](https://github.com/indaco/sley/commit/256c0ec)) ([#132](https://github.com/indaco/sley/pull/132))
- **tagmanager:** add option to control git tags creation for pre-release versions ([9c170ab](https://github.com/indaco/sley/commit/9c170ab)) ([#127](https://github.com/indaco/sley/pull/127))
- **changelog:** add option to set custom icon for breaking changes ([fa2e994](https://github.com/indaco/sley/commit/fa2e994)) ([#125](https://github.com/indaco/sley/pull/125))
- **changelog:** add Minimal release format ([4899b71](https://github.com/indaco/sley/commit/4899b71)) ([#122](https://github.com/indaco/sley/pull/122))
- **changelog:** add GitHub release format with breaking changes section ([7fe4a3c](https://github.com/indaco/sley/commit/7fe4a3c)) ([#121](https://github.com/indaco/sley/pull/121))

### 🩹 Fixes

- **bumpcmd:** add missing extension hooks to auto and release commands ([d2681c5](https://github.com/indaco/sley/commit/d2681c5)) ([#138](https://github.com/indaco/sley/pull/138))
- **extensionmgr:** respect --extension-dir when set to "." ([ec9f4ed](https://github.com/indaco/sley/commit/ec9f4ed)) ([#136](https://github.com/indaco/sley/pull/136))
- **bumpcmd:** add missing tag creation for bump auto command ([c3c9974](https://github.com/indaco/sley/commit/c3c9974)) ([#134](https://github.com/indaco/sley/pull/134))

### 💅 Refactors

- **releasegate:** remove function variables and use interfaces ([258926b](https://github.com/indaco/sley/commit/258926b)) ([#147](https://github.com/indaco/sley/pull/147))
- **versionvalidator:** add max-prerelease-iterations and require-even-minor rules ([298c1fe](https://github.com/indaco/sley/commit/298c1fe)) ([#145](https://github.com/indaco/sley/pull/145))
- **tagmanager:** add GPG signing and message templates, remove git-tagger extension ([d620e73](https://github.com/indaco/sley/commit/d620e73))

### 📖 Documentation

- update tag manager documentation and README ([07834b1](https://github.com/indaco/sley/commit/07834b1))
- update extensions documentation ([0fd5f41](https://github.com/indaco/sley/commit/0fd5f41)) ([#146](https://github.com/indaco/sley/pull/146))
- **plugins:** normalize plugin type fields and remove Version row ([5398d15](https://github.com/indaco/sley/commit/5398d15)) ([#131](https://github.com/indaco/sley/pull/131))
- **plugins:** consistent and minimal example files ([328531b](https://github.com/indaco/sley/commit/328531b)) ([#129](https://github.com/indaco/sley/pull/129))
- **plugin:** add missing config options for plugins and example files ([59ff8b6](https://github.com/indaco/sley/commit/59ff8b6)) ([#126](https://github.com/indaco/sley/pull/126))
- clarify how to get PR links in the changelog ([3c4c6fd](https://github.com/indaco/sley/commit/3c4c6fd)) ([#120](https://github.com/indaco/sley/pull/120))
- rephrased intro for changelog generator ([8c26357](https://github.com/indaco/sley/commit/8c26357))
- improve documentation consistency ([ab60933](https://github.com/indaco/sley/commit/ab60933))

### 🎨 Styling

- normalize dashes and spacing in devbox-init.sh script ([e930a2e](https://github.com/indaco/sley/commit/e930a2e))

### ✅ Tests

- **validator:** extracted validVersionValidatorRuleTypes to package-level map for tests semplification ([9b3d28c](https://github.com/indaco/sley/commit/9b3d28c))

### 🏡 Chores

- update sley config file ([690e145](https://github.com/indaco/sley/commit/690e145))
- consolidate package-sync extension into dependencycheck plugin ([ca7c23b](https://github.com/indaco/sley/commit/ca7c23b)) ([#144](https://github.com/indaco/sley/pull/144))
- update sley config file and gitignore ([080a306](https://github.com/indaco/sley/commit/080a306)) ([#139](https://github.com/indaco/sley/pull/139))
- **config:** disable automatic tag-manager plugin in sley config ([f363b84](https://github.com/indaco/sley/commit/f363b84))

### 🤖 CI

- **release:** migrate from brews to homebrew_casks in goreleaser config ([8597757](https://github.com/indaco/sley/commit/8597757))

**Full Changelog:** [v0.6.0...v0.7.0](https://github.com/indaco/sley/compare/v0.6.0...v0.7.0)

### ❤️ Contributors

- [@indaco](https://github.com/indaco)

## v0.6.0 - 2026-01-09

### 🚀 Enhancements

- **clix:** implement modules and pattern flags for monorepo ([82d84be](https://github.com/indaco/sley/commit/82d84be))
- **doctor:** add plugin status section ([eaf1c56](https://github.com/indaco/sley/commit/eaf1c56))
- **plugins:** add plugin metadata types and registry introspection ([185c202](https://github.com/indaco/sley/commit/185c202))
- **git:** add explicit timeouts for external operations ([56486f2](https://github.com/indaco/sley/commit/56486f2))
- **hooks:** add context parameter for cancellation and timeout support ([85ce94b](https://github.com/indaco/sley/commit/85ce94b))
- **changelog-generator:** new contributors section and configuration ([e97e0b7](https://github.com/indaco/sley/commit/e97e0b7))
- **changelog-generator:** use GitHub-style full changelog link format ([92570bd](https://github.com/indaco/sley/commit/92570bd))
- **changelog-generator:** add use-default-icons config option ([3f47382](https://github.com/indaco/sley/commit/3f47382))
- **pre:** add multi-module workspace support ([35751be](https://github.com/indaco/sley/commit/35751be))
- **doctor:** add configuration file validation ([2ba2691](https://github.com/indaco/sley/commit/2ba2691))
- **printer:** add styled output for validation results ([b90d2bd](https://github.com/indaco/sley/commit/b90d2bd))
- **clix:** add WithDefaultAll option for read-only commands ([5208be6](https://github.com/indaco/sley/commit/5208be6))
- **initcmd:** add --migrate flag to detect version from package.json, Cargo.toml, pyproject.toml, ecc ([07e8534](https://github.com/indaco/sley/commit/07e8534))
- **init:** add --template flag for pre-configured plugin sets ([2be8a55](https://github.com/indaco/sley/commit/2be8a55))
- **initcmd:** add --workspace flag for monorepo setup ([1faa7d7](https://github.com/indaco/sley/commit/1faa7d7))
- **init:** add --template flag for pre-configured plugin sets ([b62d992](https://github.com/indaco/sley/commit/b62d992))
- **initcmd:** add --yes, --enable, --force flags ([5e39594](https://github.com/indaco/sley/commit/5e39594))
- **initcmd:** generate .sley.yaml with inline comments ([eb11904](https://github.com/indaco/sley/commit/eb11904))
- **initcmd:** add interactive plugin selection ([182e678](https://github.com/indaco/sley/commit/182e678))
- **initcmd:** add project context detection ([da444d8](https://github.com/indaco/sley/commit/da444d8))
- **operations:** add ValidateOperation for multi-module support ([2e8d38e](https://github.com/indaco/sley/commit/2e8d38e))
- **workspace:** add DisplayNameWithPath for module disambiguation ([9951156](https://github.com/indaco/sley/commit/9951156))
- **package-sync:** add monorepo support ([8db41b4](https://github.com/indaco/sley/commit/8db41b4))
- **bumpcmd:** pass module context to extension hooks ([ee5bf1b](https://github.com/indaco/sley/commit/ee5bf1b))
- **extensionmgr:** add module context to extension hooks ([46f75d3](https://github.com/indaco/sley/commit/46f75d3))
- **modules:** show path in list output for disambiguation ([30d7a9b](https://github.com/indaco/sley/commit/30d7a9b))
- **printer:** add printer package for styled console output ([6f9e782](https://github.com/indaco/sley/commit/6f9e782))
- **changelog:** add customizable contributor format template ([7f4d1d2](https://github.com/indaco/sley/commit/7f4d1d2))
- **changelog:** add merge command for versioned changelog files ([26b7c44](https://github.com/indaco/sley/commit/26b7c44))
- **changelog:** add icon support for contributors section ([a3ee4bb](https://github.com/indaco/sley/commit/a3ee4bb))
- **changelog-generator:** add format configuration with keepachangelog support ([616b46b](https://github.com/indaco/sley/commit/616b46b))
- **extensions:** add docker-tag-sync extension for Docker image tagging ([c00266e](https://github.com/indaco/sley/commit/c00266e))
- **bump:** integrate audit-log recording in all bump commands ([c7f45bd](https://github.com/indaco/sley/commit/c7f45bd))
- **plugins:** add audit-log plugin for version history tracking ([1f1eecc](https://github.com/indaco/sley/commit/1f1eecc))
- **bump:** integrate release-gate validation in all bump commands ([d97ed3c](https://github.com/indaco/sley/commit/d97ed3c))
- **plugins:** add release-gate plugin for pre-bump validation ([a898b09](https://github.com/indaco/sley/commit/a898b09))
- **bump:** add pre command to increment pre-release versions ([22ac845](https://github.com/indaco/sley/commit/22ac845))
- **plugins:** add group-icons config for changelog-generator ([6b1205c](https://github.com/indaco/sley/commit/6b1205c))
- **plugins:** add changelog-generator plugin ([fab1ea1](https://github.com/indaco/sley/commit/fab1ea1))
- integrate changelogparser with bump auto command ([810be6e](https://github.com/indaco/sley/commit/810be6e))
- **config:** add ChangelogParserConfig for changelog-parser plugin ([0c70be3](https://github.com/indaco/sley/commit/0c70be3))
- **plugins:** add changelogparser plugin for CHANGELOG.md analysis ([deb5cf2](https://github.com/indaco/sley/commit/deb5cf2))
- integrate dependencycheck with bump commands ([7fabd92](https://github.com/indaco/sley/commit/7fabd92))
- **plugins:** register dependencycheck plugin in loader ([dcff51e](https://github.com/indaco/sley/commit/dcff51e))
- **config:** add DependencyCheckConfig for dependency-check plugin ([670fc9e](https://github.com/indaco/sley/commit/670fc9e))
- **plugins:** add dependencycheck plugin for cross-file version consistency ([9880823](https://github.com/indaco/sley/commit/9880823))
- integrate versionvalidator with bump commands ([b930c83](https://github.com/indaco/sley/commit/b930c83))
- **plugins:** register versionvalidator plugin in loader ([0f2c51f](https://github.com/indaco/sley/commit/0f2c51f))
- **config:** add VersionValidatorConfig for version-validator plugin ([a7aa710](https://github.com/indaco/sley/commit/a7aa710))
- **plugins:** add versionvalidator plugin for version policy enforcement ([f5a8b72](https://github.com/indaco/sley/commit/f5a8b72))
- integrate tagmanager with bump commands ([93dec57](https://github.com/indaco/sley/commit/93dec57))
- **plugins:** register tagmanager plugin in loader ([4806d74](https://github.com/indaco/sley/commit/4806d74))
- **config:** add TagManagerConfig for tag-manager plugin ([b896063](https://github.com/indaco/sley/commit/b896063))
- **plugins:** add tagmanager plugin for git tag operations ([f13f926](https://github.com/indaco/sley/commit/f13f926))
- add ready-to-use extensions ([3a6e2fc](https://github.com/indaco/sley/commit/3a6e2fc))
- integrate extension hooks into bump commands ([2e1ee93](https://github.com/indaco/sley/commit/2e1ee93))
- add URL-based extension installation ([041eea4](https://github.com/indaco/sley/commit/041eea4))
- add extension system core ([8688f41](https://github.com/indaco/sley/commit/8688f41))
- **cli:** integrate multi-module support into commands ([f69be19](https://github.com/indaco/sley/commit/f69be19))
- **cli:** add modules command for workspace management ([7c1ec27](https://github.com/indaco/sley/commit/7c1ec27))
- **clix:** add execution context for multi-module mode ([9723e64](https://github.com/indaco/sley/commit/9723e64))
- **tui:** add spinner for progress indicators ([b46c185](https://github.com/indaco/sley/commit/b46c185))
- **operations:** add operations for module manipulation ([2503704](https://github.com/indaco/sley/commit/2503704))
- **workspace:** add executor for multi-module operations ([aeec349](https://github.com/indaco/sley/commit/aeec349))
- **workspace:** add module detection engine ([afee872](https://github.com/indaco/sley/commit/afee872))
- **config:** add workspace config schema for monorepo support ([e1899c3](https://github.com/indaco/sley/commit/e1899c3))
- **tui:** add TUI foundation for monorepo module selection ([4cb34ab](https://github.com/indaco/sley/commit/4cb34ab))
- **cli:** rename --no-auto-init to --strict ([bf2449d](https://github.com/indaco/sley/commit/bf2449d))
- **cli:** rename bump next to bump auto ([ecf31c2](https://github.com/indaco/sley/commit/ecf31c2))
- **cmdrunner:** add context support for command execution ([9c57d2d](https://github.com/indaco/sley/commit/9c57d2d))
- **pathutil:** add path validation utilities ([340f689](https://github.com/indaco/sley/commit/340f689))
- **errors:** add custom typed errors ([9e3139f](https://github.com/indaco/sley/commit/9e3139f))
- **core:** add interfaces for dependency injection ([e0ffe8b](https://github.com/indaco/sley/commit/e0ffe8b))
- introduce hooks on the bump commands ([08879df](https://github.com/indaco/sley/commit/08879df))
- feat/extension (#26) ([c43fc7a](https://github.com/indaco/sley/commit/c43fc7a))
- **cli:** add plugin 'remove' command to unregister plugins ([083b9ab](https://github.com/indaco/sley/commit/083b9ab))
- **cli:** add plugin 'list' command ([3085a96](https://github.com/indaco/sley/commit/3085a96))
- **cli:** add plugin 'register' command to register plugins ([bf4dec6](https://github.com/indaco/sley/commit/bf4dec6))
- **plugins:** add PluginManifest definition and loader with validation ([c0f54e6](https://github.com/indaco/sley/commit/c0f54e6))
- add static registry loader ([b03593f](https://github.com/indaco/sley/commit/b03593f))
- add plugin management commands ([705f82d](https://github.com/indaco/sley/commit/705f82d))
- add plugin management commands ([eb91bf3](https://github.com/indaco/sley/commit/eb91bf3))

### 🩹 Fixes

- **init:** write correct path to generated .sley.yaml ([8eee862](https://github.com/indaco/sley/commit/8eee862))
- **cli:** show correct default for --path flag in help text ([a2b64d6](https://github.com/indaco/sley/commit/a2b64d6))
- **changelog:** link PRs and commits for new contributors in changelog generator ([467c5a6](https://github.com/indaco/sley/commit/467c5a6))
- **lint:** resolve errorlint issues in error formatting ([6968757](https://github.com/indaco/sley/commit/6968757))
- **workspace:** add context cancellation checks in directory scanning ([0bba27d](https://github.com/indaco/sley/commit/0bba27d))
- **config:** validate SLEY_PATH env var to prevent path traversal ([839d6b2](https://github.com/indaco/sley/commit/839d6b2))
- **hooks:** add mutex protection to prevent race conditions ([be11c37](https://github.com/indaco/sley/commit/be11c37))
- **semver:** add input length validation to prevent ReDoS attacks ([39c9050](https://github.com/indaco/sley/commit/39c9050))
- **security:** add path traversal validation and race testing ([6d509af](https://github.com/indaco/sley/commit/6d509af))
- **showcmd:** use strict-aware version file and verbed formatter ([a69e39c](https://github.com/indaco/sley/commit/a69e39c))
- **clix:** skip TUI module selection for read-only commands ([6c3d2e1](https://github.com/indaco/sley/commit/6c3d2e1))
- **init:** prevent interactive prompts from blocking tests ([146692c](https://github.com/indaco/sley/commit/146692c))
- **initcmd:** remove redundant path field from workspace config ([e517c64](https://github.com/indaco/sley/commit/e517c64))
- **doctorcmd:** add multi-module support to validate command ([a1de338](https://github.com/indaco/sley/commit/a1de338))
- **clix:** return all modules with matching name in filter ([b4f277e](https://github.com/indaco/sley/commit/b4f277e))
- **changelog:** normalize trailing newlines in generated files ([9e56157](https://github.com/indaco/sley/commit/9e56157))
- **clix:** respect config path in multi-module detection ([14a211a](https://github.com/indaco/sley/commit/14a211a))
- **commitparser:** improve conventional commit parsing ([a660a35](https://github.com/indaco/sley/commit/a660a35))
- **security:** use secure file permissions for config ([79fefd7](https://github.com/indaco/sley/commit/79fefd7))
- **extension:** enforce mutually exclusive url and path flags for extension install ([f34c0e6](https://github.com/indaco/sley/commit/f34c0e6))
- **pluginmanager:** improve plugin registration flow and config handling ([9fc45b4](https://github.com/indaco/sley/commit/9fc45b4))
- **pluginmanager:** properly skip entire directories like .git in copyDir ([4d17ee1](https://github.com/indaco/sley/commit/4d17ee1))

### 💅 Refactors

- **plugins:** add Description and Version to plugin interfaces ([415cd01](https://github.com/indaco/sley/commit/415cd01))
- **core:** propagate context through GitBranchReader interface ([27199b9](https://github.com/indaco/sley/commit/27199b9))
- **workspace:** reduce function complexity ([7bb167a](https://github.com/indaco/sley/commit/7bb167a))
- **config:** split validator into focused modules ([0342438](https://github.com/indaco/sley/commit/0342438))
- pass large structs by pointer to avoid copying ([bae8afb](https://github.com/indaco/sley/commit/bae8afb))
- use exec.CommandContext for cancellable command execution ([abb555c](https://github.com/indaco/sley/commit/abb555c))
- **changelog-generator:** update default changelog header ([ebc5ac1](https://github.com/indaco/sley/commit/ebc5ac1))
- **plugins:** implement core interfaces for git and file operations ([0b391aa](https://github.com/indaco/sley/commit/0b391aa))
- **hooks:** use interface-based dependency injection for PreReleaseHookRunner ([c903b8b](https://github.com/indaco/sley/commit/c903b8b))
- **config:** use interface-based dependency injection for ConfigSaver ([b6e9532](https://github.com/indaco/sley/commit/b6e9532))
- **core:** add dependency injection interfaces and mock implementations ([546fd9c](https://github.com/indaco/sley/commit/546fd9c))
- **bump:** remove unused oldVersion variable ([7778ffb](https://github.com/indaco/sley/commit/7778ffb))
- reduce function lengths and improve readability ([15893fe](https://github.com/indaco/sley/commit/15893fe))
- **apperrors:** standardize error handling patterns ([de41969](https://github.com/indaco/sley/commit/de41969))
- **core:** centralize file permission and timeout constants ([b6c4951](https://github.com/indaco/sley/commit/b6c4951))
- **config:** split config.go into focused modules ([6d87501](https://github.com/indaco/sley/commit/6d87501))
- error message context enhancement ([cb9be77](https://github.com/indaco/sley/commit/cb9be77))
- **bumpcmd:** use shared logic in all bump commands ([9d2edc3](https://github.com/indaco/sley/commit/9d2edc3))
- **bumpcmd:** extract common validation and post-action logic ([070e0dc](https://github.com/indaco/sley/commit/070e0dc))
- **plugins:** export functions for registry access ([eec7309](https://github.com/indaco/sley/commit/eec7309))
- **bumpcmd:** use registry for plugin access ([52cd181](https://github.com/indaco/sley/commit/52cd181))
- **plugins:** wire registry through loader and CLI ([8b1d19b](https://github.com/indaco/sley/commit/8b1d19b))
- **plugins:** add thread-safe PluginRegistry struct ([8cc9a50](https://github.com/indaco/sley/commit/8cc9a50))
- **cli:** wire context from commands to operations ([85ed339](https://github.com/indaco/sley/commit/85ed339))
- **workspace:** add context to detector and module discovery ([73f4c03](https://github.com/indaco/sley/commit/73f4c03))
- **operations:** propagate context through all operations ([a1e4392](https://github.com/indaco/sley/commit/a1e4392))
- **config:** add context to validation methods ([ea5a1a4](https://github.com/indaco/sley/commit/ea5a1a4))
- **semver:** propagate context through VersionManager ([84f9cf9](https://github.com/indaco/sley/commit/84f9cf9))
- **core:** add context.Context to FileSystem interface ([b61c966](https://github.com/indaco/sley/commit/b61c966))
- **tui:** improve module selection UX ([7576762](https://github.com/indaco/sley/commit/7576762))
- **cli:** apply consistent styled output across commands ([88ff3e6](https://github.com/indaco/sley/commit/88ff3e6))
- **extensions:** remove dead api/v0 package ([9bd9091](https://github.com/indaco/sley/commit/9bd9091))
- rename project from verso to sley ([de90581](https://github.com/indaco/sley/commit/de90581))
- rename project from semver-cli to verso ([0c67805](https://github.com/indaco/sley/commit/0c67805))
- **cli:** move multi-module flags to command level ([0a26aac](https://github.com/indaco/sley/commit/0a26aac))
- **git:** use context-based command execution ([3745fff](https://github.com/indaco/sley/commit/3745fff))
- **semver:** add VersionManager for dependency injection ([bc7e96a](https://github.com/indaco/sley/commit/bc7e96a))
- **errors:** rename package to apperrors to avoid stdlib conflict ([c2e59ce](https://github.com/indaco/sley/commit/c2e59ce))
- **semver:** separate domain logic from I/O ([56cc61f](https://github.com/indaco/sley/commit/56cc61f))
- **extensionmgr:** rename extension-manager to extensionmgr ([a90cf76](https://github.com/indaco/sley/commit/a90cf76))
- **plugins:** rename commit-parser to commitparser ([4ccfd76](https://github.com/indaco/sley/commit/4ccfd76))
- **clix:** return typed errors instead of cli.Exit ([c506bc6](https://github.com/indaco/sley/commit/c506bc6))
- use function vars for hooks and clix and add tests ([a6bb85c](https://github.com/indaco/sley/commit/a6bb85c))
- **bumpNext:** simplify version bump logic and preserve metadata handling ([8495387](https://github.com/indaco/sley/commit/8495387))
- **cli:** organize commands into dedicated packages ([70f09df](https://github.com/indaco/sley/commit/70f09df))
- **api:** move interface and funcs to api/v0/plugins ([d5c1a44](https://github.com/indaco/sley/commit/d5c1a44))
- **plugin:** commit parser interface and registration functions ([37e29d7](https://github.com/indaco/sley/commit/37e29d7))
- **plugins:** move plugin interface and registry to api/plugins package ([166045d](https://github.com/indaco/sley/commit/166045d))
- make LoadPluginManifest unexported to limit scope ([6cb0abe](https://github.com/indaco/sley/commit/6cb0abe))
- **plugins:** extract Plugin interface to api package and update registry to use it ([29e5804](https://github.com/indaco/sley/commit/29e5804))
- **pluginmanager:** unexport functions and variables for encapsulation ([295bffa](https://github.com/indaco/sley/commit/295bffa))
- **cli:** rename plugin 'register' command to 'add' ([dd87158](https://github.com/indaco/sley/commit/dd87158))
- **testutils:** enhance CaptureStdout to capture stderr and handle errors ([e21939b](https://github.com/indaco/sley/commit/e21939b))

### 📖 Documentation

- clarify ai assistance section ([2e0def2](https://github.com/indaco/sley/commit/2e0def2))
- simplify main doc files ([359eab2](https://github.com/indaco/sley/commit/359eab2))
- update CHANGELOG with the default header ([5838138](https://github.com/indaco/sley/commit/5838138))
- update CONTRIBUTING to include govulncheck and the right go version ([421c149](https://github.com/indaco/sley/commit/421c149))
- simplify plugins docs ([5801db3](https://github.com/indaco/sley/commit/5801db3))
- **plugins:** add metadata sections to plugin documentation ([af1e2de](https://github.com/indaco/sley/commit/af1e2de))
- **changelog-generator:** new contributors detection section ([1f63190](https://github.com/indaco/sley/commit/1f63190))
- **changelog-generator:** document use-default-icons option ([c1a3b73](https://github.com/indaco/sley/commit/c1a3b73))
- **semver:** add consistent error documentation to functions ([d2168ac](https://github.com/indaco/sley/commit/d2168ac))
- **cmdrunner:** add security documentation for command execution ([d03dd2e](https://github.com/indaco/sley/commit/d03dd2e))
- **internal:** add doc.go to undocumented packages ([455f8d4](https://github.com/indaco/sley/commit/455f8d4))
- update doctor and init command descriptions ([fc0b726](https://github.com/indaco/sley/commit/fc0b726))
- update README with new init command features ([d9cebe9](https://github.com/indaco/sley/commit/d9cebe9))
- update plugins docs to explain init defaults vs built-in defaults ([b84e071](https://github.com/indaco/sley/commit/b84e071))
- **changelog:** clarify header-template usage and examples ([76dad12](https://github.com/indaco/sley/commit/76dad12))
- **readme:** add AI Assistance section and project logo ([160db0f](https://github.com/indaco/sley/commit/160db0f))
- **readme:** add logo and Homebrew installation ([3dcf764](https://github.com/indaco/sley/commit/3dcf764))
- update README ([71c354f](https://github.com/indaco/sley/commit/71c354f))
- update to include the new format config for changelog-generator plugin ([2926872](https://github.com/indaco/sley/commit/2926872))
- **changelog-generator:** document format configuration ([229ec6c](https://github.com/indaco/sley/commit/229ec6c))
- **plugins:** add audit-log and release-gate documentation ([62790c3](https://github.com/indaco/sley/commit/62790c3))
- **plugins:** add release-gate plugin documentation ([ac4e1f8](https://github.com/indaco/sley/commit/ac4e1f8))
- document bump pre command and changelog plugin comparison ([e7c59fa](https://github.com/indaco/sley/commit/e7c59fa))
- **plugins:** add changelog-generator documentation ([f431117](https://github.com/indaco/sley/commit/f431117))
- **plugins:** add changelogparser documentation and examples ([3767bd5](https://github.com/indaco/sley/commit/3767bd5))
- add dependency-check plugin to README ([75c4b32](https://github.com/indaco/sley/commit/75c4b32))
- **plugins:** refactor plugin docs into individual files with examples ([76e5f68](https://github.com/indaco/sley/commit/76e5f68))
- update README to include plugin section ([d810a68](https://github.com/indaco/sley/commit/d810a68))
- **plugins:** update documentation for versionvalidator plugin ([0f12124](https://github.com/indaco/sley/commit/0f12124))
- **plugins:** add tagmanager documentation ([b9efa14](https://github.com/indaco/sley/commit/b9efa14))
- add plugin integration examples to READMEs ([e843bfc](https://github.com/indaco/sley/commit/e843bfc))
- add plugin system documentation ([753d666](https://github.com/indaco/sley/commit/753d666))
- add extension system documentation ([083c8af](https://github.com/indaco/sley/commit/083c8af))
- add MONOREPO.md and reference it from README ([22b31aa](https://github.com/indaco/sley/commit/22b31aa))
- update README ([90c3768](https://github.com/indaco/sley/commit/90c3768))
- update documentation for API changes ([512d85d](https://github.com/indaco/sley/commit/512d85d))
- add package-level godoc documentation ([7a19206](https://github.com/indaco/sley/commit/7a19206))
- cleanup README ([94cfa19](https://github.com/indaco/sley/commit/94cfa19))

### ⚡ Performance

- **workspace:** add CachingDetector for discovery result caching ([13fabb5](https://github.com/indaco/sley/commit/13fabb5))
- **semver:** optimize Version.String() with strings.Builder ([327f61e](https://github.com/indaco/sley/commit/327f61e))
- **semver:** optimize IncrementPreRelease by eliminating regex compilation ([86b32c1](https://github.com/indaco/sley/commit/86b32c1))

### ✅ Tests

- **changelog:** split test file into focused test files ([6d3b3d8](https://github.com/indaco/sley/commit/6d3b3d8))
- sdd error recovery tests ([b1eafca](https://github.com/indaco/sley/commit/b1eafca))
- **core:** split mocks_test.go into focused test files ([c52d0b7](https://github.com/indaco/sley/commit/c52d0b7))
- **apperrors:** increase test coverage ([fa483a2](https://github.com/indaco/sley/commit/fa483a2))
- add fuzz tests for config and commit parsing ([af96548](https://github.com/indaco/sley/commit/af96548))
- **validator:** split large test files ([60a661c](https://github.com/indaco/sley/commit/60a661c))
- **config:** split large test file into focused modules ([c3ba579](https://github.com/indaco/sley/commit/c3ba579))
- **bumpcmd:** split large test file into focused modules ([ac20ae3](https://github.com/indaco/sley/commit/ac20ae3))
- improve test coverage across packages ([5a73f29](https://github.com/indaco/sley/commit/5a73f29))
- **extension:** add tests for RegisterLocalExtension error handling ([60de196](https://github.com/indaco/sley/commit/60de196))
- **bump:** mock implementations for plugin interfaces, tests for helper functions ([f6690e7](https://github.com/indaco/sley/commit/f6690e7))
- **plugins:** add test for branch constraint error handling, additional matchBranchPattern test cases ([033956c](https://github.com/indaco/sley/commit/033956c))
- **plugins:** add tests for GetConfig method in releasegate ([c29f375](https://github.com/indaco/sley/commit/c29f375))
- **config:** add more tests for plugin configs ([73e4c83](https://github.com/indaco/sley/commit/73e4c83))
- **plugins:** add coverage for loader and auditlog ([a6e4336](https://github.com/indaco/sley/commit/a6e4336))
- **setcmd,showcmd:** add multi-module operation tests ([99f0e5c](https://github.com/indaco/sley/commit/99f0e5c))
- improve test coverage across packages ([dd43c54](https://github.com/indaco/sley/commit/dd43c54))
- **config:** improve test helpers and mocks ([01ccf56](https://github.com/indaco/sley/commit/01ccf56))
- **integration:** add e2e tests for multi-module operations ([7790294](https://github.com/indaco/sley/commit/7790294))
- **integration:** add e2e integration tests for semver CLI ([c689a53](https://github.com/indaco/sley/commit/c689a53))
- **git:** fix error message expectation ([9857929](https://github.com/indaco/sley/commit/9857929))
- add fuzz and benchmark tests ([e793c38](https://github.com/indaco/sley/commit/e793c38))
- refactor extensioncmd and config tests to use helpers for clarity ([16734bd](https://github.com/indaco/sley/commit/16734bd))
- add plugin metadata registration and listing tests ([39689ac](https://github.com/indaco/sley/commit/39689ac))
- add unit test for default path fallback in config loading ([df40b18](https://github.com/indaco/sley/commit/df40b18))
- move runCLITest to testutils package ([fccd69d](https://github.com/indaco/sley/commit/fccd69d))
- refactor replacing local file helpers with testutils ([2267679](https://github.com/indaco/sley/commit/2267679))

### 🏡 Chores

- **sley:** add sley config with enabled plugins ([7f98542](https://github.com/indaco/sley/commit/7f98542))
- **lint:** enhance golangci-lint with additional linters ([fca3daa](https://github.com/indaco/sley/commit/fca3daa))
- update .sley.yaml with use-default-icons: true ([2972ba2](https://github.com/indaco/sley/commit/2972ba2))
- **go.mod:** run go mod tidy ([346305a](https://github.com/indaco/sley/commit/346305a))
- bump version to v0.6.0-rc5 ([4f92844](https://github.com/indaco/sley/commit/4f92844))
- bump version to v0.6.0-rc4 ([b5edd68](https://github.com/indaco/sley/commit/b5edd68))
- update devbox.lock ([3262b70](https://github.com/indaco/sley/commit/3262b70))
- add header-template path for changelog generation ([d4877fe](https://github.com/indaco/sley/commit/d4877fe))
- remove changie configuration ([0821b45](https://github.com/indaco/sley/commit/0821b45))
- **config:** set contributors icon in .sley.yaml ([7076d7f](https://github.com/indaco/sley/commit/7076d7f))
- remove redundant changelog-generator extension ([5787ec4](https://github.com/indaco/sley/commit/5787ec4))
- bump version to 0.6.0-rc3 ([8c8f848](https://github.com/indaco/sley/commit/8c8f848))
- code cleanup ([185159e](https://github.com/indaco/sley/commit/185159e))
- bump version to v0.6.0-rc2 ([ca9b709](https://github.com/indaco/sley/commit/ca9b709))
- bump version to v0.6.0-rc1 ([4e060ab](https://github.com/indaco/sley/commit/4e060ab))
- **cli:** adjust semver command usage string to be more concise ([3170bbf](https://github.com/indaco/sley/commit/3170bbf))
- configure changelog-generator and remove cliff.toml ([bca4d6d](https://github.com/indaco/sley/commit/bca4d6d))
- bump version to 0.6.0-alpha.3 ([03d72bc](https://github.com/indaco/sley/commit/03d72bc))
- **just:** add modernize, lint, reportcard targets and use check for install ([daf3a63](https://github.com/indaco/sley/commit/daf3a63))
- bump version to v0.6.0-alpha.2 ([2564dbb](https://github.com/indaco/sley/commit/2564dbb))
- add git hooks and changelog configuration ([c55ab6c](https://github.com/indaco/sley/commit/c55ab6c))
- update Go to latest (1.25.4) in devbox config ([a1ed01e](https://github.com/indaco/sley/commit/a1ed01e))
- remove unused section from semver.yaml ([002b114](https://github.com/indaco/sley/commit/002b114))
- add devbox.lock file ([d0d4350](https://github.com/indaco/sley/commit/d0d4350))
- **plugin:** scaffold plugin architecture and utilities ([999daa2](https://github.com/indaco/sley/commit/999daa2))

### 🤖 CI

- separate security scanning into dedicated workflow ([5b2bd33](https://github.com/indaco/sley/commit/5b2bd33))
- add govulncheck to CI pipeline ([a4e2838](https://github.com/indaco/sley/commit/a4e2838))
- **workflows:** standardize Go version and add just tasks ([c8b46e8](https://github.com/indaco/sley/commit/c8b46e8))
- update actions to latest versions ([f016787](https://github.com/indaco/sley/commit/f016787))
- bump Go to 1.25 and update golangci-lint action to v9/latest for CI consistency ([8196de8](https://github.com/indaco/sley/commit/8196de8))

### 📦 Build

- **release:** add Homebrew tap configuration ([14651fc](https://github.com/indaco/sley/commit/14651fc))
- add optimization flags to reduce binary size ([a06d93b](https://github.com/indaco/sley/commit/a06d93b))
- **deps:** update Go dependencies ([9c8711c](https://github.com/indaco/sley/commit/9c8711c))
- update devbox configuration and add scripts ([42550e1](https://github.com/indaco/sley/commit/42550e1))
- replace Makefile/Taskfile with justfile ([300d674](https://github.com/indaco/sley/commit/300d674))
- **deps:** bump Go to 1.25.4 and update urfave/cli ([e1bd492](https://github.com/indaco/sley/commit/e1bd492))
- update go deps ([04a08e6](https://github.com/indaco/sley/commit/04a08e6))
- **deps:** update urfave/cli to v3.2.0 ([deb1a81](https://github.com/indaco/sley/commit/deb1a81))
- **deps:** replace gopkg.in/yaml.v3 with goccy/go-yaml ([5fa1b20](https://github.com/indaco/sley/commit/5fa1b20))
- fix `build` commands in Taskfile and Makefile ([adf6c64](https://github.com/indaco/sley/commit/adf6c64))

**Full Changelog:** [v0.5.0...v0.6.0](https://github.com/indaco/sley/compare/v0.5.0...v0.6.0)

### ❤️ Contributors

- [@indaco](https://github.com/indaco)

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
