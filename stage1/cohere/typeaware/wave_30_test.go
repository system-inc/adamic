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
func TestWave30AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE_30_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_30_suite.a")
	binary := h.build(stage0, "wave-30", entry, archive, false)
	oracle := volumeOracle(h, "wave-30-oracle", "oracle_wave_30.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	controls := []string{
		"export const day=new Date().toISOString().slice(0,10);",
		"declare const d:Date;const iso=d.toISOString();export const day=iso.substring(10,0);",
		"declare const d:Date;export const parts=[d.toISOString().substr(0,10),d.toISOString().slice(0,11),d.toISOString().slice(-10,10),d.toISOString().slice(0,0)];",
		"declare const d:Date;export const day=d.toISOString().split('T')[0];export const [other]=d.toISOString().split('T');",
		"declare const d:Date;export const [,time]=d.toISOString().split('T');export const [...all]=d.toISOString().split('T');export const day=d.toISOString().split('T',0)[0];",
		"class Local {toISOString(){return ''}};export const day=new Local().toISOString().slice(0,10);",
		"declare const d:Date;let iso=d.toISOString();const alias=iso;export const day=alias.slice(0,10);",
		"declare const d:Date;const iso=d.toISOString();export function shadow(iso:string){return iso.slice(0,10)};export const day=iso.slice(2,8);",
		"export function read(s:string,callback:(x:unknown)=>void){try{const value=JSON.parse(s);callback(value);}catch{}}",
		"export function read(s:string,callback:(x:unknown)=>void){try{callback(JSON.parse(s));}catch(error){console.log(error)}}",
		"export function read(s:string,callback:(x:unknown)=>void){try{callback(JSON.parse(s));}catch(error){const snapshot={error};}}",
		"export function read(s:string,callback:(x:unknown)=>void){try{callback(JSON.parse(s));}catch(error){const f=(error:number)=>error;}}",
		"export function read(s:string,callback?:(x:unknown)=>void){try{const value=JSON.parse(s);callback!(value);}catch(error){}}",
		"export function read(s:string,{callback}:{callback:(x:unknown)=>void}){try{callback(JSON.parse(s));}catch{}}",
		"export function read(s:string,callback:(x:unknown)=>void){try{try{JSON.parse(s);callback(1);}catch{}}catch{}}",
		"export function read(s:string,callback:(x:unknown)=>void){try{try{JSON.parse(s)}catch{callback(1)}}catch{}}",
		"export function read(s:string,callback:(x:unknown)=>void){try{JSON.parse(s);const f=()=>callback(1);}catch{}}",
		"export function read(s:string){return new Promise((resolve,reject)=>{try{resolve(JSON.parse(s));reject(1);}catch{}})}",
		"export function read(s:string,callback:any){try{JSON.parse(s);callback(1);}catch{}}",
		"const JSON={parse:(s:string)=>s};export function read(s:string,callback:(x:unknown)=>void){try{callback(JSON.parse(s));}catch{}}",
		"declare const d:Date;export const parts=[d.toISOString().slice(0x0,0xa),d.toISOString().slice(0,10.0),d.toISOString().substring(1_0,0)];",
		"export function read(s:string,callback:(x:unknown)=>void){try{JSON.parse(s);try{}finally{callback(1)}}catch{}}",
		"export function read(s:string,callback:(x:unknown)=>void){try{callback(JSON.parse(s));}catch({message}){}}",
		"export function read(s:string,callback:(x:unknown)=>string){try{class C {[callback(JSON.parse(s))]:number;}return C;}catch{}}",
		"/* 世界 🌍 */\r\nexport const day=(new Date().toISOString()).substring(1,9);\r\n",
	}
	paths := []string{}
	for at, source := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", at), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"nexus/consistency-no-iso-string-date-cut", "nexus/correctness-no-callback-in-parse-try"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-30-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, mutation := range []struct{ name, file, from, to string }{
		{"iso-range", "consistency_no_iso_string_date_cut.a", "end <= 10", "end <= 9"},
		{"callback-range", "correctness_no_callback_in_parse_try.a", "rules.byte(subject.end), '')", "rules.byte(subject.end) + 1, '')"},
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
		mutant := h.build(stage0, mutation.name, filepath.Join(sourceDir, "wave_30_suite.a"), archive, false)
		got := h.must(mutation.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("mutant survived: %s", mutation.name)
		}
		t.Logf("%s exits 0; independent Go bytes catch byte %d", mutation.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE_30_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE_30_COMPILER_MANIFEST")},
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
console.log(tsgoInspect(program,file,0,1,'Identifier','symbol-lineage'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released symbol-lineage handle refused with panic 70")
}
