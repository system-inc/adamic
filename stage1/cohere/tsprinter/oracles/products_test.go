package oracles

import "testing"

func TestProduct_TSPrinterPinnedCorpus(t *testing.T) {
	t.Parallel()
	PinnedCorpusFiles(t)
}

// The Go doc oracle's corpus, an oracle product (#g7a9yd3).
func TestProduct_TSPrinterDocumentCorpus(t *testing.T) {
	t.Parallel()
	DocumentCorpus(t)
}

func TestProduct_TSPrinterDocumentOracle(t *testing.T) {
	t.Parallel()
	DocumentOracle(t)
}

func TestProduct_TSPrinterMutantOracle_000(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 0) }

func TestProduct_TSPrinterMutantOracle_001(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 1) }

func TestProduct_TSPrinterMutantOracle_002(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 2) }

func TestProduct_TSPrinterMutantOracle_003(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 3) }

func TestProduct_TSPrinterMutantOracle_004(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 4) }

func TestProduct_TSPrinterMutantOracle_005(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 5) }

func TestProduct_TSPrinterMutantOracle_006(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 6) }

func TestProduct_TSPrinterMutantOracle_007(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 7) }

func TestProduct_TSPrinterMutantOracle_008(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 8) }

func TestProduct_TSPrinterMutantOracle_009(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 9) }

func TestProduct_TSPrinterMutantOracle_010(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 10) }

func TestProduct_TSPrinterMutantOracle_011(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 11) }

func TestProduct_TSPrinterMutantOracle_012(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 12) }

func TestProduct_TSPrinterMutantOracle_013(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 13) }

func TestProduct_TSPrinterMutantOracle_014(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 14) }

func TestProduct_TSPrinterMutantOracle_015(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 15) }

func TestProduct_TSPrinterMutantOracle_016(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 16) }

func TestProduct_TSPrinterMutantOracle_017(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 17) }

func TestProduct_TSPrinterMutantOracle_018(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 18) }

func TestProduct_TSPrinterMutantOracle_019(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 19) }

func TestProduct_TSPrinterMutantOracle_020(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 20) }

func TestProduct_TSPrinterMutantOracle_021(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 21) }

func TestProduct_TSPrinterMutantOracle_022(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 22) }

func TestProduct_TSPrinterMutantOracle_023(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 23) }

func TestProduct_TSPrinterMutantOracle_024(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 24) }

func TestProduct_TSPrinterMutantOracle_025(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 25) }

func TestProduct_TSPrinterMutantOracle_026(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 26) }

func TestProduct_TSPrinterMutantOracle_027(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 27) }

func TestProduct_TSPrinterMutantOracle_028(t *testing.T) { t.Parallel(); MutantOracleProduct(t, 28) }

func TestProduct_TSPrinterExpressionsCorpus(t *testing.T) { t.Parallel(); ExpressionCorpus(t, false) }

func TestProduct_TSPrinterStatementsCorpus(t *testing.T) { t.Parallel(); StatementCorpus(t) }

func TestProduct_TSPrinterTSCExpressionsCorpus(t *testing.T) {
	t.Parallel()
	printerGateTSCCorpus(t, "expressions")
}

func TestProduct_TSPrinterTSCStatementsCorpus(t *testing.T) {
	t.Parallel()
	printerGateTSCCorpus(t, "statements")
}

func TestProduct_TSPrinterTSCManifest(t *testing.T) { t.Parallel(); TSCAgreementPrepare(t) }

func TestProduct_TSPrinterStatementsNPM_000(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 0, 0)
}

func TestProduct_TSPrinterStatementsNPM_001(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 1, 0)
}

func TestProduct_TSPrinterStatementsNPM_002(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 2, 0)
}

func TestProduct_TSPrinterStatementsNPM_003(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 3, 0)
}

func TestProduct_TSPrinterStatementsNPM_004(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 4, 0)
}

func TestProduct_TSPrinterStatementsNPM_005(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 5, 0)
}

func TestProduct_TSPrinterStatementsNPM_006(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 6, 0)
}

func TestProduct_TSPrinterStatementsNPM_007(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 7, 0)
}

func TestProduct_TSPrinterStatementsNPM_008(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 8, 0)
}

func TestProduct_TSPrinterStatementsNPM_009(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 9, 0)
}

func TestProduct_TSPrinterStatementsNPM_010(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 10, 0)
}

func TestProduct_TSPrinterStatementsNPM_011(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 11, 0)
}

func TestProduct_TSPrinterStatementsNPM_012(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 12, 0)
}

func TestProduct_TSPrinterStatementsNPM_013(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 13, 0)
}

func TestProduct_TSPrinterStatementsNPM_014(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 14, 0)
}

func TestProduct_TSPrinterStatementsNPM_015(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 15, 0)
}

func TestProduct_TSPrinterStatementsEmbedded_000(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 0, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_001(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 1, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_002(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 2, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_003(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 3, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_004(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 4, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_005(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 5, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_006(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 6, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_007(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 7, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_008(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 8, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_009(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 9, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_010(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 10, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_011(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 11, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_012(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 12, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_013(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 13, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_014(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 14, 1)
}

func TestProduct_TSPrinterStatementsEmbedded_015(t *testing.T) {
	t.Parallel()
	StatementDeclaredLibraryOracle(t, 15, 1)
}

func TestProduct_TSPrinterExpressionsNPM_000(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 0, 0)
}

func TestProduct_TSPrinterExpressionsNPM_001(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 1, 0)
}

func TestProduct_TSPrinterExpressionsNPM_002(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 2, 0)
}

func TestProduct_TSPrinterExpressionsNPM_003(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 3, 0)
}

func TestProduct_TSPrinterExpressionsNPM_004(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 4, 0)
}

func TestProduct_TSPrinterExpressionsNPM_005(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 5, 0)
}

func TestProduct_TSPrinterExpressionsNPM_006(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 6, 0)
}

func TestProduct_TSPrinterExpressionsNPM_007(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 7, 0)
}

func TestProduct_TSPrinterExpressionsNPM_008(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 8, 0)
}

func TestProduct_TSPrinterExpressionsNPM_009(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 9, 0)
}

func TestProduct_TSPrinterExpressionsNPM_010(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 10, 0)
}

func TestProduct_TSPrinterExpressionsNPM_011(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 11, 0)
}

func TestProduct_TSPrinterExpressionsNPM_012(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 12, 0)
}

func TestProduct_TSPrinterExpressionsNPM_013(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 13, 0)
}

func TestProduct_TSPrinterExpressionsNPM_014(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 14, 0)
}

func TestProduct_TSPrinterExpressionsNPM_015(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 15, 0)
}

func TestProduct_TSPrinterExpressionsNPM_016(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 16, 0)
}

func TestProduct_TSPrinterExpressionsNPM_017(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 17, 0)
}

func TestProduct_TSPrinterExpressionsNPM_018(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 18, 0)
}

func TestProduct_TSPrinterExpressionsNPM_019(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 19, 0)
}

func TestProduct_TSPrinterExpressionsNPM_020(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 20, 0)
}

func TestProduct_TSPrinterExpressionsNPM_021(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 21, 0)
}

func TestProduct_TSPrinterExpressionsNPM_022(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 22, 0)
}

func TestProduct_TSPrinterExpressionsNPM_023(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 23, 0)
}

func TestProduct_TSPrinterExpressionsNPM_024(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 24, 0)
}

func TestProduct_TSPrinterExpressionsNPM_025(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 25, 0)
}

func TestProduct_TSPrinterExpressionsNPM_026(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 26, 0)
}

func TestProduct_TSPrinterExpressionsNPM_027(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 27, 0)
}

func TestProduct_TSPrinterExpressionsNPM_028(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 28, 0)
}

func TestProduct_TSPrinterExpressionsNPM_029(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 29, 0)
}

func TestProduct_TSPrinterExpressionsNPM_030(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 30, 0)
}

func TestProduct_TSPrinterExpressionsNPM_031(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 31, 0)
}

func TestProduct_TSPrinterExpressionsNPM_032(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 32, 0)
}

func TestProduct_TSPrinterExpressionsNPM_033(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 33, 0)
}

func TestProduct_TSPrinterExpressionsNPM_034(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 34, 0)
}

func TestProduct_TSPrinterExpressionsNPM_035(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 35, 0)
}

func TestProduct_TSPrinterExpressionsNPM_036(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 36, 0)
}

func TestProduct_TSPrinterExpressionsNPM_037(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 37, 0)
}

func TestProduct_TSPrinterExpressionsNPM_038(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 38, 0)
}

func TestProduct_TSPrinterExpressionsNPM_039(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 39, 0)
}

func TestProduct_TSPrinterExpressionsNPM_040(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 40, 0)
}

func TestProduct_TSPrinterExpressionsNPM_041(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 41, 0)
}

func TestProduct_TSPrinterExpressionsNPM_042(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 42, 0)
}

func TestProduct_TSPrinterExpressionsNPM_043(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 43, 0)
}

func TestProduct_TSPrinterExpressionsNPM_044(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 44, 0)
}

func TestProduct_TSPrinterExpressionsNPM_045(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 45, 0)
}

func TestProduct_TSPrinterExpressionsNPM_046(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 46, 0)
}

func TestProduct_TSPrinterExpressionsNPM_047(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 47, 0)
}

func TestProduct_TSPrinterExpressionsNPM_048(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 48, 0)
}

func TestProduct_TSPrinterExpressionsNPM_049(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 49, 0)
}

func TestProduct_TSPrinterExpressionsNPM_050(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 50, 0)
}

func TestProduct_TSPrinterExpressionsNPM_051(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 51, 0)
}

func TestProduct_TSPrinterExpressionsNPM_052(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 52, 0)
}

func TestProduct_TSPrinterExpressionsNPM_053(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 53, 0)
}

func TestProduct_TSPrinterExpressionsNPM_054(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 54, 0)
}

func TestProduct_TSPrinterExpressionsNPM_055(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 55, 0)
}

func TestProduct_TSPrinterExpressionsNPM_056(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 56, 0)
}

func TestProduct_TSPrinterExpressionsNPM_057(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 57, 0)
}

func TestProduct_TSPrinterExpressionsNPM_058(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 58, 0)
}

func TestProduct_TSPrinterExpressionsNPM_059(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 59, 0)
}

func TestProduct_TSPrinterExpressionsNPM_060(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 60, 0)
}

func TestProduct_TSPrinterExpressionsNPM_061(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 61, 0)
}

func TestProduct_TSPrinterExpressionsNPM_062(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 62, 0)
}

func TestProduct_TSPrinterExpressionsNPM_063(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 63, 0)
}

func TestProduct_TSPrinterExpressionsEmbedded_000(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 0, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_001(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 1, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_002(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 2, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_003(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 3, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_004(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 4, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_005(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 5, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_006(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 6, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_007(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 7, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_008(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 8, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_009(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 9, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_010(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 10, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_011(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 11, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_012(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 12, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_013(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 13, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_014(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 14, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_015(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 15, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_016(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 16, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_017(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 17, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_018(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 18, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_019(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 19, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_020(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 20, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_021(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 21, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_022(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 22, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_023(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 23, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_024(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 24, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_025(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 25, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_026(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 26, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_027(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 27, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_028(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 28, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_029(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 29, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_030(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 30, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_031(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 31, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_032(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 32, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_033(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 33, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_034(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 34, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_035(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 35, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_036(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 36, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_037(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 37, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_038(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 38, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_039(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 39, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_040(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 40, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_041(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 41, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_042(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 42, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_043(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 43, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_044(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 44, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_045(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 45, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_046(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 46, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_047(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 47, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_048(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 48, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_049(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 49, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_050(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 50, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_051(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 51, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_052(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 52, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_053(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 53, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_054(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 54, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_055(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 55, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_056(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 56, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_057(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 57, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_058(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 58, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_059(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 59, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_060(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 60, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_061(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 61, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_062(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 62, 1)
}

func TestProduct_TSPrinterExpressionsEmbedded_063(t *testing.T) {
	t.Parallel()
	ExpressionDeclaredLibraryOracle(t, 63, 1)
}

func TestProduct_TSPrinterGoOracle(t *testing.T) { t.Parallel(); PrinterGateGoOracle(t) }
