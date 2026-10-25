package lower

import "github.com/system-inc/adamic/internal/ir"

// checkViewLogicalWrites refuses &&=, ||= and ??= through a checked view where a computed key names the
// field, or the field is optional. Both emitters read and write a named required field through the view
// as a plain read and = through it do. A computed key has no checked form to share, since a plain read
// through the view by one is refused and a plain = isn't lowered, and the checked read has none for an
// optional field. It runs once lowering is done, since a cast lowered after the assignment can still
// make its field checked.
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
		if !ok {
			return true
		}
		read, isProperty := assignment.Read.(ir.Property)
		if assignment.Key == nil && isProperty && !read.Absent {
			return true
		}
		names := []string{write.Name}
		if assignment.Key != nil {
			names = assignment.KeyNames
		}
		for _, name := range names {
			if l.result.CheckedField(write.ViewReceiverTypeID, name) {
				// A write without a recorded site is refused all the same, at its read's place or the program's.
				where := l.result.Source
				if isProperty && read.ViewWhere != "" {
					where = read.ViewWhere
				}
				if write.Site != 0 {
					where = l.program.Where(l.writeSites[write.Site-1].node)
				}
				refused = &Refused{Where: where, What: "a logical assignment through a checked view", Fix: "write the field through the object's own type, or read it, decide, then assign with ="}
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
