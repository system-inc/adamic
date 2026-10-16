package oracles

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	EnableProductProduction()
	os.Exit(m.Run())
}
func TestProduct_YAMLFormatGo(t *testing.T)               { t.Parallel(); Formatter(t) }
func TestProduct_YAMLLexerGo(t *testing.T)                { t.Parallel(); Lexer(t) }
func TestProduct_YAMLCSTGo(t *testing.T)                  { t.Parallel(); ConcreteSyntaxTree(t) }
func TestProduct_YAMLUnistGo(t *testing.T)                { t.Parallel(); Unist(t) }
func TestProduct_YAMLWidthGo(t *testing.T)                { t.Parallel(); Width(t) }
func TestProduct_YAMLSchemaGo(t *testing.T)               { t.Parallel(); Schema(t) }
func TestProduct_YAMLScalarsGo(t *testing.T)              { t.Parallel(); Scalars(t) }
func TestProduct_YAMLPropsGo(t *testing.T)                { t.Parallel(); Properties(t) }
func TestProduct_YAMLComposeGo(t *testing.T)              { t.Parallel(); Composition(t) }
func TestProduct_YAMLFormatterMutantsOracle(t *testing.T) { t.Parallel(); FormatterMutantAnswers(t) }
