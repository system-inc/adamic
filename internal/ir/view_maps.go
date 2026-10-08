package ir

// MapCertificatePairs retains source storage and semantic subtype evidence.
// Writable maps require both directions; readonly maps permit covariance.
func MapCertificatePairs(program *Program, target ViewContractID) [][2]ViewContractID {
	c := program.ViewContracts[target-1]
	pairs := [][2]ViewContractID{}
	active := map[[2]ViewContractID]bool{}
	var allowsNullish func(ViewContractID, ViewKind) bool
	allowsNullish = func(id ViewContractID, kind ViewKind) bool {
		if id == 0 {
			return false
		}
		c := program.ViewContracts[id-1]
		if c.Kind == kind || kind == ViewNull && c.Null || kind == ViewUndefined && c.Undefined {
			return true
		}
		for _, member := range c.Members {
			if allowsNullish(member, kind) {
				return true
			}
		}
		return false
	}
	sameStorage := func(from, to ViewContractID) bool {
		return from != 0 && to != 0 && program.ViewContracts[from-1].Of == program.ViewContracts[to-1].Of
	}

	var scalarMemberAccepts func(ViewContractID, Type, *ViewLiteral) bool
	scalarMemberAccepts = func(id ViewContractID, of Type, literal *ViewLiteral) bool {
		if id == 0 {
			return false
		}
		target := program.ViewContracts[id-1]
		if target.Unsupported != "" {
			return false
		}
		if target.Kind == ViewNullable {
			return scalarMemberAccepts(target.Element, of, literal)
		}
		if target.Kind == ViewUnion {
			for _, member := range target.Members {
				if scalarMemberAccepts(member, of, literal) {
					return true
				}
			}
			return false
		}
		if target.Kind != ViewScalar || target.Of.Present() != of.Present() {
			return false
		}
		if len(target.Allowed) == 0 {
			return true
		}
		if literal == nil {
			return false
		}
		for _, allowed := range target.Allowed {
			if allowed == *literal {
				return true
			}
		}
		return false
	}
	scalarAccepts := func(source ViewContract, target ViewContractID) bool {
		if source.Undefined && !allowsNullish(target, ViewUndefined) {
			return false
		}
		values := source.Allowed
		if len(values) == 0 && source.Of.Present() == Boolean {
			values = []ViewLiteral{{Of: Boolean}, {Of: Boolean, Boolean: true}}
		}
		if len(values) == 0 {
			return scalarMemberAccepts(target, source.Of, nil)
		}
		for _, value := range values {
			if !scalarMemberAccepts(target, source.Of, &value) {
				return false
			}
		}
		return true
	}
	var accepts func(ViewContractID, ViewContractID) bool
	accepts = func(from, to ViewContractID) bool {
		if from == 0 || to == 0 {
			return false
		}
		// Readonly arrays permit element covariance; writable arrays require both
		// directions. Every recursive step also preserves physical storage.
		source, target := program.ViewContracts[from-1], program.ViewContracts[to-1]
		if source.Kind == ViewNull || source.Kind == ViewUndefined {
			return allowsNullish(to, source.Kind)
		}
		if source.Kind == ViewNullable {
			if source.Null && !allowsNullish(to, ViewNull) || source.Undefined && !allowsNullish(to, ViewUndefined) {
				return false
			}
			if source.Element == 0 {
				return true
			}
			if target.Kind == ViewNullable {
				return accepts(source.Element, target.Element)
			}
			return accepts(source.Element, to)
		}
		if source.Kind == ViewUnion {
			for _, member := range source.Members {
				if !accepts(member, to) {
					return false
				}
			}
			return len(source.Members) != 0
		}
		if source.Kind == ViewScalar && (target.Kind == ViewUnion || target.Kind == ViewNullable && source.Undefined) {
			return scalarAccepts(source, to)
		}
		if target.Kind == ViewNullable {
			return accepts(from, target.Element)
		}
		if target.Kind == ViewUnion {
			for _, member := range target.Members {
				if accepts(from, member) {
					return true
				}
			}
			return false
		}
		if source.FixedTuple || target.FixedTuple {
			if source.FixedTuple != target.FixedTuple {
				return false
			}
			if source.TupleVariable != target.TupleVariable || source.TupleVariable && source.TupleMinimum != target.TupleMinimum || (source.TupleRest == 0) != (target.TupleRest == 0) || len(source.Tuple) != len(target.Tuple) || source.ArrayReadonly && !target.ArrayReadonly {
				return false
			}
			for index, element := range source.Tuple {
				if !sameStorage(element, target.Tuple[index]) || !accepts(element, target.Tuple[index]) || !target.ArrayReadonly && !accepts(target.Tuple[index], element) {
					return false
				}
			}
			return true
		}
		if source.Kind == ViewCallable {
			if target.Kind != ViewCallable || source.Result == 0 || target.Result == 0 || len(source.Parameters) != len(target.Parameters) {
				return false
			}
			for i, parameter := range source.Parameters {
				if !sameStorage(parameter, target.Parameters[i]) || !accepts(target.Parameters[i], parameter) {
					return false
				}
			}
			return sameStorage(source.Result, target.Result) && accepts(source.Result, target.Result)
		}
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
			return sameStorage(source.Element, target.Element) && accepts(source.Element, target.Element) && (target.ArrayReadonly || accepts(target.Element, source.Element))
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
		if sameStorage(key, c.Key) && (sameStorage(value, c.Element) || c.MapReadonly && MapReadStorageCompatible(program.ViewContracts[value-1].Of, program.ViewContracts[c.Element-1].Of)) && accepts(key, c.Key) && accepts(value, c.Element) && (c.MapReadonly || accepts(c.Key, key) && accepts(c.Element, value)) {
			pairs = append(pairs, pair)
		}
	}
	return pairs
}

// Only top-level readonly entries use these adapters. Keys and nested arrays keep
// their original calling and storage conventions.
func MapReadStorageCompatible(source, target Type) bool {
	return source == target || target == Union && (source == Number || source == Boolean || source == MaybeNumber || source == MaybeBoolean || source.IsReference() && source != Weak) || target == MaybeNumber && source == Number || target == MaybeBoolean && source == Boolean
}
