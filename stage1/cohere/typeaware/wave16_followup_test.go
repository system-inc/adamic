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

var wave16FollowupRuleNames = []string{"nexus/correctness-no-global-listener-target-assertion", "nexus/correctness-no-mock-on-module-namespace", "nexus/correctness-no-leaked-number-render"}

func wave16FollowupControls() []string {
	return []string{
		"document.addEventListener('click',e=>{(e.target as HTMLInputElement).value; (e.target as HTMLElement).title; (e.currentTarget as HTMLInputElement).value;});",
		"window.addEventListener('click',function(e){const t=(e.target as unknown as HTMLTextAreaElement);function inner(e:Event){(e.target as HTMLInputElement).value;}setTimeout(()=>{(e.target! as HTMLSelectElement).value;});});",
		"const handler=(e:Event)=>{(e.target as HTMLInputElement|HTMLSelectElement|null).value;};document.addEventListener('click',handler);",
		"function handler(e:Event){if(e.target instanceof HTMLInputElement){(e.target as HTMLInputElement).value;}(e.target as SVGCircleElement).cx;}window.addEventListener('click',handler);",
		"declare const el:HTMLInputElement;el.addEventListener('click',e=>{(e.target as HTMLInputElement).value;});function f(document:{addEventListener:(s:string,h:(e:Event)=>void)=>void}){document.addEventListener('click',e=>{(e.target as HTMLInputElement).value;});}",
		"let handler=(e:Event)=>{(e.target as HTMLInputElement).value;};document.addEventListener('click',handler);",
		"interface Custom extends HTMLElement {x:number}document.addEventListener('click',e=>{(e.target as Custom).x;(e.target as HTMLInputElement|Custom).x;});",
		"import * as ns from './helper.a';import {mock} from 'node:test';mock.method(ns,'value');mock.getter((ns),'value');mock.setter(ns!,'value');mock.property(ns as unknown,'value');mock.method(ns satisfies object,'value');",
		"import * as ns from './helper.a';import {mock} from 'node:test';mock.method(ns.default,'value');const copy=ns;mock.method(copy,'value');function f(ns:object){mock.method(ns,'value');}const fake={method(o:object,k:string){}};fake.method(ns,'value');",
		"import type * as ns from './helper.a';import {mock} from 'node:test';mock.method(ns,'value');",
		"/* 世界 🌍 */\r\ndocument.addEventListener('click',e=>{(e.target as HTMLTextAreaElement).value;});\r\n",
	}
}

func wave16FollowupSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.a"))
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
		source = regexp.MustCompile(`'\./([^']+\.ts)'`).ReplaceAllString(source, "'"+filepath.Join(h.repository, "stage1/cohere/typeaware")+"/$1'")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave16_followup_suite.a"), archive, false)
}

// All output, including failures and mutants, goes to files. This suite is serial
// because sanitizer archives and corpus runs share the limited scratch volume.
func TestWave16FollowupAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE16_FOLLOWUP_ARTIFACTS"); path != "" {
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
	// The shared dispatch source stays untouched. The overlay tests the exact
	// integration patch while its registration remains pending upstream.
	registration := h.overlay("questions-registration", "bridge/tsgo/checker/facts.go", "\tout.text(mode)\n", "\tout.text(mode)\n\tif answer, handled, err := p.additionalAnswer(out, c, node, question); handled { return answer, err }\n")
	archive := h.archive("checker", registration, false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave16_followup_suite.a")
	binary := h.build(stage0, "wave16", entry, archive, false)
	oracle := volumeOracle(h, "wave16-oracle", "oracle_wave16_followup.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","moduleResolution":"Bundler","jsx":"preserve","lib":["ES2022","DOM"]},"files":["node.d.a"]}`)
	h.write("node.d.a", `declare module "node:test" { interface MockTracker {method(o:unknown,k:string):void;getter(o:unknown,k:string):void;setter(o:unknown,k:string):void;property(o:unknown,k:string):void;} export const mock:MockTracker;}`)
	paths := []string{filepath.Join(directory, "node.d.a")}
	for i, source := range wave16FollowupControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("helper.a", "export let value=1; export const object={x:1}; export default object;\n"))
	jsx := []string{
		"declare const count:number, optional:number|undefined,big:bigint,flag:boolean,label:string; const view=<p>{count && <Badge/>}{flag && optional && 'x'}{(flag ? big : false) && 'x'}{label || count && 'x'}{(count || flag) && 'x'}</p>;",
		"declare const count:number,flag:boolean;const view=<><Badge visible={count&&true}/><Badge visible={<p>{count && 'x'}</p>}/>{...(count && [])}{flag ? count && 'x' : null}</>;",
		"declare const zero:0|1,one:1|2,other:string|number,unknown:unknown;type Brand=number&{readonly unit:'x'};declare const branded:Brand;const view=<p>{zero&&'x'}{one&&'x'}{other&&'x'}{unknown&&'x'}{branded&&'x'}</p>;",
		"function View<T extends number>(p:{value:T}){return <p>{p.value&&'x'}</p>;} function Unknown<T>(p:{value:T}){return <p>{p.value&&'x'}</p>;}",
		"enum E{Off,On};enum F{One=1,Two=2};declare const e:E,f:F;const view=<p>{e&&'x'}{f&&'x'}</p>;",
		"/* 世界 🌍 */\r\ndeclare const count:number; const view=<p>{count && 'x'}</p>;\r\n",
	}
	for i, source := range jsx {
		paths = append(paths, h.write(fmt.Sprintf("jsx-%03d.tsx", i), "declare namespace JSX {interface Element{} interface IntrinsicElements{[name:string]:unknown}} declare function Badge(p:{visible?:unknown}):JSX.Element;\n"+source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range wave16FollowupRuleNames {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", registration, true)
	asan := h.build(stage0, "wave16-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Each rule mutant changes a real judgment or report range, finishes normally,
	// and is killed exclusively by the independent cohere diagnostic bytes.
	for _, change := range []struct{ name, file, from, to string }{
		{"listener-general", "no_global_listener_target_assertion.a", "['Element', 'HTMLElement', 'SVGElement', 'MathMLElement'].includes(part.name)", "['Element', 'SVGElement', 'MathMLElement'].includes(part.name)"},
		{"render-number", "no_leaked_number_render.a", "if((type.flags & (64 | 128)) !== 0) return 2;", "if((type.flags & (64 | 128)) !== 0) return 1;"},
		{"mock-range", "no_mock_on_module_namespace.a", "this.rules.byte(call.end),", "this.rules.byte(call.end) + 1,"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			mutant := wave16FollowupSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
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
	// The existing fact API is independently held by its direct checker tests.
	if manifest := os.Getenv("ADAMIC_WAVE16_REPOSITORY_MANIFEST"); manifest != "" {
		h.compare("repository", oracle, binary, filepath.Join(repository, "tsconfig.json"), manifest)
		h.compare("repository-asan", oracle, asan, filepath.Join(repository, "tsconfig.json"), manifest)
	}
	if manifest := os.Getenv("ADAMIC_WAVE16_COMPILER_MANIFEST"); manifest != "" {
		corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
		if corpus == "" {
			t.Fatal("compiler source is required")
		}
		h.compare("compiler", oracle, binary, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
		h.compare("compiler-asan", oracle, asan, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','node-symbol-details'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released program queried using the existing fact API: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, caught by the required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
