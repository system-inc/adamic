package ir

// TupleViewMembers plans the existing scalar/tuple alternatives at a selected
// read. It does not inspect tuple descendants, which remain lazy obligations.
func TupleViewMembers(program *Program, id ViewContractID) ([]ViewContractID, bool) {
	var members []ViewContractID
	active := map[ViewContractID]bool{}
	tuples := 0
	var visit func(ViewContractID) bool
	visit = func(id ViewContractID) bool {
		if id <= 0 || int(id) > len(program.ViewContracts) || active[id] {
			return false
		}
		c := program.ViewContracts[id-1]
		if c.Unsupported != "" {
			return false
		}
		active[id] = true
		defer delete(active, id)
		if c.Kind == ViewUnion {
			if len(c.Members) == 0 {
				return false
			}
			for _, child := range c.Members {
				if !visit(child) {
					return false
				}
			}
			return true
		}
		if c.FixedTuple && c.Kind == ViewObject && c.Of == Object {
			tuples++
		} else if c.Kind != ViewUndefined && (c.Kind != ViewScalar || c.Of != Number && c.Of != String && c.Of != Boolean) {
			return false
		}
		members = append(members, id)
		return true
	}
	if !visit(id) || tuples == 0 {
		return nil, false
	}
	return members, true
}
