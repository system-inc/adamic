package oracles

import "testing"

func TestProduct_TsPrinterPinnedCorpus(test *testing.T) {
	test.Parallel()
	PinnedCorpusFiles(test)
}

// The Go doc oracle's corpus, an oracle product (#g7a9yd3).
func TestProduct_TsPrinterDocumentCorpus(test *testing.T) {
	test.Parallel()
	DocumentCorpus(test)
}

func TestProduct_TSPrinterDocumentOracle(test *testing.T) {
	test.Parallel()
	DocumentOracle(test)
}

func TestProduct_TsPrinterMutantOracle_000(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 0)
}

func TestProduct_TsPrinterMutantOracle_001(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 1)
}

func TestProduct_TsPrinterMutantOracle_002(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 2)
}

func TestProduct_TsPrinterMutantOracle_003(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 3)
}

func TestProduct_TsPrinterMutantOracle_004(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 4)
}

func TestProduct_TsPrinterMutantOracle_005(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 5)
}

func TestProduct_TsPrinterMutantOracle_006(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 6)
}

func TestProduct_TsPrinterMutantOracle_007(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 7)
}

func TestProduct_TsPrinterMutantOracle_008(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 8)
}

func TestProduct_TsPrinterMutantOracle_009(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 9)
}

func TestProduct_TsPrinterMutantOracle_010(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 10)
}

func TestProduct_TsPrinterMutantOracle_011(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 11)
}

func TestProduct_TsPrinterMutantOracle_012(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 12)
}

func TestProduct_TsPrinterMutantOracle_013(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 13)
}

func TestProduct_TsPrinterMutantOracle_014(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 14)
}

func TestProduct_TsPrinterMutantOracle_015(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 15)
}

func TestProduct_TsPrinterMutantOracle_016(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 16)
}

func TestProduct_TsPrinterMutantOracle_017(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 17)
}

func TestProduct_TsPrinterMutantOracle_018(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 18)
}

func TestProduct_TsPrinterMutantOracle_019(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 19)
}

func TestProduct_TsPrinterMutantOracle_020(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 20)
}

func TestProduct_TsPrinterMutantOracle_021(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 21)
}

func TestProduct_TsPrinterMutantOracle_022(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 22)
}

func TestProduct_TsPrinterMutantOracle_023(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 23)
}

func TestProduct_TsPrinterMutantOracle_024(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 24)
}

func TestProduct_TsPrinterMutantOracle_025(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 25)
}

func TestProduct_TsPrinterMutantOracle_026(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 26)
}

func TestProduct_TsPrinterMutantOracle_027(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 27)
}

func TestProduct_TsPrinterMutantOracle_028(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 28)
}

func TestProduct_TsPrinterExpressionsCorpus(test *testing.T) {
	test.Parallel()
	ExpressionCorpus(test, false)
}

func TestProduct_TsPrinterStatementsCorpus(test *testing.T) { test.Parallel(); StatementCorpus(test) }

func TestProduct_TsPrinterTscExpressionsCorpus(test *testing.T) {
	test.Parallel()
	printerGateTscCorpus(test, "expressions")
}

func TestProduct_TsPrinterTscStatementsCorpus(test *testing.T) {
	test.Parallel()
	printerGateTscCorpus(test, "statements")
}

func TestProduct_TsPrinterTscManifest(test *testing.T) { test.Parallel(); TSCAgreementPrepare(test) }

func TestProduct_TsPrinterStatementsNpm_000(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 0, 0)
}

func TestProduct_TsPrinterStatementsNpm_001(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 1, 0)
}

func TestProduct_TsPrinterStatementsNpm_002(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 2, 0)
}

func TestProduct_TsPrinterStatementsNpm_003(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 3, 0)
}

func TestProduct_TsPrinterStatementsNpm_004(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 4, 0)
}

func TestProduct_TsPrinterStatementsNpm_005(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 5, 0)
}

func TestProduct_TsPrinterStatementsNpm_006(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 6, 0)
}

func TestProduct_TsPrinterStatementsNpm_007(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 7, 0)
}

func TestProduct_TsPrinterStatementsNpm_008(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 8, 0)
}

func TestProduct_TsPrinterStatementsNpm_009(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 9, 0)
}

func TestProduct_TsPrinterStatementsNpm_010(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 10, 0)
}

func TestProduct_TsPrinterStatementsNpm_011(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 11, 0)
}

func TestProduct_TsPrinterStatementsNpm_012(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 12, 0)
}

func TestProduct_TsPrinterStatementsNpm_013(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 13, 0)
}

func TestProduct_TsPrinterStatementsNpm_014(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 14, 0)
}

func TestProduct_TsPrinterStatementsNpm_015(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 15, 0)
}

func TestProduct_TsPrinterStatementsEmbedded_000(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 0, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_001(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 1, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_002(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 2, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_003(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 3, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_004(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 4, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_005(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 5, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_006(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 6, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_007(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 7, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_008(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 8, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_009(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 9, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_010(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 10, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_011(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 11, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_012(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 12, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_013(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 13, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_014(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 14, 1)
}

func TestProduct_TsPrinterStatementsEmbedded_015(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 15, 1)
}

func TestProduct_TsPrinterExpressionsNpm_000(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 0, 0)
}

func TestProduct_TsPrinterExpressionsNpm_001(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 1, 0)
}

func TestProduct_TsPrinterExpressionsNpm_002(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 2, 0)
}

func TestProduct_TsPrinterExpressionsNpm_003(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 3, 0)
}

func TestProduct_TsPrinterExpressionsNpm_004(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 4, 0)
}

func TestProduct_TsPrinterExpressionsNpm_005(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 5, 0)
}

func TestProduct_TsPrinterExpressionsNpm_006(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 6, 0)
}

func TestProduct_TsPrinterExpressionsNpm_007(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 7, 0)
}

func TestProduct_TsPrinterExpressionsNpm_008(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 8, 0)
}

func TestProduct_TsPrinterExpressionsNpm_009(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 9, 0)
}

func TestProduct_TsPrinterExpressionsNpm_010(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 10, 0)
}

func TestProduct_TsPrinterExpressionsNpm_011(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 11, 0)
}

func TestProduct_TsPrinterExpressionsNpm_012(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 12, 0)
}

func TestProduct_TsPrinterExpressionsNpm_013(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 13, 0)
}

func TestProduct_TsPrinterExpressionsNpm_014(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 14, 0)
}

func TestProduct_TsPrinterExpressionsNpm_015(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 15, 0)
}

func TestProduct_TsPrinterExpressionsNpm_016(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 16, 0)
}

func TestProduct_TsPrinterExpressionsNpm_017(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 17, 0)
}

func TestProduct_TsPrinterExpressionsNpm_018(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 18, 0)
}

func TestProduct_TsPrinterExpressionsNpm_019(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 19, 0)
}

func TestProduct_TsPrinterExpressionsNpm_020(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 20, 0)
}

func TestProduct_TsPrinterExpressionsNpm_021(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 21, 0)
}

func TestProduct_TsPrinterExpressionsNpm_022(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 22, 0)
}

func TestProduct_TsPrinterExpressionsNpm_023(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 23, 0)
}

func TestProduct_TsPrinterExpressionsNpm_024(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 24, 0)
}

func TestProduct_TsPrinterExpressionsNpm_025(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 25, 0)
}

func TestProduct_TsPrinterExpressionsNpm_026(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 26, 0)
}

func TestProduct_TsPrinterExpressionsNpm_027(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 27, 0)
}

func TestProduct_TsPrinterExpressionsNpm_028(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 28, 0)
}

func TestProduct_TsPrinterExpressionsNpm_029(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 29, 0)
}

func TestProduct_TsPrinterExpressionsNpm_030(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 30, 0)
}

func TestProduct_TsPrinterExpressionsNpm_031(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 31, 0)
}

func TestProduct_TsPrinterExpressionsNpm_032(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 32, 0)
}

func TestProduct_TsPrinterExpressionsNpm_033(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 33, 0)
}

func TestProduct_TsPrinterExpressionsNpm_034(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 34, 0)
}

func TestProduct_TsPrinterExpressionsNpm_035(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 35, 0)
}

func TestProduct_TsPrinterExpressionsNpm_036(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 36, 0)
}

func TestProduct_TsPrinterExpressionsNpm_037(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 37, 0)
}

func TestProduct_TsPrinterExpressionsNpm_038(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 38, 0)
}

func TestProduct_TsPrinterExpressionsNpm_039(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 39, 0)
}

func TestProduct_TsPrinterExpressionsNpm_040(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 40, 0)
}

func TestProduct_TsPrinterExpressionsNpm_041(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 41, 0)
}

func TestProduct_TsPrinterExpressionsNpm_042(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 42, 0)
}

func TestProduct_TsPrinterExpressionsNpm_043(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 43, 0)
}

func TestProduct_TsPrinterExpressionsNpm_044(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 44, 0)
}

func TestProduct_TsPrinterExpressionsNpm_045(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 45, 0)
}

func TestProduct_TsPrinterExpressionsNpm_046(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 46, 0)
}

func TestProduct_TsPrinterExpressionsNpm_047(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 47, 0)
}

func TestProduct_TsPrinterExpressionsNpm_048(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 48, 0)
}

func TestProduct_TsPrinterExpressionsNpm_049(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 49, 0)
}

func TestProduct_TsPrinterExpressionsNpm_050(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 50, 0)
}

func TestProduct_TsPrinterExpressionsNpm_051(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 51, 0)
}

func TestProduct_TsPrinterExpressionsNpm_052(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 52, 0)
}

func TestProduct_TsPrinterExpressionsNpm_053(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 53, 0)
}

func TestProduct_TsPrinterExpressionsNpm_054(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 54, 0)
}

func TestProduct_TsPrinterExpressionsNpm_055(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 55, 0)
}

func TestProduct_TsPrinterExpressionsNpm_056(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 56, 0)
}

func TestProduct_TsPrinterExpressionsNpm_057(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 57, 0)
}

func TestProduct_TsPrinterExpressionsNpm_058(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 58, 0)
}

func TestProduct_TsPrinterExpressionsNpm_059(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 59, 0)
}

func TestProduct_TsPrinterExpressionsNpm_060(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 60, 0)
}

func TestProduct_TsPrinterExpressionsNpm_061(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 61, 0)
}

func TestProduct_TsPrinterExpressionsNpm_062(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 62, 0)
}

func TestProduct_TsPrinterExpressionsNpm_063(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 63, 0)
}

func TestProduct_TsPrinterExpressionsEmbedded_000(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 0, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_001(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 1, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_002(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 2, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_003(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 3, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_004(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 4, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_005(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 5, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_006(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 6, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_007(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 7, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_008(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 8, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_009(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 9, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_010(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 10, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_011(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 11, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_012(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 12, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_013(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 13, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_014(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 14, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_015(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 15, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_016(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 16, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_017(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 17, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_018(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 18, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_019(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 19, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_020(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 20, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_021(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 21, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_022(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 22, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_023(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 23, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_024(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 24, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_025(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 25, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_026(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 26, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_027(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 27, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_028(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 28, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_029(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 29, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_030(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 30, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_031(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 31, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_032(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 32, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_033(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 33, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_034(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 34, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_035(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 35, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_036(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 36, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_037(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 37, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_038(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 38, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_039(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 39, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_040(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 40, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_041(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 41, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_042(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 42, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_043(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 43, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_044(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 44, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_045(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 45, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_046(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 46, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_047(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 47, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_048(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 48, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_049(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 49, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_050(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 50, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_051(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 51, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_052(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 52, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_053(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 53, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_054(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 54, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_055(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 55, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_056(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 56, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_057(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 57, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_058(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 58, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_059(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 59, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_060(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 60, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_061(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 61, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_062(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 62, 1)
}

func TestProduct_TsPrinterExpressionsEmbedded_063(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 63, 1)
}

func TestProduct_TSPrinterGoOracle(test *testing.T) { test.Parallel(); PrinterGateGoOracle(test) }
