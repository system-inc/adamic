package oracles

import (
	"os"
	"testing"
)

func TestMain(suite *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	EnableProductProduction()
	os.Exit(suite.Run())
}
func TestProduct_YAMLFormatGo(test *testing.T)  { test.Parallel(); Formatter(test) }
func TestProduct_YAMLLexerGo(test *testing.T)   { test.Parallel(); Lexer(test) }
func TestProduct_YAMLCSTGo(test *testing.T)     { test.Parallel(); ConcreteSyntaxTree(test) }
func TestProduct_YAMLUnistGo(test *testing.T)   { test.Parallel(); Unist(test) }
func TestProduct_YAMLWidthGo(test *testing.T)   { test.Parallel(); Width(test) }
func TestProduct_YAMLSchemaGo(test *testing.T)  { test.Parallel(); Schema(test) }
func TestProduct_YAMLScalarsGo(test *testing.T) { test.Parallel(); Scalars(test) }
func TestProduct_YAMLPropsGo(test *testing.T)   { test.Parallel(); Properties(test) }
func TestProduct_YAMLComposeGo(test *testing.T) { test.Parallel(); Composition(test) }
func TestProduct_YAMLFormatterMutantsOracle(test *testing.T) {
	test.Parallel()
	FormatterMutantAnswers(test)
}
