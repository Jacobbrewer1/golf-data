//go:build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Build mg.Namespace

// All builds all applications
func (b Build) All() error {
	if err := buildWithBazel("..."); err != nil {
		return err
	}
	return nil
}

// One builds a single application
func (b Build) One(service string) error {
	if err := buildWithBazel("cmd/" + service); err != nil {
		return err
	}
	return nil
}

// buildWithBazel builds the specified target using Bazel.
func buildWithBazel(target string) error {
	args := make([]string, 0)

	args = append(args, "build", "//"+target)

	if RemoteCacheEnabled() {
		fmt.Println("Using remote cache")
		cacheBucket := fmt.Sprintf("b3-prod-1-bazel-%s-cache", RepositoryNameOnly())
		args = append(args, fmt.Sprintf("--remote_cache=%s/%s", gcpStorageHost, cacheBucket))
		args = append(args, fmt.Sprintf("--google_credentials=%s", GCPServiceAccountJsonLocation()))
	}

	if err := sh.Run("bazel", args...); err != nil {
		return err
	}

	return nil
}
