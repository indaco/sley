package extension

import (
	"context"
	"fmt"

	"github.com/indaco/sley/internal/extensionmgr"
	"github.com/urfave/cli/v3"
)

// installCmd returns the "install" subcommand.
func installCmd() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install an extension from a remote repo or local path",
		Description: `Install an extension from a local path or remote Git repository.

Supported URL formats (any git-accessible host):
  - https://github.com/user/repo
  - https://github.com/user/repo@v1.0.0 (specific version)
  - https://github.com/user/repo@develop (from branch)
  - https://github.com/user/repo@abc123 (specific commit)
  - github.com/user/repo/path/to/extension (with subdirectory)
  - github.com/user/repo/path/to/extension@v2.0.0 (subdirectory with version)
  - https://gitlab.com/user/repo (GitLab)
  - https://git.example.com/user/repo (self-hosted)

Examples:
  # Install from local path
  sley extension install --path ./my-extension

  # Install latest from default branch
  sley extension install --url https://github.com/user/sley-ext-changelog

  # Install specific version
  sley extension install --url github.com/user/sley-ext-changelog@v1.0.0

  # Install from branch
  sley extension install --url github.com/user/sley-ext-changelog@develop

  # Install from subdirectory with version
  sley extension install --url github.com/indaco/sley/contrib/extensions/changelog-generator@v2.0.0

  # Install from self-hosted git
  sley extension install --url https://git.company.com/team/extension@main`,
		MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
			{
				Flags: [][]cli.Flag{
					{
						&cli.StringFlag{Name: "url", Usage: "Git URL to clone"},
						&cli.StringFlag{Name: "path", Usage: "Local path to copy from"},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "extension-dir", Usage: "Directory to store extensions in", Value: "."},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExtensionInstall(cmd)
		},
	}
}

// runExtensionInstall installs an extension from local or remote.
func runExtensionInstall(cmd *cli.Command) error {
	localPath := cmd.String("path")
	urlStr := cmd.String("url")

	// Check that at least one source is provided
	if localPath == "" && urlStr == "" {
		return cli.Exit("missing --path or --url for extension installation", 1)
	}

	// Get the extension directory (use the provided flag or default to current directory)
	extensionDirectory := cmd.String("extension-dir")

	// Handle URL-based installation
	if urlStr != "" {
		// Validate git is available
		if err := extensionmgr.ValidateGitAvailable(); err != nil {
			return cli.Exit(fmt.Sprintf("Error: %v", err), 1)
		}

		// Install from URL
		if err := extensionmgr.InstallFromURL(urlStr, ".sley.yaml", extensionDirectory); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to install extension from URL: %v", err), 1)
		}
		return nil
	}

	// Handle local path installation
	if localPath != "" {
		// Reject if URL is detected in --path flag
		if extensionmgr.IsURL(localPath) {
			return cli.Exit("URL detected in --path flag. Please use --url flag for remote installations.", 1)
		}

		// Proceed with normal extension registration from local path
		if err := extensionmgr.RegisterLocalExtensionFn(localPath, ".sley.yaml", extensionDirectory); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to install extension: %v", err), 1)
		}
		return nil
	}

	return cli.Exit("no installation source provided", 1)
}
