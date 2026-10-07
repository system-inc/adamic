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

var recoveryCases = []string{
	"interface I",
	"interface I { m(a: string): void;",
	"interface I { m<(a: string): void; }",
	"interface I { m<T(a: T): T; }",
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
	cases = append(cases, "f((_ x) => { g(); });", "f((x + y) => x);", "f((1) => x);", "const f = (x => x);", "a ? (x): T => x : b;", "type T = [A B];", "type T = [{ readonly x x: T; readonly y: U; }];", "type T = [, A];", "type T = [x: T y: U];", "type T = typeof", "type T = typeof ;", "type T = typeof X\n<T>;", "interface I { m() => T; }", "interface I { () => T; new() => T; }", "type T = import(\"fs\").", "type T = import(\"fs\").A.", "type T = typeof import(\"fs\").;", "class C { f() T | undefined {} }", "class C { f(): : T | undefined {} }", "class C { x y; }", "class C { x: T y; }", "class C { x: T() }", "class C { x @dec }", "class C { @dec }", "yield 1; await 1;", "yield\nx;", "function* f(){ yield }", "class C { static { await (1); } }", "function* f(){ class C { static { yield (1); } } }", "class C { static readonly", "class C { public", "class C { static {} }", "class C { public static {} }", "class C { static static {} }", "class C { static { await f(); } }", "{} = x;", "if(x) {} = y;", "const f = <T>(x => x);", "const f = <T>(1) => x;", "const f = <T>(x: ) => x;", "const f = <T>(x): => x;", "function f< <T extends N>(x: T[]): R<T> { return x; }", "const f = (x: T) => try {} catch {} };", "x => return x; }", "x => function() {};", "f(<T extends { readonly x: T; ; }>(x: T) => x);", "new Map<A, Map<B B, readonly C[] | undefined>>();", "f<A B>();", "type T = F<A B>;", "f<A,>();", "interface I { private boolean; }", "interface I { new", "interface I { get x", "class C { get x", "class C { *m", "({ get x", "interface I { set", "interface I { get", "class C { set", "({ set", "interface I { get\nx(): T; }", "catch {}", "finally {}", "try {} catch catch {}", "(x) { f(); } else { g(); }", "(x + y) { f(); }", "const f = <T>(x) { return x; };", "a + b = c;", "> = (a: T) => a;", "type T<X>> = (a: X) => X;", "({ readonly size: number; has(key: K): boolean; })", "({ public x: 1, async f() {}, readonly x })", "({ *m })", "interface I { [index: string: T; }", "interface I { [x: T y: U]: V; }", "class C { [x: T y: U]: V; }", "class C { : x; }", "class C { ) }", "class C", "class C x;", "class C extends", "class C extends {}", "class C extends {} {}", "class C extends A B {}", "class C extends , {}", "class C implements {}", "interface I extends {}", "interface I extends A B {}", "class C implements A.B {}", "interface I extends f() {}")
	cases = append(cases, "await new Promise<void>(function(resolve) { setTimeout(resolve, 500); }); export {};", "new Promise(@dec async () => {})", "class Foo { constructor(public { a }: { a: string }) { this.a = a; } }", "var yield = 5; yield: while (foo) { new Object(); }")
	cases = append(cases, "class C { get #a() { return 1; } #a = 1; m() { this.#a = 1; } }", "class C { set #a(x: number) {} }")
	cases = append(cases, "f(a @x {b});", "f(@dec class C {});", "const x = @dec class C {};", "f(@dec);", "f(@dec public class {});", "f(@a @b {x});")
	cases = append(cases, "type T = .A | B;", "type T = .;", "type T = | .A;", "type T = A & .B;", "type T = .A<B>;")
	cases = append(cases, "const x = { a: function function f() {} };", "const x = function return;", "const x = function* function f() {};", "const x = async function function f() {};")
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
	if !t.Failed() {
		t.Logf("%d focused recovery cases: Go, Node and sanitized native identical", len(cases))
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
		timeout, expressions   bool
		file                   string
	}{
		{"private-accessor", "['PrivateIdentifier', 'OpenBracketToken', 'StringLiteral', 'NumericLiteral', 'BigIntLiteral']", "['OpenBracketToken', 'StringLiteral', 'NumericLiteral', 'BigIntLiteral']", "class C { get #a() { return 1; } }", false, false, "lookahead.ts"},
		{"octal-suggestion", "const suggestion =", "const suggestion = 'wrong' +", `const s = "\123";`, false, false, "lexical.ts"},
		{"diagnostic-code", "this.error(1005, `'${tokenSpelling(kind)}' expected.`);\n            return false;", "this.error(1006, `'${tokenSpelling(kind)}' expected.`);\n            return false;", recoveryCases[0], false, false, "parser.ts"},
		{"stranded-export", "this.error(diagnostic.code, diagnostic.message);", "this.error(diagnostic.code + (context === 'source' ? 1 : 0), diagnostic.message);", "export", false, false, "parser.ts"},
		{"generic-child", "result.push(this.make('TypeParameter', pos, children));", "result.push(this.make('TypeParameter', pos, children.slice(0, 0)));", recoveryCases[3], false, false, "parser.ts"},
		{"eof-loop", "return this.scanner.kind;", "return this.scanner.kind === 'EndOfFile' ? 'Identifier' : this.scanner.kind;", recoveryCases[1], true, false, "parser.ts"},
		{"corpus-eof-loop", "return this.scanner.kind;", "return this.scanner.kind === 'EndOfFile' ? 'Identifier' : this.scanner.kind;", recoveryCases[1], true, false, "parser.ts"},
		{"speculative-roots", "this.roots.splice(state.roots);", "this.roots.splice(this.roots.length);", "type T = ({ x = 1 }) => U;", false, true, "parser.ts"},
		{"function-expression-name", "if(bindingStart(this.kind())) {\n            children.push(this.identifier());\n        }\n        const oldAwait", "if(this.kind() === 'Identifier' || this.kind().endsWith('Keyword')) {\n            children.push(this.identifier());\n        }\n        const oldAwait", "const x = { a: function function f() {} };", false, false, "parser.ts"},
		{"missing-type-qualifier", "const name = this.entityName(true);", "const name = this.missingIdentifier(1110, 'Type expected.');", "type T = .A | B;", false, false, "parser.ts"},
		{"decorated-expression", "return this.make('MissingDeclaration', pos, modifiers);", "return this.make('Identifier', pos, modifiers);", "f(a @x {b});", false, false, "parser.ts"},
		{"arrow-head-rejection", "let result = 0;", "let result = 2;", "f((_ x) => { g(); });", false, false, "lookahead.ts"},
		{"tuple-recovery", "this.delimitedList('tuple', () => this.tupleElement())", "this.delimitedList('typeArguments', () => this.tupleElement())", "type T = [A B];", false, false, "parser.ts"},
		{"type-query-diagnostic", "const name = this.entityName();", "const name = this.entityName(true);", "type T = typeof", false, false, "parser.ts"},
		{"signature-return-token", "this.kind() === 'EqualsGreaterThanToken' &&\n                ['CallSignature', 'ConstructSignature', 'MethodSignature'].includes(kind)", "false", "interface I { m() => T; }", false, false, "parser.ts"},
		{"import-qualifier", "children.push(this.entityName(true));", "children.push(this.entityName());", "type T = import(\"fs\").", false, false, "parser.ts"},
		{"class-missing-name", "1146, missing, missing,", "1003, missing, missing,", "class C { @dec }", false, false, "statements.ts"},
		{"property-semicolon", "this.expressionSemicolon(name, start);", "this.semicolon();", "class C { x y; }", false, false, "statements.ts"},
		{"contextual-literal", "['NumericLiteral', 'BigIntLiteral', 'StringLiteral'].includes(after)", "['BigIntLiteral', 'StringLiteral'].includes(after)", "yield 1; await 1;", false, false, "lookahead.ts"},
		{"static-await", "this.parser.setAwait(true);\n                this.parser.setYield(false);", "this.parser.setAwait(false);\n                this.parser.setYield(false);", "class C { static { await (1); } }", false, false, "statements.ts"},
		{"static-yield", "this.parser.setAwait(true);\n                this.parser.setYield(false);", "this.parser.setAwait(true);\n                this.parser.setYield(true);", "function* f(){ class C { static { yield (1); } } }", false, false, "statements.ts"},
		{"class-modifiers", "const member = this.parser.modifiers(true, true, true);", "const member = this.parser.modifiers(true, true, true).slice(0, 0);", "class C { static readonly", false, false, "statements.ts"},
		{"static-block-modifier", "if(stopStaticBlock &&", "if(false &&", "class C { static {} }", false, false, "parser.ts"},
		{"block-equals-diagnostic", "2809,", "1128,", "{} = x;", false, false, "statements.ts"},
		{"class-missing-brace", "if(!this.parser.expect('OpenBraceToken')) {\n            return this.make(expression", "if(!this.parser.expect('OpenBraceToken') && false) {\n            return this.make(expression", "class C x;", false, false, "statements.ts"},
		{"class-recovery", "if(!this.parser.listElement('class')) {", "if(false) {", "class C { ) }", true, false, "statements.ts"},
		{"empty-heritage", "return ['OpenBraceToken', 'ExtendsKeyword', 'ImplementsKeyword'].includes(kind);", "return ['ExtendsKeyword', 'ImplementsKeyword'].includes(kind);", "class C extends {}", false, false, "recovery.ts"},
		{"stranded-catch", "['TryKeyword', 'CatchKeyword', 'FinallyKeyword'].includes(this.parser.kind())", "this.parser.kind() === 'TryKeyword'", "catch {}", true, false, "statements.ts"},
		{"arrow-missing-block", "!this.parser.expect('OpenBraceToken') && !ignoreMissing", "!this.parser.expect('OpenBraceToken')", "x => return x; }", false, false, "statements.ts"},
		{"generic-arrow-name", "!allowAmbiguity &&\n            !bindingStart(this.kind())", "false &&\n            !bindingStart(this.kind())", "const f = <T>(1) => x;", false, false, "parser.ts"},
		{"generic-arrow-block", "let result = ['EqualsGreaterThanToken', 'OpenBraceToken'].includes(parser.kind());", "let result = ['EqualsGreaterThanToken'].includes(parser.kind());", "const f = <T>(x): U { return x; }", false, false, "lookahead.ts"},
		{"generic-arrow-parameters", "if(!parser.speculativeParameters()) {", "if(!parser.speculativeParameters() && false) {", "const f = <T>(x => x);", false, false, "lookahead.ts"},
		{"generic-arrow-return", "if(!arrowReturnTypeValid(parser, parser.returnType())) {", "if(!arrowReturnTypeValid(parser, parser.returnType()) && false) {", "const f = <T>(x): => x;", false, false, "lookahead.ts"},
		{"generic-arrow-speculation", "if(!valid) {\n        parser.rewind(state);\n        return false;", "if(!valid) {\n        parser.rewind(state);\n        return true;", "f(<T extends { readonly x: T; ; }>(x: T) => x);", false, false, "lookahead.ts"},
		{"type-argument-speculation", "if(parser.kind() === 'GreaterThanToken') {\n        parser.next();\n        const after", "if(true) {\n        parser.next();\n        const after", "new Map<A, Map<B B, readonly C[] | undefined>>();", false, false, "lookahead.ts"},
		{"type-member-modifiers", "const children = this.modifiers(false, false);", "const children = this.modifiers(false, false).slice(0, 0);", "interface I { private boolean; }", false, false, "parser.ts"},
		{"accessor-parameters", "kind === 'GetAccessor' ||\n                    kind === 'SetAccessor' ||", "false ||", "interface I { get x", false, false, "parser.ts"},
		{"construct-recognition", "this.kind() === 'NewKeyword' && ['OpenParenToken', 'LessThanToken'].includes(this.peek())", "this.kind() === 'NewKeyword'", "interface I { new", false, false, "parser.ts"},
		{"accessor-follow", "const result =\n        kind(scanner) === 'Identifier'", "const result =\n        kind(scanner) === 'EndOfFile' || kind(scanner) === 'Identifier'", "interface I { set", false, false, "lookahead.ts"},
		{"arrow-block", "let result = ['EqualsGreaterThanToken', 'OpenBraceToken'].includes(parser.kind());", "let result = ['EqualsGreaterThanToken'].includes(parser.kind());", "(x) { f(); }", false, false, "lookahead.ts"},
		{"heritage-kind", "name >= 0 ? 'TypeReference' : 'ExpressionWithTypeArguments'", "name >= 0 ? 'ExpressionWithTypeArguments' : 'ExpressionWithTypeArguments'", "class C implements A.B {}", false, false, "statements.ts"},
		{"assignment-left", "leftHandSideKind(this.node(left).kind)", "true", "a + b = c;", false, false, "parser.ts"},
		{"object-modifiers", "const children = this.modifiers(true, false);\n        let async = false;", "const children = this.modifiers(true, false).slice(0, 0);\n        let async = false;", "({ readonly size: number; })", false, false, "parser.ts"},
		{"index-parameters", "for(const parameter of this.indexParameters()) {", "for(const parameter of this.indexParameters().slice(0, 1)) {", "interface I { [x: T y: U]: V; }", false, false, "parser.ts"},
		{"heritage-comma", "types.push(this.heritageElement(typeHeritage));", "types.push(this.heritageElement(typeHeritage)); this.parser.next();", "class C extends A B {}", false, false, "statements.ts"},
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
				limit := 2 * time.Second
				if mutation.name == "corpus-eof-loop" {
					limit = incompleteDeadline
				}
				got, err := recoveryRunLimit(t, limit, path+"."+side.name, side.command, side.args...)
				if mutation.timeout {
					if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("timeout after %s", limit)) {
						t.Fatalf("%s must be caught by the deadline, got %v", side.name, err)
					}
					t.Logf("%s: nonterminating mutant caught by %s deadline; input %s", side.name, limit, path)
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
