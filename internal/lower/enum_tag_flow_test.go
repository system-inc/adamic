package lower

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// A flow node the walk doesn't know refuses the read: the rule fails closed by construction, not
// by listing every step it refuses (#drwm5pj part 1, Adamic's first condition). FlowFlagsShared
// alone names no kind of step. Mutant: the walk's default case passing turns this red.
func TestEnumTagWalkRefusesAnUnknownFlowNode(t *testing.T) {
	t.Parallel()
	walk := &enumTagWalk{seen: map[*ast.FlowNode]bool{}}
	if walk.visit(&ast.FlowNode{Flags: ast.FlowFlagsShared}) {
		t.Fatal("an unknown flow node passed the walk")
	}
}
