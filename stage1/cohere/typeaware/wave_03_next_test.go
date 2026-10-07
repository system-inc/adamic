package typeaware

import (
	"bytes"
	"encoding/json"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func wave03NextBodies(t *testing.T, repository string) []string {
	t.Helper()
	var bodies []string
	for _, name := range []string{"correctness_no_process_exit_after_output", "correctness_no_uncleared_race_timeout", "correctness_require_blocking_standard_streams"} {
		tree, err := goparser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/nexus", name+"_test.go"), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			literal, ok := node.(*goast.CompositeLit)
			if !ok {
				return true
			}
			array, ok := literal.Type.(*goast.ArrayType)
			if !ok {
				return true
			}
			element, ok := array.Elt.(*goast.Ident)
			if !ok || element.Name != "string" || len(literal.Elts) < 2 {
				return true
			}
			var lines []string
			for _, e := range literal.Elts {
				value, ok := e.(*goast.BasicLit)
				if !ok || value.Kind != token.STRING {
					return true
				}
				text, err := strconv.Unquote(value.Value)
				if err != nil {
					t.Fatal(err)
				}
				lines = append(lines, text)
			}
			source := strings.Join(lines, "\n")
			if strings.Contains(source, "{") && (strings.Contains(source, "process.") || strings.Contains(source, "Promise.race") || strings.Contains(source, "console.")) {
				bodies = append(bodies, source)
			}
			return true
		})
	}
	return bodies
}
func wave03NextMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	paths, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/wave_03_next/*.a"))
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
				h.t.Fatalf("nonunique %s mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		// Rewrite the parser imports before rewriting the one-parent imports.
		source = strings.ReplaceAll(source, "../../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "'../", "'"+filepath.Join(h.repository, "stage1/cohere/typeaware")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "main.a"), archive, false)
}

// Not parallel: builds, sanitizer runs and corpus timings share CPUs and scratch.
// Not parallel: builds bridge archives and records native/Go wall times.
func TestWave03ContinuationAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE03_NEXT_ARTIFACTS"); path != "" {
		directory = path
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("next-checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_03_next/main.a")
	binary := h.build(stage0, "next", entry, archive, false)
	oracle := volumeOracle(h, "next-oracle", "oracle_wave_03_next.go")
	types := h.write("node.d.ts", `declare namespace NodeJS { interface Stream {write(data:unknown,callback?:()=>void):boolean} interface Process {stdout:Stream;stderr:Stream;exit(code?:number):never;on(event:string,callback:(...args:any[])=>void):void;} }
declare var process:NodeJS.Process;
declare module 'node:process' {export=process;}
`)
	configBytes, err := json.Marshal(map[string]any{"compilerOptions": map[string]any{"strict": true, "target": "ES2022", "lib": []string{"ES2022", "DOM"}, "moduleDetection": "auto"}, "files": []string{types}})
	if err != nil {
		t.Fatal(err)
	}
	config := h.write("next-config.json", string(configBytes))
	prelude := "declare const flag:boolean,output:string,rows:string[],milliseconds:number,argumentsList:string[],results:unknown;declare const report:{authorized:boolean;lines:string[]};declare const result:{lines:string[];exitCode:number;allOk:boolean;sources:{ok:boolean;label:string}[]};declare const fileConsole:Console,streams:{stdout:{write(text:string):void}},runner:{exit(code:number):void},socketServer:{on(event:string,listener:(error:Error)=>void):void};declare function work():Promise<string>;declare function run():Promise<void>;declare function use(value:unknown):void;declare function onEvent(listener:()=>void):void;declare function sendResults(value:unknown):void;declare function printReport():void;declare function timeoutAfter(ms:number):Promise<never>;declare function runLinkCommand(args:string[]):Promise<number>;\n"
	bodies := append(wave03NextBodies(t, repository), []string{
		"console.log('世界🌍'); process.exit(1);",
		"function main({done}: {done:()=>void}){done();console.log(output);process.exit(1)}main({done:()=>{}});",
		"function main(){console.warn('a');if(flag)process.exit(1);process.exit(2);}main();",
		"function helper(){console.log('helper')}function main(){helper();process.exit(1)}main();",
		"Promise.race([work(),new Promise((r)=>{const timer=setTimeout(r,1);use({timer});})]);",
		"Promise.race([work(),new Promise((r)=>{const timer=setTimeout(r,1);function other(){const timer=1;use(timer);}})]);",
		"Promise.race([new Promise((r)=>setTimeout(r,1)),new Promise((r)=>setTimeout(r,2))]);",
		"function main(){while(flag){if(flag)process.exit(1);console.log(output)}}main();",
		"function main(){try{console.log(output);throw 1}catch({code=process.exit(1)}){console.warn(output);process.exit(2)}}main();",
		"#!/usr/bin/env node\nconsole.log(output); process.exit(1);",
	}...)
	var paths []string
	for i, source := range bodies {
		if strings.HasPrefix(source, "#!") {
			parts := strings.SplitN(source, "\n", 2)
			source = parts[0] + "\n" + prelude + parts[1]
		} else {
			source = prelude + source
		}
		paths = append(paths, h.write(fmt.Sprintf("next-control-%03d.ts", i), source+"\nexport {};\n"))
	}
	candidates := h.write("next-candidates.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("parse-filter", exec.Command(oracle, config, candidates, "--valid-sources"))
	manifest := h.write("next-controls.manifest", string(valid.stdout))
	t.Logf("generated controls: %d candidates, %d independently parseable", len(paths), len(strings.Fields(string(valid.stdout))))
	truth := h.compare("next-controls", oracle, binary, config, manifest)
	for _, name := range []string{"correctness-no-process-exit-after-output", "correctness-no-uncleared-race-timeout", "correctness-require-blocking-standard-streams"} {
		if !bytes.Contains(truth.stdout, []byte("\tnexus/"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	// A separate program proves import reachability, ordering, callback treatment
	// and imported entry exclusion without mixing unrelated synthetic entries.
	if err := os.MkdirAll(filepath.Join(directory, "source/system"), 0755); err != nil {
		t.Fatal(err)
	}
	h.write("source/system/StandardStreams.ts", "export function blockStandardStreams(){Reflect.get(process.stdout,'_handle');}\n")
	h.write("Writer.ts", "export function showHelp(){console.log('help');}\n")
	cases := []string{
		"import {blockStandardStreams} from './source/system/StandardStreams.js';console.log(output);process.exit(1);blockStandardStreams();",
		"import {blockStandardStreams as block} from './source/system/StandardStreams.js';block();console.log(output);process.exit(1);",
		"import {blockStandardStreams} from './source/system/StandardStreams.js';function main(){console.log(output);process.exit(1);blockStandardStreams()}main();",
		"import {blockStandardStreams} from './source/system/StandardStreams.js';function main(){blockStandardStreams();console.log(output);process.exit(1)}main();",
		"import {blockStandardStreams} from './source/system/StandardStreams.js';process.on('x',()=>{console.log(output);process.exit(1)});blockStandardStreams();",
		"import {blockStandardStreams} from './source/system/StandardStreams.js';function main(x:void){console.log(output);process.exit(1)}main(blockStandardStreams());",
		"import {blockStandardStreams} from './source/system/StandardStreams.js';async function main(){await run();console.log(output);process.exit(1)}main();blockStandardStreams();",
		"import {blockStandardStreams} from './source/system/StandardStreams.js';class Setup{static{blockStandardStreams()}}console.log(output);process.exit(1);",
		"import {showHelp} from './Writer.js';showHelp();process.exit(1);",
		"import(output).then(use);console.log(output);process.exit(1);",
	}
	var ordered []string
	for i, source := range cases {
		ordered = append(ordered, h.write(fmt.Sprintf("ordered-%02d.ts", i), prelude+source+"\nexport {};\n"))
	}
	orderedManifest := h.write("ordered.manifest", strings.Join(ordered, "\n")+"\n")
	h.compare("ordered", oracle, binary, config, orderedManifest)
	library := h.write("Imported.ts", prelude+"console.log(output);process.exit(1);export const value=1;\n")
	importer := h.write("Importer.ts", "import {value} from './Imported.js';console.log(value);\n")
	h.compare("imported", oracle, binary, config, h.write("imported.manifest", library+"\n"+importer+"\n"))
	if os.Getenv("ADAMIC_WAVE03_NEXT_QUICK") != "" {
		return
	}
	sanitized := h.archive("next-checker-asan", "", true)
	asan := h.build(stage0, "next-asan", entry, sanitized, true)
	h.compare("next-controls-asan", oracle, asan, config, manifest)
	h.compare("ordered-asan", oracle, asan, config, orderedManifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"race", "correctness_no_uncleared_race_timeout.a", "this.timer(index) && this.lost(index, executor)", "this.timer(index) && !this.lost(index, executor)"},
		{"output", "correctness_no_process_exit_after_output.a", "if(result.exits.includes(exit))", "if(!result.exits.includes(exit))"},
		{"blocking", "correctness_require_blocking_standard_streams.a", "const reason = shebang ?", "const reason = !shebang ?"},
	} {
		mutant := wave03NextMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s: builds, exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
	}
	for _, population := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE03_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE03_COMPILER_MANIFEST")},
	} {
		if population.manifest == "" {
			continue
		}
		h.compare(population.name, oracle, binary, population.config, population.manifest)
		h.compare(population.name+"-asan", oracle, asan, population.config, population.manifest)
		want := h.must(population.name+"-timed-go", exec.Command(oracle, population.config, population.manifest))
		command := exec.Command(binary, population.config, population.manifest)
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(population.name+"-timed-native", command)
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatalf("timed %s differs", population.name)
		}
		t.Logf("%s timing: native %s Go %s; native %s Go %s", population.name, got.elapsed, want.elapsed, got.stderr, want.stderr)
	}
	probe := h.write("next-released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,2,'SourceFile','program-module-edges'));`)
	source := h.write("next-probe.ts", "x;")
	stale := h.build(stage0, "next-released", probe, archive, false)
	got := h.run("next-released-run", exec.Command(stale, config, source))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released module query: required panic 70")
	overlay := h.overlay("next-released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains a released program.")
	mutantArchive := h.archive("next-released-registry", overlay, false)
	mutant := h.build(stage0, "next-released-mutant", probe, mutantArchive, false)
	h.must("next-released-mutant-run", exec.Command(mutant, config, source))
	t.Log("released registry mutant exits 0 and is caught by required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
