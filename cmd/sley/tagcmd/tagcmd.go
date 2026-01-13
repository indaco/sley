package tagcmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/indaco/sley/internal/config"
	"github.com/indaco/sley/internal/core"
	"github.com/indaco/sley/internal/plugins/tagmanager"
	"github.com/indaco/sley/internal/printer"
	"github.com/indaco/sley/internal/semver"
	"github.com/urfave/cli/v3"
)

// TagCommand handles tag operations with injected dependencies.
type TagCommand struct {
	gitOps core.GitTagOperations
}

// NewTagCommand creates a new TagCommand with the given git operations.
func NewTagCommand(gitOps core.GitTagOperations) *TagCommand {
	return &TagCommand{gitOps: gitOps}
}

// NewDefaultTagCommand creates a TagCommand with the default OS git operations.
func NewDefaultTagCommand() *TagCommand {
	return NewTagCommand(tagmanager.NewOSGitTagOperations())
}

// Run returns the "tag" command with subcommands.
func Run(cfg *config.Config) *cli.Command {
	tc := NewDefaultTagCommand()
	return &cli.Command{
		Name:  "tag",
		Usage: "Manage git tags for versions",
		Commands: []*cli.Command{
			tc.createCmd(cfg),
			tc.listCmd(cfg),
			tc.pushCmd(cfg),
			tc.deleteCmd(cfg),
		},
	}
}

// createCmd returns the "tag create" subcommand.
func (tc *TagCommand) createCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:      "create",
		Aliases:   []string{"c", "new"},
		Usage:     "Create a git tag for the current version",
		UsageText: "sley tag create [--push] [--message <msg>]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "push",
				Usage: "Push the tag to remote after creation",
			},
			&cli.StringFlag{
				Name:    "message",
				Aliases: []string{"m"},
				Usage:   "Override the tag message (for annotated/signed tags)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return tc.runCreateCmd(ctx, cmd, cfg)
		},
	}
}

// listCmd returns the "tag list" subcommand.
func (tc *TagCommand) listCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:      "list",
		Aliases:   []string{"l", "ls"},
		Usage:     "List existing version tags",
		UsageText: "sley tag list [--limit <n>]",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"n"},
				Usage:   "Limit the number of tags shown",
				Value:   0,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return tc.runListCmd(ctx, cmd, cfg)
		},
	}
}

// pushCmd returns the "tag push" subcommand.
func (tc *TagCommand) pushCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:      "push",
		Aliases:   []string{"p"},
		Usage:     "Push a tag to remote",
		UsageText: "sley tag push [tag-name]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return tc.runPushCmd(ctx, cmd, cfg)
		},
	}
}

// deleteCmd returns the "tag delete" subcommand.
func (tc *TagCommand) deleteCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Aliases:   []string{"d", "rm"},
		Usage:     "Delete a git tag",
		UsageText: "sley tag delete <tag-name> [--remote]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "remote",
				Usage: "Also delete the tag from remote",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return tc.runDeleteCmd(ctx, cmd, cfg)
		},
	}
}

// runCreateCmd creates a git tag for the current version.
func (tc *TagCommand) runCreateCmd(_ context.Context, cmd *cli.Command, cfg *config.Config) error {
	// Check if tag-manager plugin is enabled
	if !isTagManagerEnabled(cfg) {
		printer.PrintWarning("Warning: The tag-manager plugin is not enabled.")
		printer.PrintInfo("To enable it, add the following to your .sley.yaml:")
		fmt.Println("")
		fmt.Println("  plugins:")
		fmt.Println("    tag-manager:")
		fmt.Println("      enabled: true")
		fmt.Println("      auto-create: false  # Disable auto-tagging during bump")
		fmt.Println("")
		printer.PrintInfo("Proceeding with tag creation using default settings...")
		fmt.Println("")
	}

	path := getVersionPath(cmd, cfg)

	version, err := semver.ReadVersion(path)
	if err != nil {
		return fmt.Errorf("failed to read version from %s: %w", path, err)
	}

	tmConfig := buildTagManagerConfig(cfg)
	prefix := tmConfig.Prefix
	tagName := prefix + version.String()

	exists, err := tc.gitOps.TagExists(tagName)
	if err != nil {
		return fmt.Errorf("failed to check tag existence: %w", err)
	}
	if exists {
		return fmt.Errorf("tag %s already exists", tagName)
	}

	message := cmd.String("message")
	if message == "" {
		data := tagmanager.NewTemplateData(version, prefix)
		message = tagmanager.FormatMessage(tmConfig.MessageTemplate, data)
	}

	if err := tc.createTag(tagName, message, tmConfig); err != nil {
		return err
	}

	printer.PrintSuccess(fmt.Sprintf("Created tag %s", tagName))

	shouldPush := cmd.Bool("push") || tmConfig.Push
	if shouldPush {
		if err := tc.gitOps.PushTag(tagName); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}
		printer.PrintSuccess(fmt.Sprintf("Pushed tag %s to remote", tagName))
	}

	return nil
}

// createTag creates a tag based on the configuration.
func (tc *TagCommand) createTag(tagName, message string, cfg *tagmanager.Config) error {
	switch {
	case cfg.Sign:
		if err := tc.gitOps.CreateSignedTag(tagName, message, cfg.SigningKey); err != nil {
			return fmt.Errorf("failed to create signed tag: %w", err)
		}
	case cfg.Annotate:
		if err := tc.gitOps.CreateAnnotatedTag(tagName, message); err != nil {
			return fmt.Errorf("failed to create annotated tag: %w", err)
		}
	default:
		if err := tc.gitOps.CreateLightweightTag(tagName); err != nil {
			return fmt.Errorf("failed to create lightweight tag: %w", err)
		}
	}
	return nil
}

// runListCmd lists existing version tags.
func (tc *TagCommand) runListCmd(_ context.Context, cmd *cli.Command, cfg *config.Config) error {
	prefix := getTagPrefix(cfg)
	pattern := prefix + "*"

	tags, err := tc.gitOps.ListTags(pattern)
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) == 0 {
		printer.PrintInfo(fmt.Sprintf("No tags found matching pattern %q", pattern))
		return nil
	}

	sortTagsBySemver(tags, prefix)

	limit := cmd.Int("limit")
	if limit > 0 && limit < len(tags) {
		tags = tags[:limit]
	}

	for _, tag := range tags {
		fmt.Println(tag)
	}

	return nil
}

// runPushCmd pushes a tag to remote.
func (tc *TagCommand) runPushCmd(_ context.Context, cmd *cli.Command, cfg *config.Config) error {
	var tagName string

	if cmd.NArg() > 0 {
		tagName = cmd.Args().Get(0)
	} else {
		path := getVersionPath(cmd, cfg)
		version, err := semver.ReadVersion(path)
		if err != nil {
			return fmt.Errorf("failed to read version from %s: %w", path, err)
		}

		prefix := getTagPrefix(cfg)
		tagName = prefix + version.String()
	}

	exists, err := tc.gitOps.TagExists(tagName)
	if err != nil {
		return fmt.Errorf("failed to check tag existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("tag %s does not exist locally", tagName)
	}

	if err := tc.gitOps.PushTag(tagName); err != nil {
		return fmt.Errorf("failed to push tag: %w", err)
	}

	printer.PrintSuccess(fmt.Sprintf("Pushed tag %s to remote", tagName))
	return nil
}

// runDeleteCmd deletes a git tag.
func (tc *TagCommand) runDeleteCmd(_ context.Context, cmd *cli.Command, _ *config.Config) error {
	if cmd.NArg() < 1 {
		return cli.Exit("missing required tag name argument", 1)
	}

	tagName := cmd.Args().Get(0)

	exists, err := tc.gitOps.TagExists(tagName)
	if err != nil {
		return fmt.Errorf("failed to check tag existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("tag %s does not exist locally", tagName)
	}

	if err := tc.gitOps.DeleteTag(tagName); err != nil {
		return fmt.Errorf("failed to delete local tag: %w", err)
	}
	printer.PrintSuccess(fmt.Sprintf("Deleted local tag %s", tagName))

	if cmd.Bool("remote") {
		if err := tc.gitOps.DeleteRemoteTag(tagName); err != nil {
			return fmt.Errorf("failed to delete remote tag: %w", err)
		}
		printer.PrintSuccess(fmt.Sprintf("Deleted remote tag %s", tagName))
	}

	return nil
}

// isTagManagerEnabled checks if the tag-manager plugin is enabled.
func isTagManagerEnabled(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	if cfg.Plugins == nil {
		return false
	}
	if cfg.Plugins.TagManager == nil {
		return false
	}
	return cfg.Plugins.TagManager.Enabled
}

// getVersionPath returns the version file path from flags or config.
func getVersionPath(cmd *cli.Command, cfg *config.Config) string {
	if cmd.IsSet("path") {
		return cmd.String("path")
	}
	if cfg != nil && cfg.Path != "" {
		return cfg.Path
	}
	return ".version"
}

// getTagPrefix returns the tag prefix from config or default "v".
func getTagPrefix(cfg *config.Config) string {
	if cfg != nil && cfg.Plugins != nil && cfg.Plugins.TagManager != nil {
		return cfg.Plugins.TagManager.GetPrefix()
	}
	return "v"
}

// buildTagManagerConfig creates a tagmanager.Config from the sley config.
func buildTagManagerConfig(cfg *config.Config) *tagmanager.Config {
	tmConfig := tagmanager.DefaultConfig()

	if cfg == nil || cfg.Plugins == nil || cfg.Plugins.TagManager == nil {
		return tmConfig
	}

	tmCfg := cfg.Plugins.TagManager
	tmConfig.Enabled = tmCfg.Enabled
	tmConfig.AutoCreate = tmCfg.GetAutoCreate()
	tmConfig.Prefix = tmCfg.GetPrefix()
	tmConfig.Annotate = tmCfg.GetAnnotate()
	tmConfig.Push = tmCfg.Push
	tmConfig.TagPrereleases = tmCfg.GetTagPrereleases()
	tmConfig.Sign = tmCfg.GetSign()
	tmConfig.SigningKey = tmCfg.GetSigningKey()
	tmConfig.MessageTemplate = tmCfg.GetMessageTemplate()

	return tmConfig
}

// sortTagsBySemver sorts tags by semantic version in descending order (newest first).
func sortTagsBySemver(tags []string, prefix string) {
	sort.Slice(tags, func(i, j int) bool {
		vi := parseVersionFromTag(tags[i], prefix)
		vj := parseVersionFromTag(tags[j], prefix)

		if vi.Major != vj.Major {
			return vi.Major > vj.Major
		}
		if vi.Minor != vj.Minor {
			return vi.Minor > vj.Minor
		}
		if vi.Patch != vj.Patch {
			return vi.Patch > vj.Patch
		}

		if vi.PreRelease == "" && vj.PreRelease != "" {
			return true
		}
		if vi.PreRelease != "" && vj.PreRelease == "" {
			return false
		}

		return vi.PreRelease > vj.PreRelease
	})
}

// parseVersionFromTag parses a semver from a tag string, stripping the prefix.
func parseVersionFromTag(tag, prefix string) semver.SemVersion {
	versionStr := strings.TrimPrefix(tag, prefix)
	version, err := semver.ParseVersion(versionStr)
	if err != nil {
		return semver.SemVersion{}
	}
	return version
}
