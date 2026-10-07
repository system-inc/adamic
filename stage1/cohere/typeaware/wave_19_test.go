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

func wave19Controls() []string {
	return []string{
		"class Box<T>{};const a:Box<string>=new Box();const b:Box<number>=new Box;",
		"class Box<T>{};const a:(Box<string>)=new Box();const b:Box<string>=new Box<string>();const c= new Box<string>();",
		"class Box<T>{};class Holder{field:Box<string>=new Box(); constructor(p:Box<string>=new Box()){} method(p:Box<number>=new Box()){}} function f(p:Box<string>=new Box()){} const g=function(p:Box<number>=new Box()){};const arrow=(p:Box<string>=new Box())=>p;",
		"class Box<T>{};const a: /*before*/ (/*inner*/ Box /*name*/ < /*type*/ string > /*tail*/)=new Box /*callee*/ ();",
		"const a:Uint8Array<ArrayBuffer>=new Uint8Array();const b:Map<string,number>=new Map();",
		"class Uint8Array<T>{};const a:Uint8Array<string>=new Uint8Array();",
		"type Uint8Array<T>={};const a:Uint8Array<string>=new Uint8Array();",
		"class Box<T>{};const a: //before\r\n Box<string>=new Box();",
		"declare const o:any;o['key'];o[('key')];o[((`key`))];o[true];o[false];o[null];o['1'];o['é'];o[1];o[`dynamic${1}`];",
		"declare const o:any;o?.['key'];o['key'].next;o['key']instanceof Object;o[/*keep*/'key'];o['key'/*keep*/];o['x/*y'];o['\\u006bey'];",
		"5['key'];5_000['key'];0x5['key'];5.0['key'];(5)['key'];",
		"Array();new Array;new Array();Array(1,2);new Array(1, /*keep*/ 2);Array(...[1,2]);Array(1);new Array<number>(1,2);Array?.(1,2);(Array)(1,2);",
		"function f(Array:any){return [Array(),new Array(),Array(1,2)];} const g=(Array:any)=>new Array();",
		"declare const obj:{Array:any};obj.Array();new obj.Array();Array(/*lost*/);Array/*lost*/();",
		"/* 世界 🌍 */\r\nclass Box<T>{};const é:Box<string>=new Box;declare const 漢:any;漢['key'];Array(1,2);\r\n",
		"class Box<T>{};const {a}:Box<string>=new Box();class C{['field']:Box<string>=new Box();accessor value:Box<string>=new Box();}",
	}
}

// Extract source inputs, not expected verdicts or edits, from the pinned Go
// production-rule tables. Every extracted input is judged anew under defaults.
func wave19ProductionInputs(t *testing.T, repository string) []string {
	t.Helper()
	unique := map[string]bool{}
	for _, name := range []string{"typescript/consistent_generic_constructors_test.go", "typescript/dot_notation_corpus_data_test.go", "core/no_array_constructor_test.go"} {
		path := filepath.Join(repository, "cohere/internal/lint/rules", name)
		tree, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			var value goast.Expr
			if pair, ok := node.(*goast.KeyValueExpr); ok {
				if key, ok := pair.Key.(*goast.Ident); ok && key.Name == "sourceText" {
					value = pair.Value
				}
			}
			if row, ok := node.(*goast.CompositeLit); ok {
				if strings.Contains(name, "dot_notation_corpus") && len(row.Elts) == 7 {
					value = row.Elts[2]
				}
				if strings.Contains(name, "no_array_constructor") && len(row.Elts) == 3 {
					if first, ok := row.Elts[0].(*goast.BasicLit); ok && first.Kind == token.STRING {
						value = row.Elts[1]
					}
				}
			}
			if literal, ok := value.(*goast.BasicLit); ok && literal.Kind == token.STRING {
				text, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatal(err)
				}
				if strings.TrimSpace(text) != "let.if();" {
					unique[text] = true
				}
			}
			return true
		})
	}
	result := make([]string, 0, len(unique))
	for source := range unique {
		result = append(result, source)
	}
	sort.Strings(result)
	if len(result) < 100 {
		t.Fatalf("only %d production inputs", len(result))
	}
	return result
}

func wave19Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	h.t.Helper()
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.a"))
	if err != nil {
		h.t.Fatal(err)
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			h.t.Fatal(err)
		}
		source := string(data)
		if filepath.Base(path) == file {
			if strings.Count(source, from) != 1 {
				h.t.Fatalf("nonunique %s mutant", name)
			}
			source = strings.Replace(source, from, to, 1)
		}
		// Existing helpers stay at their original absolute paths; new .a modules
		// stay together in scratch so the compiled mutation is the actual rule.
		source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
		entries, _ := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware/*.ts"))
		for _, entry := range entries {
			source = strings.ReplaceAll(source, "./"+filepath.Base(entry), entry)
		}
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0600); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_19.a"), archive, false)
}

// Not parallel: builds, sanitizers, mutants and timings share this machine.
func TestWave19AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE19_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_19.a")
	binary := h.build(stage0, "wave19", entry, archive, false)
	oracle := volumeOracle(h, "wave19-oracle", "oracle_wave_19.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	sources := append(wave19Controls(), wave19ProductionInputs(t, repository)...)
	t.Logf("%d targeted and production-table inputs before independent parse filtering", len(sources))
	for i, source := range sources {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	valid := h.must("controls-parse-filter", exec.Command(oracle, config, manifest, "--valid-sources"))
	validCount := len(strings.Split(strings.TrimSpace(string(valid.stdout)), "\n"))
	if validCount < 100 {
		t.Fatalf("only %d parse-valid inputs", validCount)
	}
	t.Logf("%d parse-valid controls; %d excluded by independent Go parser", validCount, len(paths)-validCount)
	manifest = h.write("controls-valid.manifest", string(valid.stdout))
	gap := h.write("contextual-let-gap.a", "let.if();\nexport {};\n")
	gapManifest := h.write("contextual-let-gap.manifest", gap+"\n")
	gapGo := h.must("contextual-let-go", exec.Command(oracle, config, gapManifest))
	if summary(gapGo.stdout) != "findings 0" {
		t.Fatal("contextual-let Go control changed")
	}
	refused := h.run("contextual-let-native", exec.Command(binary, config, gapManifest))
	if code, ok := refused.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(refused.stderr, []byte("parser slice expected semicolon at 4")) {
		t.Fatalf("contextual-let gap escaped: %v %s", refused.err, refused.stderr)
	}
	t.Log("one production input excluded from agreement: let.if(); is refused by the existing Adamic parser with panic 70, semicolon at 4; Go defaults report zero")

	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"consistent-generic-constructors", "dot-notation", "no-array-constructor"} {
		if !bytes.Contains(truth.stdout, []byte("\t@typescript-eslint/"+name+"\t")) {
			t.Fatalf("missing positive %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave19-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"generic", "consistent_generic_constructors.a", "unwrapped.children.length < 2", "unwrapped.children.length < 3"},
		{"dot", "dot_notation.a", "const prefix = optional ? ''", "const prefix = optional ? '.'"},
		{"array", "no_array_constructor.a", "const replacement = count === 0 ? '[]'", "const replacement = count === 0 ? '[0]'"},
	} {
		mutant := wave19Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	// Compiler-option branches have their own positive and negative controls.
	isolated := h.write("isolated.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ES2022"],"isolatedDeclarations":true,"declaration":true,"noEmit":true},"files":["control-000.a"]}`)
	indexConfig := h.write("index.json", `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"noPropertyAccessFromIndexSignature":true},"files":["control-000.a"]}`)
	indexSource := h.write("index.a", "declare const o:{[key:string]:number;fixed:number};o['dynamic'];o['fixed'];declare const n:{[key:number]:number};n['name'];declare const a:any;a['name'];declare const s:{[key:symbol]:number};s['name'];\nexport {};\n")
	indexManifest := h.write("index.manifest", indexSource+"\n")
	for _, pair := range []struct{ name, config, manifest string }{{"isolated", isolated, manifest}, {"index", indexConfig, indexManifest}} {
		h.compare(pair.name, oracle, binary, pair.config, pair.manifest)
		h.compare(pair.name+"-asan", oracle, asan, pair.config, pair.manifest)
	}

	for _, change := range []struct{ name, file, from, to, config, manifest string }{
		{"isolated-option", "isolated_declarations.go", "out.yes(p.Compiler.Options().IsolatedDeclarations.IsTrue())", "out.yes(!p.Compiler.Options().IsolatedDeclarations.IsTrue())", config, manifest},
		{"index-key", "index_signature_access.go", "out.number(uint64(info.KeyType().Flags()))", "_ = info; out.number(0)", indexConfig, indexManifest},
		{"symbol-origin", "node_symbol_origin.go", "writeSymbolOrigin(out, p, c.GetSymbolAtLocation(node))", "_ = c; writeSymbolOrigin(out, p, nil)", config, manifest},
	} {
		overlay := h.overlay(change.name, "bridge/tsgo/checker/"+change.file, change.from, change.to)
		mutantArchive := h.archive(change.name+"-checker", overlay, false)
		mutant := h.build(stage0, change.name+"-native", entry, mutantArchive, false)
		want := h.must(change.name+"-go", exec.Command(oracle, change.config, change.manifest))
		got := h.must(change.name+"-mutant", exec.Command(mutant, change.config, change.manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, want.stdout) {
			t.Fatalf("%s checker mutant survived", change.name)
		}
		t.Logf("%s checker mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, want.stdout))
		os.Remove(mutantArchive)
		os.Remove(mutant)
	}

	for _, population := range []struct{ name, root, config, manifest string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json"), "repository.manifest"},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), "compiler.manifest"},
	} {
		if population.name == "compiler" && population.root != "" {
			pin := h.must("compiler-pin", exec.Command("git", "-C", population.root, "rev-parse", "HEAD"))
			if strings.TrimSpace(string(pin.stdout)) != compilerCommit {
				t.Fatal("compiler corpus pin differs")
			}
		}
		if population.root == "" {
			t.Fatal("ADAMIC_TYPESCRIPT_SOURCE required for compiler comparison")
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.manifest))
		if err != nil {
			t.Fatal(err)
		}
		var roots []string
		for _, line := range strings.Split(string(data), "\n") {
			if line != "" {
				roots = append(roots, filepath.Join(population.root, line))
			}
		}
		if (population.name == "compiler" && len(roots) != 77) || (population.name == "repository" && len(roots) != 287) {
			t.Fatalf("unexpected %s roots %d", population.name, len(roots))
		}
		manifest := h.write(population.manifest, strings.Join(roots, "\n")+"\n")
		h.compare(population.name, oracle, binary, population.config, manifest)
		h.compare(population.name+"-asan", oracle, asan, population.config, manifest)
		for _, implementation := range []struct{ name, path string }{{"Go", oracle}, {"native", binary}} {
			cmd := exec.Command(implementation.path, population.config, manifest)
			cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			got := h.must("timed-"+population.name+"-"+implementation.name, cmd)
			t.Logf("%s %s whole process %.6fs; %s; %s", population.name, implementation.name, got.elapsed.Seconds(), summary(got.stdout), strings.TrimSpace(string(got.stderr)))
		}
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,2,'SourceFile','isolated-declarations'));`)
	probe := h.write("probe.a", "x;")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released handle: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released program.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	got = h.must("released-mutant-run", exec.Command(mutant, config, probe))
	if len(got.stderr) != 0 {
		t.Fatal("released mutant stderr")
	}
	t.Log("released-registry mutant: exit 0 caught by required panic 70")
	os.Remove(mutantArchive)
	os.Remove(mutant)
}
