package typeaware

import (
	"bytes"
	"encoding/json"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func suiteSources(t *testing.T, repository string) []string {
	t.Helper()
	unique := map[string]bool{}
	for _, name := range []string{"no_unsafe_unary_minus", "related_getter_setter_pairs", "no_unsafe_declaration_merging", "no_unsafe_argument", "restrict_plus_operands", "no_unnecessary_boolean_literal_compare"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules/typescript", name+"_test.go"), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		add := func(n goast.Expr) {
			literal, ok := n.(*goast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return
			}
			text, err := strconv.Unquote(literal.Value)
			if err == nil && (strings.Contains(text, ";") || strings.Contains(text, "\n") || strings.HasPrefix(text, "class ") || strings.HasPrefix(text, "interface ")) {
				unique[text] = true
			}
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			literal, ok := n.(*goast.CompositeLit)
			if !ok {
				return true
			}
			if array, ok := literal.Type.(*goast.ArrayType); ok {
				if kind, ok := array.Elt.(*goast.Ident); ok && kind.Name == "string" {
					for _, element := range literal.Elts {
						add(element)
					}
				}
			}
			if len(literal.Elts) > 1 {
				add(literal.Elts[1])
			}
			for _, element := range literal.Elts {
				if field, ok := element.(*goast.KeyValueExpr); ok {
					if key, ok := field.Key.(*goast.Ident); ok && (key.Name == "sourceText" || key.Name == "source" || key.Name == "code") {
						add(field.Value)
					}
				}
			}
			return true
		})
	}
	// Independent controls pin direction, local declarations, overload selection,
	// union membership, nullable defaults, recursive generics and literal text.
	for _, text := range []string{
		"interface A { get x(): string; set x(v:string|undefined); } interface B { get x(): never; set x(v:string); } interface C { get x(): string; set x(v:never); }",
		"interface A {} export class A {} class B {} interface B {} interface B {} namespace C {} interface C {} class C {}",
		"declare function f(x:number):void; declare function f(x:string):void; declare function g(x:any):void; declare const x:any; f(x);g(x);",
		"declare const x:symbol|string; declare const nr:number|bigint; x+1; nr+nr; /x/+1; /x/+'s'; declare const i: string & {a:1};i+'s';",
		"declare const b:boolean; b===true; !(b!==false); (b||b)===false; declare const x:boolean|undefined; x===true;x!==false; declare const y:boolean|string|undefined;y===true;",
		"/* 世界 🌍 */\r\ninterface é {} export class é {}\r\ndeclare function f(x:'🌍'):void; declare const x:any; f(x);",
		"function generic<T extends Array<any>>(x:T):void { const take=(v:Array<number>):void=>{};take(x); }",
		"type R = Array<R>; declare function r(a:R):void; declare const rr:R;r(rr);",
		"declare function tag(t:TemplateStringsArray,x:number):void; declare const x:any;tag`世界${x}`;",
	} {
		unique[text] = true
	}
	keys := make([]string, 0, len(unique))
	for text := range unique {
		keys = append(keys, text)
	}
	sort.Strings(keys)
	return keys
}

func suiteTSMutant(h *harness, stage0, normal, entry, name, relative, from, to string, imports ...string) string {
	h.t.Helper()
	path := filepath.Join(h.repository, "stage1/cohere/typeaware", relative)
	text, err := os.ReadFile(path)
	if err != nil {
		h.t.Fatal(err)
	}
	if strings.Count(string(text), from) != 1 {
		h.t.Fatalf("nonunique %s anchor", name)
	}
	source := strings.Replace(string(text), from, to, 1)
	// A cloned nominal class must be shared by its decoder and every typed caller.
	for index := 0; index < len(imports); index += 2 {
		if strings.Count(source, imports[index]) != 1 {
			h.t.Fatalf("nonunique %s import", name)
		}
		source = strings.Replace(source, imports[index], imports[index+1], 1)
	}
	source = strings.ReplaceAll(source, "../../typescript", filepath.Join(h.repository, "stage1/typescript"))
	for _, file := range []string{"facts.ts", "diagnostic.ts", "unary_minus.ts", "rules.ts", "flags.ts", "frames.ts", "types.ts", "type_fact.ts", "parameters.ts"} {
		source = strings.ReplaceAll(source, "./"+file, filepath.Join(h.repository, "stage1/cohere/typeaware", file))
	}
	mutant := h.write(name+".ts", source)
	driver, err := os.ReadFile(entry)
	if err != nil {
		h.t.Fatal(err)
	}
	main := strings.Replace(string(driver), "./"+relative, mutant, 1)
	main = strings.ReplaceAll(main, "../../typescript", filepath.Join(h.repository, "stage1/typescript"))
	for _, file := range []string{"facts.ts", "diagnostic.ts", "unary_minus.ts", "rules.ts", "flags.ts", "frames.ts", "types.ts", "type_fact.ts", "parameters.ts"} {
		main = strings.ReplaceAll(main, "./"+file, filepath.Join(h.repository, "stage1/cohere/typeaware", file))
	}
	return h.build(stage0, name, h.write(name+"-main.ts", main), normal, false)
}

// Not parallel: archive builds, sanitizer runs and benchmarks share the machine.
func TestSixRuleAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_SIX_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	normal := h.archive("checker", "", false)
	sanitized := h.archive("checker-asan", "", true)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/suite.ts")
	binary := h.build(stage0, "suite-asan", entry, sanitized, true)
	optimized := h.build(stage0, "suite", entry, normal, false)
	oracle := filepath.Join(directory, "oracle")
	virtual := filepath.Join(repository, "cohere/adamic_six_oracle.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(repository, "stage1/cohere/typeaware/testdata/oracle_six.go")}})
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-overlay", h.write("oracle-overlay.json", string(data)), "-o", oracle, virtual)
	cmd.Dir = filepath.Join(repository, "cohere")
	h.must("oracle-build", cmd)
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var paths []string
	for i, source := range suiteSources(t, repository) {
		paths = append(paths, h.write(fmt.Sprintf("suite-case-%03d.ts", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("empty.ts", ""))
	manifest := h.write("generated.manifest", strings.Join(paths, "\n")+"\n")
	// Table extraction can also pick rendered messages and deliberately malformed
	// fixtures. The independent Go parser names the syntactically valid sources.
	valid := h.must("source-filter", exec.Command(oracle, config, manifest, "--valid-sources"))
	before := len(paths)
	paths = strings.Split(strings.TrimSpace(string(valid.stdout)), "\n")
	manifest = h.write("generated.manifest", strings.Join(paths, "\n")+"\n")
	t.Logf("fixture candidates: %d syntactically valid sources; %d nonsource strings or malformed fixtures excluded", len(paths), before-len(paths))
	truth := h.compare("six-generated", oracle, binary, config, manifest)
	for _, name := range []string{"no-unsafe-unary-minus", "related-getter-setter-pairs", "no-unsafe-declaration-merging", "no-unsafe-argument", "restrict-plus-operands", "no-unnecessary-boolean-literal-compare"} {
		marker := []byte("\t@typescript-eslint/" + name + "\t")
		count := bytes.Count(truth.stdout, marker)
		if count == 0 {
			t.Fatalf("no reporting control for %s", name)
		}
		t.Logf("%s: %d generated findings", name, count)
	}
	t.Logf("six-rule generated coverage: %d files", len(paths))
	nonstrict := h.write("nonstrict.json", `{"compilerOptions":{"strict":false,"target":"ES2022","module":"NodeNext","lib":["ES2022"]}}`)
	one := h.write("one.manifest", paths[0]+"\n")
	strictTruth := h.compare("six-nonstrict", oracle, binary, nonstrict, one)
	if !bytes.Contains(strictTruth.stdout, []byte("\tnoStrictNullCheck\t")) {
		t.Fatal("strict-null configuration control missing")
	}

	globalConfig := h.write("global.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","moduleDetection":"legacy","lib":["ES2022"]}}`)
	globalManifest := h.write("global.manifest", h.write("global-interface.ts", "/* first file */ interface Shared {}\n")+"\n"+h.write("global-class.ts", "class Shared {}\n")+"\n")
	globalTruth := h.compare("six-cross-file", oracle, binary, globalConfig, globalManifest)
	if !bytes.Contains(globalTruth.stdout, []byte("\tunsafeMerging\t")) {
		t.Fatal("cross-file declaration control missing")
	}
	globalOverlay := h.overlay("declaration-source", "bridge/tsgo/checker/facts.go", "scanner.GetTokenPosOfNode(name, source, false)", "scanner.GetTokenPosOfNode(name, f, false)")
	globalArchive := h.archive("declaration-source-checker", globalOverlay, false)
	globalMutant := h.build(stage0, "declaration-source-suite", entry, globalArchive, false)
	globalGot := h.must("declaration-source-run", exec.Command(globalMutant, globalConfig, globalManifest))
	if len(globalGot.stderr) != 0 || bytes.Equal(globalGot.stdout, globalTruth.stdout) {
		t.Fatal("cross-file declaration source mutant survived oracle")
	}
	t.Logf("declaration-source: production cohere oracle catches byte %d", firstDifference(globalGot.stdout, globalTruth.stdout))

	refusal := h.run("inspect-unlinked", exec.Command(stage0, "build", h.write("inspect-unlinked.ts", "import {tsgoInspect} from 'adamic'; console.log(tsgoInspect(1,'x',0,1,'Identifier','type'));\n"), "-o", filepath.Join(directory, "unlinked")))
	if refusal.err == nil || !bytes.Contains(refusal.stderr, []byte("unlinked typescript-go library call")) {
		t.Fatal("inspect accepted without bridge")
	}
	for _, change := range []struct{ name, from, to string }{
		{"wrong-node", "this.ask(left, 'base-type')", "this.ask(index, 'base-type')"},
		{"last-declaration", "if(use && pos < earliestPos)", "if(use)"},
		{"nullable-default", "if(!plain || nullable)", "if(!plain && !nullable)"},
	} {
		mutant := suiteTSMutant(h, stage0, normal, entry, change.name, "rules.ts", change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s not caught by oracle alone", change.name)
		}
		t.Logf("%s: production cohere oracle catches byte %d; %s", change.name, firstDifference(got.stdout, truth.stdout), summary(got.stdout))
	}
	// Union test mutates the facts consumer and reroutes the Rules import too.
	// A separate driver avoids a mutant whose unused module compiles but never runs.
	factsText, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/types.ts"))
	if err != nil {
		t.Fatal(err)
	}
	from := "return (type.flags & mask) !== 0 ? type.parts : [type.id];"
	if strings.Count(string(factsText), from) != 1 {
		t.Fatal("union anchor changed")
	}
	factsSource := string(factsText)
	for _, file := range []string{"flags.ts", "type_fact.ts"} {
		factsSource = strings.ReplaceAll(factsSource, "./"+file, filepath.Join(repository, "stage1/cohere/typeaware", file))
	}
	factsPath := h.write("union-mutant.ts", strings.Replace(factsSource, from, "return mask === 0 ? type.parts : [type.id];", 1))
	decoder, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/facts.ts"))
	if err != nil {
		t.Fatal(err)
	}
	decoderText := strings.Replace(string(decoder), "from './types.ts'", "from '"+factsPath+"'", 1)
	for _, file := range []string{"frames.ts", "type_fact.ts"} {
		decoderText = strings.ReplaceAll(decoderText, "./"+file, filepath.Join(repository, "stage1/cohere/typeaware", file))
	}
	decoderPath := h.write("union-decoder.ts", decoderText)
	parameters, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/parameters.ts"))
	if err != nil {
		t.Fatal(err)
	}
	parametersPath := h.write("union-parameters.ts", strings.Replace(string(parameters), "from './types.ts'", "from '"+factsPath+"'", 1))
	unionBinary := suiteTSMutant(h, stage0, normal, entry, "union-members", "rules.ts", "from './facts.ts'", "from '"+decoderPath+"'", "from './types.ts'", "from '"+factsPath+"'", "from './parameters.ts'", "from '"+parametersPath+"'")
	unionGot := h.must("union-members-run", exec.Command(unionBinary, config, manifest))
	if len(unionGot.stderr) != 0 || bytes.Equal(unionGot.stdout, truth.stdout) {
		t.Fatal("union mutant not caught by oracle alone")
	}
	t.Logf("union-members: production cohere oracle catches byte %d", firstDifference(unionGot.stdout, truth.stdout))
	for _, change := range []struct{ name, from, to string }{
		{"type-name", "out.text(c.TypeToString(p.typesByID[id-1]))", "out.text(c.TypeToString(p.typesByID[0]))"},
		{"raw-shape-constraint", `if mode != "raw-type" && mode != "raw-shape"`, `if mode != "raw-type"`},
		{"assignability-direction", "checker.Checker_isTypeAssignableTo(c, c.GetTypeAtLocation(node), c.GetTypeAtLocation(target))", "checker.Checker_isTypeAssignableTo(c, c.GetTypeAtLocation(target), c.GetTypeAtLocation(node))"},
		{"resolved-signature", "signature := c.GetResolvedSignature(node)", `var firstCall *ast.Node; var findCall func(*ast.Node) bool; findCall = func(n *ast.Node) bool { if n.Kind == ast.KindCallExpression { firstCall = n; return true }; return n.ForEachChild(findCall) }; findCall(source.AsNode()); selected := node; if node.Kind == ast.KindCallExpression && firstCall != nil { selected = firstCall }; signature := c.GetResolvedSignature(selected)`},
	} {
		overlay := h.overlay(change.name, "bridge/tsgo/checker/facts.go", change.from, change.to)
		archive := h.archive(change.name+"-checker", overlay, false)
		mutant := h.build(stage0, change.name+"-suite", entry, archive, false)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s not caught by oracle alone", change.name)
		}
		t.Logf("%s: production cohere oracle catches byte %d; %s", change.name, firstDifference(got.stdout, truth.stdout), summary(got.stdout))
	}
	probe := h.write("probe.ts", "-1;\n")
	releasedEntry := h.write("released-inspect.ts", "import {panic,programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const a=programArguments();const p=tsgoProgram(a[0]??panic('config'),[a[1]??panic('file')]);tsgoRelease(p);console.log(tsgoInspect(p,a[1]??panic('file'),0,2,'PrefixUnaryExpression','type'));\n")
	released := h.build(stage0, "released-inspect", releasedEntry, normal, false)
	got := h.run("released-inspect-run", exec.Command(released, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte("invalid or released checker handle")) {
		t.Fatalf("released inspect escaped: %v %s", got.err, got.stderr)
	}
	staleOverlay := h.overlay("retained-handle", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released roots.")
	staleArchive := h.archive("retained-handle-checker", staleOverlay, false)
	stale := h.build(stage0, "retained-handle", releasedEntry, staleArchive, false)
	h.must("retained-handle-run", exec.Command(stale, config, probe))
	t.Log("released program: panic 70; retained-handle mutant exits 0 and fails the stale-query expectation")
	// Explicit C allocation length, not the native path's extra terminator.
	lengthOverlay := h.overlay("facts-length", "bridge/tsgo/archive/main.go", "*facts = buffer(answer)", "*facts = buffer(answer); facts.length++")
	lengthArchive := h.archive("facts-length-checker", lengthOverlay, true)
	length := h.build(stage0, "facts-length", entry, lengthArchive, true)
	got = h.run("facts-length-run", exec.Command(length, config, manifest))
	if got.err == nil || !bytes.Contains(got.stderr, []byte("AddressSanitizer: heap-buffer-overflow")) {
		t.Fatalf("facts length mutant escaped ASan: %v %s", got.err, got.stderr)
	}
	t.Log("facts UTF-8 length + 1: AddressSanitizer heap-buffer-overflow")
	for _, change := range []struct{ name, value, message string }{
		{"facts-empty", "", "invalid checker facts frame"},
		{"facts-integer", "1\nx", "invalid checker facts integer"},
		{"facts-length", "-1\n", "invalid checker facts length"},
		{"facts-version", "1\n2", "unsupported checker facts schema"},
	} {
		overlay := h.overlay(change.name+"-bad", "bridge/tsgo/archive/main.go", "*facts = buffer(answer)", "_ = answer; *facts = buffer("+strconv.Quote(change.value)+")")
		archive := h.archive(change.name+"-bad-checker", overlay, false)
		malformed := h.build(stage0, change.name+"-bad", entry, archive, false)
		r := h.run(change.name+"-bad-run", exec.Command(malformed, config, manifest))
		if code, ok := r.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(r.stderr, []byte(change.message)) {
			t.Fatalf("%s malformed facts escaped: %v %s", change.name, r.err, r.stderr)
		}
		t.Logf("%s: panic 70, %s", change.name, change.message)
	}
	nativeCost := h.build(stage0, "facts-cost", filepath.Join(repository, "stage1/cohere/typeaware/testdata/fact_cost.ts"), normal, false)
	directCost := filepath.Join(directory, "direct-cost")
	h.must("facts-cost-go-build", exec.Command("go", "build", "-o", directCost, "./bridge/tsgo/cost"))
	costSource := "-1;\ninterface Pair { get value(): number; set value(v: string); }\ninterface Merge {}\nclass Merge {}\ndeclare function accept(x:number):void;\ndeclare const x:any;\naccept(x);\ndeclare const nullable:boolean|undefined;\nnullable === true;\ndeclare const nr:number|bigint;\nnr + nr;\n"
	costProbe := h.write("facts-probe.ts", costSource)
	getter := strings.Index(costSource, "get value") - 1
	getterEnd := getter + strings.Index(costSource[getter:], ";") + 1
	setter := strings.Index(costSource, "v: string")
	declaration := strings.Index(costSource, "class Merge") - 1
	declarationEnd := declaration + strings.Index(costSource[declaration:], "{}") + 2
	call := strings.Index(costSource, "\naccept(x)")
	callEnd := call + 1 + len("accept(x)")
	arg := call + 1 + len("accept(")
	nullable := strings.Index(costSource, "\nnullable ===")
	base := strings.Index(costSource, "\nnr + nr;")
	cases := []struct {
		name, kind, question string
		start, end           int
	}{
		{"assignable", "GetAccessor", fmt.Sprintf("assignable\n%d\n%d\nParameter", setter, setter+len("v: string")), getter, getterEnd},
		{"declarations", "ClassDeclaration", "declarations", declaration, declarationEnd},
		{"signature", "CallExpression", "signature", call, callEnd},
		{"raw-type", "Identifier", "raw-type", arg, arg + 1},
		{"nullable", "Identifier", "type", nullable, nullable + 1 + len("nullable")},
		{"union", "Identifier", "base-type", base, base + 1 + len("nr")},
		{"signature-shape", "CallExpression", "signature-shape", call, callEnd},
		{"raw-shape", "Identifier", "raw-shape", arg, arg + 1},
		{"type-shape", "Identifier", "type-shape", nullable, nullable + 1 + len("nullable")},
		{"options", "SourceFile", "options", 0, len(costSource)},
	}
	for _, probe := range cases {
		for round := 1; round <= 3; round++ {
			if round > 1 && os.Getenv("ADAMIC_TYPEAWARE_BENCH") != "1" {
				break
			}
			args := []string{config, costProbe, strconv.Itoa(probe.start), strconv.Itoa(probe.end), probe.kind, probe.question, "10000"}
			cmd := exec.Command(nativeCost, args...)
			cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			native := h.must(fmt.Sprintf("facts-cost-%s-%d-native", probe.name, round), cmd)
			direct := h.must(fmt.Sprintf("facts-cost-%s-%d-go", probe.name, round), exec.Command(directCost, args...))
			if !bytes.Equal(native.stdout, direct.stdout) {
				t.Fatalf("%s fact cost checksums differ", probe.name)
			}
			t.Logf("fact cost %s round %d: native %s; Go %s; identical units=%s", probe.name, round, strings.TrimSpace(string(native.stderr)), strings.TrimSpace(string(direct.stderr)), strings.TrimSpace(string(native.stdout)))
		}
	}
	corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if corpus == "" {
		t.Log("compiler corpus skipped: set ADAMIC_TYPESCRIPT_SOURCE")
		return
	}
	pin := h.must("corpus-pin", exec.Command("git", "-C", corpus, "rev-parse", "HEAD"))
	if strings.TrimSpace(string(pin.stdout)) != compilerCommit {
		t.Fatal("compiler corpus pin differs")
	}
	paths = nil
	err = filepath.WalkDir(filepath.Join(corpus, "src/compiler"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	if len(paths) != 77 {
		t.Fatalf("want all 77 files, got %d", len(paths))
	}
	compilerManifest := h.write("compiler.manifest", strings.Join(paths, "\n")+"\n")
	compilerConfig := filepath.Join(corpus, "src/compiler/tsconfig.json")
	h.compare("six-compiler", oracle, binary, compilerConfig, compilerManifest)
	if os.Getenv("ADAMIC_TYPEAWARE_BENCH") == "1" {
		for round := 1; round <= 3; round++ {
			binaries := []string{optimized, oracle}
			if round%2 == 0 {
				binaries[0], binaries[1] = binaries[1], binaries[0]
			}
			var expected []byte
			for _, binary := range binaries {
				cmd := exec.Command(binary, compilerConfig, compilerManifest, "--count")
				cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
				r := h.must(fmt.Sprintf("six-bench-%d-%s", round, filepath.Base(binary)), cmd)
				if expected != nil && !bytes.Equal(r.stdout, expected) {
					t.Fatal("timed counts differ")
				}
				expected = r.stdout
				t.Logf("six-rule all-77 round %d %s: process_s=%.6f %s; %s", round, filepath.Base(binary), r.elapsed.Seconds(), summary(r.stdout), strings.TrimSpace(string(r.stderr)))
			}
		}
	}
}

func TestSixPinnedFlags(t *testing.T) {
	t.Parallel()
	for _, pair := range []struct {
		name     string
		actual   checker.TypeFlags
		expected uint32
	}{
		{"union", checker.TypeFlagsUnion, 134217728}, {"intersection", checker.TypeFlagsIntersection, 268435456}, {"boolean", checker.TypeFlagsBooleanLike, 8448}, {"nullable", checker.TypeFlagsNullable, 12}, {"number", checker.TypeFlagsNumberLike, 67648}, {"bigint", checker.TypeFlagsBigIntLike, 4224}, {"string", checker.TypeFlagsStringLike, 12583968},
	} {
		if uint32(pair.actual) != pair.expected {
			t.Fatalf("update %s mask: %d", pair.name, pair.actual)
		}
	}
}

// Not parallel: this builds and executes a native decoder with dynamic inputs.
func TestFactsDecoderGuards(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_SIX_GUARD_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	binary := filepath.Join(directory, "decoder")
	h.must("decoder-build", exec.Command(stage0, "build", filepath.Join(repository, "stage1/cohere/typeaware/testdata/facts_decode.ts"), "-o", binary, "--sanitize"))
	// Complete schema, one number type. Each record owns no other record.
	fields := []string{"1", "raw-type", "1", "1", "1", "1", "0", "1", "1", "64", "number", "0", "0", "0", "-1", "0", "0", "0", "0"}
	frame := func(values []string) string {
		var out strings.Builder
		for _, value := range values {
			fmt.Fprintf(&out, "%d\n%s", len([]rune(value)), value)
		}
		return out.String()
	}
	good := h.must("decoder-valid", exec.Command(binary, frame(fields)))
	if string(good.stdout) != "64\n" || len(good.stderr) != 0 {
		t.Fatal("valid decoder control failed")
	}
	cases := []struct{ name, message, payload string }{{"empty", "invalid checker facts frame", ""}, {"length-negative", "invalid checker facts length", "-1\n"}, {"length-short", "invalid checker facts length", "99\nx"}, {"trailing", "trailing checker facts", frame(fields) + "x"}}
	for _, change := range []struct {
		name, message, value string
		index                int
	}{
		{"version", "unsupported checker facts schema", "2", 0},
		{"question", "unsupported checker facts schema", "type", 1},
		{"not-integer", "invalid checker facts integer", "x", 9},
		{"rounded-integer", "invalid checker facts integer", "9007199254740993", 9},
		{"noncanonical-integer", "invalid checker facts integer", "064", 9},
		{"negative-natural", "negative checker fact", "-1", 9},
		{"bad-boolean", "invalid checker facts boolean", "2", 2},
		{"zero-id", "invalid checker type record", "0", 8},
		{"negative-tuple", "invalid checker type record", "-2", 14},
		{"missing-root", "missing checker type identity", "2", 5},
		{"missing-constraint", "missing checker type identity", "2", 15},
		{"missing-element", "missing checker type identity", "2", 16},
	} {
		values := append([]string{}, fields...)
		values[change.index] = change.value
		cases = append(cases, struct{ name, message, payload string }{change.name, change.message, frame(values)})
	}
	duplicate := append([]string{}, fields...)
	duplicate[7] = "2"
	duplicate = append(duplicate, fields[8:]...)
	cases = append(cases, struct{ name, message, payload string }{"duplicate-id", "invalid checker type record", frame(duplicate)})
	for _, offset := range []int{17, 18} {
		values := append([]string{}, fields[:offset]...)
		values = append(values, "1", "2")
		values = append(values, fields[offset+1:]...)
		cases = append(cases, struct{ name, message, payload string }{fmt.Sprintf("missing-link-%d", offset), "missing checker type identity", frame(values)})
	}
	for _, mutant := range cases {
		got := h.run(mutant.name, exec.Command(binary, mutant.payload))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte(mutant.message)) {
			t.Fatalf("%s escaped: %v %s", mutant.name, got.err, got.stderr)
		}
		t.Logf("wire mutant %s: panic 70, %s", mutant.name, mutant.message)
	}
}

// Not parallel: each mutant is a separately linked archive and native program.
func TestInspectRequestRefusals(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_SIX_REQUEST_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	normal := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/fact_cost.ts")
	binary := h.build(stage0, "request", entry, normal, false)
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	probe := h.write("probe.ts", "-1;\n")
	h.must("request-valid", exec.Command(binary, config, probe, "0", "2", "PrefixUnaryExpression", "raw-type", "2"))
	releasedEntry := h.write("released-name.ts", "import {panic,programArguments,tsgoProgram,tsgoInspect,tsgoRelease} from 'adamic'; const a=programArguments(); const config=a[0]??panic('config');const file=a[1]??panic('file');const p=tsgoProgram(config,[file]); console.log(tsgoInspect(p,file,0,2,'PrefixUnaryExpression','raw-shape')); tsgoRelease(p); console.log(tsgoInspect(p,file,0,2,'PrefixUnaryExpression','name\\n1'));\n")
	released := h.build(stage0, "released-name", releasedEntry, normal, false)
	stale := h.run("released-name-run", exec.Command(released, config, probe))
	if code, ok := stale.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(stale.stderr, []byte("invalid or released checker handle")) {
		t.Fatalf("released name escaped: %v %s", stale.err, stale.stderr)
	}
	t.Log("released type-name query: panic 70, invalid or released checker handle")
	for _, change := range []struct{ name, kind, question, message, from, to string }{
		{"wrong-kind", "Identifier", "raw-type", "no exact Identifier node", `candidate.Kind.String() == "Kind"+kind`, `kind != ""`},
		{"unknown-question", "PrefixUnaryExpression", "unknown", "unsupported checker question", `return "", fmt.Errorf("unsupported checker question: %s", question)`, `return out.String(), nil`},
	} {
		args := []string{config, probe, "0", "2", change.kind, change.question, "2"}
		got := h.run(change.name+"-refused", exec.Command(binary, args...))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte(change.message)) {
			t.Fatalf("%s escaped: %v %s", change.name, got.err, got.stderr)
		}
		overlay := h.overlay(change.name, "bridge/tsgo/checker/facts.go", change.from, change.to)
		archive := h.archive(change.name+"-checker", overlay, false)
		mutant := h.build(stage0, change.name, entry, archive, false)
		h.must(change.name+"-mutant-run", exec.Command(mutant, args...))
		t.Logf("%s: panic 70; guard mutant exits 0 and fails the refusal expectation", change.name)
	}
}
