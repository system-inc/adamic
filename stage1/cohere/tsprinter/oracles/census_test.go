package oracles

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/leancensus"
)

// censusExports is the export data of this package and every package it and its tests import, built ahead
// (TestProduct_TSPrinterOraclesCensusExports), so the library oracles' census type-checks with real types and runs no
// go.
func censusExports(test testing.TB) string {
	test.Helper()
	return buildcache.GoExports(test, "./stage1/cohere/tsprinter/oracles")
}

func TestProduct_TSPrinterOraclesCensusExports(test *testing.T) {
	test.Parallel()
	censusExports(test)
}

// The library oracles' builds and their shards' (#zb9268t), each a LeanProductOf build of what it's handed, read no
// package variable written outside its initializer, here or in any package they reach, and reach nothing the oracles
// binary's recipe doesn't key: a lean caller's build keeps the settling binary's whole recipe, so every package of the
// tree the binary links is its slice. The guard that keeps a process from building an oracle is read outside them
// (oracleFound). Mutants that each fail it: a build reading oracleProducer; one reading pinnedCorpus through
// tsPrinterLabelRoots.
func TestALibraryOraclesBuildReadsNoStateButItsInitializers(test *testing.T) {
	test.Parallel()
	// TestMain runs the package's tests from stage1/cohere/tsprinter, where Repository is the checkout.
	root, absolutePathError := filepath.Abs(Repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	exports, tree, _, importsError := leancensus.Imports(censusExports(test))
	if importsError != nil {
		test.Fatal(importsError)
	}
	census := leancensus.Census{
		Root:           root,
		Directory:      filepath.Join(root, "stage1", "cohere", "tsprinter", "oracles"),
		ImportPath:     leancensus.ModulePath + "/stage1/cohere/tsprinter/oracles",
		Entries:        []string{"libraryOracleBuild", "statementShardsBuild", "expressionShardsBuild"},
		Declared:       []string{"stage1/cohere/tsprinter/oracles/oracles.go"},
		Slice:          tree,
		Exports:        exports,
		Tree:           tree,
		TestsBuildLean: true,
	}
	findings, findingsError := census.Findings()
	if findingsError != nil {
		test.Fatal(findingsError)
	}
	if len(findings) > 0 {
		test.Fatalf("a library oracle's build reaches past what its key holds:\n%s", strings.Join(findings, "\n"))
	}
}

// A process that may not build an oracle product only finds it (oracleFound): under a trace and outside the oracles
// binary, a missing product fails, naming it, and nothing is built; the oracles binary, or any process outside a trace,
// isn't handed one, so its caller builds. Mutants that each fail it: the find-only lookup handed a build; a
// non-producer under a trace let through to build; the producer made to look.
// Not parallel: test.Setenv.
func TestAProcessThatMayNotBuildAnOracleOnlyFindsIt(test *testing.T) {
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", test.TempDir())
	test.Setenv("ADAMIC_BUILD_OUTPUT_DIR", "")
	test.Setenv("ADAMIC_BUILD_STORE", "off")
	inputs := oracleInputs(buildcache.Inputs{Name: "tsprinter-oracle-found-test", Flags: []string{"never built"}})
	for _, each := range []struct {
		name              string
		producer, tracing bool
		refused           bool
	}{
		{"the oracles binary under a trace builds", true, true, false},
		{"a process outside a trace builds", false, false, false},
		{"another process under a trace only looks, and a missing product refuses it", false, true, true},
	} {
		test.Run(each.name, func(test *testing.T) {
			directory, found, oracleFoundError := oracleFound(inputs, each.producer, each.tracing)
			switch {
			case each.refused && (oracleFoundError == nil || !strings.Contains(oracleFoundError.Error(), inputs.Name)):
				test.Fatalf(
					"oracleFound: found %v at %q, err %v; want it refused, naming %s",
					found,
					directory,
					oracleFoundError,
					inputs.Name,
				)
			case !each.refused && (oracleFoundError != nil || found):
				test.Fatalf(
					"oracleFound: found %v at %q, err %v; want its caller to build",
					found,
					directory,
					oracleFoundError,
				)
			}
		})
	}
}
