package wave12next

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const nodeDeclarations = `declare namespace NodeJS {
 interface WritableStream { write(chunk:string|Uint8Array, callback?:(error?:Error|null)=>void):boolean }
 interface Process {stdout:WritableStream;stderr:WritableStream;exit(code?:number):never;exitCode:number;on(event:string,listener:(...args:any[])=>void):this}
 interface Timeout {unref():this;ref():this}
}
declare var process:NodeJS.Process;
declare var console:{log(...x:any[]):void;info(...x:any[]):void;debug(...x:any[]):void;warn(...x:any[]):void;error(...x:any[]):void;trace(...x:any[]):void;table(...x:any[]):void;dir(...x:any[]):void;dirxml(...x:any[]):void;time(...x:any[]):void;group(...x:any[]):void};
declare function setTimeout<T extends any[]>(callback:(...args:T)=>void,delay?:number,...args:T):NodeJS.Timeout;
declare function clearTimeout(handle:NodeJS.Timeout|undefined):void;
declare module 'node:process' { export = process; }
declare module 'process' {import process = require('node:process');export = process;}
`

func ownControls() []string {
	return []string{
		"export {};console.log('x');process.exit(0);",
		"export {};process.exit(0);console.log('x');process.exit(1);",
		"export {};if(flag){console.log('x')}else{process.exit(1)};",
		"export {};console.log('x');try{run()}catch{process.exit(1)}",
		"export {};try{console.log('x')}catch{process.exit(1)}",
		"export {};for(const x of rows){if(flag)process.exit(1);console.log(x)}",
		"export {};function help(){console.log('x')};help();process.exit(0);",
		"export {};process.stdout.write('x',()=>process.exit(0));",
		"export {};const timeout=new Promise<never>((resolve,reject)=>{setTimeout(reject,10)});Promise.race([work(),timeout]);",
		"export {};Promise.race([work(),new Promise<never>((resolve,reject)=>setTimeout(reject,10))]);",
		"export {};Promise.race([work(),new Promise<never>((resolve,reject)=>{const timer=setTimeout(reject,10)})]);",
		"export {};Promise.race([work(),new Promise<never>((resolve,reject)=>{const timer=setTimeout(reject,10);use(timer)})]);",
		"export {};let timer:NodeJS.Timeout;Promise.race([work(),new Promise<never>((resolve,reject)=>{timer=setTimeout(reject,10)})]);",
		"export {};let timer:NodeJS.Timeout;Promise.race([work(),new Promise<never>((resolve,reject)=>{timer=setTimeout(reject,10)})]);use(timer);",
		"export {};Promise.race([work(),new Promise<never>((resolve,reject)=>{setTimeout(reject,10).unref()})]);",
		"#!/usr/bin/env tsx\nexport {};console.error('usage');process.exit(1);",
		"export {};function main(){console.log('x');process.exit(1)};main();",
		"export {};function main(){console.log('x');process.exit(1)};",
		"export {};onEvent(()=>{console.log('x');process.exit(1)});",
		"export {};import {blockStandardStreams} from '../../libraries/nexus/source/system/StandardStreams';console.log('x');process.exit(1);blockStandardStreams();",
		"export {};import {blockStandardStreams} from '../../libraries/nexus/source/system/StandardStreams';blockStandardStreams();console.log('x');process.exit(1);",
		"export {}; /* 世界 🌍 */\r\nconsole.log('é');process.exit(0);\r\n",
		"export {};if(false){console.log('x')};process.exit(0);",
		"export {};while(false){console.log('x')};process.exit(0);",
		"export {};console.log('x');while(true){};process.exit(0);",
		"export {};for(;;){if(flag){console.log('x');break}}process.exit(0);",
		"export {};for(let i=console.log('x');flag;use(i)){process.exit(0)};",
		"export {};if(flag&&console.log('x'))process.exit(0);",
		"export {};if(flag||console.log('x'))process.exit(0);",
		"export {};console.log(flag?process.exit(0):'x');process.exit(0);",
		"export {};function f(x=console.log('x')){process.exit(0)};f();",
		"export {};function f(x=process.exit(0)){console.log('x');process.exit(1)};f();",
		"export {};try{process.exit(0)}catch{console.log('x');process.exit(1)};",
		"export {};function f(){console.log('x');try{return;}catch{process.exit(1)}}f();",
		"export {};function f(){console.log('x');try{return work()}catch{process.exit(1)}}f();",
		"export {};console.log('x');while(/x/){};process.exit(0);",
		"export {};console.log('x');while(0x0n){if(flag)break};process.exit(0);",
		"export {};console.log('x');while(0x1n){};process.exit(0);",
		"export {};let x:boolean|void;x ||= process.exit(0);console.log('x');process.exit(1);",
	}
}

func oracle(h *harness) string {
	virtual := filepath.Join(h.repository, "cohere/adamic_wave12_next_oracle.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(h.repository, "stage1/cohere/typeaware/wave12_next/oracle.go")}})
	if err != nil {
		h.t.Fatal(err)
	}
	overlay := h.write("oracle-overlay.json", string(data))
	binary := filepath.Join(h.directory, "oracle")
	command := exec.Command("go", "build", "-overlay", overlay, "-o", binary, virtual)
	command.Dir = filepath.Join(h.repository, "cohere")
	h.must("oracle-build", command)
	return binary
}

// Not parallel: native archives, sanitizer subprocesses, and timings share resources.
func TestAgreement(t *testing.T) {
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE12_NEXT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave12_next/suite.a")
	binary := h.build(stage0, "native", entry, archive, false)
	truth := oracle(h)
	declarations := h.write("node.d.ts", nodeDeclarations)
	config := h.write("tsconfig.json", fmt.Sprintf(`{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ES2022"],"noEmit":true},"files":[%q]}`, declarations))
	modules := filepath.Join(directory, "modules/subject")
	streams := filepath.Join(directory, "libraries/nexus/source/system")
	commandLine := filepath.Join(directory, "libraries/nexus/source/command-line")
	for _, path := range []string{modules, streams, commandLine} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	// These are TypeScript reference input fixtures, not Adamic implementation modules.
	if err := os.WriteFile(filepath.Join(streams, "StandardStreams.ts"), []byte("export function blockStandardStreams(){Reflect.get(process.stdout,'_handle')};\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(commandLine, "CommandLineInterface.ts"), []byte("import {blockStandardStreams} from '../system/StandardStreams';export abstract class CommandLineInterface{abstract run():Promise<void>};export function runCommandLineInterface(x:CommandLineInterface){blockStandardStreams();x.run().catch((error)=>{console.error(error);process.exit(1)})};\n"), 0644); err != nil {
		t.Fatal(err)
	}
	prelude := "declare const rows:string[];declare const flag:boolean;declare function work():Promise<string>;declare function run():Promise<void>;declare function use(x:unknown):void;declare function onEvent(f:()=>void):void;\n"
	var paths []string
	for at, source := range append(ownControls(), referenceControls(t, repository)...) {
		if at < len(ownControls()) {
			if strings.HasPrefix(source, "#!") {
				newline := strings.Index(source, "\n")
				source = source[:newline+1] + prelude + source[newline+1:]
			} else {
				source = prelude + source
			}
		}
		path := filepath.Join(modules, fmt.Sprintf("control-%03d.a", at))
		if err := os.WriteFile(path, []byte(source+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	want := h.compare("controls", truth, binary, config, manifest)
	for _, rule := range []string{"no-process-exit-after-output", "no-uncleared-race-timeout", "require-blocking-standard-streams"} {
		if !bytes.Contains(want.stdout, []byte("nexus/correctness-"+rule+"\t")) {
			t.Fatal("missing positive", rule)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "native-asan", entry, sanitized, true)
	h.compare("controls-asan", truth, asan, config, manifest)
	for _, mutant := range []struct{ file, from, to string }{
		{"no_process_exit_after_output.a", "facts.write(call)", "facts.directWrite(call)"},
		{"no_uncleared_race_timeout.a", "this.lost(executor, index)", "!this.lost(executor, index)"},
		{"require_blocking_standard_streams.a", "const reason = shebang ?", "const reason = !shebang ?"},
	} {
		own := filepath.Join(directory, mutant.file+"-source")
		if err := os.MkdirAll(own, 0755); err != nil {
			t.Fatal(err)
		}
		files, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware/wave12_next/*.a"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range files {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if filepath.Base(path) == mutant.file {
				if strings.Count(source, mutant.from) != 1 {
					t.Fatal("nonunique mutant", mutant.file)
				}
				source = strings.Replace(source, mutant.from, mutant.to, 1)
			}
			source = strings.ReplaceAll(source, "'../", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
			source = strings.ReplaceAll(source, filepath.Join(repository, "stage1/cohere/typeaware")+"/../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
			if err := os.WriteFile(filepath.Join(own, filepath.Base(path)), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
		}
		changed := h.build(stage0, mutant.file+"-mutant", filepath.Join(own, "suite.a"), archive, false)
		got := h.must(mutant.file+"-run", exec.Command(changed, config, manifest))
		if len(got.stderr) > 0 || bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("mutant survived", mutant.file)
		}
		t.Logf("%s mutant exits 0, empty stderr, byte comparison catches byte %d", mutant.file, firstDifference(got.stdout, want.stdout))
		os.Remove(changed)
	}
	for _, mutation := range []struct{ name, file, from, to string }{
		{"ancestry-declaration", "bridge/tsgo/checker/symbol_ancestry.go", "out.yes(file.IsDeclarationFile)", "out.yes(false)"},
		{"signature-body", "bridge/tsgo/checker/call_declaration.go", "out.yes(declaration.Body() != nil)", "out.yes(false)"},
		{"import-target", "bridge/tsgo/checker/program_imports.go", "resolvedFile = target.FileName()", `resolvedFile = ""`},
	} {
		overlay := h.overlay(mutation.name, mutation.file, mutation.from, mutation.to)
		mutatedArchive := h.archive(mutation.name, overlay, false)
		changed := h.build(stage0, mutation.name+"-native", entry, mutatedArchive, false)
		got := h.must(mutation.name+"-run", exec.Command(changed, config, manifest))
		if len(got.stderr) > 0 || bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("question mutant survived", mutation.name)
		}
		t.Logf("%s raw question mutant exits 0, empty stderr, byte comparison catches byte %d", mutation.name, firstDifference(got.stdout, want.stdout))
		os.Remove(changed)
		os.Remove(mutatedArchive)
	}
	probe := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,3,'SourceFile',args[2]??''));`)
	sample := h.write("probe.a", "x;\n")
	released := h.build(stage0, "released", probe, archive, false)
	for _, question := range []string{"symbol-ancestry", "call-declaration", "program-imports"} {
		got := h.run("released-"+question, exec.Command(released, config, sample, question))
		code, ok := got.err.(*exec.ExitError)
		if !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released %s: %v %s", question, got.err, got.stderr)
		}
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "")
	retainedArchive := h.archive("released-registry", overlay, false)
	retained := h.build(stage0, "retained", probe, retainedArchive, false)
	h.must("retained-run", exec.Command(retained, config, sample, "program-imports"))
	t.Log("all three new questions reject released handles with panic 70; registry retention mutant exits 0 and is caught")
	os.Remove(retained)
	os.Remove(retainedArchive)

	for _, corpus := range []string{"COMPILER", "REPOSITORY"} {
		roots := os.Getenv("ADAMIC_WAVE12_NEXT_" + corpus + "_MANIFEST")
		if roots == "" {
			continue
		}
		cfg := filepath.Join(repository, "tsconfig.json")
		if corpus == "COMPILER" {
			cfg = filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")
		}
		h.compare(strings.ToLower(corpus), truth, binary, cfg, roots)
		h.compare(strings.ToLower(corpus)+"-asan", truth, asan, cfg, roots)
	}
}
