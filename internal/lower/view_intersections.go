package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The shared dispatcher must store these ids as a conjunction, never a union.
// This helper does not enable cast admission or certify any runtime payload.
func (l *lowering) viewIntersectionContracts(node *ast.Node, target *checker.Type, build viewContractBuilder) ([]ir.ViewContractID, error) {
	if target.Flags()&checker.TypeFlagsIntersection == 0 || l.phantomBase(target) != nil {
		return nil, l.notYet(node, "a nonprimitive intersection checked-view contract")
	}
	if build == nil {
		return nil, l.notYet(node, "an intersection checked view without a member builder")
	}
	var members []ir.ViewContractID
	for _, part := range target.Types() {
		if l.viewIntersectionPhantom(part) {
			continue
		}
		child, err := build(part)
		if err != nil {
			return nil, err
		}
		if child <= 0 || int(child) > len(l.result.ViewContracts) || l.result.ViewContracts[int(child)-1].Kind == ir.ViewUnknown {
			return nil, l.notYet(node, "an unavailable intersection member contract for "+l.checker.TypeToString(part))
		}
		members = append(members, child)
	}
	if len(members) == 0 {
		return nil, l.notYet(node, "an intersection without a runtime constituent")
	}
	return members, nil
}

// Reuse the brand lane's field rule. An empty object is not evidence of a brand,
// and callable/indexed objects must retain their runtime obligations.
func (l *lowering) viewIntersectionPhantom(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsObject == 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 0 || len(l.checker.GetIndexInfosOfType(target)) != 0 {
		return false
	}
	fields := l.checker.GetPropertiesOfType(target)
	if len(fields) == 0 {
		return false
	}
	for _, field := range fields {
		if !phantomField(l.checker.GetTypeOfSymbol(field), field.Flags&ast.SymbolFlagsOptional != 0) {
			return false
		}
	}
	return true
}

// Structural intersections share object storage, but retain every constituent's
// read obligations. Checker properties resolve duplicate fields by intersection,
// so their child descriptor retains the conjunction rather than picking an arm.
func (l *lowering) structuralViewIntersection(target *checker.Type) bool {
	if target.Flags()&checker.TypeFlagsIntersection == 0 || l.phantomBase(target) != nil {
		return false
	}
	runtimeParts := 0
	for _, part := range target.Types() {
		if l.viewIntersectionPhantom(part) {
			continue
		}
		runtimeParts++
		if part.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(part) || l.checker.IsArrayType(part) || checker.IsTupleType(part) || l.callableViewContract(part) || len(l.checker.GetIndexInfosOfType(part)) != 0 {
			return false
		}
	}
	return runtimeParts > 0
}

// The lazy dispatcher reserves this descriptor before walking descendants.
// Unsupported children remain obligations at their own reads, never cast gates.
func (l *lowering) internStructuralViewIntersection(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	if !l.structuralViewIntersection(target) {
		return 0, l.notYet(node, "a structural object intersection view")
	}
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	if id := l.result.ViewContractTypes[int(target.Id())]; id != 0 {
		return id, nil
	}
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	contract := ir.ViewContract{Intersection: true, Kind: ir.ViewObject, Of: ir.Object, Name: l.checker.TypeToString(target)}
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id
	runtimeFields := map[string]bool{}
	for _, part := range target.Types() {
		if l.viewIntersectionPhantom(part) {
			continue
		}
		child, err := l.viewContract(node, part)
		if err != nil {
			return 0, err
		}
		contract.Members = append(contract.Members, child)
		for _, field := range l.checker.GetPropertiesOfType(part) {
			runtimeFields[field.Name] = true
		}
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		if !runtimeFields[field.Name] {
			continue
		}
		child, err := l.viewContract(node, l.checker.GetTypeOfSymbol(field))
		if err != nil {
			return 0, err
		}
		contract.Fields = append(contract.Fields, ir.ViewFieldContract{Name: field.Name, Contract: child, Optional: field.Flags&ast.SymbolFlagsOptional != 0, Readonly: l.checker.IsReadonlySymbol(field)})
	}
	if l.recursiveIntersectionPayload(target) {
		contract.Unsupported = "recursive intersection payload"
	}
	l.result.ViewContracts[id-1] = contract
	return id, nil
}

// Do not admit combinations the production matcher cannot validate yet. This
// metadata is consumed by lazy demand, so unread casts still remain admitted.
func (l *lowering) viewIntersectionReadFamily(id ir.ViewContractID, target *checker.Type) string {
	root := l.result.ViewContracts[id-1]
	if root.Kind == ir.ViewUnion {
		hasIntersection := false
		for _, member := range root.Members {
			hasIntersection = hasIntersection || l.result.ViewContracts[member-1].Intersection
		}
		if tag := l.viewIntersectionUnionTagCandidate(id, hasIntersection || root.Unsupported == "untagged object union"); tag != "" && !l.recursiveIntersectionPayload(target) {
			l.result.ViewContracts[id-1].IntersectionTag = tag
			if root.Unsupported == "untagged object union" {
				l.result.ViewContracts[id-1].Unsupported = ""
			}
			return ""
		}
		for _, member := range root.Members {
			if l.result.ViewContracts[member-1].Intersection {
				for _, part := range target.Types() {
					if l.recursiveIntersectionPayload(part) {
						return "union intersection"
					}
				}
				if tag := l.viewIntersectionUnionTag(id); tag != "" {
					l.result.ViewContracts[id-1].IntersectionTag = tag
					return ""
				}
				return "union intersection"
			}
		}
	}
	if root.Intersection && l.recursiveIntersectionPayload(target) && l.recursiveIntersectionSupported(id, map[ir.ViewContractID]bool{}) {
		l.result.ViewContracts[id-1].Unsupported = ""
		l.result.ViewContracts[id-1].IntersectionRecursive = true
		return ""
	}
	if !root.Intersection {
		return ""
	}
	active := map[ir.ViewContractID]bool{}
	var visit func(ir.ViewContractID) string
	visit = func(id ir.ViewContractID) string {
		contract := l.result.ViewContracts[id-1]
		// Descendant unsupported families keep their own read-site obligations.
		if contract.Unsupported != "" {
			return ""
		}
		if active[id] {
			return "recursive intersection payload"
		}
		active[id] = true
		defer delete(active, id)
		switch contract.Kind {
		case ir.ViewScalar:
			if contract.Of == ir.Union {
				return "mixed intersection payload"
			}
		case ir.ViewObject:
			for _, field := range contract.Fields {
				if family := visit(field.Contract); family != "" {
					return family
				}
			}
		case ir.ViewCallable:
			return "" // Kind now; the shared callable proof is demanded by member reads.
		case ir.ViewUnknown:
			return "" // Lazy demand never treats Unknown as a read certificate.
		default:
			return "compound intersection payload"
		}
		return ""
	}
	return visit(id)
}

// Inspect checker types too: an optional recursive descriptor can be copied
// while its canonical parent is still being reserved, before Fields are filled.
func (l *lowering) recursiveIntersectionPayload(target *checker.Type) bool {
	active := map[*checker.Type]bool{}
	complete := map[*checker.Type]bool{}
	var visit func(*checker.Type) bool
	visit = func(target *checker.Type) bool {
		target = l.checker.GetNonNullableType(target)
		if target.Flags()&checker.TypeFlagsUnion != 0 {
			for _, part := range target.Types() {
				if visit(part) {
					return true
				}
			}
			return false
		}
		if target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || l.checker.IsArrayType(target) || l.callableViewContract(target) {
			return false
		}
		if active[target] {
			return true
		}
		if complete[target] {
			return false
		}
		active[target] = true
		defer delete(active, target)
		for _, field := range l.checker.GetPropertiesOfType(target) {
			if visit(l.checker.GetTypeOfSymbol(field)) {
				return true
			}
		}
		complete[target] = true
		return false
	}
	return visit(target)
}

// A shared tag may select an arm only if its literal sets are disjoint and
// every selected arm is in the finite runtime family already implemented.
func (l *lowering) viewIntersectionUnionTag(id ir.ViewContractID) string {
	root := l.result.ViewContracts[id-1]
	if root.Kind != ir.ViewUnion || root.Of != ir.Object || len(root.Members) == 0 {
		return ""
	}
	var finite func(ir.ViewContractID, map[ir.ViewContractID]bool) bool
	finite = func(id ir.ViewContractID, active map[ir.ViewContractID]bool) bool {
		c := l.result.ViewContracts[id-1]
		if c.Unsupported != "" {
			return false
		}
		if active[id] {
			return false
		}
		active[id] = true
		defer delete(active, id)
		if c.Kind == ir.ViewScalar {
			return c.Of != ir.Union
		}
		if c.Kind != ir.ViewObject {
			return false
		}
		for _, f := range c.Fields {
			child := l.result.ViewContracts[f.Contract-1]
			if child.Unsupported != "" {
				continue
			}
			if !finite(f.Contract, active) {
				return false
			}
		}
		return true
	}
	// The checker may retain Identifier and LHS & Identifier as two arms.
	// Coalesce only byte-identical field obligations; overlapping unequal arms
	// still require a general selector and remain refused.
	members := []ir.ViewContractID{}
	for _, candidate := range root.Members {
		own := l.result.ViewContracts[candidate-1]
		duplicate := false
		for i, existing := range members {
			prior := l.result.ViewContracts[existing-1]
			same := own.Kind == prior.Kind && own.Unsupported == prior.Unsupported && len(own.Fields) == len(prior.Fields)
			for _, field := range own.Fields {
				found := false
				for _, other := range prior.Fields {
					if field == other {
						found = true
					}
				}
				same = same && found
			}
			if same {
				duplicate = true
				if own.Intersection {
					members[i] = candidate
				}
				break
			}
		}
		if !duplicate {
			members = append(members, candidate)
		}
	}
	root.Members = members
	for _, member := range root.Members {
		if !finite(member, map[ir.ViewContractID]bool{}) {
			return ""
		}
	}
	for _, field := range root.Fields {
		tag := l.result.ViewContracts[field.Contract-1]
		if field.Optional || tag.Kind != ir.ViewScalar || tag.Of != ir.Number && tag.Of != ir.String && tag.Of != ir.Boolean {
			continue
		}
		used := map[ir.ViewLiteral]bool{}
		valid := true
		for _, member := range root.Members {
			c := l.result.ViewContracts[member-1]
			found := false
			for _, own := range c.Fields {
				if own.Name != field.Name {
					continue
				}
				part := l.result.ViewContracts[own.Contract-1]
				if own.Optional || part.Kind != ir.ViewScalar || part.Of != tag.Of || len(part.Allowed) == 0 {
					break
				}
				found = true
				for _, literal := range part.Allowed {
					if used[literal] {
						valid = false
					}
					used[literal] = true
				}
			}
			if !found {
				valid = false
			}
		}
		if valid {
			l.result.ViewContracts[id-1].Members = root.Members
			return field.Name
		}
	}
	return ""
}

func (l *lowering) viewIntersectionUnionTagCandidate(id ir.ViewContractID, eligible bool) string {
	if !eligible {
		return ""
	}
	return l.viewIntersectionUnionTag(id)
}

// Runtime recursion uses canonical optional descriptors, never a reservation copy.
// Nullable, union, array and callable recursive payloads remain demand refusals.
func (l *lowering) recursiveIntersectionSupported(id ir.ViewContractID, seen map[ir.ViewContractID]bool) bool {
	c := l.result.ViewContracts[id-1]
	if c.ObjectPresent != 0 {
		return l.recursiveIntersectionSupported(c.ObjectPresent, seen)
	}
	if seen[id] {
		return true
	}
	seen[id] = true
	if c.Unsupported != "" && c.Unsupported != "recursive intersection payload" {
		return true
	} // descendant remains lazy
	if c.Kind == ir.ViewScalar {
		return c.Of != ir.Union
	}
	if c.Kind != ir.ViewObject || c.Nominal != "" {
		return false
	}
	for _, f := range c.Fields {
		if !l.recursiveIntersectionSupported(f.Contract, seen) {
			return false
		}
	}
	return true
}

// A resolved intersection read emits its own runtime checks. A same-named
// unsupported field in a wider source carrier cannot replace that obligation.
// Unknown or unsupported read contracts still use the conservative fallback.
func (l *lowering) viewIntersectionReadChecks(id ir.ViewContractID) bool {
	if id == 0 {
		return false
	}
	c := l.result.ViewContracts[id-1]
	if c.Unsupported != "" {
		return false
	}
	if c.Kind == ir.ViewNullable {
		return l.viewIntersectionReadChecks(c.Element)
	}
	return c.Intersection || c.IntersectionTag != ""
}

// Lane 7's demand refusals were decided while descriptors were still being
// reserved. Once every descriptor is complete, a refused intersection or tagged
// union whose whole reachable graph the bounded walk validates is admitted.
// A refused descriptor it reaches is walked only if it is admitted too, so the
// emitters, which defer every remaining refusal, walk exactly what was proven.
func (l *lowering) finishBoundedIntersections() {
	program := l.result
	pending := func(family string) bool {
		switch family {
		case "union intersection", "recursive intersection payload", "mixed intersection payload", "compound intersection payload":
			return true
		}
		return false
	}
	tags := map[ir.ViewContractID]string{}
	candidates := map[ir.ViewContractID]bool{}
	for i, c := range program.ViewContracts {
		id := ir.ViewContractID(i + 1)
		if !pending(c.Unsupported) || c.ObjectPresent != 0 {
			continue
		}
		switch {
		case c.Kind == ir.ViewObject && c.Intersection:
			candidates[id] = true
		case c.Kind == ir.ViewUnion:
			if tag, _ := ir.IntersectionUnionArms(program, id); tag != "" {
				candidates[id] = true
				tags[id] = tag
			}
		}
	}
	walkable := func(id ir.ViewContractID) bool {
		c := program.ViewContracts[id-1]
		return pending(c.Unsupported) && candidates[ir.RecursiveIntersectionPresent(program, id)]
	}
	for changed := true; changed; {
		changed = false
		for id := range candidates {
			if _, valid := ir.BoundedIntersectionContracts(program, id, walkable); !valid {
				delete(candidates, id)
				changed = true
			}
		}
	}
	for i := range program.ViewContracts {
		id := ir.ViewContractID(i + 1)
		if !walkable(id) {
			continue
		}
		c := &program.ViewContracts[i]
		c.Unsupported = ""
		c.IntersectionBounded = true
		c.IntersectionRecursive = false
		if tag := tags[id]; tag != "" {
			c.IntersectionTag = tag
		}
	}
}
