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

func jsxFailure(t *testing.T, name string, args ...string) (int, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	stdout, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer stdout.Close()
	var stderr strings.Builder
	command.Stdout = stdout
	command.Stderr = &stderr
	err = command.Run()
	exit, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected parser rejection from %s: %v %s", name, err, &stderr)
	}
	return exit.ExitCode(), stderr.String()
}

// Not parallel: one control binary holds every name boundary before mutant builds.
func TestJsxNameBoundaryRejections(t *testing.T) {
	oracle := goOracle(t)
	directory, _ := filepath.Abs(".")
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	binary := buildPort(t, directory, true)
	cases := []struct{ name, source, marker string }{
		{"private member", `const x = <Foo.#bar/>;`, "JSX expected name"},
		{"escaped tag", `const x = <Fo\u006f/>;`, "JSX Unicode escape not allowed"},
		{"escaped member", `const x = <Foo.b\u0061r/>;`, "JSX Unicode escape not allowed"},
		{"escaped namespace", `const x = <ns:b\u{61}r/>;`, "JSX Unicode escape not allowed"},
		{"escaped attribute", `const x = <Foo b\u0061r/>;`, "JSX Unicode escape not allowed"},
	}
	paths := map[string]string{}
	for _, item := range cases {
		path := filepath.Join(t.TempDir(), "name.tsx")
		paths[item.name] = path
		if err := os.WriteFile(path, []byte(item.source), 0644); err != nil {
			t.Fatal(err)
		}
		code, diagnostics := jsxFailure(t, oracle, path, "--whole")
		if code != 1 || !strings.Contains(diagnostics, "parser diagnostic") {
			t.Fatalf("Go accepted %s: %s", item.name, diagnostics)
		}
		code, nodeError := jsxFailure(t, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), path, "--whole")
		if code != 70 || !strings.Contains(nodeError, item.marker) {
			t.Fatalf("Node accepted %s: %d %s", item.name, code, nodeError)
		}
		code, nativeError := jsxFailure(t, binary, path, "--whole")
		if code != 70 || nativeError != nodeError {
			t.Fatalf("native refusal differs for %s", item.name)
		}
		t.Logf("%s: Go diagnoses; Node/native refuse identically", item.name)
	}
	changes := []struct{ name, from, to, control string }{
		{"private name accepted", "if(this.parser.kind() !== 'Identifier' && !this.parser.kind().endsWith('Keyword'))", "if(this.parser.kind() !== 'Identifier' && !this.parser.kind().endsWith('Keyword') && this.parser.kind() !== 'PrivateIdentifier')", "private member"},
		{"escaped name accepted", "if((this.parser.scanner.flags & 1032) !== 0)", "if((this.parser.scanner.flags & 0) !== 0)", "escaped tag"},
	}
	for _, change := range changes {
		t.Run(change.name, func(t *testing.T) {
			mutant := copyPort(t, "jsx.ts", change.from, change.to)
			mutantBinary := buildPort(t, mutant, true)
			path := paths[change.control]
			for _, side := range []struct {
				name string
				run  execution
			}{
				{"Node", execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(mutant, "main.ts"), path, "--whole")},
				{"native", execute(t, "", mutantBinary, path, "--whole")},
			} {
				if !strings.Contains(string(side.run.output), "JsxSelfClosingElement") {
					t.Fatal("permissive mutant did not parse its input")
				}
				t.Logf("%s compiled permissive mutant exits 0; Go-backed refusal check catches it", side.name)
			}
		})
	}
}
