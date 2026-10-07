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

func wave26FifthControls() []string {
	controls := []string{
		"Promise.reject('failed');Promise.reject();Promise.reject(5);Promise.reject(undefined);",
		"Promise.reject(new Error());Promise.reject(error);Promise.reject(getError());",
		"Promise['reject'](false);Promise[`reject`](null);(Promise?.reject)({});",
		"(Promise as any).reject(5);(Promise.reject as any)(false);",
		"function f(Promise:any){Promise.reject(5);new Promise((a,b)=>b(5));}",
		"function f(undefined:any){Promise.reject(undefined);}Promise.reject((undefined));",
		"Promise.reject(foo&&5);Promise.reject(5&&foo);Promise.reject(foo||5);Promise.reject(5??foo);",
		"Promise.reject(x=5);Promise.reject(x&&=5);Promise.reject(x||=foo);Promise.reject((foo,5));",
		"Promise.reject(x+=1);Promise.reject(1?2:foo);Promise.reject(1?2:3);",
		"Promise.reject((error as Error));Promise.reject((5 as number));",
		"new Promise((resolve,reject)=>reject(5));",
		"new Promise(function(resolve,reject){reject('x');reject(new Error());});",
		"new Promise((resolve,reject)=>{function f(){reject({});}const x=()=>reject();});",
		"new Promise((resolve,reject)=>{function f(reject:any){reject(5);} {const reject=foo;reject(5);}});",
		"new Promise(function(reject,reject){reject(5);});",
		"new Promise((resolve,{apply})=>apply(5));new Promise(({foo},reject)=>reject(5));",
		"new Promise((resolve,reject,x=reject(5))=>{});",
		"new Promise(foo,(resolve,reject)=>reject(5));new Promise((resolve)=>resolve(5));",
		"new (Promise)((resolve,reject)=>reject(5));new foo.Promise((resolve,reject)=>reject(5));",
		"new Promise((resolve,reject)=>{(reject as any)(5);(reject)?.(undefined);resolve(5,reject);});",
		"function f(){arguments;arguments[0];arguments.length;arguments.callee;}",
		"arguments;const f=()=>arguments;",
		"function f(arguments:any){arguments;arguments[0];}function g(){var arguments;arguments;}",
		"function f(){const arrow=()=>arguments[0];function g(){arguments;}}",
		"function f(arguments:any){function g(){arguments;}const arrow=()=>arguments;}",
		"function f(){try{}catch(arguments){arguments;}}",
		"function f(){({arguments});(arguments).length;arguments['length'];}",
		"function f(){const x={arguments:1};x.arguments;}",
		"RegExp('abc');new RegExp('abc','g');",
		"new RegExp('a/b');new RegExp('');",
		"RegExp(`abc`);new RegExp(String.raw`\\d`,'g');",
		"new RegExp(String['raw']`a/b`);new RegExp((String?.raw)`a`);",
		"new RegExp();new RegExp(pattern);new RegExp('a'+'b');new RegExp('a',flags);",
		"function f(RegExp:any){new RegExp('a');}function g(String:any){new RegExp(String.raw`a`);}",
		"const R=RegExp;R('a');new R('b');",
		"let R=RegExp;R=1;R('a');",
		"const R=flag?RegExp:other;R('a');(flag?RegExp:RegExp)('b');",
		"(0,RegExp)('a');(RegExp as any)('b');RegExp?.('c');",
		"globalThis.RegExp('a');window['RegExp']('b');self[`RegExp`]('c');",
		"globalThis['Reg'+'Exp']('a');globalThis[`Reg${'Exp'}`]('b');",
		"const {RegExp:R}=globalThis;R('a');const {RegExp}=window;RegExp('b');",
		"let R;({RegExp:R}=globalThis);R('a');",
		"const {['Reg'+'Exp']:R}=globalThis;new R('a');",
		"const G=globalThis;G.RegExp('a');",
		"RegExp=foo;RegExp('a');globalThis=foo;globalThis.RegExp('b');",
		"function f(globalThis:any){globalThis.RegExp('a');}const R=RegExp;function g(R:any){R('a');}R('b');",
		"RegExp('a'/* note */);RegExp(/* comment */'b');",
		"new RegExp('http://a');new RegExp('/*a*/');",
		"const r=/*before*/RegExp('a');const s=RegExp('b');//after",
		"x=foo\nRegExp('a');",
		"function f(){return RegExp('a');}const x=foo/RegExp('b');",
		"RegExp('a')in obj;",
		"RegExp('+');RegExp('[');RegExp('(');RegExp('a','gg');RegExp('a','uv');",
		"RegExp('a{2}');RegExp('a{2,3}');RegExp('{');RegExp('{','u');",
		"RegExp('a\\n/b\\t');RegExp(String.raw`a\\b`);",
		"RegExp('é');RegExp('a','v');RegExp('[a/b]');",
		"RegExp('(?:a)');RegExp('(?<x>a)');RegExp('a*?');",
		"#!/usr/bin/env node\nRegExp('a');",
		"/* 世界 🌍 */\r\nnew RegExp('a/b');Promise.reject('x');function f(){arguments[0];}",
		"function f(R=RegExp){R('a');}let a=RegExp;a=a;new a('b');",
		"let R;R||=RegExp;R('a');let S;S+=RegExp;S('b');",
		"const {RegExp:R=RegExp}=globalThis;R('a');",
		"new RegExp(/a/);new RegExp(/a/,'g');",
		"Promise.reject(error<number>);",
		"function f(){function arguments(){}arguments;}function g(){let arguments=1;arguments;}",
	}
	for i := range controls {
		controls[i] += "\nexport {};\n"
	}
	return controls
}

func wave26FifthMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_26_fifth.a"), archive, false)
}

// Not parallel: builds, sanitizers and cost observations share a machine.
func TestWave26FifthAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE26_FIFTH_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_26_fifth.a")
	binary := h.build(stage0, "wave26", entry, archive, false)
	oracle := volumeOracle(h, "wave26-oracle", "oracle_wave_26_fifth.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"jsx":"preserve","target":"ES2022","module":"ESNext","lib":["ES2022","DOM"],"noEmit":true},"files":["control-000.a","configured.d.ts"]}`)
	h.write("configured.d.ts", "")
	h.write("exports.d.ts", "export declare const Function:any,String:any,Symbol:any;\n")
	var paths []string
	for i, source := range wave26FifthControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d%s", i, func() string {
			if strings.Contains(source, "<div") {
				return ".tsx"
			}
			return ".a"
		}()), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"prefer-promise-reject-errors", "prefer-regex-literals", "prefer-rest-params"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave26-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"promise", "prefer_promise_reject_errors.a", "'AmpersandAmpersandToken'", "'AsteriskToken'"},
		{"regex", "prefer_regex_literals.a", "replacement = '\\\\/'", "replacement = '/'"},
		{"rest", "prefer_rest_params.a", "=== 'PropertyAccessExpression'", "=== 'ElementAccessExpression'"},
	} {
		mutant := wave26FifthMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
	}
	for _, change := range []struct{ name, file, from, to string }{
		{"syntax", "bridge/tsgo/checker/preference_structure.go", "args = n.Arguments()", "args = nil"},
		{"binding", "bridge/tsgo/checker/preference_binding.go", "out.number(p.symbolID(symbol))", "out.number(0)"},
		{"pattern", "bridge/tsgo/checker/regex_pattern.go", "out.yes(parsed)", "out.yes(parsed && false)"},
		{"comments", "bridge/tsgo/checker/source_comments.go", "ranges := comments.All(source)", "ranges := comments.All(source); ranges = nil"},
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
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);console.log(tsgoInspect(program,file,0,1,'Identifier','preference-binding'));tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','preference-binding'));`)
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
