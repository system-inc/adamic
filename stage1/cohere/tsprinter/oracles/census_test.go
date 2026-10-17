package oracles

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/leancensus"
)

// censusExports is the export data of this package and every package it and its tests import, built ahead
// (TestProduct_TSPrinterOraclesCensusExports), so the library oracles' census type-checks with real types and runs no go.
func censusExports(t testing.TB) string {
	t.Helper()
	return buildcache.GoExports(t, "./stage1/cohere/tsprinter/oracles")
}

func TestProduct_TSPrinterOraclesCensusExports(t *testing.T) {
	t.Parallel()
	censusExports(t)
}

// The library oracles' builds and their shards' (#zb9268t), each a LeanProductOf build of what it's handed, read no
// package variable written outside its initializer, here or in any package they reach, and reach nothing the oracles
// binary's recipe doesn't key: a lean caller's build keeps the settling binary's whole recipe, so every package of the
// tree the binary links is its slice. The guard that keeps a process from building an oracle is read outside them
// (oracleFound). Mutants that each fail it: a build reading oracleProducer; one reading pinnedCorpus through
// tsPrinterLabelRoots.
func TestALibraryOraclesBuildReadsNoStateButItsInitializers(t *testing.T) {
	t.Parallel()
	// TestMain runs the package's tests from stage1/cohere/tsprinter, where Repository is the checkout.
	root, err := filepath.Abs(Repository)
	if err != nil {
		t.Fatal(err)
	}
	exports, tree, _, err := leancensus.Imports(censusExports(t))
	if err != nil {
		t.Fatal(err)
	}
	census := leancensus.Census{
		Root: root, Directory: filepath.Join(root, "stage1", "cohere", "tsprinter", "oracles"), ImportPath: leancensus.ModulePath + "/stage1/cohere/tsprinter/oracles",
		Entries:  []string{"libraryOracleBuild", "statementShardsBuild", "expressionShardsBuild"},
		Declared: []string{"stage1/cohere/tsprinter/oracles/oracles.go"},
		Slice:    tree,
		Exports:  exports, Tree: tree, TestsBuildLean: true,
	}
	findings, err := census.Findings()
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) > 0 {
		t.Fatalf("a library oracle's build reaches past what its key holds:\n%s", strings.Join(findings, "\n"))
	}
}

// A process that may not build an oracle product only finds it (oracleFound): under a trace and outside the oracles
// binary, a missing product fails, naming it, and nothing is built; the oracles binary, or any process outside a trace,
// isn't handed one, so its caller builds. Mutants that each fail it: the find-only lookup handed a build; a non-producer
// under a trace let through to build; the producer made to look.
// Not parallel: t.Setenv.
func TestAProcessThatMayNotBuildAnOracleOnlyFindsIt(t *testing.T) {
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", t.TempDir())
	t.Setenv("ADAMIC_BUILD_OUTPUT_DIR", "")
	t.Setenv("ADAMIC_BUILD_STORE", "off")
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
		t.Run(each.name, func(t *testing.T) {
			directory, found, err := oracleFound(inputs, each.producer, each.tracing)
			switch {
			case each.refused && (err == nil || !strings.Contains(err.Error(), inputs.Name)):
				t.Fatalf("oracleFound: found %v at %q, err %v; want it refused, naming %s", found, directory, err, inputs.Name)
			case !each.refused && (err != nil || found):
				t.Fatalf("oracleFound: found %v at %q, err %v; want its caller to build", found, directory, err)
			}
		})
	}
}
