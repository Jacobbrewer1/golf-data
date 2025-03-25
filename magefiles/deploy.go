//go:build mage

package magefiles

import (
	"fmt"
	"os"
	"sync"

	"github.com/jacobbrewer1/utils"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Deploy mg.Namespace

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
		for _, err := range multiErr.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	return nil
}

// Deploy a single app
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

	tag, err := getTagFromCommit()
	if err != nil {
		return fmt.Errorf("failed to get tag from commit: %w", err)
	}

	fmt.Println("Deploying", appName, "with tag", tag)

	args = append(args, "image.tag="+tag)

	if err := sh.Run("helm", args...); err != nil {
		return fmt.Errorf("failed to deploy: %w", err)
	}

	return nil
}
