package typeaware

import (
	"bytes"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func wave07Sources(t *testing.T, repository string) []string {
	t.Helper()
	sources := map[string]bool{}
	for _, path := range []string{
		"core/no_undef_init_test.go", "typescript/prefer_for_of_test.go", "typescript/consistent_indexed_object_style_test.go",
	} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules", path), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			literal, ok := n.(*goast.CompositeLit)
			if !ok {
				return true
			}
			add := func(expr goast.Expr) {
				value, ok := expr.(*goast.BasicLit)
				if !ok || value.Kind != token.STRING {
					return
				}
				text, err := strconv.Unquote(value.Value)
				if err != nil {
					t.Fatal(err)
				}
				sources[text] = true
			}
			if array, ok := literal.Type.(*goast.ArrayType); ok {
				if name, ok := array.Elt.(*goast.Ident); ok && name.Name == "string" {
					for _, element := range literal.Elts {
						add(element)
					}
				}
			}
			for _, element := range literal.Elts {
				if keyed, ok := element.(*goast.KeyValueExpr); ok {
					if name, ok := keyed.Key.(*goast.Ident); ok && name.Name == "sourceText" {
						add(keyed.Value)
					}
				}
			}
			// no-undef-init's one-field positional source rows.
			if path == "core/no_undef_init_test.go" && len(literal.Elts) == 1 {
				add(literal.Elts[0])
			}
			return true
		})
	}
	for _, source := range []string{
		"/* 世界 🌍 */\r\nlet é: number | undefined = undefined;\r\nlet other! = undefined;",
		"let a/**/=undefined;let b=/**/undefined;let c=undefined/*keep*/;function f(undefined:number){let d=undefined;}",
		"export interface Dictionary<T> { readonly [key:string]: T }; export declare interface Declared {[x:number]: string}",
		"type Nested = {[key:string]: {[key:string]:Nested}}; interface Self {[key:string]:Self};interface Wrapped {[key:string]:Wrapped[]}",
		"type Mapped = {+readonly [K in 'a'|'b']-?:number};type Minus={-readonly[K in 'a']?:number};type Remap={[K in 'a' as `x${K}`]:number}",
		"type Commented = {/*drop*/[key:string]:number};type Preserved={[key:string /*keep*/]:number};type MapComment={//drop\n[K in 'a']:number}",
		"declare const arr:number[];for(let i=0;i<arr.length;i++){const value=arr[i];const object={i};}for(let i=0x0;i<arr.length;i+=0x1){arr[i]}for(let i=0;i<arr.length;i=1+i){arr[i]}for(let i=0;i<arr.length;i++){let i=3;console.log(i);}",
		"declare const arr:number[];for(let i=0;i<arr.length;i++){({foo:arr[i]}= {foo:1});}for(let i=0;i<arr.length;i++){({foo:arr[i]})={foo:1};}",
	} {
		sources[source] = true
	}
	var result []string
	for source := range sources {
		if source != "" {
			result = append(result, source)
		}
	}
	sort.Strings(result)
	return result
}

func wave07Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := os.ReadDir(filepath.Join(h.repository, "stage1/cohere/typeaware"))
	if err != nil {
		h.t.Fatal(err)
	}
	for _, fileInfo := range files {
		filename := fileInfo.Name()
		if !strings.HasSuffix(filename, ".ts") && !strings.HasSuffix(filename, ".a") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(h.repository, "stage1/cohere/typeaware", filename))
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filename == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique %s mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		if err := os.WriteFile(filepath.Join(directory, filename), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_07_suite.a"), archive, false)
}

// Not parallel: builds, sanitizer processes and timing runs share the machine.
func TestWave07AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_WAVE07_ARTIFACTS"); path != "" {
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_07_suite.a")
	binary := h.build(stage0, "wave07", entry, archive, false)
	oracle := volumeOracle(h, "wave07-oracle", "oracle_wave_07.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range wave07Sources(t, repository) {
		// These imported error-recovery probes contain an empty index signature,
		// which the native parser explicitly refuses rather than recovering.
		if strings.Contains(source, "{ [] }") || strings.Contains(source, "\n  [];") {
			t.Logf("excluded empty-index-signature recovery control: %q", source)
			continue
		}
		if strings.Contains(source, "yield ") {
			source = "function* generated() {\n" + source + "\n}"
		}
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid-sources", exec.Command(oracle, config, manifest, "--valid-sources"))
	manifest = h.write("valid-controls.manifest", string(valid.stdout))
	t.Logf("controls: %d extracted sources, %d parse-valid", len(paths), strings.Count(string(valid.stdout), "\n"))
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-undef-init", "@typescript-eslint/prefer-for-of", "@typescript-eslint/consistent-indexed-object-style"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave07-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"undef", "no_undef_init.a", "this.bindings.rules.byte(removal), this.bindings.rules.byte(node.end)", "this.bindings.rules.byte(removal) + 1, this.bindings.rules.byte(node.end)"},
		{"forof", "prefer_for_of.a", "if(this.body(node.children[3] ?? -1, declaration, array))", "if(!this.body(node.children[3] ?? -1, declaration, array))"},
		{"indexed", "consistent_indexed_object_style.a", "'A record is preferred over an index signature.'", "'A record is preferred over an index signature!'"},
		{"graph", "type_reference_graph.a", "node.symbol === anchor", "node.symbol !== anchor"},
	} {
		mutant := wave07Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s: exit 0, empty stderr, independent Go bytes catch byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	graphOverlay := h.overlay("graph-symbol", "bridge/tsgo/checker/type_reference_graph.go", "r.symbol = p.symbolID(symbol)", "r.symbol = 0")
	graphArchive := h.archive("graph-symbol", graphOverlay, false)
	graphMutant := h.build(stage0, "graph-symbol-mutant", entry, graphArchive, false)
	graphResult := h.must("graph-symbol-run", exec.Command(graphMutant, config, manifest))
	if len(graphResult.stderr) != 0 || bytes.Equal(graphResult.stdout, truth.stdout) {
		t.Fatal("Go graph symbol mutant survived")
	}
	t.Logf("Go graph symbol mutant: exit 0, empty stderr, independent Go bytes catch byte %d", firstDifference(graphResult.stdout, truth.stdout))
	os.Remove(graphArchive)
	os.Remove(graphMutant)
	for _, corpus := range []struct{ name, config, env string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), "ADAMIC_WAVE07_REPOSITORY_MANIFEST"},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), "ADAMIC_WAVE07_COMPILER_MANIFEST"},
	} {
		if manifest := os.Getenv(corpus.env); manifest != "" {
			h.compare(corpus.name, oracle, binary, corpus.config, manifest)
			h.compare(corpus.name+"-asan", oracle, asan, corpus.config, manifest)
			// A quiet alternating pair with full output, separate from builds and checks.
			goRun := h.must(corpus.name+"-timing-go", exec.Command(oracle, corpus.config, manifest))
			cmd := exec.Command(binary, corpus.config, manifest)
			cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			nativeRun := h.must(corpus.name+"-timing-native", cmd)
			if !bytes.Equal(goRun.stdout, nativeRun.stdout) {
				t.Fatal("timing byte mismatch")
			}
			t.Logf("%s: Go process %s, native process %s; Go %s native %s", corpus.name, goRun.elapsed, nativeRun.elapsed, goRun.stderr, nativeRun.stderr)
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','type-reference-graph'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released program: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	if len(got.stderr) != 0 {
		t.Fatal("released mutant stderr")
	}
	t.Log("released-registry mutant: exit 0, caught by required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
