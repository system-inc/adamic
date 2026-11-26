package lower

import (
	"errors"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// structuralViewCast never answers nil without an error (#bgm34gh). castRoadValue takes what it returns as the
// cast's value, so a nil there went on as one, and expression's noteViewArrayStorage dereferenced it: a compiler
// panic over all of tsc in first-link trap mode, at utilities.ts:5163's `node.parent as Declaration`, whose
// operand was held as a Union. No small program reproduces that representation (J-ybnhz24, J-n4czx8s: every
// cut-down shape holds the operand as an Object and lowers), so this judges the cast on a kept lowering directly,
// as graph_view_guard_test does for pairs no source program makes. Derived is a structural subtype of Base, the
// pair castProof answers structuralView for. Mutant (structural_view_cast.mutants.md): the Union arm's NotYet
// made `return nil, nil` again.
func TestStructuralViewCastNeverAnswersNil(t *testing.T) {
	t.Parallel()
	lowered, types, nodes := graphViewGuardLowering(t)
	source, target, node := types["Base"], types["Derived"], nodes["Derived"]
	if source == nil || target == nil || node == nil {
		t.Fatal("Base or Derived isn't declared")
	}
	for _, row := range []struct {
		name  string
		value ir.Expression
		want  string
	}{
		{"a value held as a Union", ir.Read{Of: ir.Union}, "a checked cast from Base to Derived whose value isn't a plain object"},
		{"a value held as a number", ir.Read{Of: ir.Number}, "a checked cast from Base to Derived whose value isn't a plain object"},
	} {
		t.Run(row.name, func(t *testing.T) {
			got, err := lowered.structuralViewCast(node, row.value, source, target)
			var notYet *NotYet
			if got != nil || !errors.As(err, &notYet) || notYet.What != row.want {
				t.Fatalf("got value %v and error %v; want no value and NotYet %q", got, err, row.want)
			}
		})
	}
	// The pair the other way, Derived over Base, isn't a downcast castProof would prove, so it stops too, never nil.
	got, err := lowered.structuralViewCast(node, ir.Read{Of: ir.Object}, target, source)
	var notYet *NotYet
	if got != nil || !errors.As(err, &notYet) {
		t.Fatalf("an unproven pair: got value %v and error %v; want no value and a NotYet", got, err)
	}
}
