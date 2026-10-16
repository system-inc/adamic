package printer

import (
	"github.com/system-inc/adamic/stage1/cohere/graphql/printer/oracles"
	"testing"
)

func printerOracle(t *testing.T) string                    { return oracles.Printer(t) }
func TestProduct_GraphQLPrinterFindsGoOracle(t *testing.T) { t.Parallel(); printerOracle(t) }
