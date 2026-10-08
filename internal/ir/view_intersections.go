package ir

// RecursiveIntersectionObjects returns only the object descriptors reachable
// from this read. Optional copies resolve to their completed canonical object.
func RecursiveIntersectionObjects(program *Program, root ViewContractID) []ViewContractID {
	ids := []ViewContractID{}
	seen := map[ViewContractID]bool{}
	var visit func(ViewContractID)
	visit = func(id ViewContractID) {
		contract := program.ViewContracts[id-1]
		if contract.ObjectPresent != 0 {
			visit(contract.ObjectPresent)
			return
		}
		if contract.Kind != ViewObject || seen[id] || contract.Unsupported != "" && contract.Unsupported != "recursive intersection payload" {
			return
		}
		seen[id] = true
		ids = append(ids, id)
		for _, field := range contract.Fields {
			visit(field.Contract)
		}
	}
	visit(root)
	return ids
}

func RecursiveIntersectionPresent(program *Program, id ViewContractID) ViewContractID {
	for program.ViewContracts[id-1].ObjectPresent != 0 {
		id = program.ViewContracts[id-1].ObjectPresent
	}
	return id
}
