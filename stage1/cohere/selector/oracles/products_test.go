package oracles

import (
	"os"
	"testing"
)

func TestMain(suite *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	EnableProduction()
	os.Exit(suite.Run())
}
func TestProduct_SelectorGoOracle(test *testing.T) { test.Parallel(); Parser(test) }

func TestProduct_SelectorGoAnswers(test *testing.T) { test.Parallel(); Cases(test) }
