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

var wave16ThirdRuleNames = []string{"nexus/security-no-interpolated-shell-command", "nexus/security-no-interpolated-sql-string", "no-alert"}

func wave16ThirdControls() []string {
	return []string{
		"alert('x'); confirm('x');prompt('x');window.alert();globalThis['prompt']();window[`confirm`]();window[('alert')]();(alert)();alert?.();window?.alert();this.alert();window[alert]();",
		"function alert(){}alert();function f(window:{alert:()=>void},confirm:()=>void){window.alert();confirm();globalThis.prompt();}const o={alert(){}};o.alert();import('./helper.a');",
		"import {execSync,spawn,execFileSync} from 'node:child_process';declare const input:string,n:number,mode:'status'|'diff';execSync(`echo ${input}`);execSync('echo '+input);execSync(`echo ${n}`);execSync(`git ${mode}`);execFileSync('echo',[input]);spawn('echo',[input],{shell:true});spawn(`echo ${input}`,[],{shell:'sh'});",
		"import * as cp from 'node:child_process';declare const input:string;const prefix='echo ';const command=prefix+input;cp.execSync(command);let opaque=`echo ${input}`;cp.execSync(opaque);function f(command:string){cp.execSync(command);}function execSync(x:string){}execSync(`echo ${input}`);",
		"import {exec,spawn} from 'node:child_process';import {promisify} from 'node:util';declare const input:string,n:number,flag:boolean;promisify(exec)(`echo ${input}`);spawn('echo',[input,n,,...[]],{shell:false});spawn('echo',[input],{shell:true});spawn('echo',[input],{shell:flag});spawn('echo',[input],{shell:''});",
		"import {execSync} from 'node:child_process';declare const data:string;execSync(`cat <<'EOF'\\n${data}\\nEOF`);execSync(`cat <<EOF\\n${data}\\nEOF`);execSync(`cat <<-'EOF'\\n${data}\\n\\tEOF`);execSync(`cat <<\\\\EOF\\n${data}\\nEOF`);",
		"import {execSync} from 'node:child_process';declare const data:string;execSync(`cat <<'EOF'\\n${data}\\nEOF\\necho ${data}`);execSync(`cat <<'EOF' <<SECOND\\n${data}\\nEOF\\n${data}\\nSECOND`);execSync(`cat <<< ${data}`);execSync(`cat <<${data}`);",
		"import {execSync} from 'node:child_process';declare const flag:boolean,input:string,n:number;execSync(flag?`echo ${input}`:'echo safe');execSync(`echo ${flag?'safe':'ok'}`);execSync(`echo ${flag?`a ${input}`:`n ${n}`}`);",
		"import {execSync} from 'node:child_process';declare const opener:'x'|'<<EOF',value:string;execSync(`echo ${opener} ${value}`);function f<T extends 'x'|'y'>(p:T){execSync(`echo ${p}`);}declare const u:unknown;execSync(`echo ${u}`);",
		"declare const word:string,n:number,status:'Active'|'Paused';`SELECT name FROM t WHERE name LIKE '%${word}%'`;`SELECT * FROM t WHERE id = '${n}' AND status='${status}'`;`WHERE id = ${word}`;`SELECT * FROM t WHERE id=${word}`;`WHERE id=${n}`;",
		"declare const word:string;`(LOWER(name) LIKE '%${word}%' OR last LIKE '${word}')`;`Matching names LIKE '${word}'`;`SELECT '${word}' to continue`;`delete refused: '${word}'`;`echo 'WHERE name LIKE '${word}' OR all'`;",
		"declare const word:string;`SELECT * FROM t -- ${word}\\nWHERE name='${word}'`;`SELECT * FROM t /* '${word}' */ WHERE name='${word}'`;`SELECT * FROM t WHERE name='${word}`;`SELECT * FROM t # name='${word}'`;`SELECT * FROM t WHERE x--1='${word}'`;",
		"declare const word:string;`SELECT * FROM t WHERE name='a''${word}'`;`SELECT * FROM t WHERE name='a\\\\${word}'`;`SELECT * FROM t WHERE name=\"${word}\"`;`SELECT * FROM t WHERE name='${word}' AND fragment=${word}`;",
		"declare const word:string;const escaped=word.replaceAll('\\\\','\\\\\\\\').replaceAll(\"'\",\"''\");`WHERE name='${escaped}'`;`WHERE name='${word.replaceAll(\"'\",\"''\")}'`;`WHERE name='${word.replace(/\\\\\\\\/g,'\\\\\\\\').replace(/'/g,\"''\")}'`;",
		"declare const literal:\"O'Brien\",safe:'ok'|'yes',unknown:unknown,any:any;`WHERE name='${literal}'`;`WHERE name='${safe}'`;`WHERE name='${unknown}'`;`WHERE name='${any}'`;",
		"type Brand=string&{readonly brand:'x'};declare const branded:Brand,numeric:`${number}-${number}`,textual:`id-${string}`;`WHERE name='${branded}'`;`WHERE name='${numeric}'`;`WHERE name='${textual}'`;`WHERE id=${textual}`;",
		"function f<T extends string>(x:T){return `WHERE name='${x}'`;}function g<T>(x:T){return `WHERE id=${x}`;}declare const sql:(s:TemplateStringsArray,...x:unknown[])=>unknown,word:string;sql`WHERE name='${word}'`;",
		"declare const word:string;`WHERE id=${word.replaceAll('\\\\','\\\\\\\\').replaceAll(\"'\",\"''\")}`;`SELECT * FROM ${word} WHERE id IN (${word}) AND ${word}`;`WHERE id <> ${word}`;",
		"/* 世界 🌍 */\r\nimport {execSync} from 'node:child_process';declare const word:string;execSync(`echo ${word}`);`WHERE name='${word}'`;window.alert();\r\n",
	}
}

func wave16ThirdSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave16_third_suite.a"), archive, false)
}

// All output, including failures and mutants, goes to files. This suite is serial
// because sanitizer archives and corpus runs share the limited scratch volume.
func TestWave16ThirdAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE16_THIRD_ARTIFACTS"); path != "" {
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
	// The shared dispatch source stays untouched. The overlay tests the exact
	// integration patch while its registration remains pending upstream.
	registration := h.overlay("questions-registration", "bridge/tsgo/checker/facts.go", "\tout.text(mode)\n", "\tout.text(mode)\n\tif answer, handled, err := p.additionalAnswer(out, c, node, question); handled { return answer, err }\n")
	archive := h.archive("checker", registration, false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave16_third_suite.a")
	binary := h.build(stage0, "wave16", entry, archive, false)
	oracle := volumeOracle(h, "wave16-oracle", "oracle_wave16_third.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","moduleResolution":"Bundler","jsx":"preserve","lib":["ES2022","DOM"]},"files":["node.d.a"]}`)
	h.write("node.d.a", "declare module 'node:child_process' {\n export function exec(command:string,...args:unknown[]):unknown;\n export namespace exec {export function __promisify__(command:string):unknown;}\n export function execSync(command:string,...args:unknown[]):unknown;\n export function spawn(command:string,...args:unknown[]):unknown;\n export function spawnSync(command:string,...args:unknown[]):unknown;\n export function execFile(command:string,...args:unknown[]):unknown;\n export function execFileSync(command:string,...args:unknown[]):unknown;\n}\ndeclare module 'node:util' {export function promisify(f:typeof import('node:child_process').exec):typeof import('node:child_process').exec.__promisify__;}\n")
	paths := []string{filepath.Join(directory, "node.d.a")}
	for i, source := range wave16ThirdControls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("helper.a", "export let value=1; export const object={x:1}; export default object;\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range wave16ThirdRuleNames {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", registration, true)
	asan := h.build(stage0, "wave16-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Each rule mutant changes a real judgment or report range, finishes normally,
	// and is killed exclusively by the independent cohere diagnostic bytes.
	for _, change := range []struct{ name, file, from, to string }{
		{"shell-heredoc", "no_interpolated_shell_command.a", "quoted.has(i)", "false"},
		{"sql-string", "no_interpolated_sql_string.a", "if((flags & (32 | 8388608)) !== 0) return true;", "if((flags & (32 | 8388608)) !== 0) return false;"},
		{"alert-range", "no_alert.a", "this.rules.byte(call.end)", "this.rules.byte(call.end) + 1"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			mutant := wave16ThirdSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
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
