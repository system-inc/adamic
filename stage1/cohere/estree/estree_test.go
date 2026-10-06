package estree

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func execute(t *testing.T, dir, name string, args ...string) []byte {
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
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "oracle")
	execute(t, filepath.Join(repo, "cohere"), "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}
func build(t *testing.T, path string, sanitize bool) (string, string) {
	t.Helper()
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "port")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: sanitize}); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "port.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(lowered)), 0644); err != nil {
		t.Fatal(err)
	}
	return binary, script
}
func onNode(t *testing.T, path string, args ...string) []byte {
	t.Helper()
	argv := []string{"--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), path}
	argv = append(argv, args...)
	return execute(t, "", "node", argv...)
}
func generated() []string {
	return []string{
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
func TestThreePortMutants(t *testing.T) {
	list := manifest(t, generated())
	want := execute(t, "", goOracle(t), "--manifest", list)
	for _, item := range []struct{ name, file, from, to string }{
		{"member-computed", "convert.ts", "boolValue(node.kind === 'ElementAccessExpression')", "boolValue(node.kind === 'PropertyAccessExpression')"},
		{"logical-rebalance", "postprocess.ts", "return this.rebalance(id);", "return id;"},
		{"merged-jsdoc-value", "postprocess.ts", "*//*", "*/ /*"},
	} {
		t.Run(item.name, func(t *testing.T) {
			path := mutantPort(t, item.file, item.from, item.to)
			got := onNode(t, path, "--manifest", list)
			diff := firstDifference(want, got)
			if diff == "" {
				t.Fatal("source Node mutant survived")
			}
			t.Logf("source Node finished; byte comparison caught %s", diff)
			binary, _ := build(t, path, true)
			got = execute(t, "", binary, "--manifest", list)
			diff = firstDifference(want, got)
			if diff == "" {
				t.Fatal("native mutant survived")
			}
			t.Logf("sanitized native finished; byte comparison caught %s", diff)
		})
	}
}
