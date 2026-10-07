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

var wave02RuleNames = []string{"logical-assignment-operators", "@typescript-eslint/no-unsafe-return", "nexus/correctness-no-nullish-stripping-assertion"}

func wave02Controls() []string {
	return []string{
		"export function f(a:string|undefined,b:string){a=a||b;return a;}",
		"export function f(a:string|undefined,b:string,c:string){a=(a||b||c);a=(a||b)||c;return a;}",
		"declare const o:{p:string|undefined};declare const b:string; o.p=o.p??b; o.p||(o.p=b);",
		"declare const o:{p:{q:string|undefined}};declare const b:string; o.p.q||(o.p.q=b);o.p.q=o.p.q||b;",
		"declare const o:{p:string|undefined};declare const b:string; o.p=o['p']||b; o?.p||(o.p=b);",
		"export function f(a:string|undefined,b:string){a=a/*keep*/||b;a||(a=b);return a;}",
		"declare function call(...x:unknown[]):void;let a:string;call(a||(a='x'));call(1,a||(a='x'));const array=[a||(a='x')];const object={p:a||(a='x')};",
		"declare function use2():void;export function useThings(){use2();let a:string; a=a||'x';return a;}export function plain(){let a:string;a=a||'x';return a;}",
		"declare const anyValue:any;export function plain(){return anyValue;}export function annotated():any{return anyValue;}export function safe():unknown{return anyValue;}",
		"declare const anyRows:any[];export function array(){return anyRows;}export function arraySafe():unknown[]{return anyRows;}export function explicit():any[]{return anyRows;}",
		"declare const anyValue:any;export const f=()=>anyValue;export const safe:()=>unknown=()=>anyValue;export const generic:()=>Set<string>=()=>new Set<any>();",
		"declare const promise:Promise<any>;export async function f(){return promise;}export function sync(){return promise;}export async function safe():Promise<unknown>{return promise;}export async function explicit():Promise<any>{return promise;}",
		"declare const item:any;export async function safe():Promise<any>{return item;}export function error(){return missing;}export function empty(){}",
		"declare const items:Set<any>;export function f():Set<string>{return items;}export function noContext(){return items;}export function emptyMap():Map<string,string>{return new Map();}",
		"declare const items:Map<string,Set<any>>;export function f():Map<string,Set<number>>{return items;}",
		"declare const x:string|number|undefined;export const narrowed=x as string;export const pure=x as string|number;export const nullable=x as string|undefined;export const impossible=x as never;",
		"declare const x:'a'|'b'|null;export const y=x as 'a';export const z=x as 'a'|'b';",
		"declare const x:null|undefined;export const y=x as number;declare const v:unknown;export const z=v as string;",
		"export function f<T>(x:string|number|undefined){return x as T;}export function g<T extends string>(x:string|number|undefined){return x as T;}",
		"export function f<M,K extends keyof M>(x:string|number|undefined){return x as M[K];}",
		"declare let x:string|number|undefined;(x as string)='x';export const pure=<string|number>x;export const down=<string>x;",
		"/* 世界 🌍 */\r\nexport function f(é:any):number{return é;}declare const 漢:string|number|null;export const down=漢 as string;\r\n",
	}
}

func wave02SourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.ts"))
	additional, _ := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.a"))
	files = append(files, additional...)
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
				h.t.Fatalf("nonunique %s mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_02_suite.a"), archive, false)
}

// All output, including failures and mutants, goes to files. This suite is serial
// because sanitizer archives and corpus runs share the limited scratch volume.
func TestWave02AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE02_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_02_suite.a")
	binary := h.build(stage0, "coverage", entry, archive, false)
	oracle := volumeOracle(h, "coverage-oracle", "oracle_wave_02.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	paths := []string{}
	for i, source := range wave02Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("helper.ts", "export const Used=1,Unused=2,AlsoUnused=3;\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range wave02RuleNames {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "coverage-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Each rule mutant changes a real judgment or report range, finishes normally,
	// and is killed exclusively by the independent cohere diagnostic bytes.
	for _, change := range []struct{ name, file, from, to string }{
		{"logical", "logical_assignment_operators.a", "this.safe(this.rules.skip(node.children[0] ?? -1))", "false"},
		{"unsafe-return", "no_unsafe_return.a", "if(anyKind === 1 && !async)", "if(anyKind === 1 && async)"},
		{"nullish", "no_nullish_stripping_assertion.a", "present.includes(id)", "present.includes(id + 1)"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			mutant := wave02SourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
			got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
			if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
				t.Fatalf("%s mutant survived", change.name)
			}
			t.Logf("%s: exit 0, Go byte oracle catches byte %d; %s", change.name, firstDifference(got.stdout, truth.stdout), summary(got.stdout))
			if err := os.Remove(mutant); err != nil {
				t.Fatal(err)
			}
		})
	}
	// New raw questions are also independently held to their direct checker tests.
	if manifest := os.Getenv("ADAMIC_WAVE02_REPOSITORY_MANIFEST"); manifest != "" {
		h.compare("repository", oracle, binary, filepath.Join(repository, "tsconfig.json"), manifest)
		h.compare("repository-asan", oracle, asan, filepath.Join(repository, "tsconfig.json"), manifest)
	}
	if manifest := os.Getenv("ADAMIC_WAVE02_COMPILER_MANIFEST"); manifest != "" {
		corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
		if corpus == "" {
			t.Fatal("compiler source is required")
		}
		h.compare("compiler", oracle, binary, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
		h.compare("compiler-asan", oracle, asan, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoInspect(program,file,0,1,'Identifier','raw-shape');tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','awaited-shape\n1'));
`)
	probe := h.write("probe.ts", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released program queried using a new question: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, caught by the required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
