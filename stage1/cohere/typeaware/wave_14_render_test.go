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

func renderCompare(h *harness, name, oracle, binary, config, manifest string, extra ...string) result {
	args := append([]string{config, manifest}, extra...)
	want := h.must(name+"-go", exec.Command(oracle, args...))
	got := h.must(name+"-native", exec.Command(binary, args...))
	if len(got.stderr) != 0 || !bytes.Equal(got.stdout, want.stdout) {
		h.t.Fatalf("%s differs at %d: native %s; Go %s; stderr %s", name, firstDifference(got.stdout, want.stdout), got.stdout, want.stdout, got.stderr)
	}
	h.t.Logf("%s: %d identical bytes; %s", name, len(want.stdout), summary(want.stdout))
	return want
}

// Not parallel: native archives, sanitizer builds and corpus runs share scratch.
func TestWave14RenderJudgmentsAndMutant(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE14_RENDER_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_14_render.a")
	binary := h.build(stage0, "render", entry, archive, false)
	oracle := volumeOracle(h, "render-oracle", "oracle_wave_14_render.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	controls := []string{
		"declare const count:number;count&&'some';",
		"/* 世界 🌍 */\r\ndeclare const count:number;((count))&&'some';\r\n",
		"declare const count:number|null|undefined;count&&'some';",
		"declare const count:number|false;count&&'some';",
		"declare const count:string|number;count&&'some';",
		"declare const count:boolean;count&&'some';",
		"declare const count:bigint;count&&'some';",
		"declare const count:0n|1n;count&&'some';",
		"type Zero=0n;declare const count:Zero;count&&'some';",
		"declare const count:1n|2n;count&&'some';",
		"declare const count:0|1;count&&'some';",
		"declare const count:1|2;count&&'some';",
		"enum E{Off,On};declare const count:E;count&&'some';",
		"enum E{One=1,Two=2};declare const count:E;count&&'some';",
		"declare const count:number&{readonly brand:1};count&&'some';",
		"declare const count:any;count&&'some';declare const value:unknown;value&&'some';",
		"declare const count:number;declare const flag:boolean;flag&&count&&'some';",
		"declare const count:number;declare const flag:boolean;(flag||count)&&'some';(count||flag)&&'some';",
		"declare const count:number;declare const flag:boolean;(flag?count:false)&&'some';flag?count&&'some':null;",
		"declare const count:number|null;declare const next:number;(count??next)&&'some';count??(next&&'some');",
		"declare const count:number;declare const label:string;(count&&label)||'fallback';",
		"export function value<T extends number>(count:T){count&&'some';}export function unconstrained<T>(count:T){count&&'some';}",
	}
	var paths []string
	for i, source := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := renderCompare(h, "synthetic-children", oracle, binary, config, manifest, "--synthetic-children")
	if !bytes.Contains(truth.stdout, []byte("\tleakedNumberRender\t")) {
		t.Fatal("missing positive judgment")
	}
	renderCompare(h, "synthetic-attributes", oracle, binary, config, manifest, "--synthetic-attribute")
	renderCompare(h, "ordinary-source", oracle, binary, config, manifest)
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "render-asan", entry, sanitized, true)
	renderCompare(h, "synthetic-children-asan", oracle, asan, config, manifest, "--synthetic-children")
	unused := wave14NextMutant(h, stage0, archive, "render-verdict", "no_leaked_number_render.a", "this.union(subject, facts.root().id, 0) === 2", "this.union(subject, facts.root().id, 0) === 1")
	os.Remove(unused)
	mutant := h.build(stage0, "render-mutant", filepath.Join(directory, "render-verdict-source/wave_14_render.a"), archive, false)
	got := h.must("render-mutant-run", exec.Command(mutant, config, manifest, "--synthetic-children"))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
		t.Fatal("render judgment mutant survived")
	}
	t.Logf("render mutant exits 0 with empty stderr; byte comparison catches byte %d", firstDifference(got.stdout, truth.stdout))
	os.Remove(mutant)
	for _, corpus := range []struct{ name, config, manifest string }{
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE14_NEXT_COMPILER_MANIFEST")},
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE14_NEXT_REPOSITORY_MANIFEST")},
	} {
		if corpus.manifest != "" {
			renderCompare(h, corpus.name, oracle, binary, corpus.config, corpus.manifest)
			renderCompare(h, corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		}
	}
	t.Log("Real JSX source remains blocked by the shared native parser; synthetic-context agreement proves the judgments, not JSX parsing")
}
