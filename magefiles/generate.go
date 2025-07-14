//go:build mage

package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	modelFileSuffix = ".xo.go"
)

func GenerateAll() {
	mg.Deps(Init)
	mg.Deps(mg.F(Generate.mock, false))
	mg.Deps(mg.F(Generate.openAPI, false))
	mg.Deps(mg.F(Generate.models, false))
}

type Generate mg.Namespace

// Diff returns an error if the generated code is not up to date.
func (g Generate) Diff() error {
	mg.Deps(GenerateAll)

	// Check for any changes or new/removed files
	got, err := sh.Output("git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to get git status: %w", err)
	}

	if got != "" {
		log(slog.LevelError, "There are uncommitted changes in the generated code:")
		fmt.Println(got)
		return fmt.Errorf("there are uncommitted changes, please run 'mage generateall' and commit the changes")
	}

	return nil
}

// Mock generates mock files for the application.
func (Generate) Mock() error {
	mg.Deps(mg.F(Generate.mock, true))

	// Cover any new dependencies that may have been added by the code generation.
	return VendorDeps()
}

// mock generates mock files for the application.
func (Generate) mock(shouldVendor bool) error {
	mg.Deps(Init)

	args := []string{
		"run",
		"//:gen_mock",
	}

	if err := sh.Run("bazel", args...); err != nil {
		return fmt.Errorf("error writing generated mock files: %w", err)
	}

	if shouldVendor {
		if err := VendorDeps(); err != nil {
			return err
		}
	}

	return nil
}

// Models generates model files for the application.
func (g Generate) Models() error {
	if err := g.models(false); err != nil {
		return fmt.Errorf("error generating models: %w", err)
	}

	return VendorDeps()
}

// models generates model files for the application.
func (g Generate) models(shouldVendor bool) error {
	mg.Deps(Init)

	if err := g.removeExistingModels(); err != nil {
		return fmt.Errorf("error removing existing models: %w", err)
	}

	if err := sh.Run("go", "generate", "pkg/models"); err != nil {
		return fmt.Errorf("error generating models: %w", err)
	}

	if err := g.formatModels(); err != nil {
		return fmt.Errorf("error formatting models: %w", err)
	}

	if shouldVendor {
		if err := VendorDeps(); err != nil {
			return fmt.Errorf("error vendoring dependencies after model generation: %w", err)
		}
	}

	return nil
}

// removeExistingModels removes existing model files from the pkg/models directory.
func (Generate) removeExistingModels() error {
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

// formatModels formats the model files in the pkg/models directory.
func (Generate) formatModels() error {
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

func (Generate) OpenAPI() error {
	mg.Deps(mg.F(Generate.openAPI, true))

	// Cover any new dependencies that may have been added by the code generation.
	return VendorDeps()
}

func (Generate) openAPI(shouldVendor bool) error {
	mg.Deps(Init)

	args := []string{
		"run",
		"//:gen_openapi",
	}

	if err := sh.Run("bazel", args...); err != nil {
		return fmt.Errorf("error writing generated open-api files: %w", err)
	}

	if shouldVendor {
		return VendorDeps()
	}

	return nil
}
