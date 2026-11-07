package oracles

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if chdirError := os.Chdir(".."); chdirError != nil {
		panic(chdirError)
	}
	producer = true
	os.Exit(m.Run())
}
func TestProduct_SuppressionGoOracle(test *testing.T) { test.Parallel(); Suppression(test) }
