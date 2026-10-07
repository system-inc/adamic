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

// Not parallel: native archives and sanitizer subprocesses share the scratch budget.
func TestWave21ProcessRules(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE21_PROCESS_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_21_process_suite.a")
	binary := h.build(stage0, "wave21-process", entry, archive, false)
	oracle := volumeOracle(h, "wave21-process-oracle", "oracle_wave_21_process.go")
	h.write("timers.d.ts", `export {}; declare global {namespace NodeJS { interface Timeout {unref():this;ref():this;} }
 function setTimeout<TArgs extends any[]>(callback:(...args:TArgs)=>void,delay?:number,...args:TArgs):NodeJS.Timeout;
 namespace setTimeout {const __promisify__:unknown;} function clearTimeout(timeout:NodeJS.Timeout|string|number|undefined):void;}`)
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ES2022"],"noEmit":true},"include":["*.d.ts"]}`)
	paths := wave21ProcessFixtures(h, "correctness_no_uncleared_race_timeout", "correctnessNoUnclearedRaceTimeout")
	paths = append(paths, wave21ProcessFixtures(h, "correctness_no_process_exit_after_output", "correctnessNoProcessExitAfterOutput")...)
	paths = append(paths, wave21BlockingFixtures(h)...)
	paths = append(paths, h.write("dom.a", `/// <reference lib="dom" />
 declare function work():Promise<string>; export async function load(ms:number) {return Promise.race([work(), new Promise<never>((_r,reject)=>{window.setTimeout(reject,ms);})]);}`))
	for i, body := range []string{
		"setTimeout(reject,1);", "void(setTimeout(reject,1));", "const timer=setTimeout(reject,1);", "let timer;timer=setTimeout(reject,1);", "let timer;timer=setTimeout(reject,1); use({timer});", "const timer=setTimeout(reject,1); function nested(){use(timer);}", "let timer;use(timer=setTimeout(reject,1));", "let timer;((timer))=setTimeout(reject,1);", "class X { m(){setTimeout(reject,1);} }", "setTimeout(reject,1).unref();",
	} {
		paths = append(paths, h.write(fmt.Sprintf("timer-boundary-%03d.a", i), "declare function work():Promise<string>;declare function use(x:unknown):void; export const race=Promise.race([work(),new Promise<never>((_r,reject)=>{"+body+"})]);"))
	}
	for i, body := range []string{
		"console.log('outside');try {void 0;}catch {process.exit(1);}",
		"declare const callback:((x:unknown)=>void)|undefined; callback?.(process.exit(0));console.log('after');process.exit(1);",
		"export function* work(){try {console.log('before');yield 0;while(true){}} finally{process.exit(1);}}work();",
		"console.log('outside');try {void flag;}catch {process.exit(1);}",
		"declare const callback:((x:unknown)=>void)|undefined; callback?.(console.log('arg'));process.exit(1);",
	} {
		paths = append(paths, h.write(fmt.Sprintf("process-boundary-%03d.a", i), body+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("\tunclearedRaceTimeout\t")) {
		t.Fatal("no timer positive control")
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave21-process-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	siblings := filepath.Join(repository, "stage1/cohere/typeaware")
	for _, change := range []struct{ name, file, from, to string }{
		{"timer", "correctness_no_uncleared_race_timeout.a", "&& this.lost(executor, index)", "&& !this.lost(executor, index)"},
		{"process-exit", "correctness_no_process_exit_after_output.a", "if(state.length > 0 && !reported.includes(index))", "if(state.length === 0 && !reported.includes(index))"},
		{"blocking", "correctness_require_blocking_standard_streams.a", "if(!shebang && this.index.imported.includes(this.rules.path))", "if(!shebang && !this.index.imported.includes(this.rules.path))"},
	} {
		data, err := os.ReadFile(filepath.Join(siblings, change.file))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(data), change.from) != 1 {
			t.Fatal("nonunique mutant", change.name)
		}
		source := strings.Replace(string(data), change.from, change.to, 1)
		files, err := os.ReadDir(siblings)
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			source = strings.ReplaceAll(source, "'./"+file.Name()+"'", "'"+filepath.Join(siblings, file.Name())+"'")
		}
		if change.name == "process-exit" {
			for _, name := range []string{"NoProcessExitAfterOutput", "OutputWalk", "OutputPosition"} {
				source = strings.ReplaceAll(source, name, name+"Mutant")
			}
		}
		mutantFile := h.write(change.name+"_mutant.a", source)
		main, err := os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		source = strings.Replace(string(main), "'./"+change.file+"'", "'"+mutantFile+"'", 1)
		for _, file := range files {
			source = strings.ReplaceAll(source, "'./"+file.Name()+"'", "'"+filepath.Join(siblings, file.Name())+"'")
		}
		source = strings.ReplaceAll(source, "'../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")
		if change.name == "process-exit" {
			dependency, err := os.ReadFile(filepath.Join(siblings, "correctness_require_blocking_standard_streams.a"))
			if err != nil {
				t.Fatal(err)
			}
			dependent := strings.Replace(string(dependency), "'./correctness_no_process_exit_after_output.a'", "'"+mutantFile+"'", 1)
			for _, file := range files {
				dependent = strings.ReplaceAll(dependent, "'./"+file.Name()+"'", "'"+filepath.Join(siblings, file.Name())+"'")
			}
			for _, name := range []string{"NoProcessExitAfterOutput", "OutputWalk", "OutputPosition"} {
				dependent = strings.ReplaceAll(dependent, name, name+"Mutant")
			}
			dependentFile := h.write("process-exit-dependent-blocking.a", dependent)
			source = strings.Replace(source, "'"+filepath.Join(siblings, "correctness_require_blocking_standard_streams.a")+"'", "'"+dependentFile+"'", 1)
			source = strings.ReplaceAll(source, "NoProcessExitAfterOutput", "NoProcessExitAfterOutputMutant")
		}
		mutant := h.build(stage0, change.name+"-mutant", h.write(change.name+"_mutant_main.a", source), archive, false)
		got := h.must(change.name+"-mutant-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) > 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived or failed outside comparison", change.name)
		}
		t.Logf("%s judgment mutant: exit 0, comparison catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	released := h.write("released_queries.a", `import {panic,programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const a=programArguments();const file=a[1]??panic('file');const p=tsgoProgram(a[0]??panic('config'),[file]);const q=a[2]??panic('question');tsgoRelease(p);console.log(tsgoInspect(p,file,0,3,'CallExpression',q));`)
	probe := h.write("released_probe.a", "f();")
	stale := h.build(stage0, "released-queries", released, archive, false)
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutantStale := h.build(stage0, "released-queries-mutant", released, mutantArchive, false)
	for _, question := range []string{"node-symbol-context", "node-symbol-context\nfollow-alias", "program-modules", "resolved-declaration"} {
		observed := h.run("released-query-run", exec.Command(stale, config, probe, question))
		if code, ok := observed.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(observed.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released query escaped: %v %s", observed.err, observed.stderr)
		}
		h.must("released-query-mutant-run", exec.Command(mutantStale, config, probe, question))
		t.Logf("%s: released panic 70; retaining registry mutant exits 0 and is caught", strings.ReplaceAll(question, "\n", "/"))
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"compiler", os.Getenv("ADAMIC_WAVE21_COMPILER_CONFIG"), os.Getenv("ADAMIC_WAVE21_COMPILER_MANIFEST")},
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE21_REPOSITORY_MANIFEST")},
	} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
			native := h.must(corpus.name+"-timed-native", exec.Command(binary, corpus.config, corpus.manifest, "--count"))
			direct := h.must(corpus.name+"-timed-go", exec.Command(oracle, corpus.config, corpus.manifest, "--count"))
			if !bytes.Equal(native.stdout, direct.stdout) {
				t.Fatal("timed counts differ")
			}
			t.Logf("%s quiet timing native %.6fs Go %.6fs", corpus.name, native.elapsed.Seconds(), direct.elapsed.Seconds())
		}
	}
}
