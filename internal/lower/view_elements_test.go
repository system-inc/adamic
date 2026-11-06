package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A failed element check is a panic, exit 70 on both backends, which no catch takes. So a checked
// read is checking, and still no throw: it moves no function onto the throwing paths.
func TestViewElementFailureSummary(t *testing.T) {
	t.Parallel()
	l := &lowering{result: &ir.Program{}}
	read := ir.ViewElement{Array: ir.Undefined{Of: ir.Array}, Index: ir.NumberConstant{}, Contract: 1, Of: ir.Object}
	if viewReadOf(l.result, read) != viewReadChecks {
		t.Fatal("element read is not classified as checking")
	}
	if l.throwsOut([]ir.Statement{ir.Evaluate{Value: read}}) {
		t.Fatal("an element check's panic is summarized as a throw a catch could take")
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
		if l.throwsOut([]ir.Statement{statement}) {
			t.Errorf("%T's element check is summarized as a throw a catch could take", node)
		}
	}
}
