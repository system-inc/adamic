package lower

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// A flow node the walk doesn't know refuses the read: the rule fails closed by construction, not
// by listing every step it refuses (#drwm5pj part 1, Adamic's first condition). FlowFlagsShared
// alone names no kind of step. Mutant: the walk's default case passing turns this red.
func TestEnumTagWalkRefusesAnUnknownFlowNode(test *testing.T) {
	test.Parallel()
	walk := &enumTagWalk{seen: map[*ast.FlowNode]bool{}}
	if walk.visit(&ast.FlowNode{Flags: ast.FlowFlagsShared}) {
		test.Fatal("an unknown flow node passed the walk")
	}
}
