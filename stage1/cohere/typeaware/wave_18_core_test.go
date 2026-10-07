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

func wave18CoreMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_18_core_suite.a"), archive, false)
}

// Not parallel: archive and sanitizer builds share a bounded scratch filesystem.
func TestWave18CoreAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE18_CORE_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_18_core_suite.a")
	binary := h.build(stage0, "wave18-core", entry, archive, false)
	oracle := volumeOracle(h, "wave18-oracle", "oracle_wave_18_core.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","lib":["ESNext"],"jsx":"preserve","noEmit":true},"sourceExtensions":[".a"],"include":["*.d.ts","*.a","*.tsx"]}`)
	h.write("anchor.d.ts", "export {};\n")
	var paths []string
	var controls []string
	data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/testdata/core_controls.json"))
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
	for _, name := range []string{"no-class-assign", "no-const-assign", "no-constant-binary-expression"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("missing positive control %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave18-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Pin the optional relational arm with the same complete diagnostic serializer.
	for _, mode := range []struct{ name, binary string }{{"relational", binary}, {"relational-asan", asan}} {
		want := h.must(mode.name+"-go", exec.Command(oracle, config, manifest, "--relational"))
		got := h.must(mode.name+"-native", exec.Command(mode.binary, config, manifest, "--relational"))
		if len(got.stderr) != 0 || !bytes.Equal(got.stdout, want.stdout) {
			t.Fatalf("%s disagreement at byte %d: %s", mode.name, firstDifference(got.stdout, want.stdout), got.stderr)
		}
		if !bytes.Contains(want.stdout, []byte("\tconstantRelationalComparison\t")) {
			t.Fatal("relational option has no positive control")
		}
		t.Logf("%s: %d identical finding bytes; %s", mode.name, len(want.stdout), summary(want.stdout))
	}
	for _, m := range []struct{ name, file, from, to string }{
		{"class-assign", "no_class_assign.a", "this.facts.anchor(index) === anchor", "this.facts.anchor(index) !== anchor"},
		{"const-assign", "no_const_assign.a", "this.anchors.includes(this.facts.anchor(index))", "!this.anchors.includes(this.facts.anchor(index))"},
		{"constant-binary", "no_constant_binary_expression.a", "this.fixedNullish(left)", "!this.fixedNullish(left)"},
	} {
		mutant := wave18CoreMutant(h, stage0, archive, m.name, m.file, m.from, m.to)
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
	t.Logf("released registry mutant caught: expected stale handle panic, got %v %s", got.err, strings.TrimSpace(string(got.stderr)))
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
