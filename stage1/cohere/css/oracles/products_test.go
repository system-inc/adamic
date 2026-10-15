package oracles

import "testing"

func TestProduct_CSSParserPostCSSAnswers_0(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 0) }

func TestProduct_CSSParserPostCSSAnswers_1(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 1) }

func TestProduct_CSSParserPostCSSAnswers_2(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 2) }

func TestProduct_CSSParserPostCSSAnswers_3(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 3) }

func TestProduct_CSSParserPostCSSAnswers_4(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 4) }

func TestProduct_CSSParserPostCSSAnswers_5(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 5) }

func TestProduct_CSSParserPostCSSAnswers_6(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 6) }

func TestProduct_CSSParserPostCSSAnswers_7(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 7) }

func TestProduct_CSSPrinterParserOracle(t *testing.T) { t.Parallel(); ComparisonOracle(t, false) }

func TestProduct_CSSPrinterOracle(t *testing.T) { t.Parallel(); ComparisonOracle(t, true) }

func TestProduct_CSSParserGoOracle(t *testing.T) { t.Parallel(); cssParserOracle(t) }

func TestProduct_CSSParserCorpus(t *testing.T) { t.Parallel(); prepareCSSParserTopCorpus(t) }

func TestProduct_CSSCompositionCorpusOracle(t *testing.T) {
	t.Parallel()
	CompositionOracle(t, "css-composition-corpus-oracle", "testdata/composition_corpus_side_test.go", "internal/format/css/postcss/adamic_port_side_test.go", "./internal/format/css/postcss")
}

func TestProduct_CSSCompositionGoOracle(t *testing.T) {
	t.Parallel()
	CompositionOracle(t, "css-composition-go-oracle", "testdata/compose_side_test.go", "internal/format/css/adamic_compose_side_test.go", "./internal/format/css")
}
