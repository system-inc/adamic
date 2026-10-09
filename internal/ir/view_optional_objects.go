package ir

// Only a complete optional plain-object union uses the shared absent-slot probe.
// Every present object still runs union selection before its payload is used.
func OptionalObjectViewContract(contracts []ViewContract, id ViewContractID) bool {
	if id <= 0 || int(id) > len(contracts) {
		return false
	}
	root := contracts[id-1]
	if root.Kind != ViewUnion || root.Of != Object || root.Unsupported != "" || !root.Undefined {
		return false
	}
	objects, undefined := 0, false
	for _, member := range root.Members {
		if member <= 0 || int(member) > len(contracts) {
			return false
		}
		child := contracts[member-1]
		if child.Unsupported != "" {
			return false
		}
		if child.Kind == ViewUndefined {
			undefined = true
			continue
		}
		if child.Kind != ViewObject || child.Of != Object || child.Nominal != "" || child.FixedTuple {
			return false
		}
		objects++
	}
	return objects == 1 && undefined
}
