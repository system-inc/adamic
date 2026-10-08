package ir

// DictionaryReadKinds is the common read adapter boundary. Unsupported child
// families and mixed finite alternatives retain a demanded-read refusal.
func DictionaryReadKinds(program *Program, id ViewContractID) ([]string, bool) {
	if id <= 0 || int(id) > len(program.ViewContracts) {
		return nil, false
	}
	contract := program.ViewContracts[id-1]
	if contract.Unsupported != "" {
		return nil, false
	}
	var kinds []string
	switch contract.Kind {
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
			values, ok := DictionaryReadKinds(program, child)
			if !ok {
				return nil, false
			}
			kinds = append(kinds, values...)
		}
	default:
		return nil, false
	}
	if contract.Undefined {
		kinds = append(kinds, "undefined")
	}
	return kinds, true
}
