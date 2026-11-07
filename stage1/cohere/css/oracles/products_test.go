package oracles

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestProduct_CSSParserPostCSSAnswers_0(test *testing.T) {
	test.Parallel()
	ParserPostCSSAnswers(test, 0)
}

func TestProduct_CSSParserPostCSSAnswers_1(test *testing.T) {
	test.Parallel()
	ParserPostCSSAnswers(test, 1)
}

func TestProduct_CSSParserPostCSSAnswers_2(test *testing.T) {
	test.Parallel()
	ParserPostCSSAnswers(test, 2)
}

func TestProduct_CSSParserPostCSSAnswers_3(test *testing.T) {
	test.Parallel()
	ParserPostCSSAnswers(test, 3)
}

func TestProduct_CSSParserPostCSSAnswers_4(test *testing.T) {
	test.Parallel()
	ParserPostCSSAnswers(test, 4)
}

func TestProduct_CSSParserPostCSSAnswers_5(test *testing.T) {
	test.Parallel()
	ParserPostCSSAnswers(test, 5)
}

func TestProduct_CSSParserPostCSSAnswers_6(test *testing.T) {
	test.Parallel()
	ParserPostCSSAnswers(test, 6)
}

func TestProduct_CSSParserPostCSSAnswers_7(test *testing.T) {
	test.Parallel()
	ParserPostCSSAnswers(test, 7)
}

func TestProduct_CSSPrinterParserOracle(test *testing.T) {
	test.Parallel()
	requireComparisonOracle(test, ComparisonOracle(test, false))
}

func TestProduct_CSSPrinterOracle(test *testing.T) {
	test.Parallel()
	requireComparisonOracle(test, ComparisonOracle(test, true))
}

// requireComparisonOracle proves a Go comparison oracle exists and runs: the test binary lists no tests (#n1g5604).
func requireComparisonOracle(test *testing.T, binary string) {
	test.Helper()
	buildcache.RequireArtifacts(
		test,
		filepath.Dir(binary),
		buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"-test.list=^$"}},
	)
}

func TestProduct_CSSParserGoOracle(test *testing.T) { test.Parallel(); cssParserOracle(test) }

func TestProduct_CSSParserCorpus(test *testing.T) { test.Parallel(); prepareCSSParserTopCorpus(test) }

func TestProduct_CSSCompositionCorpusOracle(test *testing.T) {
	test.Parallel()
	CompositionOracle(
		test,
		"css-composition-corpus-oracle",
		"testdata/composition_corpus_side_test.go",
		"internal/format/css/postcss/adamic_port_side_test.go",
		"./internal/format/css/postcss",
	)
}

func TestProduct_CSSCompositionGoOracle(test *testing.T) {
	test.Parallel()
	CompositionOracle(
		test,
		"css-composition-go-oracle",
		"testdata/compose_side_test.go",
		"internal/format/css/adamic_compose_side_test.go",
		"./internal/format/css",
	)
}
