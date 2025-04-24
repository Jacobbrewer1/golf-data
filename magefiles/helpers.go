//go:build mage

package main

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	gcpStorageHost = "https://storage.googleapis.com"
)

const (
	envCI                            = "CI"
	envGithubActions                 = "GITHUB_ACTIONS"
	envRepositoryName                = "GITHUB_REPOSITORY"
	envRepositoryOwner               = "GITHUB_REPOSITORY_OWNER"
	envGCPServiceAccountJsonLocation = "GOOGLE_APPLICATION_CREDENTIALS_JSON_PATH"
	envBazelRemoteCacheBucket        = "BAZEL_REMOTE_CACHE_BUCKET"
)

var RemoteCacheEnabled = sync.OnceValue(func() bool {
	ciRunner, _ := strconv.ParseBool(os.Getenv(envCI))
	githubRunner, _ := strconv.ParseBool(os.Getenv(envGithubActions))
	if ciRunner || githubRunner {
		return true
	}
	return false
})

var RemoteCacheBucket = sync.OnceValue(func() string {
	bucket := os.Getenv(envBazelRemoteCacheBucket)
	if bucket == "" {
		return ""
	}
	return bucket
})

// RepositoryName returns the name of the repository.
var RepositoryName = sync.OnceValue(func() string {
	name := os.Getenv(envRepositoryName)
	if name == "" {
		return ""
	}
	return name
})

// RepositoryOwner returns the owner of the repository.
var RepositoryOwner = sync.OnceValue(func() string {
	owner := os.Getenv(envRepositoryOwner)
	if owner == "" {
		return ""
	}
	return owner
})

// RepositoryNameOnly returns the name of the repository without the owner.
var RepositoryNameOnly = sync.OnceValue(func() string {
	name := RepositoryName()
	if name == "" {
		return ""
	}

	owner := RepositoryOwner()
	if owner == "" {
		return name
	}

	name = strings.TrimPrefix(name, owner+"/")
	return name
})

var GCPServiceAccountJsonLocation = sync.OnceValue(func() string {
	location := os.Getenv(envGCPServiceAccountJsonLocation)
	if location == "" {
		return ""
	}
	return location
})
