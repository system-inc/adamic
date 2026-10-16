package css

import (
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// These finders put the ordinary tests' oracle reads in this package's product
// list. The oracles package owns production; product-order waits for those owners.
func TestProduct_CSSFindsParserPostCSSAnswers_0(t *testing.T) {
	t.Parallel()
	oracles.ParserPostCSSAnswers(t, 0)
}

func TestProduct_CSSFindsParserPostCSSAnswers_1(t *testing.T) {
	t.Parallel()
	oracles.ParserPostCSSAnswers(t, 1)
}

func TestProduct_CSSFindsParserPostCSSAnswers_2(t *testing.T) {
	t.Parallel()
	oracles.ParserPostCSSAnswers(t, 2)
}

func TestProduct_CSSFindsParserPostCSSAnswers_3(t *testing.T) {
	t.Parallel()
	oracles.ParserPostCSSAnswers(t, 3)
}

func TestProduct_CSSFindsParserPostCSSAnswers_4(t *testing.T) {
	t.Parallel()
	oracles.ParserPostCSSAnswers(t, 4)
}

func TestProduct_CSSFindsParserPostCSSAnswers_5(t *testing.T) {
	t.Parallel()
	oracles.ParserPostCSSAnswers(t, 5)
}

func TestProduct_CSSFindsParserPostCSSAnswers_6(t *testing.T) {
	t.Parallel()
	oracles.ParserPostCSSAnswers(t, 6)
}

func TestProduct_CSSFindsParserPostCSSAnswers_7(t *testing.T) {
	t.Parallel()
	oracles.ParserPostCSSAnswers(t, 7)
}

func TestProduct_CSSFindsPrinterParserOracle(t *testing.T) {
	t.Parallel()
	oracles.ComparisonOracle(t, false)
}

func TestProduct_CSSFindsPrinterOracle(t *testing.T) { t.Parallel(); oracles.ComparisonOracle(t, true) }

func TestProduct_CSSFindsParserGoOracleAnswers(t *testing.T) {
	t.Parallel()
	oracles.CachedParserOutputs(t)
}

func TestProduct_CSSFindsParserCorpus(t *testing.T) { t.Parallel(); oracles.LoadParserCorpus(t) }

func TestProduct_CSSFindsCompositionCorpusOracle(t *testing.T) {
	t.Parallel()
	oracles.CompositionOracle(t, "css-composition-corpus-oracle", "testdata/composition_corpus_side_test.go", "internal/format/css/postcss/adamic_port_side_test.go", "./internal/format/css/postcss")
}

func TestProduct_CSSFindsCompositionGoOracle(t *testing.T) {
	t.Parallel()
	oracles.CompositionOracle(t, "css-composition-go-oracle", "testdata/compose_side_test.go", "internal/format/css/adamic_compose_side_test.go", "./internal/format/css")
}
