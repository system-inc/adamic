package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var recoveryCases = []string{
	"interface I",
	"interface I { m(a: string): void;",
	"interface I { m<(a: string): void; }",
	"interface I { m<T(a: T): T; }",
}

// Each process has its own deadline. Inputs and answers survive a failing test.
func recoveryRun(t *testing.T, artifact, name string, args ...string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	output, err := os.Create(artifact + ".stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	stderr, err := os.Create(artifact + ".stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()
	command.Stdout, command.Stderr = output, stderr
	err = command.Run()
	if ctx.Err() != nil {
		err = fmt.Errorf("timeout after 2s: %w", ctx.Err())
	}
	data, readErr := os.ReadFile(output.Name())
	if readErr != nil {
		t.Fatal(readErr)
	}
	if info, statErr := stderr.Stat(); statErr != nil {
		t.Fatal(statErr)
	} else if info.Size() != 0 && err == nil {
		err = fmt.Errorf("unexpected stderr: %s", stderr.Name())
	}
	return data, err
}

func TestMethodRecoveryAgrees(t *testing.T) {
	oracle := goOracle(t)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	binary := buildPort(t, directory, true)
	artifacts := os.Getenv("ADAMIC_RECOVERY_ARTIFACTS")
	if artifacts == "" {
		artifacts = "/tmp/adamic-parser-recovery"
	}
	if err := os.MkdirAll(artifacts, 0755); err != nil {
		t.Fatal(err)
	}
	cases := append([]string(nil), recoveryCases...)
	cases = append(cases, "function", "function () {}", "export function function f() {}", "export default function () {}", "const {a b} = obj;", "const {a,,b} = obj;", "const [a b] = obj;", "type T = A | | B;", "type T = A & & B;", "type T = | A;", "function f(x, { return g(); }", "const { return x } = obj;", "x[];", "x?.[];", "f()[ ];", "import { A B } from \"m\";", "import { A,, B } from \"m\";", "export { A B };", "import { A from \"m\";", "function f(x:: T | undefined, y: U) {}", "function f(| T) {}", "function f(1) {}", "`unterminated", "const s = \"unterminated", "/* unterminated", "const r = /unterminated", "const f = ( => g();", "=> x", "const x = { a.b };", "module`M` {}", "tag`x` x;", "const x = { a: 1; };", "const x = { a: 1 b: 2 };", "type T = Omit<A \"b\"> & { b: string | false | undefined; };", "type T = A[", "function f(x: A[", "type T = A[;", "* from \"../moduleSpecifiers.js\";", "export export * from \"../moduleSpecifiers.js\";", "export from \"../moduleSpecifiers.js\";", "import import * as p from \"p.js\";", "import", "import\r\n  ", "}", ") ;", "x ?", "x ? y", "x ? y :", "x?.", "`x${", "`x${y", "type T = `x${", "interface I { : I | undefined; name: string; }", "interface I { x: T y: U }", "f(x x);", "f(, x);", "function f(){ g(x }", "function f(){ export }", "default x;", "switch(x){case 1:: f();}", "switch(x){default:: f();}", "switch(x){ f(); }", "namespace N { : f(); }", "function f(x: T y: U) {}", "function f(, x: T) {}", "function f(return) {}", "function f(const) {}", "function f(this: T) {}", "function f(x:: T) {}", "function f() x;", "break break;", "continue continue;", "break 1;", "throw", "throw\nx;", "throw x y;", "setParent setParent(x, y);", "retrun 1;", "Set 1;", "globl 1;", "İmport 1;", "export {", "import {", "f(", "[", "({", "function f() {", "class C {", "type T =", "type T = A<", "type T = (x:", "type T = ()", "type T = <", "function f(): (x: T, y:", "const {", "const const x = 1;", "const x = 1 y = 2;", "const , x = 1;", "const", "function f() { const", "const f = (x:", "const f = (...", "const f = ():", "const f = () {}", "export *", "export * from", "export * from\r\n  ", "export default", "export =", "export", "export\r\n  ", "export export", "export\nconst x = 1;", "export x;", "/* π 💡 */\r\nexport")
	for _, source := range recoveryCases {
		cases = append(cases, "/* π 💡 */\r\n"+source, source+"\r\n  ")
	}
	for i, source := range cases {
		path := filepath.Join(artifacts, fmt.Sprintf("method-%d.ts", i))
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		want, err := recoveryRun(t, path+".go", oracle, path, "--whole", "--recovery")
		if err != nil {
			t.Fatal(err)
		}
		for _, side := range []struct {
			name, command string
			args          []string
		}{
			{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), path, "--whole", "--recovery"}},
			{"native", binary, []string{path, "--whole", "--recovery"}},
		} {
			got, err := recoveryRun(t, path+"."+side.name, side.command, side.args...)
			if err != nil {
				t.Errorf("%s %s: %v (input and output saved)", side.name, path, err)
				continue
			}
			if !bytes.Equal(got, want) {
				t.Errorf("%s %s: %s", side.name, path, difference(got, want))
			}
		}
	}
}

// The three checks are independent: diagnostic bytes, recovered child shape,
// and termination. A build error or crash does not kill a comparison mutant.
func TestRecoveryMutants(t *testing.T) {
	oracle := goOracle(t)
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	artifacts := "/tmp/adamic-parser-recovery-mutants"
	if err := os.MkdirAll(artifacts, 0755); err != nil {
		t.Fatal(err)
	}
	mutations := []struct {
		name, from, to, source string
		timeout                bool
	}{
		{"diagnostic-code", "this.error(1005, `'${tokenSpelling(kind)}' expected.`);\n            return false;", "this.error(1006, `'${tokenSpelling(kind)}' expected.`);\n            return false;", recoveryCases[0], false},
		{"stranded-export", "this.error(diagnostic.code, diagnostic.message);", "this.error(diagnostic.code + (context === 'source' ? 1 : 0), diagnostic.message);", "export", false},
		{"generic-child", "result.push(this.make('TypeParameter', pos, children));", "result.push(this.make('TypeParameter', pos, children.slice(0, 0)));", recoveryCases[3], false},
		{"eof-loop", "return this.scanner.kind;", "return this.scanner.kind === 'EndOfFile' ? 'Identifier' : this.scanner.kind;", recoveryCases[1], true},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			path := filepath.Join(artifacts, mutation.name+".ts")
			if err := os.WriteFile(path, []byte(mutation.source), 0644); err != nil {
				t.Fatal(err)
			}
			want, err := recoveryRun(t, path+".go", oracle, path, "--whole", "--recovery")
			if err != nil {
				t.Fatal(err)
			}
			mutant := copyPort(t, "parser.ts", mutation.from, mutation.to)
			binary := buildPort(t, mutant, true)
			for _, side := range []struct {
				name, command string
				args          []string
			}{
				{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(mutant, "main.ts"), path, "--whole", "--recovery"}},
				{"native", binary, []string{path, "--whole", "--recovery"}},
			} {
				got, err := recoveryRun(t, path+"."+side.name, side.command, side.args...)
				if mutation.timeout {
					if err == nil || !strings.Contains(err.Error(), "timeout after 2s") {
						t.Fatalf("%s must be caught by the deadline, got %v", side.name, err)
					}
					t.Logf("%s: EOF loop caught by 2s deadline; input %s", side.name, path)
				} else {
					if err != nil {
						t.Fatalf("%s mutant must finish normally: %v", side.name, err)
					}
					if bytes.Equal(got, want) {
						t.Fatalf("%s mutant survived", side.name)
					}
					t.Logf("%s: %s", side.name, difference(got, want))
				}
			}
		})
	}
}

func TestRecoveredLintCasesAgree(t *testing.T) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("../../cohere/lint/testdata/oracle.go")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(root, "adamic_lint_recovery_oracle.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(t.TempDir(), "lint-oracle")
	execute(t, root, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	directory, err := filepath.Abs("../../cohere/lint")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	binary := buildPort(t, directory, true)
	artifacts := "/tmp/adamic-parser-recovery-lint"
	if err := os.MkdirAll(artifacts, 0755); err != nil {
		t.Fatal(err)
	}
	for i, source := range recoveryCases {
		path := filepath.Join(artifacts, fmt.Sprintf("method-%d.ts", i))
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		for _, style := range []string{"method", "property"} {
			manifest := path + "." + style + ".manifest"
			row := path + "\t@typescript-eslint/method-signature-style\t\t\tfalse\t{\"Style\":\"" + style + "\"}\trecovery\n"
			if err := os.WriteFile(manifest, []byte(row), 0644); err != nil {
				t.Fatal(err)
			}
			want, err := recoveryRun(t, manifest+".go", oracle, "--manifest", manifest)
			if err != nil {
				t.Fatal(err)
			}
			for _, side := range []struct {
				name, command string
				args          []string
			}{
				{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", manifest}},
				{"native", binary, []string{"--manifest", manifest}},
			} {
				got, err := recoveryRun(t, manifest+"."+side.name, side.command, side.args...)
				if err != nil {
					t.Fatalf("%s case %d %s: %v", side.name, i, style, err)
				}
				if !bytes.Equal(got, want) {
					t.Fatalf("%s case %d %s: %s", side.name, i, style, difference(got, want))
				}
			}
		}
	}
	t.Log("all eight original source/style combinations match real cohere findings and proposed repairs")
}
