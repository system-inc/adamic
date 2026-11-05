package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestViewElementFailureSummary(t *testing.T) {
	t.Parallel()
	l := &lowering{result: &ir.Program{}}
	read := ir.ViewElement{Array: ir.Undefined{Of: ir.Array}, Index: ir.NumberConstant{}, Contract: 1, Of: ir.Object}
	if viewReadOf(l.result, read) != viewReadChecks {
		t.Fatal("element read is not classified as checking")
	}
	if !l.throwsOut([]ir.Statement{ir.Evaluate{Value: read}}) {
		t.Fatal("element check failure is missing from the function summary")
	}
	if !l.throwsOut([]ir.Statement{ir.Evaluate{Value: ir.Comma{Left: read, Right: ir.Defined{Value: ir.Undefined{Of: ir.Object}, Message: "invariant"}}}}) {
		t.Fatal("a later nonthrowing read erased the element check failure")
	}
}

func TestViewElementImplicitFailureSummary(t *testing.T) {
	t.Parallel()
	l := &lowering{result: &ir.Program{}}
	check := &ir.ViewElementCheck{Contract: 1}
	for _, node := range []any{ir.ForOf{ElementCheck: check}, ir.ArrayMap{ElementCheck: check}, ir.ArrayVisit{ElementCheck: check}, ir.ArrayReduce{ElementCheck: check}} {
		var statement ir.Statement
		if value, ok := node.(ir.Statement); ok {
			statement = value
		} else {
			statement = ir.Evaluate{Value: node.(ir.Expression)}
		}
		if !l.throwsOut([]ir.Statement{statement}) {
			t.Errorf("%T lost its element-check failure", node)
		}
	}
}
