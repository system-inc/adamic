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
	"sync"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func suiteSources(test *testing.T, repository string) []string {
	test.Helper()
	names := []string{
		"no_unsafe_unary_minus",
		"related_getter_setter_pairs",
		"no_unsafe_declaration_merging",
		"no_unsafe_argument",
		"restrict_plus_operands",
		"no_unnecessary_boolean_literal_compare",
	}
	pinnedFixtureFiles(test, repository, names)
	unique := map[string]bool{}
	for _, name := range names {
		tree, parseFileError := parser.ParseFile(
			token.NewFileSet(),
			filepath.Join(repository, "cohere/internal/lint/rules/typescript", name+"_test.go"),
			nil,
			0,
		)
		if parseFileError != nil {
			test.Fatal(parseFileError)
		}
		add := func(expr goast.Expr) {
			literal, ok := expr.(*goast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return
			}
			text, unquoteError := strconv.Unquote(literal.Value)
			if unquoteError == nil && (strings.Contains(text, ";") ||
				strings.Contains(text, "\n") ||
				strings.HasPrefix(text, "class ") ||
				strings.HasPrefix(text, "interface ")) {
				unique[text] = true
			}
		}
		goast.Inspect(tree, func(node goast.Node) bool {
			literal, ok := node.(*goast.CompositeLit)
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
					if key, ok := field.Key.(*goast.Ident); ok &&
						(key.Name == "sourceText" || key.Name == "source" || key.Name == "code") {
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

func suiteTsMutant(harness *harness, normal, entry, name, relative, from, to string, imports ...string) string {
	harness.t.Helper()
	path := filepath.Join(harness.repository, "stage1/cohere/typeaware", relative)
	text, readError := os.ReadFile(path)
	if readError != nil {
		harness.t.Fatal(readError)
	}
	if strings.Count(string(text), from) != 1 {
		harness.t.Fatalf("nonunique %s anchor", name)
	}
	source := strings.Replace(string(text), from, to, 1)
	// A cloned nominal class must be shared by its decoder and every typed caller.
	for index := 0; index < len(imports); index += 2 {
		if strings.Count(source, imports[index]) != 1 {
			harness.t.Fatalf("nonunique %s import", name)
		}
		source = strings.Replace(source, imports[index], imports[index+1], 1)
	}
	source = strings.ReplaceAll(source, "../../typescript", filepath.Join(harness.repository, "stage1/typescript"))
	for _, file := range []string{
		"facts.ts",
		"diagnostic.ts",
		"unary_minus.ts",
		"rules.ts",
		"flags.ts",
		"frames.ts",
		"types.ts",
		"type_fact.ts",
		"parameters.ts",
	} {
		source = strings.ReplaceAll(
			source,
			"./"+file,
			filepath.Join(harness.repository, "stage1/cohere/typeaware", file),
		)
	}
	mutant := harness.write(name+".ts", source)
	driver, readError := os.ReadFile(entry)
	if readError != nil {
		harness.t.Fatal(readError)
	}
	main := strings.Replace(string(driver), "./"+relative, mutant, 1)
	main = strings.ReplaceAll(main, "../../typescript", filepath.Join(harness.repository, "stage1/typescript"))
	for _, file := range []string{
		"facts.ts",
		"diagnostic.ts",
		"unary_minus.ts",
		"rules.ts",
		"flags.ts",
		"frames.ts",
		"types.ts",
		"type_fact.ts",
		"parameters.ts",
	} {
		main = strings.ReplaceAll(main, "./"+file, filepath.Join(harness.repository, "stage1/cohere/typeaware", file))
	}
	return harness.mutantNativeOf(
		name,
		harness.loweredMutant(harness.write(name+"-main.ts", main), name, "C six mutant "+name),
		normal,
		false,
	)
}

// Builds are prepared once and shared before the parallel check units start.
// Each six-build log is an input-build unit for the developer-tools cache helper.
// ADAMIC_TEST_SHARD=i/n (zero based) selects unit indices modulo n; unset runs
// every unit. Full program roots and each mutant's original corpus are retained.
const testSixRuleAgreementAndMutantsShards = 95

func prepareSixRuleAgreementAndMutants(test *testing.T) *typeAwarePlan {
	started := time.Now()
	if _, _, sixSelectionError := sixSelection(os.Getenv("ADAMIC_TEST_SHARD")); sixSelectionError != nil {
		test.Fatal(sixSelectionError)
	}
	var shards []sixShard
	add := func(name string, ids []string, run func(*harness)) {
		shards = append(shards, sixShard{name: name, ids: ids, run: run})
	}
	repository, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	directory := filepath.Join(productDirectory, "six-plan")
	if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if path := os.Getenv("ADAMIC_SIX_ARTIFACTS"); path != "" {
		directory, absolutePathError = filepath.Abs(path)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
	}
	h := &harness{
		t:            test,
		repository:   repository,
		directory:    directory,
		sixBuilds:    true,
		sharedSetup:  true,
		setupStarted: started,
	}
	traceGroup(test)
	stage0 := "" // Native builds use the in-process compiler; only refusal shards need stage zero.
	normal := func(harness *harness) string { return typeAwareArchive(harness, "checker", false) }
	sanitized := func(harness *harness) string { return typeAwareArchive(harness, "checker-asan", true) }
	entry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/sharded_suite.ts")
	binary := func(harness *harness) string {
		return typeAwareBuild(harness, stage0, "suite-asan", entry, sanitized(harness), true)
	}
	optimized := func(harness *harness) string {
		return typeAwareBuild(harness, stage0, "suite", entry, normal(harness), false)
	}
	// Common products are ready before any case deadline starts.
	var readyNormal, readySanitized string
	var builds sync.WaitGroup
	builds.Add(3)
	go func() { defer builds.Done(); typeAwareStage0(h) }()
	go func() { defer builds.Done(); readyNormal = normal(h) }()
	go func() { defer builds.Done(); readySanitized = sanitized(h) }()
	builds.Wait()
	normal = func(*harness) string { return readyNormal }
	sanitized = func(*harness) string { return readySanitized }
	var readyBinary, readyOptimized string
	builds.Add(2)
	go func() { defer builds.Done(); readyBinary = binary(h) }()
	go func() { defer builds.Done(); readyOptimized = optimized(h) }()
	builds.Wait()
	binary = func(*harness) string { return readyBinary }
	optimized = func(*harness) string { return readyOptimized }
	// The corpus, its oracle and truth, and the 13 mutants that run over it are prepared apart from the rest of this
	// plan, so a witness product builds without the plan's natives (sixCorpusOnce).
	generated := sixCorpusOnce(test)
	oracle, configuration, paths, manifest, truth := generated.oracle, generated.config, generated.paths, generated.manifest, generated.truth
	// Generated fixtures are fixed by cohere 7945d102a6c18dd36adf9114a758ce646e8b2359
	// and the explicit controls in suiteSources; contiguous ranges are stable.
	for index, group := range sixRanges(paths, 32) {
		add(
			fmt.Sprintf("agreement/generated-%02d", index),
			sixIds("agreement/generated", group),
			func(harness *harness) {
				harness.sixCompare("generated", oracle, binary(harness), configuration, manifest, group)
			},
		)
	}
	add("controls/reporting", sixRuleControlIds(), func(harness *harness) {
		test := harness.t
		for _, name := range []string{
			"no-unsafe-unary-minus",
			"related-getter-setter-pairs",
			"no-unsafe-declaration-merging",
			"no-unsafe-argument",
			"restrict-plus-operands",
			"no-unnecessary-boolean-literal-compare",
		} {
			marker := []byte("\t@typescript-eslint/" + name + "\t")
			count := bytes.Count(truth.stdout, marker)
			if count == 0 {
				test.Fatalf("no reporting control for %s", name)
			}
			test.Logf("%s: %d generated findings", name, count)
		}
		test.Logf("six-rule generated coverage: %d files", len(paths))
	})
	nonstrict := h.write(
		"nonstrict.json",
		`{"compilerOptions":{"strict":false,"target":"ES2022","module":"NodeNext","lib":["ES2022"]}}`,
	)
	one := h.write("one.manifest", paths[0]+"\n")
	add("agreement/nonstrict", sixIds("agreement/nonstrict", paths[:1]), func(harness *harness) {
		test := harness.t
		strictTruth := harness.compare("six-nonstrict", oracle, binary(harness), nonstrict, one)
		if !bytes.Contains(strictTruth.stdout, []byte("\tnoStrictNullCheck\t")) {
			test.Fatal("strict-null configuration control missing")
		}
	})

	globalConfiguration := h.write(
		"global.json",
		`{"compilerOptions":{"strict":true,"target":"ES2022","module":"ESNext","moduleDetection":"legacy","lib":["ES2022"]}}`,
	)
	globalManifest := h.write("global.manifest", h.write(
		"global-interface.ts",
		"/* first file */ interface Shared {}\n",
	)+"\n"+h.write("global-class.ts", "class Shared {}\n")+"\n")
	globalPaths := strings.Split(strings.TrimSpace(globalManifestText(h, globalManifest)), "\n")
	globalTruth := h.must("global-truth", exec.Command(oracle, globalConfiguration, globalManifest))
	add("agreement/cross-file", sixIds("agreement/cross-file", globalPaths), func(harness *harness) {
		test := harness.t
		actual := harness.sixCompare(
			"cross-file",
			oracle,
			binary(harness),
			globalConfiguration,
			globalManifest,
			globalPaths,
		)
		if !bytes.Contains(actual.stdout, []byte("\tunsafeMerging\t")) {
			test.Fatal("cross-file declaration control missing")
		}
	})

	add("mutant/declaration-source", sixIds("mutant/declaration-source", globalPaths), func(harness *harness) {
		test := harness.t
		globalArchive := typeAwareMutantArchive(harness, "declaration-source")
		globalMutant := typeAwareMutantNative(harness, "declaration-source-suite", entry, globalArchive, false)
		globalGot := harness.must(
			"declaration-source-run",
			exec.Command(globalMutant, globalConfiguration, globalManifest),
		)
		if len(globalGot.stderr) != 0 || bytes.Equal(globalGot.stdout, globalTruth.stdout) {
			test.Fatal("cross-file declaration source mutant survived oracle")
		}
		test.Logf(
			"declaration-source: production cohere oracle catches byte %d",
			firstDifference(globalGot.stdout, globalTruth.stdout),
		)
	})

	add("refusal/unlinked", []string{"refusal/unlinked"}, func(harness *harness) {
		stage0 := typeAwareStage0(harness)
		test := harness.t
		refusal := harness.run("inspect-unlinked", exec.Command(stage0, "build", harness.write(
			"inspect-unlinked.ts",
			"import {tsgoInspect} from 'adamic'; console.log(tsgoInspect(1,'x',0,1,'Identifier','type'));\n",
		), "-o", filepath.Join(directory, "unlinked")))
		if refusal.err == nil || !bytes.Contains(refusal.stderr, []byte("unlinked typescript-go library call")) {
			test.Fatal("inspect accepted without bridge")
		}
	})

	for _, name := range []string{"wrong-node", "last-declaration", "nullable-default"} {
		add("mutant/"+name, sixIds("mutant/"+name, paths), func(harness *harness) {
			mutant, cases := generated.witnessRun(harness, name)
			generated.caught(harness, name, mutant, cases)
		})
	}
	add("mutant/union-members", sixIds("mutant/union-members", paths), func(harness *harness) {
		mutant, cases := generated.witnessRun(harness, "union-members")
		generated.caught(harness, "union-members", mutant, cases)
	})

	for _, name := range []string{
		"type-name",
		"raw-shape-constraint",
		"assignability-direction",
		"resolved-signature",
	} {
		add("mutant/"+name, sixIds("mutant/"+name, paths), func(harness *harness) {
			mutant, cases := generated.witnessRun(harness, name)
			generated.caught(harness, name, mutant, cases)
		})
	}
	probe := h.write("probe.ts", "-1;\n")
	releasedEntry := h.write(
		"released-inspect.ts",
		"import {panic,programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const a=programArguments();const p=tsgoProgram(a[0]??panic('config'),[a[1]??panic('file')]);tsgoRelease(p);console.log(tsgoInspect(p,a[1]??panic('file'),0,2,'PrefixUnaryExpression','type'));\n",
	)
	add("control/released-handle", []string{"control/released-handle"}, func(harness *harness) {
		test := harness.t
		released := typeAwareBuild(harness, stage0, "released-inspect", releasedEntry, normal(harness), false)
		got := harness.run("released-inspect-run", exec.Command(released, configuration, probe))
		if code, ok := got.err.(*exec.ExitError); !ok ||
			code.ExitCode() != 70 ||
			!bytes.Contains(got.stderr, []byte("invalid or released checker handle")) {
			test.Fatalf("released inspect escaped: %v %s", got.err, got.stderr)
		}
	})

	add("mutant/retained-handle", []string{"mutant/retained-handle"}, func(harness *harness) {
		test := harness.t
		staleArchive := typeAwareMutantArchive(harness, "retained-handle")
		stale := typeAwareBuild(harness, stage0, "retained-handle", releasedEntry, staleArchive, false)
		harness.must("retained-handle-run", exec.Command(stale, configuration, probe))
		test.Log("released program: panic 70; retained-handle mutant exits 0 and fails the stale-query expectation")
	})

	// Explicit C allocation length, not the native path's extra terminator.
	add("mutant/facts-length-asan", sixIds("mutant/facts-length-asan", paths), func(harness *harness) {
		test := harness.t
		length, cases := generated.witnessRun(harness, "facts-length-asan")
		got := harness.run("facts-length-run", exec.Command(length, configuration, cases))
		if !generated.mutants["facts-length-asan"].dies(got) {
			test.Fatalf("facts length mutant escaped ASan: %v %s", got.err, got.stderr)
		}
		test.Log("facts UTF-8 length + 1: AddressSanitizer heap-buffer-overflow")
	})

	// Each frame's value is planted by its checker mutant (checker_mutants_test.go).
	for _, change := range []struct{ name, message string }{
		{"facts-empty", "invalid checker facts frame"},
		{"facts-integer", "invalid checker facts integer"},
		{"facts-length", "invalid checker facts length"},
		{"facts-version", "unsupported checker facts schema"},
	} {
		add("mutant/"+change.name+"-bad", sixIds("mutant/"+change.name+"-bad", paths), func(harness *harness) {
			test := harness.t
			malformed, cases := generated.witnessRun(harness, change.name+"-bad")
			result := harness.run(change.name+"-bad-run", exec.Command(malformed, configuration, cases))
			if !generated.mutants[change.name+"-bad"].dies(result) {
				test.Fatalf("%s malformed facts escaped: %v %s", change.name, result.err, result.stderr)
			}
			test.Logf("%s: panic 70, %s", change.name, change.message)
		})
	}
	nativeCost := func(harness *harness) string {
		return typeAwareBuild(
			harness,
			stage0,
			"facts-cost",
			filepath.Join(repository, "stage1/cohere/typeaware/testdata/fact_cost.ts"),
			normal(harness),
			false,
		)
	}
	directCost := func(harness *harness) string {
		return sixFactsCostGo(harness)
	}
	readyNativeCost, readyDirectCost := nativeCost(h), directCost(h)
	nativeCost = func(*harness) string { return readyNativeCost }
	directCost = func(*harness) string { return readyDirectCost }
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
		{
			"assignable",
			"GetAccessor",
			fmt.Sprintf("assignable\n%d\n%d\nParameter", setter, setter+len("v: string")),
			getter,
			getterEnd,
		},
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
	rounds := 1
	if os.Getenv("ADAMIC_TYPEAWARE_BENCH") == "1" {
		rounds = 3
	}
	for _, probe := range cases {
		for round := 1; round <= 3; round++ {
			ids := []string{fmt.Sprintf("cost/%s-%d", probe.name, round)}
			if round > rounds {
				ids = nil
			}
			add(fmt.Sprintf("cost/%s-%d", probe.name, round), ids, func(harness *harness) {
				if round > rounds {
					harness.t.Skip("set ADAMIC_TYPEAWARE_BENCH=1")
				}
				test := harness.t
				args := []string{
					configuration,
					costProbe,
					strconv.Itoa(probe.start),
					strconv.Itoa(probe.end),
					probe.kind,
					probe.question,
					"10000",
				}
				command := exec.Command(nativeCost(harness), args...)
				command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
				native := harness.must(fmt.Sprintf("facts-cost-%s-%d-native", probe.name, round), command)
				direct := harness.must(
					fmt.Sprintf("facts-cost-%s-%d-go", probe.name, round),
					exec.Command(directCost(harness), args...),
				)
				if !bytes.Equal(native.stdout, direct.stdout) {
					test.Fatalf("%s fact cost checksums differ", probe.name)
				}
				test.Logf(
					"fact cost %s round %d: native %s; Go %s; identical units=%s",
					probe.name,
					round,
					strings.TrimSpace(string(native.stderr)),
					strings.TrimSpace(string(direct.stderr)),
					strings.TrimSpace(string(native.stdout)),
				)
			})
		}
	}
	compilerPaths := []string(nil)
	corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if corpus == "" {
		test.Log("compiler corpus skipped: set ADAMIC_TYPESCRIPT_SOURCE (#xq2ecw6)")
	} else {
		compilerPaths = pinnedCompilerFiles(test, corpus)
		sort.Strings(compilerPaths)
	}
	compilerManifest := h.write("compiler.manifest", strings.Join(compilerPaths, "\n")+"\n")
	compilerConfiguration := filepath.Join(corpus, "src/compiler/tsconfig.json")
	groups := make([][]string, 8)
	if len(compilerPaths) != 0 {
		// compilerCommit fixes these 77 roots; repository growth cannot move them.
		groups = sixCompilerRanges(test, compilerPaths, 8)
	}
	for index, group := range groups {
		add(fmt.Sprintf("agreement/compiler-%02d", index), sixIds("agreement/compiler", group), func(harness *harness) {
			if len(group) == 0 {
				harness.t.Skip("set ADAMIC_TYPESCRIPT_SOURCE")
			}
			harness.sixCompare("compiler", oracle, binary(harness), compilerConfiguration, compilerManifest, group)
		})
		for round := 1; round <= 3; round++ {
			ids := sixIds(fmt.Sprintf("bench/compiler-%d", round), group)
			if rounds != 3 {
				ids = nil
			}
			add(fmt.Sprintf("bench/compiler-%02d-%d", index, round), ids, func(harness *harness) {
				if len(group) == 0 || rounds != 3 {
					harness.t.Skip("set ADAMIC_TYPESCRIPT_SOURCE and ADAMIC_TYPEAWARE_BENCH=1")
				}
				test := harness.t
				selected := harness.write("selected.manifest", strings.Join(group, "\n")+"\n")
				binaries := []string{optimized(harness), oracle}
				if round%2 == 0 {
					binaries[0], binaries[1] = binaries[1], binaries[0]
				}
				var expected []byte
				for _, binary := range binaries {
					command := exec.Command(
						binary,
						compilerConfiguration,
						compilerManifest,
						"--files",
						selected,
						"--count",
					)
					command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
					result := harness.must(fmt.Sprintf("six-bench-%d-%s", round, filepath.Base(binary)), command)
					if expected != nil && !bytes.Equal(result.stdout, expected) {
						test.Fatal("timed counts differ")
					}
					expected = result.stdout
					test.Logf(
						"six-rule round %d %s: process_s=%.6f %s; %s",
						round,
						filepath.Base(binary),
						result.elapsed.Seconds(),
						summary(result.stdout),
						strings.TrimSpace(string(result.stderr)),
					)
				}
			})
		}
	}
	// Independent unsplit enumeration: membership is not derived from assignments.
	expected := sixUnsplitIds(paths, globalPaths, compilerPaths, rounds, os.Getenv("ADAMIC_TYPEAWARE_BENCH") == "1")
	return newTypeAwarePlan(test, h, expected, shards, testSixRuleAgreementAndMutantsShards)
}

func TestSixPinnedFlags(test *testing.T) {
	test.Parallel()
	for _, pair := range []struct {
		name     string
		actual   checker.TypeFlags
		expected uint32
	}{
		{"union", checker.TypeFlagsUnion, 134217728},
		{"intersection", checker.TypeFlagsIntersection, 268435456},
		{"boolean", checker.TypeFlagsBooleanLike, 8448},
		{"nullable", checker.TypeFlagsNullable, 12},
		{"number", checker.TypeFlagsNumberLike, 67648},
		{"bigint", checker.TypeFlagsBigIntLike, 4224},
		{"string", checker.TypeFlagsStringLike, 12583968},
	} {
		if uint32(pair.actual) != pair.expected {
			test.Fatalf("update %s mask: %d", pair.name, pair.actual)
		}
	}
}

// Not parallel: this builds and executes a native decoder with dynamic inputs.
func TestFactsDecoderGuards(test *testing.T) {
	repository, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	directory := test.TempDir()
	if path := os.Getenv("ADAMIC_SIX_GUARD_ARTIFACTS"); path != "" {
		directory, absolutePathError = filepath.Abs(path)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
	}
	harness := &harness{t: test, repository: repository, directory: directory}
	// The stage 0 compiler is the package's shared product, built ahead by TestProduct_stage0.
	stage0 := harness.stage0()
	binary := filepath.Join(directory, "decoder")
	harness.must("decoder-build", exec.Command(
		stage0,
		"build",
		filepath.Join(repository, "stage1/cohere/typeaware/testdata/facts_decode.ts"),
		"-o",
		binary,
		"--sanitize",
	))
	// Complete schema, one number type. Each record owns no other record.
	fields := []string{
		"1",
		"raw-type",
		"1",
		"1",
		"1",
		"1",
		"0",
		"1",
		"1",
		"64",
		"number",
		"0",
		"0",
		"0",
		"-1",
		"0",
		"0",
		"0",
		"0",
	}
	frame := func(values []string) string {
		var out strings.Builder
		for _, value := range values {
			fmt.Fprintf(&out, "%d\n%s", len([]rune(value)), value)
		}
		return out.String()
	}
	good := harness.must("decoder-valid", exec.Command(binary, frame(fields)))
	if string(good.stdout) != "64\n" || len(good.stderr) != 0 {
		test.Fatal("valid decoder control failed")
	}
	cases := []struct{ name, message, payload string }{
		{"empty", "invalid checker facts frame", ""},
		{"length-negative", "invalid checker facts length", "-1\n"},
		{"length-short", "invalid checker facts length", "99\nx"},
		{"trailing", "trailing checker facts", frame(fields) + "x"},
	}
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
	cases = append(
		cases,
		struct{ name, message, payload string }{"duplicate-id", "invalid checker type record", frame(duplicate)},
	)
	for _, offset := range []int{17, 18} {
		values := append([]string{}, fields[:offset]...)
		values = append(values, "1", "2")
		values = append(values, fields[offset+1:]...)
		cases = append(cases, struct{ name, message, payload string }{
			fmt.Sprintf("missing-link-%d", offset),
			"missing checker type identity",
			frame(values),
		})
	}
	for _, mutant := range cases {
		got := harness.run(mutant.name, exec.Command(binary, mutant.payload))
		if code, ok := got.err.(*exec.ExitError); !ok ||
			code.ExitCode() != 70 ||
			!bytes.Contains(got.stderr, []byte(mutant.message)) {
			test.Fatalf("%s escaped: %v %s", mutant.name, got.err, got.stderr)
		}
		test.Logf("wire mutant %s: panic 70, %s", mutant.name, mutant.message)
	}
}

// sixCorpus is what the generated agreement shards and every planted mutant's witnesses read: the six oracle, the
// syntactically valid generated cases and the oracle's truth over them, and the 13 mutants that run over them. It is
// prepared once per process apart from the rest of the six plan, so a witness product builds without the plan's
// natives.
//
// A planted mutant proves the oracle comparison can fail, and the cases its run differs on prove that as well as the
// whole corpus does: the 13 mutants that read the generated corpus were 73% of the six family's seconds run over all of
// it (#testaudit row 16, Loom's ruling (a), Oct 14). Each mutant shard runs its mutant over its witnesses, which
// witnesses finds once for each change to anything the run reads and keeps as a product, never a hand list.
type sixCorpus struct {
	paths                    []string
	oracle, config, manifest string
	truth                    result
	mutants                  map[string]sixMutant
}

// sixMutant is a planted mutant over the generated corpus: how to build it and, for one that must die, how it dies.
type sixMutant struct {
	build func(*harness) string
	dies  func(result) bool
}

var sixCorpusShared struct {
	once   sync.Once
	corpus *sixCorpus
}

func sixCorpusOnce(test *testing.T) *sixCorpus {
	test.Helper()
	shared := &sixCorpusShared
	shared.once.Do(func() {
		repository, absolutePathError := filepath.Abs("../../..")
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		directory := filepath.Join(productDirectory, "six-corpus")
		if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		harness := &harness{
			t:            test,
			repository:   repository,
			directory:    directory,
			sixBuilds:    true,
			sharedSetup:  true,
			setupStarted: time.Now(),
		}
		oracle := typeAwareOracle(harness, "six")
		configuration := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
		var paths []string
		for index, source := range suiteSources(test, repository) {
			paths = append(paths, harness.write(fmt.Sprintf("suite-case-%03d.ts", index), source+"\nexport {};\n"))
		}
		paths = append(paths, harness.write("empty.ts", ""))
		manifest := harness.write("generated.manifest", strings.Join(paths, "\n")+"\n")
		// Table extraction can also pick rendered messages and deliberately malformed
		// fixtures. The independent Go parser names the syntactically valid sources.
		valid := harness.must("source-filter", exec.Command(oracle, configuration, manifest, "--valid-sources"))
		before := len(paths)
		paths = strings.Split(strings.TrimSpace(string(valid.stdout)), "\n")
		manifest = harness.write("generated.manifest", strings.Join(paths, "\n")+"\n")
		test.Logf(
			"fixture candidates: %d syntactically valid sources; %d nonsource strings or malformed fixtures excluded",
			len(paths),
			before-len(paths),
		)
		truth := harness.must("generated-truth", exec.Command(oracle, configuration, manifest))
		shared.corpus = &sixCorpus{
			paths:    paths,
			oracle:   oracle,
			config:   configuration,
			manifest: manifest,
			truth:    truth,
			mutants:  sixMutants(harness, repository),
		}
	})
	if shared.corpus == nil {
		test.Fatal("six corpus preparation failed")
	}
	return shared.corpus
}

// sixMutants are the 13 planted mutants that run over the generated corpus, by name.
func sixMutants(h *harness, repository string) map[string]sixMutant {
	test := h.t
	normal := func(harness *harness) string { return typeAwareArchive(harness, "checker", false) }
	entry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/sharded_suite.ts")
	mutants := map[string]sixMutant{}
	for _, change := range []struct{ name, from, to string }{
		{"wrong-node", "this.ask(left, 'base-type')", "this.ask(index, 'base-type')"},
		{"last-declaration", "if(use && pos < earliestPos)", "if(use)"},
		{"nullable-default", "if(!plain || nullable)", "if(!plain && !nullable)"},
	} {
		mutants[change.name] = sixMutant{build: func(harness *harness) string {
			return typeAwareMutant(harness, normal(harness), entry, change.name, "rules.ts", change.from, change.to)
		}}
	}
	// Union test mutates the facts consumer and reroutes the Rules import too.
	// A separate driver avoids a mutant whose unused module compiles but never runs.
	factsText, readError := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/types.ts"))
	if readError != nil {
		test.Fatal(readError)
	}
	from := "return (type.flags & mask) !== 0 ? type.parts : [type.id];"
	if strings.Count(string(factsText), from) != 1 {
		test.Fatal("union anchor changed")
	}
	factsSource := string(factsText)
	for _, file := range []string{"flags.ts", "type_fact.ts"} {
		factsSource = strings.ReplaceAll(
			factsSource,
			"./"+file,
			filepath.Join(repository, "stage1/cohere/typeaware", file),
		)
	}
	factsPath := h.write(
		"union-mutant.ts",
		strings.Replace(factsSource, from, "return mask === 0 ? type.parts : [type.id];", 1),
	)
	decoder, readError := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/facts.ts"))
	if readError != nil {
		test.Fatal(readError)
	}
	decoderText := strings.Replace(string(decoder), "from './types.ts'", "from '"+factsPath+"'", 1)
	for _, file := range []string{"frames.ts", "type_fact.ts"} {
		decoderText = strings.ReplaceAll(
			decoderText,
			"./"+file,
			filepath.Join(repository, "stage1/cohere/typeaware", file),
		)
	}
	decoderPath := h.write("union-decoder.ts", decoderText)
	parameters, readError := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/parameters.ts"))
	if readError != nil {
		test.Fatal(readError)
	}
	parametersPath := h.write(
		"union-parameters.ts",
		strings.Replace(string(parameters), "from './types.ts'", "from '"+factsPath+"'", 1),
	)
	mutants["union-members"] = sixMutant{build: func(harness *harness) string {
		return typeAwareMutant(
			harness,
			normal(harness),
			entry,
			"union-members",
			"rules.ts",
			"from './facts.ts'",
			"from '"+decoderPath+"'",
			"from './types.ts'",
			"from '"+factsPath+"'",
			"from './parameters.ts'",
			"from '"+parametersPath+"'",
		)
	}}
	for _, name := range []string{
		"type-name",
		"raw-shape-constraint",
		"assignability-direction",
		"resolved-signature",
	} {
		mutants[name] = sixMutant{build: func(harness *harness) string {
			return typeAwareMutantNative(harness, name+"-suite", entry, typeAwareMutantArchive(harness, name), false)
		}}
	}
	// Explicit C allocation length, not the native path's extra terminator.
	mutants["facts-length-asan"] = sixMutant{build: func(harness *harness) string {
		lengthArchive := typeAwareMutantArchive(harness, "facts-length")
		return typeAwareSetup(
			harness,
			"facts-length",
			func() string { return harness.mutantNative("facts-length", entry, lengthArchive, true) },
		)
	}, dies: func(result result) bool {
		return result.err != nil && bytes.Contains(result.stderr, []byte("AddressSanitizer: heap-buffer-overflow"))
	}}
	for _, change := range []struct{ name, message string }{
		{"facts-empty", "invalid checker facts frame"},
		{"facts-integer", "invalid checker facts integer"},
		{"facts-length", "invalid checker facts length"},
		{"facts-version", "unsupported checker facts schema"},
	} {
		mutants[change.name+"-bad"] = sixMutant{build: func(harness *harness) string {
			return typeAwareMutantNative(
				harness,
				change.name+"-bad",
				entry,
				typeAwareMutantArchive(harness, change.name+"-bad"),
				false,
			)
		}, dies: func(result result) bool {
			code, ok := result.err.(*exec.ExitError)
			return ok && code.ExitCode() == 70 && bytes.Contains(result.stderr, []byte(change.message))
		}}
	}
	return mutants
}

// witnessRun builds a mutant and writes its witnesses as this shard's manifest.
func (corpus *sixCorpus) witnessRun(harness *harness, name string) (string, string) {
	harness.t.Helper()
	mutant := corpus.mutants[name].build(harness)
	cases := corpus.witnesses(harness, name, func() string { return mutant })
	return mutant, harness.write(name+"-witnesses.manifest", strings.Join(cases, "\n")+"\n")
}

// caught runs a mutant the oracle must catch by output over its witnesses, against the oracle's own run over them.
func (corpus *sixCorpus) caught(harness *harness, name, mutant, cases string) {
	harness.t.Helper()
	want := harness.must(name+"-witness-truth", exec.Command(corpus.oracle, corpus.config, cases))
	got := harness.must(name+"-run", exec.Command(mutant, corpus.config, cases))
	if len(got.stderr) != 0 || bytes.Equal(got.stdout, want.stdout) {
		harness.t.Fatalf("%s not caught by oracle alone", name)
	}
	harness.t.Logf(
		"%s: production cohere oracle catches byte %d over its witnesses; %s",
		name,
		firstDifference(got.stdout, want.stdout),
		summary(got.stdout),
	)
}

// sixDyingWitnesses is how many cases a mutant that must die keeps as witnesses. It dies on nearly every case that asks
// the checker a question, so recording them all would rerun the corpus the witnesses replace.
const sixDyingWitnesses = 8

// witnesses are a planted mutant's witness cases, a product keyed by everything its run reads: the shared
// typeaware inputs hold the checker bridge its archive is built from, the rules, the mutant's own definition in this
// package, cohere's oracle and the fixtures the corpus is drawn from, and the toolchains. Building it runs the mutant
// once. A mutant the oracle must catch by output (dies nil) runs over the whole generated corpus and records every case
// whose findings differ from the oracle's; a mutant that must die runs case by case in corpus order and records the
// first sixDyingWitnesses it dies on. An empty set fails: a mutant no case catches guards nothing.
func (corpus *sixCorpus) witnesses(harness *harness, name string, mutant func() string) []string {
	harness.t.Helper()
	paths, configuration, manifest, truth, dies := corpus.paths, corpus.config, corpus.manifest, corpus.truth, corpus.mutants[name].dies
	record, cachedTypeAwareProductError := cachedTypeAwareProduct("six witnesses "+
		name, func(directory string) (string, error) {
		binary := mutant()
		var found []string
		if dies == nil {
			found = sixDifferingCases(
				harness.must(name+"-witness-search", exec.Command(binary, configuration, manifest)).stdout,
				truth.stdout,
			)
		} else {
			for _, path := range paths {
				one := harness.write(name+"-witness-search.manifest", path+"\n")
				if dies(harness.run(name+"-witness-search", exec.Command(binary, configuration, one))) {
					found = append(found, filepath.Base(path))
					if len(found) == sixDyingWitnesses {
						break
					}
				}
			}
		}
		file := filepath.Join(directory, "witnesses.txt")
		return file, os.WriteFile(file, []byte(strings.Join(found, "\n")), 0644)
	})
	if cachedTypeAwareProductError != nil {
		harness.t.Fatal(cachedTypeAwareProductError)
	}
	data, readError := os.ReadFile(record)
	if readError != nil {
		harness.t.Fatal(readError)
	}
	byName := make(map[string]string, len(paths))
	for _, path := range paths {
		byName[filepath.Base(path)] = path
	}
	var witnesses []string
	for _, base := range strings.Fields(string(data)) {
		path, ok := byName[base]
		if !ok {
			harness.t.Fatalf("%s witness %s is not a generated case", name, base)
		}
		witnesses = append(witnesses, path)
	}
	if len(witnesses) == 0 {
		harness.t.Fatalf("%s: no generated case witnesses this mutant, so it guards nothing", name)
	}
	harness.t.Logf("%s: %d witnesses of %d generated cases", name, len(witnesses), len(paths))
	return witnesses
}

// sixDifferingCases are the generated cases whose findings differ between two suite outputs: each case's block, from
// its "file" line to the next, compared by the case's file name. The closing "findings N" total is left out, since it
// moves whenever any case does and names none.
func sixDifferingCases(got, want []byte) []string {
	blocks := func(output []byte) map[string]string {
		result := map[string]string{}
		name := ""
		for _, line := range strings.Split(string(output), "\n") {
			if path, ok := strings.CutPrefix(line, "file\t"); ok {
				name = filepath.Base(path)
				result[name] = ""
				continue
			}
			if name != "" && !strings.HasPrefix(line, "findings ") {
				result[name] += line + "\n"
			}
		}
		return result
	}
	gotBlocks, wantBlocks := blocks(got), blocks(want)
	var names []string
	for name, block := range gotBlocks {
		if wanted, ok := wantBlocks[name]; !ok || wanted != block {
			names = append(names, name)
		}
	}
	for name := range wantBlocks {
		if _, ok := gotBlocks[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// sixWitnessProduct builds one mutant's witnesses ahead of the test units, from the corpus the shards read, without
// the rest of the six plan.
func sixWitnessProduct(test *testing.T, name string) {
	corpus := sixCorpusOnce(test)
	repository, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	directory := filepath.Join(productDirectory, "six-witnesses", test.Name())
	if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	// The witness search is the product's fill, shared setup the unit's own budget bounds, as six's corpus harness is,
	// never typeAwareChildLimit's 90 s for case work: a mutant caught by output runs over the whole generated corpus in
	// one command, and a builder's load would decide the product (#h4hrxpe).
	harness := &harness{t: test, repository: repository, directory: directory, parallel: true, sharedSetup: true}
	corpus.witnesses(harness, name, func() string { return corpus.mutants[name].build(harness) })
}
func TestProduct_six_witnesses_wrong_node(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "wrong-node")
}
func TestProduct_six_witnesses_last_declaration(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "last-declaration")
}
func TestProduct_six_witnesses_nullable_default(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "nullable-default")
}
func TestProduct_six_witnesses_union_members(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "union-members")
}
func TestProduct_six_witnesses_type_name(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "type-name")
}
func TestProduct_six_witnesses_raw_shape_constraint(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "raw-shape-constraint")
}
func TestProduct_six_witnesses_assignability_direction(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "assignability-direction")
}
func TestProduct_six_witnesses_resolved_signature(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "resolved-signature")
}
func TestProduct_six_witnesses_facts_length_asan(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "facts-length-asan")
}
func TestProduct_six_witnesses_facts_empty_bad(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "facts-empty-bad")
}
func TestProduct_six_witnesses_facts_integer_bad(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "facts-integer-bad")
}
func TestProduct_six_witnesses_facts_length_bad(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "facts-length-bad")
}
func TestProduct_six_witnesses_facts_version_bad(test *testing.T) {
	test.Parallel()
	sixWitnessProduct(test, "facts-version-bad")
}

// sixMutantNativeProduct builds one planted mutant's native ahead of the test units, by its sixMutants recipe, and a
// rules mutant's emitted C with it (loweredMutant), so its shard and its witness product fetch them instead of lowering
// and compiling inside their own time. Built in the shard, these natives held their shards at 16 to 26 s under load
// (typeaware#1/7, future-87bcfd8d), while facts-length-asan's, already a product (TestProduct_six_facts_length_native),
// took 3 s.
func sixMutantNativeProduct(test *testing.T, name string) {
	harness := typeAwareProductHarness(test)
	mutant, ok := sixMutants(harness, harness.repository)[name]
	if !ok {
		test.Fatalf("no six mutant %s", name)
	}
	mutant.build(harness)
}
func TestProduct_six_wrong_node_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "wrong-node")
}
func TestProduct_six_last_declaration_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "last-declaration")
}
func TestProduct_six_nullable_default_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "nullable-default")
}
func TestProduct_six_union_members_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "union-members")
}
func TestProduct_six_type_name_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "type-name")
}
func TestProduct_six_raw_shape_constraint_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "raw-shape-constraint")
}
func TestProduct_six_assignability_direction_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "assignability-direction")
}
func TestProduct_six_resolved_signature_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "resolved-signature")
}
func TestProduct_six_facts_empty_bad_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "facts-empty-bad")
}
func TestProduct_six_facts_integer_bad_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "facts-integer-bad")
}
func TestProduct_six_facts_length_bad_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "facts-length-bad")
}
func TestProduct_six_facts_version_bad_native(test *testing.T) {
	test.Parallel()
	sixMutantNativeProduct(test, "facts-version-bad")
}
