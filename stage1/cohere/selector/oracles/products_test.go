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
func TestProduct_SelectorGoOracle(t *testing.T) { t.Parallel(); Parser(t) }

func TestProduct_SelectorGoAnswers(t *testing.T) { t.Parallel(); Cases(t) }
