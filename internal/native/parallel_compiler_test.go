package native

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// This inspects generated code without linking an unavailable runtime ABI.
func TestParallelMapCompilerABI(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Source: "parallel.a", Locals: []ir.Local{{Name: "items", Type: ir.Array, Global: true}, {Name: "work", Type: ir.Closure, Global: true}}, Main: []ir.Statement{
		ir.Evaluate{Value: ir.ParallelMap{Items: ir.Read{Local: 0, Of: ir.Array}, Work: ir.Read{Local: 1, Of: ir.Closure}}},
	}}
	code := C(program)
	if strings.Count(code, "adamic_parallel_map(") != 1 {
		t.Fatal("want exactly one runtime call")
	}
	if strings.Contains(code, "adamic_parallel_map(adamic_array") {
		t.Fatal("compiler emitted a stand-in ABI definition")
	}
}
