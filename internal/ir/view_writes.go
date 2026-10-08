package ir

// ScalarWriteContracts includes complete scalar and structural object contracts.
// Callable and array families still require their shared contract adapters.
func ScalarWriteContracts(program *Program, written ViewContractID) []ViewContractID {
	var accepted []ViewContractID
	var assignable func(ViewContractID, ViewContractID, map[[2]ViewContractID]bool) bool
	base := func(of Type) Type {
		if of == MaybeNumber {
			return Number
		}
		if of == MaybeBoolean {
			return Boolean
		}
		return of
	}
	assignable = func(from, to ViewContractID, seen map[[2]ViewContractID]bool) bool {
		if from <= 0 || to <= 0 || int(from) > len(program.ViewContracts) || int(to) > len(program.ViewContracts) {
			return false
		}
		source, target := program.ViewContracts[from-1], program.ViewContracts[to-1]
		source = nominalReferenceSlot(program, source)
		target = nominalReferenceSlot(program, target)
		if source.Kind == ViewNull {
			return target.Kind == ViewNull || target.Null
		}
		if source.Null && !target.Null {
			return false
		}
		if source.Kind == ViewUndefined {
			return target.Kind == ViewUndefined || target.Undefined
		}
		if source.Undefined && !target.Undefined {
			return false
		}
		if source.Kind != target.Kind || base(source.Of) != base(target.Of) {
			return false
		}
		if target.Nominal != "" && source.Nominal != target.Nominal {
			ancestor := false
			for _, base := range source.NominalBases {
				ancestor = ancestor || base == target.Nominal
			}
			if !ancestor {
				return false
			}
		}
		pair := [2]ViewContractID{from, to}
		if seen[pair] {
			return true
		}
		seen[pair] = true
		if source.Kind == ViewScalar {
			if len(target.Allowed) == 0 {
				return true
			}
			if len(source.Allowed) == 0 {
				return false
			}
			for _, value := range source.Allowed {
				found := false
				for _, allowed := range target.Allowed {
					found = found || value == allowed
				}
				if !found {
					return false
				}
			}
			return true
		}
		if source.Kind != ViewObject {
			return false
		}
		for _, field := range target.Fields {
			found := false
			for _, own := range source.Fields {
				if own.Name == field.Name {
					found = true
					if own.Optional && !field.Optional || !assignable(own.Contract, field.Contract, seen) {
						return false
					}
					if !field.Readonly && (own.Readonly || !assignable(field.Contract, own.Contract, seen)) {
						return false
					}
				}
			}
			if !found && !field.Optional {
				return false
			}
		}
		return true
	}
	for index := range program.ViewContracts {
		target := ViewContractID(index + 1)
		if assignable(written, target, map[[2]ViewContractID]bool{}) {
			accepted = append(accepted, target)
		}
	}
	return accepted
}

// Nullable class-bearing object slots preserve counted references. Keep both
// nullish flags while comparing their complete private object descriptors.
func nominalReferenceSlot(program *Program, contract ViewContract) ViewContract {
	if contract.Kind != ViewNullable || (contract.Of != Object && contract.Of != Union) || contract.Element == 0 {
		return contract
	}
	present := program.ViewContracts[contract.Element-1]
	if present.Kind != ViewObject || !HasMapNominalWitness(program, contract.Element) || present.Unsupported != "" {
		return contract
	}
	present.Null = present.Null || contract.Null
	present.Undefined = present.Undefined || contract.Undefined
	return present
}
