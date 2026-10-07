package typeaware

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
)

// Not parallel: native builds and sanitizer processes share the machine.
// The dispatch is a Go build overlay, not a change to the shared file.
func TestWave19TimeoutPendingRegistration(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE19_TIMEOUT_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	overlay := h.overlay("ancestry-registration", "bridge/tsgo/checker/facts.go", "switch mode {", "switch mode {\n\tcase \"declaration-ancestry\": return p.declarationAncestry(out, c, node, question)")
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("ancestry", overlay, false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_19_timeout.a")
	binary := h.build(stage0, "timeout", entry, archive, false)
	oracle := volumeOracle(h, "timeout-oracle", "oracle_wave_19_timeout.go")
	declarations := h.write("timers.d.ts", "export {};declare global { namespace NodeJS { interface Timeout {unref():this;ref():this;} } function setTimeout(callback:(...args:any[])=>void,delay?:number,...args:any[]):NodeJS.Timeout;namespace setTimeout{const __promisify__:unknown;}function clearTimeout(value:any):void;}")
	_ = declarations
	var paths []string
	prefix := "declare function work():Promise<string>;declare function use(x:unknown):void;"
	bodies := []string{
		"setTimeout(reject,10);", "const timer=setTimeout(reject,10);", "let timer;timer=setTimeout(reject,10);", "void setTimeout(reject,10);", "globalThis.setTimeout(reject,10);",
		"setTimeout(reject,10).unref();", "use(setTimeout(reject,10));", "const timer=setTimeout(reject,10);use(timer);", "const timer=setTimeout(reject,10);use({timer});", "const timer=setTimeout(reject,10);work().then(()=>clearTimeout(timer));", "const nested=()=>setTimeout(reject,10);use(nested);", "let timer;timer=setTimeout(reject,10);timer=setTimeout(reject,20);",
	}
	for _, body := range bodies {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", len(paths)), prefix+"export function run(){return Promise.race([work(),new Promise<never>((resolve,reject)=>{"+body+"})]);}"))
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", len(paths)), prefix+"export function run(){const timeout=new Promise<never>((resolve,reject)=>{"+body+"});return Promise.race([work(),timeout]);}"))
	}
	for _, source := range []string{
		"export function run(){return Promise.race([new Promise<void>((resolve)=>setTimeout(resolve,10))]);}",
		"export function run(){let timer;try{return Promise.race([new Promise<void>((resolve)=>{timer=setTimeout(resolve,10);})]);}finally{clearTimeout(timer);}}",
		"export function run(){const setTimeout=(callback:any,n:number)=>n;return Promise.race([new Promise<void>((resolve)=>{setTimeout(resolve,10);})]);}",
		"export function run(){let timeout=new Promise<void>((resolve)=>{setTimeout(resolve,10);});return Promise.race([timeout]);}",
		"export function run(){return Promise.all([new Promise<void>((resolve)=>{setTimeout(resolve,10);})]);}",
		"export function run(){return Promise.race([new Promise<void>((resolve)=>{const timer=setTimeout(resolve,10);{const timer=3;use(timer);}})]);}",
		"export function run(){return Promise.race([new Promise<void>((resolve)=>{let timer;use(timer=setTimeout(resolve,10));})]);}",
	} {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", len(paths)), prefix+source))
	}

	// Reconstruct the pinned production table's source lines, not expected diagnostics.
	tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/nexus/correctness_no_uncleared_race_timeout_test.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	production := 0
	goast.Inspect(tree, func(node goast.Node) bool {
		row, ok := node.(*goast.CompositeLit)
		if !ok || len(row.Elts) < 2 || len(row.Elts) > 3 {
			return true
		}
		label, ok := row.Elts[0].(*goast.BasicLit)
		if !ok || label.Kind != token.STRING {
			return true
		}
		lines, ok := row.Elts[1].(*goast.CompositeLit)
		if !ok {
			return true
		}
		var values []string
		for _, element := range lines.Elts {
			literal, ok := element.(*goast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			values = append(values, value)
		}
		if len(values) == 0 {
			return true
		}
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", len(paths)), prefix+"declare function timeoutAfter(n:number):Promise<never>;declare const milliseconds:number;\n"+strings.Join(values, "\n")+"\n"))
		production++
		return true
	})
	if production < 15 {
		t.Fatalf("only %d production table controls", production)
	}
	t.Logf("%d production table controls; %d controls total", production, len(paths))
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]},"files":["control-000.a","timers.d.ts"]}`)
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("unclearedRaceTimeout")) {
		t.Fatal("missing positive")
	}
	sanitized := h.archive("ancestry-asan", overlay, true)
	asan := h.build(stage0, "timeout-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Clone only this unit's modules; preserve imports of unchanged helpers.
	mutantDirectory := filepath.Join(directory, "mutant-source")
	if err := os.MkdirAll(mutantDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"wave_19_timeout.a", "no_uncleared_race_timeout.a", "declaration_ancestry.a", "node_symbol_origin.a"} {
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware", name))
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		for _, helper := range []string{"rules.ts", "bindings.ts", "diagnostic.ts", "frames.ts", "unary_minus.ts"} {
			source = strings.ReplaceAll(source, "'./"+helper+"'", "'"+filepath.ToSlash(filepath.Join(repository, "stage1/cohere/typeaware", helper))+"'")
		}
		source = strings.ReplaceAll(source, "'../../typescript/", "'"+filepath.ToSlash(filepath.Join(repository, "stage1/typescript"))+"/")
		if name == "no_uncleared_race_timeout.a" {
			source = strings.Replace(source, "this.globalTimer(index) && this.lost(executor, index)", "this.globalTimer(index) && false", 1)
		}
		if err := os.WriteFile(filepath.Join(mutantDirectory, name), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	mutant := h.build(stage0, "timeout-mutant", filepath.Join(mutantDirectory, "wave_19_timeout.a"), archive, false)
	got := h.must("timeout-mutant-run", exec.Command(mutant, config, manifest))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("timeout mutant survived")
	}
	t.Logf("timeout mutant: exit 0, empty stderr, Go bytes catch byte %d", firstDifference(got.stdout, truth.stdout))

	for _, population := range []struct{ name, root, config string }{
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"},
		{"repository", repository, "tsconfig.json"},
	} {
		if population.root == "" {
			t.Log("compiler corpus absent; explicitly skipped")
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.name+".manifest"))
		if err != nil {
			t.Fatal(err)
		}
		var roots []string
		for _, path := range strings.Split(string(data), "\n") {
			if path != "" {
				roots = append(roots, filepath.Join(population.root, path))
			}
		}
		corpusManifest := h.write(population.name+".manifest", strings.Join(roots, "\n")+"\n")
		corpusConfig := filepath.Join(population.root, population.config)
		h.compare(population.name, oracle, binary, corpusConfig, corpusManifest)
		h.compare(population.name+"-asan", oracle, asan, corpusConfig, corpusManifest)
		for _, implementation := range []struct{ name, binary string }{{"Go", oracle}, {"native", binary}} {
			command := exec.Command(implementation.binary, corpusConfig, corpusManifest)
			command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			got := h.must("timed-"+population.name+"-"+implementation.name, command)
			t.Logf("%s %s whole process %.6fs; %s", population.name, implementation.name, got.elapsed.Seconds(), strings.TrimSpace(string(got.stderr)))
		}
	}
	probe := h.write("probe.a", "x;")
	releasedSource := h.write("released.a", `import {panic,programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const path=args[1]??panic('path');const program=tsgoProgram(args[0]??panic('config'),[path]);tsgoRelease(program);console.log(tsgoInspect(program,path,0,1,'Identifier','declaration-ancestry'));`)
	released := h.build(stage0, "released", releasedSource, archive, false)
	got = h.run("released-run", exec.Command(released, config, probe))
	if got.err == nil || !strings.Contains(string(got.stderr), "invalid or released checker handle") {
		t.Fatal("released handle accepted")
	}
	t.Log("released handle correctly refused")
	registryOverlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant: retain the released handle.")
	var combined map[string]map[string]string
	registrationData, err := os.ReadFile(overlay)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(registrationData, &combined); err != nil {
		t.Fatal(err)
	}
	var registry map[string]map[string]string
	registryData, err := os.ReadFile(registryOverlay)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(registryData, &registry); err != nil {
		t.Fatal(err)
	}
	for original, replacement := range registry["Replace"] {
		combined["Replace"][original] = replacement
	}
	combinedData, err := json.Marshal(combined)
	if err != nil {
		t.Fatal(err)
	}
	combinedOverlay := h.write("combined-release.json", string(combinedData))
	mutantArchive := h.archive("released-mutant-checker", combinedOverlay, false)
	releaseMutant := h.build(stage0, "released-mutant", releasedSource, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(releaseMutant, config, probe))
	if len(got.stderr) != 0 {
		t.Fatal("registry mutant failed outside refusal check")
	}
	t.Log("released-registry mutant: exit 0; required panic check kills it")
	unregisteredArchive := h.archive("unregistered-checker", "", false)
	unregistered := h.build(stage0, "timeout-unregistered", entry, unregisteredArchive, false)
	got = h.run("unregistered-run", exec.Command(unregistered, config, manifest))
	if got.err == nil || !strings.Contains(string(got.stderr), "unsupported checker question: declaration-ancestry") {
		t.Fatalf("pending registration did not produce its explicit refusal: %s", got.stderr)
	}
	t.Log("unregistered production archive explicitly refuses declaration-ancestry; this is the integration blocker")

	// Save the precise pending dispatch change for integration; no source edit.
	data, err := os.ReadFile(overlay)
	if err != nil {
		t.Fatal(err)
	}
	var mapping map[string]map[string]string
	if err := json.Unmarshal(data, &mapping); err != nil {
		t.Fatal(err)
	}
	t.Log("ancestry dispatch is pending integration; comparisons above use only a scratch Go build overlay")
}
