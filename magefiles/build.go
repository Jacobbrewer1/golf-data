//go:build mage

package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Build mg.Namespace

// All builds all applications
func (b Build) All() error {
	mg.Deps(Init)
	log(slog.LevelInfo, "Building all code")

	start := time.Now()

	args := []string{
		"build",
	}

	if isCIRunner() {
		args = appendBazelArgs(args)
	}
	args = append(args, "//...")

	if err := sh.Run("bazel", args...); err != nil {
		return fmt.Errorf("error building all code: %w", err)
	}

	log(slog.LevelInfo, fmt.Sprintf("Build completed in %s", time.Since(start)))
	return nil
}

// One builds a single application
func (b Build) One(service string) error {
	mg.Deps(Init)
	log(slog.LevelInfo, fmt.Sprintf("Building %s", service))

	start := time.Now()

	args := []string{
		"build",
	}

	if isCIRunner() {
		args = appendBazelArgs(args)
	}
	args = append(args, "//cmd/"+service)

	if err := sh.Run("bazel", args...); err != nil {
		return fmt.Errorf("error building %s: %w", service, err)
	}

	log(slog.LevelInfo, fmt.Sprintf("Build of %s completed in %s", service, time.Since(start)))
	return nil
}
