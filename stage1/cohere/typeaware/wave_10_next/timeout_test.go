package wave10next_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type result struct {
	stdout, stderr []byte
	elapsed        time.Duration
	err            error
}
type harness struct {
	t                     *testing.T
	repository, directory string
	next                  int
}

func (h *harness) run(name string, command *exec.Cmd) result {
	h.t.Helper()
	h.next++
	if command.Dir == "" {
		command.Dir = h.repository
	}
	stem := filepath.Join(h.directory, fmt.Sprintf("%03d-%s", h.next, name))
	out, err := os.Create(stem + ".stdout")
	if err != nil {
		h.t.Fatal(err)
	}
	defer out.Close()
	report, err := os.Create(stem + ".stderr")
	if err != nil {
		h.t.Fatal(err)
	}
	defer report.Close()
	command.Stdout = out
	command.Stderr = report
	started := time.Now()
	runError := command.Run()
	elapsed := time.Since(started)
	stdout, err := os.ReadFile(out.Name())
	if err != nil {
		h.t.Fatal(err)
	}
	stderr, err := os.ReadFile(report.Name())
	if err != nil {
		h.t.Fatal(err)
	}
	return result{stdout, stderr, elapsed, runError}
}
func (h *harness) must(name string, command *exec.Cmd) result {
	h.t.Helper()
	r := h.run(name, command)
	if r.err != nil {
		h.t.Fatalf("%s: %v\n%s\n%s", name, r.err, r.stdout, r.stderr)
	}
	return r
}
func (h *harness) write(name, text string) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name)
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		h.t.Fatal(err)
	}
	return path
}
func (h *harness) overlay(name, path, from, to string) string {
	h.t.Helper()
	original := filepath.Join(h.repository, path)
	data, err := os.ReadFile(original)
	if err != nil {
		h.t.Fatal(err)
	}
	if strings.Count(string(data), from) != 1 {
		h.t.Fatalf("nonunique mutant %s", name)
	}
	side := h.write(name+filepath.Ext(path), strings.Replace(string(data), from, to, 1))
	data, err = json.Marshal(map[string]any{"Replace": map[string]string{original: side}})
	if err != nil {
		h.t.Fatal(err)
	}
	return h.write(name+".json", string(data))
}
func (h *harness) archive(name, overlay string, sanitize bool) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name+".a")
	args := []string{"build", "-buildmode=c-archive", "-o", path}
	if overlay != "" {
		args = append(args, "-overlay", overlay)
	}
	args = append(args, "./bridge/tsgo/archive")
	cmd := exec.Command("go", args...)
	if sanitize {
		cmd.Env = append(os.Environ(), "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
	}
	h.must(name, cmd)
	return path
}
func (h *harness) build(stage0, name, entry, archive string, sanitize bool) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name)
	args := []string{"build", entry, "-o", path, "--tsgo", archive}
	if sanitize {
		args = append(args, "--sanitize")
	}
	h.must(name, exec.Command(stage0, args...))
	return path
}
func firstDifference(a, b []byte) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return min(len(a), len(b))
}
func (h *harness) compare(name, oracle, binary, config, manifest string) result {
	h.t.Helper()
	want := h.must(name+"-go", exec.Command(oracle, config, manifest))
	got := h.must(name+"-native", exec.Command(binary, config, manifest))
	if len(got.stderr) != 0 {
		h.t.Fatalf("sanitizer stderr: %s", got.stderr)
	}
	if !bytes.Equal(got.stdout, want.stdout) {
		i := firstDifference(got.stdout, want.stdout)
		h.t.Fatalf("%s mismatch byte %d: native %q Go %q", name, i, got.stdout[max(0, i-50):min(len(got.stdout), i+250)], want.stdout[max(0, i-50):min(len(want.stdout), i+250)])
	}
	h.t.Logf("%s: %d identical finding bytes; %s", name, len(want.stdout), summary(want.stdout))
	return want
}
func summary(data []byte) string {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	return lines[len(lines)-1]
}

func timeoutControls(t *testing.T, repository string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/nexus/correctness_no_uncleared_race_timeout_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var sources []string
	goast.Inspect(file, func(node goast.Node) bool {
		literal, ok := node.(*goast.CompositeLit)
		if !ok {
			return true
		}
		array, ok := literal.Type.(*goast.ArrayType)
		if !ok {
			return true
		}
		typ, ok := array.Elt.(*goast.Ident)
		if !ok || typ.Name != "string" {
			return true
		}
		var lines []string
		for _, element := range literal.Elts {
			text, ok := element.(*goast.BasicLit)
			if !ok || text.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(text.Value)
			if err != nil {
				t.Fatal(err)
			}
			lines = append(lines, value)
		}
		source := strings.Join(lines, "\n")
		if strings.Contains(source, "Promise.race") || strings.Contains(source, "Promise.all") {
			sources = append(sources, source)
		}
		return true
	})
	const race = "Promise.race([work(),new Promise((_r,reject)=>{BODY})]);"
	for _, body := range []string{
		"const timer=setTimeout(reject,10);use({timer});",
		"let timer;timer=setTimeout(reject,10);timer=undefined;",
		"let timer;timer=setTimeout(reject,10);timer++;",
		"let timer;timer=setTimeout(reject,10);use(()=>timer);",
		"let timer;((timer))=((setTimeout(reject,10)));",
		"let timer;use(timer=setTimeout(reject,10));",
		"const {timer}=setTimeout(reject,10);",
		"return setTimeout(reject,10);",
		"class Nested{method(){setTimeout(reject,10)}}",
		"void ((setTimeout(reject,10)));",
		"let timer;timer=timer=setTimeout(reject,10);",
		"const timer=setTimeout(reject,10);{const timer=1;use(timer);}",
	} {
		sources = append(sources, strings.Replace(race, "BODY", body, 1))
	}
	sources = append(sources,
		"const timeout=new Promise((_r,reject)=>setTimeout(reject,10));Promise.race([timeout,timeout]);",
		"import {Promise} from './library';Promise.race([new Promise((_r,reject)=>setTimeout(reject,10))]);",
		"import {setTimeout} from './timers';Promise.race([new Promise((_r,reject)=>setTimeout(reject,10))]);",
		"let timeout=new Promise((_r,reject)=>setTimeout(reject,10));Promise.race([timeout]);",
		"const timeout=new Promise((_r,reject)=>setTimeout(reject,10));Promise.race([...timeout]);",
		"/* 世界 🌍 */\r\nPromise.race([new Promise((_r,reject)=>{setTimeout(reject,10);})]);\r\n",
		"/// <reference lib=\"dom\" />\nPromise.race([new Promise((_r,reject)=>{window.setTimeout(reject,10);})]);",
	)
	return sources
}

func TestTimeoutAgreementAndMutants(t *testing.T) {
	if os.Getenv("ADAMIC_WAVE10_NEXT_VALIDATE") == "" {
		t.Skip("set ADAMIC_WAVE10_NEXT_VALIDATE for native checks")
	}
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE10_NEXT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_10_next/timeout_suite.a")
	binary := h.build(stage0, "timeout", entry, archive, false)
	virtual := filepath.Join(repository, "cohere/adamic_wave10_timeout_oracle.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(repository, "stage1/cohere/typeaware/wave_10_next/testdata/oracle_timeout.go")}})
	if err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(directory, "oracle")
	cmd := exec.Command("go", "build", "-overlay", h.write("oracle.json", string(data)), "-o", oracle, virtual)
	cmd.Dir = filepath.Join(repository, "cohere")
	h.must("oracle-build", cmd)
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["node.d.ts"]}`)
	h.write("node.d.ts", `export {};declare global {namespace NodeJS{interface Timeout{unref():this;ref():this}} function setTimeout<TArgs extends any[]>(callback:(...args:TArgs)=>void,delay?:number,...args:TArgs):NodeJS.Timeout;namespace setTimeout{const __promisify__:unknown;} function clearTimeout(timeout:NodeJS.Timeout|string|number|undefined):void;}`)
	h.write("library.d.ts", `export { Promise };`)
	h.write("timers.d.ts", `export { setTimeout };`)
	prelude := `declare function generateImage(prompt:string):Promise<{path:string}>;declare function work():Promise<string>;declare function timeoutAfter(ms:number):Promise<never>;declare function use(value:unknown):void;declare const milliseconds:number;`
	var paths []string
	for i, source := range timeoutControls(t, repository) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), prelude+"\n"+source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("\tunclearedRaceTimeout\t")) {
		t.Fatal("no positive timeout finding")
	}
	t.Logf("%d controls", len(paths))
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "timeout-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Mutate a native decision only. Both processes must still succeed.
	mutantDirectory := filepath.Join(directory, "mutant-source")
	if err := os.MkdirAll(mutantDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"timeout_suite.a", "runtime_context.a", "no_uncleared_race_timeout.a"} {
		source, err := os.ReadFile(filepath.Join(filepath.Dir(entry), name))
		if err != nil {
			t.Fatal(err)
		}
		text := strings.ReplaceAll(string(source), "'../", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
		// The runner's parser imports are farther above the rule directory.
		text = strings.ReplaceAll(text, filepath.Join(repository, "stage1/cohere/typeaware")+"/../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
		if name == "no_uncleared_race_timeout.a" {
			from := "this.globalTimer(index) && this.lost(executor, index)"
			if strings.Count(text, from) != 1 {
				t.Fatal("nonunique timeout mutant")
			}
			text = strings.Replace(text, from, "this.globalTimer(index) && !this.lost(executor, index)", 1)
		}
		if err := os.WriteFile(filepath.Join(mutantDirectory, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	mutant := h.build(stage0, "timeout-mutant", filepath.Join(mutantDirectory, "timeout_suite.a"), archive, false)
	got := h.must("timeout-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("timeout mutant survived")
	}
	t.Logf("timeout lost-handle mutant: exit 0, empty stderr, byte oracle catches byte %d", firstDifference(got.stdout, truth.stdout))
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE10_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE10_COMPILER_MANIFEST")},
	} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,1,'Identifier','runtime-context\norigin'));`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got = h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released runtime-context query: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	retained := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(retained, config, probe))
	t.Log("released-registry mutant exits 0; required panic check catches it")
}
