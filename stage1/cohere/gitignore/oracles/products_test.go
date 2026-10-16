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
func TestProduct_GitignoreOracle(t *testing.T) { t.Parallel(); Binary(t) }
