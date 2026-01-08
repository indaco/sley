package core

import (
	"io/fs"
	"time"
)

// File permission constants for consistent security across the codebase.
// These follow the principle of least privilege.
const (
	// PermOwnerRW is read/write for owner only (0600).
	// Use for sensitive files: config, version files, audit logs.
	PermOwnerRW fs.FileMode = 0600

	// PermOwnerRWGroupR is read/write for owner, read for group (0640).
	// Use for files that need to be readable by group members.
	PermOwnerRWGroupR fs.FileMode = 0640

	// PermPublicRead is read/write for owner, read for all (0644).
	// Use for public files: changelogs, documentation.
	PermPublicRead fs.FileMode = 0644

	// PermDirDefault is read/write/execute for owner, read/execute for others (0755).
	// Use for directories that need to be traversable.
	PermDirDefault fs.FileMode = 0755

	// PermDirPrivate is read/write/execute for owner only (0700).
	// Use for directories containing sensitive data.
	PermDirPrivate fs.FileMode = 0700

	// PermExecutable is the executable bit mask (0111).
	// Use to check if a file has any executable permission.
	PermExecutable fs.FileMode = 0111
)

// Timeout constants for external operations.
const (
	// TimeoutDefault is the default timeout for external commands (30 seconds).
	TimeoutDefault = 30 * time.Second

	// TimeoutShort is a shorter timeout for quick operations (5 seconds).
	TimeoutShort = 5 * time.Second

	// TimeoutLong is a longer timeout for potentially slow operations (2 minutes).
	TimeoutLong = 2 * time.Minute

	// TimeoutGit is the default timeout for git operations (60 seconds).
	// Git operations like clone and pull can take longer than regular commands.
	TimeoutGit = 60 * time.Second

	// TimeoutExtension is the default timeout for extension script execution (30 seconds).
	// This matches TimeoutDefault for consistency with other external operations.
	TimeoutExtension = TimeoutDefault
)

// Discovery constants for workspace module discovery.
const (
	// MaxDiscoveryDepth is the default maximum directory depth for module discovery.
	MaxDiscoveryDepth = 10
)
