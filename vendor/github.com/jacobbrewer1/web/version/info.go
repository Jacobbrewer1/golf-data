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
	//
	// This function uses sync.OnceValue to ensure the value is computed only once
	// and cached for subsequent calls. It retrieves build information using
	// debug.ReadBuildInfo and searches for the revisionKey in the build settings.
	//
	// Returns:
	//   - string: The git commit hash if found, otherwise an empty string.
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
	//
	// This function uses sync.OnceValue to ensure the value is computed only once
	// and cached for subsequent calls. It retrieves build information using
	// debug.ReadBuildInfo and searches for the commitTimestampKey in the build settings.
	// If the timestamp is found, it is parsed using the RFC3339 format.
	// If parsing fails or the key is not found, it returns the zero value of time.Time.
	//
	// Returns:
	//   - time.Time: The commit timestamp if found and successfully parsed, otherwise the zero value of time.Time.
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
	//
	// This function uses sync.OnceValue to ensure the value is computed only once
	// and cached for subsequent calls. It retrieves build information using
	// debug.ReadBuildInfo and searches for the modifiedKey in the build settings.
	// If the key is found, its value is parsed as a boolean. If parsing fails or the key is not found,
	// the function returns false.
	//
	// Returns:
	//   - bool: True if uncommitted changes were present during the build, otherwise false.
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
