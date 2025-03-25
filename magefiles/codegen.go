//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	modelFileSuffix = ".xo.go"
)

type Codegen mg.Namespace

func (c Codegen) Generate() error {
	// Is goschema and goimports installed?
	if _, err := exec.LookPath("goschema"); err != nil {
		return fmt.Errorf("goschema is not installed: %w", err)
	} else if _, err := exec.LookPath("goimports"); err != nil {
		return fmt.Errorf("goimports is not installed: %w", err)
	}

	if err := sh.Run("go", "generate", "./..."); err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	if err := c.formatModels(); err != nil {
		return fmt.Errorf("failed to format models: %w", err)
	}

	return nil
}

func (c Codegen) removeExistingModels() error {
	fp := filepath.Join("pkg", "models")

	// Get files in the models directory
	files, err := os.ReadDir(fp)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Loop through the files
	for _, file := range files {
		// Check if the file is a go file
		if !strings.HasSuffix(file.Name(), modelFileSuffix) {
			continue
		}

		// Remove the file
		if err := os.Remove(filepath.Join(fp, file.Name())); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
	}

	return nil
}

func (c Codegen) formatModels() error {
	fp := filepath.Join("pkg", "models")

	// Get files in the models directory
	files, err := os.ReadDir(fp)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Loop through the files
	for _, file := range files {
		// Check if the file is a go file
		if !strings.HasSuffix(file.Name(), modelFileSuffix) {
			continue
		}

		// Format the file
		if err := sh.Run("goimports", "-w", filepath.Join(fp, file.Name())); err != nil {
			return fmt.Errorf("failed to format file: %w", err)
		}
	}

	return nil
}

func (c Codegen) Cicd() error {
	mg.Deps(c.Generate, c.removeExistingModels)

	// Check for any changes or new/removed files
	got, err := sh.Output("git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}

	if got != "" {
		fmt.Println("Found changes:\n")
		fmt.Println(got)
		return fmt.Errorf("there are uncommitted changes, please run 'mage codegen:generate' and commit the changes")
	}

	return nil
}
