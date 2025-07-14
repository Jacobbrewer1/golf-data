//go:build mage

package main

import (
	"sync"

	"github.com/magefile/mage/sh"
)

// binDirectory returns the absolute path to the root bazel-bin directory.
var binDirectory = sync.OnceValue(func() string {
	dir, err := sh.Output("bazel", "info", "bazel-bin")
	if err != nil {
		panic(err)
	}

	return dir
})

// outputDirectory returns the absolute path to the root output path.
var outputDirectory = sync.OnceValue(func() string {
	dir, err := sh.Output("bazel", "info", "output_path")
	if err != nil {
		panic(err)
	}

	return dir
})

// appendBazelArgs appends additional Bazel arguments to the provided slice of arguments.
func appendBazelArgs(args []string) []string {
	args = append(args, "--@io_bazel_rules_go//go/config:race")
	return args
}
