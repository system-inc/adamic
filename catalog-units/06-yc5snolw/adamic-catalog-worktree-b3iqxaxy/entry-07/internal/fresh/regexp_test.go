package fresh_test

import (
	"testing"

	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

// Keep the regex IR proof testable on main before the native regex slice merges.
// The source fixture in testdata/regexp_tree.ts is also run by the full Node oracle
// on the CSS scratch merge; docs/regex-cycle-proof.md records that command.
func TestRegexOperationsDoNotPoisonTreeWrites(t *testing.T) {
	t.Parallel()
	expressions := []ir.Expression{
		ir.RegExpNew{}, ir.Null{}, ir.IsNull{Value: ir.Null{}},
		ir.RegExpProperty{Array: ir.Read{Local: 0, Of: ir.Array}, Name: "groups", Of: ir.Object},
		ir.RegExpGroup{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "word", Of: ir.String},
	}
	for _, method := range []string{"test", "exec", "match", "matchAll", "next", "iteratorDone", "replace", "replaceAll", "split", "search"} {
		for _, returns := range []ir.Type{ir.Boolean, ir.String, ir.Array, ir.Object} {
			expressions = append(expressions, ir.RegExpCall{Value: ir.RegExpNew{}, Method: method, Returns: returns})
		}
	}
	for _, expression := range expressions {
		program := regexTreeProgram(expression)
		writes := fresh.ProveWrites(program)
		if len(writes) != 1 || !writes[0].Proven {
			t.Errorf("%T %+v: want one proven tree write, got %+v", expression, expression, writes)
		}
	}
}

func regexTreeProgram(expression ir.Expression) *ir.Program {
	return &ir.Program{
		Locals: []ir.Local{{Type: ir.Object, Function: 0}},
		Functions: []ir.Function{{Name: "tree", Parameters: []int{0}, Body: []ir.Statement{
			ir.Evaluate{Value: expression},
			ir.SetProperty{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "next", Value: ir.ObjectLiteral{}, Site: 1},
			ir.Return{},
		}}},
		Main: []ir.Statement{ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.ObjectLiteral{}}}}},
	}
}

func TestRegexOperandsStillJudgeCycleWrites(t *testing.T) {
	t.Parallel()
	closes := ir.ArrayPush{Array: ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "nodes", Of: ir.Array},
		Value: ir.Read{Local: 0, Of: ir.Object}, Element: ir.Object, Site: 2}
	for _, expression := range []ir.Expression{
		ir.RegExpNew{Arguments: []ir.Expression{ir.StringFromCodes{Codes: []ir.Expression{closes}}}},
		ir.RegExpCall{Value: ir.RegExpNew{}, Method: "test", Returns: ir.Boolean,
			Arguments: []ir.Expression{ir.StringFromCodes{Codes: []ir.Expression{closes}}}},
	} {
		program := regexTreeProgram(expression)
		program.Main = []ir.Statement{
			ir.Declare{Local: 1, Value: ir.ObjectLiteral{Fields: []ir.Field{{Name: "nodes", Value: ir.ArrayLiteral{Element: ir.Object}}}}},
			ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.Read{Local: 1, Of: ir.Object}}}},
		}
		program.Locals = append(program.Locals, ir.Local{Type: ir.Object, Function: -1})
		found := false
		for _, write := range fresh.ProveWrites(program) {
			if write.Site == 2 && !write.Proven && write.Kind == fresh.WriteElement {
				found = true
			}
		}
		if !found {
			t.Errorf("%T lost the cycle-closing operand write", expression)
		}
	}
}

func TestFutureRegexMethodRemainsUnknown(t *testing.T) {
	t.Parallel()
	program := regexTreeProgram(ir.RegExpCall{Value: ir.RegExpNew{}, Method: "futureCallback", Returns: ir.Object})
	for _, write := range fresh.ProveWrites(program) {
		if write.Kind == fresh.WriteUnknown {
			return
		}
	}
	t.Fatal("a regex method without a proved effect became silently safe")
}
