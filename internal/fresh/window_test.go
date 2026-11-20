package fresh_test

import (
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

// windowFactory is a factory over the construction window's locals: it allocates o, runs between, and returns
// o. Its parameters are debugging, a boolean, and visit, a callback the walk can't see into; kept is a global,
// and function 1, keep, a direct callee that does nothing.
func windowFactory(between ...ir.Statement) *ir.Program {
	body := []ir.Statement{ir.Declare{Local: 0, Value: ir.ObjectLiteral{}}}
	body = append(body, between...)
	body = append(body, ir.Return{Value: ir.Read{Local: 0, Of: ir.Object}})
	return &ir.Program{
		Locals: []ir.Local{
			{Name: "o", Type: ir.Object, Function: 0},
			{Name: "debugging", Type: ir.Boolean, Function: 0},
			{Name: "visit", Type: ir.Closure, Function: 0},
			{Name: "kept", Type: ir.Object, Function: -1, Global: true},
			{Name: "held", Type: ir.Object, Function: 1},
		},
		Functions: []ir.Function{
			{Name: "factory", Parameters: []int{1, 2}, Returns: ir.Object, Body: body},
			{Name: "keep", Parameters: []int{4}, Body: []ir.Statement{ir.Return{}}},
		},
	}
}

// windowWrite is o.name = 1, a statement.
func windowWrite(name string) ir.Statement {
	return ir.SetProperty{Object: ir.Read{Local: 0, Of: ir.Object}, Name: name, Value: ir.NumberConstant{Value: 1}, Site: 1}
}

// The predicate over (allocation, field), asked directly of a factory's body (#w8nanzw (B), rulings.md:78 applied):
// a write on every path, on one path, in a loop that runs zero times and in a short circuit's right side is in
// the window, on every path or some; a write after the object is stored in a global, handed to a callback (Adamic's
// first attack), or handed to any call, which the window takes as one it can't see into, is after the window and
// writes nothing; a call the object isn't handed leaves the window open. Settled holds only where every path
// writes the same fields in one order.
// Mutant (a): the window ignoring escape (confined answering true once the object is made) writes checker in the
// global, callback and direct-call rows.
func TestConstructionWindowWrites(t *testing.T) {
	t.Parallel()
	debugging := ir.Read{Local: 1, Of: ir.Boolean}
	handed := func(arguments ...ir.Expression) ir.Statement {
		return ir.Evaluate{Value: ir.CallClosure{Closure: ir.Read{Local: 2, Of: ir.Closure}, Arguments: arguments}}
	}
	o := ir.Read{Local: 0, Of: ir.Object}
	for _, row := range []struct {
		name    string
		between []ir.Statement
		writes  map[string]fresh.WindowWrite
		order   []string // nil where the window isn't settled
	}{
		{"every path", []ir.Statement{windowWrite("flags"), windowWrite("checker")}, map[string]fresh.WindowWrite{"flags": fresh.WindowEveryPath, "checker": fresh.WindowEveryPath}, []string{"flags", "checker"}},
		{"one path", []ir.Statement{windowWrite("flags"), ir.If{Condition: debugging, Then: []ir.Statement{windowWrite("checker")}}}, map[string]fresh.WindowWrite{"flags": fresh.WindowEveryPath, "checker": fresh.WindowSomePaths}, nil},
		{"both paths, in two orders", []ir.Statement{ir.If{Condition: debugging, Then: []ir.Statement{windowWrite("flags"), windowWrite("checker")}, Else: []ir.Statement{windowWrite("checker"), windowWrite("flags")}}}, map[string]fresh.WindowWrite{"flags": fresh.WindowEveryPath, "checker": fresh.WindowEveryPath}, nil},
		{"a loop that runs zero times", []ir.Statement{windowWrite("flags"), ir.Loop{Condition: debugging, Body: []ir.Statement{windowWrite("checker")}}}, map[string]fresh.WindowWrite{"flags": fresh.WindowEveryPath, "checker": fresh.WindowSomePaths}, nil},
		{"a short circuit", []ir.Statement{ir.Evaluate{Value: ir.Logical{Left: debugging, Right: ir.Effects{Body: []ir.Statement{windowWrite("checker")}, Result: ir.BooleanConstant{Value: true}}, Of: ir.Boolean}}}, map[string]fresh.WindowWrite{"checker": fresh.WindowSomePaths}, nil},
		{"after a store in a global", []ir.Statement{windowWrite("flags"), ir.Assign{Local: 3, Value: o}, windowWrite("checker")}, map[string]fresh.WindowWrite{"flags": fresh.WindowEveryPath, "checker": fresh.WindowUnwritten}, []string{"flags"}},
		{"after a callback is handed it", []ir.Statement{windowWrite("flags"), handed(o), windowWrite("checker")}, map[string]fresh.WindowWrite{"flags": fresh.WindowEveryPath, "checker": fresh.WindowUnwritten}, []string{"flags"}},
		{"after a direct call is handed it", []ir.Statement{ir.Evaluate{Value: ir.Call{Function: 1, Arguments: []ir.Expression{o}}}, windowWrite("checker")}, map[string]fresh.WindowWrite{"checker": fresh.WindowUnwritten}, []string{}},
		{"after a call it isn't handed", []ir.Statement{handed(), windowWrite("checker")}, map[string]fresh.WindowWrite{"checker": fresh.WindowEveryPath}, []string{"checker"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			program := windowFactory(row.between...)
			window := fresh.ConstructionWindow(program, 0, &program.Functions[0].Body[0])
			for name, want := range row.writes {
				if got := window.Writes(name); got != want {
					t.Errorf("Writes(%q) = %d, want %d (window %+v)", name, got, want, window)
				}
			}
			if window.Settled != (row.order != nil) || row.order != nil && !slices.Equal(window.Order, row.order) {
				t.Errorf("Settled %v Order %q, want order %q (nil: not settled)", window.Settled, window.Order, row.order)
			}
		})
	}
}

// An allocation the walk sees escape as it's made, into a global, has a window that writes nothing, and one
// statement that isn't the allocation names no window.
func TestConstructionWindowOfAnEscapedAllocation(t *testing.T) {
	t.Parallel()
	program := windowFactory(windowWrite("checker"))
	program.Functions[0].Body[0] = ir.Declare{Local: 3, Value: ir.ObjectLiteral{}}
	program.Functions[0].Body[1] = ir.SetProperty{Object: ir.Read{Local: 3, Of: ir.Object}, Name: "checker", Value: ir.NumberConstant{Value: 1}, Site: 1}
	if got := fresh.ConstructionWindow(program, 0, &program.Functions[0].Body[0]).Writes("checker"); got != fresh.WindowUnwritten {
		t.Errorf("a global's allocation: Writes(checker) = %d, want WindowUnwritten", got)
	}
	program = windowFactory(windowWrite("checker"))
	if got := fresh.ConstructionWindow(program, 0, &program.Functions[0].Body[1]).Writes("checker"); got != fresh.WindowUnwritten {
		t.Errorf("a write named as the allocation: Writes(checker) = %d, want WindowUnwritten", got)
	}
}
