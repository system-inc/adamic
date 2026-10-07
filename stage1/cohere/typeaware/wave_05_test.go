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

// Not parallel: archive builds, sanitizers and corpus timings share a machine.
func TestWave05AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE05_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_05.a")
	binary := h.build(stage0, "wave05", entry, archive, false)
	oracle := volumeOracle(h, "wave05-oracle", "oracle_wave_05.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	sources := []string{
		"declare const object:{}; export const s=`${object}`;",
		"declare const u:unknown; export const s=`${u}`;",
		"declare const a:string[]; export const s=`${a}`;",
		"declare const n:never; export const s=`${n}`;",
		"declare const a:any,b:boolean,n:number,bi:bigint,z:null,u:undefined,r:RegExp; export const s=`${a}${b}${n}${bi}${z}${u}${r}`;",
		"declare const x:string|{}; export const s=`${x}`;",
		"declare const x:string & {brand:1}; export const s=`${x}`;",
		"export const s=`${new Error('a')}`; class Derived extends Error {} class Twice extends Derived {} export const t=`${new Twice()}`;",
		"class Error {} export const s=`${new Error()}`;",
		"declare function tag(parts:TemplateStringsArray,...args:unknown[]):string; export const s=tag`${{}}`;",
		"export function constrained<T extends string,U extends {}>(t:T,u:U){return `${t}${u}`;}",
		"/* 世界 🌍 */\r\ndeclare const é:unknown;export const s=`x${ /* hello */ é }y`;\r\n",
		"export function once<T>(x:T):void {};export function identity<T>(x:T):T{return x};",
		"declare function unused<T>(x:string):void;declare function assertion<T>(x:string):T;",
		"export const witness=<T,>():T=>null as T;export class Brand<T>{declare private value:T;};",
		"export const onceArray=<T,>(x:T[])=>null;export const returnArray=<T,>(x:string):T[]=>[];",
		"export function dead(c:boolean){let v=1;console.log(v);v=2;} export function conditional(c:boolean){let v=1;if(c){v=2;return;}console.log(v);}",
		"export function loop(c:boolean){let v=1;while(c){console.log(v);v=2;}console.log(v);v=3;}",
		"export function shorthand(){let v=1;const o={v};v=2;return o;}export function captured(){let v=1;const f=()=>v;v=2;return f;}",
		"export function caught(){let v=1;try{unknownCall();v=2;}catch{}return v;}declare function unknownCall():void;",
		"export function finallyRead(){let v;try{}finally{v=2;}console.log(v);}export function updates(){let v=1;console.log(v);v++;}",
		"export function destructured(){let [a,b]=[1,2];a=3;console.log(a,b);}",
		"export function ternary(c:boolean){let v=1;console.log(v);v=2;return c?0:v;}export function assign(){let a=1,b=2;[a,b]=[3,4];return a+b;}",
		"type Alias = RegExp; declare const x:Alias; export const s=`${x}`;",
	}
	paths := []string{}
	for i, source := range sources {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("\trestrict-template-expressions\t")) && !bytes.Contains(truth.stdout, []byte("\t@typescript-eslint/restrict-template-expressions\t")) {
		t.Fatal("no positive control")
	}
	asanArchive := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave05-asan", entry, asanArchive, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"template", "restrict_template_expressions.a", "if((subject.flags & stringLike) !== 0)", "if((subject.flags & (stringLike | 2)) !== 0)"},
		{"parameter", "no_unnecessary_type_parameters.a", "count > 2", "count > 1"},
		{"dead-store", "no_useless_assignment.a", "dead &&", "!dead &&"},
	} {
		mutantDirectory := filepath.Join(directory, change.name+"-source")
		if err := os.MkdirAll(mutantDirectory, 0755); err != nil {
			t.Fatal(err)
		}
		files, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware/*.a"))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			original, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			text := string(original)
			if filepath.Base(file) == change.file {
				if strings.Count(text, change.from) != 1 {
					t.Fatal("nonunique mutant")
				}
				text = strings.Replace(text, change.from, change.to, 1)
			}
			text = strings.ReplaceAll(text, "'../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")
			dependencies, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware/*.ts"))
			if err != nil {
				t.Fatal(err)
			}
			for _, dependency := range dependencies {
				text = strings.ReplaceAll(text, "'./"+filepath.Base(dependency)+"'", "'"+dependency+"'")
			}
			if err := os.WriteFile(filepath.Join(mutantDirectory, filepath.Base(file)), []byte(text), 0644); err != nil {
				t.Fatal(err)
			}
		}
		mutant := h.build(stage0, change.name+"-mutant", filepath.Join(mutantDirectory, "wave_05.a"), archive, false)
		got := h.must(change.name+"-mutant-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0; independent Go byte comparison catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE05_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE05_COMPILER_MANIFEST")},
	} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','type-shape'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	r := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := r.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(r.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", r.err, r.stderr)
	}
	t.Log("released handle: exit 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released program.")
	registryArchive := h.archive("registry-mutant", overlay, false)
	registry := h.build(stage0, "released-registry", released, registryArchive, false)
	survived := h.must("released-registry-run", exec.Command(registry, config, probe))
	if len(survived.stderr) != 0 {
		t.Fatalf("released registry mutant did not exit normally: %s", survived.stderr)
	}
	t.Log("released registry mutant: exits 0; required exit 70 catches it")
	if os.Getenv("ADAMIC_WAVE05_BENCH") != "" {
		for _, corpus := range []struct{ name, config, manifest string }{
			{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE05_REPOSITORY_MANIFEST")},
			{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE05_COMPILER_MANIFEST")},
		} {
			if corpus.manifest == "" {
				continue
			}
			for round := 0; round < 3; round++ {
				implementations := []string{oracle, binary}
				if round%2 == 1 {
					implementations = []string{binary, oracle}
				}
				for _, implementation := range implementations {
					name := "native"
					if implementation == oracle {
						name = "go"
					}
					command := exec.Command(implementation, corpus.config, corpus.manifest, "--count")
					command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
					measured := h.must(fmt.Sprintf("%s-%s-round-%d", corpus.name, name, round), command)
					t.Logf("%s %s round %d: %.6f seconds; %s; %s", corpus.name, name, round, measured.elapsed.Seconds(), summary(measured.stdout), strings.TrimSpace(string(measured.stderr)))
				}
			}
		}
	}
}
