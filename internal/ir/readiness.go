package ir

// Throws identifies language dead-zone reads. Placeholder proof failures remain terminal.
func (r Read) Throws(program *Program) bool {
	return r.Readiness == "" && (r.Checked && !program.Locals[r.Local].Hoisted || r.Unset && program.Locals[r.Local].Global && !program.Locals[r.Local].Hoisted)
}

// A TDZ write checks readiness after evaluating its right side.
func (a Assign) Throws(program *Program) bool {
	local := program.Locals[a.Local]
	return a.Checked || local.Uninitialized && local.Global && !local.Hoisted
}
