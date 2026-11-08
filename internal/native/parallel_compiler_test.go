package native

import (
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Inspect both counted-reference ABI cases; the oracle executes the generated programs.
func TestParallelMapCompilerAbi(test *testing.T) {
	test.Parallel()
	for _, result := range []ir.Type{ir.Number, ir.String} {
		program := &ir.Program{Source: "parallel.a", Locals: []ir.Local{
			{Name: "items", Type: ir.Array, Global: true},
			{Name: "work", Type: ir.Closure, Global: true},
		}, Main: []ir.Statement{
			ir.Evaluate{Value: ir.ParallelMap{
				Items:  ir.Read{Local: 0, Of: ir.Array},
				Work:   ir.Read{Local: 1, Of: ir.Closure},
				Result: result,
			}},
		}}
		code := C(program)
		if !strings.Contains(code, "adamic_closure_call(work, arguments, 2)") {
			test.Fatal("adapter must call the original closure with its original self")
		}
		element := arrayElement(result)
		if !regexp.MustCompile(`adamic_parallel_map\([^,;\n]+, [^,;\n]+, ` + element + `\)`).MatchString(code) {
			test.Fatalf("runtime call missing static result element %s", element)
		}
		if strings.Contains(code, "->result_references") {
			test.Fatal("result ownership must be conveyed by the ABI argument")
		}
		declaration := "adamic_array *adamic_parallel_map(adamic_array *items, adamic_closure *work, enum adamic_ir_type element);"
		if strings.Count(code, declaration) != 1 || strings.Count(code, "adamic_parallel_map(") != 2 {
			test.Fatal("want one ABI declaration and one runtime call, without a stand-in definition")
		}
	}
}
