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

func wave22NextSources(h *harness) []string {
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
	for _, name := range []string{"no_misused_promises", "no_misused_spread"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/typescript", name+"_test.go"), nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			if field, ok := node.(*goast.KeyValueExpr); ok {
				if key, ok := field.Key.(*goast.Ident); ok && (key.Name == "sourceText" || key.Name == "Code") {
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
							if fieldName.Name == "sourceText" {
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
	for _, name := range []string{"concurrency_no_lost_update_test.go", "concurrency_no_lost_update_corpus_test.go"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/nexus", name), nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		prelude := ""
		goast.Inspect(tree, func(node goast.Node) bool {
			if declaration, ok := node.(*goast.ValueSpec); ok && len(declaration.Names) > 0 && declaration.Names[0].Name == "concurrencyNoLostUpdatePrelude" && len(declaration.Values) > 0 {
				if join, ok := declaration.Values[0].(*goast.CallExpr); ok && len(join.Args) > 0 {
					if lines, ok := join.Args[0].(*goast.CompositeLit); ok {
						var texts []string
						for _, line := range lines.Elts {
							if literal, ok := line.(*goast.BasicLit); ok {
								text, err := strconv.Unquote(literal.Value)
								if err != nil {
									h.t.Fatal(err)
								}
								texts = append(texts, text)
							}
						}
						prelude = strings.Join(texts, "\n")
					}
				}
			}
			return true
		})
		goast.Inspect(tree, func(node goast.Node) bool {
			if field, ok := node.(*goast.KeyValueExpr); ok {
				if key, ok := field.Key.(*goast.Ident); ok && key.Name == "source" {
					add(field.Value)
				}
			}
			if call, ok := node.(*goast.CallExpr); ok {
				if name, ok := call.Fun.(*goast.Ident); ok && name.Name == "concurrencyNoLostUpdateSource" {
					var lines []string
					for _, argument := range call.Args {
						if literal, ok := argument.(*goast.BasicLit); ok && literal.Kind == token.STRING {
							text, err := strconv.Unquote(literal.Value)
							if err != nil {
								h.t.Fatal(err)
							}
							lines = append(lines, text)
						} else {
							h.t.Fatal("unsupported production lost-update source argument")
						}
					}
					unique[prelude+strings.Join(lines, "\n")+"\n"] = true
				}
			}
			return true
		})
	}
	for _, source := range []string{
		"declare const p:Promise<number>;if(p){}const o={...p};",
		"declare const p:Promise<number>;declare const q:Promise<number>|undefined;if(q){}p&&1;p??1;if(p&&p){}!p;",
		"declare function use(cb:()=>void):void;use(async()=>{});const x:()=>void=async()=>{};",
		"const map=new Map<string,number>();const a={...(map)};const b={...(map),x:1};const c={...(map,map)};",
		"class Base{x=1;}class Data extends Base{constructor(public y=1){super();}}const a={...new Data()};class Proto{f(){}}const b={...new Proto()};const c={...Proto};",
		"/* 世界 🌍 */\r\ndeclare const é:Promise<number>;if(é){}const a={...(é||{})};\r\n",
		"declare function use(cb:()=>void|Promise<void>):void;use(async()=>{});declare function v(...cb:(()=>void)[]):void;v(async()=>{},async()=>{});",
		"interface B{f():void;}class D implements B{async f(){}}const a:B={async f(){}};",
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

func wave22NextSourceMutant(h *harness, stage0, archive, name, file, from, to string) string {
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
	return h.build(stage0, name, filepath.Join(directory, "wave_22_next_suite.a"), archive, false)
}

// Not parallel: native builds, sanitizers and timing share the worker.
func TestWave22NextAgreement(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE22_NEXT_ARTIFACTS")
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
	binary := h.build(stage0, "wave-22-next", filepath.Join(repository, "stage1/cohere/typeaware/wave_22_next_suite.a"), archive, false)
	oracle := volumeOracle(h, "wave-22-next-oracle", "oracle_wave_22_next.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for at, source := range wave22NextSources(h) {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", at), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("valid-controls", exec.Command(oracle, config, manifest, "--valid-sources"))
	accepted := strings.Fields(string(valid.stdout))
	t.Logf("controls: %d accepted, %d rejected by independent parser (including JSX under .a)", len(accepted), len(paths)-len(accepted))
	manifest = h.write("valid-controls.manifest", strings.Join(accepted, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"@typescript-eslint/no-misused-promises", "@typescript-eslint/no-misused-spread", "nexus/concurrency-no-lost-update"} {
		if !bytes.Contains(truth.stdout, []byte("\t"+name+"\t")) {
			t.Fatalf("no positive controls for %s", name)
		}
	}
	if os.Getenv("ADAMIC_WAVE22_NEXT_CONTROLS_ONLY") != "" {
		return
	}
	asanArchive := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave-22-next-asan", filepath.Join(repository, "stage1/cohere/typeaware/wave_22_next_suite.a"), asanArchive, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"promise-condition", "no_misused_promises.a", "this.report(index, 'conditional',", "this.report(index, 'conditionalMutant',"},
		{"spread-await-edit", "no_misused_spread.a", "new Repair(end, end, ')')", "new Repair(end + 1, end + 1, ')')"},
		{"lost-write-span", "concurrency_no_lost_update.a", "rules.byte(node.end), 'nexus/'", "rules.byte(node.end) + 1, 'nexus/'"},
	} {
		mutant := wave22NextSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s survived", change.name)
		}
		t.Logf("%s: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, question := range []string{"binding-state", "type-operations\nraw\n0"} {
		probe := h.write("released-probe.a", "x;\n")
		source := h.write("released-next.a", "import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';\nconst args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,1,'Identifier',"+strconv.Quote(question)+"));\n")
		stale := h.build(stage0, "released-next", source, archive, false)
		got := h.run("released-next-run", exec.Command(stale, config, probe))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
		}
		if question == "binding-state" {
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
