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

// Not parallel: native archives and sanitizer subprocesses share the scratch budget.
func TestWave21CoreRules(t *testing.T) {
	repository, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	directory := os.Getenv("ADAMIC_WAVE21_CORE_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if e := os.MkdirAll(directory, 0755); e != nil {
		t.Fatal(e)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_21_core_suite.a")
	binary := h.build(stage0, "wave21-core", entry, archive, false)
	oracle := volumeOracle(h, "wave21-core-oracle", "oracle_wave_21_core.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","moduleDetection":"force","lib":["ES2022"],"jsx":"preserve","noEmit":true}}`)
	defaults, allowVoid, jsx := wave21CoreFixtures(h)
	manifest := h.write("controls.manifest", strings.Join(defaults, "\n")+"\n")
	compare := func(name, exe string, paths []string, flags ...string) result {
		t.Helper()
		list := h.write(name+".manifest", strings.Join(paths, "\n")+"\n")
		args := append([]string{config, list}, flags...)
		want := h.must(name+"-go", exec.Command(oracle, args...))
		got := h.must(name+"-native", exec.Command(exe, args...))
		if len(got.stderr) > 0 {
			t.Fatalf("native/sanitizer stderr: %s", got.stderr)
		}
		if !bytes.Equal(got.stdout, want.stdout) {
			i := firstDifference(got.stdout, want.stdout)
			t.Fatalf("%s mismatch byte %d: native %q Go %q", name, i, got.stdout[max(0, i-50):min(len(got.stdout), i+300)], want.stdout[max(0, i-50):min(len(want.stdout), i+300)])
		}
		t.Logf("%s: %d identical finding/fix/suggestion bytes; %s", name, len(want.stdout), summary(want.stdout))
		return want
	}
	truth := compare("controls", binary, defaults)
	compare("allow-void", binary, allowVoid, "--allow-void")
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave21-core-asan", entry, sanitized, true)
	compare("controls-asan", asan, defaults)
	compare("allow-void-asan", asan, allowVoid, "--allow-void")
	siblings := filepath.Join(repository, "stage1/cohere/typeaware")
	for _, change := range []struct{ name, file, from, to string }{
		{"obj-calls", "no_obj_calls.a", "const direct = name === global;", "const direct = name !== global;"},
		{"object-constructor", "no_object_constructor.a", "replacement = ';({})';", "replacement = '({})';"},
		{"promise-return", "no_promise_executor_return.a", "if(!['FunctionExpression', 'ClassExpression'].includes(b.kind)", "if(['FunctionExpression', 'ClassExpression'].includes(b.kind)"},
	} {
		data, e := os.ReadFile(filepath.Join(siblings, change.file))
		if e != nil {
			t.Fatal(e)
		}
		if strings.Count(string(data), change.from) != 1 {
			t.Fatal("nonunique mutant", change.name)
		}
		source := strings.Replace(string(data), change.from, change.to, 1)
		files, e := os.ReadDir(siblings)
		if e != nil {
			t.Fatal(e)
		}
		for _, file := range files {
			source = strings.ReplaceAll(source, "'./"+file.Name()+"'", "'"+filepath.Join(siblings, file.Name())+"'")
		}
		mutantFile := h.write(change.name+"_mutant.a", source)
		main, e := os.ReadFile(entry)
		if e != nil {
			t.Fatal(e)
		}
		source = strings.Replace(string(main), "'./"+change.file+"'", "'"+mutantFile+"'", 1)
		for _, file := range files {
			source = strings.ReplaceAll(source, "'./"+file.Name()+"'", "'"+filepath.Join(siblings, file.Name())+"'")
		}
		source = strings.ReplaceAll(source, "'../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")
		mutant := h.build(stage0, change.name+"-mutant", h.write(change.name+"_mutant_main.a", source), archive, false)
		got := h.must(change.name+"-mutant-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) > 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived or failed outside comparison", change.name)
		}
		t.Logf("%s judgment/edit mutant: exit 0; byte comparison catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	released := h.write("released_queries.a", `import {panic,programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const a=programArguments();const f=a[1]??panic('file');const p=tsgoProgram(a[0]??panic('config'),[f]);tsgoRelease(p);console.log(tsgoInspect(p,f,0,3,'CallExpression','reference-node'));`)
	probe := h.write("released_probe.a", "f();")
	stale := h.build(stage0, "released-queries", released, archive, false)
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutantStale := h.build(stage0, "released-queries-mutant", released, mutantArchive, false)
	observed := h.run("released-query-run", exec.Command(stale, config, probe))
	if code, ok := observed.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(observed.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released query escaped: %v %s", observed.err, observed.stderr)
	}
	h.must("released-query-mutant-run", exec.Command(mutantStale, config, probe))
	t.Log("reference-node: released panic 70; retaining registry mutant exits 0 and is caught")
	for i, path := range jsx {
		list := h.write(fmt.Sprintf("jsx-%d.manifest", i), path+"\n")
		want := h.must("jsx-go", exec.Command(oracle, config, list))
		got := h.run("jsx-native", exec.Command(binary, config, list))
		code, ok := got.err.(*exec.ExitError)
		if !ok || code.ExitCode() != 70 || !strings.Contains(string(got.stderr), "adamic: panic: parser slice expected") {
			t.Fatalf("parser gap changed: %v %s %s", got.err, got.stdout, got.stderr)
		}
		t.Logf("shared parser gap %d: Go %s; native %v: %s", i, summary(want.stdout), got.err, strings.TrimSpace(string(got.stderr)))
	}
	strictOverlay := h.overlay("reference-suffix-mutant", "bridge/tsgo/checker/reference_node.go", `if question != "reference-node" {`, `if false {`)
	strict := h.run("reference-suffix-mutant-test", exec.Command("go", "test", "-overlay", strictOverlay, "./bridge/tsgo/checker", "-run", "^TestReferenceNodeFacts$", "-count=1"))
	if strict.err == nil || !bytes.Contains(strict.stdout, []byte("malformed suffix accepted")) {
		t.Fatalf("reference suffix mutant survived or failed outside assertion: %v %s %s", strict.err, strict.stdout, strict.stderr)
	}
	t.Log("reference suffix mutant compiles; malformed-question assertion catches it")
	for _, corpus := range []struct{ name, config, manifest string }{{"compiler", os.Getenv("ADAMIC_WAVE21_COMPILER_CONFIG"), os.Getenv("ADAMIC_WAVE21_COMPILER_MANIFEST")}, {"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE21_REPOSITORY_MANIFEST")}} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
			native := h.must(corpus.name+"-timed-native", exec.Command(binary, corpus.config, corpus.manifest, "--count"))
			direct := h.must(corpus.name+"-timed-go", exec.Command(oracle, corpus.config, corpus.manifest, "--count"))
			if !bytes.Equal(native.stdout, direct.stdout) {
				t.Fatal("timed counts differ")
			}
			t.Logf("%s quiet timing native %.6fs Go %.6fs", corpus.name, native.elapsed.Seconds(), direct.elapsed.Seconds())
		}
	}
}
