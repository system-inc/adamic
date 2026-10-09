package lower

import (
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// Carry the entire referenced group through every intervening lexical closure.
// Its layout may grow when a later sibling body is lowered, so completion also
// forwards those cells. The identity cell ensures even capture-free activations
// have distinct canonical function values.
func (l *lowering) nestedReferenceEnvironment(index int) {
	owner := l.result.Functions[index].NestedParent - 1
	if l.result.Functions[owner].FrameIdentity == 0 {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "nested_identity", Type: ir.Number, Function: owner, Captured: true, Preallocated: true})
		l.result.Functions[owner].FrameIdentity = local + 1
	}
	l.touch(l.result.Functions[owner].FrameIdentity - 1)
	for _, local := range l.result.Functions[index].Environment {
		l.touch(local)
	}
	start := 0
	for position, closure := range l.closures {
		if closure == owner {
			start = position + 1
		}
	}
	for _, closure := range l.closures[start:] {
		target := &l.result.Functions[closure]
		if !slices.Contains(target.ReferenceParents, owner+1) {
			target.ReferenceParents = append(target.ReferenceParents, owner+1)
		}
	}
}
