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

func wave14NextControls() []string {
	return []string{
		"document.addEventListener('click',event=>{(event.target as HTMLInputElement).value;});window.addEventListener('keydown',function(event){(event.target as HTMLTextAreaElement).selectionStart;});",
		"/* 世界 🌍 */\r\ndocument.addEventListener('click',event=>((event.target as unknown) as HTMLInputElement).value);\r\n",
		"function named(event:KeyboardEvent){(<HTMLInputElement>event.target).value;}document.addEventListener('keydown',named);const arrow=(event:Event)=>{(event.target as HTMLInputElement).value;};window.addEventListener('click',arrow);",
		"document.addEventListener('click',event=>{(event.target as HTMLInputElement|HTMLTextAreaElement|null).value;(event.target as HTMLInputElement|HTMLElement).tagName;(event.target as HTMLInputElement|{}).value;});",
		"document.addEventListener('click',event=>{if(event.target instanceof HTMLInputElement){(event.target as HTMLInputElement).value;}(event.target as HTMLElement).closest('x');(event.currentTarget as Document).title;});",
		"document.addEventListener('click',event=>{(()=>{(event.target as HTMLInputElement).value;})();function nested(event:Event){(event.target as HTMLInputElement).value;}});",
		"declare const textarea:HTMLTextAreaElement;textarea.addEventListener('click',event=>{(event.target as HTMLInputElement).value;});function local(document:{addEventListener(name:string,callback:(event:Event)=>void):void}){document.addEventListener('click',event=>{(event.target as HTMLInputElement).value;});}",
		"interface Custom extends HTMLInputElement{};document.addEventListener('click',event=>{(event.target as Custom).value;});let mutable=(event:Event)=>{(event.target as HTMLInputElement).value;};document.addEventListener('click',mutable);",
		"import * as Store from 'store';import {mock} from 'node:test';mock.method(Store,'load');mock.getter(Store,'version');mock.setter(Store,'version');mock.property(Store,'version',2);",
		"import * as Store from 'store';import {mock} from 'node:test';/* 世界 🌍 */\r\nmock.method(((Store as typeof Store)!), 'load');mock.method(Store satisfies typeof Store,'load');mock.method(<typeof Store>Store,'load');\r\n",
		"import * as Store from 'store';import type * as Types from 'store';import {mock} from 'node:test';declare const copy:typeof Store;mock.method(copy,'load');mock.method(Types,'load');mock.method(Store.default,'load');function local(Store:object){mock.method(Store,'load');}",
		"import * as Store from 'store';interface MockTracker{method(object:object,name:string):void};declare const mock:MockTracker;mock.method(Store,'load');",
		"import * as Store from 'store';import {mock} from 'other-mock';mock.method(Store,'load');",
		"import * as Store from 'store';import {tracker} from 'test';tracker.property(Store,'version');",
	}
}

func wave14NextMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_14_next.a"), archive, false)
}

// Not parallel: corpus comparisons and sanitizer archives share scratch space.
func TestWave14NextAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE14_NEXT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_14_next.a")
	binary := h.build(stage0, "wave-14-next", entry, archive, false)
	oracle := volumeOracle(h, "wave-14-next-oracle", "oracle_wave_14_next.go")
	node := h.write("node.a", `declare module "node:test" {export interface MockTracker {method(object:object,name:string):void;getter(object:object,name:string):void;setter(object:object,name:string):void;property(object:object,name:string,value?:unknown):void;}export const mock:MockTracker;}
declare module "store" {export function load():string;export const version:number;const value:{load():string};export default value;}
declare module "other-mock" {export interface MockTracker{method(object:object,name:string):void;}export const mock:MockTracker;}
declare module "test" {export namespace nested {export class MockTracker{property(object:object,name:string):void;}}export const tracker:nested.MockTracker;}
`)
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","lib":["ES2022","DOM"]},"files":["node.a"]}`)
	paths := []string{node}
	for i, source := range wave14NextControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"correctness-no-global-listener-target-assertion", "correctness-no-mock-on-module-namespace"} {
		if !bytes.Contains(truth.stdout, []byte("\tnexus/"+name+"\t")) {
			t.Fatal("missing positive control", name)
		}
	}
	for _, module := range []string{"CommonJS", "NodeNext", "Preserve", "ES2020"} {
		alternative := h.write("tsconfig-"+module+".json", strings.ReplaceAll(`{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","lib":["ES2022","DOM"]},"files":["node.a"]}`, "ESNext", module))
		h.compare("module-"+module, oracle, binary, alternative, manifest)
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-14-next-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"listener", "no_global_listener_target_assertion.a", "return assignable ? -1 : outer;", "return assignable ? outer : -1;"},
		{"namespace", "no_mock_on_module_namespace.a", "tracker = ['node:test', 'test'].includes(signature.names[at] ?? '');", "tracker = ['other'].includes(signature.names[at] ?? '');"},
	} {
		mutant := wave14NextMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
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
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE14_NEXT_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE14_NEXT_COMPILER_MANIFEST")},
	} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
	jsx := h.write("jsx-witness.a", "declare const count:number;export const view=<p>{count&&'some'}</p>;\n")
	jsxManifest := h.write("jsx-witness.manifest", jsx+"\n")
	jsxResult := h.run("jsx-native-refusal", exec.Command(binary, config, jsxManifest))
	if code, ok := jsxResult.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(jsxResult.stderr, []byte("parser slice expected CloseBraceToken, got AmpersandAmpersandToken")) {
		t.Fatalf("JSX boundary changed: %v %s", jsxResult.err, jsxResult.stderr)
	}
	virtual := filepath.Join(repository, "cohere/internal/lint/rules/nexus/adamic_wave14_jsx_test.go")
	witnessSource := filepath.Join(repository, "stage1/cohere/typeaware/testdata/jsx_wave_14_next_test.go")
	witnessOverlay := h.write("jsx-overlay.json", fmt.Sprintf("{\"Replace\":{%q:%q}}", virtual, witnessSource))
	witness := exec.Command("go", "test", "-overlay", witnessOverlay, "./internal/lint/rules/nexus", "-run", "^TestAdamicWave14JSXWitness$", "-count=1", "-v", "-timeout=30m")
	witness.Dir = filepath.Join(repository, "cohere")
	witness.Env = append(os.Environ(), "ADAMIC_WAVE14_JSX_WITNESS="+jsx)
	oracleWitness := h.must("jsx-go-positive", witness)
	if !bytes.Contains(oracleWitness.stdout, []byte("leakedNumberRender")) {
		t.Fatal("missing Go-positive JSX witness")
	}
	t.Log("leaked-number-render: Go reports the .a witness as virtual TSX; native refuses parsing with panic 70, so this rule remains unported")
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','symbol-context'));
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
