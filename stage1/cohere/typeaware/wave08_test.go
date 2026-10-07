package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var wave08Names = []string{"no-loop-func", "nexus/correctness-no-import-cycle-load-time-read", "@typescript-eslint/no-require-imports"}

func wave08Controls() []string {
	return []string{
		"export function f(){for(var i=0;i<3;i++){consume(()=>i)}}declare function consume(x:()=>number):void;",
		"for(let i=0;i<3;i++){consume(()=>i)}declare function consume(x:()=>number):void;",
		"let a=0;for(let i=0;i<3;i++){consume(()=>a)}a++;declare function consume(x:()=>number):void;",
		"let a=0;for(let i=0;i<3;i++){consume(()=>a)}declare function consume(x:()=>number):void;",
		"for(var i=0;i<3;i++){(()=>i)();(function named(){named;i})()}for(var [a,b] of [[1,2]]){consume(()=>a+b)}declare function consume(x:()=>number):void;",
		"for(var i=0;i<3;i++){const o={method(){return i},get x(){return i},set x(v:number){i=v}};class C{constructor(){console.log(i)}method(){return i}}}",
		"for(var i=0;i<3;i++){(()=>{consume(()=>i)})();(async()=>i)();(function*(){yield i})();}declare function consume(x:()=>number):void;",
		"for(var i=0,f=()=>i;i<3;i++){};let a=0;for(const x of [1,2]){for(const y of [1]){consume(()=>a+x+y)}a++}declare function consume(x:()=>number):void;",
		"for(var i=0;i<3;i++){consume((i:number)=>i);consume(()=>{const i=1;return i})}declare function consume(x:any):void;",
		"require('one');(require)('two');require?.('three');import c=require('four');",
		"function require(x:string){return x}require('local');export {};",
		"declare function require(x:string):unknown;require('ambient');export {};",
		"declare global{function require(x:string):unknown}require('global');export {};",
		"function use(require:(s:string)=>unknown){require('parameter')}export {use};",
		"/* 世界 🌍 */\r\nfor(var é=0;é<2;é++){consume(()=>é)}require('unicode');declare function consume(x:()=>number):void;",
	}
}

// Not parallel: archives, sanitizers and timings share the limited scratch disk.
func TestWave08AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE08_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave08_suite.a")
	binary := h.build(stage0, "wave08", entry, archive, false)
	oracle := volumeOracle(h, "wave08-oracle", "oracle_wave08.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range wave08Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	for name, source := range map[string]string{
		"cycle-a.ts":     "import {b,hoisted} from './cycle-b.ts';export const a=b;export const called=hoisted();export function later(){return b}",
		"cycle-b.ts":     "import {a} from './cycle-a.ts';export const b=a;export function hoisted(){return 1}",
		"namespace-a.ts": "import * as b from './namespace-b.ts';export const a=b.b;export class Derived extends b.Base{}",
		"namespace-b.ts": "import {a} from './namespace-a.ts';export const b=a;export class Base{}",
		"type-a.ts":      "import type {B} from './type-b.ts';export const a=1;export type A=B;",
		"type-b.ts":      "import {a} from './type-a.ts';export const b=a;export type B=number;",
		"elided-a.ts":    "import {b} from './elided-b.ts';export const a=1;export type A=typeof b;",
		"elided-b.ts":    "import {a} from './elided-a.ts';export const b=a;",
		"static-a.ts":    "import {b} from './static-b.ts';export const a=1;export class C{static x=b;static{console.log(b)}instance=b;method(){return b}}",
		"static-b.ts":    "import {a} from './static-a.ts';export const b=a;",
		"server-a.ts":    "import {b} from './server-b.ts';export const a=b;",
		"server-b.ts":    "'use server';import {a} from './server-a.ts';export const b=a;",
		"reexport-a.ts":  "import {b} from './reexport-c.ts';export const a={b};",
		"reexport-b.ts":  "import {a} from './reexport-a.ts';export const b=a;",
		"reexport-c.ts":  "export {b} from './reexport-b.ts';",
	} {
		paths = append(paths, h.write(name, source+"\n"))
	}
	// Stable file order also fixes graph path ordering in the oracle's output.
	// Graph edges themselves retain source statement order, as Go does.
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range wave08Names {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave08-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"loop", "no_loop_func.a", "if(names.length > 0)", "if(names.length > 1)"},
		{"require", "no_require_imports.a", "(declaration.flags & 8388608) === 0", "(declaration.flags & 8388608) !== 0"},
		{"cycle", "no_import_cycle_load_time_read.a", "if(!uninitialized)", "if(uninitialized)"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			scratch := filepath.Join(directory, change.name+"-source")
			if err := os.MkdirAll(scratch, 0755); err != nil {
				t.Fatal(err)
			}
			for _, extension := range []string{"*.ts", "*.a"} {
				files, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware", extension))
				if err != nil {
					t.Fatal(err)
				}
				for _, file := range files {
					data, err := os.ReadFile(file)
					if err != nil {
						t.Fatal(err)
					}
					source := string(data)
					if filepath.Base(file) == change.file {
						if strings.Count(source, change.from) != 1 {
							t.Fatal("nonunique mutant")
						}
						source = strings.Replace(source, change.from, change.to, 1)
					}
					source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
					source = strings.ReplaceAll(source, "../lint/", filepath.Join(repository, "stage1/cohere/lint")+"/")
					if err := os.WriteFile(filepath.Join(scratch, filepath.Base(file)), []byte(source), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			mutant := h.build(stage0, change.name+"-mutant", filepath.Join(scratch, "wave08_suite.a"), archive, false)
			got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
			if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
				t.Fatalf("%s mutant survived", change.name)
			}
			t.Logf("%s: exit 0, empty stderr, Go byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
			if err := os.Remove(mutant); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE08_COMPILER_MANIFEST")},
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE08_REPOSITORY_MANIFEST")},
	} {
		if corpus.manifest == "" {
			t.Logf("%s corpus not requested", corpus.name)
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const path=args[1]??'';const program=tsgoProgram(args[0]??'',[path]);tsgoRelease(program);console.log(tsgoInspect(program,path,0,1,'Identifier','wave08-symbol'));`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released handle panics 70; retaining-registry mutant exits 0 and fails that expectation")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}

// Not parallel: isolated compiler overlays reuse the machine's build cache.
func TestWave08CheckerFactMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE08_FACT_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	h.must("facts-control", exec.Command("go", "test", "./bridge/tsgo/checker", "-run", "^TestWave08Facts$", "-count=1", "-v"))
	for _, change := range []struct{ name, file, from, to, signal string }{
		{"symbol-flags", "wave08_symbol.go", "out.number(uint64(declaration.End()))\n\t\tout.number(uint64(declaration.Flags))", "out.number(uint64(declaration.End()))\n\t\tout.number(uint64(declaration.Flags) ^ uint64(ast.NodeFlagsAmbient))", "declaration differs"},
		{"type-syntax", "syntax_metadata.go", "out.yes(ast.IsTypeNode(node))", "out.yes(!ast.IsTypeNode(node))", "syntax differs"},
		{"module-target", "module_links.go", "target = found.FileName()", "target = \"\"", "resolved module absent"},
	} {
		overlay := h.overlay(change.name, "bridge/tsgo/checker/"+change.file, change.from, change.to)
		got := h.run(change.name, exec.Command("go", "test", "-overlay", overlay, "./bridge/tsgo/checker", "-run", "^TestWave08Facts$", "-count=1", "-v"))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 1 || !bytes.Contains(got.stdout, []byte(change.signal)) || bytes.Contains(got.stdout, []byte("build failed")) {
			t.Fatalf("%s mutant not caught by its fact check: %v %s %s", change.name, got.err, got.stdout, got.stderr)
		}
		t.Logf("%s: compiled, direct checker comparison caught %s", change.name, change.signal)
	}
}
