package utils

import (
	"runtime/debug"
	"strconv"
	"time"
)

const (
	// revisionKey is the key for the revision.
	revisionKey = "vcs.revision"

	// buildDateKey is the key for the build date.
	buildDateKey = "vcs.time" // Timestamp of the commit

	// modifiedKey is the key for the modified flag.
	modifiedKey = "vcs.modified" // Set to true (as a string) if the binary was built from a working directory containing uncommitted changes.
)

var GitCommit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key != revisionKey {
				continue
			}
			return setting.Value
		}
	}
	return ""
}

var CommitTimestamp = func() time.Time {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key != buildDateKey {
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
}

var IsModified = func() bool {
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
}
