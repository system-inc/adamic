package ir

// A shared required array field does not select an object-union arm. Its read
// proves the same complete array contract in every arm; other fields stay lazy.
func SharedUntaggedArrayRead(program *Program, receiver ViewContractID, name string, child ViewContractID) bool {
	if receiver <= 0 || int(receiver) > len(program.ViewContracts) || child <= 0 || int(child) > len(program.ViewContracts) {
		return false
	}
	root := program.ViewContracts[receiver-1]
	array := program.ViewContracts[child-1]
	if root.Kind != ViewUnion || root.Of != Object || root.Unsupported != "untagged object union" || len(root.Members) < 2 || array.Kind != ViewArray || array.Unsupported != "" || array.Undefined || array.Null || !PrimitiveArrayContract(program, array.Element) {
		return false
	}
	for _, member := range root.Members {
		if member <= 0 || int(member) > len(program.ViewContracts) {
			return false
		}
		object := program.ViewContracts[member-1]
		if object.Kind != ViewObject || object.Of != Object || object.NominalClass != 0 || object.Unsupported != "" {
			return false
		}
		found := false
		for _, field := range object.Fields {
			if field.Name == name && !field.Optional && field.Contract == child {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
