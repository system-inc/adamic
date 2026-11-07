package tsprinter

import (
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

// These finders put the ordinary tests' oracle reads in this package's product list: a Loom unit fetches only the
// products its own package's TestProduct_ tests built or found. The oracles package owns production (#2ba0c20), and
// cloud/fast-gate/product-order.json starts each finder after its owner and the owners its key reads through. The
// TSPrinterFinds names keep them distinct from the owners, since adamic-gate requires same-named product declarations
// to agree. The two TSC output corpora have none: only the TSC manifest's build reads them, never a hit.

func TestProduct_TSPrinterFindsPinnedCorpus(test *testing.T) {
	test.Parallel()
	oracles.PinnedCorpusFiles(test)
}

func TestProduct_TSPrinterFindsDocumentCorpus(test *testing.T) {
	test.Parallel()
	oracles.DocumentCorpus(test)
}

func TestProduct_TSPrinterFindsDocumentOracle(test *testing.T) {
	test.Parallel()
	oracles.DocumentOracle(test)
}

func TestProduct_TSPrinterFindsMutantOracle_000(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 0)
}

func TestProduct_TSPrinterFindsMutantOracle_001(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 1)
}

func TestProduct_TSPrinterFindsMutantOracle_002(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 2)
}

func TestProduct_TSPrinterFindsMutantOracle_003(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 3)
}

func TestProduct_TSPrinterFindsMutantOracle_004(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 4)
}

func TestProduct_TSPrinterFindsMutantOracle_005(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 5)
}

func TestProduct_TSPrinterFindsMutantOracle_006(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 6)
}

func TestProduct_TSPrinterFindsMutantOracle_007(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 7)
}

func TestProduct_TSPrinterFindsMutantOracle_008(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 8)
}

func TestProduct_TSPrinterFindsMutantOracle_009(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 9)
}

func TestProduct_TSPrinterFindsMutantOracle_010(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 10)
}

func TestProduct_TSPrinterFindsMutantOracle_011(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 11)
}

func TestProduct_TSPrinterFindsMutantOracle_012(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 12)
}

func TestProduct_TSPrinterFindsMutantOracle_013(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 13)
}

func TestProduct_TSPrinterFindsMutantOracle_014(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 14)
}

func TestProduct_TSPrinterFindsMutantOracle_015(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 15)
}

func TestProduct_TSPrinterFindsMutantOracle_016(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 16)
}

func TestProduct_TSPrinterFindsMutantOracle_017(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 17)
}

func TestProduct_TSPrinterFindsMutantOracle_018(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 18)
}

func TestProduct_TSPrinterFindsMutantOracle_019(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 19)
}

func TestProduct_TSPrinterFindsMutantOracle_020(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 20)
}

func TestProduct_TSPrinterFindsMutantOracle_021(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 21)
}

func TestProduct_TSPrinterFindsMutantOracle_022(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 22)
}

func TestProduct_TSPrinterFindsMutantOracle_023(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 23)
}

func TestProduct_TSPrinterFindsMutantOracle_024(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 24)
}

func TestProduct_TSPrinterFindsMutantOracle_025(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 25)
}

func TestProduct_TSPrinterFindsMutantOracle_026(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 26)
}

func TestProduct_TSPrinterFindsMutantOracle_027(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 27)
}

func TestProduct_TSPrinterFindsMutantOracle_028(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 28)
}

func TestProduct_TSPrinterFindsExpressionsCorpus(test *testing.T) {
	test.Parallel()
	oracles.ExpressionCorpus(test, false)
}

func TestProduct_TSPrinterFindsStatementsCorpus(test *testing.T) {
	test.Parallel()
	oracles.StatementCorpus(test)
}

func TestProduct_TSPrinterFindsTSCManifest(test *testing.T) {
	test.Parallel()
	oracles.TSCAgreementPrepare(test)
}

func TestProduct_TSPrinterFindsStatementsNPM_000(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 0, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_001(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 1, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_002(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 2, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_003(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 3, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_004(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 4, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_005(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 5, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_006(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 6, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_007(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 7, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_008(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 8, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_009(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 9, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_010(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 10, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_011(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 11, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_012(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 12, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_013(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 13, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_014(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 14, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_015(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 15, 0)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_000(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 0, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_001(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 1, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_002(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 2, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_003(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 3, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_004(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 4, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_005(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 5, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_006(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 6, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_007(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 7, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_008(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 8, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_009(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 9, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_010(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 10, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_011(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 11, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_012(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 12, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_013(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 13, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_014(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 14, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_015(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 15, 1)
}

func TestProduct_TSPrinterFindsExpressionsNPM_000(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 0, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_001(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 1, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_002(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 2, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_003(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 3, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_004(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 4, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_005(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 5, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_006(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 6, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_007(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 7, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_008(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 8, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_009(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 9, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_010(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 10, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_011(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 11, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_012(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 12, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_013(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 13, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_014(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 14, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_015(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 15, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_016(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 16, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_017(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 17, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_018(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 18, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_019(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 19, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_020(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 20, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_021(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 21, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_022(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 22, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_023(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 23, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_024(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 24, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_025(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 25, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_026(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 26, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_027(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 27, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_028(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 28, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_029(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 29, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_030(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 30, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_031(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 31, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_032(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 32, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_033(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 33, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_034(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 34, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_035(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 35, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_036(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 36, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_037(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 37, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_038(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 38, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_039(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 39, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_040(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 40, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_041(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 41, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_042(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 42, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_043(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 43, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_044(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 44, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_045(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 45, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_046(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 46, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_047(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 47, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_048(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 48, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_049(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 49, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_050(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 50, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_051(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 51, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_052(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 52, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_053(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 53, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_054(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 54, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_055(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 55, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_056(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 56, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_057(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 57, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_058(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 58, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_059(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 59, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_060(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 60, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_061(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 61, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_062(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 62, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_063(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 63, 0)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_000(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 0, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_001(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 1, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_002(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 2, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_003(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 3, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_004(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 4, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_005(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 5, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_006(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 6, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_007(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 7, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_008(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 8, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_009(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 9, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_010(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 10, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_011(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 11, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_012(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 12, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_013(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 13, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_014(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 14, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_015(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 15, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_016(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 16, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_017(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 17, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_018(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 18, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_019(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 19, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_020(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 20, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_021(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 21, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_022(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 22, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_023(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 23, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_024(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 24, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_025(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 25, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_026(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 26, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_027(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 27, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_028(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 28, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_029(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 29, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_030(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 30, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_031(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 31, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_032(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 32, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_033(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 33, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_034(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 34, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_035(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 35, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_036(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 36, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_037(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 37, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_038(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 38, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_039(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 39, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_040(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 40, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_041(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 41, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_042(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 42, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_043(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 43, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_044(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 44, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_045(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 45, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_046(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 46, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_047(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 47, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_048(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 48, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_049(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 49, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_050(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 50, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_051(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 51, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_052(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 52, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_053(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 53, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_054(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 54, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_055(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 55, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_056(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 56, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_057(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 57, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_058(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 58, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_059(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 59, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_060(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 60, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_061(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 61, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_062(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 62, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_063(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 63, 1)
}

func TestProduct_TSPrinterFindsGoOracle(test *testing.T) {
	test.Parallel()
	oracles.PrinterGateGoOracle(test)
}
