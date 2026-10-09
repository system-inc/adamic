package estree

import (
	"bytes"
	"context"
	"crypto/sha256"
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
	"runtime"
	"strings"
	"testing"
	"time"
)

func execute(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	if filepath.Base(name) == "oracle" && len(args) == 2 && args[0] == "--manifest" {
		return oracleOutput(t, dir, name, args[1])
	}
	return executeUncached(t, dir, name, args...)
}

func oracleOutput(t *testing.T, working, oracle, manifest string) []byte {
	t.Helper()
	binary, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "estree-oracle-output", Files: []string{"stage1/cohere/estree/estree_test.go"}, Flags: []string{fmt.Sprintf("oracle-sha256=%x", sha256.Sum256(binary)), "working=" + working, "--manifest"}, Toolchain: []string{runtime.GOOS, runtime.GOARCH}}
	for i, path := range strings.Fields(string(paths)) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("case-%03d=%x", i, sha256.Sum256(data)))
	}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		output := executeUncached(t, working, oracle, "--manifest", manifest)
		return os.WriteFile(filepath.Join(directory, "canonical"), output, 0644)
	})
	data, err := os.ReadFile(filepath.Join(directory, "canonical"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func executeUncached(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = dir
	output, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil || stderr.Len() != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func root(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	return path
}
func goOracle(t *testing.T) string {
	t.Helper()
	repo := root(t)
	source, err := filepath.Abs("testdata/oracle.go")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(repo, "cohere/adamic_estree_oracle.go")
	dir := prepareOverlayOracle(t, func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		command := exec.Command("go", "build", "-overlay="+path, "-o", filepath.Join(dir, "oracle"), virtual)
		command.Dir = filepath.Join(repo, "cohere")
		output, err := command.CombinedOutput()
		if err != nil || len(output) != 0 {
			return fmt.Errorf("go oracle: %v: %s", err, output)
		}
		return nil
	})
	return filepath.Join(dir, "oracle")
}

func build(t *testing.T, path string, sanitize bool) (string, string) {
	t.Helper()
	inputs := portBuildInputs(t, sanitize)
	relative, err := filepath.Rel(root(t), path)
	if err != nil {
		t.Fatal(err)
	}
	// Temporary mutant directories are snapshotted by content in the key and
	// copied into the product before lowering. Their random path is not an input.
	if strings.HasPrefix(relative, "..") {
		inputs.Name += "-temporary-source"
		files, err := filepath.Glob(filepath.Join(filepath.Dir(path), "*.ts"))
		if err != nil {
			t.Fatal(err)
		}
		snapshot := make(map[string][]byte, len(files))
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Base(file)
			snapshot[name] = data
			inputs.Flags = append(inputs.Flags, fmt.Sprintf("source-%s=%x", name, sha256.Sum256(data)))
		}
		inputs.Flags = append(inputs.Flags, "entry="+filepath.Base(path))
		dir := buildcache.Product(t, inputs, func(dir string) error {
			source := filepath.Join(dir, "source")
			if err := os.Mkdir(source, 0755); err != nil {
				return err
			}
			for name, data := range snapshot {
				if err := os.WriteFile(filepath.Join(source, name), data, 0644); err != nil {
					return err
				}
			}
			return buildPortInto(t, filepath.Join(source, filepath.Base(path)), sanitize, dir)
		})
		return filepath.Join(dir, "port"), filepath.Join(dir, "port.mjs")
	}
	inputs.Flags = append(inputs.Flags, "entry="+filepath.ToSlash(relative))
	dir := buildcache.Product(t, inputs, func(dir string) error { return buildPortInto(t, path, sanitize, dir) })
	return filepath.Join(dir, "port"), filepath.Join(dir, "port.mjs")
}

func buildPortInto(t *testing.T, path string, sanitize bool, dir string) error {
	start := time.Now()
	program, err := load.Load([]string{path})
	if err != nil {
		return err
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return err
	}
	t.Logf("cold build load/lower: %.3fs", time.Since(start).Seconds())
	start = time.Now()
	c := native.C(lowered)
	if err := os.WriteFile(filepath.Join(dir, "program.c"), []byte(c), 0644); err != nil {
		return err
	}
	if err := native.Build(c, filepath.Join(dir, "port"), native.Options{Sanitize: sanitize}); err != nil {
		return err
	}
	t.Logf("cold build sanitized=%t native: %.3fs", sanitize, time.Since(start).Seconds())
	start = time.Now()
	err = os.WriteFile(filepath.Join(dir, "port.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
	t.Logf("cold build emitted JS: %.3fs", time.Since(start).Seconds())
	return err
}

func onNode(t *testing.T, path string, args ...string) []byte {
	t.Helper()
	argv := []string{"--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), path}
	argv = append(argv, args...)
	return execute(t, "", "node", argv...)
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
		"", "// only\n", "/* only */", "#! /bin/runtime\nx;", "'use strict';\n'a\\x20b'; x; 'no';",
		"x;", "x /* 😀é */ ;", "x\u00a0;", "x\u2028;", "x\r\n;", "{ 'not a directive'; x; }",
		"1; 1.5; 0x10; 0b10; 0o10; 1_000; 'é😀'; true; false; null;",
		"1n; 0b101n; 0o17n;", "/a\\/*b/g;", "`hello`;", "`a${x}b${y}c`;", "`é${1}😀`;",
		"a+b*c;", "a=b+=c;", "a && (b && (c && d));", "a || (b || c);", "a ?? (b ?? c);", "a,(b,c),d;",
		"a?b:c;", "++x; x--; -x; +x; !x; ~x; typeof x; void x; delete x.a;", "this; super.x;",
		"a.b; a[b]; a?.b.c; (a?.b).c; a?.[b]; a?.(b); a?.b(); new A(x,y); import('x', options);",
		"[1,,2,...a]; ({ a: 1, b, [x]: 2, ...rest });", "[a,b]=c; ({a,b:c,...d}=e);",
		"let a; const b=1,c=2; var d=3; using e=f(); await using g=h();", "let a!: number; let b: string = 'x';",
		"if(x) a; else b;", "while(x) a;", "do x; while(y);", "with(x) y;", "label: while(x) { break label; continue; debugger; }",
		"return x; throw y;", "x; // / comment\n y;", "x; /* block */ y;", "x; /**\n * one\n */ /**\n * two\n */ y;",
		"x; /**\n * one\n *//**\n * two\n */ y;", "'/* no */ // no'; /\\/\\//; `/* no */`; // yes\n x;",
		"let x: number[]; let y: A[B]; let z: (A | B) & C; let q: [string, number];",
		"let x: A extends B ? C : D; let y: keyof A; let z: readonly string[];",
		"x as string; y satisfies number; <string>x; x!; a?.b!;", "let x: null; let y: true | 1 | 'a';", "let x: | A; let y: & B;",
	}
}
func manifest(t *testing.T, cases []string) string {
	t.Helper()
	dir := t.TempDir()
	var listing strings.Builder
	for index, text := range cases {
		path := filepath.Join(dir, fmt.Sprintf("%04d.ts", index))
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		listing.WriteString(path + "\n")
	}
	path := filepath.Join(dir, "manifest")
	if err := os.WriteFile(path, []byte(listing.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
func firstDifference(want, got []byte) string {
	if bytes.Equal(want, got) {
		return ""
	}
	left, right := strings.Split(string(want), "\n"), strings.Split(string(got), "\n")
	for i := 0; i < len(left) && i < len(right); i++ {
		if left[i] != right[i] {
			return fmt.Sprintf("line %d: Go %q, port %q", i+1, left[i], right[i])
		}
	}
	return fmt.Sprintf("length: Go %d, port %d", len(want), len(got))
}
func TestGeneratedAgreement(t *testing.T) {
	path, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	cases := generated()
	list := manifest(t, cases)
	want := execute(t, "", goOracle(t), "--manifest", list)
	if diff := firstDifference(want, onNode(t, path, "--manifest", list)); diff != "" {
		t.Fatal("source Node: " + diff)
	}
	binary, script := build(t, path, true)
	if diff := firstDifference(want, execute(t, "", binary, "--manifest", list)); diff != "" {
		t.Fatal("sanitized native: " + diff)
	}
	if diff := firstDifference(want, onNode(t, script, "--manifest", list)); diff != "" {
		t.Fatal("emitted JS: " + diff)
	}
	t.Logf("%d generated files: %d identical bytes on Go, source Node, sanitized native, emitted JS", len(cases), len(want))
}

func mutantPort(t *testing.T, file, from, to string) string {
	t.Helper()
	directory := t.TempDir()
	return mutantPortInto(t, directory, file, from, to)
}

func mutantPortInto(t *testing.T, directory, file, from, to string) string {
	t.Helper()
	files, err := filepath.Glob("*.ts")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if name == file {
			if strings.Count(text, from) != 1 {
				t.Fatalf("mutant anchor %q occurs %d times", from, strings.Count(text, from))
			}
			text = strings.Replace(text, from, to, 1)
		}
		text = strings.ReplaceAll(text, "'../../typescript/", "'"+filepath.ToSlash(filepath.Join(root(t), "stage1/typescript"))+"/")
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(directory, "main.ts")
}

func TestOriginalLibraries(t *testing.T) {
	checkOriginalLibraries(t, generated(), 3)
}
func checkOriginalLibraries(t *testing.T, cases []string, postprocessedGaps int) {
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	checkOriginalManifest(t, manifest(t, cases), cases, postprocessedGaps)
}
func checkOriginalManifest(t *testing.T, list string, cases []string, postprocessedGaps int) {
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	oracle := goOracle(t)
	for _, mode := range []string{"raw", "postprocessed"} {
		flag := "--raw-json"
		if mode != "raw" {
			flag = "--json"
		}
		want := strings.Split(strings.TrimSpace(string(execute(t, "", oracle, flag, list))), "\n")
		script, err := filepath.Abs("testdata/library.mjs")
		if err != nil {
			t.Fatal(err)
		}
		got := strings.Split(strings.TrimSpace(string(execute(t, "", "node", script, library, mode, list))), "\n")
		if len(got) != len(want) {
			t.Fatalf("library %s: %d versus %d files", mode, len(got), len(want))
		}
		differences := 0
		for index := range want {
			var left, right any
			if err := json.Unmarshal([]byte(want[index]), &left); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(got[index]), &right); err != nil {
				t.Fatal(err)
			}
			leftBytes, _ := json.Marshal(left)
			rightBytes, _ := json.Marshal(right)
			if !bytes.Equal(leftBytes, rightBytes) {
				source := cases[index]
				known := mode == "postprocessed" && (source == "x\u00a0;" || source == "x\u2028;" || source == "x\r\n;")
				if !known {
					t.Errorf("unrecorded %s case %d: Go %s; library %s", mode, index, leftBytes, rightBytes)
					continue
				}
				ast := left.(map[string]any)["ast"].(map[string]any)
				statement := ast["body"].([]any)[0].(map[string]any)
				if source == "x\r\n;" {
					if !bytes.Equal(mustJSON(t, ast["range"]), []byte("[0,4]")) || !bytes.Equal(mustJSON(t, statement["range"]), []byte("[0,4]")) {
						t.Fatal("CRLF gap changed")
					}
					ast["range"] = []int{0, 3}
					statement["range"] = []int{0, 3}
				} else {
					if statement["__contentEnd"] != float64(2) {
						t.Fatal("multibyte whitespace gap changed")
					}
					statement["__contentEnd"] = 1
				}
				if !bytes.Equal(mustJSON(t, left), rightBytes) {
					t.Fatalf("known gap has additional differences: %s case %d: Go after exact known delta %s; library %s", mode, index, mustJSON(t, left), rightBytes)
				}
				differences++
				t.Logf("proved %s case %d known gap for %q", mode, index, source)
			}
		}
		t.Logf("%s: %d files, %d identical, %d differing", mode, len(want), len(want)-differences, differences)
		expected := 0
		if mode == "postprocessed" {
			expected = postprocessedGaps
		}
		if differences != expected {
			t.Errorf("expected exactly %d documented gaps, got %d", expected, differences)
		}
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
