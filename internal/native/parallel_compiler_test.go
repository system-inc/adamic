package native

import (
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Inspect both counted-reference ABI cases; the oracle executes the generated programs.
func TestParallelMapCompilerABI(t *testing.T) {
	t.Parallel()
	for _, result := range []ir.Type{ir.Number, ir.String} {
		program := &ir.Program{Source: "parallel.a", Locals: []ir.Local{{Name: "items", Type: ir.Array, Global: true}, {Name: "work", Type: ir.Closure, Global: true}}, Main: []ir.Statement{
			ir.Evaluate{Value: ir.ParallelMap{Items: ir.Read{Local: 0, Of: ir.Array}, Work: ir.Read{Local: 1, Of: ir.Closure}, Result: result}},
		}}
		code := C(program)
		if !strings.Contains(code, "adamic_closure_call(work, arguments, 2)") {
			t.Fatal("adapter must call the original closure with its original self")
		}
		flag := "false"
		if result.IsReference() {
			flag = "true"
		}
		if !regexp.MustCompile(`adamic_parallel_map\([^,;\n]+, [^,;\n]+, ` + flag + `\)`).MatchString(code) {
			t.Fatalf("runtime call missing static result reference flag %s", flag)
		}
		if strings.Contains(code, "->result_references") {
			t.Fatal("result ownership must be conveyed by the ABI argument")
		}
		declaration := "adamic_array *adamic_parallel_map(adamic_array *items, adamic_closure *work, bool references);"
		if strings.Count(code, declaration) != 1 || strings.Count(code, "adamic_parallel_map(") != 2 {
			t.Fatal("want one ABI declaration and one runtime call, without a stand-in definition")
		}
	}
}
