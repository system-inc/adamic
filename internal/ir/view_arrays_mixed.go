package ir

// MixedArrayContract describes represented scalar and flat structural object
// alternatives. It certifies selection capability, never array contents or writes.
func MixedArrayContract(program *Program, id ViewContractID) bool {
	if id <= 0 || int(id) > len(program.ViewContracts) {
		return false
	}
	root := program.ViewContracts[id-1]
	if root.Kind != ViewUnion || root.Of != Union || root.Null || len(root.Members) == 0 || root.Unsupported != "" && root.Unsupported != "union intersection" {
		return false
	}
	objects, scalars := 0, 0
	for _, member := range root.Members {
		if member <= 0 || int(member) > len(program.ViewContracts) {
			return false
		}
		c := program.ViewContracts[member-1]
		if !c.Null && c.Kind != ViewNull && PrimitiveArrayContract(program, member) {
			scalars++
			continue
		}
		if c.Kind != ViewObject || c.Of != Object || c.Nominal != "" || c.FixedTuple || c.Unsupported != "" && c.Unsupported != "compound intersection payload" {
			return false
		}
		for _, field := range c.Fields {
			if !PrimitiveArrayContract(program, field.Contract) {
				return false
			}
		}
		objects++
	}
	return objects > 0 && scalars > 0
}
