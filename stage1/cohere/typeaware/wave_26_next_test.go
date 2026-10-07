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

const wave26NextDeclarations = `declare module "node:test" {
 export class MockTracker {
  method(object:unknown,name:string,fn?:unknown):void;
  getter(object:unknown,name:string,fn?:unknown):void;
  setter(object:unknown,name:string,fn?:unknown):void;
  property(object:unknown,name:string,value?:unknown):void;
 }
 export const mock:MockTracker;
}
declare module "other" { export const mock:{method(object:unknown,name:string):void}; }
declare namespace JSX { interface IntrinsicElements { div:any; } }
`

func wave26NextControls() []string {
	controls := []string{
		`document.addEventListener('click',(event)=>{(event.target as HTMLInputElement).value;});`,
		`window.addEventListener('click',function(event){(event.target as HTMLTextAreaElement).value;});`,
		`document.addEventListener('click',(event)=>{(event.target as HTMLElement).closest('p');});`,
		`document.addEventListener('click',(event)=>{if(event.target instanceof HTMLInputElement) {(event.target as HTMLInputElement).value;}});`,
		`document.body.addEventListener('click',(event)=>{(event.target as HTMLInputElement).value;});`,
		`const document={addEventListener(n:string,f:(e:Event)=>void){}};document.addEventListener('click',(event)=>{(event.target as HTMLInputElement).value;});`,
		`function listener(event:Event){(event.target as HTMLInputElement).value;}document.addEventListener('click',listener);`,
		`const listener=(event:Event)=>{(event.target as HTMLInputElement).value;};document.addEventListener('click',listener);`,
		`let listener=(event:Event)=>{(event.target as HTMLInputElement).value;};document.addEventListener('click',listener);`,
		`document.addEventListener('click',(event)=>{(event.target as unknown as HTMLInputElement).value;(event.currentTarget as HTMLInputElement).value;});`,
		`document.addEventListener('click',(event)=>{(()=>{(event.target as HTMLInputElement).value;})();function nested(event:Event){(event.target as HTMLInputElement).value;}});`,
		`interface Custom extends HTMLElement {mine:number}document.addEventListener('click',(event)=>{(event.target as Custom).mine;});`,
		`document.addEventListener('click',(event)=>{(event.target as HTMLInputElement|HTMLTextAreaElement|null).value;});`,
		`document.addEventListener('click',(event)=>{(event.target as HTMLInputElement|HTMLElement).title;});`,
		`/* 世界 🌍 */` + "\r\n" + `document.addEventListener('click',(événement)=>{(événement.target as HTMLInputElement).value;});`,
	}
	for _, typ := range []string{"number", "bigint", "0", "0n", "1 | 2", "boolean", "number | null", "number | string", "any", "unknown", "never", "void", "0 | false | null | undefined", "number & {brand:true}"} {
		controls = append(controls, `declare const count:`+typ+`;const view=<div>{count && <div/>}</div>;`)
	}
	controls = append(controls,
		`declare const count:number,flag:boolean;const view=<div>{flag && count && <div/>}</div>;`,
		`declare const count:number,flag:boolean;const view=<div>{flag ? count && <div/> : null}</div>;`,
		`declare const count:number;const view=<div visible={count && true}>{count > 0 && <div/>}</div>;`,
		`function component<T extends number>(count:T){return <div>{count && <div/>}</div>;}`,
		`declare const count:number,other:number;const view=<div>{(count || other) && <div/>}{(count ?? other) && <div/>}</div>;`,
	)
	for _, method := range []string{"method", "getter", "setter", "property"} {
		controls = append(controls, `import {mock} from 'node:test';import * as Store from './exports';mock.`+method+`(Store,'value');`)
	}
	controls = append(controls,
		`import {mock} from 'node:test';import type * as Store from './exports';mock.method(Store,'value');`,
		`import {mock} from 'other';import * as Store from './exports';mock.method(Store,'value');`,
		`import {mock} from 'node:test';import * as Store from './exports';mock.method(Store.default,'value');`,
		`import {mock} from 'node:test';import * as Store from './exports';mock.method((Store as unknown)!,'value');`,
		`import {mock} from 'node:test';import Store from './exports';mock.method(Store,'value');`,
	)
	for i := range controls {
		controls[i] += "\nexport {};\n"
	}
	return controls
}

func wave26NextMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
					h.t.Fatalf("nonunique mutant %s", name)
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
	return h.build(stage0, name, filepath.Join(directory, "wave_26_next.a"), archive, false)
}

// Not parallel: builds, sanitizers and cost observations share a machine.
func TestWave26NextAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE26_NEXT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_26_next.a")
	binary := h.build(stage0, "wave26", entry, archive, false)
	oracle := volumeOracle(h, "wave26-oracle", "oracle_wave_26_next.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"jsx":"preserve","target":"ES2022","module":"ESNext","lib":["ES2022","DOM"],"noEmit":true},"files":["control-000.a","configured.d.ts"]}`)
	h.write("configured.d.ts", wave26NextDeclarations)
	h.write("exports.d.ts", "export const value:number; export default {value:1};\n")
	var paths []string
	for i, source := range wave26NextControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d%s", i, func() string {
			if strings.Contains(source, "<div") {
				return ".tsx"
			}
			return ".a"
		}()), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"global-listener-target-assertion", "leaked-number-render", "mock-on-module-namespace"} {
		if !bytes.Contains(truth.stdout, []byte("\tnexus/correctness-no-"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave26-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"listener", "global_listener_target_assertion.a", "['Element', 'HTMLElement', 'SVGElement', 'MathMLElement']", "['Element', 'SVGElement', 'MathMLElement']"},
		{"render", "leaked_number_render.a", "['0', '-0', 'NaN']", "['-0', 'NaN']"},
		{"mock", "mock_on_module_namespace.a", "declaration.typeOnly", "false"},
	} {
		mutant := wave26NextMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
	}
	for _, change := range []struct{ name, file, from, to string }{
		{"syntax", "bridge/tsgo/checker/node_structure.go", `operator = strings.TrimPrefix(b.OperatorToken.Kind.String(), "Kind")`, `operator = "BarBarToken"`},
		{"lineage", "bridge/tsgo/checker/declaration_lineage.go", "out.yes(p.Compiler.IsSourceFileDefaultLibrary(f.Path()))", "out.yes(false)"},
		{"literal", "bridge/tsgo/checker/literal_string.go", `text = value.String()`, `text = value.String(); if text == "0" {text = "1"}`},
		{"transform", "bridge/tsgo/checker/transformed_shape.go", "t = checker.Checker_getBaseConstraintOfType(c, t)", "t = p.typesByID[id-1]"},
	} {
		overlay := h.overlay(change.name, change.file, change.from, change.to)
		mutatedArchive := h.archive(change.name, overlay, false)
		mutant := h.build(stage0, change.name+"-question", entry, mutatedArchive, false)
		got := h.must(change.name+"-question-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s question mutant survived", change.name)
		}
		t.Logf("%s question mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutatedArchive)
		os.Remove(mutant)
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE26_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE26_COMPILER_MANIFEST")},
	} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		want := h.must(corpus.name+"-timed-go", exec.Command(oracle, corpus.config, corpus.manifest))
		native := exec.Command(binary, corpus.config, corpus.manifest)
		native.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(corpus.name+"-timed-native", native)
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatal("timed output mismatch")
		}
		t.Logf("%s whole process native %s Go %s; native phases %s; Go phases %s", corpus.name, got.elapsed, want.elapsed, got.stderr, want.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);console.log(tsgoInspect(program,file,0,1,'Identifier','raw-shape'));tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','declaration-lineage\nnode'));`)
	probe := h.write("probe.a", "x;\n")
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
	t.Log("released-registry mutant: exit 0, required panic 70 catches it")
}
