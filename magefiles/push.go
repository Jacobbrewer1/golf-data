//go:build mage

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// containerRegistry returns the container registry to use for pushing images.
var containerRegistry = sync.OnceValue(func() string {
	// Check if there's an override registry set in the environment
	overrideRegistry := os.Getenv("CONTAINER_REGISTRY")
	if overrideRegistry != "" {
		return overrideRegistry
	}

	// Default to the GitHub Container Registry
	repoPath := os.Getenv("GITHUB_REPOSITORY")
	if repoPath == "" {
		panic("GITHUB_REPOSITORY environment variable is not set")
	}

	repoPath = strings.ToLower(repoPath)

	// repo path current format is "owner/repo"

	// Append the GitHub container registry URL
	return "ghcr.io/" + repoPath
})

// commitTag returns the latest git tag or commit hash if no tags are found.
var commitTag = sync.OnceValue(func() string {
	got, err := sh.Output("git", "describe", "--tags", "--abbrev=0")
	if err == nil && got != "" {
		return strings.TrimSpace(got)
	}

	got, err = sh.Output("git", "describe", "--tags")
	if err == nil && got != "" {
		return strings.TrimSpace(got)
	}

	// Fallback to the commit hash if no tags are found
	got, err = sh.Output("git", "rev-parse", "--short", "HEAD")
	if err == nil && got != "" {
		return strings.TrimSpace(got)
	}

	panic("could not determine git tag")
})

// shouldPush determines if the image should be pushed to the container registry.
var shouldPush = sync.OnceValue(func() bool {
	overrideStr := os.Getenv("PUSH_IMAGES")
	if overrideStr != "" {
		log(slog.LevelDebug, "Overriding push images with "+overrideStr)
		overrideVal, _ := strconv.ParseBool(overrideStr)
		return overrideVal
	}

	return isCIRunner()
})

type Push mg.Namespace

// All builds and pushes all container targets to the target environment container registry.
func (Push) All() error {
	mg.Deps(Build.All)

	// Grab all the targets from the cmd directory
	targets, err := os.ReadDir("cmd")
	if err != nil {
		return fmt.Errorf("reading cmd directory: %w", err)
	}

	// Filter out the directories that are not services
	serviceTargets := make([]string, 0, len(targets))
	for _, target := range targets {
		if target.IsDir() {
			serviceTargets = append(serviceTargets, target.Name())
		}
	}

	if err := pushImages(serviceTargets, commitTag()); err != nil {
		return fmt.Errorf("pushing images: %w", err)
	}

	return nil
}

// One builds and pushes the specified container target to the target environment container registry.
func (Push) One(target string) error {
	mg.Deps(mg.F(Build.One, target))

	// Grab all the targets from the cmd directory
	targets, err := os.ReadDir("cmd")
	if err != nil {
		return fmt.Errorf("reading cmd directory: %w", err)
	}

	// Filter out the directories that are not services
	serviceTargets := make([]string, 0, len(targets))
	for _, target := range targets {
		if target.IsDir() {
			serviceTargets = append(serviceTargets, target.Name())
		}
	}

	if err := pushImages(serviceTargets, commitTag()); err != nil {
		return fmt.Errorf("pushing images: %w", err)
	}

	return nil
}

// pushImages pushes the specified images to the target environment container registry.
func pushImages(targets []string, tag string) error {
	errChan := make(chan error, 1)
	wg := new(sync.WaitGroup)
	wg.Add(len(targets))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, target := range targets {
		go func(target string) {
			defer wg.Done()

			log(slog.LevelDebug, "Pushing image "+target)
			if err := pushImage(ctx, "cmd/"+target, tag); err != nil {
				errChan <- fmt.Errorf("pushing image %s: %w", target, err)
			}
		}(target)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	log(slog.LevelDebug, "Watching push errors channel")
	errs := make([]error, 0, len(targets))
	for err := range errChan {
		errs = append(errs, err)
	}

	log(slog.LevelDebug, "Waiting for all push jobs to finish")
	wg.Wait()

	return errors.Join(errs...)
}

// pushImage pushes the specified image to the target environment container registry.
func pushImage(ctx context.Context, target, tag string) error {
	tarballPath := filepath.Join(binDirectory(), target, "oci.tar", "tarball.tar")
	if err := sh.Run(
		"docker",
		"load",
		"--input", tarballPath,
	); err != nil {
		return fmt.Errorf("loading tarball %s to image: %w", tarballPath, err)
	}

	target = strings.TrimPrefix(target, "cmd/")

	// Is the context cancelled?
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Prevent blocking
	}

	log(slog.LevelDebug, "Tagging images with "+tag)

	// Tag the image
	if err := sh.Run(
		"docker",
		"tag",
		target,
		fmt.Sprintf("%s/%s:%s", containerRegistry(), target, tag),
	); err != nil {
		return fmt.Errorf("tagging image %s: %w", target, err)
	}

	// Tag the image with the Latest tag
	if err := sh.Run(
		"docker",
		"tag",
		target,
		fmt.Sprintf("%s/%s:latest", containerRegistry(), target),
	); err != nil {
		return fmt.Errorf("tagging image %s with latest tag: %w", target, err)
	}

	if !shouldPush() {
		log(slog.LevelInfo, "Skipping push of image "+target)
		return nil
	}

	// Push the image
	if err := sh.Run(
		"docker",
		"push",
		fmt.Sprintf("%s/%s:%s", containerRegistry(), target, tag),
	); err != nil {
		return fmt.Errorf("pushing image %s: %w", target, err)
	}

	// Push the image with the Latest tag
	if err := sh.Run(
		"docker",
		"push",
		fmt.Sprintf("%s/%s:latest", containerRegistry(), target),
	); err != nil {
		return fmt.Errorf("pushing image %s with latest tag: %w", target, err)
	}

	return nil
}
