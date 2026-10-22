package lower

import "github.com/system-inc/adamic/internal/ir"

// checkViewLogicalWrites refuses &&=, ||= and ??= through a checked view. Both emitters read and write a
// logical assignment's field as its static type says, without the view's check or the boxing a union slot
// needs, so a number written through a view into a string | number slot is stored as a raw double. It runs
// once lowering is done, since a cast lowered after the assignment can still make its field checked.
func (l *lowering) checkViewLogicalWrites() error {
	var refused error
	visit := func(node any) bool {
		if refused != nil {
			return false
		}
		assignment, ok := node.(ir.LogicalAssignment)
		if !ok {
			return true
		}
		write, ok := assignment.Write.(ir.SetProperty)
		if !ok || write.Site == 0 {
			return true
		}
		names := []string{write.Name}
		if assignment.Key != nil {
			names = assignment.KeyNames
		}
		for _, name := range names {
			if l.result.CheckedField(write.ViewReceiverTypeID, name) {
				refused = &Refused{Where: l.program.Where(l.writeSites[write.Site-1].node), What: "a logical assignment through a checked view", Fix: "write the field through the object's own type, or read it, decide, then assign with ="}
				return false
			}
		}
		return true
	}
	walk(l.result.Main, visit)
	for _, function := range l.result.Functions {
		walk(function.Body, visit)
	}
	return refused
}
