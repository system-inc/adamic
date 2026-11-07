package oracles

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	EnableProduction()
	os.Exit(m.Run())
}
func TestProduct_GitignoreOracle(test *testing.T) { test.Parallel(); Binary(test) }
