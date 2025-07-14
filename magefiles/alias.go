//go:build mage

package main

var Aliases = map[string]interface{}{
	"fixit":    VendorDeps,
	"build":    Build.All,
	"test":     Test.Unit,
	"install":  Dep.Install,
	"generate": GenerateAll,
}
