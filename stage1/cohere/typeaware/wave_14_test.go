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

func wave14Controls() []string {
	return []string{
		"declare const rows:number[];delete rows[0];delete ((rows[(console.log('x'), 1)]));",
		"/* 世界 🌍 */\r\ndeclare const rows:readonly [number,string];delete /*a*/ rows /*b*/ [ /*c*/ 0 /*d*/ ];\r\n",
		"declare const rows:number[]|string[];delete rows[0];declare const mixed:number[]|{x:number};delete mixed[0];declare const intersect:number[]&{tag:1};delete intersect[0];",
		"export function del<T extends number[]>(x:T){delete x[0];}declare const anyValue:any;delete anyValue[0];declare const obj:{x?:number};delete obj['x'];delete rows.length;declare const rows:number[];",
		"declare const obj:{};String(obj);`${obj}`;obj.toString();obj.toLocaleString();''+obj;obj+'';",
		"declare const rows:{}[];rows.join(',');declare const tuple:[string,{}];tuple.join();declare const useful:string[];useful.join();",
		"declare const union:{}|string;`${union}`;declare const intersection:{}&{toString():string};`${intersection}`;",
		"export function stringify<T extends {}>(x:T){return `${x}`;}export function unknown<T>(x:T){return `${x}`;}declare const value:unknown;String(value);",
		"class ErrorChild<T> extends Error {}declare const err:ErrorChild<number>;String(err);declare const ignored:URL;String(ignored);interface Primitive {[Symbol.toPrimitive]():string};declare const p:Primitive;String(p);",
		"type Recursive=Recursive[];declare const recursive:Recursive;String(recursive);const String=(x:{})=>x;String({});",
		"String(({}));({}).toString();declare const tag:any;tag`${{}}`;String(...[{}]);",
		"class Empty{};class Ctor{constructor(){}};class Static{static x=1};class Real{x=1};class Param{constructor(public x:number){}};class Sub extends Empty{};",
		"interface I{};class Implements implements I{};abstract class Abstract{abstract x:number};const Named=class Name{static f(){}};export default class{static x=1};",
		"class Token{};function register(x:new()=>unknown){}register(Token);class Unused{};class Used{};console.log(Used);",
		"const Token=class Named{};function register(x:new()=>unknown){}register(Token);register(class{});class Generic{};function generic<T>(x:T){}generic(Generic);",
		"class Exported{};export {Exported};export class Direct{};class Base{};class Child extends Base{};let Reassigned=class{};Reassigned=class{};",
		"class Both{constructor(){}static x=1};class Block{static{}};class Semicolon{;};class Index{[x:string]:unknown};class Accessor{static get x(){return 1}};",
	}
}

func wave14Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, pattern := range []string{"*.a"} {
		files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware", pattern))
		if err != nil {
			h.t.Fatal(err)
		}
		for _, path := range files {
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
			for _, unchanged := range []string{"rules", "frames", "facts", "shadow", "diagnostic", "unary_minus", "suggestion", "repair"} {
				source = strings.ReplaceAll(source, "./"+unchanged+".ts", filepath.Join(h.repository, "stage1/cohere/typeaware", unchanged+".ts"))
			}
			source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
			source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
			if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_14.a"), archive, false)
}

// Not parallel: corpus comparisons and sanitizer archives share scratch space.
func TestWave14AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE14_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_14.a")
	binary := h.build(stage0, "wave-14", entry, archive, false)
	oracle := volumeOracle(h, "wave-14-oracle", "oracle_wave_14.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range wave14Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-array-delete", "no-base-to-string", "no-extraneous-class"} {
		if !bytes.Contains(truth.stdout, []byte("\t@typescript-eslint/"+name+"\t")) {
			t.Fatal("missing positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-14-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"delete", "no_array_delete.a", "if(!underlying)", "if(underlying)"},
		{"stringify", "no_base_to_string.a", "return metadata.parentKinds.length > 0 ? 'will' : 'always';", "return metadata.parentKinds.length > 0 ? 'may' : 'always';"},
		{"class", "no_extraneous_class.a", "valueUses === slots", "valueUses !== slots"},
	} {
		mutant := wave14Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE14_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE14_COMPILER_MANIFEST")},
	} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','reference-context'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released handle: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, required panic catches it")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
