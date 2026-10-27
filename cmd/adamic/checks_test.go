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
func TestExplainChecksOutput(t *testing.T) {
	t.Parallel()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []string{"overload_some_empty", "overload_some_false_read", "overload_some_true_only", "overload_erased", "overload_assertion"} {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			want, err := os.ReadFile("testdata/checks/" + fixture + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			witness := "internal/lower/testdata/predicates/" + fixture + ".a"
			source, err := os.ReadFile(filepath.Join(repository, witness))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "main.ts")
			if err = os.WriteFile(path, source, 0644); err != nil {
				t.Fatal(err)
			}
			for _, backend := range []string{"c", "js", "build"} {
				t.Run(backend, func(t *testing.T) {
					t.Parallel()
					arguments := []string{backend, path, "--explain-checks"}
					if backend == "build" {
						arguments = append(arguments, "-o", filepath.Join(t.TempDir(), "program"), "--sanitize")
					}
					command := exec.Command(os.Args[0], append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, arguments...)...)
					command.Dir = repository
					command.Env = explainDriverEnvironment(t, os.Environ())
					var output, diagnostic bytes.Buffer
					command.Stdout, command.Stderr = &output, &diagnostic
					if err := command.Run(); err != nil {
						t.Fatalf("%s: %v, %s", backend, err, diagnostic.Bytes())
					}
					if diagnosticString := strings.ReplaceAll(diagnostic.String(), path, witness); diagnosticString != string(want) {
						t.Fatalf("%s explanation:\n%s\nwant:\n%s", backend, diagnostic.Bytes(), want)
					}
					if backend != "build" && output.Len() == 0 {
						t.Fatalf("%s lost generated source", backend)
					}
					if backend == "build" && output.Len() != 0 {
						t.Fatalf("build wrote report to stdout: %s", output.Bytes())
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
func explainDriverEnvironment(t *testing.T, environment []string) []string {
	t.Helper()
	cache, err := location.BuildCache()
	if err != nil {
		t.Fatal(err)
	}
	return append(slices.Clone(environment), "ADAMIC_EXPLAIN_TEST_DRIVER=1", "ADAMIC_BUILD_CACHE_DIR="+cache)
}

// A driver reads the build cache this process does, whatever the box sets. Mutant: the fresh XDG_CACHE_HOME back in
// place of the pin, and a box without ADAMIC_BUILD_CACHE_DIR has the driver look in an empty cache.
func TestExplainDriversReadTheTestsBuildCache(t *testing.T) {
	t.Parallel()
	cache, err := location.BuildCache()
	if err != nil {
		t.Fatal(err)
	}
	environment := explainDriverEnvironment(t, []string{"XDG_CACHE_HOME=" + t.TempDir(), "ADAMIC_BUILD_CACHE_DIR="})
	pinned := ""
	for _, entry := range environment {
		if value, found := strings.CutPrefix(entry, "ADAMIC_BUILD_CACHE_DIR="); found {
			// The last setting of a name is the one a child process sees.
			pinned = value
		}
	}
	if pinned != cache {
		t.Fatalf("a driver's ADAMIC_BUILD_CACHE_DIR is %q, want this process's build cache %q: %q", pinned, cache, environment)
	}
}

// Not parallel: run writes process-wide stdout/stderr and native.runtimeBuilds in the driver process.
func TestExplainChecksDriver(t *testing.T) {
	if os.Getenv("ADAMIC_EXPLAIN_TEST_DRIVER") != "1" {
		return
	}
	for index, argument := range os.Args {
		if argument == "--" {
			os.Exit(run(os.Args[index+1:]))
		}
	}
	t.Fatal("missing driver arguments")
}
