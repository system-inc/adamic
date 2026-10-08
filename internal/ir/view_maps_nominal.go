package ir

// HasMapNominalWitness finds class identities beneath a private entry descriptor.
func HasMapNominalWitness(program *Program, id ViewContractID) bool {
	seen := map[ViewContractID]bool{}
	var visit func(ViewContractID) bool
	visit = func(id ViewContractID) bool {
		if id == 0 || seen[id] {
			return false
		}
		seen[id] = true
		c := program.ViewContracts[id-1]
		if c.NominalClass != 0 {
			return true
		}
		if visit(c.Element) {
			return true
		}
		for _, field := range c.Fields {
			if visit(field.Contract) {
				return true
			}
		}
		for _, member := range c.Members {
			if visit(member) {
				return true
			}
		}
		return false
	}
	return visit(id)
}

// Both representations keep counted class references. The shared element reader
// preserves them, while MapCertificatePairs separately proves logical covariance.
func NominalArrayUnionStorage(program *Program, source, target ViewContractID) bool {
	if source == 0 || target == 0 {
		return false
	}
	from, to := program.ViewContracts[source-1], program.ViewContracts[target-1]
	if from.Of != Object || to.Of != Union {
		return false
	}
	present := func(c ViewContract) ViewContract {
		if c.Kind == ViewNullable && c.Element != 0 {
			return program.ViewContracts[c.Element-1]
		}
		return c
	}
	from, to = present(from), present(to)
	return from.NominalClass != 0 && to.NominalClass != 0 && from.Unsupported == "" && to.Unsupported == ""
}
