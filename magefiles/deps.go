//go:build mage

package magefiles

import (
	"fmt"

	"github.com/magefile/mage/sh"
)

func InstallDeps() error {
	fmt.Println("Installing Deps...")
	if err := sh.Run("go", "install", "github.com/jacobbrewer1/goschema@latest"); err != nil {
		return fmt.Errorf("failed to install goschema: %w", err)
	}

	if err := sh.Run("go", "install", "golang.org/x/tools/cmd/goimports@latest"); err != nil {
		return fmt.Errorf("failed to install goimports: %w", err)
	}

	return nil
}
