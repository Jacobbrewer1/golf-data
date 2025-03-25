//go:build mage

package magefiles

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/jacobbrewer1/utils"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	imageAppSeparator = "/"
)

var (
	dockerRegistry = os.Getenv("DOCKER_REGISTRY")
	envTags        = os.Getenv("TAGS")
	toPushEnv      = os.Getenv("DOCKER_PUSH")
	toPush         = false
)

type Image mg.Namespace

// A build step that requires additional params, or platform specific steps for example
func (i Image) All() error {
	wg := new(sync.WaitGroup)
	multiErr := utils.NewMultiError()

	// Get all directory names in the ./cmd directory
	cmds, err := os.ReadDir("./cmd")
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Iterate over each directory
	for _, cmd := range cmds {
		if cmd.IsDir() {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()

				if err := i.One(name); err != nil {
					multiErr.Add(err)
				}
			}(cmd.Name())
		}
	}

	wg.Wait()

	if multiErr.Err() != nil {
		for _, err := range multiErr.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	return nil
}

func (i Image) One(appName string) error {
	toPush, _ = strconv.ParseBool(toPushEnv)
	fmt.Println("Push images set to: ", toPush)

	if err := i.buildImage(appName); err != nil {
		return fmt.Errorf("failed to build image for %s: %w", appName, err)
	}

	if toPush {
		if err := i.pushImage(appName); err != nil {
			return fmt.Errorf("failed to push image for %s: %w", appName, err)
		}
	}

	return nil
}

func (i Image) buildImage(appName string) error {
	applicationDockerRegistry := dockerRegistry + imageAppSeparator + appName
	fmt.Println(applicationDockerRegistry)

	tags := i.imageTags(applicationDockerRegistry)
	fmt.Println(tags)

	commitTag, err := getTagFromCommit()
	if err != nil {
		return fmt.Errorf("failed to get commit tag: %w", err)
	}

	commitTag = applicationDockerRegistry + ":" + commitTag

	if !slices.Contains(tags, commitTag) {
		fmt.Println("Adding commit tag to image tags: ", commitTag)
		tags = append(tags, commitTag)
	}

	cmd := exec.Command("docker", "build")

	for _, tag := range tags {
		cmd.Args = append(cmd.Args, "-t", tag)
	}

	cmd.Args = append(cmd.Args, ".")
	cmd.Args = append(cmd.Args, "--build-arg", "APP_NAME="+appName)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	return nil
}

func (i Image) pushImage(appName string) error {
	applicationDockerRegistry := dockerRegistry + imageAppSeparator + appName
	fmt.Println(applicationDockerRegistry)

	tags := i.imageTags(applicationDockerRegistry)
	fmt.Println(tags)
	for _, tag := range tags {
		if err := sh.Run("docker", "push", tag); err != nil {
			return fmt.Errorf("failed to push image: %w", err)
		}
	}

	return nil
}

func (i Image) imageTags(registry string) []string {
	envSplit := strings.Split(envTags, ",")
	tags := make([]string, 0)
	for _, tag := range envSplit {
		tags = append(tags, registry+":"+tag)
	}
	return tags
}

// Get the tag for the image based on the current commit.
func (i Image) CommitTag() error {
	tag, err := getTagFromCommit()
	if err != nil {
		return fmt.Errorf("failed to get commit tag: %w", err)
	}
	fmt.Println(tag)
	return nil
}

func getTagFromCommit() (string, error) {
	commit, err := sh.Output("git", "describe", "--tags")
	if err != nil {
		return "", fmt.Errorf("failed to get commit tag: %w", err)
	}
	return commit, nil
}
