//go:build mage

package main

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/jacobbrewer1/utils"
)

type Deploy mg.Namespace

// All deploys all apps in the ./cmd directory
func (d Deploy) All(environment string) error {
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

				if err := d.One(name, environment); err != nil {
					multiErr.Add(err)
				}
			}(cmd.Name())
		}
	}

	wg.Wait()

	if multiErr.Err() != nil {
		log(slog.LevelError, "Errors occurred during deployment:")
		for _, err := range multiErr.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	return nil
}

// One deploys a single app
func (d Deploy) One(appName, environment string) error {
	args := []string{
		"upgrade",
		"--install",
		appName,
		"./charts",
		"--values",
		"charts/valueFiles/" + environment + "/golf-data-" + appName + ".yaml",
		"--set",
	}

	log(slog.LevelInfo, fmt.Sprintf("Deploying %s to %s with tag %s", appName, environment, commitTag()))

	args = append(args, "image.tag="+commitTag())

	if err := sh.Run("helm", args...); err != nil {
		return fmt.Errorf("failed to deploy: %w", err)
	}

	return nil
}
