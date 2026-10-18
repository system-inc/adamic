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

func TestProduct_ScannerOracle(t *testing.T) { t.Parallel(); ScannerOracleProduct(t) }
