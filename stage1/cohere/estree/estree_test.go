package estree

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func execute(test *testing.T, dir, name string, args ...string) []byte {
	test.Helper()
	command := exec.Command(name, args...)
	command.Dir = dir
	output, temporaryFileError := os.CreateTemp(test.TempDir(), "output")
	if temporaryFileError != nil {
		test.Fatal(temporaryFileError)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if runError := command.Run(); runError != nil || stderr.Len() != 0 {
		test.Fatalf("%s %v: %v\n%s", name, args, runError, &stderr)
	}
	data, readError := os.ReadFile(output.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	return data
}
func root(test *testing.T) string {
	test.Helper()
	// Reads: internal/native/runtime/*.[ch]
	path, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return path
}

// build is the port at path built natively, and its emitted JavaScript. A sanitized build is miscBuild's product, made
// ahead by a TestProduct_ for every path a test builds, so path must be a file of this checkout or of a product (a
// temporary copy would key a product no TestProduct_ makes). Only the unsanitized build, which TestThroughput alone
// asks for, is still made here.
func build(test *testing.T, path string, sanitize bool) (string, string) {
	test.Helper()
	if sanitize {
		return miscBuild(test, path)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	dir := test.TempDir()
	binary := filepath.Join(dir, "port")
	if buildError := native.Build(native.C(lowered), binary, native.Options{Sanitize: sanitize}); buildError != nil {
		test.Fatal(buildError)
	}
	script := filepath.Join(dir, "port.mjs")
	if writeError := os.WriteFile(script, []byte(javascript.JavaScript(lowered)), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return binary, script
}
func onNode(test *testing.T, path string, args ...string) []byte {
	test.Helper()
	argv := []string{"--disable-warning=ExperimentalWarning", filepath.Join(root(test), "oracle/node.mjs"), path}
	argv = append(argv, args...)
	return execute(test, "", "node", argv...)
}
func generated() []string {
	return []string{
		"type X = import('x').A<T>; type Y = typeof import('x', { with: { type: 'json' } }).A; type Z = typeof this.a;",
		"const x = tag`\\xZ`; const y = tag`\\uZZZZ`; const z = tag`\\u{61}`;",
		"let x: [A?, ...B[]]; type X = unique symbol; type Y<in T, out U> = T & U;",
		"namespace A.B { export let x = 1; } declare module 'x' { export const x: number; } declare global { interface X {} }",
		"type X<T> = { -readonly [K in keyof T as K]-?: T[K] }; type Y = { [K in A]?: B };",
		"for(;;) x; for(let i=0;i<3;i++) x; for(;x;) y; for(x;;y) z; for(;x;y) z;",
		"type T = [a: string, b?: number, ...rest: A[]]; type U = [A?, ...B[]];",
		"function f(x: unknown): x is string { return true; } function g(x: unknown): asserts x is A {}",
		"const x = tag<T>`a${b}c`; new.target; import.meta;",
		"import X, { type A as B, C } from 'x' with { type: 'json' }; import * as N from 'y'; import 'z';",
		"import type { A } from 'x'; import defer * as X from 'x'; import Y = require('y');",
		"export { type A as B, C }; export * as X from 'x'; export type * from 'y'; export default x; export = x;",
		"export declare const x: number; export let y = 1;",
		"class C extends A<B> implements X.Y<Z> { ; x?: number; y!: string; static readonly z = 1; constructor(public x: number) {} m<T>(a: T): T { return a; } get value() { return 1; } set value(x: number) {} }",
		"abstract class C { abstract x: number; abstract m(): void; accessor y = 1; static { 'x'; y; } }",
		"interface X<T> extends A.B<T> { readonly x?: number; m<U>(a: U): T; get value(): string; [x: string]: number; }",
		"const o = { m<T>(x: T): T { return x; }, get a() { return 1; }, set a(x: number) {} };",
		"type X<T> = (x: T) => T; type Y = abstract new (x: number) => A; type Z = { (x: number): A; new(): B; };",
		"enum E { A, B = 2 } export const enum F { X = 'x' }",
		"f<A, B,>(x); f?.<A>(x); new A<B>(x);",
		"function f(a: number, b?: string): number { 'use strict'; return a; }",
		"declare function f<T>(x: T): T; export default function g() {}",
		"async function* f() { yield* x; await x; }",
		"let x = (a: number = 1, ...b: string[]) => a; let y = async x => x;",
		"let [a,,b=1,...c]=d; let {a,b:c=2,...d}=e;",
		"type X = string; export type Y<T extends A = B> = T[];",
		"let x: A<B>; let y: A<B<C>, D,>;",
		"for(const x of y) x; for await(const x of y) x; for(x in y) x;",
		"switch(x) { case 1: y; break; default: z; }",
		"try { x; } catch(e: unknown) { y; } finally { z; }",
		"try {} catch {}",
		"",
		"// only\n",
		"/* only */",
		"#! /bin/runtime\nx;",
		"'use strict';\n'a\\x20b'; x; 'no';",
		"x;",
		"x /* 😀é */ ;",
		"x\u00a0;",
		"x\u2028;",
		"x\r\n;",
		"{ 'not a directive'; x; }",
		"1; 1.5; 0x10; 0b10; 0o10; 1_000; 'é😀'; true; false; null;",
		"1n; 0b101n; 0o17n;",
		"/a\\/*b/g;",
		"`hello`;",
		"`a${x}b${y}c`;",
		"`é${1}😀`;",
		"a+b*c;",
		"a=b+=c;",
		"a && (b && (c && d));",
		"a || (b || c);",
		"a ?? (b ?? c);",
		"a,(b,c),d;",
		"a?b:c;",
		"++x; x--; -x; +x; !x; ~x; typeof x; void x; delete x.a;",
		"this; super.x;",
		"a.b; a[b]; a?.b.c; (a?.b).c; a?.[b]; a?.(b); a?.b(); new A(x,y); import('x', options);",
		"[1,,2,...a]; ({ a: 1, b, [x]: 2, ...rest });",
		"[a,b]=c; ({a,b:c,...d}=e);",
		"let a; const b=1,c=2; var d=3; using e=f(); await using g=h();",
		"let a!: number; let b: string = 'x';",
		"if(x) a; else b;",
		"while(x) a;",
		"do x; while(y);",
		"with(x) y;",
		"label: while(x) { break label; continue; debugger; }",
		"return x; throw y;",
		"x; // / comment\n y;",
		"x; /* block */ y;",
		"x; /**\n * one\n */ /**\n * two\n */ y;",
		"x; /**\n * one\n *//**\n * two\n */ y;",
		"'/* no */ // no'; /\\/\\//; `/* no */`; // yes\n x;",
		"let x: number[]; let y: A[B]; let z: (A | B) & C; let q: [string, number];",
		"let x: A extends B ? C : D; let y: keyof A; let z: readonly string[];",
		"x as string; y satisfies number; <string>x; x!; a?.b!;",
		"let x: null; let y: true | 1 | 'a';",
		"let x: | A; let y: & B;",
	}
}
func manifest(test *testing.T, cases []string) string {
	test.Helper()
	dir := test.TempDir()
	var listing strings.Builder
	for index, text := range cases {
		path := filepath.Join(dir, fmt.Sprintf("%04d.ts", index))
		if writeError := os.WriteFile(path, []byte(text), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		listing.WriteString(path + "\n")
	}
	path := filepath.Join(dir, "manifest")
	if writeError := os.WriteFile(path, []byte(listing.String()), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return path
}
func firstDifference(want, got []byte) string {
	if bytes.Equal(want, got) {
		return ""
	}
	left, right := strings.Split(string(want), "\n"), strings.Split(string(got), "\n")
	for index := 0; index < len(left) && index < len(right); index++ {
		if left[index] != right[index] {
			return fmt.Sprintf("line %d: Go %q, port %q", index+1, left[index], right[index])
		}
	}
	return fmt.Sprintf("length: Go %d, port %d", len(want), len(got))
}

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestGeneratedAgreement(test *testing.T) {
	path, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	cases := generated()
	list := manifest(test, cases)
	want := execute(test, "", goOracle(test), "--manifest", list)
	if diff := firstDifference(want, onNode(test, path, "--manifest", list)); diff != "" {
		test.Fatal("source Node: " + diff)
	}
	binary, script := build(test, path, true)
	if diff := firstDifference(want, execute(test, "", binary, "--manifest", list)); diff != "" {
		test.Fatal("sanitized native: " + diff)
	}
	if diff := firstDifference(want, onNode(test, script, "--manifest", list)); diff != "" {
		test.Fatal("emitted JS: " + diff)
	}
	test.Logf(
		"%d generated files: %d identical bytes on Go, source Node, sanitized native, emitted JS",
		len(cases),
		len(want),
	)
}

func mutantPort(test *testing.T, file, from, to string) string {
	test.Helper()
	directory := test.TempDir()
	files, globError := filepath.Glob("*.ts")
	if globError != nil {
		test.Fatal(globError)
	}
	for _, name := range files {
		data, readError := os.ReadFile(name)
		if readError != nil {
			test.Fatal(readError)
		}
		text := string(data)
		if name == file {
			if strings.Count(text, from) != 1 {
				test.Fatalf("mutant anchor %q occurs %d times", from, strings.Count(text, from))
			}
			text = strings.Replace(text, from, to, 1)
		}
		text = strings.ReplaceAll(
			text,
			"'../../typescript/",
			"'"+filepath.ToSlash(filepath.Join(root(test), "stage1/typescript"))+"/",
		)
		if writeError := os.WriteFile(filepath.Join(directory, name), []byte(text), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return filepath.Join(directory, "main.ts")
}

// resolvedPort is name inside a product holding a mutant port, in the copy whose imports name this checkout, which
// node and the loader read; the product itself names it <repository> (#tqrqx60).
func resolvedPort(test *testing.T, product, name string) string {
	test.Helper()
	directory, resolvedError := buildcache.Resolved(product)
	if resolvedError != nil {
		test.Fatal(resolvedError)
	}
	return filepath.Join(directory, name)
}

func TestOriginalLibraries(test *testing.T) {
	test.Parallel()
	checkOriginalLibraries(test, generated(), 3)
}
func checkOriginalLibraries(test *testing.T, cases []string, postprocessedGaps int) {
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		test.Skip(
			"set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	checkOriginalManifest(test, manifest(test, cases), cases, postprocessedGaps)
}
func checkOriginalManifest(test *testing.T, list string, cases []string, postprocessedGaps int) {
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	oracle := goOracle(test)
	for _, mode := range []string{"raw", "postprocessed"} {
		flag := "--raw-json"
		if mode != "raw" {
			flag = "--json"
		}
		want := strings.Split(strings.TrimSpace(string(execute(test, "", oracle, flag, list))), "\n")
		script, absolutePathError := filepath.Abs("testdata/library.mjs")
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		got := strings.Split(strings.TrimSpace(string(execute(test, "", "node", script, library, mode, list))), "\n")
		if len(got) != len(want) {
			test.Fatalf("library %s: %d versus %d files", mode, len(got), len(want))
		}
		differences := 0
		for index := range want {
			var left, right any
			if decodeError := json.Unmarshal([]byte(want[index]), &left); decodeError != nil {
				test.Fatal(decodeError)
			}
			if decodeError := json.Unmarshal([]byte(got[index]), &right); decodeError != nil {
				test.Fatal(decodeError)
			}
			leftBytes, _ := json.Marshal(left)
			rightBytes, _ := json.Marshal(right)
			if !bytes.Equal(leftBytes, rightBytes) {
				source := cases[index]
				known := mode == "postprocessed" && (source == "x\u00a0;" || source == "x\u2028;" || source == "x\r\n;")
				if !known {
					test.Errorf("unrecorded %s case %d: Go %s; library %s", mode, index, leftBytes, rightBytes)
					continue
				}
				ast := left.(map[string]any)["ast"].(map[string]any)
				statement := ast["body"].([]any)[0].(map[string]any)
				if source == "x\r\n;" {
					if !bytes.Equal(mustJSON(test, ast["range"]), []byte("[0,4]")) ||
						!bytes.Equal(mustJSON(test, statement["range"]), []byte("[0,4]")) {
						test.Fatal("CRLF gap changed")
					}
					ast["range"] = []int{0, 3}
					statement["range"] = []int{0, 3}
				} else {
					if statement["__contentEnd"] != float64(2) {
						test.Fatal("multibyte whitespace gap changed")
					}
					statement["__contentEnd"] = 1
				}
				if !bytes.Equal(mustJSON(test, left), rightBytes) {
					test.Fatalf(
						"known gap has additional differences: %s case %d: Go after exact known delta %s; library %s",
						mode,
						index,
						mustJSON(test, left),
						rightBytes,
					)
				}
				differences++
				test.Logf("proved %s case %d known gap for %q", mode, index, source)
			}
		}
		test.Logf("%s: %d files, %d identical, %d differing", mode, len(want), len(want)-differences, differences)
		expected := 0
		if mode == "postprocessed" {
			expected = postprocessedGaps
		}
		if differences != expected {
			test.Errorf("expected exactly %d documented gaps, got %d", expected, differences)
		}
	}
}

func mustJSON(test *testing.T, value any) []byte {
	test.Helper()
	data, encodeError := json.Marshal(value)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	return data
}
