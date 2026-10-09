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

func profileSourceMutant(h *harness, stage0, archive, name, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	sourceDirectory := filepath.Join(h.repository, "stage1/cohere/typeaware")
	files, err := filepath.Glob(filepath.Join(sourceDirectory, "*.ts"))
	if err != nil {
		h.t.Fatal(err)
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filepath.Base(path) == "shadow.ts" {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique mutant %s", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "volume_suite.ts"), archive, false)
}

// Not parallel: profile controls and optional corpus runs stay separate from
// other suites. Each independent mutant has a parallel child harness.
func TestVolumeProfileAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_VOLUME_PROFILE_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	traceGroup(t)
	stage0 := h.stage0()
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/volume_suite.ts")
	binary := h.build(stage0, "volume", entry, archive, false)
	oracle := volumeOracle(h, "volume-oracle", "oracle_volume.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	sources := append(volumeControls(),
		"function first(){const value=1;}function second(){const value=2;}",
		"const X=1;const C=class X<X>{m(){const X=2;}};",
		"const a=1;const b=2;function f(a:number,b:number){return a+b;}")
	var paths []string
	for i, text := range sources {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), text+"\nexport {};\n"))
	}
	paths = append(paths, h.write("native-globals.d.ts", "declare const console: {log():void};\n"), h.write("native-console.ts", "const detached=console.log;\nexport {};\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.must("controls-go", exec.Command(oracle, config, manifest))
	h.compare("controls", oracle, binary, config, manifest)
	for _, change := range []struct{ name, from, to string }{
		{"binding-slot", "candidates.push(slot);", "candidates.push(0);"},
		{"scope-containment", "if(node.pos > container.pos || node.end < container.end)", "if(node.pos > container.pos)"},
		{"first-binding", "if(candidates[candidates.length - 2] !== at)", "if(true)"},
	} {
		t.Run(change.name, func(t *testing.T) {
			t.Parallel()
			h := h.child(t)
			mutant := profileSourceMutant(h, stage0, archive, change.name, change.from, change.to)
			got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
			if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
				t.Fatalf("%s mutant survived byte oracle", change.name)
			}
			t.Logf("%s: exit 0, independent Go byte oracle catches byte %d; %s", change.name, firstDifference(got.stdout, truth.stdout), summary(got.stdout))
			if err := os.Remove(mutant); err != nil {
				t.Fatal(err)
			}
		})
	}
	// Index kind/range/file separation is independently held to the compiler AST.
	for _, change := range []struct{ name, from, to string }{
		{"index-kind", `candidate.Kind.String() == "Kind"+kind`, `kind != ""`},
		{"root-kind", `start == uint64(source.Pos()) && end == uint64(source.End()) && kind == "SourceFile"`, `start == uint64(source.Pos()) && end == uint64(source.End()) && kind != ""`},
		{"index-end", `nodeRange{uint64(candidate.Pos()), uint64(candidate.End())}`, `nodeRange{uint64(candidate.Pos()), 0}`},
	} {
		t.Run(change.name, func(t *testing.T) {
			t.Parallel()
			h := h.child(t)
			overlay := h.overlay(change.name, "bridge/tsgo/checker/facts.go", change.from, change.to)
			result := h.run(change.name+"-compiler-node-test", exec.Command("go", "test", "-overlay", overlay, "./bridge/tsgo/checker", "-run", "^TestExactIndexMatchesCompilerNodes$", "-count=1"))
			if result.err == nil || !bytes.Contains(result.stdout, []byte("exact index changed compiler node")) {
				t.Fatalf("%s not caught by AST identity: %v %s %s", change.name, result.err, result.stdout, result.stderr)
			}
			t.Logf("%s: compiler AST identity oracle catches wrong selector", change.name)
		})
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "volume-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, corpus := range []struct{ name, manifest, config string }{
		{"repository", os.Getenv("ADAMIC_VOLUME_REPOSITORY_MANIFEST"), filepath.Join(repository, "tsconfig.json")},
		{"compiler", os.Getenv("ADAMIC_VOLUME_COMPILER_MANIFEST"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")},
	} {
		if corpus.manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
}

// Not parallel: a separately linked mutant needs an archive build.
func TestShadowIndexMissingBinding(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_SHADOW_REFUSAL_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	traceGroup(t)
	stage0 := h.stage0()
	archive := h.archive("checker", "", false)
	mutant := profileSourceMutant(h, stage0, archive, "missing-binding", "this.namedScopes.set(binding.name, fresh);", "this.namedScopes.set('wrong', fresh);")
	source := h.write("input.ts", "const x=1;\nexport {};\n")
	manifest := h.write("manifest", source+"\n")
	got := h.run("missing-binding-run", exec.Command(mutant, filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json"), manifest))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte("missing binding index")) {
		t.Fatalf("missing binding index was not refused: %v %s", got.err, got.stderr)
	}
	t.Log("omitted binding key: compiled mutant panics 70, missing binding index")
}
