package native

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// compilerNames are the C compilers a Go file can run by name.
var compilerNames = []string{"clang", "clang++", "cc", "gcc"}

// compilerSites are the functions outside this package that name a C compiler, each with why it may: every product a
// test builds by compiling C goes through Build, BuildFlags or BuildTSGo, which give clang no absolute path to read. A
// product site that ran clang itself kept its build directory in a sanitized binary's globals, random per build, twice
// (landable-15's markdown list layout, landable-18's malformed events). A version query, buildcache.Tool(name,
// "--version" or "-v"), passes without a reason.
var compilerSites = map[string]string{
	"internal/fuzz/confirm_test.go:TestFindingReproducesAlone":                                         "names clang as a fixture failure stage; compiles nothing",
	"internal/fuzz/confirm_test.go:flakyCheckout":                                                      "writes fixture tools that simulate failed executions; compiles nothing",
	"bench/run.go:describeMachine":                                                                     "prints clang's version for a benchmark's header; compiles nothing",
	"cmd/adamic-gate/main.go:buildFlags":                                                               "records clang's version in a gate's machine line; compiles nothing",
	"stage1/cohere/cssnumbers/shards_test.go:numbersClangToolchain":                                    "checks clang runs before naming it in a key; compiles nothing",
	"stage1/cohere/cssstrings/shards_test.go:stringsClangToolchain":                                    "checks clang runs before naming it in a key; compiles nothing",
	"bridge/tsgo/products_test.go:buildBridgeProduct":                                                  "bridge's own C test driver, compiled from the repository by relative names with its compilation directory named \".\"",
	"cloud/reports/decode-ascii/build.go:main":                                                         "a report tool run by hand, never a product",
	"cmd/adamic-gate/resume.go:executionIdentity":                                                      "asks clang for its resource directory, to name the machine; compiles nothing",
	"cmd/adamic-test262/cache.go:resolveTools":                                                         "resolves the clang test262's runner runs and its own cache is keyed by; compiles nothing",
	"cmd/adamic-test262/corpus_test.go:TestMiniRunner":                                                 "skips when clang is missing; compiles nothing",
	"cmd/adamic-test262/run.go:attempt":                                                                "test262's runner links each case into its own run directory, never a product",
	"cmd/adamic-test262/run_test.go:TestProgramCPUDeadline":                                            "compiles a fixture into a t.TempDir(), never a product",
	"internal/buildcache/precise_test.go:TestWhatABuildReadsOfTheMachineIsKeyed":                       "a fake system directory's cc link, a path in a trace; nothing runs",
	"internal/buildcache/precise_test.go:TestWhatABuildOnlyLooksForOrRunsThroughALinkIsKeyed":          "a fake tools directory's clang script, run through a link in a trace; nothing compiles",
	"internal/toolchain/toolchain_test.go:TestAToolsCompilersAreClangUnlessTheCallerNamesOne":          "checks Env names clang and clang++ as CC and CXX; compiles nothing",
	"internal/fuzz/run.go:tryFile":                                                                     "the fuzzer compiles each case into its own scratch directory, never a product",
	"internal/oracle/runtimes_test.go:TestARuntimeProductKeyIsTheSameThroughEitherLink":                "finds clang to reach it through two links and key a runtime; compiles nothing",
	"internal/oracle/wasi_test.go:TestWASIRunnerCatchesMutants":                                        "the WASI SDK's clang, into a t.TempDir(), never a product",
	"internal/oracle/wasi_test.go:runWASIEmissionShard":                                                "the WASI SDK's clang, into a t.TempDir(), never a product",
	"stage1/cohere/css/profile_test.go:TestCSSProfileArtifacts":                                        "a profiling build into a scratch directory, never a product",
	"stage1/cohere/css/setup_products_test.go:TestCSSGoOraclesRekeyCompiler":                           "keys the Go oracles with named compilers and compares keys; compiles nothing",
	"stage1/cohere/css/setup_products_test.go:TestCSSGoOraclesKeyTheCompilersEnvGivesGo":               "sets CC and CXX to check the key reads them from toolchain.Env; compiles nothing",
	"stage1/cohere/cssnumbers/setup_products_test.go:TestCSSNumbersGoOracleRekeyCompiler":              "keys the Go oracle with named compilers and compares keys; compiles nothing",
	"stage1/cohere/cssnumbers/setup_products_test.go:TestCSSNumbersGoOracleKeysTheCompilersEnvGivesGo": "sets CC and CXX to check the key reads them from toolchain.Env; compiles nothing",
	"stage1/cohere/estree/misc_shards_test.go:miscArchiverTool":                                        "finds the archiver beside clang to name it in a key; compiles nothing",
	"stage1/cohere/estree/recovery_cache_test.go:recoveryStableNative":                                 "compiles from its own work directory by relative names, with that directory's prefix maps",
	"stage1/cohere/json/port_matches_prepared_split_test.go:portMatchesJsonClangToolchain":             "names clang and its archiver in a key; compiles nothing",
	"stage1/cohere/lint/jsx_shards_test.go:TestJsxSetupKeyNamesTheBuildsCompiler":                      "sets CC to clang or gcc for child go env runs to check the key names it; compiles nothing",
	"stage1/cohere/lint/profile_compilation_main_test.go:compilationProductKinds":                      "the profiled build, unsanitized, with its directory prefix-mapped: debug info is all that names a path",
	"stage1/cohere/lint/profile_test.go:buildProfile":                                                  "a profiling build into a scratch directory, never a product",
	"stage1/cohere/markdownblocks/support_test.go:nativeMutant":                                        "a mutant compiled into a t.TempDir(), never a product",
	"stage1/cohere/markdownblocks/go_compilers_test.go:TestMarkdownGoBuildsRekeyCompiler":              "keys the Go builds with named compilers and compares keys; compiles nothing",
	"stage1/cohere/markdownblocks/go_compilers_test.go:TestMarkdownGoBuildsKeyTheCompilersEnvGivesGo":  "sets CC and CXX to check the keys read them from toolchain.Env; compiles nothing",
	"stage1/cohere/typeaware/typeaware_cache_test.go:TestTypeAwareKeysNameTheBuildsCompiler":           "sets CC to check the keys name the compiler cgo runs; compiles nothing",
	"stage1/cohere/typeaware/volume_profile_Controls_test.go:typeAwareCompilerKey":                     "names cgo's compiler in a key; compiles nothing",
	"stage1/typescript/scanner/profile_test.go:TestProfileArtifacts":                                   "a profiling build into a scratch directory, never a product",
}

// Every Go file outside this package that names a C compiler does so in a function compilerSites gives a reason for:
// a new product site that runs clang itself reds here, where a grep for one call shape let the second one through
// (#rmmw95x). Each reason is a judgment, so a new site is one someone made. A listed function that no longer names a
// compiler reds too, so the list says what is. Mutants: a hand-rolled clang call in a product; a stale entry.
func TestEveryCCompilerRunsThroughNative(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	var unlisted []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			if relative == "cohere" || relative == "internal/native" || entry.Name() == "testdata" || entry.Name() == "node_modules" || strings.HasPrefix(entry.Name(), ".") && relative != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(relative, ".go") {
			return nil
		}
		files := token.NewFileSet()
		parsed, err := parser.ParseFile(files, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			site := relative + ":" + function.Name.Name
			ast.Inspect(function.Body, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok && versionQuery(call) {
					return false
				}
				literal, ok := node.(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					return true
				}
				if value, err := strconv.Unquote(literal.Value); err == nil && slices.Contains(compilerNames, value) {
					if _, listed := compilerSites[site]; listed {
						seen[site] = true
					} else {
						unlisted = append(unlisted, files.Position(literal.Pos()).String()+" in "+function.Name.Name)
					}
				}
				return true
			})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, place := range unlisted {
		t.Errorf("%s names a C compiler: build a product's C through native.Build, BuildFlags or BuildTSGo, or list the function in compilerSites with why it may", strings.TrimPrefix(place, root+"/"))
	}
	for site := range compilerSites {
		if !seen[site] {
			t.Errorf("compilerSites lists %s, which names no C compiler now: take it off", site)
		}
	}
}

// versionQuery is buildcache.Tool(name, "--version"), (name, "-v") or (name, "-print-..."): a tool's report for a
// key, never a compile.
func versionQuery(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Tool" || len(call.Args) != 2 {
		return false
	}
	if pkg, ok := selector.X.(*ast.Ident); !ok || pkg.Name != "buildcache" {
		return false
	}
	flag, ok := call.Args[1].(*ast.BasicLit)
	return ok && (flag.Value == `"--version"` || flag.Value == `"-v"` || strings.HasPrefix(flag.Value, `"-print-`))
}
