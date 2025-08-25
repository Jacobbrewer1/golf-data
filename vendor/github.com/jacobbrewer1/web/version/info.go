package version

import (
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

const (
	// vcsKey is the key for identifying the version control system.
	vcsKey = "vcs"

	// revisionKey is the key for retrieving the revision (commit hash) of the build.
	revisionKey = vcsKey + ".revision"

	// commitTimestampKey is the key for retrieving the timestamp of the commit.
	commitTimestampKey = vcsKey + ".time"

	// modifiedKey is the key for indicating if the binary was built from a working directory
	// containing uncommitted changes. The value is set to "true" as a string if changes are present.
	modifiedKey = vcsKey + ".modified"
)

var (
	// GitCommit returns the git commit hash of the current build.
	GitCommit = sync.OnceValue(func() string {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == revisionKey {
					return setting.Value
				}
			}
		}
		return ""
	})

	// CommitTimestamp returns the timestamp of the commit.
	CommitTimestamp = sync.OnceValue(func() time.Time {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key != commitTimestampKey {
					continue
				}

				t, err := time.Parse(time.RFC3339, setting.Value)
				if err != nil {
					return time.Time{}
				}

				return t
			}
		}
		return time.Time{}
	})

	// IsModified returns whether the binary was built from a working directory containing uncommitted changes.
	IsModified = sync.OnceValue(func() bool {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key != modifiedKey {
					continue
				}

				containsChanges, err := strconv.ParseBool(setting.Value)
				if err != nil {
					return false
				}
				return containsChanges
			}
		}
		return false
	})
)
