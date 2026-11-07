package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestCEndsInNewline(test *testing.T) {
	test.Parallel()
	program := &ir.Program{Source: "newline fixture"}
	for _, large := range []bool{false, true} {
		if large {
			for index := 0; index < 6000; index++ {
				program.Main = append(program.Main, ir.Evaluate{Value: ir.ArrayReverse{
					Array: ir.ArrayLiteral{Element: ir.Number, Elements: []ir.Expression{ir.NumberConstant{Value: 1}}},
				}})
			}
		}
		source := C(program)
		if !strings.HasSuffix(source, "\n") {
			test.Fatal("generated translation unit lacks a final newline")
		}
		if large && len(source) <= 256<<10 {
			test.Fatal("large fixture did not exceed runner's old cap")
		}
	}
}
