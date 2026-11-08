package printer

import (
	"github.com/system-inc/adamic/stage1/cohere/graphql/printer/oracles"
	"testing"
)

func printerOracle(test *testing.T) string                    { return oracles.Printer(test) }
func TestProduct_GraphQlPrinterFindsGoOracle(test *testing.T) { test.Parallel(); printerOracle(test) }
