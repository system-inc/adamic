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
func TestProduct_GraphQLGoOracle(test *testing.T) { test.Parallel(); Parser(test) }
