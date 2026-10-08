package ir

// RecursiveIntersectionObjects returns only the object descriptors reachable
// from this read. Optional copies resolve to their completed canonical object.
func RecursiveIntersectionObjects(program *Program, root ViewContractID) []ViewContractID {
	ids := []ViewContractID{}
	seen := map[ViewContractID]bool{}
	var visit func(ViewContractID)
	visit = func(id ViewContractID) {
		contract := program.ViewContracts[id-1]
		if contract.ObjectPresent != 0 {
			visit(contract.ObjectPresent)
			return
		}
		if contract.Kind != ViewObject || seen[id] || contract.Unsupported != "" && contract.Unsupported != "recursive intersection payload" {
			return
		}
		seen[id] = true
		ids = append(ids, id)
		for _, field := range contract.Fields {
			visit(field.Contract)
		}
	}
	visit(root)
	return ids
}

func RecursiveIntersectionPresent(program *Program, id ViewContractID) ViewContractID {
	for program.ViewContracts[id-1].ObjectPresent != 0 {
		id = program.ViewContracts[id-1].ObjectPresent
	}
	return id
}

// BoundedField says how a bounded intersection walk treats one field's contract.
type BoundedField int

const (
	// BoundedDeferred leaves the field to its own checked read: an unsupported
	// family refuses there at compile time, and a mixed scalar union is checked there.
	BoundedDeferred BoundedField = iota
	// BoundedKind checks presence, runtime kind and literal membership only.
	// Array elements, Map entries and callable signatures keep their own reads.
	BoundedKind
	// BoundedObject checks the kind, then every field of the canonical object.
	BoundedObject
	// BoundedArms checks the kind, then selects a union arm by its literal tag.
	BoundedArms
	// BoundedRefused is a shape the bounded walk cannot validate.
	BoundedRefused
)

// BoundedIntersectionField classifies a field contract for the bounded walk and
// returns the descriptor to walk next. walkable names refused descriptors the
// caller is still deciding about; nil treats every refusal as deferred.
func BoundedIntersectionField(program *Program, id ViewContractID, walkable func(ViewContractID) bool) (BoundedField, ViewContractID) {
	c := program.ViewContracts[id-1]
	if c.Unsupported != "" && (walkable == nil || !walkable(id)) {
		return BoundedDeferred, 0
	}
	if c.ObjectPresent != 0 {
		present := RecursiveIntersectionPresent(program, id)
		return BoundedIntersectionField(program, present, walkable)
	}
	switch c.Kind {
	case ViewUnknown, ViewUndefined, ViewNull, ViewDictionary:
		return BoundedDeferred, 0
	case ViewScalar:
		if c.Of == Union {
			return BoundedDeferred, 0
		}
		return BoundedKind, 0
	case ViewObject:
		if c.Nominal != "" || c.Of != Object {
			return BoundedRefused, 0
		}
		return BoundedObject, id
	case ViewUnion:
		switch c.Of {
		case Union:
			return BoundedDeferred, 0
		case Object:
			if tag, _ := IntersectionUnionArms(program, id); tag != "" {
				return BoundedArms, id
			}
			if tag, _ := UnionDiscriminant(program, id); tag != "" {
				return BoundedArms, id
			}
			// Supported untagged descendants retain presence and object kind.
			// Their member selection is checked when the descendant is read.
			return BoundedKind, 0
		}
		return BoundedKind, 0
	case ViewNullable:
		if c.Null || !c.Undefined || c.Element == 0 {
			return BoundedRefused, 0
		}
		return BoundedIntersectionField(program, c.Element, walkable)
	case ViewMap:
		if c.Of != Map {
			return BoundedRefused, 0
		}
		return BoundedKind, 0
	case ViewArray:
		if c.Of != Array {
			return BoundedRefused, 0
		}
		return BoundedKind, 0
	case ViewCallable:
		if c.Of != Closure {
			return BoundedRefused, 0
		}
		return BoundedKind, 0
	}
	return BoundedRefused, 0
}

// BoundedIntersectionContracts returns the object and selected-union descriptors a
// bounded walk from root can reach, root first, or false if any is refused.
func BoundedIntersectionContracts(program *Program, root ViewContractID, walkable func(ViewContractID) bool) ([]ViewContractID, bool) {
	ids := []ViewContractID{}
	seen := map[ViewContractID]bool{}
	valid := true
	var visit func(ViewContractID)
	visit = func(id ViewContractID) {
		if seen[id] || !valid {
			return
		}
		seen[id] = true
		ids = append(ids, id)
		c := program.ViewContracts[id-1]
		children := []ViewContractID{}
		if c.Kind == ViewUnion {
			tag, arms := IntersectionUnionArms(program, id)
			if tag == "" {
				return // A discriminant-only union checks its tag; members keep their own reads.
			}
			children = arms
		} else {
			for _, field := range c.Fields {
				children = append(children, field.Contract)
			}
		}
		for _, child := range children {
			switch kind, next := BoundedIntersectionField(program, child, walkable); kind {
			case BoundedRefused:
				valid = false
			case BoundedObject, BoundedArms:
				visit(next)
			}
		}
	}
	visit(RecursiveIntersectionPresent(program, root))
	return ids, valid
}

// IntersectionUnionArms selects the arms of an object union by one required
// literal field whose sets are disjoint. Arms with identical field obligations
// are coalesced, keeping the intersection, since the checker may list Identifier
// and LeftHandSideExpression & Identifier separately.
func IntersectionUnionArms(program *Program, id ViewContractID) (string, []ViewContractID) {
	root := program.ViewContracts[id-1]
	if root.Kind != ViewUnion || root.Of != Object || len(root.Members) == 0 {
		return "", nil
	}
	arms := []ViewContractID{}
	for _, candidate := range root.Members {
		candidate = RecursiveIntersectionPresent(program, candidate)
		own := program.ViewContracts[candidate-1]
		if own.Kind != ViewObject || own.Of != Object || own.Nominal != "" {
			return "", nil
		}
		duplicate := false
		for i, existing := range arms {
			prior := program.ViewContracts[existing-1]
			same := len(own.Fields) == len(prior.Fields)
			for _, field := range own.Fields {
				found := false
				for _, other := range prior.Fields {
					found = found || field.Name == other.Name && field.Optional == other.Optional && field.Readonly == other.Readonly && boundedSame(program, field.Contract, other.Contract)
				}
				same = same && found
			}
			if same {
				duplicate = true
				if own.Intersection {
					arms[i] = candidate
				}
				break
			}
		}
		if !duplicate {
			arms = append(arms, candidate)
		}
	}
	for _, field := range root.Fields {
		tag := program.ViewContracts[field.Contract-1]
		if field.Optional || tag.Kind != ViewScalar || tag.Of != Number && tag.Of != String && tag.Of != Boolean {
			continue
		}
		used := map[ViewLiteral]bool{}
		valid := true
		for _, arm := range arms {
			allowed := armLiterals(program, arm, field.Name, tag.Of)
			if len(allowed) == 0 {
				valid = false
			}
			for _, literal := range allowed {
				valid = valid && !used[literal]
				used[literal] = true
			}
		}
		if valid {
			return field.Name, arms
		}
	}
	return "", nil
}

// Two field obligations are the same to the bounded walk when they name one
// canonical descriptor, or are arrays it checks by kind alone: an optional copy
// or a reservation copy of an array must not split identical arms.
func boundedSame(program *Program, a, b ViewContractID) bool {
	if a == b {
		return true
	}
	left, right := program.ViewContracts[a-1], program.ViewContracts[b-1]
	if left.Undefined != right.Undefined || left.Name != right.Name || left.Unsupported != right.Unsupported {
		return false
	}
	if RecursiveIntersectionPresent(program, a) == RecursiveIntersectionPresent(program, b) {
		return true
	}
	return left.Kind == ViewArray && right.Kind == ViewArray && left.Of == right.Of && len(left.Allowed) == 0 && len(right.Allowed) == 0
}

// IntersectionArmLiterals returns an arm's required literal set for the tag.
func IntersectionArmLiterals(program *Program, arm ViewContractID, tag string) []ViewLiteral {
	for _, field := range program.ViewContracts[arm-1].Fields {
		if field.Name == tag {
			return program.ViewContracts[field.Contract-1].Allowed
		}
	}
	return nil
}

func armLiterals(program *Program, arm ViewContractID, name string, of Type) []ViewLiteral {
	for _, field := range program.ViewContracts[arm-1].Fields {
		if field.Name != name {
			continue
		}
		part := program.ViewContracts[field.Contract-1]
		if field.Optional || part.Kind != ViewScalar || part.Of != of {
			return nil
		}
		return part.Allowed
	}
	return nil
}

// UnionDiscriminant is the plain object union's checked discriminant: its first
// field with a finite literal set, as the union read itself checks it.
func UnionDiscriminant(program *Program, id ViewContractID) (string, ViewContractID) {
	for _, field := range program.ViewContracts[id-1].Fields {
		tag := program.ViewContracts[field.Contract-1]
		if tag.Kind == ViewScalar && len(tag.Allowed) != 0 {
			return field.Name, field.Contract
		}
	}
	return "", 0
}
