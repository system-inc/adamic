package ir

// MapCertificatePairs retains source storage and semantic subtype evidence.
// Writable maps require both directions; readonly maps permit covariance.
func MapCertificatePairs(program *Program, target ViewContractID) [][2]ViewContractID {
	c := program.ViewContracts[target-1]
	pairs := [][2]ViewContractID{}
	active := map[[2]ViewContractID]bool{}
	var accepts func(ViewContractID, ViewContractID) bool
	accepts = func(from, to ViewContractID) bool {
		if from == 0 || to == 0 {
			return false
		}
		if program.ViewContracts[from-1].Of != program.ViewContracts[to-1].Of {
			return false
		}
		// Readonly arrays permit element covariance; writable arrays require both
		// directions. Every recursive step also preserves physical storage.
		source, target := program.ViewContracts[from-1], program.ViewContracts[to-1]
		if source.Kind == ViewArray {
			pair := [2]ViewContractID{from, to}
			if active[pair] {
				return true
			}
			active[pair] = true
			defer delete(active, pair)
			if target.Kind != ViewArray || source.ArrayReadonly && !target.ArrayReadonly {
				return false
			}
			return accepts(source.Element, target.Element) && (target.ArrayReadonly || accepts(target.Element, source.Element))
		}
		for _, id := range ScalarWriteContracts(program, from) {
			if id == to {
				return true
			}
		}
		return false
	}
	for _, pair := range program.MapCertificates {
		key, value := pair[0], pair[1]
		if accepts(key, c.Key) && accepts(value, c.Element) && (c.MapReadonly || accepts(c.Key, key) && accepts(c.Element, value)) {
			pairs = append(pairs, pair)
		}
	}
	return pairs
}
