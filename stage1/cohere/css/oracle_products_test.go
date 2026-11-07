package css

import (
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// These finders put the ordinary tests' oracle reads in this package's product
// list. The oracles package owns production; product-order waits for those owners.
func TestProduct_CSSFindsParserPostCSSAnswers_0(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 0)
}

func TestProduct_CSSFindsParserPostCSSAnswers_1(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 1)
}

func TestProduct_CSSFindsParserPostCSSAnswers_2(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 2)
}

func TestProduct_CSSFindsParserPostCSSAnswers_3(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 3)
}

func TestProduct_CSSFindsParserPostCSSAnswers_4(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 4)
}

func TestProduct_CSSFindsParserPostCSSAnswers_5(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 5)
}

func TestProduct_CSSFindsParserPostCSSAnswers_6(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 6)
}

func TestProduct_CSSFindsParserPostCSSAnswers_7(test *testing.T) {
	test.Parallel()
	oracles.ParserPostCSSAnswers(test, 7)
}

func TestProduct_CSSFindsPrinterParserOracle(test *testing.T) {
	test.Parallel()
	oracles.ComparisonOracle(test, false)
}

func TestProduct_CSSFindsPrinterOracle(test *testing.T) {
	test.Parallel()
	oracles.ComparisonOracle(test, true)
}

func TestProduct_CSSFindsParserGoOracleAnswers(test *testing.T) {
	test.Parallel()
	oracles.CachedParserOutputs(test)
}

func TestProduct_CSSFindsParserCorpus(test *testing.T) {
	test.Parallel()
	oracles.LoadParserCorpus(test)
}

func TestProduct_CSSFindsCompositionCorpusOracle(test *testing.T) {
	test.Parallel()
	oracles.CompositionOracle(
		test,
		"css-composition-corpus-oracle",
		"testdata/composition_corpus_side_test.go",
		"internal/format/css/postcss/adamic_port_side_test.go",
		"./internal/format/css/postcss",
	)
}

func TestProduct_CSSFindsCompositionGoOracle(test *testing.T) {
	test.Parallel()
	oracles.CompositionOracle(
		test,
		"css-composition-go-oracle",
		"testdata/compose_side_test.go",
		"internal/format/css/adamic_compose_side_test.go",
		"./internal/format/css",
	)
}
