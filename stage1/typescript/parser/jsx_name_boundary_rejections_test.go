package parser

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func jsxFailure(test *testing.T, name string, args ...string) (int, string) {
	test.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	stdout, temporaryFileError := os.CreateTemp(test.TempDir(), "stdout-")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	defer stdout.Close()
	var stderr strings.Builder
	command.Stdout = stdout
	command.Stderr = &stderr
	runError := command.Run()
	exit, ok := runError.(*exec.ExitError)
	if !ok {
		test.Fatalf("expected parser rejection from %s: %v %s", name, runError, &stderr)
	}
	return exit.ExitCode(), stderr.String()
}

func TestJsxNameBoundaryRejections(test *testing.T) {
	test.Parallel()
	oracle := goOracle(test)
	directory, _ := filepath.Abs(".")
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	binary := buildPort(test, directory, true)
	cases := []struct{ name, source, marker string }{
		{"private member", `const x = <Foo.#bar/>;`, "JSX expected name"},
		{"escaped tag", `const x = <Fo\u006f/>;`, "JSX Unicode escape not allowed"},
		{"escaped member", `const x = <Foo.b\u0061r/>;`, "JSX Unicode escape not allowed"},
		{"escaped namespace", `const x = <ns:b\u{61}r/>;`, "JSX Unicode escape not allowed"},
		{"escaped attribute", `const x = <Foo b\u0061r/>;`, "JSX Unicode escape not allowed"},
	}
	paths := map[string]string{}
	for _, item := range cases {
		path := filepath.Join(test.TempDir(), "name.tsx")
		paths[item.name] = path
		if writeError := os.WriteFile(path, []byte(item.source), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		code, diagnostics := jsxFailure(test, oracle, path, "--whole")
		if code != 1 || !strings.Contains(diagnostics, "parser diagnostic") {
			test.Fatalf("Go accepted %s: %s", item.name, diagnostics)
		}
		code, nodeError := jsxFailure(
			test,
			"node",
			"--disable-warning=ExperimentalWarning",
			runner,
			filepath.Join(directory, "main.ts"),
			path,
			"--whole",
		)
		if code != 70 || !strings.Contains(nodeError, item.marker) {
			test.Fatalf("Node accepted %s: %d %s", item.name, code, nodeError)
		}
		code, nativeError := jsxFailure(test, binary, path, "--whole")
		if code != 70 || nativeError != nodeError {
			test.Fatalf("native refusal differs for %s", item.name)
		}
		test.Logf("%s: Go diagnoses; Node/native refuse identically", item.name)
	}
	for _, change := range jsxNameBoundaryChanges() {
		test.Run(change.name, func(test *testing.T) {
			mutant := copyPort(test, "jsx.ts", change.from, change.to)
			mutantBinary := buildPort(test, mutant, true)
			path := paths[change.control]
			for _, side := range []struct {
				name string
				run  execution
			}{
				{"Node", execute(
					test,
					"",
					"node",
					"--disable-warning=ExperimentalWarning",
					runner,
					filepath.Join(mutant, "main.ts"),
					path,
					"--whole",
				)},
				{"native", execute(test, "", mutantBinary, path, "--whole")},
			} {
				if !strings.Contains(string(side.run.output), "JsxSelfClosingElement") {
					test.Fatal("permissive mutant did not parse its input")
				}
				test.Logf("%s compiled permissive mutant exits 0; Go-backed refusal check catches it", side.name)
			}
		})
	}
}
