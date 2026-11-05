package fresh

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestViewElementPreservesFreshness(t *testing.T) {
	t.Parallel()
	a := &analysis{proof: &freshness{program: &ir.Program{Locals: []ir.Local{{Type: ir.Array}}}}, state: newState()}
	a.state.locals[0] = value{strong: objects{newest(0): true}}
	a.state.store(newest(0), elementKey, value{strong: objects{newest(1): true}})
	got := a.value(ir.ViewElement{Array: ir.Read{Local: 0, Of: ir.Array}, Index: ir.NumberConstant{}, Contract: 1, Of: ir.Object})
	if len(got.strong) != 1 || !got.strong[newest(1)] || a.state.clobbered || a.writeCount != 0 || len(a.state.escaped) != 0 {
		t.Fatalf("a check lost the value's freshness or invented a write: %#v", got)
	}
}
