package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExplainOverloadValueChecks(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, backend := range []string{"c", "js", "build"} {
		args := []string{backend, "internal/oracle/testdata/overload_values/narrow-valid.ts", "--explain-checks"}
		if backend == "build" {
			args = append(args, "-o", filepath.Join(t.TempDir(), "program"), "--sanitize")
		}
		command := exec.Command(os.Args[0], append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, args...)...)
		command.Dir = root
		command.Env = append(os.Environ(), "ADAMIC_EXPLAIN_TEST_DRIVER=1")
		var output, diagnostic bytes.Buffer
		command.Stdout = &output
		command.Stderr = &diagnostic
		if err := command.Run(); err != nil {
			t.Fatalf("%s: %v %s", backend, err, diagnostic.String())
		}
		text := diagnostic.String()
		if strings.Count(text, "result.value: checked;") != 2 || !strings.Contains(text, "predicate checks: proven 0 checked 2 unobservable 0") {
			t.Fatalf("%s: %s", backend, text)
		}
	}
}
