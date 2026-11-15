package native

import (
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A conditional held as a number or a boolean with a bare undefined arm is a compiler bug said out
// loud, never NULL in a double (clang refuses it) or in a bool (it reads false, #qgthc6g). The same
// arm in a conditional held as a reference, and a maybe's absent arm, are the forms lowering makes,
// so they emit.
func TestConditionalRefusesUndefinedInAScalar(t *testing.T) {
	t.Parallel()
	emits := func(conditional ir.Conditional) (failure string) {
		defer func() {
			if recovered := recover(); recovered != nil {
				failure = fmt.Sprint(recovered)
			}
		}()
		e := &emitter{program: &ir.Program{}, scopes: [][]string{nil}}
		e.conditional(conditional)
		return ""
	}
	condition := ir.BooleanConstant{Value: true}
	for _, of := range []ir.Type{ir.Number, ir.Boolean} {
		present := ir.Expression(ir.NumberConstant{Value: 1})
		if of == ir.Boolean {
			present = ir.BooleanConstant{Value: true}
		}
		for _, conditional := range []ir.Conditional{
			{Condition: condition, WhenTrue: present, WhenNot: ir.Undefined{}, Of: of},
			{Condition: condition, WhenTrue: ir.Undefined{}, WhenNot: present, Of: of},
		} {
			if failure := emits(conditional); !strings.Contains(failure, "compiler bug: a conditional held as") {
				t.Fatalf("held as %s, an undefined arm emitted (%q), want the compiler bug", cType(of), failure)
			}
		}
	}
	for _, conditional := range []ir.Conditional{
		{Condition: condition, WhenTrue: ir.ArrayLiteral{Element: ir.Number}, WhenNot: ir.Undefined{Of: ir.Array}, Of: ir.Array},
		{Condition: condition, WhenTrue: ir.MaybeOf{Value: ir.NumberConstant{Value: 1}, Of: ir.MaybeNumber}, WhenNot: ir.MaybeOf{Of: ir.MaybeNumber}, Of: ir.MaybeNumber},
	} {
		if failure := emits(conditional); failure != "" {
			t.Fatalf("held as %s, want it emitted, got %q", cType(conditional.Of), failure)
		}
	}
}
