package commitparser

import (
	"errors"
	"regexp"
	"strings"
)

// Conventional commit patterns for parsing.
var (
	// Matches breaking change indicator: type! or type(scope)!
	breakingExclamationRe = regexp.MustCompile(`^[a-z]+(\([a-z0-9_-]+\))?!:`)
	// Matches BREAKING CHANGE: or BREAKING-CHANGE: footer in commit body
	breakingFooterRe = regexp.MustCompile(`(?i)\nBREAKING[- ]CHANGE:`)
	// Matches feat or feat(scope):
	featRe = regexp.MustCompile(`^feat(\([a-z0-9_-]+\))?:`)
	// Matches fix or fix(scope):
	fixRe = regexp.MustCompile(`^fix(\([a-z0-9_-]+\))?:`)
)

/* ------------------------------------------------------------------------- */
/* INTERFACES                                                                */
/* ------------------------------------------------------------------------- */

// CommitParser defines the interface for parsing a list of commit messages
// and determining the corresponding semver bump type.

type CommitParser interface {
	Name() string
	Parse(commits []string) (string, error)
}

type CommitParserPlugin struct{}

func (CommitParserPlugin) Name() string { return "commit-parser" }
func (CommitParserPlugin) Description() string {
	return "Parses conventional commits to infer bump type"
}
func (CommitParserPlugin) Version() string { return "v0.1.0" }

/* ------------------------------------------------------------------------- */
/* IMPLEMENTATION                                                            */
/* ------------------------------------------------------------------------- */

// NewCommitParser returns a new Conventional Commits parser.
func NewCommitParser() CommitParser {
	return &CommitParserPlugin{}
}

// Parse analyzes a slice of commit messages and infers the semver bump type.
// It returns "major", "minor", "patch", or an error if no inference is possible.
func (p *CommitParserPlugin) Parse(commits []string) (string, error) {
	hasBreaking := false
	hasFeat := false
	hasFix := false

	for _, commit := range commits {
		lowerCommit := strings.ToLower(commit)

		// Check for breaking changes: feat!:, fix!:, or BREAKING CHANGE: footer
		if breakingExclamationRe.MatchString(lowerCommit) || breakingFooterRe.MatchString(commit) {
			hasBreaking = true
			continue
		}

		// Check for feat or feat(scope):
		if featRe.MatchString(lowerCommit) {
			hasFeat = true
			continue
		}

		// Check for fix or fix(scope):
		if fixRe.MatchString(lowerCommit) {
			hasFix = true
			continue
		}

		// Fallback: check for "breaking change" anywhere in message
		if strings.Contains(lowerCommit, "breaking change") {
			hasBreaking = true
		}
	}

	switch {
	case hasBreaking:
		return "major", nil
	case hasFeat:
		return "minor", nil
	case hasFix:
		return "patch", nil
	default:
		return "", errors.New("no bump type could be inferred")
	}
}

/* ------------------------------------------------------------------------- */
/* REGISTRATION                                                              */
/* ------------------------------------------------------------------------- */

// Register registers the commit parser plugin with the sley plugin system.
func Register() {
	RegisterCommitParserFn(&CommitParserPlugin{})
}
