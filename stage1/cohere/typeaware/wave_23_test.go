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

func wave23Controls() []string {
	return []string{
		"throw 'bad';throw (undefined);throw null;throw 0;throw {};",
		"declare const a:any;declare const u:unknown;declare const n:never;throw a;throw u;throw n;",
		"throw new Error('ok');throw new TypeError('ok');class Custom extends Error {} throw new Custom();",
		"declare const union:Error|string;throw union;declare const both:Error&{tag:1};throw both;",
		"function f<T extends Error>(e:T){throw e;}function g<T extends string>(e:T){throw e;}",
		"try{}catch(e){throw (e);}try{}catch({message}){throw message;}",
		"declare const p:Promise<number>;p.catch((e:string)=>{throw e;});p.then(x=>x,(e:string)=>{throw e;});p.then((e:string)=>{throw e;});",
		"declare const p:Promise<number>;p['catch']((e:string)=>{throw e;});p.catch(function(e:string){throw e;});p.catch((a:string,e:string)=>{throw e;});p.catch((...e:string[])=>{throw e;});",
		"declare const p:{then(...callbacks:((x:number)=>void)[]):void;catch(cb:(e:string)=>void):void};p.catch(e=>{throw e;});",
		"const fake={catch(cb:(x:string)=>void){}};fake.catch(e=>{throw e;});",
		"class Thenable {then(cb:(v:number)=>void){return this;} #catch(cb:(e:string)=>void){} run(){this.#catch(e=>{throw e;});}}",
		"Promise.reject();Promise.reject(5);Promise.reject(new Error());Promise.reject((new TypeError()));",
		"declare const a:any;declare const u:unknown;declare const e:Readonly<Error>;declare const nested:Readonly<Readonly<Error>>;Promise.reject(a);Promise.reject(u);Promise.reject(e);Promise.reject(nested);",
		"Promise['reject']('bad');Promise[`reject`]('bad');const k='reject';Promise[k]('ignored');",
		"const fake={reject(x:unknown){}};fake.reject(5);class Error{} Promise.reject(new Error());throw new Error();",
		"new Promise((resolve,reject)=>{reject();reject(1);reject(new Error());const f=()=>reject('nested');f();});",
		"new Promise(function(resolve,reject){function shadow(reject:(x:number)=>void){reject(5);}reject('bad');(reject)('bad');});",
		"new Promise((resolve,...reject)=>{reject(5);});new Promise((resolve,{reject}:any)=>reject(5));",
		"declare const args:any[];Promise.reject(...args);declare const union:Error|string;Promise.reject(union);",
		"[1,2].reduce((a,n)=>a.concat(n),[] as number[]);[1].reduce((a,n)=>a.concat(n),<number[]>[]);",
		"[1].reduce<number[]>((a,n)=>a.concat(n),[] as number[]);([1].reduce)((a,n)=>a.concat(n),([] as number[]));",
		"[1]['reduce']((a,n)=>a.concat(n),[] as number[]);[1]?.reduce((a,n)=>a.concat(n),[] as number[]);",
		"[1].reduce((a,n)=>a,{} as Record<'a'|'b',number>);[1].reduce((a,n)=>a,{} as Record<string,number>);[1].reduce((a,n)=>a,[] as const);",
		"declare const tuple:[number,number];tuple.reduce((a,n)=>a.concat(n),[] as number[]);function f<T extends number[]>(a:T){a.reduce((a,n)=>a.concat(n),[] as number[]);}",
		"declare const mixed:number[]|{reduce:Function};mixed.reduce((a,n)=>a,[] as number[]);declare const intersection:number[]&{tag:1};intersection.reduce((a,n)=>a,[] as number[]);",
		"/* 世界 🌍 */\r\nthrow ('é');\r\n[1].reduce((a,n)=>a, /* trivia */ ({} as Record<string, number>));\r\n",
	}
}

func wave23Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(h.repository, "stage1/cohere/typeaware"))
	if err != nil {
		h.t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || (filepath.Ext(entry.Name()) != ".a" && filepath.Ext(entry.Name()) != ".ts") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(h.repository, "stage1/cohere/typeaware", entry.Name()))
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if entry.Name() == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatal("nonunique mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_23_suite.a"), archive, false)
}

// Not parallel: builds, sanitizer archives and measured runs share one machine.
func TestWave23AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE23_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_23_suite.a")
	binary := h.build(stage0, "native", entry, archive, false)
	oracle := volumeOracle(h, "wave23-oracle", "oracle_wave_23.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range wave23Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"only-throw-error", "prefer-promise-reject-errors", "prefer-reduce-type-parameter"} {
		if !bytes.Contains(truth.stdout, []byte("\t@typescript-eslint/"+name+"\t")) {
			t.Fatal("no positive control", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "native-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"throw", "only_throw_error.a", "type.flags & 4", "type.flags & 8"},
		{"reject", "prefer_promise_reject_errors.a", "this.rules.add('prefer-promise-reject-errors', 'rejectAnError', message, index);", "this.rules.add('prefer-promise-reject-errors', 'rejectAnError', message + 'wrong', index);"},
		{"reduce", "prefer_reduce_type_parameter.a", "this.rules.byte(valueNode.end), this.rules.byte(expression.end), ''", "this.rules.byte(valueNode.end) + 1, this.rules.byte(expression.end), ''"},
	} {
		mutant := wave23Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}

	for _, change := range []struct{ name, file, from, to string }{
		{"alias-argument", "bridge/tsgo/checker/type_alias_info.go", "for _, argument := range alias.TypeArguments() {", "for _, argument := range alias.TypeArguments()[:0] {"},
		{"rest-callback", "bridge/tsgo/checker/then_callback_parameters.go", "t = checker.Checker_getIndexTypeOfType(c, t, checker.Checker_numberType(c))", "t = nil"},
	} {
		overlay := h.overlay(change.name, change.file, change.from, change.to)
		mutantArchive := h.archive(change.name, overlay, false)
		mutant := h.build(stage0, change.name+"-native", entry, mutantArchive, false)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("fact mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, Go byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE23_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE23_COMPILER_MANIFEST")},
	} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		goRun := h.must(corpus.name+"-timing-go", exec.Command(oracle, corpus.config, corpus.manifest))
		command := exec.Command(binary, corpus.config, corpus.manifest)
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		nativeRun := h.must(corpus.name+"-timing-native", command)
		if !bytes.Equal(goRun.stdout, nativeRun.stdout) {
			t.Fatal("timed bytes differ")
		}
		t.Logf("%s timing: native %s, Go %s; native stderr %s; Go stderr %s", corpus.name, nativeRun.elapsed, goRun.elapsed, nativeRun.stderr, goRun.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoInspect(program,file,0,1,'Identifier','raw-shape');tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','type-alias-info\n1'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// mutant keeps released program")
	staleArchive := h.archive("released-registry", overlay, false)
	staleBinary := h.build(stage0, "released-registry-probe", released, staleArchive, false)
	staleGot := h.run("released-registry-run", exec.Command(staleBinary, config, probe))
	// The retained registry mutant answers normally; only the required refusal catches it.
	if staleGot.err != nil || len(staleGot.stderr) != 0 {
		t.Fatal("released registry mutant survived")
	}
	t.Logf("released handle: normal exit 70; registry mutant caught by required released-handle error")
}
