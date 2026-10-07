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

func wave03Controls() []string {
	return []string{
		"declare const x:any;x();new x();x`tag`; (x)(); new (x)();(x)`tag`;",
		"declare const unresolved:NotKnown;unresolved();new unresolved();unresolved`tag`;",
		"declare const f:Function;f();new f();f`tag`;interface VoidCall extends Function{():void};declare const v:VoidCall;v();new v();v`tag`;",
		"interface Returns extends Function{():string};declare const r:Returns;r();new r();interface Constructs extends Function{new():object};declare const c:Constructs;c();new c();c`tag`;",
		"type Function=()=>void;declare const local:Function;local();declare const n:number;`${n}`;",
		"import { Shape, value, Unused } from './helper.js';export const x:Shape={x:1};export {value};",
		"import { Shape } from './helper.js';export type T=Shape;import type { Other } from './helper.js';export type U=Other;",
		"import { Shape, type Other } from './helper.js';export type T=Shape;export type U=Other;",
		"import Default, { Shape, value, Other } from './helper.js';export type T=Default;export type U=Shape;export const v=value;export type W=Other;",
		"import Default, * as NS from './helper.js';export type T=Default;export type W=NS.Shape;",
		"import * as NS from './helper.js';export type T=NS.Shape;export type U=import('./helper.js').Shape;",
		"import {Shape as S} from './helper.js';export type T=S;export function f(){const S=1;return S;}",
		"import {Shape} from './helper.js';export {type Shape};",
		"declare const s:string;export const a=`${s}`;export const b=`${s||'x'}`.length;export const c=`${s}${s}`;",
		"export const a=`head${'mid'}tail`;export const b=`${1}${true}${null}${undefined}${Infinity}${NaN}`;",
		"export const a=`${'    '}\n`;export const b=`${/*keep*/'x'}`;export const c=`${'x'/*keep*/}`;",
		"export const a=`$${'{'}${'$'}{`;export const b=`${`$`}${'{'}`;export const c=`${'`'}${'${'}`;",
		"export const a=`${0o25}${0b1010}${1_0}${1e21}${10n}${/x\\d/}`;",
		"type A=`${'x'}`;type B=`${number}`;type C<S extends string>=`${S}`;enum E{X='x'};type D=`${E.X}`;",
		"/* 世界 🌍 */\r\ndeclare const é:any;é();export const a=`${    '漢'    }`;\r\n",
	}
}

func wave03Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*"))
	if err != nil {
		h.t.Fatal(err)
	}
	for _, path := range files {
		if filepath.Ext(path) != ".a" && filepath.Ext(path) != ".ts" {
			continue
		}
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
	return h.build(stage0, name, filepath.Join(directory, "wave_03_suite.a"), archive, false)
}

// Not parallel: sanitizer builds and whole-corpus timings share scratch and CPUs.
func TestWave03AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE03_ARTIFACTS"); path != "" {
		directory = path
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_03_suite.a")
	binary := h.build(stage0, "wave03", entry, archive, false)
	oracle := volumeOracle(h, "wave03-oracle", "oracle_wave_03.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range wave03Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("helper.ts", "export interface Shape{x:number};export interface Other{y:number};export default interface Default{z:number};export const value=1,Unused=2;\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-unsafe-call", "consistent-type-imports", "no-unnecessary-template-expression"} {
		if !bytes.Contains(truth.stdout, []byte("\t@typescript-eslint/"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}

	looseConfig := h.write("loose.json", `{"compilerOptions":{"strict":true,"noImplicitThis":false,"target":"ES2022","lib":["ES2022"]}}`)
	loose := h.write("loose.ts", "export function f(){this.method();this.tag`x`;new this.Constructor();}export function g(this:NotKnown){this.method();}\n")
	looseManifest := h.write("loose.manifest", loose+"\n")
	h.compare("loose-this", oracle, binary, looseConfig, looseManifest)
	if os.Getenv("ADAMIC_WAVE03_QUICK") != "" {
		return
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave03-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"unsafe-call", "no_unsafe_call.a", "signatures.constructs > 0", "signatures.constructs > 1000000"},
		{"imports", "consistent_type_imports.a", "found = true;", "found = false;"},
		{"template", "no_unnecessary_template_expression.a", "this.rules.start(expression) - 2", "this.rules.start(expression) - 1"},
	} {
		mutant := wave03Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s: exit 0, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}

	metadataConfig := h.write("metadata.json", `{"compilerOptions":{"strict":true,"experimentalDecorators":true,"emitDecoratorMetadata":true,"target":"ES2022","lib":["ES2022"]}}`)
	metadataSource := h.write("metadata.ts", "export const value=1;\n")
	metadataManifest := h.write("metadata.manifest", metadataSource+"\n")
	refused := h.run("metadata-refusal", exec.Command(binary, metadataConfig, metadataManifest))
	if code, ok := refused.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(refused.stderr) != "adamic: panic: wave 03 requires emitDecoratorMetadata disabled; metadata imports are not yet ported\n" {
		t.Fatalf("metadata guard escaped: %v %s", refused.err, refused.stderr)
	}
	guard := wave03Mutant(h, stage0, archive, "metadata-guard", "import_runtime_options.a", "if(metadata)", "if(metadata && factory === 'impossible')")
	h.must("metadata-guard-run", exec.Command(guard, metadataConfig, metadataManifest))
	t.Log("metadata guard mutant exits 0, caught by required panic 70")
	if err := os.Remove(guard); err != nil {
		t.Fatal(err)
	}
	for _, population := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE03_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE03_COMPILER_MANIFEST")},
	} {
		if population.manifest == "" {
			continue
		}
		h.compare(population.name, oracle, binary, population.config, population.manifest)
		h.compare(population.name+"-asan", oracle, asan, population.config, population.manifest)
		want := h.must(population.name+"-timed-go", exec.Command(oracle, population.config, population.manifest))
		command := exec.Command(binary, population.config, population.manifest)
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(population.name+"-timed-native", command)
		if !bytes.Equal(got.stdout, want.stdout) {
			t.Fatalf("timed %s mismatch", population.name)
		}
		t.Logf("%s timing: native %s Go %s; native %s Go %s", population.name, got.elapsed, want.elapsed, got.stderr, want.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','node-binding'));
`)
	probe := h.write("probe.ts", "x;\n")
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
	t.Log("released registry mutant exits 0 and is caught by required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
