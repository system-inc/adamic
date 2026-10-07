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

func wave22ThirdSources(h *harness) []string {
	unique := map[string]bool{}
	add := func(expr goast.Expr) {
		if literal, ok := expr.(*goast.BasicLit); ok && literal.Kind == token.STRING {
			value, err := strconv.Unquote(literal.Value)
			if err != nil {
				h.t.Fatal(err)
			}
			unique[value] = true
		}
	}
	for _, name := range []string{"no_global_assign", "no_implicit_globals", "no_implied_eval"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/core", name+"_test.go"), nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			if value, ok := node.(*goast.ValueSpec); ok && len(value.Names) == 1 && value.Names[0].Name == "impliedEvalAmbientGlobals" && len(value.Values) == 1 {
				if literal, ok := value.Values[0].(*goast.BasicLit); ok {
					text, err := strconv.Unquote(literal.Value)
					if err != nil {
						h.t.Fatal(err)
					}
					h.write("ambient-globals.d.ts", text)
				}
			}

			if field, ok := node.(*goast.KeyValueExpr); ok {
				if key, ok := field.Key.(*goast.Ident); ok && (key.Name == "sourceText" || key.Name == "Code" || key.Name == "source") {
					add(field.Value)
				}
			}
			if assignment, ok := node.(*goast.AssignStmt); ok {
				for _, rhs := range assignment.Rhs {
					values, ok := rhs.(*goast.CompositeLit)
					if !ok {
						continue
					}
					array, ok := values.Type.(*goast.ArrayType)
					if !ok {
						continue
					}
					structure, ok := array.Elt.(*goast.StructType)
					if !ok {
						continue
					}
					sourceIndex := -1
					slot := 0
					for _, field := range structure.Fields.List {
						for _, fieldName := range field.Names {
							if fieldName.Name == "sourceText" || fieldName.Name == "source" {
								sourceIndex = slot
							}
							slot++
						}
					}
					if sourceIndex < 0 {
						continue
					}
					for _, entry := range values.Elts {
						if row, ok := entry.(*goast.CompositeLit); ok && sourceIndex < len(row.Elts) {
							if _, keyed := row.Elts[sourceIndex].(*goast.KeyValueExpr); !keyed {
								add(row.Elts[sourceIndex])
							}
						}
					}
				}
			}
			return true
		})
	}
	for _, source := range []string{
		"/* 世界 🌍 */\r\nObject = 1;String++;for(Array of []){}\r\n",
		"Object.x=1;function f(Object){Object=1;}({Object=0}= {});",
		"var a=1;function f(){}let b=1;const c=2;class D{}",
		"var a=1;export {};",
		"leaked = 1;for(other in {}){}for(third of []){}[fourth,...fifth]=[];",
		"'use strict';a=1;function f(){b=2;}",
		"function f(){'use strict';a=1;}function g(){b=1;}class D{f(){c=1;}}",
		"setTimeout('x',1);setInterval(`x${1}`,1);globalThis.setTimeout('x');window.window['setTimeout']('x');",
		"function f(setTimeout,window){setTimeout('x');window.setTimeout('x');}setTimeout(()=>{},1);",
		"({ Object = String } = {});({Object: Array}= {});[...Math] = [];",
		"async function forAwait(){for await(Object of []){}for await(leakAwait of []){}}",
	} {
		unique[source] = true
	}
	result := make([]string, 0, len(unique))
	for source := range unique {
		result = append(result, source)
	}
	sort.Strings(result)
	return result
}

func wave22ThirdSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_22_third_suite.a"), archive, false)
}

// Not parallel: native builds, sanitizers and timing share the worker.
func TestWave22ThirdAgreement(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE22_THIRD_ARTIFACTS")
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
	binary := h.build(stage0, "wave-22-third", filepath.Join(repository, "stage1/cohere/typeaware/wave_22_third_suite.a"), archive, false)
	oracle := volumeOracle(h, "wave-22-third-oracle", "oracle_wave_22_third.go")
	config := h.write("controls-tsconfig.json", `{ "compilerOptions": {"strict":true,"target":"ES2022","lib":["ES2022","DOM"],"moduleDetection":"auto","allowJs":true,"types":[]} }`)
	var paths []string
	for at, source := range wave22ThirdSources(h) {
		for _, extension := range []string{"js", "a"} {
			paths = append(paths, h.write(fmt.Sprintf("control-%03d.%s", at, extension), source+"\n"))
		}
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid-controls", exec.Command(oracle, config, manifest, "--valid-sources"))
	accepted := strings.Fields(string(valid.stdout))
	t.Logf("controls: %d accepted, %d rejected by independent parser", len(accepted), len(paths)-len(accepted))
	manifest = h.write("valid-controls.manifest", strings.Join(accepted, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"no-global-assign", "no-implicit-globals", "no-implied-eval"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive controls for %s", name)
		}
	}
	if os.Getenv("ADAMIC_WAVE22_THIRD_CONTROLS_ONLY") != "" {
		return
	}
	asanArchive := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-22-third-asan", filepath.Join(repository, "stage1/cohere/typeaware/wave_22_third_suite.a"), asanArchive, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for at, source := range []string{
		"Object=1;({Object,String=0}= {});[...Array]=[];async function f(){for await(Math of []){}};export {};",
		"setTimeout('x'+value);execScript('x');window.window['execScript']('x');self.self.setInterval(`x${value}`);function f(setTimeout){setTimeout('x');}export {};",
		"var isolated = 1;function topLevel(){}let skipped=2;leaky=1;function f(){'use strict';declined=1;}class C{f(){declinedClass=1;}}",
		"using resource = null;var plain=1;let lexical=2;",
	} {
		for _, extension := range []string{"js", "a"} {
			file := h.write(fmt.Sprintf("isolated-%d.%s", at, extension), source+"\n")
			list := h.write("isolated.manifest", file+"\n")
			h.compare(fmt.Sprintf("isolated-%d-%s", at, extension), oracle, binary, config, list)
			h.compare(fmt.Sprintf("isolated-%d-%s-asan", at, extension), oracle, asan, config, list)
		}
	}

	for _, change := range []struct{ name, file, from, to string }{
		{"global-provenance", "no_global_assign.a", "facts.declarations[0] !== true", "facts.declarations[0] === true"},
		{"global-declaration-span", "no_implicit_globals.a", "this.nonlexical(declaration);", "this.nonlexical(name);"},
		{"timer-string", "no_implied_eval.a", "this.kind(node.children[1] ?? -1) === 'PlusToken'", "this.kind(node.children[1] ?? -1) === 'MinusToken'"},
	} {
		mutant := wave22ThirdSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s survived", change.name)
		}
		t.Logf("%s: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, question := range []string{"global-binding\nnode", "global-source"} {
		probe := h.write("released-probe.a", "x;\n")
		kind, end := "Identifier", 1
		if question == "global-source" {
			kind, end = "SourceFile", 3
		}
		source := h.write("released-next.a", "import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';\nconst args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,"+strconv.Itoa(end)+","+strconv.Quote(kind)+","+strconv.Quote(question)+"));\n")
		stale := h.build(stage0, "released-next", source, archive, false)
		got := h.run("released-next-run", exec.Command(stale, config, probe))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
		}
		if question == "global-binding\nnode" {
			overlay := h.overlay("next-released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released handle.")
			mutantArchive := h.archive("next-released-registry", overlay, false)
			mutant := h.build(stage0, "released-next-mutant", source, mutantArchive, false)
			h.must("released-next-mutant-run", exec.Command(mutant, config, probe))
		}
		t.Logf("%q released handle: panic 70; registry retention checked separately", question)
	}

	for _, corpus := range []struct{ name, root, config string }{{"repository", repository, filepath.Join(repository, "tsconfig.json")}, {"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), ""}} {
		if corpus.root == "" {
			t.Log("compiler omitted: ADAMIC_TYPESCRIPT_SOURCE unset")
			continue
		}
		corpusConfig := corpus.config
		if corpus.name == "compiler" {
			corpusConfig = filepath.Join(corpus.root, "src/compiler/tsconfig.json")
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
		list := h.write(corpus.name+".manifest", strings.Join(paths, "\n")+"\n")
		h.compare(corpus.name, oracle, binary, corpusConfig, list)
		h.compare(corpus.name+"-asan", oracle, asan, corpusConfig, list)
	}
}
