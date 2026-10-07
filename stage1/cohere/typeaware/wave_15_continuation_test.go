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

// Not parallel: compiler builds, sanitizers and measurements share the machine.
func TestWave15ContinuationAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE15_CONT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_15_continuation_suite.a")
	binary := h.build(stage0, "wave15", entry, archive, false)
	oracle := volumeOracle(h, "wave15-oracle", "oracle_wave_15_continuation.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	nodeTypes := h.write("node.d.ts", `declare module "node:test" { interface MockTracker {method(object:object,key:string):void;getter(object:object,key:string):void;setter(object:object,key:string):void;property(object:object,key:string,value?:unknown):void;} export const mock:MockTracker; }
declare module "store" {export const value:number;}
declare module "other" {interface MockTracker {method(object:object,key:string):void;} export const mock:MockTracker;}`)
	config = h.write("tsconfig.json", fmt.Sprintf(`{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","types":[],"lib":["ES2022","DOM"]},"files":[%q]}`, nodeTypes))
	controls := []string{
		"import * as Store from 'store';import {mock} from 'node:test'; mock.method(Store,'value');mock.getter((Store),'value');mock.setter(Store!,'value');mock.property(Store as typeof Store,'value');",
		"import type * as Store from 'store';import {mock} from 'node:test';mock.method(Store,'value');",
		"import * as Store from 'store';import {mock} from 'other';mock.method(Store,'value');",
		"import * as Store from 'store';import {mock} from 'node:test';const copy={...Store};mock.method(copy,'value');function f(Store:object){mock.method(Store,'value');}",
		"/* 世界 🌍 */\r\nimport * as Store from 'store';import {mock} from 'node:test';mock.method(Store satisfies object,'value');\r\n",
		"document.addEventListener('click',event=>{const target=event.target as HTMLInputElement;});",
		"const handler=(event:Event)=>{event.target as HTMLTextAreaElement;};document.addEventListener('click',handler);",
		"function handler(event:Event){(<HTMLInputElement>event.target).value;}window.addEventListener('click',handler);",
		"document.addEventListener('click',event=>{event.target as HTMLElement;if(event.target instanceof HTMLInputElement){event.target as HTMLInputElement;}});",
		"declare const textarea:HTMLTextAreaElement;textarea.addEventListener('click',event=>{event.target as HTMLInputElement;});",
		"let handler=(event:Event)=>{event.target as HTMLInputElement;};document.addEventListener('click',handler);",
		"document.addEventListener('click',event=>{(()=>{event.target as SVGPathElement;})();const f=(event:Event)=>{event.target as HTMLInputElement;};});",
		"document.addEventListener('click',event=>{((event.target) as unknown as HTMLInputElement|HTMLTextAreaElement|null)?.value;});",
		"interface CustomElement extends HTMLElement {field:number;}document.addEventListener('click',event=>{event.target as CustomElement;});",
		"function f(window:HTMLTextAreaElement){window.addEventListener('click',event=>{event.target as HTMLInputElement;});}document.addEventListener('click',(event:{target:unknown})=>{event.target as HTMLInputElement;});",
		"/* 世界 🌍 */\r\ndocument.addEventListener('click',event=>{event.target! as HTMLInputElement;});\r\n",
	}
	paths := []string{}
	for i, s := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), s+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave15-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, module := range []string{"commonjs", "nodenext", "preserve", "es2020"} {
		otherConfig := h.write("tsconfig-"+module+".json", fmt.Sprintf(`{"compilerOptions":{"strict":true,"target":"ES2022","module":%q,"types":[],"lib":["ES2022","DOM"]},"files":[%q]}`, module, nodeTypes))
		h.compare("controls-"+module, oracle, binary, otherConfig, manifest)
	}

	for _, change := range []struct{ name, file, from, to string }{
		{"listener", "no_global_listener_target_assertion.a", "if(this.facts.compare(index, source.root().id, target.root().id))", "if(!this.facts.compare(index, source.root().id, target.root().id))"},
		{"mock", "no_mock_on_module_namespace.a", "chain.names[1] !== 'MockTracker'", "chain.names[1] !== 'OtherTracker'"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			scratch := filepath.Join(directory, change.name+"-source")
			if err := os.MkdirAll(scratch, 0755); err != nil {
				t.Fatal(err)
			}
			files, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware/*"))
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				if filepath.Ext(file) != ".a" && filepath.Ext(file) != ".ts" {
					continue
				}
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				s := string(data)
				if filepath.Base(file) == change.file {
					if strings.Count(s, change.from) != 1 {
						t.Fatal("nonunique mutant")
					}
					s = strings.Replace(s, change.from, change.to, 1)
				}
				s = strings.ReplaceAll(s, "../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
				s = strings.ReplaceAll(s, "../lint/", filepath.Join(repository, "stage1/cohere/lint")+"/")
				h.write(filepath.Join(change.name+"-source", filepath.Base(file)), s)
			}
			mutant := h.build(stage0, change.name+"-mutant", filepath.Join(scratch, "wave_15_continuation_suite.a"), archive, false)
			got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
			if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
				t.Fatal("mutant survived")
			}
			t.Logf("%s: exit 0, independent Go bytes catch byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		})
	}
	for _, corpus := range []string{"repository", "compiler"} {
		manifest := os.Getenv("ADAMIC_WAVE15_" + strings.ToUpper(corpus) + "_MANIFEST")
		if manifest == "" {
			continue
		}
		config := filepath.Join(repository, "tsconfig.json")
		if corpus == "compiler" {
			config = filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")
		}
		h.compare(corpus, oracle, binary, config, manifest)
		h.compare(corpus+"-asan", oracle, asan, config, manifest)
		goResult := h.must(corpus+"-timed-go", exec.Command(oracle, config, manifest))
		command := exec.Command(binary, config, manifest)
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		native := h.must(corpus+"-timed-native", command)
		if !bytes.Equal(goResult.stdout, native.stdout) {
			t.Fatal("timed bytes differ")
		}
		t.Logf("%s native=%s Go=%s; %s %s", corpus, native.elapsed, goResult.elapsed, native.stderr, goResult.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','call-declaration-chain'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released handle: panic 70, invalid or released checker handle")
}
