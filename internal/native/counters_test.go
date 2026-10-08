package native

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A temporary or lookup added to one body must not renumber a later body or main.
func TestFunctionCountersDoNotMove(t *testing.T) {
	t.Parallel()
	property := ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "value", Of: ir.Number}
	program := &ir.Program{
		Source: "counters.a",
		Locals: []ir.Local{{Name: "object", Type: ir.Object, Function: 0}, {Name: "object", Type: ir.Object, Function: 1}},
		Functions: []ir.Function{
			{Name: "first", Parameters: []int{0}, Returns: ir.Number, Body: []ir.Statement{ir.Return{Value: property}}},
			{Name: "second", Parameters: []int{1}, Returns: ir.Number, Body: []ir.Statement{ir.Return{Value: ir.Property{Object: ir.Read{Local: 1, Of: ir.Object}, Name: "value", Of: ir.Number}}}},
		},
	}
	program.Main = []ir.Statement{ir.Evaluate{Value: ir.Length{Array: ir.ArrayLiteral{Element: ir.Number, Elements: []ir.Expression{ir.NumberConstant{Value: 1}}}}}}
	original := C(program)
	body := func(source, name string) string {
		start := strings.Index(source, "static double "+name+"(")
		if start < 0 {
			t.Fatalf("missing signature %s", name)
		}
		// Skip the prototype.
		start += strings.Index(source[start:], ";") + 1
		start += strings.Index(source[start:], "static double "+name+"(")
		end := strings.Index(source[start:], "\n}\n")
		if end < 0 {
			t.Fatalf("missing body %s", name)
		}
		return source[start : start+end+3]
	}
	name := (&emitter{program: program}).functionName(1)
	second := body(original, name)
	main := original[strings.Index(original, "\nint main("):]
	for _, added := range []ir.Expression{ir.Length{Array: ir.ArrayLiteral{Element: ir.Number, Elements: []ir.Expression{ir.NumberConstant{Value: 14}}}}, property} {
		program.Functions[0].Body = []ir.Statement{ir.Evaluate{Value: added}, ir.Return{Value: property}}
		changed := C(program)
		if got := body(changed, name); got != second {
			t.Fatalf("adding %T moved second body\nbefore:\n%s\nafter:\n%s", added, second, got)
		}
		if got := changed[strings.Index(changed, "\nint main("):]; got != main {
			t.Fatalf("adding %T moved main\nbefore:\n%s\nafter:\n%s", added, main, got)
		}
		cache := "static adamic_slot_cache " + stableName("adamic_cache", "slot", name) + "_1;"
		if !strings.Contains(changed, cache) {
			t.Fatalf("later function cache moved: missing %s", cache)
		}
	}
}
