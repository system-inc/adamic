package ir

// DictionaryReadKinds is the common read adapter boundary. Unsupported child
// families and mixed finite alternatives retain a demanded-read refusal.
func DictionaryReadKinds(program *Program, id ViewContractID) ([]string, bool) {
	return dictionaryReadKinds(program, id, map[ViewContractID]bool{})
}

func dictionaryReadKinds(program *Program, id ViewContractID, active map[ViewContractID]bool) ([]string, bool) {
	if id <= 0 || int(id) > len(program.ViewContracts) || active[id] {
		return nil, false
	}
	contract := program.ViewContracts[id-1]
	if contract.Unsupported != "" {
		return nil, false
	}
	active[id] = true
	defer delete(active, id)
	var kinds []string
	switch contract.Kind {
	case ViewNullable:
		values, ok := dictionaryReadKinds(program, contract.Element, active)
		if !ok {
			return nil, false
		}
		kinds = values
	case ViewNull:
		kinds = []string{"null"}
	case ViewUndefined:
		kinds = []string{"undefined"}
	case ViewScalar:
		switch contract.Of.Present() {
		case Number:
			kinds = []string{"number"}
		case Boolean:
			kinds = []string{"boolean"}
		case String:
			kinds = []string{"string"}
		default:
			return nil, false
		}
	case ViewArray:
		kinds = []string{"array"}
	case ViewObject, ViewDictionary:
		if contract.Nominal != "" {
			return nil, false
		}
		kinds = []string{"object"}
	case ViewUnion:
		booleanValues := map[bool]bool{}
		for _, child := range contract.Members {
			member := program.ViewContracts[child-1]
			for _, literal := range member.Allowed {
				if literal.Of == Boolean {
					booleanValues[literal.Boolean] = true
				}
			}
		}
		for _, child := range contract.Members {
			member := program.ViewContracts[child-1]
			if len(member.Allowed) != 0 && !(member.Of == Boolean && booleanValues[false] && booleanValues[true]) {
				return nil, false
			}
			values, ok := dictionaryReadKinds(program, child, active)
			if !ok {
				return nil, false
			}
			kinds = append(kinds, values...)
		}
	default:
		return nil, false
	}
	if contract.Null {
		kinds = append(kinds, "null")
	}
	if contract.Undefined {
		kinds = append(kinds, "undefined")
	}
	return kinds, true
}

// A selected read certificate admits only these runtime kinds. Its original
// declared contract and type ID remain available for all reference obligations.
func PrimitiveDictionaryReadCertificate(program *Program, read Property) bool {
	if !read.DictionaryPrimitive || read.DictionaryKey == nil {
		return false
	}
	kinds, ok := DictionaryReadKinds(program, read.ViewContract)
	if !ok || len(kinds) == 0 {
		return false
	}
	for _, kind := range kinds {
		switch kind {
		case "string", "number", "boolean", "null", "undefined":
		default:
			return false
		}
	}
	return true
}

// A complete immediate selector preserves the original child graph for later reads.
func SelectedDictionaryReadCertificate(program *Program, read Property) bool {
	if !read.DictionaryPrimitive && !read.DictionaryReference || read.DictionaryKey == nil {
		return false
	}
	kinds, ok := DictionaryReadKinds(program, read.ViewContract)
	return ok && len(kinds) > 0
}
