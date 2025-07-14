//go:build mage

package main

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Test mg.Namespace

// Unit runs unit tests for the repository.
func (Test) Unit() error {
	mg.Deps(Init)
	log(slog.LevelInfo, "Running unit tests")

	args := []string{
		"test",
	}

	if isCIRunner() {
		args = appendBazelArgs(args)
	}
	args = append(args, "--test_arg=--test.shuffle=on")

	args = append(args, "//...")

	err := sh.Run("bazel", args...)
	if err != nil {
		return err
	}

	return nil
}

// Bench runs benchmarks for the repository.
func (Test) Bench() error {
	mg.Deps(Init)
	got, err := benchmark()
	if err != nil {
		return fmt.Errorf("error running benchmarks: %w", err)
	}

	fmt.Println(got)
	return nil
}

// benchmark executes the benchmarks for the repository and returns the output.
func benchmark() (string, error) {
	mg.Deps(Init)
	log(slog.LevelInfo, "Running benchmarks")

	args := []string{
		"test",
		"--test_arg=-bench=.",
		"--test_arg=-run=^$",
	}

	args = append(args, "//...")

	got, err := sh.Output("bazel", args...)
	if err != nil {
		return "", fmt.Errorf("error running benchmarks: %w", err)
	}

	return got, nil
}

type Coverage mg.Namespace

// Run runs unit tests for the repository with code coverage enabled.
func (Coverage) Run() error {
	mg.Deps(Init)
	log(slog.LevelInfo, "Running unit tests with coverage")

	args := []string{
		"coverage",
		"//...",
	}

	err := sh.Run("bazel", args...)
	if err != nil {
		return fmt.Errorf("error running unit tests: %w", err)
	}

	return nil
}

// View views the code coverage for this project.
// Information on coverage can be viewed with the lcov genhtml tool (i.e. `brew install lcov`).
func (Coverage) View() error {
	mg.Deps(Coverage.Run)

	// Source: https://bazel.build/configure/coverage#viewing_coverage
	err := sh.Run(
		"genhtml",
		"--branch-coverage",
		"--output",
		"genhtml",
		filepath.Join(outputDirectory(), "_coverage", "_coverage_report.dat"),
	)
	if err != nil {
		return fmt.Errorf("unable to generate HTML coverage report: %w", err)
	}

	return sh.Run("open", filepath.Join("genhtml", "index.html"))
}
