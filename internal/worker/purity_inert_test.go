package worker

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// madeUpExpression is an operation the IR doesn't have, standing for one added later that CheckPure
// has no case for.
type madeUpExpression struct{ Value ir.Expression }

func (madeUpExpression) Type() ir.Type { return ir.Number }

// An expression CheckPure has no case for is refused unless it is ir.Inert, so an operation added
// later, which could call code or reach the host, fails closed (Compiler's and Adamic's ruling on
// #qk8rztp S2c-2b). Mutant: purity.go's default arm made `if false && ...`, and the made-up
// operation passes as pure.
func TestCheckPureRefusesAnOperationNoReviewNames(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Functions: []ir.Function{{Name: "compute", Body: []ir.Statement{ir.Return{Value: madeUpExpression{Value: ir.NumberConstant{Value: 1}}}}}}}
	if err := CheckPure(program, 0); err == nil || !strings.Contains(err.Error(), "impure: an operation no purity review names") {
		t.Fatalf("want the made-up operation refused, got %v", err)
	}
	inert := &ir.Program{Functions: []ir.Function{{Name: "compute", Body: []ir.Statement{ir.Return{Value: ir.Binary{Left: ir.NumberConstant{Value: 1}, Right: ir.NumberConstant{Value: 2}}}}}}}
	if err := CheckPure(inert, 0); err != nil {
		t.Fatalf("want an inert operation pure, got %v", err)
	}
}

// A generator's step runs its body, which is inspected through the step's own Function, since the
// call that made the frame may be outside what CheckPure walks. Mutant: the GeneratorNext arm's
// inspect dropped, and the body's console write passes.
func TestCheckPureInspectsAGeneratorStepsBody(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Functions: []ir.Function{
		{Name: "compute", Body: []ir.Statement{ir.Return{Value: ir.GeneratorNext{Generator: ir.Read{Local: 0, Of: ir.Object}, Function: 1, Of: ir.Number}}}},
		{Name: "numbers", Body: []ir.Statement{ir.WriteLine{Stream: ir.Stdout, Value: ir.StringConstant{Index: 0}}}},
	}, Locals: []ir.Local{{Name: "iterator", Type: ir.Object, Function: 0}}}
	if err := CheckPure(program, 0); err == nil || !strings.Contains(err.Error(), "calls console") {
		t.Fatalf("want the stepped body's console write refused, got %v", err)
	}
}

// A spread runs the getters lowering recorded for it (ir.ObjectLiteral's SpreadGetters), so each must
// be pure (Memory's and Compiler's ruling on #qk8rztp S2c-2b). Mutant: the ObjectLiteral arm's
// inspect made `if err := error(nil); getter < 0 && err != nil {`, and the getter's console write
// passes.
func TestCheckPureInspectsASpreadsGetters(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Functions: []ir.Function{
		{Name: "compute", Body: []ir.Statement{ir.Evaluate{Value: ir.ObjectLiteral{Spread: ir.Read{Local: 0, Of: ir.Object}, SpreadGetters: []int{1}}}}},
		{Name: "get value", Body: []ir.Statement{ir.WriteLine{Stream: ir.Stdout, Value: ir.StringConstant{Index: 0}}}},
	}, Locals: []ir.Local{{Name: "extra", Type: ir.Object, Function: 0}}}
	if err := CheckPure(program, 0); err == nil || !strings.Contains(err.Error(), "calls console") {
		t.Fatalf("want the spread's getter's console write refused, got %v", err)
	}
}
