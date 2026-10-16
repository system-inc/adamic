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
func TestProduct_GraphQLPrinterGoOracle(t *testing.T) { t.Parallel(); Printer(t) }
func TestProduct_GraphQLPrinterDefaults(t *testing.T) { t.Parallel(); Cases(t, "defaults") }
func TestProduct_GraphQLPrinterNarrow(t *testing.T)   { t.Parallel(); Cases(t, "narrow") }
func TestProduct_GraphQLPrinterTight(t *testing.T)    { t.Parallel(); Cases(t, "tight") }
func TestProduct_GraphQLPrinterTabs(t *testing.T)     { t.Parallel(); Cases(t, "tabs") }
