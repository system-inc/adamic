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

func wave24Controls() []string {
	return []string{
		"export class Foo { f(): Foo { return this; } g(): Foo { return new Foo(); } h(): this {return this;} }",
		"export class Foo { f = (): Foo => this; g = (): Foo => (this); h = function(): Foo { return this; }; }",
		"export class Foo { f(): Foo | undefined { if(Math.random()) return; return this; } g(): Foo {const self=this;return self;} }",
		"export class Base{};export class Derived extends Base { f():Base{return this;} }",
		"export class Foo { f(): Foo {if(Math.random())return new Foo();return this;} g(): Foo & {} {return this;} }",
		"export class Foo { f(): (Foo|undefined)|null {return this;} g(this:Foo):Foo{return this;} static h():Foo{return (this);} }",
		"export const Named=class Bar {f():Bar{return this;}};export class Foo{get f():Foo{return this;} g():Foo{()=>{return this;};return this;}}",
		"declare const p:Promise<number>;export async function ordinary(){return await p;} export async function primitive(){return await 1;}",
		"declare const p:Promise<number>;export async function handled(){try{return p;}catch{ return 0; }}",
		"declare const p:Promise<number>;export async function f(){try{return await p;}catch{return p;}finally{}}",
		"declare const p:Promise<number>;export async function f(){try{return p;}finally{return await p;}}",
		"export const f=async()=>await /* keep */ 1;export const g=async()=>await {a:1}.a;export const h=async()=>((await 1));",
		"declare const p:Promise<number>;export async function f(){return Math.random()?await p:await 1;}",
		"export async function f<T>(p:T){return await p;}export async function g(p:any){return await p;}export async function h(p:unknown){return await p;}",
		"declare const p:Promise<number>;export async function f(){using resource=null;return p;} export async function g(){const resource=null;return await p;}",
		"declare const p:Promise<number>;p.catch(error=>{});p.catch((error:any)=>{});p.catch((error:unknown)=>{});p.then(undefined,(error:Error)=>{});",
		"declare const p:Promise<number>;p.catch(([error])=>{});p.catch(({message})=>{});p.catch((...error:any[])=>{});p.catch((...error:[unknown])=>{});",
		"declare const p:Promise<number>;p.catch((error=1)=>{});p.catch((error?:any)=>{});p.catch((...error)=>{});p.catch(()=>{});",
		"declare const p:Promise<number>;p.catch(Math.random()?err=>{}:(err:any)=>{});p.catch((null,err=>{}));p.catch((err=>{}) as (e:any)=>void);",
		"declare const p:Promise<number>;const key='catch';p[key](err=>{});p['catch'](err=>{});p[`catch`](err=>{});let widened='catch';p[widened](err=>{});",
		"declare const p:Promise<number>;const fn=(err:any)=>{};p.catch(fn);p.catch(...[fn]);p.then(...[undefined,fn]);({catch:(f:(e:any)=>void)=>{}}).catch(err=>{});",
		"export class Foo {f():Foo|string{if(Math.random())return this;const value:Foo|string=Math.random()?new Foo():'x';return value;}g():Foo|string{if(Math.random())return this;const value:number|string=Math.random()?1:'x';return value;}}",
		"declare const p:{then:(...callbacks:((x:number)=>void)[])=>unknown;catch:(f:(e:any)=>void)=>unknown};p.catch(err=>{});export async function f(){return await p;}",
		"/* 世界 🌍 */\r\nexport class 漢 {f():漢 {return this;}}\r\ndeclare const p:Promise<number>;p.catch(é=>{});\r\n",
	}
}

func wave24Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, extension := range []string{"*.ts", "*.a"} {
		paths, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware", extension))
		if err != nil {
			h.t.Fatal(err)
		}
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				h.t.Fatal(err)
			}
			source := string(data)
			if filepath.Base(path) == file {
				if strings.Count(source, from) != 1 {
					h.t.Fatal("nonunique mutant", name)
				}
				source = strings.Replace(source, from, to, 1)
			}
			source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
			source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
			if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0644); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_24_suite.a"), archive, false)
}

// Not parallel: corpus timings and sanitizer builds share this machine.
func TestWave24AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE24_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_24_suite.a")
	native := h.build(stage0, "wave24", entry, archive, false)
	oracle := volumeOracle(h, "wave24-oracle", "oracle_wave_24.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range wave24Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, native, config, manifest)
	for _, name := range []string{"prefer-return-this-type", "return-await", "use-unknown-in-catch-callback-variable"} {
		if !bytes.Contains(truth.stdout, []byte("\t@typescript-eslint/"+name+"\t")) {
			t.Fatal("missing positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave24-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"receiver", "prefer_return_this_type.a", "!returnedClass && returnedThis", "returnedClass || returnedThis"},
		{"await", "return_await.a", "handling && !awaited", "!handling && !awaited"},
		{"unknown", "use_unknown_in_catch_callback_variable.a", "(parameter.flags & 2) === 0", "(parameter.flags & 2) !== 0"},
	} {
		mutant := wave24Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", change.name)
		}
		t.Logf("%s mutant exits 0, empty stderr; byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	for _, change := range []struct{ name, path, from, to string }{
		{"class-question", "bridge/tsgo/checker/class_this_types.go", "out.number(g.id(receiver))", "out.number(g.id(receiver) + 1)"},
		{"then-question", "bridge/tsgo/checker/then_callback_signatures.go", "out.ids(counts)", "out.ids(nil)"},
		{"handler-question", "bridge/tsgo/checker/handler_parameter_types.go", "out.number(flags)", "out.number(flags | uint64(checker.TypeFlagsUnknown))"},
	} {
		overlay := h.overlay(change.name, change.path, change.from, change.to)
		mutantArchive := h.archive(change.name, overlay, false)
		mutant := h.build(stage0, change.name+"-native", entry, mutantArchive, false)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("question mutant survived", change.name)
		}
		t.Logf("%s mutant exits 0, empty stderr; byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutantArchive); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}

	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE24_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE24_COMPILER_MANIFEST")},
	} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name, oracle, native, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		for _, item := range []struct{ name, binary string }{{"go", oracle}, {"native", native}} {
			got := h.must(corpus.name+"-timing-"+item.name, exec.Command(item.binary, corpus.config, corpus.manifest))
			t.Logf("%s %s whole process %s: %s", corpus.name, item.name, got.elapsed, summary(got.stdout))
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,4,'ArrowFunction','handler-parameter-types'));`)
	probe := h.write("probe.a", "x=>x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released handle: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, caught by the required panic 70")
	if err := os.Remove(mutantArchive); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(mutant); err != nil {
		t.Fatal(err)
	}
}
