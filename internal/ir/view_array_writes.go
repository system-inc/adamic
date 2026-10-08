package ir

// FlatArrayRecordContract admits required scalar data fields only. Unknown,
// callable, nested, optional and nominal declarations never certify a write.
func FlatArrayRecordContract(program *Program, id ViewContractID) bool {
	if id <= 0 || int(id) > len(program.ViewContracts) {
		return false
	}
	contract := program.ViewContracts[id-1]
	if contract.Kind != ViewObject || contract.Unsupported != "" || contract.Nominal != "" || contract.Undefined {
		return false
	}
	for _, field := range contract.Fields {
		if field.Optional || field.Contract <= 0 || int(field.Contract) > len(program.ViewContracts) {
			return false
		}
		child := program.ViewContracts[field.Contract-1]
		if child.Kind != ViewScalar || child.Unsupported != "" || child.Undefined || (child.Of != Number && child.Of != Boolean && child.Of != String) {
			return false
		}
	}
	return true
}

// ArrayRecordWritePairs uses the shared structural assignment rule, including
// bidirectional mutable fields. Physical pointer storage is never a certificate.
func ArrayRecordWritePairs(program *Program) [][2]ViewContractID {
	var pairs [][2]ViewContractID
	for index := range program.ViewContracts {
		source := ViewContractID(index + 1)
		if !FlatArrayRecordContract(program, source) {
			continue
		}
		for _, target := range ScalarWriteContracts(program, source) {
			if FlatArrayRecordContract(program, target) {
				pairs = append(pairs, [2]ViewContractID{source, target})
			}
		}
	}
	return pairs
}
