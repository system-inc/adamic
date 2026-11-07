package oracles

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	producer = true
	os.Exit(m.Run())
}

func TestProduct_ScannerOracle(test *testing.T) { test.Parallel(); ScannerOracleProduct(test) }
