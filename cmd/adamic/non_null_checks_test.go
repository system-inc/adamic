package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestNonNullExplainChecks(t *testing.T) {
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ name, position, status, counts string }{
		{"narrowed", "2:43", "proven", "proven 1 checked 0"},
		{"present", "1:69", "checked", "proven 0 checked 1"},
		{"undefined", "3:24", "checked", "proven 0 checked 1"},
		{"null", "3:24", "checked", "proven 0 checked 1"},
		{"initializer", "3:26", "checked", "proven 0 checked 1"},
		{"let_initializer", "3:26", "checked", "proven 0 checked 1"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := "internal/oracle/testdata/non_null_checked_" + probe.name + ".ts"
			want := "adamic: predicate checks: proven 0 checked 0 unobservable 0\n" + path + ":" + probe.position + ": non-null assertion value!: " + probe.status + "\nadamic: non-null checks: " + probe.counts + "\nadamic: checks: " + probe.counts + " unobservable 0\n"
			for _, backend := range []string{"c", "js", "build"} {
				args := []string{backend, path, "--explain-checks"}
				if backend == "build" {
					args = append(args, "-o", filepath.Join(t.TempDir(), "program"), "--sanitize")
				}
				command := exec.Command(os.Args[0], append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, args...)...)
				command.Dir = repository
				command.Env = append(os.Environ(), "ADAMIC_EXPLAIN_TEST_DRIVER=1")
				var output, diagnostic bytes.Buffer
				command.Stdout = &output
				command.Stderr = &diagnostic
				if err := command.Run(); err != nil {
					t.Fatalf("%s: %v %s", backend, err, diagnostic.Bytes())
				}
				if diagnostic.String() != want {
					t.Fatalf("%s explanation:\n%s\nwant:\n%s", backend, diagnostic.Bytes(), want)
				}
			}
		})
	}
}
