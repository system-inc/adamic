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
func TestProduct_GraphQLPrinterGoOracle(test *testing.T) { test.Parallel(); Printer(test) }
func TestProduct_GraphQlPrinterDefaults(test *testing.T) { test.Parallel(); Cases(test, "defaults") }
func TestProduct_GraphQlPrinterNarrow(test *testing.T)   { test.Parallel(); Cases(test, "narrow") }
func TestProduct_GraphQlPrinterTight(test *testing.T)    { test.Parallel(); Cases(test, "tight") }
func TestProduct_GraphQlPrinterTabs(test *testing.T)     { test.Parallel(); Cases(test, "tabs") }
