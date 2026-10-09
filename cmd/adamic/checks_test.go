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
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []string{"overload_some_empty", "overload_some_false_read", "overload_some_true_only", "overload_erased", "overload_assertion"} {
		t.Run(fixture, func(t *testing.T) {
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
				arguments := []string{backend, path, "--explain-checks"}
				if backend == "build" {
					arguments = append(arguments, "-o", filepath.Join(t.TempDir(), "program"), "--sanitize")
				}
				command := exec.Command(os.Args[0], append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, arguments...)...)
				command.Dir = repository
				command.Env = append(os.Environ(), "ADAMIC_EXPLAIN_TEST_DRIVER=1")
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
			}
		})
	}
}

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
