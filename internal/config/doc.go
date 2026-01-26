// Package config handles configuration loading and saving for sley.
//
// The configuration system uses a priority hierarchy for determining
// the version file path:
//
//  1. --path flag (highest priority)
//  2. SLEY_PATH environment variable
//  3. .sley.yaml configuration file
//  4. Default ".version" (lowest priority)
//
// # Configuration File Format
//
// The configuration file (.sley.yaml) supports the following options:
//
//	# Path to the version file
//	path: internal/version/.version
//
//	# Plugin configuration
//	plugins:
//	  commit-parser: true
//
//	# Extension configuration
//	# Extensions are external scripts that hook into version lifecycle events.
//	# See docs/EXTENSIONS.md for details on creating extensions.
//	extensions:
//	  - name: my-extension        # Extension identifier
//	    path: ./extensions/my-ext # Path to extension directory
//	    enabled: true             # Whether extension is active
//
//	  - name: git-changelog
//	    path: /home/user/.sley-extensions/git-changelog
//	    enabled: true
//
//	  - name: legacy-notifier
//	    path: ./extensions/notifier
//	    enabled: false            # Disabled extensions are ignored
//
//	# Pre-release hooks (executed before version bumps)
//	pre-release-hooks:
//	  - command: "make test"
//	  - command: "make lint"
//
// # Workspace Configuration (Monorepo Support)
//
// The workspace section configures multi-module/monorepo behavior:
//
//	workspace:
//	  # Auto-discovery settings (all optional, shown with defaults)
//	  discovery:
//	    enabled: true      # Enable module auto-discovery
//	    recursive: true    # Search subdirectories
//	    module_max_depth: 10      # Maximum directory depth
//	    exclude:           # Additional paths to exclude
//	      - custom_dir
//
//	  # Explicit module definitions (overrides discovery)
//	  modules:
//	    - name: module-a
//	      path: ./module-a/.version
//	    - name: module-b
//	      path: ./services/module-b/.version
//	      enabled: false   # Disable this module
//
// Discovery is zero-config by default. The following patterns are excluded:
// node_modules, .git, vendor, tmp, build, dist, .cache, __pycache__
//
// Helper methods:
//   - GetDiscoveryConfig() returns discovery settings with defaults applied
//   - GetExcludePatterns() returns merged default + configured excludes
//   - HasExplicitModules() returns true if modules are explicitly defined
//   - IsModuleEnabled(name) checks if a specific module is enabled
//
// # Security
//
// Configuration files are created with 0600 permissions (owner read/write only)
// to protect any sensitive hook commands from being readable by other users.
package config
