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
func TestProduct_YamlFormatGo(test *testing.T)  { test.Parallel(); Formatter(test) }
func TestProduct_YamlLexerGo(test *testing.T)   { test.Parallel(); Lexer(test) }
func TestProduct_YamlcstGo(test *testing.T)     { test.Parallel(); ConcreteSyntaxTree(test) }
func TestProduct_YamlUnistGo(test *testing.T)   { test.Parallel(); Unist(test) }
func TestProduct_YamlWidthGo(test *testing.T)   { test.Parallel(); Width(test) }
func TestProduct_YamlSchemaGo(test *testing.T)  { test.Parallel(); Schema(test) }
func TestProduct_YamlScalarsGo(test *testing.T) { test.Parallel(); Scalars(test) }
func TestProduct_YamlPropsGo(test *testing.T)   { test.Parallel(); Properties(test) }
func TestProduct_YamlComposeGo(test *testing.T) { test.Parallel(); Composition(test) }
func TestProduct_YamlFormatterMutantsOracle(test *testing.T) {
	test.Parallel()
	FormatterMutantAnswers(test)
}
