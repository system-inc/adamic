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
func TestProduct_GraphQLPrinterGoOracle(test *testing.T) { test.Parallel(); Printer(test) }
func TestProduct_GraphQLPrinterDefaults(test *testing.T) { test.Parallel(); Cases(test, "defaults") }
func TestProduct_GraphQLPrinterNarrow(test *testing.T)   { test.Parallel(); Cases(test, "narrow") }
func TestProduct_GraphQLPrinterTight(test *testing.T)    { test.Parallel(); Cases(test, "tight") }
func TestProduct_GraphQLPrinterTabs(test *testing.T)     { test.Parallel(); Cases(test, "tabs") }
