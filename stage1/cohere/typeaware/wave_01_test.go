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

func wave01Controls() []string {
	return []string{
		"/** @deprecated \u0085Reason\u0085 */ export const nel=1;nel;\n/** @deprecated Reason\ufeff */ export const bom=1;bom;",
		"namespace NS {\n/** @deprecated Use Plain. */\nexport function Old(){}\nexport function Plain(){}\n}\nimport Alias = NS.Old;import Plain = NS.Plain;Alias();Plain();",
		`export interface Data {
/** @deprecated Use fresh. */
old:number;
fresh:number;
}
declare const data:Data;
data.old;data['old'];const key='old';data[key];const broad:string='old';data[broad];const {old}=data;const {old:renamed}=data;
`,
		`export class C {
/** @deprecated Use modern. */
method(){}
/** @deprecated Getter reason. */
get value(){return 1;}
/** @deprecated Private reason. */
#old=1;
use(){this.#old;}
}
const c=new C();c.method();c.value;
`,
		`export class C {
constructor();
/** @deprecated String overload. */
constructor(x:string);
constructor(x?:string){}
}
new C();new C('x');
`,
		`/** @deprecated */
export function overloaded(x:string):void;
export function overloaded(x:number):void;
export function overloaded(x:string|number){}
overloaded('x');overloaded(1);overloaded;
`,

		"export function partial(b:boolean) { if(b) return 1; }",
		"export function exhaustive(k:'a'|'b') { switch(k) {case 'a':return 1;case 'b':return 2;} }",
		"declare function stop():never; export function stopped(b:boolean) { if(b)return 1;stop(); }",
		"export function empty(){};export function voidReturn(b:boolean):void {if(b)return;} export function anyReturn(b:boolean):any {if(b)return 1;} export function undefinedReturn(b:boolean):undefined {if(b)return undefined;}",
		"export function annotated(b:boolean):number|undefined {if(b)return 1;} export function typeError(b:boolean):number {if(b)return 1;} export function neverReturn():never {}",
		"export const arrow=(b:boolean)=>{if(b)return 1;}; export const expression=(b:boolean)=>b?1:2; export class C {method(b:boolean){if(b)return 1;}get x(){if(Date.now())return 1;}}",
		"export async function asyncPartial(b:boolean) {if(b)return 1;} export function* gen(b:boolean):Generator<number,number|undefined> {yield 1;if(b)return 2;}",
		"export function bare(b:boolean) {if(b)return 1;return;} export function nested(b:boolean) {const f=()=>{return;};if(b)return 2;} export class C{constructor(b:boolean){if(b)return;}}",
		"/** @deprecated Use modern. */ export function old(){} old(); export function modern(){}modern();",
		"/** @deprecated */ export const old=1;const obj={old};export {obj};",
		"export interface Data {/** @deprecated Use fresh. */ old:number; fresh:number};declare const data:Data;data.old;data.fresh;data['old'];const key='old';data[key];const broad:string='old';data[broad];const {old}=data;const {old:renamed}=data;",
		"/** @deprecated Use newCall. */ function overloaded(x:string):void;function overloaded(x:number):void;function overloaded(x:string|number){};overloaded('x');overloaded(1);overloaded;export {overloaded};",
		"export class C {/** @deprecated */ method(){};/** @deprecated getter reason */get value(){return 1};/** @deprecated */ #old=1;use(){this.#old;}};const c=new C();c.method();c.value;",
		"/** @deprecated linked {@link replacement} reason */ export const old=1;old;",
		"export function f(b:boolean){if(b){return 1;}else{return 2;}}",
		"export function f(a:boolean,b:boolean){if(a)return 1;else if(b)return 2;else return 3;} export function incomplete(a:boolean,b:boolean){if(a)return 1;else if(b)return 2;}",
		"export function collision(a:number,b:boolean){if(b){return 1;}else{let a:number=2;return a;}} export function clean(b:boolean){if(b){return 1;}else{const x:number=2;return x;}}",
		"export function captured(b:boolean){if(b){return 1;}else{let missing=1;}missing;}export function nested(b:boolean){if(b){return 1;}else{{let b=1;}}}",
		"export function front(b:boolean){if(b)return 1\nelse{[1,2,3].join();}} export function back(b:boolean){if(b)return 1;else{console.log(1)}console.log(2)}",
		"export function f(b:boolean){if(b){return 1;}else function x(){};} export function naive(b:boolean,c:boolean){if(b){if(c){return 1;}else{return 2;}console.log(1);}else{return 3;}}",
		"export function f(b:boolean){if(b){return 1;}else{class X{};const {a:renamed}={a:1};return renamed;}}",
		"/* 世界 🌍 */\r\nexport function é(b:boolean){if(b)return 1;}\r\n/** @deprecated 世界 🌍 */ export const 漢=1;漢;\r\nexport function f(b:boolean){if(b)return 1;else{return 2;}}\r\n",
	}
}

func wave01Mutant(h *harness, stage0, archive, name, file, from, to string) string {
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
		if !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".a") {
			continue
		}
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
		for _, dependency := range files {
			if strings.HasSuffix(dependency, ".ts") {
				base := filepath.Base(dependency)
				source = strings.ReplaceAll(source, "'./"+base+"'", "'./"+strings.TrimSuffix(base, ".ts")+".a'")
			}
		}
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) + ".a"
		if err := os.WriteFile(filepath.Join(directory, base), []byte(source), 0644); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_01_suite.a"), archive, false)
}

// Not parallel: builds, sanitizer runs and corpus timings share this machine.
func TestWave01AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE01_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_01_suite.a")
	binary := h.build(stage0, "wave01", entry, archive, false)
	oracle := volumeOracle(h, "wave01-oracle", "oracle_wave_01.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ES2022"],"noEmit":true,"allowImportingTsExtensions":true},"sourceExtensions":[".a"],"files":["control-000.a"],"include":["*.a"]}`)
	var paths []string
	for i, source := range wave01Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, positive := range []string{"`Alias` is deprecated. Use Plain.", "`#old` is deprecated. Private reason.", "`method` is deprecated. Use modern.", "`old` is deprecated. Use fresh."} {
		if !bytes.Contains(truth.stdout, []byte(positive)) {
			t.Fatalf("missing documentation positive control: %s", positive)
		}
	}
	for _, name := range []string{"nexus/correctness-no-implicit-return", "@typescript-eslint/no-deprecated", "no-else-return"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	relaxed := h.write("relaxed.json", `{"compilerOptions":{"strict":false,"target":"ES2022","module":"NodeNext","lib":["ES2022"],"noEmit":true,"allowImportingTsExtensions":true},"sourceExtensions":[".a"],"files":["control-000.a"],"include":["*.a"]}`)
	h.compare("relaxed", oracle, binary, relaxed, manifest)
	covered := h.write("covered.json", `{"compilerOptions":{"strict":true,"noImplicitReturns":true,"target":"ES2022","module":"NodeNext","lib":["ES2022"],"noEmit":true,"allowImportingTsExtensions":true},"sourceExtensions":[".a"],"files":["control-000.a"],"include":["*.a"]}`)
	h.compare("covered", oracle, binary, covered, manifest)
	if os.Getenv("ADAMIC_WAVE01_CONTROLS_ONLY") == "1" {
		return
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave01-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"implicit", "no_implicit_return.a", "facts.implicit\n", "!facts.implicit\n"},
		{"deprecated", "no_deprecated.a", "tag.reason === '' ? 'deprecated' : 'deprecatedWithReason'", "tag.reason === '' ? 'deprecated' : 'deprecated'"},
		{"else", "no_else_return.a", "this.rules.byte(node.end), source.slice(start, end)", "this.rules.byte(node.end) + 1, source.slice(start, end)"},
	} {
		mutant := wave01Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	if corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"); corpus != "" {
		pin := h.must("compiler-pin", exec.Command("git", "-C", corpus, "rev-parse", "HEAD"))
		if strings.TrimSpace(string(pin.stdout)) != compilerCommit {
			t.Fatal("compiler pin differs")
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage/compiler.manifest"))
		if err != nil {
			t.Fatal(err)
		}
		var roots []string
		for _, path := range strings.Split(string(data), "\n") {
			if path != "" {
				roots = append(roots, filepath.Join(corpus, path))
			}
		}
		compilerManifest := h.write("compiler.manifest", strings.Join(roots, "\n")+"\n")
		compilerConfig := filepath.Join(corpus, "src/compiler/tsconfig.json")
		h.compare("compiler", oracle, binary, compilerConfig, compilerManifest)
		h.compare("compiler-asan", oracle, asan, compilerConfig, compilerManifest)
	}
	data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage/repository.manifest"))
	if err != nil {
		t.Fatal(err)
	}
	var roots []string
	for _, path := range strings.Split(string(data), "\n") {
		if path != "" {
			roots = append(roots, filepath.Join(repository, path))
		}
	}
	repositoryManifest := h.write("repository.manifest", strings.Join(roots, "\n")+"\n")
	h.compare("repository", oracle, binary, filepath.Join(repository, "tsconfig.json"), repositoryManifest)
	h.compare("repository-asan", oracle, asan, filepath.Join(repository, "tsconfig.json"), repositoryManifest)
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	for _, question := range []string{"function-return-flow", "symbol-documentation\nnode", "scope-symbol-declarations"} {
		probe := h.write("released.a", "import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,21,'FunctionDeclaration',"+fmt.Sprintf("%q", question)+"));\n")
		input := h.write("probe.a", "function f(){return;}")
		stale := h.build(stage0, "released", probe, archive, false)
		got := h.run("released-run", exec.Command(stale, config, input))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
		}
		t.Logf("released handle %s: panic 70", question)
		mutant := h.build(stage0, "released-mutant", probe, mutantArchive, false)
		got = h.must("released-mutant-run", exec.Command(mutant, config, input))
		if len(got.stderr) != 0 {
			t.Fatalf("released-registry mutant stderr: %s", got.stderr)
		}
		t.Logf("released-registry %s: mutant exits 0, caught by required panic 70", question)
	}
}
