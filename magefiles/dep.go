//go:build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Dep mg.Namespace

// Get retrieves a golang dependency and ensures the library is configured correctly in Bazel.
func (Dep) Get(dep string) error {
	err := sh.Run(
		"go",
		"get",
		dep,
	)
	if err != nil {
		return fmt.Errorf("error retrieving dependency: %w", err)
	}

	return VendorDeps()
}

// Install installs a golang dependency via go install.
func (Dep) Install(dep string) error {
	err := sh.Run(
		"go",
		"install",
		dep,
	)
	if err != nil {
		return fmt.Errorf("error installing dependency: %w", err)
	}

	return nil
}

// VendorDeps manages vendoring of Golang dependencies.
func VendorDeps() error {
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return err
	}

	if err := sh.Run("go", "mod", "verify"); err != nil {
		return err
	}

	if err := sh.Run("go", "mod", "vendor"); err != nil {
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
