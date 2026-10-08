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
