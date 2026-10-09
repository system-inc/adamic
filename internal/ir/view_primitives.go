package ir

// PrimitiveViewMembers returns only complete, nonrecursive primitive alternatives.
// A mixed object, unsupported or missing child prevents primitive-only dispatch.
// The payload's storage kind never supplies a missing declared contract.
func PrimitiveViewMembers(program *Program, id ViewContractID) ([]ViewContract, bool) {
	var members []ViewContract
	active := map[ViewContractID]bool{}
	var visit func(ViewContractID) bool
	visit = func(id ViewContractID) bool {
		if id <= 0 || int(id) > len(program.ViewContracts) || active[id] {
			return false
		}
		contract := program.ViewContracts[id-1]
		if contract.Unsupported != "" {
			return false
		}
		active[id] = true
		defer delete(active, id)
		switch contract.Kind {
		case ViewUnion:
			if len(contract.Members) == 0 {
				return false
			}
			for _, child := range contract.Members {
				if !visit(child) {
					return false
				}
			}
			return true
		case ViewUndefined:
			members = append(members, contract)
			return true
		case ViewScalar:
			if contract.Of == MaybeNumber {
				contract.Of = Number
			}
			if contract.Of == MaybeBoolean {
				contract.Of = Boolean
			}
			if contract.Of != Number && contract.Of != Boolean && contract.Of != String {
				return false
			}
			for _, literal := range contract.Allowed {
				if literal.Of != contract.Of {
					return false
				}
			}
			members = append(members, contract)
			if contract.Undefined {
				members = append(members, ViewContract{Kind: ViewUndefined, Undefined: true})
			}
			return true
		default:
			return false
		}
	}
	if !visit(id) {
		return nil, false
	}
	return members, true
}
