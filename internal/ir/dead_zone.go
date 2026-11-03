package ir

// DeadZoneMayThrow says a node can throw JavaScript's ReferenceError, "Cannot access 'name' before
// initialization", for a binding touched in its temporal dead zone: a read or write whose readiness
// is checked as it runs (Checked), a read of a global that isn't hoisted made before its declaration
// can have run (Unset), and a write of such a global whose declaration has a dead zone of its own
// (its Local's Uninitialized). These are exactly where the native emitter tests readiness and throws
// (emit_locals.go's read, emit_statements.go's Assign), so every analysis that asks what can throw
// (lower's throwsOut, flow's CanThrow) counts the same nodes, and no throw goes untested by a
// caller or unlanded by a try (#bxmdk73). A Read's Readiness check, "read before assignment", is
// Adamic's own stop and never throws.
func (p *Program) DeadZoneMayThrow(node any) bool {
	switch node := node.(type) {
	case Read:
		return node.Checked || node.Unset && p.Locals[node.Local].Global && !p.Locals[node.Local].Hoisted
	case Assign:
		local := p.Locals[node.Local]
		return node.Checked || local.Uninitialized && local.Global && !local.Hoisted
	}
	return false
}
