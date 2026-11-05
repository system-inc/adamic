package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestViewElementFailureSummary(t *testing.T) {
	t.Parallel()
	l := &lowering{result: &ir.Program{}}
	read := ir.ViewElement{Value: ir.Undefined{Of: ir.Object}, Contract: 1, Of: ir.Object}
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
