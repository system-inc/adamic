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

// Not parallel: native builds, sanitizers and measured subprocesses share a machine.
func TestWave30ThirdAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_30_THIRD_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_30_third_suite.a")
	binary := h.build(stage0, "wave-30-third", entry, archive, false)
	oracle := volumeOracle(h, "wave-30-third-oracle", "oracle_wave_30_third.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	controls := []string{
		"Promise.race([Promise.resolve(),new Promise(r=>{setTimeout(r,1)})]);",
		"Promise.race([new Promise(r=>setTimeout(r,1))]);",
		"Promise.race([new Promise(r=>{void (setTimeout(r,1));const timer=setTimeout(r,2);let other;other=setTimeout(r,3)})]);",
		"const timeout=new Promise(r=>{setTimeout(r,1)});Promise.race([timeout,timeout]);",
		"let timeout=new Promise(r=>{setTimeout(r,1)});Promise.race([timeout]);",
		"Promise.race([new Promise(r=>{const timer=setTimeout(r,1);clearTimeout(timer)})]);",
		"Promise.race([new Promise(r=>{const timer=setTimeout(r,1);const value={timer}})]);",
		"Promise.race([new Promise(r=>{const timer=setTimeout(r,1);function later(){return timer}})]);",
		"Promise.race([new Promise(r=>{function later(){setTimeout(r,1)};class C{m(){setTimeout(r,1)}}})]);",
		"Promise.race([new Promise(function(r){setTimeout(r,1);return setTimeout(r,2)})]);",
		"Promise.race([new Promise(r=>{globalThis.setTimeout(r,1);window.setTimeout(r,2)})]);",
		"Promise.race([new Promise(r=>{let timer;timer=setTimeout(r,1);timer=setTimeout(r,2)})]);",
		"Promise.race([new Promise(r=>{let timer;let second;second=timer=setTimeout(r,1)})]);",
		"const setTimeout=(r:unknown,n:number)=>0;Promise.race([new Promise(r=>{setTimeout(r,1)})]);",
		"const Promise={race:(x:unknown)=>x};Promise.race([]);",
		"/* 世界 🌍 */\r\n(Promise).race([(new Promise(r=>(setTimeout(r,1))))]);",
		"Promise.race([new Promise(r=>{const timer=setTimeout(r,1);{const timer=1;console.log(timer)}})]);",
		"Promise.race([new Promise(r=>{let timer;(timer)=setTimeout(r,1)})]);",
		"Promise.race([new Promise(r=>{setTimeout(r,1).toString();const {value}=setTimeout(r,2) as any})]);",
		"Promise.race([Promise.resolve()]);Promise.all([new Promise(r=>{setTimeout(r,1)})]);",
	}
	paths := []string{}
	for at, source := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", at), source+"\nexport {};\n"))
	}

	config = h.write("controls-tsconfig.json", `{"files":["control-000.a"],"compilerOptions":{"strict":true,"target":"ESNext","module":"ESNext","moduleResolution":"Bundler","lib":["ESNext","DOM"]}}`)

	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("\tnexus/correctness-no-uncleared-race-timeout\t")) {
		t.Fatal("missing timer positive control")
	}
	nodeGlobals := h.write("node-globals.a", `export {};declare global {namespace NodeJS {interface Timeout {unref():this;ref():this}}function setTimeout<T extends any[]>(cb:(...args:T)=>void,ms?:number,...args:T):NodeJS.Timeout;namespace setTimeout {const __promisify__:unknown}function clearTimeout(t:NodeJS.Timeout|string|number|undefined):void;}`)
	nodeAlias := filepath.Join(directory, "node-globals.d.ts")
	if err := os.Symlink(filepath.Base(nodeGlobals), nodeAlias); err != nil {
		if target, e := os.Readlink(nodeAlias); e != nil || target != filepath.Base(nodeGlobals) {
			t.Fatal(err)
		}
	}
	nodeConfig := h.write("node-controls-tsconfig.json", `{"files":["node-globals.d.ts","control-000.a"],"compilerOptions":{"strict":true,"target":"ESNext","module":"ESNext","moduleResolution":"Bundler","lib":["ESNext"]}}`)
	nodeTruth := h.compare("node-controls", oracle, binary, nodeConfig, manifest)
	if !bytes.Contains(nodeTruth.stdout, []byte("\tnexus/correctness-no-uncleared-race-timeout\t")) {
		t.Fatal("missing Node ambient positive control")
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-30-third-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	h.compare("node-controls-asan", oracle, asan, nodeConfig, manifest)
	for _, mutation := range []struct{ name, file, from, to string }{
		{"timer-range", "correctness_no_uncleared_race_timeout.a", "rules.byte(node.end), '')", "rules.byte(node.end) + 1, '')"},
	} {
		sourceDir := filepath.Join(directory, mutation.name+"-source")
		if err := os.MkdirAll(sourceDir, 0755); err != nil {
			t.Fatal(err)
		}
		for _, pattern := range []string{"*.a", "*.ts"} {
			files, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware", pattern))
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				source := string(data)
				if filepath.Base(file) == mutation.file {
					if strings.Count(source, mutation.from) != 1 {
						t.Fatal("nonunique mutant")
					}
					source = strings.Replace(source, mutation.from, mutation.to, 1)
				}
				source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
				source = strings.ReplaceAll(source, "../lint/", filepath.Join(repository, "stage1/cohere/lint")+"/")
				if err := os.WriteFile(filepath.Join(sourceDir, filepath.Base(file)), []byte(source), 0600); err != nil {
					t.Fatal(err)
				}
			}
		}
		mutant := h.build(stage0, mutation.name, filepath.Join(sourceDir, "wave_30_third_suite.a"), archive, false)
		got := h.must(mutation.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("mutant survived: %s", mutation.name)
		}
		t.Logf("%s exits 0; independent Go bytes catch byte %d", mutation.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE_30_THIRD_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE_30_THIRD_COMPILER_MANIFEST")},
	} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		goRun := h.must(corpus.name+"-timed-go", exec.Command(oracle, corpus.config, corpus.manifest))
		cmd := exec.Command(binary, corpus.config, corpus.manifest)
		cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		native := h.must(corpus.name+"-timed-native", cmd)
		if !bytes.Equal(goRun.stdout, native.stdout) {
			t.Fatal("timed bytes differ")
		}
		t.Logf("%s whole process native=%s Go=%s; %s", corpus.name, native.elapsed, goRun.elapsed, native.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier',args[2]??''));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)

	for _, question := range []string{"symbol-lineage", "symbol-identities", "binding-declarations"} {
		got := h.run("released-"+strings.Split(question, "\n")[0], exec.Command(stale, config, probe, question))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
		}
		t.Logf("released %s handle refused with panic 70", strings.Split(question, "\n")[0])
	}
}

// This verifies the reusable native path state, not either incomplete process rule.
func TestWave30ProcessOutputStateAgreement(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_30_process_state_suite.a")
	binary := h.build(stage0, "process-state", entry, archive, false)
	oracle := filepath.Join(directory, "go-state")
	h.must("oracle-build", exec.Command("go", "build", "-o", oracle, filepath.Join(repository, "stage1/cohere/typeaware/testdata/oracle_wave_30_process_state.go")))
	steps := []string{}
	for _, left := range []string{",", ",0,", ",1,", ",0,1,", ",1,10,", ",0,1,10,"} {
		for _, right := range []string{",", ",0,", ",1,", ",0,1,", ",1,10,", ",0,1,10,"} {
			steps = append(steps, "reset", "w"+left, "w"+right, "c0", "c1", "c10")
		}
	}
	truth := h.must("go-state", exec.Command(oracle, steps...))
	got := h.must("native-state", exec.Command(binary, steps...))
	if len(got.stderr) != 0 || !bytes.Equal(truth.stdout, got.stdout) {
		t.Fatal("native state differs from Go")
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "process-state-asan", entry, sanitized, true)
	got = h.must("asan-state", exec.Command(asan, steps...))
	if len(got.stderr) != 0 || !bytes.Equal(truth.stdout, got.stdout) {
		t.Fatal("sanitized state differs")
	}
	source, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/process_output_state.a"))
	if err != nil {
		t.Fatal(err)
	}
	h.write("process_output_state.a", strings.Replace(string(source), "!held.includes(`,${index},`)", "held.includes(`,${index},`)", 1))
	suite, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	mutantEntry := h.write("mutant-state.a", string(suite))
	mutant := h.build(stage0, "state-mutant", mutantEntry, archive, false)
	got = h.must("state-mutant-run", exec.Command(mutant, steps...))
	if len(got.stderr) != 0 || bytes.Equal(truth.stdout, got.stdout) {
		t.Fatal("state mutant survived")
	}
	t.Logf("216 state transitions agree; catch-state mutant exits 0 and Go catches byte %d", firstDifference(got.stdout, truth.stdout))
}
