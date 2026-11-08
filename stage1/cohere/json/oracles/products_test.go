package oracles

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"testing"
)

// TestMain is shared with the parent json tests.
func TestMain(suite *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	EnableProductProduction()
	os.Exit(suite.Run())
}

// TestProduct_JsonGoOracle is shared with the parent json tests.
func TestProduct_JsonGoOracle(test *testing.T) { test.Parallel(); FormatterOracle(test) }

// TestProduct_JsonAuditOracle is shared with the parent json tests. AuditOracle returns the oracle binary's path, as
// buildcache.GoTest does: its product directory holds it.
func TestProduct_JsonAuditOracle(test *testing.T) {
	test.Parallel()
	binary := AuditOracle(test)
	buildcache.RequireArtifacts(
		test,
		filepath.Dir(binary),
		buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"-test.list=^$"}},
	)
}

// TestProduct_JsonClassOrderAnswer is shared with the parent json tests.
func TestProduct_JsonClassOrderAnswer(test *testing.T) { test.Parallel(); ClassOrderAnswer(test) }

// TestProduct_JsonUpstreamParityVerdict is shared with the parent json tests.
func TestProduct_JsonUpstreamParityVerdict(test *testing.T) { test.Parallel(); UpstreamParity(test) }

// TestProduct_JsonPrinterMutant_0 is shared with the parent json tests.
func TestProduct_JsonPrinterMutant_0(test *testing.T) {
	test.Parallel()
	printerMutation := printerMutations[0]
	PrinterMutant(
		test,
		[]buildcache.Mutation{{
			File:   "cohere/internal/format/javascript/" + printerMutation.File,
			Before: printerMutation.From,
			After:  printerMutation.To,
		}},
	)
}

// TestProduct_JsonPrinterMutant_1 is shared with the parent json tests.
func TestProduct_JsonPrinterMutant_1(test *testing.T) {
	test.Parallel()
	printerMutation := printerMutations[1]
	PrinterMutant(
		test,
		[]buildcache.Mutation{{
			File:   "cohere/internal/format/javascript/" + printerMutation.File,
			Before: printerMutation.From,
			After:  printerMutation.To,
		}},
	)
}

// TestProduct_JsonPrinterMutant_2 is shared with the parent json tests.
func TestProduct_JsonPrinterMutant_2(test *testing.T) {
	test.Parallel()
	printerMutation := printerMutations[2]
	PrinterMutant(
		test,
		[]buildcache.Mutation{{
			File:   "cohere/internal/format/javascript/" + printerMutation.File,
			Before: printerMutation.From,
			After:  printerMutation.To,
		}},
	)
}
