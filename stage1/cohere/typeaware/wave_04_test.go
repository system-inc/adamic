package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func wave04Controls() []string {
	return []string{
		`function value(b:boolean):string|undefined {if(b)return;return "x";}`,
		`function nothing():void {return undefined;}`,
		`const f:()=>string|undefined=()=>{return;};`,
		`declare function takes(f:()=>void):void;takes(()=>1);`,
		`declare function takes(f:()=>void):void;takes(()=>{return 1;});`,
		`declare function takes(f:()=>void):void;takes(async()=>{});`,
		`class Base {cb:()=>void=()=>{}}; class Sub extends Base {cb=()=>1;}`,
		`let cb:(()=>void)|null=null; cb??=()=>1;`,
		`const object:{cb:()=>void}={cb(){return 1;}};`,
		`function nested():void {function inner():undefined {return;}return undefined;}`,
		`function comments():void {return /*keep*/undefined;}`,
		`function parenthesized():void {return (undefined);}`,
		`function* gen():Generator<number,string|undefined>{yield 1;return;}`,
		`async function load():Promise<void>{return undefined;}`,
		`function local(undefined:number):void{return undefined;}`,
		`const LocalValue=1; console.log(LocalValue);`,
		`const ALL_CAPS_=1;console.log(ALL_CAPS_);`,
		"function spaces():void{return\uFEFFundefined;} function tabs():void{return\tundefined;}",
		`function create(){return ()=>1;}export const Created=create();`,
		`export const exportedValue=1;`,
		`export const ExportedFunction=()=>1;`,
		`class Host{};export const HostInstance=new Host();`,
		`const LocalValue=1;const holder={LocalValue};console.log(holder);`,
		`const LocalValue=1;const localValue=2;console.log(LocalValue,localValue);`,
		`const Long_name=1;console.log(Long_name);`,
		`const Éclair=1;console.log(Éclair);`,
		`const _LocalValue=1;console.log(_LocalValue);`,
		`const ALL_CAPS=1;const Do_=1;declare const AmbientValue:number;`,
		`const Kind={First:'First',Second:2} as const;const nativeMap=globalThis.Map;`,
		`class Rows extends Array<number>{};export const rows=new Rows();`,
		`function factory():()=>number{return ()=>1;}export const FactoryResult=factory();`,
		`const Factory=()=>1;const Alias=Factory;console.log(Alias());`,
		`const LocalValue=1;export {LocalValue as PublicValue};`,
		`declare const input:{callback:()=>void};input.callback=(x=>1);`,
		`declare function make(cb:()=>void):void;make(function():number{return 1;});`,
		`declare function make(cb:()=>void):void;make(function*(){yield 1;});`,
		`declare function choose(cb:()=>void):void;declare function choose(cb:()=>number):void;choose(()=>1);`,
		`function overload(x:number):void;function overload(x:string):string;function overload(x:any){return;}`,
		`function generic<T>():T|void{return undefined as any;} function unknown():any{return;}`,
		`declare const source:[number,()=>void];let second:()=>void;[,second]=source;`,
		"/* 世界 🌍 */\r\nexport function unicode():string|undefined {return;}\r\n",
	}
}
func wave04SourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.a"))
	if err != nil {
		h.t.Fatal(err)
	}
	relativeTS := regexp.MustCompile(`from './([^']+\.ts)'`)
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filepath.Base(path) == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique %s mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = relativeTS.ReplaceAllString(source, "from '"+filepath.Join(h.repository, "stage1/cohere/typeaware")+"/$1'")
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_04_suite.a"), archive, false)
}

// Not parallel: archive compilation, sanitizers and corpus runs share scratch space.
func TestWave04AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE04_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_04_suite.a")
	binary := h.build(stage0, "wave04", entry, archive, false)
	oracle := volumeOracle(h, "wave04-oracle", "oracle_wave_04.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for at, source := range wave04Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", at), source+"\nexport {};\n"))
	}
	helper := h.write("helper.a", "export const importedValue=1,unusedValue=2;export const ImportedFn=()=>1;\n")
	paths = append(paths, helper, h.write("consumer.a", "import {importedValue,ImportedFn} from './helper.a';console.log(importedValue,ImportedFn());\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"nexus/consistency-require-constant-casing", "nexus/consistency-require-matching-return-type", "@typescript-eslint/strict-void-return"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("missing positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave04-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"casing", "constant_casing.a", "this.rules.byte(named.end), '')", "this.rules.byte(named.end) + 1, '')"},
		{"matching", "matching_return_type.a", "verdict.kind === 2 ? 'bareReturnInValueFunction'", "verdict.kind === 2 ? 'wrongReturnInValueFunction'"},
		{"void", "strict_void_return.a", "this.rules.byte(end)))", "this.rules.byte(end) + 1))"},
	} {
		mutant := wave04SourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, Go byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err = os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE04_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE04_COMPILER_MANIFEST")},
	} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,2,'SourceFile','program-imports'));
`)
	probe := h.write("probe.a", "x;")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released handle: panic 70; registry mutant: exit 0, caught by required panic")
}
