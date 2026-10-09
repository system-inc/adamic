package tsprinter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProduct_TSPrinterExpressionsCorpus(t *testing.T) { t.Parallel(); expressionCorpus(t, false) }
func TestProduct_TSPrinterStatementsCorpus(t *testing.T)  { t.Parallel(); statementCorpus(t) }
func printerGateTSCCorpus(t *testing.T, family string) {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	files, err := trackedRootFiles(t, root, "stage3/drivers/tsc/corpus")
	if err != nil {
		t.Fatal(err)
	}
	gaps, _ := filepath.Abs("testdata/notyet.json")
	tsPrinterOracleOutputs(t, root, files, gaps, family, printerGateGoOracle(t)+"/oracle")
}
func TestProduct_TSPrinterTSCExpressionsCorpus(t *testing.T) {
	t.Parallel()
	printerGateTSCCorpus(t, "expressions")
}
func TestProduct_TSPrinterTSCStatementsCorpus(t *testing.T) {
	t.Parallel()
	printerGateTSCCorpus(t, "statements")
}
func TestProduct_TSPrinterTSCManifest(t *testing.T) { t.Parallel(); tscAgreementPrepare(t) }

func statementDeclaredLibraryOracle(t *testing.T, number, side int) {
	t.Helper()
	cases, want, specs := statementCorpus(t)
	shard := statementShards(t, cases, want, specs)[number]
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	compilerHash := statementFileHash(t, executable)
	name, script, files, flags := "npm", "testdata/expressions.mjs", []string{"stage1/cohere/tsprinter/testdata/expressions.mjs", "stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json", "stage1/cohere/tsprinter/testdata/prettier-differences.json"}, []string{"npm bytes sha256=" + statementDirectoryHash(t, library)}
	if side == 1 {
		name, script = "embedded", "testdata/embedded.mjs"
		library, _ = filepath.Abs(filepath.Join(repository, "cohere/internal/format/prettier/bundles"))
		files = []string{"stage1/cohere/tsprinter/testdata/embedded.mjs", "stage1/cohere/tsprinter/testdata/tsc-upstream-differences.json", "cohere/internal/format/prettier/bundles"}
		flags = nil
	}
	script, _ = filepath.Abs(script)
	statementLibraryOracle(t, number, side, shard, compilerHash, "", name, script, library, files, flags)
}
func TestProduct_TSPrinterStatementsNPM_000(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 0, 0)
}
func TestProduct_TSPrinterStatementsNPM_001(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 1, 0)
}
func TestProduct_TSPrinterStatementsNPM_002(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 2, 0)
}
func TestProduct_TSPrinterStatementsNPM_003(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 3, 0)
}
func TestProduct_TSPrinterStatementsNPM_004(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 4, 0)
}
func TestProduct_TSPrinterStatementsNPM_005(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 5, 0)
}
func TestProduct_TSPrinterStatementsNPM_006(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 6, 0)
}
func TestProduct_TSPrinterStatementsNPM_007(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 7, 0)
}
func TestProduct_TSPrinterStatementsNPM_008(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 8, 0)
}
func TestProduct_TSPrinterStatementsNPM_009(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 9, 0)
}
func TestProduct_TSPrinterStatementsNPM_010(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 10, 0)
}
func TestProduct_TSPrinterStatementsNPM_011(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 11, 0)
}
func TestProduct_TSPrinterStatementsNPM_012(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 12, 0)
}
func TestProduct_TSPrinterStatementsNPM_013(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 13, 0)
}
func TestProduct_TSPrinterStatementsNPM_014(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 14, 0)
}
func TestProduct_TSPrinterStatementsNPM_015(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 15, 0)
}
func TestProduct_TSPrinterStatementsEmbedded_000(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 0, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_001(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 1, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_002(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 2, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_003(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 3, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_004(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 4, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_005(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 5, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_006(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 6, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_007(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 7, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_008(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 8, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_009(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 9, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_010(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 10, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_011(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 11, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_012(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 12, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_013(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 13, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_014(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 14, 1)
}
func TestProduct_TSPrinterStatementsEmbedded_015(t *testing.T) {
	t.Parallel()
	statementDeclaredLibraryOracle(t, 15, 1)
}
