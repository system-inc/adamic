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

// Extract input strings from upstream tests, not their expected diagnostics.
func wave17NextControls(h *harness) []string {
	h.t.Helper()
	var paths []string
	for _, rule := range []string{"collection_misuse", "discarded_pure_result", "discarded_outcome"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/nexus/correctness_no_"+rule+"_test.go"), nil, 0)
		if err != nil {
			h.t.Fatal(err)
		}
		constants := map[string]string{}
		var literal func(goast.Expr) string
		literal = func(e goast.Expr) string {
			switch x := e.(type) {
			case *goast.BasicLit:
				s, _ := strconv.Unquote(x.Value)
				return s
			case *goast.Ident:
				return constants[x.Name]
			case *goast.BinaryExpr:
				return literal(x.X) + literal(x.Y)
			}
			return ""
		}
		var list func(goast.Expr) []string
		list = func(e goast.Expr) []string {
			var ss []string
			if x, ok := e.(*goast.CompositeLit); ok {
				for _, v := range x.Elts {
					ss = append(ss, literal(v))
				}
			}
			return ss
		}
		prefix := "correctnessNo" + map[string]string{"collection_misuse": "CollectionMisuse", "discarded_pure_result": "DiscardedPureResult", "discarded_outcome": "DiscardedOutcome"}[rule]
		prelude := ""
		for _, decl := range tree.Decls {
			g, ok := decl.(*goast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range g.Specs {
				v, ok := spec.(*goast.ValueSpec)
				if !ok || len(v.Values) != 1 {
					continue
				}
				name := v.Names[0].Name
				constants[name] = literal(v.Values[0])
				if name == prefix+"Prelude" {
					call := v.Values[0].(*goast.CallExpr)
					prelude = strings.Join(list(call.Args[0]), literal(call.Args[1]))
				}
				if name == prefix+"NexusFiles" {
					m := v.Values[0].(*goast.CompositeLit)
					for _, entry := range m.Elts {
						pair := entry.(*goast.KeyValueExpr)
						call := pair.Value.(*goast.CallExpr)
						remote := literal(pair.Key)
						suffix := strings.SplitN(remote, "/nexus/source/", 2)[1]
						alias := filepath.Join(h.directory, "libraries/structure/libraries/nexus/source", suffix)
						if err := os.MkdirAll(filepath.Dir(alias), 0755); err != nil {
							h.t.Fatal(err)
						}
						// Fixture sources are .a; a .ts symlink preserves the exact declaration
						// paths required by production cohere's Nexus identity contract.
						physical := strings.TrimSuffix(alias, ".ts") + ".a"
						if err := os.WriteFile(physical, []byte(strings.Join(list(call.Args[0]), literal(call.Args[1]))), 0644); err != nil {
							h.t.Fatal(err)
						}
						os.Remove(alias)
						if err := os.Symlink(filepath.Base(physical), alias); err != nil {
							h.t.Fatal(err)
						}
					}
				}
			}
		}
		unique := map[string]bool{}
		goast.Inspect(tree, func(n goast.Node) bool {
			if call, ok := n.(*goast.CallExpr); ok {
				if fun, ok := call.Fun.(*goast.Ident); ok && fun.Name == prefix+"Source" {
					ss := []string{}
					for _, a := range call.Args {
						if s := literal(a); s != "" {
							ss = append(ss, s)
						}
					}
					if len(ss) > 0 {
						unique[strings.Join(ss, "\n")] = true
					}
				}
			}
			if row, ok := n.(*goast.CompositeLit); ok && len(row.Elts) >= 2 {
				if _, ok := row.Elts[0].(*goast.BasicLit); ok {
					if lines, ok := row.Elts[1].(*goast.CompositeLit); ok {
						ss := list(lines)
						if len(ss) > 0 {
							unique[strings.Join(ss, "\n")] = true
						}
					}
				}
			}
			return true
		})
		keys := []string{}
		for s := range unique {
			keys = append(keys, s)
		}
		sort.Strings(keys)
		for i, s := range keys {
			name := fmt.Sprintf("%s-%03d.a", rule, i)
			if rule == "discarded_outcome" {
				name = "modules/meta/" + name
			}
			if err := os.MkdirAll(filepath.Dir(filepath.Join(h.directory, name)), 0755); err != nil {
				h.t.Fatal(err)
			}
			paths = append(paths, h.write(name, prelude+s+"\nexport {};\n"))
		}
	}
	return paths
}

func TestWave17NextAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE17_NEXT_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_17_next_suite.a")
	binary := h.build(stage0, "wave17-next", entry, archive, false)
	oracle := volumeOracle(h, "wave17-next-oracle", "oracle_wave_17_next.go")
	h.write("config-root.d.ts", "export {};\n")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ESNext","module":"NodeNext","lib":["ESNext"],"noEmit":true},"files":["config-root.d.ts"]}`)
	paths := wave17NextControls(h)
	paths = append(paths, h.write("edges.a", `/* 世界 🌍 */
export {};
declare const a: string[]; declare const m: Map<string,number>; declare const key: 'x'|'y'; declare const mixed: 'x'|'size';
'00' in a; '4294967294' in a; '4294967295' in a; '-1' in a; '' in a; '1e0' in a; '01' in a;
a.length < -0; a.length < 0x0; a.length === -0x1; a.length < -(1_0); a.length < +0; a.length < -(-1);a.length < 0X1;a.length < 0B1;a.length < 0O1;a.length === -0X1;a.length < 0B0;
m[key];m[mixed];m[null];m[true];m[1n];m[undefined];
declare const union: Map<string,number>|Set<string>; union['missing'];
function generic<T extends string[]>(v:T){ 'missing' in v; v.length <= -1; }
declare const s: string; declare const cb: ((x:string)=>string)|string;declare const unknownArg: unknown;
s.replace('x',cb);s.replace('x',unknownArg);s.trim();(s.trim());s?.trim();s.trim?.();
declare const numbers: readonly number[];numbers.toSorted();numbers.toSorted((a,b)=>a-b);numbers.toReversed();
`))
	// Outcome declaration ancestry, narrowing, generic instantiation, paths,
	// parentheses, promise unwrapping, and deterministic priority across aliases.
	outcome := `import type {JsonFileWriteOutcomeType as W,JsonFileReadOutcomeType as R} from '../../libraries/structure/libraries/nexus/source/structured-text/json/JsonFile';
import type {JsonParseOutcomeType as P} from '../../libraries/structure/libraries/nexus/source/structured-text/json/Json';
export {};declare function both():W|P;both();declare function narrowed():Extract<W,{outcome:'Written'}>;narrowed();
declare function complete():W|undefined; complete();declare function asyncValue():Promise<W>;async function f(){await (asyncValue());asyncValue();void await asyncValue();}
declare function read():R<{世界:string}>;read();
`
	paths = append(paths, h.write("modules/meta/ancestry.a", outcome))
	// Same file suffix but namespace-nested and parenthesized aliases are
	// separate declaration witnesses. Never infer identity from names alone.
	alias := filepath.Join(directory, "extra/nexus/source/structured-text/json/Json.a")
	if err := os.MkdirAll(filepath.Dir(alias), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(alias, []byte(`export type JsonParseOutcomeType = (({outcome:'Parsed'})) | ({outcome:'Invalid'});export declare function parse():JsonParseOutcomeType;
export namespace Nested {export type JsonParseOutcomeType={outcome:'Parsed'}|{outcome:'Invalid'};export declare function parse():JsonParseOutcomeType;}
`), 0644); err != nil {
		t.Fatal(err)
	}
	link := strings.TrimSuffix(alias, ".a") + ".ts"
	os.Remove(link)
	if err := os.Symlink(filepath.Base(alias), link); err != nil {
		t.Fatal(err)
	}
	paths = append(paths, h.write("extra-use.a", `import {parse,Nested} from './extra/nexus/source/structured-text/json/Json';parse();Nested.parse();`))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"collection-misuse", "discarded-outcome", "discarded-pure-result"} {
		if !bytes.Contains(truth.stdout, []byte("\tnexus/correctness-no-"+name+"\t")) {
			t.Fatal("missing positive " + name)
		}
	}
	t.Logf("%d upstream and edge control files", len(paths))
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave17-next-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, m := range []struct{ name, file, from, to string }{
		{"collection", "no_collection_misuse.a", "return value <= 0;", "return value < 0;"},
		{"outcome", "no_discarded_outcome.a", "(seen.get(key)?.length ?? 0) === counts.get(key)", "(seen.get(key)?.length ?? 0) >= 1"},
		{"pure", "no_discarded_pure_result.a", "if(arguments_.some((argument) => this.callable(argument)))", "if(false && arguments_.some((argument) => this.callable(argument)))"},
	} {
		original, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware", m.file))
		if err != nil {
			t.Fatal(err)
		}
		s := strings.ReplaceAll(string(original), "'./", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
		if strings.Count(s, m.from) != 1 {
			t.Fatal("mutant anchor " + m.name)
		}
		rule := h.write("mutant-"+m.name+".a", strings.Replace(s, m.from, m.to, 1))
		data, err := os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		s = strings.ReplaceAll(string(data), "'./", "'"+filepath.Join(repository, "stage1/cohere/typeaware")+"/")
		s = strings.ReplaceAll(s, "'../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")
		s = strings.Replace(s, filepath.Join(repository, "stage1/cohere/typeaware", m.file), rule, 1)
		e := h.write("mutant-suite-"+m.name+".a", s)
		b := h.build(stage0, "mutant-"+m.name, e, sanitized, true)
		got := h.must("mutant-"+m.name+"-run", exec.Command(b, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatal("mutant survived or sanitizer caught " + m.name)
		}
		t.Logf("%s mutant exits 0, empty stderr; Go byte oracle catches byte %d", m.name, firstDifference(got.stdout, truth.stdout))
	}
	// Mutate the new checker operation independently of the Adamic rule.
	awaitOverlay := h.overlay("ancestry-await", "bridge/tsgo/checker/type_declaration_ancestry.go", `if parts[2] == "1" {`, `if false && parts[2] == "1" {`)
	awaitArchive := h.archive("ancestry-await-checker", awaitOverlay, true)
	awaitBinary := h.build(stage0, "ancestry-await-mutant", entry, awaitArchive, true)
	gotAwait := h.must("ancestry-await-mutant-run", exec.Command(awaitBinary, config, manifest))
	if len(gotAwait.stderr) != 0 || bytes.Equal(gotAwait.stdout, truth.stdout) {
		t.Fatal("awaited ancestry mutant survived or sanitizer caught it")
	}
	t.Logf("ancestry await mutant exits 0, empty stderr; Go byte oracle catches byte %d", firstDifference(gotAwait.stdout, truth.stdout))
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const p=tsgoProgram(args[0]??'',[file]);tsgoInspect(p,file,0,1,'Identifier','raw-shape');tsgoRelease(p);console.log(tsgoInspect(p,file,0,1,'Identifier','type-declaration-ancestry\n1\n0'));`)
	probe := h.write("released-input.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if exit, ok := got.err.(*exec.ExitError); !ok || exit.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("new ancestry question rejects released program: panic 70")
	registryOverlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released roots.")
	registryArchive := h.archive("released-registry-checker", registryOverlay, false)
	registryBinary := h.build(stage0, "released-registry-mutant", released, registryArchive, true)
	gotRegistry := h.must("released-registry-mutant-run", exec.Command(registryBinary, config, probe))
	if len(gotRegistry.stderr) != 0 {
		t.Fatal("released registry mutant did not finish cleanly")
	}
	t.Log("released registry mutant exits 0, empty stderr; required panic 70 catches it")

	for _, population := range []struct{ name, root, config, manifest string }{
		{"repository", repository, filepath.Join(repository, "tsconfig.json"), "repository.manifest"},
		{"compiler", os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), "compiler.manifest"},
	} {
		if population.root == "" {
			t.Fatal("ADAMIC_TYPESCRIPT_SOURCE required")
		}
		if population.name == "compiler" {
			pin := h.must("corpus-pin", exec.Command("git", "-C", population.root, "rev-parse", "HEAD"))
			if strings.TrimSpace(string(pin.stdout)) != compilerCommit {
				t.Fatal("wrong TypeScript pin")
			}
		}
		data, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/validation-coverage", population.manifest))
		if err != nil {
			t.Fatal(err)
		}
		paths = nil
		for _, p := range strings.Split(string(data), "\n") {
			if p != "" {
				paths = append(paths, filepath.Join(population.root, p))
			}
		}
		list := h.write(population.name+".manifest", strings.Join(paths, "\n")+"\n")
		h.compare(population.name, oracle, binary, population.config, list)
		h.compare(population.name+"-asan", oracle, asan, population.config, list)
	}
}
