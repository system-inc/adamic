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
		if source.Undefined && !target.Undefined {
			return false
		}
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
			for _, field := range target.Fields {
				found := false
				for _, own := range source.Fields {
					if own.Name != field.Name {
						continue
					}
					found = true
					if own.Optional && !field.Optional || !accepts(own.Contract, field.Contract) {
						return false
					}
					if !field.Readonly && (own.Readonly || !accepts(field.Contract, own.Contract)) {
						return false
					}
				}
				// A producer lacking a declared optional own field can hide
				// an incompatible extra property. Its absence is not evidence.
				if !found {
					return false
				}
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
