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

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

func recoveryInputs(t *testing.T) []string {
	t.Helper()
	var cases []string
	for _, name := range []string{"interface-eof.ts.txt", "method-eof.ts.txt", "generic-method-open.ts.txt", "generic-method-close.ts.txt"} {
		cases = append(cases, recoveryInput(t, name))
	}
	return cases
}

func recoveryInput(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "recovery", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// Each process has its own deadline. Inputs and answers survive a failing test.
func recoveryRun(t *testing.T, artifact, name string, args ...string) ([]byte, error) {
	t.Helper()
	return recoveryRunLimit(t, 2*time.Second, artifact, name, args...)
}

func recoveryRunLimit(t *testing.T, limit time.Duration, artifact, name string, args ...string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), limit)
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
		err = fmt.Errorf("timeout after %s: %w", limit, ctx.Err())
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
	recoveryCases := recoveryInputs(t)
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
	cases = append(cases, recoveryInput(t, "unterminated-string.ts.txt"), recoveryInput(t, "reserved-binding.ts.txt"))
	cases = append(cases, "(readonly T)", "(readonly T[])[];", "(public x);", "(readonly as U);", "async (readonly T)", "((x): number => x)(1);", "type T = ({x = 1}) => U;", "type T = ({x = 1});", "export function min< <T>(items: readonly [T, ...T[]], compare: Comparer<T>): T;", "const x = <T>(x: T): T;", "const x = <T>(x: T) => x;", "const x = { has( (element: T): boolean {} };", "const x = { values() ): T {} };", "const x = { a: 1; x.y };", "class C { m() x; }", "const x = { a: 1\n; };", "const x = { a: 1; if (x) {} };", "f(x :);", "const x = [a :];", "function f(export function g() {}", "function f(static x: T) {}", "function f(abstract x: T) {}", "function f(public @dec x: T) {}", "function f<T, U = T T>() {}", "function f<T T>() {}", "function f<T,, U>() {}", "function f<out>", "function f<const>", "function f<T extends>", "const a = [x x, y];", "const a = [x,,];", "const a = [;];", "for (const a = [x in y];;) {}", "try", "try {", "try {x;}", "try {} catch {}", "try {} finally", "try {} x;", "interface {}", "interface 1 {}", "namespace {}", "namespace 1 {}", "type = A;", "type 1 = A;", "is T;", "declare", "declare x;", "abstract", "undefined: x;", "type: x;", "let: x;", "async function f(){ await: x; }", "enum E { A B }", "enum E { A,, B }", "enum E", "enum E {", "let", "let;", "let = 1;", "let\nx;", "function f(){ let }", "export let", "namespace N", "namespace A.B", "namespace N;", "declare module \"m\";", "function", "function () {}", "export function function f() {}", "export default function () {}", "const {a b} = obj;", "const {a,,b} = obj;", "const [a b] = obj;", "type T = A | | B;", "type T = A & & B;", "type T = | A;", "function f(x, { return g(); }", "const { return x } = obj;", "x[];", "x?.[];", "f()[ ];", "import { A B } from \"m\";", "import { A,, B } from \"m\";", "export { A B };", "import { A from \"m\";", "function f(x:: T | undefined, y: U) {}", "function f(| T) {}", "function f(1) {}", "`unterminated", "const s = \"unterminated", "/* unterminated", "const r = /unterminated", "const f = ( => g();", "=> x", "const x = { a.b };", "module`M` {}", "tag`x` x;", "const x = { a: 1; };", "const x = { a: 1 b: 2 };", "type T = Omit<A \"b\"> & { b: string | false | undefined; };", "type T = A[", "function f(x: A[", "type T = A[;", "* from \"../moduleSpecifiers.js\";", "export export * from \"../moduleSpecifiers.js\";", "export from \"../moduleSpecifiers.js\";", "import import * as p from \"p.js\";", "import", "import\r\n  ", "}", ") ;", "x ?", "x ? y", "x ? y :", "x?.", "`x${", "`x${y", "type T = `x${", "interface I { : I | undefined; name: string; }", "interface I { x: T y: U }", "f(x x);", "f(, x);", "function f(){ g(x }", "function f(){ export }", "default x;", "switch(x){case 1:: f();}", "switch(x){default:: f();}", "switch(x){ f(); }", "namespace N { : f(); }", "function f(x: T y: U) {}", "function f(, x: T) {}", "function f(return) {}", "function f(const) {}", "function f(this: T) {}", "function f(x:: T) {}", "function f() x;", "break break;", "continue continue;", "break 1;", "throw", "throw\nx;", "throw x y;", "setParent setParent(x, y);", "retrun 1;", "Set 1;", "globl 1;", "İmport 1;", "export {", "import {", "f(", "[", "({", "function f() {", "class C {", "type T =", "type T = A<", "type T = (x:", "type T = ()", "type T = <", "function f(): (x: T, y:", "const {", "const const x = 1;", "const x = 1 y = 2;", "const , x = 1;", "const", "function f() { const", "const f = (x:", "const f = (...", "const f = ():", "const f = () {}", "export *", "export * from", "export * from\r\n  ", "export default", "export =", "export", "export\r\n  ", "export export", "export\nconst x = 1;", "export x;", "/* π 💡 */\r\nexport")
	cases = append(cases, `const s = "\1\12\123\377\400\8\9";`, "const x = 0123;", "const x = -0123;", "const x = - 0123;", "const x = 0000;", "const x = 0777777777777777777777777777777777;", "/* π 💡 */ const x = 0123;")
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
	recoveryCases := recoveryInputs(t)
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
		name, file, from, to, source string
		timeout, expressions         bool
	}{
		{"scanner-error-hook", "parser.ts", "this.scanner.errors.length !== errors", "this.scanner.errors.length === errors", recoveryInput(t, "unterminated-string.ts.txt"), false, false},
		{"reserved-token-switch", "grammar.ts", "case 'ReturnKeyword':", "case 'Identifier':", recoveryInput(t, "reserved-binding.ts.txt"), false, false},
		{"octal-suggestion", "lexical.ts", "const suggestion =", "const suggestion = 'wrong' +", `const s = "\123";`, false, false},
		{"diagnostic-code", "parser.ts", "this.error(1005, `'${tokenSpelling(kind)}' expected.`);\n            return false;", "this.error(1006, `'${tokenSpelling(kind)}' expected.`);\n            return false;", recoveryCases[0], false, false},
		{"stranded-export", "parser.ts", "this.error(diagnostic.code, diagnostic.message);", "this.error(diagnostic.code + (context === 'source' ? 1 : 0), diagnostic.message);", "export", false, false},
		{"generic-child", "parser.ts", "result.push(this.make('TypeParameter', pos, children));", "result.push(this.make('TypeParameter', pos, children.slice(0, 0)));", recoveryCases[3], false, false},
		{"eof-loop", "parser.ts", "return this.scanner.kind;", "return this.scanner.kind === 'EndOfFile' ? 'Identifier' : this.scanner.kind;", recoveryCases[1], true, false},
		{"speculative-roots", "parser.ts", "this.roots.splice(state.roots);", "this.roots.splice(this.roots.length);", "type T = ({ x = 1 }) => U;", false, true},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			path := filepath.Join(artifacts, mutation.name+".ts")
			if err := os.WriteFile(path, []byte(mutation.source), 0644); err != nil {
				t.Fatal(err)
			}
			args := []string{path, "--recovery"}
			if !mutation.expressions {
				args = append(args, "--whole")
			}
			want, err := recoveryRun(t, path+".go", oracle, args...)
			if err != nil {
				t.Fatal(err)
			}
			mutant := copyPort(t, mutation.file, mutation.from, mutation.to)
			binary := buildPort(t, mutant, true)
			for _, side := range []struct {
				name, command string
				args          []string
			}{
				{"Node", "node", append([]string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(mutant, "main.ts")}, args...)},
				{"native", binary, args},
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
	recoveryCases := recoveryInputs(t)
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	directory, err := filepath.Abs("../../cohere/lint")
	if err != nil {
		t.Fatal(err)
	}
	descriptors, err := registry.Generate(directory)
	if err != nil {
		t.Fatal(err)
	}
	replacements := map[string]string{}
	var virtualFiles []string
	add := func(name, source string) {
		virtual := filepath.Join(root, "adamic_lint_recovery_"+name+".go")
		replacements[virtual] = source
		virtualFiles = append(virtualFiles, virtual)
	}
	add("oracle", filepath.Join(directory, "testdata/oracle.go"))
	add("registry", filepath.Join(directory, ".generated/registry.go"))
	for _, descriptor := range descriptors {
		add(strings.ReplaceAll(descriptor.Slug, "-", "_"), filepath.Join(directory, "rules", descriptor.Slug, "oracle.go"))
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(t.TempDir(), "lint-oracle")
	buildArgs := append([]string{"build", "-overlay=" + overlayPath, "-o", oracle}, virtualFiles...)
	execute(t, root, "go", buildArgs...)
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
