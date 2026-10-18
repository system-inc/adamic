package oracles

import (
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"testing"
)

// TestMain is shared with the parent json tests.
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	EnableProductProduction()
	os.Exit(m.Run())
}

// TestProduct_JSONGoOracle is shared with the parent json tests.
func TestProduct_JSONGoOracle(t *testing.T) { t.Parallel(); FormatterOracle(t) }

// TestProduct_JSONAuditOracle is shared with the parent json tests. AuditOracle returns the oracle binary's path, as
// buildcache.GoTest does: its product directory holds it.
func TestProduct_JSONAuditOracle(t *testing.T) {
	t.Parallel()
	binary := AuditOracle(t)
	buildcache.RequireArtifacts(t, filepath.Dir(binary), buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"-test.list=^$"}})
}

// TestProduct_JSONClassOrderAnswer is shared with the parent json tests.
func TestProduct_JSONClassOrderAnswer(t *testing.T) { t.Parallel(); ClassOrderAnswer(t) }

// TestProduct_JSONUpstreamParityVerdict is shared with the parent json tests.
func TestProduct_JSONUpstreamParityVerdict(t *testing.T) { t.Parallel(); UpstreamParity(t) }

// TestProduct_JSONPrinterMutant_0 is shared with the parent json tests.
func TestProduct_JSONPrinterMutant_0(t *testing.T) {
	t.Parallel()
	m := printerMutations[0]
	PrinterMutant(t, []buildcache.Mutation{{File: "cohere/internal/format/javascript/" + m.File, Before: m.From, After: m.To}})
}

// TestProduct_JSONPrinterMutant_1 is shared with the parent json tests.
func TestProduct_JSONPrinterMutant_1(t *testing.T) {
	t.Parallel()
	m := printerMutations[1]
	PrinterMutant(t, []buildcache.Mutation{{File: "cohere/internal/format/javascript/" + m.File, Before: m.From, After: m.To}})
}

// TestProduct_JSONPrinterMutant_2 is shared with the parent json tests.
func TestProduct_JSONPrinterMutant_2(t *testing.T) {
	t.Parallel()
	m := printerMutations[2]
	PrinterMutant(t, []buildcache.Mutation{{File: "cohere/internal/format/javascript/" + m.File, Before: m.From, After: m.To}})
}
