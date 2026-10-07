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

var wave16FourthRuleNames = []string{"no-new-func", "no-new-native-nonconstructor", "no-new-wrappers"}

func wave16FourthControls() []string {
	return []string{
		"new Function('x','return x');Function('return 1');Function.call(null,'return 1');Function.bind(null,'return 1')();Function.apply(null,['return 1']);Function['call'](null);Function[`bind`](null);(Function?.call)(null);new (Function)();Function.toString();Function.call;new Function.call(null);",
		"new Symbol('x');new BigInt(1);new (Symbol)('x');new ((BigInt))(1);Symbol('x');BigInt(1);",
		"new String('x');new Number(1);new Boolean(false);new (String)('x');new ((Number))(1);String('x');Number(1);Boolean(false);",
		"function f(Function:any,Symbol:any,BigInt:any,String:any,Number:any,Boolean:any){new Function();Function.call();new Symbol();new BigInt();new String();new Number();new Boolean();}",
		"class Function{};class Symbol{};class BigInt{};class String{};class Number{};class Boolean{};new Function();new Symbol();new BigInt();new String();new Number();new Boolean();",
		"function inner(){class Function{};class Number{};new Function();new Number();}new Function();new Number();",
		"import Function from './helper.a';import {value as Number} from './helper.a';new Function();new Number();",
		"function Symbol(){};function BigInt(){};function Boolean(){};new Symbol();new BigInt();new Boolean();",
		"function f(){if(true){new Boolean();}else{var Boolean=class{};}} {const String=class{};new String();}new String();",
		"declare const foo:any;foo.Function.call(null);(foo,Function).call(null);Function.Function.call(null);foo().Function.call(null);new window.String();new globalThis.Number();new Object();Function[0](null);",
		"Function?.();Function?.call();(Function).call();new Function;function nested(){return function Symbol(){};}new Symbol();",
		"function nested(){return function Function(){Function();};}new Function();Function.bind;Function.notAMethod();",
		"/* 世界 🌍 */new Boolean(false);new Symbol();Function['call'](null);declare const key:string;Function[key](null);Function[('apply')](null);new String;new Number;new Boolean;",
	}
}

func wave16FourthSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave16_fourth_suite.a"), archive, false)
}

// All output, including failures and mutants, goes to files. This suite is serial
// because sanitizer archives and corpus runs share the limited scratch volume.
func TestWave16FourthAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE16_FOURTH_ARTIFACTS"); path != "" {
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave16_fourth_suite.a")
	binary := h.build(stage0, "wave16", entry, archive, false)
	oracle := volumeOracle(h, "wave16-oracle", "oracle_wave16_fourth.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","moduleResolution":"Bundler","jsx":"preserve","lib":["ES2022","DOM"]},"files":["node.d.a"]}`)
	h.write("node.d.a", "declare module 'node:child_process' {\n export function exec(command:string,...args:unknown[]):unknown;\n export namespace exec {export function __promisify__(command:string):unknown;}\n export function execSync(command:string,...args:unknown[]):unknown;\n export function spawn(command:string,...args:unknown[]):unknown;\n export function spawnSync(command:string,...args:unknown[]):unknown;\n export function execFile(command:string,...args:unknown[]):unknown;\n export function execFileSync(command:string,...args:unknown[]):unknown;\n}\ndeclare module 'node:util' {export function promisify(f:typeof import('node:child_process').exec):typeof import('node:child_process').exec.__promisify__;}\n")
	paths := []string{filepath.Join(directory, "node.d.a")}
	for i, source := range wave16FourthControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("helper.a", "export let value=1; export const object={x:1}; export default object;\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range wave16FourthRuleNames {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave16-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Each rule mutant changes a real judgment or report range, finishes normally,
	// and is killed exclusively by the independent cohere diagnostic bytes.
	for _, change := range []struct{ name, file, from, to string }{
		{"function-range", "no_new_func.a", "this.rules.byte(node.end)", "this.rules.byte(node.end) + 1"},
		{"native-range", "no_new_native_nonconstructor.a", "const anchor = callee;", "const anchor = node;"},
		{"wrapper-range", "no_new_wrappers.a", "this.rules.byte(anchor.end)", "this.rules.byte(anchor.end) + 1"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			mutant := wave16FourthSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
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
