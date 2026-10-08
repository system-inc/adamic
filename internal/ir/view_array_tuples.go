package ir

// Array-to-tuple views retain their homogeneous source array. The marker is an
// existing reference conversion; no new allocation or ownership operand exists.
func HasArrayTupleViews(program *Program) bool {
	for _, origin := range program.ViewOrigins {
		if value, ok := origin.(Narrow); ok && value.Tuple && value.To == Object && value.Value.Type() == Array {
			return true
		}
	}
	return false
}

func ArrayTupleReceiver(program *Program, property Property) ViewContractID {
	if !HasArrayTupleViews(program) {
		return 0
	}
	id := program.ViewContractTypes[property.ViewReceiverTypeID]
	if id <= 0 || int(id) > len(program.ViewContracts) || !program.ViewContracts[id-1].FixedTuple {
		return 0
	}
	return id
}
