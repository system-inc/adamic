package oracles

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"testing"
)

// TestMain is shared with the parent json tests.
func TestMain(m *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	EnableProductProduction()
	os.Exit(m.Run())
}

// TestProduct_JSONGoOracle is shared with the parent json tests.
func TestProduct_JSONGoOracle(test *testing.T) { test.Parallel(); FormatterOracle(test) }

// TestProduct_JSONAuditOracle is shared with the parent json tests. AuditOracle returns the oracle binary's path, as
// buildcache.GoTest does: its product directory holds it.
func TestProduct_JSONAuditOracle(test *testing.T) {
	test.Parallel()
	binary := AuditOracle(test)
	buildcache.RequireArtifacts(
		test,
		filepath.Dir(binary),
		buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"-test.list=^$"}},
	)
}

// TestProduct_JSONClassOrderAnswer is shared with the parent json tests.
func TestProduct_JSONClassOrderAnswer(test *testing.T) { test.Parallel(); ClassOrderAnswer(test) }

// TestProduct_JSONUpstreamParityVerdict is shared with the parent json tests.
func TestProduct_JSONUpstreamParityVerdict(test *testing.T) { test.Parallel(); UpstreamParity(test) }

// TestProduct_JSONPrinterMutant_0 is shared with the parent json tests.
func TestProduct_JSONPrinterMutant_0(test *testing.T) {
	test.Parallel()
	m := printerMutations[0]
	PrinterMutant(
		test,
		[]buildcache.Mutation{{File: "cohere/internal/format/javascript/" + m.File, Before: m.From, After: m.To}},
	)
}

// TestProduct_JSONPrinterMutant_1 is shared with the parent json tests.
func TestProduct_JSONPrinterMutant_1(test *testing.T) {
	test.Parallel()
	m := printerMutations[1]
	PrinterMutant(
		test,
		[]buildcache.Mutation{{File: "cohere/internal/format/javascript/" + m.File, Before: m.From, After: m.To}},
	)
}

// TestProduct_JSONPrinterMutant_2 is shared with the parent json tests.
func TestProduct_JSONPrinterMutant_2(test *testing.T) {
	test.Parallel()
	m := printerMutations[2]
	PrinterMutant(
		test,
		[]buildcache.Mutation{{File: "cohere/internal/format/javascript/" + m.File, Before: m.From, After: m.To}},
	)
}
