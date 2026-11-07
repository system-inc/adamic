package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache/location"
)

// Each subprocess owns its stdout/stderr; other driver tests capture those globals.
func TestExplainChecksOutput(test *testing.T) {
	test.Parallel()
	repository, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, fixture := range []string{
		"overload_some_empty",
		"overload_some_false_read",
		"overload_some_true_only",
		"overload_erased",
		"overload_assertion",
	} {
		test.Run(fixture, func(test *testing.T) {
			test.Parallel()
			want, readError := os.ReadFile("testdata/checks/" + fixture + ".txt")
			if readError != nil {
				test.Fatal(readError)
			}
			witness := "internal/lower/testdata/predicates/" + fixture + ".a"
			source, readError := os.ReadFile(filepath.Join(repository, witness))
			if readError != nil {
				test.Fatal(readError)
			}
			path := filepath.Join(test.TempDir(), "main.ts")
			if writeError := os.WriteFile(path, source, 0644); writeError != nil {
				test.Fatal(writeError)
			}
			for _, backend := range []string{"c", "js", "build"} {
				test.Run(backend, func(test *testing.T) {
					test.Parallel()
					arguments := []string{backend, path, "--explain-checks"}
					if backend == "build" {
						arguments = append(arguments, "-o", filepath.Join(test.TempDir(), "program"), "--sanitize")
					}
					command := exec.Command(
						os.Args[0],
						append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, arguments...)...,
					)
					command.Dir = repository
					command.Env = explainDriverEnvironment(test, os.Environ())
					var output, diagnostic bytes.Buffer
					command.Stdout, command.Stderr = &output, &diagnostic
					if runError := command.Run(); runError != nil {
						test.Fatalf("%s: %v, %s", backend, runError, diagnostic.Bytes())
					}
					if diagnosticString := strings.ReplaceAll(
						diagnostic.String(),
						path,
						witness,
					); diagnosticString != string(want) {
						test.Fatalf("%s explanation:\n%s\nwant:\n%s", backend, diagnostic.Bytes(), want)
					}
					if backend != "build" && output.Len() == 0 {
						test.Fatalf("%s lost generated source", backend)
					}
					if backend == "build" && output.Len() != 0 {
						test.Fatalf("build wrote report to stdout: %s", output.Bytes())
					}
				})
			}
		})
	}
}

// explainDriverEnvironment is a driver's environment: the test's own, with the build cache pinned to the one this
// process reads, where the TestProduct_CommandRuntime tests put the runtime archives a driver's build is supplied.
// Without ADAMIC_BUILD_CACHE_DIR the cache is the user cache directory's, so a driver given a fresh XDG_CACHE_HOME
// found none of them and failed wherever Loom hadn't set the variable (#r9baa2d). A driver only reads the archives
// (commandRuntime never builds), so the drivers can share them.
func explainDriverEnvironment(test *testing.T, environment []string) []string {
	test.Helper()
	cache, buildCacheError := location.BuildCache()
	if buildCacheError != nil {
		test.Fatal(buildCacheError)
	}
	return append(slices.Clone(environment), "ADAMIC_EXPLAIN_TEST_DRIVER=1", "ADAMIC_BUILD_CACHE_DIR="+cache)
}

// A driver reads the build cache this process does, whatever the box sets. Mutant: the fresh XDG_CACHE_HOME back in
// place of the pin, and a box without ADAMIC_BUILD_CACHE_DIR has the driver look in an empty cache.
func TestExplainDriversReadTheTestsBuildCache(test *testing.T) {
	test.Parallel()
	cache, buildCacheError := location.BuildCache()
	if buildCacheError != nil {
		test.Fatal(buildCacheError)
	}
	environment := explainDriverEnvironment(
		test,
		[]string{"XDG_CACHE_HOME=" + test.TempDir(), "ADAMIC_BUILD_CACHE_DIR="},
	)
	pinned := ""
	for _, entry := range environment {
		if value, found := strings.CutPrefix(entry, "ADAMIC_BUILD_CACHE_DIR="); found {
			// The last setting of a name is the one a child process sees.
			pinned = value
		}
	}
	if pinned != cache {
		test.Fatalf(
			"a driver's ADAMIC_BUILD_CACHE_DIR is %q, want this process's build cache %q: %q",
			pinned,
			cache,
			environment,
		)
	}
}

// Not parallel: run writes process-wide stdout/stderr and native.runtimeBuilds in the driver process.
func TestExplainChecksDriver(test *testing.T) {
	if os.Getenv("ADAMIC_EXPLAIN_TEST_DRIVER") != "1" {
		return
	}
	for index, argument := range os.Args {
		if argument == "--" {
			os.Exit(run(os.Args[index+1:]))
		}
	}
	test.Fatal("missing driver arguments")
}

// Placeholder checks share the public report in every compiler entry point.
func TestPlaceholderExplainChecksOutputC(test *testing.T) {
	test.Parallel()
	assertPlaceholderExplainChecksOutput(test, "c")
}

func TestPlaceholderExplainChecksOutputJavaScript(test *testing.T) {
	test.Parallel()
	assertPlaceholderExplainChecksOutput(test, "js")
}

func TestPlaceholderExplainChecksOutputBuild(test *testing.T) {
	test.Parallel()
	assertPlaceholderExplainChecksOutput(test, "build")
}

func assertPlaceholderExplainChecksOutput(test *testing.T, backend string) {
	test.Helper()
	repository, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}

	arguments := []string{backend, "internal/oracle/testdata/placeholder_nonnull_saved_leak.a", "--explain-checks"}
	if backend == "build" {
		arguments = append(arguments, "-o", filepath.Join(test.TempDir(), "program"), "--sanitize")
	}
	command := exec.Command(os.Args[0], append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, arguments...)...)
	command.Dir = repository
	command.Env = append(os.Environ(), "ADAMIC_EXPLAIN_TEST_DRIVER=1")
	var output, diagnostic bytes.Buffer
	command.Stdout, command.Stderr = &output, &diagnostic
	if runError := command.Run(); runError != nil {
		test.Fatalf("%v: %s", runError, diagnostic.Bytes())
	}
	for _, want := range []string{
		"placeholder value: checked at argument via saved",
		"adamic: placeholder checks: proven 0 checked 1",
	} {
		if !strings.Contains(diagnostic.String(), want) {
			test.Fatalf("missing %q: %s", want, diagnostic.Bytes())
		}
	}
}
