//go:build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/sh"
)

// VendorDeps manages vendoring of Golang dependencies.
func VendorDeps() error {
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return err
	}

	if err := sh.Run("go", "mod", "vendor"); err != nil {
		return err
	}

	if err := sh.Run("go", "mod", "verify"); err != nil {
		return err
	}

	if err := sh.Run("bazel", "mod", "tidy"); err != nil {
		return err
	}

	if err := sh.Run("bazel", "run", "//:gazelle"); err != nil {
		return err
	}

	return nil
}

// InstallDeps installs the required dependencies for the project.
func InstallDeps() error {
	fmt.Println("[DEBUG] Installing Deps...")
	if err := sh.Run("go", "install", "github.com/jacobbrewer1/goschema@latest"); err != nil {
		return fmt.Errorf("failed to install goschema: %w", err)
	}

	if err := sh.Run("go", "install", "golang.org/x/tools/cmd/goimports@latest"); err != nil {
		return fmt.Errorf("failed to install goimports: %w", err)
	}

	if err := sh.Run("go", "install", "github.com/bazelbuild/bazelisk@latest"); err != nil {
		return fmt.Errorf("failed to install bazelisk: %w", err)
	}

	return nil
}
