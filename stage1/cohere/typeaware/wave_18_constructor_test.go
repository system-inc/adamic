package typeaware

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

func wave18ConstructorMutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*"))
	if err != nil {
		h.t.Fatal(err)
	}
	changed := false
	for _, path := range files {
		extension := filepath.Ext(path)
		if extension != ".a" && extension != ".ts" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filepath.Base(path) == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique mutant %s", name)
			}
			source = strings.Replace(source, from, to, 1)
			changed = true
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	if !changed {
		h.t.Fatal("mutant did not change a source")
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_18_constructor_suite.a"), archive, false)
}

// Not parallel: archive and sanitizer builds share a bounded scratch filesystem.
func TestWave18ConstructorAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE18_CONSTRUCTOR_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_18_constructor_suite.a")
	binary := h.build(stage0, "wave18-constructor", entry, archive, false)
	oracle := volumeOracle(h, "wave18-oracle", "oracle_wave_18_constructor.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","lib":["ESNext"],"jsx":"preserve","noEmit":true},"sourceExtensions":[".a"],"include":["*.d.ts","*.a","*.tsx"]}`)
	h.write("anchor.d.ts", "export {};\n")
	var paths []string
	helper := h.write("constructor-shadows.a", `export default class Str {constructor(_value?:number){}} export class Symbol{} export class BigInt{} export class Number{} export class Boolean{} export function Function(){return 1;}`)
	paths = append(paths, helper)
	var controls []string
	data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/testdata/constructor_controls.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &controls); err != nil {
		t.Fatal(err)
	}
	for i, source := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-new-func", "no-new-native-nonconstructor", "no-new-wrappers"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("missing positive control %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave18-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	noLib := h.write("no-lib.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","noLib":true,"noEmit":true},"sourceExtensions":[".a"],"include":["anchor.d.ts"]}`)
	noLibrary := h.compare("no-lib", oracle, binary, noLib, manifest)
	if summary(noLibrary.stdout) != "findings 0" {
		t.Fatal("unresolved globals have findings")
	}
	h.compare("no-lib-asan", oracle, asan, noLib, manifest)
	// Declaration-file origin is the production test, not default-library origin.
	h.write("ambient.d.ts", "declare const Symbol:any,BigInt:any,String:any,Number:any,Boolean:any,Function:any;\n")
	ambientConfig := h.write("ambient.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","noLib":true,"noEmit":true},"sourceExtensions":[".a"],"files":["ambient.d.ts"]}`)
	ambientTruth := h.compare("ambient", oracle, binary, ambientConfig, manifest)
	for _, name := range []string{"no-new-func", "no-new-native-nonconstructor", "no-new-wrappers"} {
		if !bytes.Contains(ambientTruth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("missing ambient positive %s", name)
		}
	}
	h.compare("ambient-asan", oracle, asan, ambientConfig, manifest)
	for _, m := range []struct{ name, file, from, to string }{
		{"new-func", "no_new_func.a", "this.facts.global(object)", "!this.facts.global(object)"},
		{"native-nonconstructor", "no_new_native_nonconstructor.a", "this.facts.report(callee,", "this.facts.report(index,"},
		{"new-wrappers", "no_new_wrappers.a", "this.facts.global(callee)", "!this.facts.global(callee)"},
	} {
		mutant := wave18ConstructorMutant(h, stage0, archive, m.name, m.file, m.from, m.to)
		got := h.must(m.name+"-mutant-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("mutant %s survived", m.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, Go byte oracle caught byte %d", m.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
	}
	for _, name := range []string{"repository", "compiler"} {
		variable := "ADAMIC_WAVE18_" + strings.ToUpper(name) + "_MANIFEST"
		corpusManifest := os.Getenv(variable)
		if corpusManifest == "" {
			continue
		}
		corpusConfig := filepath.Join(repository, "tsconfig.json")
		if name == "compiler" {
			corpusConfig = filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")
		}
		h.compare(name, oracle, binary, corpusConfig, corpusManifest)
		h.compare(name+"-asan", oracle, asan, corpusConfig, corpusManifest)
		for round := 0; round < 3; round++ {
			commands := []struct{ name, binary string }{{"native", binary}, {"go", oracle}}
			if round%2 == 1 {
				commands[0], commands[1] = commands[1], commands[0]
			}
			for _, command := range commands {
				cmd := exec.Command(command.binary, corpusConfig, corpusManifest, "--count")
				cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
				got := h.must(fmt.Sprintf("%s-%d-%s-timing", name, round, command.name), cmd)
				t.Logf("%s round %d %s: %.6fs %s %s", name, round, command.name, got.elapsed.Seconds(), strings.TrimSpace(string(got.stdout)), strings.TrimSpace(string(got.stderr)))
			}
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoInspect(program,file,0,1,'Identifier','raw-shape');tsgoRelease(program);console.log(tsgoInspect(program,file,0,1,'Identifier','binding-declarations'));`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.run("released-mutant-run", exec.Command(mutant, config, probe))
	// With a retained registry the live type request incorrectly succeeds.
	if got.err != nil || len(got.stderr) != 0 {
		t.Fatal("released registry mutant survived")
	}
	t.Logf("released registry mutant caught: expected stale handle panic, got %v", got.err)
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
