package css

import (
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// These finders put the ordinary tests' oracle reads in this package's product
// list. The oracles package owns production; product-order waits for those owners.
func TestProduct_CssFindsParserPostCssAnswers_0(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 0)
}

func TestProduct_CssFindsParserPostCssAnswers_1(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 1)
}

func TestProduct_CssFindsParserPostCssAnswers_2(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 2)
}

func TestProduct_CssFindsParserPostCssAnswers_3(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 3)
}

func TestProduct_CssFindsParserPostCssAnswers_4(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 4)
}

func TestProduct_CssFindsParserPostCssAnswers_5(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 5)
}

func TestProduct_CssFindsParserPostCssAnswers_6(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 6)
}

func TestProduct_CssFindsParserPostCssAnswers_7(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 7)
}

func TestProduct_CssFindsPrinterParserOracle(test *testing.T) {
	test.Parallel()
	oracles.ComparisonOracle(test, false)
}

func TestProduct_CssFindsPrinterOracle(test *testing.T) {
	test.Parallel()
	oracles.ComparisonOracle(test, true)
}

func TestProduct_CssFindsParserGoOracleAnswers(test *testing.T) {
	test.Parallel()
	oracles.CachedParserOutputs(test)
}

func TestProduct_CssFindsParserCorpus(test *testing.T) {
	test.Parallel()
	oracles.LoadParserCorpus(test)
}

func TestProduct_CssFindsCompositionCorpusOracle(test *testing.T) {
	test.Parallel()
	oracles.CompositionOracle(
		test,
		"css-composition-corpus-oracle",
		"testdata/composition_corpus_side_test.go",
		"internal/format/css/postcss/adamic_port_side_test.go",
		"./internal/format/css/postcss",
	)
}

func TestProduct_CssFindsCompositionGoOracle(test *testing.T) {
	test.Parallel()
	oracles.CompositionOracle(
		test,
		"css-composition-go-oracle",
		"testdata/compose_side_test.go",
		"internal/format/css/adamic_compose_side_test.go",
		"./internal/format/css",
	)
}
