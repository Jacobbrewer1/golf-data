//go:build mage

package main

var Aliases = map[string]interface{}{
	"fixit":  VendorDeps,
	"deps":   InstallDeps,
	"binary": FindBinary,
	"build":  Build.All,
}
