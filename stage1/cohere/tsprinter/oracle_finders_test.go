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

func TestProduct_TSPrinterFindsPinnedCorpus(t *testing.T) { t.Parallel(); oracles.PinnedCorpusFiles(t) }

func TestProduct_TSPrinterFindsDocumentCorpus(t *testing.T) { t.Parallel(); oracles.DocumentCorpus(t) }

func TestProduct_TSPrinterFindsDocumentOracle(t *testing.T) { t.Parallel(); oracles.DocumentOracle(t) }

func TestProduct_TSPrinterFindsMutantOracle_000(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 0)
}

func TestProduct_TSPrinterFindsMutantOracle_001(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 1)
}

func TestProduct_TSPrinterFindsMutantOracle_002(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 2)
}

func TestProduct_TSPrinterFindsMutantOracle_003(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 3)
}

func TestProduct_TSPrinterFindsMutantOracle_004(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 4)
}

func TestProduct_TSPrinterFindsMutantOracle_005(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 5)
}

func TestProduct_TSPrinterFindsMutantOracle_006(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 6)
}

func TestProduct_TSPrinterFindsMutantOracle_007(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 7)
}

func TestProduct_TSPrinterFindsMutantOracle_008(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 8)
}

func TestProduct_TSPrinterFindsMutantOracle_009(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 9)
}

func TestProduct_TSPrinterFindsMutantOracle_010(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 10)
}

func TestProduct_TSPrinterFindsMutantOracle_011(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 11)
}

func TestProduct_TSPrinterFindsMutantOracle_012(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 12)
}

func TestProduct_TSPrinterFindsMutantOracle_013(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 13)
}

func TestProduct_TSPrinterFindsMutantOracle_014(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 14)
}

func TestProduct_TSPrinterFindsMutantOracle_015(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 15)
}

func TestProduct_TSPrinterFindsMutantOracle_016(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 16)
}

func TestProduct_TSPrinterFindsMutantOracle_017(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 17)
}

func TestProduct_TSPrinterFindsMutantOracle_018(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 18)
}

func TestProduct_TSPrinterFindsMutantOracle_019(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 19)
}

func TestProduct_TSPrinterFindsMutantOracle_020(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 20)
}

func TestProduct_TSPrinterFindsMutantOracle_021(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 21)
}

func TestProduct_TSPrinterFindsMutantOracle_022(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 22)
}

func TestProduct_TSPrinterFindsMutantOracle_023(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 23)
}

func TestProduct_TSPrinterFindsMutantOracle_024(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 24)
}

func TestProduct_TSPrinterFindsMutantOracle_025(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 25)
}

func TestProduct_TSPrinterFindsMutantOracle_026(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 26)
}

func TestProduct_TSPrinterFindsMutantOracle_027(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 27)
}

func TestProduct_TSPrinterFindsMutantOracle_028(t *testing.T) {
	t.Parallel()
	oracles.MutantOracleProduct(t, 28)
}

func TestProduct_TSPrinterFindsExpressionsCorpus(t *testing.T) {
	t.Parallel()
	oracles.ExpressionCorpus(t, false)
}

func TestProduct_TSPrinterFindsStatementsCorpus(t *testing.T) {
	t.Parallel()
	oracles.StatementCorpus(t)
}

func TestProduct_TSPrinterFindsTSCManifest(t *testing.T) {
	t.Parallel()
	oracles.TSCAgreementPrepare(t)
}

func TestProduct_TSPrinterFindsStatementsNPM_000(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 0, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_001(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 1, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_002(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 2, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_003(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 3, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_004(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 4, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_005(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 5, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_006(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 6, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_007(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 7, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_008(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 8, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_009(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 9, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_010(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 10, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_011(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 11, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_012(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 12, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_013(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 13, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_014(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 14, 0)
}

func TestProduct_TSPrinterFindsStatementsNPM_015(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 15, 0)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_000(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 0, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_001(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 1, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_002(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 2, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_003(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 3, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_004(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 4, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_005(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 5, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_006(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 6, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_007(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 7, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_008(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 8, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_009(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 9, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_010(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 10, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_011(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 11, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_012(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 12, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_013(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 13, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_014(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 14, 1)
}

func TestProduct_TSPrinterFindsStatementsEmbedded_015(t *testing.T) {
	t.Parallel()
	oracles.StatementDeclaredLibraryOracle(t, 15, 1)
}

func TestProduct_TSPrinterFindsExpressionsNPM_000(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 0, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_001(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 1, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_002(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 2, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_003(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 3, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_004(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 4, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_005(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 5, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_006(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 6, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_007(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 7, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_008(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 8, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_009(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 9, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_010(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 10, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_011(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 11, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_012(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 12, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_013(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 13, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_014(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 14, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_015(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 15, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_016(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 16, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_017(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 17, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_018(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 18, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_019(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 19, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_020(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 20, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_021(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 21, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_022(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 22, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_023(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 23, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_024(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 24, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_025(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 25, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_026(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 26, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_027(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 27, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_028(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 28, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_029(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 29, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_030(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 30, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_031(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 31, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_032(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 32, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_033(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 33, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_034(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 34, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_035(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 35, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_036(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 36, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_037(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 37, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_038(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 38, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_039(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 39, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_040(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 40, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_041(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 41, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_042(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 42, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_043(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 43, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_044(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 44, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_045(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 45, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_046(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 46, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_047(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 47, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_048(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 48, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_049(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 49, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_050(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 50, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_051(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 51, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_052(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 52, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_053(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 53, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_054(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 54, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_055(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 55, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_056(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 56, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_057(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 57, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_058(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 58, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_059(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 59, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_060(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 60, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_061(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 61, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_062(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 62, 0)
}

func TestProduct_TSPrinterFindsExpressionsNPM_063(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 63, 0)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_000(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 0, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_001(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 1, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_002(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 2, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_003(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 3, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_004(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 4, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_005(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 5, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_006(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 6, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_007(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 7, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_008(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 8, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_009(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 9, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_010(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 10, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_011(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 11, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_012(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 12, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_013(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 13, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_014(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 14, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_015(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 15, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_016(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 16, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_017(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 17, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_018(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 18, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_019(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 19, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_020(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 20, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_021(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 21, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_022(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 22, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_023(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 23, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_024(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 24, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_025(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 25, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_026(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 26, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_027(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 27, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_028(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 28, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_029(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 29, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_030(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 30, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_031(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 31, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_032(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 32, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_033(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 33, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_034(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 34, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_035(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 35, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_036(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 36, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_037(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 37, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_038(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 38, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_039(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 39, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_040(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 40, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_041(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 41, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_042(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 42, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_043(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 43, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_044(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 44, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_045(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 45, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_046(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 46, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_047(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 47, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_048(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 48, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_049(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 49, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_050(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 50, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_051(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 51, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_052(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 52, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_053(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 53, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_054(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 54, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_055(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 55, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_056(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 56, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_057(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 57, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_058(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 58, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_059(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 59, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_060(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 60, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_061(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 61, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_062(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 62, 1)
}

func TestProduct_TSPrinterFindsExpressionsEmbedded_063(t *testing.T) {
	t.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(t, 63, 1)
}

func TestProduct_TSPrinterFindsGoOracle(t *testing.T) { t.Parallel(); oracles.PrinterGateGoOracle(t) }
