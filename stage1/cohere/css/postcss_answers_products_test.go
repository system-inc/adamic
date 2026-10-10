package css

import (
	"os"
	"path/filepath"
	"testing"
)

// cssParserPostCSSAnswersProducts is how many product tests share the agreement units' PostCSS answers, one node run
// per unit (388 units on Oct 10), so each product test stays well inside a product's 600 s.
const cssParserPostCSSAnswersProducts = 8

// cssParserPostCSSAnswers builds the PostCSS answers that every agreement unit whose ordinal is part modulo
// cssParserPostCSSAnswersProducts reads, each keyed as that unit asks for it: its own cases, the pinned library
// (ADAMIC_CSS_LIBRARY) and testdata/library.mjs. Without the library the units skip, and so does this.
func cssParserPostCSSAnswers(t *testing.T, part int) {
	t.Helper()
	library := os.Getenv("ADAMIC_CSS_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_CSS_LIBRARY to the pinned npm scratch directory")
	}
	script, err := filepath.Abs("testdata/library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	identity := cssParserOracleIdentity(t, library)
	prepareCSSParserTopCorpus(t)
	corpus := cssParserTopCorpus(t)
	ordinal, built := 0, 0
	for _, root := range cssParserShards(corpus.keys) {
		for _, unit := range root.groups {
			if unit.kind != "agreement" {
				continue
			}
			if ordinal%cssParserPostCSSAnswersProducts == part {
				cases, _ := writeCSSParserSlice(t, corpus, unit.cases)
				cachedCSSParserPostCSSAnswers(t, script, library, cases, identity)
				built++
			}
			ordinal++
		}
	}
	if built == 0 {
		t.Fatalf("part %d of %d holds no agreement unit, of %d", part, cssParserPostCSSAnswersProducts, ordinal)
	}
	t.Logf("PostCSS answers for %d of %d agreement units", built, ordinal)
}

func TestProduct_CSSParserPostCSSAnswers_0(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 0) }
func TestProduct_CSSParserPostCSSAnswers_1(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 1) }
func TestProduct_CSSParserPostCSSAnswers_2(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 2) }
func TestProduct_CSSParserPostCSSAnswers_3(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 3) }
func TestProduct_CSSParserPostCSSAnswers_4(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 4) }
func TestProduct_CSSParserPostCSSAnswers_5(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 5) }
func TestProduct_CSSParserPostCSSAnswers_6(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 6) }
func TestProduct_CSSParserPostCSSAnswers_7(t *testing.T) { t.Parallel(); cssParserPostCSSAnswers(t, 7) }
