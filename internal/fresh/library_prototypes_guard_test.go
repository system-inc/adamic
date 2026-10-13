package fresh

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestPrototypeCallsCaptureOwnershipGraph(t *testing.T) {
	t.Parallel()
	for _, method := range []string{"create", "setPrototypeOf", "getPrototypeOf"} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			analysis := analysis{state: newState()}
			held := analysis.objectCall(ir.ObjectCall{Method: method, Returns: ir.Object})
			if !analysis.state.clobbered || !held.strong[outside] {
				t.Fatal("prototype topology was treated as confined scalar reflection")
			}
		})
	}
}
