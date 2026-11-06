package native

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// madeUpExpression is an operation the IR doesn't have, standing for one added later that reuse's
// global proof has no case for.
type madeUpExpression struct{}

func (madeUpExpression) Type() ir.Type { return ir.Number }

// The move of a global into a consumed parameter, and the lending of one to a borrowed parameter, ask
// whether anything that runs meanwhile may touch the global. An operation reachesGlobal has no case
// for counts as touching it unless it is ir.Inert, so one added later fails closed (Compiler's and
// Adamic's ruling on #qk8rztp S2c-2b). Mutant: reachesGlobal's last line made `return false`, and the
// made-up operation, in the expression and in a function it calls, touches nothing.
func TestReuseTakesAnOperationNoReviewNamesAsTouchingTheGlobal(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals:    []ir.Local{{Name: "tree", Type: ir.Object, Global: true, Function: -1}},
		Functions: []ir.Function{{Name: "helper", Body: []ir.Statement{ir.Evaluate{Value: madeUpExpression{}}}}, {Name: "plain", Body: []ir.Statement{ir.Evaluate{Value: ir.Binary{Left: ir.NumberConstant{Value: 1}, Right: ir.NumberConstant{Value: 2}}}}}},
	}
	if !reachesGlobal(program, madeUpExpression{}, 0, map[int]bool{}) {
		t.Fatal("a made-up operation is taken to touch nothing; reachesGlobal's default must be closed")
	}
	if !touches(program, 0, 0, map[int]bool{}) {
		t.Fatal("a function evaluating a made-up operation is taken to touch nothing")
	}
	if touches(program, 1, 0, map[int]bool{}) {
		t.Fatal("a function of inert operations is taken to touch the global")
	}
}

// A generator's step runs its body, which touches finds the global in through the step's Function.
// Mutant: the GeneratorNext arm made `return false`.
func TestReuseFollowsAGeneratorStepIntoItsBody(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Name: "tree", Type: ir.Object, Global: true, Function: -1}, {Name: "frame", Type: ir.Generator, Function: 0}},
		Functions: []ir.Function{
			{Name: "steps", Body: []ir.Statement{ir.Evaluate{Value: ir.GeneratorNext{Generator: ir.Read{Local: 1, Of: ir.Generator}, Function: 1, Of: ir.Object}}}},
			{Name: "numbers", Generator: true, Body: []ir.Statement{ir.Evaluate{Value: ir.Read{Local: 0, Of: ir.Object}}}},
		},
	}
	if !touches(program, 0, 0, map[int]bool{}) {
		t.Fatal("a step of a generator whose body reads the global is taken to touch nothing")
	}
}

// A spread runs the getters lowering recorded for it (ir.ObjectLiteral's SpreadGetters), so a
// function whose spread may run a getter that writes the global touches it, and a spread with none
// doesn't (Memory's and Compiler's ruling on #qk8rztp S2c-2b). Mutant: the ObjectLiteral arm's
// touches made `false &&`, and the getter's write goes unseen.
func TestReuseTakesASpreadsGettersAsWhatItRuns(t *testing.T) {
	t.Parallel()
	spread := func(getters []int) ir.Function {
		return ir.Function{Name: "describe", Body: []ir.Statement{ir.Evaluate{Value: ir.ObjectLiteral{Spread: ir.Read{Local: 1, Of: ir.Object}, SpreadGetters: getters}}}}
	}
	program := &ir.Program{
		Locals: []ir.Local{{Name: "tree", Type: ir.Object, Global: true, Function: -1}, {Name: "extra", Type: ir.Object, Function: 0}},
		Functions: []ir.Function{
			spread([]int{1}),
			{Name: "get value", Body: []ir.Statement{ir.Assign{Local: 0, Value: ir.ObjectLiteral{}}}},
			spread(nil),
		},
	}
	if !touches(program, 0, 0, map[int]bool{}) {
		t.Fatal("a spread whose getter writes the global is taken to touch nothing")
	}
	if touches(program, 2, 0, map[int]bool{}) {
		t.Fatal("a spread with no getter is taken to touch the global")
	}
}
