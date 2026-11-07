package oracles

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	EnableProduction()
	os.Exit(m.Run())
}
func TestProduct_SelectorGoOracle(test *testing.T) { test.Parallel(); Parser(test) }

func TestProduct_SelectorGoAnswers(test *testing.T) { test.Parallel(); Cases(test) }
