package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
			path := "internal/lower/testdata/predicates/" + fixture + ".a"
			for _, backend := range []string{"c", "js", "build"} {
				t.Run(backend, func(t *testing.T) {
					t.Parallel()
					arguments := []string{backend, path, "--explain-checks"}
					if backend == "build" {
						arguments = append(arguments, "-o", filepath.Join(t.TempDir(), "program"), "--sanitize")
					}
					command := exec.Command(os.Args[0], append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, arguments...)...)
					command.Dir = repository
					command.Env = append(os.Environ(), "ADAMIC_EXPLAIN_TEST_DRIVER=1", "XDG_CACHE_HOME="+t.TempDir())
					var output, diagnostic bytes.Buffer
					command.Stdout, command.Stderr = &output, &diagnostic
					if err := command.Run(); err != nil {
						t.Fatalf("%s: %v, %s", backend, err, diagnostic.Bytes())
					}
					if !bytes.Equal(diagnostic.Bytes(), want) {
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

// Placeholder checks share the public report in every compiler entry point.
func TestPlaceholderExplainChecksOutput(t *testing.T) {
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, backend := range []string{"c", "js", "build"} {
		t.Run(backend, func(t *testing.T) {
			arguments := []string{backend, "internal/oracle/testdata/placeholder_nonnull_saved_leak.a", "--explain-checks"}
			if backend == "build" {
				arguments = append(arguments, "-o", filepath.Join(t.TempDir(), "program"), "--sanitize")
			}
			command := exec.Command(os.Args[0], append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, arguments...)...)
			command.Dir = repository
			command.Env = append(os.Environ(), "ADAMIC_EXPLAIN_TEST_DRIVER=1")
			var output, diagnostic bytes.Buffer
			command.Stdout, command.Stderr = &output, &diagnostic
			if err := command.Run(); err != nil {
				t.Fatalf("%v: %s", err, diagnostic.Bytes())
			}
			for _, want := range []string{"placeholder value: checked at argument via saved", "adamic: placeholder checks: proven 0 checked 1"} {
				if !strings.Contains(diagnostic.String(), want) {
					t.Fatalf("missing %q: %s", want, diagnostic.Bytes())
				}
			}
		})
	}
}
