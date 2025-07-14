//go:build mage

package main

import (
	"os"
	"strconv"
	"sync"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// isCIRunner determines if the current process is running in a Continuous Integration (CI) environment.
// It checks the values of the "CI" and "GITHUB_ACTIONS" environment variables.
//
// GitHub Documentation: https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#default-environment-variables
var isCIRunner = sync.OnceValue(func() bool {
	// Check if running in CI
	ciRunner, _ := strconv.ParseBool(os.Getenv("CI"))
	githubRunner, _ := strconv.ParseBool(os.Getenv("GITHUB_ACTIONS"))
	return ciRunner || githubRunner
})

// isDebugMode checks debug mode is enabled
var isDebugMode = sync.OnceValue(func() bool {
	gotStr := os.Getenv("RUNNER_DEBUG")
	got, _ := strconv.ParseBool(gotStr)

	// Enable debug logging if the RUNNER_DEBUG environment variable is set to true or if not running in a CI environment.
	return got || !isCIRunner()
})

// hostArch returns the host CPU architecture. Use this instead of runtime.GOARCH as the latter succumbs to cross-compilation, so can lie.
var hostArch = sync.OnceValue(func() string {
	out, err := sh.Output("uname", "-m")
	if err != nil {
		panic(err)
	}

	switch out {
	case "arm64":
		return "arm64"
	case "x86_64":
		return "amd64"
	default:
		panic("unsupported host arch: " + out)
	}
})

// Init initializes the mage environment
func Init() error {
	// This is a workaround to prevent overriding local binaries
	if !isCIRunner() {
		return nil
	}

	mg.Deps(
		mg.F(Dep.Install, "github.com/bazelbuild/bazelisk@latest"),
	)
	return nil
}
