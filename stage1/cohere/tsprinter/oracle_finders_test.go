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

func TestProduct_TsPrinterFindsPinnedCorpus(test *testing.T) {
	test.Parallel()
	oracles.PinnedCorpusFiles(test)
}

func TestProduct_TsPrinterFindsDocumentCorpus(test *testing.T) {
	test.Parallel()
	oracles.DocumentCorpus(test)
}

func TestProduct_TsPrinterFindsDocumentOracle(test *testing.T) {
	test.Parallel()
	oracles.DocumentOracle(test)
}

func TestProduct_TsPrinterFindsMutantOracle_000(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 0)
}

func TestProduct_TsPrinterFindsMutantOracle_001(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 1)
}

func TestProduct_TsPrinterFindsMutantOracle_002(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 2)
}

func TestProduct_TsPrinterFindsMutantOracle_003(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 3)
}

func TestProduct_TsPrinterFindsMutantOracle_004(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 4)
}

func TestProduct_TsPrinterFindsMutantOracle_005(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 5)
}

func TestProduct_TsPrinterFindsMutantOracle_006(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 6)
}

func TestProduct_TsPrinterFindsMutantOracle_007(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 7)
}

func TestProduct_TsPrinterFindsMutantOracle_008(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 8)
}

func TestProduct_TsPrinterFindsMutantOracle_009(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 9)
}

func TestProduct_TsPrinterFindsMutantOracle_010(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 10)
}

func TestProduct_TsPrinterFindsMutantOracle_011(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 11)
}

func TestProduct_TsPrinterFindsMutantOracle_012(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 12)
}

func TestProduct_TsPrinterFindsMutantOracle_013(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 13)
}

func TestProduct_TsPrinterFindsMutantOracle_014(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 14)
}

func TestProduct_TsPrinterFindsMutantOracle_015(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 15)
}

func TestProduct_TsPrinterFindsMutantOracle_016(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 16)
}

func TestProduct_TsPrinterFindsMutantOracle_017(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 17)
}

func TestProduct_TsPrinterFindsMutantOracle_018(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 18)
}

func TestProduct_TsPrinterFindsMutantOracle_019(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 19)
}

func TestProduct_TsPrinterFindsMutantOracle_020(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 20)
}

func TestProduct_TsPrinterFindsMutantOracle_021(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 21)
}

func TestProduct_TsPrinterFindsMutantOracle_022(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 22)
}

func TestProduct_TsPrinterFindsMutantOracle_023(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 23)
}

func TestProduct_TsPrinterFindsMutantOracle_024(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 24)
}

func TestProduct_TsPrinterFindsMutantOracle_025(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 25)
}

func TestProduct_TsPrinterFindsMutantOracle_026(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 26)
}

func TestProduct_TsPrinterFindsMutantOracle_027(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 27)
}

func TestProduct_TsPrinterFindsMutantOracle_028(test *testing.T) {
	test.Parallel()
	oracles.MutantOracleProduct(test, 28)
}

func TestProduct_TsPrinterFindsExpressionsCorpus(test *testing.T) {
	test.Parallel()
	oracles.ExpressionCorpus(test, false)
}

func TestProduct_TsPrinterFindsStatementsCorpus(test *testing.T) {
	test.Parallel()
	oracles.StatementCorpus(test)
}

func TestProduct_TsPrinterFindsTscManifest(test *testing.T) {
	test.Parallel()
	oracles.TSCAgreementPrepare(test)
}

func TestProduct_TsPrinterFindsStatementsNpm_000(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 0, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_001(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 1, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_002(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 2, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_003(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 3, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_004(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 4, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_005(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 5, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_006(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 6, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_007(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 7, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_008(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 8, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_009(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 9, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_010(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 10, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_011(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 11, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_012(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 12, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_013(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 13, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_014(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 14, 0)
}

func TestProduct_TsPrinterFindsStatementsNpm_015(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 15, 0)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_000(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 0, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_001(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 1, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_002(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 2, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_003(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 3, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_004(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 4, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_005(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 5, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_006(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 6, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_007(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 7, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_008(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 8, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_009(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 9, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_010(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 10, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_011(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 11, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_012(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 12, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_013(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 13, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_014(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 14, 1)
}

func TestProduct_TsPrinterFindsStatementsEmbedded_015(test *testing.T) {
	test.Parallel()
	oracles.StatementDeclaredLibraryOracle(test, 15, 1)
}

func TestProduct_TsPrinterFindsExpressionsNpm_000(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 0, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_001(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 1, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_002(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 2, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_003(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 3, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_004(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 4, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_005(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 5, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_006(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 6, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_007(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 7, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_008(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 8, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_009(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 9, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_010(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 10, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_011(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 11, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_012(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 12, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_013(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 13, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_014(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 14, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_015(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 15, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_016(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 16, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_017(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 17, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_018(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 18, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_019(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 19, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_020(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 20, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_021(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 21, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_022(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 22, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_023(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 23, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_024(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 24, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_025(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 25, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_026(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 26, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_027(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 27, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_028(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 28, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_029(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 29, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_030(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 30, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_031(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 31, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_032(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 32, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_033(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 33, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_034(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 34, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_035(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 35, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_036(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 36, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_037(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 37, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_038(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 38, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_039(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 39, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_040(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 40, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_041(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 41, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_042(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 42, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_043(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 43, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_044(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 44, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_045(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 45, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_046(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 46, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_047(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 47, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_048(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 48, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_049(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 49, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_050(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 50, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_051(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 51, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_052(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 52, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_053(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 53, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_054(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 54, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_055(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 55, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_056(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 56, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_057(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 57, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_058(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 58, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_059(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 59, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_060(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 60, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_061(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 61, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_062(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 62, 0)
}

func TestProduct_TsPrinterFindsExpressionsNpm_063(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 63, 0)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_000(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 0, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_001(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 1, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_002(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 2, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_003(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 3, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_004(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 4, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_005(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 5, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_006(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 6, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_007(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 7, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_008(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 8, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_009(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 9, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_010(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 10, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_011(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 11, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_012(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 12, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_013(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 13, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_014(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 14, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_015(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 15, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_016(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 16, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_017(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 17, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_018(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 18, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_019(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 19, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_020(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 20, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_021(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 21, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_022(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 22, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_023(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 23, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_024(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 24, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_025(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 25, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_026(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 26, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_027(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 27, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_028(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 28, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_029(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 29, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_030(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 30, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_031(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 31, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_032(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 32, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_033(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 33, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_034(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 34, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_035(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 35, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_036(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 36, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_037(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 37, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_038(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 38, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_039(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 39, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_040(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 40, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_041(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 41, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_042(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 42, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_043(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 43, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_044(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 44, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_045(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 45, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_046(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 46, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_047(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 47, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_048(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 48, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_049(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 49, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_050(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 50, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_051(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 51, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_052(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 52, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_053(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 53, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_054(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 54, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_055(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 55, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_056(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 56, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_057(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 57, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_058(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 58, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_059(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 59, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_060(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 60, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_061(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 61, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_062(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 62, 1)
}

func TestProduct_TsPrinterFindsExpressionsEmbedded_063(test *testing.T) {
	test.Parallel()
	oracles.ExpressionDeclaredLibraryOracle(test, 63, 1)
}

func TestProduct_TsPrinterFindsGoOracle(test *testing.T) {
	test.Parallel()
	oracles.PrinterGateGoOracle(test)
}
