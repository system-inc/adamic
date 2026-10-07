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

var continuation2RuleNames = []string{"nexus/correctness-no-uncleared-race-timeout", "nexus/correctness-no-process-exit-after-output", "nexus/correctness-require-blocking-standard-streams"}

func continuation2Controls() []string {
	return []string{
		`console.log('x');process.exit(0);`,
		`console.log('x');if(flag){process.exit(0);}process.exit(1);`,
		`console['error']('x');process.exit(1);`,
		`process.stdout.write('x');process.exit(0);`,
		`import {stdout,exit} from 'node:process';stdout.write('x');exit(0);`,
		`import * as NodeProcess from 'node:process';NodeProcess.stderr.write('x');NodeProcess.exit(1);`,
		`if(flag){console.log('x');process.exit(0);}process.exit(1);`,
		`if(flag){process.exit(0);}console.log('x');`,
		`while(flag){process.exit(0);console.log('x');}`,
		`while(flag){if(flag)process.exit(0);console.log('x');}`,
		`while(true){console.log('x');break;}process.exit(0);`,
		`function f(){console.log('x');return;process.exit(0);}f();`,
		`try{console.log('x');run();}catch{process.exit(1);}`,
		`console.log('x');try{run();}catch{process.exit(1);}`,
		`try{run();}catch{console.error('x');process.exit(1);}`,
		`function f(){console.log('x');try{return;}finally{process.exit(0);}}f();`,
		`console.log(process.exit(1));process.exit(0);`,
		`console.log('x');try{run();}catch({code=process.exit(1)}){process.exit(0);}`,
		`function f(){console.log('x');}f();process.exit(0);`,
		`import {print} from './writer';print();process.exit(0);`,
		`import {halt} from './writer';halt();process.exit(0);`,
		`import {twice} from './writer';twice();process.exit(0);`,
		`async function f(){console.log('x');}f();process.exit(0);`,
		`async function f(){console.log('x');}await f();process.exit(0);`,
		`function* f(){console.log('x');}f();process.exit(0);`,
		`function f(){console.log('x');process.exit(1);}onEvent(f);`,
		`onEvent(()=>{console.error('x');process.exit(1);});`,
		`process.stdout.write('x',()=>process.exit(0));`,
		`const console={log(){}};console.log('x');process.exit(0);`,
		`const process={exit(){},stdout:{write(){}}};console.log('x');process.exit();`,
		`console.log('x');process['exit'](0);`,
		`consolee.log('x');process.exit(0);`,
		`class A{field=(console.log('x'),process.exit(0));static{console.log('x');process.exit(0);}}`,
		`class A{['x'+(console.log('x'),process.exit(0))](){}}`,
		`function f(flag:boolean){if(flag){console.log('x');}else{return;}process.exit(0);}f(flag);`,
		`outer:for(let i=0;i<2;i++){console.log('x');continue outer;}process.exit(0);`,
		`switch(flag){case true:console.log('x');break;default:process.exit(1);}process.exit(0);`,
		`/* 世界 🌍 */
console.log('é');process.exit(0);`,
		`#!/usr/bin/env tsx
console.log('x');process.exit(0);`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';blockStandardStreams();console.log('x');process.exit(0);`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';console.log('x');process.exit(0);blockStandardStreams();`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';function main(){console.log('x');process.exit(0);blockStandardStreams();}main();`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';function main(){blockStandardStreams();console.log('x');process.exit(0);}main();`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';onEvent(()=>{console.log('x');process.exit(0);});blockStandardStreams();`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';async function main(){await run();console.log('x');process.exit(0);}main();blockStandardStreams();`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';function main(){console.log('x');process.exit(0);}main(blockStandardStreams());`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';const main=()=>{console.log('x');process.exit(0);};main();blockStandardStreams();`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';function main(){if(flag){console.log('x');process.exit(0);}blockStandardStreams();console.log('x');process.exit(1);}main();`,
		`import './load-block';console.log('x');process.exit(0);`,
		`import './load-function';console.log('x');process.exit(0);`,
		`import type {Marker} from './load-block';console.log('x');process.exit(0);`,
		`export {blockStandardStreams} from './nexus/source/system/StandardStreams';console.log('x');process.exit(0);`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';await run();console.log('x');process.exit(0);`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';await using resource={ [Symbol.asyncDispose]:async()=>{} };console.log('x');process.exit(0);`,
		`import {blockStandardStreams} from './nexus/source/system/StandardStreams';for await(const item of asyncItems){console.log(item);process.exit(0);}`,
		`declare const moduleName:string;import(moduleName);console.log('x');process.exit(0);`,
		`declare const moduleName:string;console.log('x');process.exit(0);import(moduleName);`,
		`declare const moduleName:string;require(moduleName);console.log('x');process.exit(0);`,
		`declare const moduleName:string;console.log('x');process.exit(0);require(moduleName);`,
		`function main(){console.log('x');process.exit(0);}main();`,
		`function main(){console.log('x');process.exit(0);}`,
		`const main=()=>{console.log('x');process.exit(0);};onEvent(main);`,
		`new Host(()=>{console.log('x');process.exit(0);});`,
		`Promise.race([work(),new Promise((resolve,reject)=>{setTimeout(reject,ms);})]);`,
		`Promise.race([work(),new Promise((resolve,reject)=>setTimeout(reject,ms))]);`,
		`const timeout=new Promise((resolve,reject)=>{const timer=setTimeout(reject,ms);});Promise.race([work(),timeout]);`,
		`let timeout=new Promise((resolve,reject)=>{setTimeout(reject,ms);});Promise.race([work(),timeout]);`,
		`Promise.race([new Promise((resolve,reject)=>{void setTimeout(reject,ms);})]);`,
		`Promise.race([new Promise((resolve,reject)=>{globalThis.setTimeout(reject,ms);})]);`,
		`Promise.race([new Promise((resolve,reject)=>{const timer=setTimeout(reject,ms);use(timer);})]);`,
		`let timer:NodeJS.Timeout|undefined;Promise.race([new Promise((resolve,reject)=>{timer=setTimeout(reject,ms);})]);`,
		`let timer:NodeJS.Timeout|undefined;Promise.race([new Promise((resolve,reject)=>{timer=setTimeout(reject,ms);})]);clearTimeout(timer);`,
		`Promise.race([new Promise((resolve,reject)=>{const timer=setTimeout(reject,ms);const obj={timer};})]);`,
		`Promise.race([new Promise((resolve,reject)=>{const timer=setTimeout(reject,ms);{const timer=1;use(timer);}})]);`,
		`Promise.race([new Promise((resolve,reject)=>{let timer=setTimeout(reject,ms);timer=undefined;})]);`,
		`Promise.race([new Promise((resolve,reject)=>{const timer=setTimeout(reject,ms);()=>timer;})]);`,
		`Promise.race([new Promise((resolve,reject)=>{return setTimeout(reject,ms);})]);`,
		`Promise.race([new Promise((resolve,reject)=>{setTimeout(reject,ms).unref();})]);`,
		`Promise.race([new Promise((resolve,reject)=>{onEvent(()=>{setTimeout(reject,ms);});})]);`,
		`Promise.race([new Promise(function(resolve,reject){setTimeout(reject,ms);setTimeout(reject,ms);})]);`,
		`const obj={timer:undefined as NodeJS.Timeout|undefined};Promise.race([new Promise((resolve,reject)=>{obj.timer=setTimeout(reject,ms);})]);`,
		`const Promise={race(items:unknown[]){return items;}};Promise.race([work(),new globalThis.Promise((resolve,reject)=>{setTimeout(reject,ms);})]);`,
		`function f(setTimeout:typeof globalThis.setTimeout){Promise.race([new Promise((resolve,reject)=>{setTimeout(reject,ms);})]);}`,
		`import {setTimeout} from 'node:timers/promises';Promise.race([new Promise((resolve,reject)=>{setTimeout(ms);})]);`,
		`const timeout=(new Promise((resolve,reject)=>{setTimeout(reject,ms);}));Promise.race([work(),(timeout)]);`,
		`Promise.race<unknown>([new Promise<unknown>((resolve,reject)=>(setTimeout(reject,ms)))]);`,
		`Promise.race([new Promise((resolve,reject)=>{let a,b;a=b=setTimeout(reject,ms);})]);`,
		`Promise.race([new Promise((resolve,reject)=>{class Nested{static{setTimeout(reject,ms);}}})]);`,
	}
}

const continuation2Platform = "declare var process: NodeJS.Process;\ndeclare module \"node:process\" {\n    global {\n        var process: NodeJS.Process;\n        namespace NodeJS {\n            interface WritableStream { write(chunk: string | Uint8Array, callback?: (error?: Error | null) => void): boolean }\n            interface WriteStream extends WritableStream { columns: number }\n            interface ReadStream { isTTY?: boolean; setRawMode(mode: boolean): this; resume(): this; pause(): this }\n            interface Process {\n                stdout: WriteStream & { fd: 1 };\n                stderr: WriteStream & { fd: 2 };\n                stdin: ReadStream & { fd: 0 };\n                exitCode: number | string | null | undefined;\n                exit(code?: number | string | null): never;\n                on(event: string, listener: (...args: any[]) => void): this;\n            }\n        }\n    }\n    export = process;\n}\ndeclare module \"process\" {\n    import process = require(\"node:process\");\n    export = process;\n}\ndeclare module \"node:console\" {\n    namespace console {\n        interface Console {\n            log(...data: any[]): void;\n            info(...data: any[]): void;\n            debug(...data: any[]): void;\n            warn(...data: any[]): void;\n            error(...data: any[]): void;\n            trace(...data: any[]): void;\n            table(tabularData?: any, properties?: string[]): void;\n            dir(item?: any): void;\n            dirxml(...data: any[]): void;\n            group(...data: any[]): void;\n            time(label?: string): void;\n        }\n    }\n    var console: console.Console;\n    export = console;\n}\ndeclare namespace NodeJS {interface Timeout{unref():this;ref():this;}}\ndeclare function require(name:string):unknown;\ndeclare const flag:boolean;declare const ms:number;declare const asyncItems:AsyncIterable<string>;\ndeclare function run():Promise<void>;declare function work():Promise<string>;\ndeclare function use(value:unknown):void;declare function onEvent(callback:()=>void):void;\ndeclare class Host {constructor(callback:()=>void);}\ndeclare module 'node:timers/promises' {export function setTimeout(delay:number):Promise<void>;}\n"

const continuation2Globals = `export {};import * as console from 'node:console';declare global {interface Console extends console.Console{}var console:Console;
function setTimeout<TArgs extends any[]>(callback:(...args:TArgs)=>void,delay?:number,...args:TArgs):NodeJS.Timeout;
namespace setTimeout{const __promisify__:unknown;}function clearTimeout(timeout:NodeJS.Timeout|undefined):void;}
`

func continuation2SourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.ts"))
	additional, _ := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.a"))
	files = append(files, additional...)
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
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_02_continuation_2_suite.a"), archive, false)
}

// Not parallel: all output, including failures and mutants, goes to files. This suite is serial
// because sanitizer archives and corpus runs share the limited scratch volume.
func TestWave02Continuation2AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE02_CONTINUATION2_ARTIFACTS"); path != "" {
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_02_continuation_2_suite.a")
	binary := h.build(stage0, "coverage", entry, archive, false)
	oracle := volumeOracle(h, "coverage-oracle", "oracle_wave_02_continuation_2.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","lib":["ESNext"],"noEmit":true},"include":["*.ts"]}`)
	h.write("platform.d.ts", continuation2Platform)
	h.write("globals.d.ts", continuation2Globals)
	for _, directory := range []string{"nexus/source/system"} {
		if err := os.MkdirAll(filepath.Join(h.directory, directory), 0755); err != nil {
			t.Fatal(err)
		}
	}
	h.write("nexus/source/system/StandardStreams.ts", `export function blockStandardStreams(){Reflect.get(process.stdout,'_handle');}`)
	h.write("writer.ts", `export function print(){console.log('x');}export function halt():never{console.log('x');process.exit(0);}export function twice(){print();}`)
	h.write("load-block.ts", `import {blockStandardStreams} from './nexus/source/system/StandardStreams';blockStandardStreams();export type Marker=string;`)
	h.write("load-function.ts", `import {blockStandardStreams} from './nexus/source/system/StandardStreams';export function later(){blockStandardStreams();}`)
	paths := []string{}
	for i, source := range continuation2Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}

	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range continuation2RuleNames {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "coverage-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Each rule mutant changes a real judgment or report range, finishes normally,
	// and is killed exclusively by the independent cohere diagnostic bytes.
	for _, change := range []struct{ name, file, from, to string }{
		{"timeout", "no_uncleared_race_timeout.a", "return this.kind(name) === 'Identifier' && this.neverRead(name);", "return this.kind(name) === 'Identifier' && !this.neverRead(name);"},
		{"exit", "no_process_exit_after_output.a", "if(state.length > 0 && !reported.includes(index))", "if(state.length === 0 && !reported.includes(index))"},
		{"blocking", "require_blocking_standard_streams.a", "exits.length > 1", "exits.length > 2"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			mutant := continuation2SourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
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
	// Each raw bridge-question mutant is killed by independent diagnostic bytes.
	for _, change := range []struct{ name, file, from, to string }{
		{"graph", "execution_graph.go", "out.yes(block.Reachable)", "out.yes(true)"},
		{"context", "ast_context.go", "out.yes(ast.IsGlobalScopeAugmentation(current))", "out.yes(false)"},
		{"signature", "resolved_call.go", "out.number(uint64(result.Flags()))", "out.number(uint64(result.Flags()) | 131072)"},
		{"modules", "program_modules.go", "out.yes(ref.typeOnly)", "out.yes(false)"},
	} {
		overlay := h.overlay(change.name, "bridge/tsgo/checker/"+change.file, change.from, change.to)
		alteredArchive := h.archive(change.name, overlay, false)
		alteredBinary := h.build(stage0, change.name, entry, alteredArchive, false)
		altered := h.must(change.name+"-run", exec.Command(alteredBinary, config, manifest))
		if len(altered.stderr) != 0 || bytes.Equal(altered.stdout, truth.stdout) {
			t.Fatalf("%s question mutant survived", change.name)
		}
		t.Logf("%s question mutant: exit 0, Go byte oracle catches byte %d; %s", change.name, firstDifference(altered.stdout, truth.stdout), summary(altered.stdout))
		os.Remove(alteredArchive)
		os.Remove(alteredBinary)
	}

	if manifest := os.Getenv("ADAMIC_WAVE02_CONTINUATION2_REPOSITORY_MANIFEST"); manifest != "" {
		h.compare("repository", oracle, binary, filepath.Join(repository, "tsconfig.json"), manifest)
		h.compare("repository-asan", oracle, asan, filepath.Join(repository, "tsconfig.json"), manifest)
	}
	if manifest := os.Getenv("ADAMIC_WAVE02_CONTINUATION2_COMPILER_MANIFEST"); manifest != "" {
		corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
		if corpus == "" {
			t.Fatal("compiler source is required")
		}
		h.compare("compiler", oracle, binary, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
		h.compare("compiler-asan", oracle, asan, filepath.Join(corpus, "src/compiler/tsconfig.json"), manifest)
	}
	domDirectory := filepath.Join(directory, "dom")
	if err := os.MkdirAll(domDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	domConfig := h.write("dom/tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ESNext","lib":["ESNext","DOM"]},"include":["*.ts"]}`)
	domFile := h.write("dom/input.ts", `Promise.race([new Promise((resolve,reject)=>{setTimeout(reject,10);window.setTimeout(reject,10);globalThis.setTimeout(reject,10);})]);`)
	domManifest := h.write("dom/manifest", domFile+"\n")
	domTruth := h.compare("dom", oracle, binary, domConfig, domManifest)
	if !bytes.Contains(domTruth.stdout, []byte("findings 3\n")) {
		t.Fatal("DOM timer positive controls did not fire")
	}
	h.compare("dom-asan", oracle, asan, domConfig, domManifest)

	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoInspect(program,file,0,1,'Identifier','raw-shape');tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','ast-context'));
`)
	probe := h.write("probe.ts", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released program queried using a new question: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant exits 0, caught by the required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
