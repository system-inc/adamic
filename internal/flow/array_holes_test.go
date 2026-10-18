package flow

import (
	"os/exec"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestArrayHolesAliasingIsFresh(t *testing.T) {
	t.Parallel()
	n := &inference{}
	s := n.value(ir.ArrayHoles{Length: ir.NumberConstant{Value: 1}, Element: ir.Object})
	if !s.fresh || s.unknown || len(s.roots()) != 0 {
		t.Fatalf("constructor must create empty storage without an opaque call: %+v", s)
	}
}

func TestArrayHolesOperandMutations(t *testing.T) {
	t.Parallel()
	push := ir.ArrayPush{Array: ir.Read{Local: 0, Of: ir.Array}, Value: ir.NumberConstant{Value: 1}, Element: ir.Number, Site: 1}
	for name, expression := range map[string]ir.Expression{
		"length":   ir.ArrayHoles{Length: push, Element: ir.Number},
		"identity": ir.ArrayRangeErrorIs{Value: ir.Comma{Left: push, Right: ir.ObjectLiteral{}}},
	} {
		t.Run(name, func(t *testing.T) {
			p := &ir.Program{Locals: []ir.Local{{Type: ir.Array, Function: 0}, {Type: expression.Type(), Function: 0}}, Functions: []ir.Function{{Name: "operand", Parameters: []int{0}, Body: []ir.Statement{ir.Declare{Local: 1, Value: expression}, ir.Return{}}}}}
			g := Build(p, 0)
			Construct(g)
			if bad := VerifySSA(g); len(bad) != 0 {
				t.Fatal(bad)
			}
			effects := InferAliasingEffects(g)
			found := false
			for _, instruction := range g.Instructions {
				for _, effect := range effects.Get(instruction.Id) {
					found = found || effect.Kind.IsMutation()
				}
			}
			if !found {
				t.Fatal("operand mutation was discarded")
			}
		})
	}
}

func TestArrayHolesThrowAndIdentityRead(t *testing.T) {
	t.Parallel()
	p := &ir.Program{}
	for _, sample := range []struct {
		expression ir.Expression
		throws     bool
	}{
		{ir.ArrayHoles{Length: ir.NumberConstant{Value: 1}, Element: ir.Number}, true},
		{ir.ArrayRangeErrorIs{Value: ir.ObjectLiteral{}}, false},
		{ir.ArrayRangeErrorIs{Value: ir.Comma{Left: ir.ArrayHoles{Length: ir.NumberConstant{Value: -1}, Element: ir.Number}, Right: ir.ObjectLiteral{}}}, true},
	} {
		statement := ir.Statement(ir.Evaluate{Value: sample.expression})
		instruction := &Instruction{At: &statement, Expression: sample.expression}
		if got := CanThrow(p, instruction); got != sample.throws {
			t.Fatalf("%T throws=%v, want %v", sample.expression, got, sample.throws)
		}
	}
}

// Snapshots must distinguish holes from present undefined values without
// expanding a sparse array into billions of absent slots.
func TestArrayHolesTraceSnapshots(t *testing.T) {
	t.Parallel()
	script := traceRuntime + `
const cells = new Array(4294967295);
console.log(adamicPrint(cells, new Set()));
cells[0] = undefined;
console.log(adamicPrint(cells, new Set()));
cells[4294967294] = 7;
console.log(adamicPrint(cells, new Set()));
delete cells[0];
cells.length--;
console.log(adamicPrint(cells, new Set()));
`
	got, err := exec.Command("node", "-e", script).CombinedOutput()
	if err != nil {
		t.Fatalf("Node snapshot: %v\n%s", err, got)
	}
	want := "array(4294967295){}\narray(4294967295){0:undefined}\narray(4294967295){0:undefined,4294967294:7}\narray(4294967294){}\n"
	if string(got) != want {
		t.Fatalf("sparse snapshot: got %q, want %q", got, want)
	}
}
