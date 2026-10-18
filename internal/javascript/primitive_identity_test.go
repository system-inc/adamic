package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestJavaScriptObjectOperandAdapterIdentity(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Locals: []ir.Local{{Name: "left", Type: ir.Object}, {Name: "right", Type: ir.Object}}}
	e := &emitter{program: program}
	left, right := ir.Read{Local: 0, Of: ir.Object}, ir.Read{Local: 1, Of: ir.Object}
	equal := e.value(ir.Binary{Operator: ir.Equal, Left: left, Right: right})
	same := e.value(ir.ObjectCall{Method: "is", Arguments: []ir.Expression{left, right}, Returns: ir.Boolean})
	setup := viewCallableAdapterRuntime + "const " + e.variable(0) + " = {}; const " + e.variable(1) + " = {}; adamicViewAdapterRoots.set(" + e.variable(0) + ", " + e.variable(1) + ");\n"
	code := "console.log(" + equal + ", " + same + ");"
	runViewNode(t, setup+code, "true true\n", "", 0)
	if strings.Count(equal, "adamicViewAdapterUnderlying(") != 2 || strings.Count(same, "adamicViewAdapterUnderlying(") != 2 {
		t.Fatal("object comparisons lost normalization", equal, same)
	}
	// Omit lookups at object-typed operand sites only, retaining the root map.
	mutant := strings.ReplaceAll(code, "adamicViewAdapterUnderlying(", "(")
	runViewNode(t, setup+mutant, "false false\n", "", 0)
	t.Log("object-typed lookup omission caught: expected true true, mutant false false")
}

func TestJavaScriptPrimitiveIdentityRepresentations(t *testing.T) {
	t.Parallel()
	e := &emitter{program: &ir.Program{Locals: []ir.Local{{Name: "value"}}}}
	for _, of := range []ir.Type{ir.Number, ir.String, ir.Boolean, ir.MaybeNumber, ir.MaybeBoolean} {
		if strings.Contains(e.identityOperand(ir.Read{Local: 0, Of: of}, false), "adamicViewAdapterUnderlying") {
			t.Fatalf("primitive representation %v normalizes", of)
		}
	}
	for _, of := range []ir.Type{ir.Object, ir.Closure, ir.Union, ir.Array} {
		if !strings.Contains(e.identityOperand(ir.Read{Local: 0, Of: of}, false), "adamicViewAdapterUnderlying") {
			t.Fatalf("object-capable representation %v bypasses", of)
		}
	}
	if strings.Contains(e.identityOperand(ir.Read{Local: 0, Of: ir.Union}, true), "adamicViewAdapterUnderlying") {
		t.Fatal("checker primitive-union proof ignored")
	}
}
