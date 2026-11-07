package lower

import (
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// Carry the entire referenced group through every intervening lexical closure.
// Its layout may grow when a later sibling body is lowered, so completion also
// forwards those cells. The identity cell ensures even capture-free activations
// have distinct canonical function values.
func (lowering *lowering) nestedReferenceEnvironment(index int) {
	owner := lowering.result.Functions[index].NestedParent - 1
	if lowering.result.Functions[owner].FrameIdentity == 0 {
		local := len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: "nested_identity", Type: ir.Number, Function: owner, Captured: true, Preallocated: true},
		)
		lowering.result.Functions[owner].FrameIdentity = local + 1
	}
	lowering.touch(lowering.result.Functions[owner].FrameIdentity - 1)
	for _, local := range lowering.result.Functions[index].Environment {
		lowering.touch(local)
	}
	start := 0
	for position, closure := range lowering.closures {
		if closure == owner {
			start = position + 1
		}
	}
	for _, closure := range lowering.closures[start:] {
		target := &lowering.result.Functions[closure]
		if !slices.Contains(target.ReferenceParents, owner+1) {
			target.ReferenceParents = append(target.ReferenceParents, owner+1)
		}
	}
}
