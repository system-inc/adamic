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

var wave22Files = []string{"no_non_null_asserted_nullish_coalescing", "no_unnecessary_qualifier", "no_unused_private_class_members"}

func wave22Sources(h *harness) []string {
	unique := map[string]bool{}
	add := func(expression goast.Expr) {
		if literal, ok := expression.(*goast.BasicLit); ok && literal.Kind == token.STRING {
			text, err := strconv.Unquote(literal.Value)
			if err != nil {
				h.t.Fatal(err)
			}
			unique[text] = true
		}
	}
	for _, name := range wave22Files {
		file := filepath.Join(h.repository, "cohere/internal/lint/rules/typescript", name+"_test.go")
		tree, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			if field, ok := node.(*goast.KeyValueExpr); ok {
				if key, ok := field.Key.(*goast.Ident); ok && key.Name == "sourceText" {
					add(field.Value)
				}
			}
			if assignment, ok := node.(*goast.AssignStmt); ok && len(assignment.Lhs) == 1 && len(assignment.Rhs) == 1 {
				if name, ok := assignment.Lhs[0].(*goast.Ident); ok && name.Name == "cases" {
					if values, ok := assignment.Rhs[0].(*goast.CompositeLit); ok {
						if array, ok := values.Type.(*goast.ArrayType); ok {
							if kind, ok := array.Elt.(*goast.Ident); ok && kind.Name == "string" {
								for _, value := range values.Elts {
									add(value)
								}
							}
						}
					}
				}
			}
			return true
		})
	}
	for _, text := range []string{
		"/* 世界 🌍 */\r\nlet é:string|undefined='a'; (é /* mark */ !) ?? '世界';\r\n",
		"namespace 漢 { export type T=number; export const é=1;const x:漢.T=漢.é; }",
		"class Unicode { private é=1;constructor(private 世界:string='🌍'){} }",
		"namespace A {export namespace B {export const x=1;const y=A.B.x;} const y=A.B.x;}",
		"namespace A {export const x=1;} import B=A; namespace A {const y=B.x;}",
		"let x:string;[...x]=[];x!??'';",
		"let x:string;({a:x}={a:'x'});x!??'';",
		"let x:string;({a=x}={});x!??'';",
		"let x:string;function g(){let x='y';x='z';}x!??'';",
		"let x:string;function g(){x='z';}x!??'';",
		"class C {private a=1;method(){this.a++;return (this.a+=1);}}",
		"class C {private a=1;method(){this.a++;}}",
		"class C {private a=1;method(){class Inner{read(o:C){return o.a;}}}}",
		"class C {#a=1;method(){return class D{#a=2;read(){return this.#a;}};}}",
		"class C {declare private readonly brand:'C';private ['computed']=1;private 123=1;private 'literal'=1;}",
	} {
		unique[text] = true
	}
	result := make([]string, 0, len(unique))
	for text := range unique {
		result = append(result, text)
	}
	sort.Strings(result)
	return result
}

func wave22SourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	// The mutation bundle uses .a for every Adamic source, including unchanged
	// copies of older .ts helpers. External imports still name existing files.
	var paths []string
	for _, extension := range []string{"*.ts", "*.a"} {
		matches, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware", extension))
		if err != nil {
			h.t.Fatal(err)
		}
		paths = append(paths, matches...)
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filepath.Base(path) == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique %s mutation", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		for _, dependency := range paths {
			base := filepath.Base(dependency)
			if strings.HasSuffix(base, ".ts") {
				source = strings.ReplaceAll(source, "./"+base, "./"+strings.TrimSuffix(base, ".ts")+".a")
			}
		}
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
		base := filepath.Base(path)
		if strings.HasSuffix(base, ".ts") {
			base = strings.TrimSuffix(base, ".ts") + ".a"
		}
		if err := os.WriteFile(filepath.Join(directory, base), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_22_suite.a"), archive, false)
}

// Not parallel: sanitizer builds, byte comparisons and timing share this machine.
func TestWave22AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE22_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_22_suite.a")
	binary := h.build(stage0, "wave-22", entry, archive, false)
	oracle := volumeOracle(h, "wave-22-oracle", "oracle_wave_22.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for index, source := range wave22Sources(h) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", index), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	// The independent parser rejects malformed imported sources explicitly.
	valid := h.must("valid-controls", exec.Command(oracle, config, manifest, "--valid-sources"))
	if count := len(strings.Fields(string(valid.stdout))); count != len(paths) {
		t.Fatalf("%d of %d controls parse", count, len(paths))
	}
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range wave22Files {
		rule := "@typescript-eslint/" + strings.ReplaceAll(name, "_", "-")
		if !bytes.Contains(truth.stdout, []byte("\t"+rule+"\t")) {
			t.Fatalf("missing positive control %s", rule)
		}
	}
	t.Logf("%d independent control sources", len(paths))
	if os.Getenv("ADAMIC_WAVE22_CONTROLS_ONLY") != "" {
		return
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-22-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"nullish-suggestion", "no_non_null_asserted_nullish_coalescing.a", "rules.byte(rules.scanner.pos), ''", "rules.byte(rules.scanner.pos) + 1, ''"},
		{"qualifier-fix", "no_unnecessary_qualifier.a", "finding.fixEnd = rules.byte(rules.start(rules.parser.node(name)));", "finding.fixEnd = rules.byte(rules.start(rules.parser.node(name))) + 1;"},
		{"private-read", "no_unused_private_class_members.a", "if(member.used) { continue; }", "if(!member.used) { continue; }"},
	} {
		mutant := wave22SourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s survived", change.name)
		}
		t.Logf("%s: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	scopeOverlay := h.overlay("scope-export-mutant", "bridge/tsgo/checker/scope_export_symbols.go",
		"scoped = c.GetExportSymbolOfSymbol(candidate)", "scoped = candidate")
	scopeArchive := h.archive("scope-export-mutant", scopeOverlay, false)
	scopeMutant := h.build(stage0, "scope-export-mutant", entry, scopeArchive, false)
	scopeGot := h.must("scope-export-mutant-run", exec.Command(scopeMutant, config, manifest))
	if len(scopeGot.stderr) != 0 || bytes.Equal(scopeGot.stdout, truth.stdout) {
		t.Fatal("export normalization mutant survived")
	}
	t.Logf("scope-export-mutant: exit 0, empty stderr, byte oracle catches byte %d", firstDifference(scopeGot.stdout, truth.stdout))
	os.Remove(scopeArchive)
	os.Remove(scopeMutant)
	for _, corpus := range []struct{ name, root, config string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json")},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), ""},
	} {
		if corpus.root == "" {
			t.Log("compiler corpus omitted: ADAMIC_TYPESCRIPT_SOURCE unset")
			continue
		}
		config := corpus.config
		if corpus.name == "compiler" {
			config = filepath.Join(corpus.root, "src/compiler/tsconfig.json")
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", corpus.name+".manifest"))
		if err != nil {
			t.Fatal(err)
		}
		var paths []string
		for _, path := range strings.Split(string(data), "\n") {
			if path != "" {
				paths = append(paths, filepath.Join(corpus.root, path))
			}
		}
		manifest := h.write(corpus.name+".manifest", strings.Join(paths, "\n")+"\n")
		h.compare(corpus.name, oracle, binary, config, manifest)
		h.compare(corpus.name+"-asan", oracle, asan, config, manifest)
		// Optimized process timings include load, parsing, diagnostics and teardown.
		want := h.must(corpus.name+"-timed-go", exec.Command(oracle, config, manifest))
		cmd := exec.Command(binary, config, manifest)
		cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(corpus.name+"-timed-native", cmd)
		if !bytes.Equal(got.stdout, want.stdout) {
			t.Fatal("timed streams differ")
		}
		t.Logf("%s time: native %.6fs Go %.6fs; native %s; Go %s", corpus.name, got.elapsed.Seconds(), want.elapsed.Seconds(), strings.TrimSpace(string(got.stderr)), strings.TrimSpace(string(want.stderr)))
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','scope-export-symbols\n2\nx'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released handle: panic 70; registry mutant exits 0, caught by required refusal")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
