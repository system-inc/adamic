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

func wave18Controls() []string {
	return []string{
		`export async function f(){ await 0; await Promise.resolve(1); await (()=>1); await (null); }`,
		`declare const a:any,b:unknown,n:never,p:number|Promise<number>; export async function f(){await a;await b;await n;await p;}`,
		`export async function f<T,U extends number,V extends PromiseLike<number>>(a:T,b:U,c:V){await a;await b;await c;}`,
		`declare const bad:{then(x:number):void}, good:{then(f:(x:number)=>void):void}, rest:{then(...fs:((x:number)=>void)[]):void};export async function f(){await bad;await good;await rest;}`,
		`export async function f(){for await(const n of [1,2]) {} for await(const p of [Promise.resolve(1)]) {} }`,
		`declare const a:AsyncIterable<number>,b:Iterable<number>|AsyncIterable<number>,c:any;export async function f(){for await(const x of a){} for await(const x of b){} for await(const x of c){}}`,
		`declare const s:Disposable,a:AsyncDisposable,b:any;export async function f(){await using x=s;await using y=a;await using z=b;}`,
		`declare const s:Disposable,a:AsyncDisposable;export async function f(){await using x=s,y=a,z=s;}`,
		`declare const mixed:number|Promise<number>;export const a=Promise.all([1,Promise.resolve(2),mixed]);export const b=Promise.all(([((3))]));`,
		`declare const rows:(number|Promise<number>)[],t:[Promise<number>,number],it:Iterable<number>;export const a=Promise.race(rows),b=Promise.any(t),c=Promise.allSettled(it);`,
		`const other={all(x:unknown){return x}};export const a=other.all([1]);const alias:typeof Promise=Promise;export const b=alias['all']([0]);const key='all';export const c=Promise[key]([1]);`,
		"/* 世界 🌍 */\r\nexport async function f(){\r\n  await /* comment */ 0;\r\n}\r\n",
		`export class A {get value(){return 1;}get string():string {return 'x';}get flag(){return true;}get nil(){return null;}get negative(){return -1;}get more(){const x=1;return 1;} }`,
		`export class A {public static get x()   :   number   {return 1;} private get p() /* c */ {return 'p';} protected get q(){return false;}}`,
		`export class A {get x(){return 1;}set 'x'(v:number){} get ["y"](){return 2;}set y(v:number){} get unpaired(){return 3;}}`,
		`let p='key';const q='key';export class A {get [p](){return 1;}set [q](v:number){} get ["other"](){return 2;}}`,
		`class Base{get x(){return 0;}} export class A extends Base {get x(){return 1;}override get other(){return 2;}static get tag(){return 'x';}}`,
		`abstract class Base{abstract get x():number;}export class A extends Base{get x(){return 1;}}interface I {readonly x:number};export class B implements I {get x(){return 2;}}`,
		`declare function dec(...args:unknown[]):void;export class A {@dec get x(){return 1;}get #secret(){return 'x';}get [123](){return 4;}}`,
		"declare const tag:(parts:TemplateStringsArray)=>string;export class A{get x(){return `hi`;}get y(){return tag`hi`;}get z(){return /x/g;}get n(){return 1n;}get template(){return `x${1}`;}}",
	}
}

func wave18Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*"))
	if err != nil {
		h.t.Fatal(err)
	}
	changed := false
	for _, path := range files {
		extension := filepath.Ext(path)
		if extension != ".a" && extension != ".ts" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filepath.Base(path) == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique mutant %s", name)
			}
			source = strings.Replace(source, from, to, 1)
			changed = true
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	if !changed {
		h.t.Fatal("mutant did not change a source")
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_18_suite.a"), archive, false)
}

// Not parallel: archive and sanitizer builds share a bounded scratch filesystem.
func TestWave18AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE18_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_18_suite.a")
	binary := h.build(stage0, "wave18", entry, archive, false)
	oracle := volumeOracle(h, "wave18-oracle", "oracle_wave_18.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","lib":["ESNext"],"jsx":"preserve","noEmit":true},"sourceExtensions":[".a"],"include":["*.d.ts","*.a","*.tsx"]}`)
	h.write("anchor.d.ts", "export {};\n")
	var paths []string
	for i, source := range wave18Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"@typescript-eslint/await-thenable", "@typescript-eslint/class-literal-property-style"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("missing positive control %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave18-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, m := range []struct{ name, file, from, to string }{
		{"await", "await_thenable.a", "return !this.thenable(index, subject.id);", "return this.thenable(index, subject.id);"},
		{"class", "class_literal_property_style.a", "this.rules.byte(node.end)", "this.rules.byte(node.end) + 1"},
	} {
		mutant := wave18Mutant(h, stage0, archive, m.name, m.file, m.from, m.to)
		got := h.must(m.name+"-mutant-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("mutant %s survived", m.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, Go byte oracle caught byte %d", m.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	for _, m := range []struct{ name, file, from, to string }{
		{"iteration-facts", "bridge/tsgo/checker/iteration_type_facts.go", "out.yes(checker.Checker_getPropertyOfType(c, t, checker.Checker_getPropertyNameForKnownSymbolName(c, split[2])) != nil)", "out.yes(false)"},
		{"base-facts", "bridge/tsgo/checker/base_member_facts.go", "out.number(uint64(b.Flags))", "out.number(uint64(ast.SymbolFlagsProperty))"},
	} {
		overlay := h.overlay(m.name, m.file, m.from, m.to)
		mutantArchive := h.archive(m.name, overlay, false)
		mutant := h.build(stage0, m.name, entry, mutantArchive, false)
		got := h.must(m.name+"-mutant-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("checker mutant %s survived", m.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, Go byte oracle caught byte %d", m.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
		os.Remove(mutantArchive)
	}
	// The missing JSX grammar is measured against a positive Go rule control.
	jsx := h.write("title-positive.tsx", `import {Head} from 'next/document';export const page=<Head><title>x</title></Head>;`)
	jsxManifest := h.write("jsx.manifest", jsx+"\n")
	positive := h.must("jsx-go-positive", exec.Command(oracle, config, jsxManifest))
	if !bytes.Contains(positive.stdout, []byte("\t@next/next/no-title-in-document-head\t")) {
		t.Fatal("missing Go JSX positive control")
	}
	negative := h.run("jsx-native-gap", exec.Command(binary, config, jsxManifest))
	if code, ok := negative.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(negative.stderr, []byte("parser slice expected GreaterThanToken, got Identifier")) {
		t.Fatalf("JSX gap changed: %v %s", negative.err, negative.stderr)
	}
	t.Logf("Next rule not ported: Go positive, native JSX parser refuses with %s", strings.TrimSpace(string(negative.stderr)))
	for _, name := range []string{"repository", "compiler"} {
		variable := "ADAMIC_WAVE18_" + strings.ToUpper(name) + "_MANIFEST"
		corpusManifest := os.Getenv(variable)
		if corpusManifest == "" {
			continue
		}
		corpusConfig := filepath.Join(repository, "tsconfig.json")
		if name == "compiler" {
			corpusConfig = filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")
		}
		h.compare(name, oracle, binary, corpusConfig, corpusManifest)
		h.compare(name+"-asan", oracle, asan, corpusConfig, corpusManifest)
		for round := 0; round < 3; round++ {
			commands := []struct{ name, binary string }{{"native", binary}, {"go", oracle}}
			if round%2 == 1 {
				commands[0], commands[1] = commands[1], commands[0]
			}
			for _, command := range commands {
				cmd := exec.Command(command.binary, corpusConfig, corpusManifest, "--count")
				cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
				got := h.must(fmt.Sprintf("%s-%d-%s-timing", name, round, command.name), cmd)
				t.Logf("%s round %d %s: %.6fs %s %s", name, round, command.name, got.elapsed.Seconds(), strings.TrimSpace(string(got.stdout)), strings.TrimSpace(string(got.stderr)))
			}
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoInspect(program,file,0,1,'Identifier','raw-shape');tsgoRelease(program);console.log(tsgoInspect(program,file,0,1,'Identifier','iteration-type-facts\n1\niterator'));`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.run("released-mutant-run", exec.Command(mutant, config, probe))
	// With a retained registry the live type request incorrectly succeeds.
	if got.err != nil || len(got.stderr) != 0 {
		t.Fatal("released registry mutant survived")
	}
	t.Logf("released registry mutant caught: expected stale handle panic, got %v %s", got.err, strings.TrimSpace(string(got.stderr)))
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
