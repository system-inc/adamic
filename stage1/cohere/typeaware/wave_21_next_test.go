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
	"strconv"
	"strings"
	"testing"
)

// Read the Go rule's own fixture sources, preserving their declarations and
// cases rather than replacing default-library and import-origin tests with names.
func wave21NextFixtures(h *harness, stem, prefix string) []string {
	h.t.Helper()
	tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/nexus", stem+"_test.go"), nil, 0)
	if err != nil {
		h.t.Fatal(err)
	}
	values := map[string]goast.Expr{}
	parents := map[goast.Node]goast.Node{}
	var stack []goast.Node
	goast.Inspect(tree, func(n goast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if len(stack) > 0 {
			parents[n] = stack[len(stack)-1]
		}
		stack = append(stack, n)
		if v, ok := n.(*goast.ValueSpec); ok {
			for i, name := range v.Names {
				if i < len(v.Values) {
					values[name.Name] = v.Values[i]
				}
			}
		}
		return true
	})
	var evaluate func(goast.Expr) string
	evaluate = func(e goast.Expr) string {
		switch n := e.(type) {
		case *goast.BasicLit:
			value, err := strconv.Unquote(n.Value)
			if err != nil {
				h.t.Fatal(err)
			}
			return value
		case *goast.Ident:
			return evaluate(values[n.Name])
		case *goast.BinaryExpr:
			if n.Op == token.ADD {
				return evaluate(n.X) + evaluate(n.Y)
			}
		case *goast.CallExpr:
			if s, ok := n.Fun.(*goast.SelectorExpr); ok && s.Sel.Name == "Join" {
				list := n.Args[0].(*goast.CompositeLit)
				var lines []string
				for _, e := range list.Elts {
					lines = append(lines, evaluate(e))
				}
				return strings.Join(lines, evaluate(n.Args[1]))
			}
		}
		h.t.Fatalf("unsupported fixture expression %T", e)
		return ""
	}
	prelude := evaluate(values[prefix+"Prelude"])
	if fixture, ok := values[prefix+"NexusFiles"].(*goast.CompositeLit); ok {
		for _, e := range fixture.Elts {
			kv := e.(*goast.KeyValueExpr)
			path := filepath.Join(h.directory, evaluate(kv.Key))
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				h.t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(evaluate(kv.Value)), 0644); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	var sources []string
	goast.Inspect(tree, func(n goast.Node) bool {
		if call, ok := n.(*goast.CallExpr); ok {
			if name, ok := call.Fun.(*goast.Ident); ok && name.Name == prefix+"Source" && call.Ellipsis == token.NoPos {
				var lines []string
				for _, e := range call.Args {
					if _, ok := e.(*goast.BasicLit); !ok {
						return true
					}
					lines = append(lines, evaluate(e))
				}
				sources = append(sources, strings.Join(lines, "\n"))
			}
		}
		list, ok := n.(*goast.CompositeLit)
		if !ok {
			return true
		}
		typ, ok := list.Type.(*goast.ArrayType)
		if !ok || typ.Len != nil {
			return true
		}
		element, ok := typ.Elt.(*goast.Ident)
		if !ok || element.Name != "string" {
			return true
		}
		take := false
		switch parent := parents[n].(type) {
		case *goast.CompositeLit:
			if len(parent.Elts) > 1 && parent.Elts[1] == list {
				_, take = parent.Elts[0].(*goast.BasicLit)
			}
		case *goast.KeyValueExpr:
			if name, ok := parent.Key.(*goast.Ident); ok {
				take = name.Name == "before" || name.Name == "after"
			}
		}
		if take {
			var lines []string
			for _, e := range list.Elts {
				lines = append(lines, evaluate(e))
			}
			sources = append(sources, strings.Join(lines, "\n"))
		}
		return true
	})
	var paths []string
	for i, source := range sources {
		name := fmt.Sprintf("%s-%03d.a", stem, i)
		if stem == "correctness_no_discarded_outcome" {
			name = filepath.Join("repository/modules/meta", name)
			if err := os.MkdirAll(filepath.Join(h.directory, "repository/modules/meta"), 0755); err != nil {
				h.t.Fatal(err)
			}
		}
		paths = append(paths, h.write(name, prelude+source+"\nexport {};\n"))
	}
	if len(paths) == 0 {
		h.t.Fatal("no upstream fixtures imported")
	}
	h.t.Logf("%s: %d upstream fixture sources", stem, len(paths))
	return paths
}

// Not parallel: native builds and sanitizer runs share the scratch and CPU budget.
func TestWave21NextAgreement(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE21_NEXT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_21_next_suite.a")
	binary := h.build(stage0, "wave21-next", entry, archive, false)
	oracle := volumeOracle(h, "wave21-next-oracle", "oracle_wave_21_next.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","lib":["ESNext"],"noEmit":true},"include":["**/*.ts"]}`)
	var paths []string
	for _, subject := range []struct{ stem, prefix string }{
		{"correctness_no_collection_misuse", "correctnessNoCollectionMisuse"},
		{"correctness_no_discarded_outcome", "correctnessNoDiscardedOutcome"},
		{"correctness_no_discarded_pure_result", "correctnessNoDiscardedPureResult"},
	} {
		paths = append(paths, wave21NextFixtures(h, subject.stem, subject.prefix)...)
	}
	paths = append(paths, h.write("boundaries.a", `/* 世界 🌍 */
 declare const a: string[]; declare const m: Map<string,number>; declare const s: Set<string>;
 declare const k: 'absent' | 'missing'; declare const optional: Map<string,number> | undefined;
 const tests=['01' in a, '-1' in a, '4294967294' in a, '4294967295' in a, k in a,
 a.length < -0, a.length < 0x0, a.length > -(0x1), -1 <= a.length, a.length < +0,
 m[true], m[null], m[undefined], m[1n], m[k], optional?.['absent']];
 ('漢'.trim()); a.toReversed(); a.toSorted(); a.toSpliced(1,1); a.with(0,'x');
 declare const callback: string | ((x:string)=>string); 'x'.replace('x', callback);
 declare const uncertain: unknown; 'x'.replace('x', uncertain);
 declare const mixed: string[] | {slice(n:number):void}; mixed.slice(1);
 export {};`))

	// Parentheses, namespace scope, incomplete arms, and wrong-origin aliases are
	// boundaries of the declaration-based rule, independent of the alias spelling.
	declarationPath := filepath.Join(directory, "parentheses/nexus/source/structured-text/json/JsonFile.ts")
	if err := os.MkdirAll(filepath.Dir(declarationPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(declarationPath, []byte(`
export type JsonFileWriteOutcomeType = (({ outcome: 'Written' }) | (({ outcome: 'Unwritable'; message: string })));
export declare function write(): JsonFileWriteOutcomeType;
export declare function success(): Extract<JsonFileWriteOutcomeType, { outcome: 'Written' }>;
export namespace Local {
 export type JsonFileWriteOutcomeType = { outcome: 'Written' } | { outcome: 'Unwritable'; message: string };
 export declare function write(): JsonFileWriteOutcomeType;
}
`), 0644); err != nil {
		t.Fatal(err)
	}
	lookalike := filepath.Join(directory, "lookalike/JsonFile.ts")
	if err := os.MkdirAll(filepath.Dir(lookalike), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lookalike, []byte(`export type JsonFileWriteOutcomeType = { outcome: 'Written' } | { outcome: 'Unwritable'; message: string }; export declare function write(): JsonFileWriteOutcomeType;`), 0644); err != nil {
		t.Fatal(err)
	}
	paths = append(paths, h.write("declaration_boundaries.a", `
import {write, success, Local} from './parentheses/nexus/source/structured-text/json/JsonFile';
import {write as elsewhere} from './lookalike/JsonFile';
/* 世界 🌍 */ (write()); success(); Local.write(); elsewhere();
declare const optional: (() => ReturnType<typeof write>) | undefined; optional?.();
declare function later(): Promise<ReturnType<typeof write>>;
async function run() { await (later()); later(); void write(); }
export {};`))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, rule := range []string{"correctness-no-collection-misuse", "correctness-no-discarded-outcome", "correctness-no-discarded-pure-result"} {
		if !bytes.Contains(truth.stdout, []byte("\tnexus/"+rule+"\t")) {
			t.Fatalf("no positive control for %s", rule)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave21-next-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	siblings := filepath.Join(repository, "stage1/cohere/typeaware")
	// Mutate one production decision per rule. Each must still compile and exit 0;
	// the independent oracle's full diagnostic bytes are the only failure signal.
	for _, change := range []struct{ name, file, from, to string }{
		{"collection", "correctness_no_collection_misuse.a", "return value <= 0", "return value < 0"},
		{"outcome", "correctness_no_discarded_outcome.a", "g.seen.length === g.count", "g.seen.length !== g.count"},
		{"pure", "correctness_no_discarded_pure_result.a", "d.kind === 'MethodSignature'", "d.kind !== 'MethodSignature'"},
	} {
		data, err := os.ReadFile(filepath.Join(siblings, change.file))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(data), change.from) != 1 {
			t.Fatal("nonunique mutant", change.name)
		}
		source := strings.Replace(string(data), change.from, change.to, 1)
		files, err := os.ReadDir(siblings)
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			source = strings.ReplaceAll(source, "'./"+file.Name()+"'", "'"+filepath.Join(siblings, file.Name())+"'")
		}
		mutantFile := h.write(change.name+"_mutant.a", source)
		main, err := os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
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
		t.Logf("%s mutant: exit 0, comparison catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	released := h.write("released_queries.a", `import {panic,programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
 const a=programArguments();const file=a[1]??panic('file');const p=tsgoProgram(a[0]??panic('config'),[file]);
 const q=a[2]??panic('question');tsgoRelease(p);console.log(tsgoInspect(p,file,0,3,'CallExpression',q));`)
	probe := h.write("released_probe.a", "f();")
	stale := h.build(stage0, "released-queries", released, archive, false)
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutantStale := h.build(stage0, "released-queries-mutant", released, mutantArchive, false)
	for _, question := range []string{"awaited-shape", "type-declaration-ancestors\n1"} {
		if strings.HasPrefix(question, "type-declaration") {
			// Register the identity before release; only the live registry is mutated.
			data, err := os.ReadFile(released)
			if err != nil {
				t.Fatal(err)
			}
			patched := strings.Replace(string(data), "tsgoRelease(p);", "console.log(tsgoInspect(p,file,0,3,'CallExpression','raw-shape'));tsgoRelease(p);", 1)
			newFile := h.write("released_ancestors.a", patched)
			stale = h.build(stage0, "released-ancestors", newFile, archive, false)
			mutantStale = h.build(stage0, "released-ancestors-mutant", newFile, mutantArchive, false)
		}
		observed := h.run("released-query-run", exec.Command(stale, config, probe, question))
		if code, ok := observed.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(observed.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released query escaped: %v %s", observed.err, observed.stderr)
		}
		h.must("released-query-mutant-run", exec.Command(mutantStale, config, probe, question))
		t.Logf("%s: released panic 70; retaining registry mutant exits 0 and is caught", strings.Split(question, "\n")[0])
	}
	for _, corpus := range []struct{ name, config, manifest string }{
		{"compiler", os.Getenv("ADAMIC_WAVE21_COMPILER_CONFIG"), os.Getenv("ADAMIC_WAVE21_COMPILER_MANIFEST")},
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE21_REPOSITORY_MANIFEST")},
	} {
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
