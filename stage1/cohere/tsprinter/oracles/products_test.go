package oracles

import "testing"

func TestProduct_TSPrinterPinnedCorpus(test *testing.T) {
	test.Parallel()
	PinnedCorpusFiles(test)
}

// The Go doc oracle's corpus, an oracle product (#g7a9yd3).
func TestProduct_TSPrinterDocumentCorpus(test *testing.T) {
	test.Parallel()
	DocumentCorpus(test)
}

func TestProduct_TSPrinterDocumentOracle(test *testing.T) {
	test.Parallel()
	DocumentOracle(test)
}

func TestProduct_TSPrinterMutantOracle_000(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 0)
}

func TestProduct_TSPrinterMutantOracle_001(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 1)
}

func TestProduct_TSPrinterMutantOracle_002(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 2)
}

func TestProduct_TSPrinterMutantOracle_003(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 3)
}

func TestProduct_TSPrinterMutantOracle_004(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 4)
}

func TestProduct_TSPrinterMutantOracle_005(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 5)
}

func TestProduct_TSPrinterMutantOracle_006(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 6)
}

func TestProduct_TSPrinterMutantOracle_007(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 7)
}

func TestProduct_TSPrinterMutantOracle_008(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 8)
}

func TestProduct_TSPrinterMutantOracle_009(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 9)
}

func TestProduct_TSPrinterMutantOracle_010(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 10)
}

func TestProduct_TSPrinterMutantOracle_011(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 11)
}

func TestProduct_TSPrinterMutantOracle_012(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 12)
}

func TestProduct_TSPrinterMutantOracle_013(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 13)
}

func TestProduct_TSPrinterMutantOracle_014(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 14)
}

func TestProduct_TSPrinterMutantOracle_015(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 15)
}

func TestProduct_TSPrinterMutantOracle_016(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 16)
}

func TestProduct_TSPrinterMutantOracle_017(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 17)
}

func TestProduct_TSPrinterMutantOracle_018(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 18)
}

func TestProduct_TSPrinterMutantOracle_019(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 19)
}

func TestProduct_TSPrinterMutantOracle_020(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 20)
}

func TestProduct_TSPrinterMutantOracle_021(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 21)
}

func TestProduct_TSPrinterMutantOracle_022(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 22)
}

func TestProduct_TSPrinterMutantOracle_023(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 23)
}

func TestProduct_TSPrinterMutantOracle_024(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 24)
}

func TestProduct_TSPrinterMutantOracle_025(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 25)
}

func TestProduct_TSPrinterMutantOracle_026(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 26)
}

func TestProduct_TSPrinterMutantOracle_027(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 27)
}

func TestProduct_TSPrinterMutantOracle_028(test *testing.T) {
	test.Parallel()
	MutantOracleProduct(test, 28)
}

func TestProduct_TSPrinterExpressionsCorpus(test *testing.T) {
	test.Parallel()
	ExpressionCorpus(test, false)
}

func TestProduct_TSPrinterStatementsCorpus(test *testing.T) { test.Parallel(); StatementCorpus(test) }

func TestProduct_TSPrinterTSCExpressionsCorpus(test *testing.T) {
	test.Parallel()
	printerGateTSCCorpus(test, "expressions")
}

func TestProduct_TSPrinterTSCStatementsCorpus(test *testing.T) {
	test.Parallel()
	printerGateTSCCorpus(test, "statements")
}

func TestProduct_TSPrinterTSCManifest(test *testing.T) { test.Parallel(); TSCAgreementPrepare(test) }

func TestProduct_TSPrinterStatementsNPM_000(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 0, 0)
}

func TestProduct_TSPrinterStatementsNPM_001(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 1, 0)
}

func TestProduct_TSPrinterStatementsNPM_002(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 2, 0)
}

func TestProduct_TSPrinterStatementsNPM_003(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 3, 0)
}

func TestProduct_TSPrinterStatementsNPM_004(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 4, 0)
}

func TestProduct_TSPrinterStatementsNPM_005(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 5, 0)
}

func TestProduct_TSPrinterStatementsNPM_006(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 6, 0)
}

func TestProduct_TSPrinterStatementsNPM_007(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 7, 0)
}

func TestProduct_TSPrinterStatementsNPM_008(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 8, 0)
}

func TestProduct_TSPrinterStatementsNPM_009(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 9, 0)
}

func TestProduct_TSPrinterStatementsNPM_010(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 10, 0)
}

func TestProduct_TSPrinterStatementsNPM_011(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 11, 0)
}

func TestProduct_TSPrinterStatementsNPM_012(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 12, 0)
}

func TestProduct_TSPrinterStatementsNPM_013(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 13, 0)
}

func TestProduct_TSPrinterStatementsNPM_014(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 14, 0)
}

func TestProduct_TSPrinterStatementsNPM_015(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 15, 0)
}

func TestProduct_TSPrinterStatementsEmbedded_000(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 0, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_001(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 1, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_002(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 2, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_003(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 3, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_004(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 4, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_005(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 5, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_006(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 6, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_007(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 7, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_008(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 8, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_009(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 9, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_010(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 10, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_011(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 11, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_012(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 12, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_013(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 13, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_014(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 14, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_015(test *testing.T) {
	test.Parallel()
	StatementDeclaredLibraryOracle(test, 15, 1)
}

func TestProduct_TSPrinterExpressionsNPM_000(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 0, 0)
}

func TestProduct_TSPrinterExpressionsNPM_001(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 1, 0)
}

func TestProduct_TSPrinterExpressionsNPM_002(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 2, 0)
}

func TestProduct_TSPrinterExpressionsNPM_003(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 3, 0)
}

func TestProduct_TSPrinterExpressionsNPM_004(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 4, 0)
}

func TestProduct_TSPrinterExpressionsNPM_005(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 5, 0)
}

func TestProduct_TSPrinterExpressionsNPM_006(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 6, 0)
}

func TestProduct_TSPrinterExpressionsNPM_007(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 7, 0)
}

func TestProduct_TSPrinterExpressionsNPM_008(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 8, 0)
}

func TestProduct_TSPrinterExpressionsNPM_009(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 9, 0)
}

func TestProduct_TSPrinterExpressionsNPM_010(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 10, 0)
}

func TestProduct_TSPrinterExpressionsNPM_011(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 11, 0)
}

func TestProduct_TSPrinterExpressionsNPM_012(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 12, 0)
}

func TestProduct_TSPrinterExpressionsNPM_013(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 13, 0)
}

func TestProduct_TSPrinterExpressionsNPM_014(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 14, 0)
}

func TestProduct_TSPrinterExpressionsNPM_015(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 15, 0)
}

func TestProduct_TSPrinterExpressionsNPM_016(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 16, 0)
}

func TestProduct_TSPrinterExpressionsNPM_017(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 17, 0)
}

func TestProduct_TSPrinterExpressionsNPM_018(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 18, 0)
}

func TestProduct_TSPrinterExpressionsNPM_019(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 19, 0)
}

func TestProduct_TSPrinterExpressionsNPM_020(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 20, 0)
}

func TestProduct_TSPrinterExpressionsNPM_021(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 21, 0)
}

func TestProduct_TSPrinterExpressionsNPM_022(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 22, 0)
}

func TestProduct_TSPrinterExpressionsNPM_023(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 23, 0)
}

func TestProduct_TSPrinterExpressionsNPM_024(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 24, 0)
}

func TestProduct_TSPrinterExpressionsNPM_025(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 25, 0)
}

func TestProduct_TSPrinterExpressionsNPM_026(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 26, 0)
}

func TestProduct_TSPrinterExpressionsNPM_027(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 27, 0)
}

func TestProduct_TSPrinterExpressionsNPM_028(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 28, 0)
}

func TestProduct_TSPrinterExpressionsNPM_029(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 29, 0)
}

func TestProduct_TSPrinterExpressionsNPM_030(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 30, 0)
}

func TestProduct_TSPrinterExpressionsNPM_031(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 31, 0)
}

func TestProduct_TSPrinterExpressionsNPM_032(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 32, 0)
}

func TestProduct_TSPrinterExpressionsNPM_033(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 33, 0)
}

func TestProduct_TSPrinterExpressionsNPM_034(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 34, 0)
}

func TestProduct_TSPrinterExpressionsNPM_035(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 35, 0)
}

func TestProduct_TSPrinterExpressionsNPM_036(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 36, 0)
}

func TestProduct_TSPrinterExpressionsNPM_037(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 37, 0)
}

func TestProduct_TSPrinterExpressionsNPM_038(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 38, 0)
}

func TestProduct_TSPrinterExpressionsNPM_039(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 39, 0)
}

func TestProduct_TSPrinterExpressionsNPM_040(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 40, 0)
}

func TestProduct_TSPrinterExpressionsNPM_041(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 41, 0)
}

func TestProduct_TSPrinterExpressionsNPM_042(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 42, 0)
}

func TestProduct_TSPrinterExpressionsNPM_043(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 43, 0)
}

func TestProduct_TSPrinterExpressionsNPM_044(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 44, 0)
}

func TestProduct_TSPrinterExpressionsNPM_045(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 45, 0)
}

func TestProduct_TSPrinterExpressionsNPM_046(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 46, 0)
}

func TestProduct_TSPrinterExpressionsNPM_047(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 47, 0)
}

func TestProduct_TSPrinterExpressionsNPM_048(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 48, 0)
}

func TestProduct_TSPrinterExpressionsNPM_049(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 49, 0)
}

func TestProduct_TSPrinterExpressionsNPM_050(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 50, 0)
}

func TestProduct_TSPrinterExpressionsNPM_051(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 51, 0)
}

func TestProduct_TSPrinterExpressionsNPM_052(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 52, 0)
}

func TestProduct_TSPrinterExpressionsNPM_053(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 53, 0)
}

func TestProduct_TSPrinterExpressionsNPM_054(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 54, 0)
}

func TestProduct_TSPrinterExpressionsNPM_055(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 55, 0)
}

func TestProduct_TSPrinterExpressionsNPM_056(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 56, 0)
}

func TestProduct_TSPrinterExpressionsNPM_057(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 57, 0)
}

func TestProduct_TSPrinterExpressionsNPM_058(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 58, 0)
}

func TestProduct_TSPrinterExpressionsNPM_059(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 59, 0)
}

func TestProduct_TSPrinterExpressionsNPM_060(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 60, 0)
}

func TestProduct_TSPrinterExpressionsNPM_061(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 61, 0)
}

func TestProduct_TSPrinterExpressionsNPM_062(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 62, 0)
}

func TestProduct_TSPrinterExpressionsNPM_063(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 63, 0)
}

func TestProduct_TSPrinterExpressionsEmbedded_000(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 0, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_001(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 1, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_002(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 2, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_003(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 3, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_004(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 4, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_005(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 5, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_006(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 6, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_007(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 7, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_008(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 8, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_009(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 9, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_010(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 10, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_011(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 11, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_012(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 12, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_013(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 13, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_014(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 14, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_015(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 15, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_016(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 16, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_017(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 17, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_018(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 18, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_019(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 19, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_020(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 20, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_021(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 21, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_022(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 22, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_023(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 23, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_024(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 24, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_025(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 25, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_026(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 26, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_027(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 27, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_028(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 28, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_029(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 29, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_030(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 30, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_031(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 31, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_032(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 32, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_033(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 33, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_034(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 34, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_035(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 35, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_036(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 36, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_037(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 37, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_038(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 38, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_039(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 39, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_040(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 40, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_041(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 41, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_042(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 42, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_043(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 43, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_044(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 44, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_045(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 45, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_046(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 46, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_047(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 47, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_048(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 48, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_049(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 49, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_050(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 50, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_051(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 51, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_052(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 52, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_053(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 53, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_054(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 54, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_055(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 55, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_056(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 56, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_057(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 57, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_058(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 58, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_059(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 59, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_060(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 60, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_061(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 61, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_062(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 62, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_063(test *testing.T) {
	test.Parallel()
	ExpressionDeclaredLibraryOracle(test, 63, 1)
}

func TestProduct_TSPrinterGoOracle(test *testing.T) { test.Parallel(); PrinterGateGoOracle(test) }
