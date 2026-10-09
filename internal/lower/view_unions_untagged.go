package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// UntaggedViewMember retains the selected contract for later checked reads.
// Tags only filter candidates. A successful tag never certifies other fields.
type UntaggedViewMember struct {
	Contract ir.ViewContractID
	Tags     []UntaggedViewTag
}

type UntaggedViewTag struct {
	Field   string
	Allowed []ir.ViewLiteral
}

// UntaggedViewMembers plans an already interned object union at a read. It does
// not admit casts or certify a deferred member. Shared dispatch owns those jobs.
func UntaggedViewMembers(contracts []ir.ViewContract, union ir.ViewContractID) ([]UntaggedViewMember, error) {
	lookup := func(id ir.ViewContractID) (ir.ViewContract, error) {
		if id <= 0 || int(id) > len(contracts) {
			return ir.ViewContract{}, fmt.Errorf("unavailable checked-view contract %d", id)
		}
		return contracts[int(id)-1], nil
	}
	root, err := lookup(union)
	if err != nil {
		return nil, err
	}
	if root.Kind != ir.ViewUnion || len(root.Members) == 0 {
		return nil, fmt.Errorf("%s is not a nonempty object union", root.Name)
	}
	members := make([]UntaggedViewMember, 0, len(root.Members))
	for _, id := range root.Members {
		contract, err := lookup(id)
		if err != nil {
			return nil, err
		}
		if contract.Kind != ir.ViewObject || contract.Of != ir.Object {
			return nil, fmt.Errorf("union member %s needs its owning family adapter", contract.Name)
		}
		member := UntaggedViewMember{Contract: id}
		for _, field := range contract.Fields {
			// An absent optional tag is legal, so it cannot exclude this candidate.
			if field.Optional {
				continue
			}
			tag, err := lookup(field.Contract)
			if err != nil {
				return nil, err
			}
			if tag.Kind != ir.ViewScalar || len(tag.Allowed) == 0 {
				continue
			}
			for _, literal := range tag.Allowed {
				if literal.Of != ir.Number && literal.Of != ir.String && literal.Of != ir.Boolean {
					return nil, fmt.Errorf("unsupported tag contract %s.%s", contract.Name, field.Name)
				}
			}
			member.Tags = append(member.Tags, UntaggedViewTag{Field: field.Name, Allowed: append([]ir.ViewLiteral(nil), tag.Allowed...)})
		}
		members = append(members, member)
	}
	return members, nil
}

// Shared lazy dispatch calls this only for a demanded union read. Tagged
// members defer unread descendants. Field-only membership needs a complete,
// recursive plain-data contract; unavailable adapters remain named obligations.
func (l *lowering) supportsUntaggedRead(root ir.ViewContract) bool {
	seen := map[ir.ViewContractID]bool{}
	var supported func(ir.ViewContractID) bool
	supported = func(id ir.ViewContractID) bool {
		if id <= 0 || int(id) > len(l.result.ViewContracts) {
			return false
		}
		contract := l.result.ViewContracts[id-1]
		if (contract.Unsupported != "" && contract.Unsupported != "untagged object union") || contract.Nominal != "" {
			return false
		}
		if contract.Kind == ir.ViewScalar {
			return contract.Of == ir.Number || contract.Of == ir.String || contract.Of == ir.Boolean || contract.Of == ir.MaybeNumber || contract.Of == ir.MaybeBoolean
		}
		if contract.Kind == ir.ViewCallable {
			return contract.Result != 0 || contract.DiscardResult
		}
		if contract.Kind == ir.ViewUndefined {
			return true
		}
		if seen[id] {
			return true
		}
		seen[id] = true
		defer delete(seen, id)
		if contract.Kind == ir.ViewArray {
			// Array membership checks represented elements; payload reads stay lazy.
			return contract.Element != 0 && supported(contract.Element)
		}
		if contract.Kind == ir.ViewUnion {
			if len(contract.Members) == 0 {
				return false
			}
			for _, member := range contract.Members {
				if !supported(member) {
					return false
				}
			}
			return true
		}
		if contract.Kind != ir.ViewObject {
			return false
		}
		tagged := false
		for _, field := range contract.Fields {
			if field.Optional || field.Contract <= 0 || int(field.Contract) > len(l.result.ViewContracts) {
				continue
			}
			child := l.result.ViewContracts[field.Contract-1]
			tagged = tagged || child.Kind == ir.ViewScalar && (len(child.Allowed) != 0 || field.Name == "kind" && (child.Of == ir.Number || child.Of == ir.String || child.Of == ir.Boolean))
		}
		if tagged {
			return true
		}
		for _, field := range contract.Fields {
			if !supported(field.Contract) {
				return false
			}
		}
		return true
	}
	if root.Kind != ir.ViewUnion || len(root.Members) == 0 {
		return false
	}
	for _, id := range root.Members {
		if !supported(id) {
			return false
		}
	}
	return true
}

// untaggedArrayElement joins only array members' logical element contracts.
// It never chooses an arbitrary member's physical representation.
func (l *lowering) untaggedArrayElement(target *checker.Type) *checker.Type {
	if element := l.viewArrayElementType(target); element != nil {
		return element
	}
	if target.Flags()&checker.TypeFlagsUnion == 0 {
		return nil
	}
	var elements []*checker.Type
	for _, member := range target.Types() {
		element := l.viewArrayElementType(member)
		if element == nil {
			return nil
		}
		elements = append(elements, element)
	}
	if len(elements) == 0 {
		return nil
	}
	return l.checker.GetUnionType(elements)
}

func (l *lowering) completeUntaggedRecursiveContracts() {
	// Optional array descriptors may be copied while their element is reserved.
	// Only a completed canonical descriptor can supply the missing element.
	for index, contract := range l.result.ViewContracts {
		if contract.Kind == ir.ViewArray && contract.Element == 0 && contract.ObjectPresent > 0 {
			canonical := l.result.ViewContracts[contract.ObjectPresent-1]
			if canonical.Kind == ir.ViewArray && canonical.Element > 0 {
				l.result.ViewContracts[index].Element = canonical.Element
			}
		}
	}

	for index, contract := range l.result.ViewContracts {
		if contract.Unsupported == "untagged object union" && l.supportsUntaggedRead(contract) {
			l.result.ViewContracts[index].Unsupported = ""
		}
	}
}

// Callable union membership uses each producer-certified member signature;
// the checker's synthesized union signature is not an implementation contract.
func (l *lowering) prepareUntaggedCallableUnionRead(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	id := l.result.ViewContractTypes[int(target.Id())]
	if id == 0 {
		id = ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{})
		l.result.ViewContractTypes[int(target.Id())] = id
	}
	contract := ir.ViewContract{Kind: ir.ViewUnion, Of: ir.Closure, Name: l.checker.TypeToString(target)}
	for _, member := range target.Types() {
		if !l.callableViewContract(member) {
			return 0, l.notYet(node, "a callable union member without a producer signature")
		}
		child, err := l.prepareUntaggedCallableMember(node, member)
		if err != nil {
			return 0, err
		}
		if l.untaggedCallableTargets == nil {
			l.untaggedCallableTargets = map[ir.ViewContractID]*checker.Type{}
		}
		l.untaggedCallableTargets[child] = member
		l.result.ViewContracts[child-1].Unsupported = ""
		contract.Members = append(contract.Members, child)
	}
	l.result.ViewContracts[id-1] = contract
	return id, nil
}

// Prepare nested callable obligations using lane 5's existing descriptor builder.
// Only represented fixed scalar signatures are certified; errors stay lazy.
func (l *lowering) prepareUntaggedStructuralRead(node *ast.Node, target *checker.Type, seen map[*checker.Type]bool) {
	if target == nil || seen[target] {
		return
	}
	seen[target] = true
	if l.callableViewContract(target) {
		if l.untaggedCallableShape(target) {
			if id, err := l.prepareUntaggedCallableMember(node, target); err == nil {
				l.result.ViewContracts[id-1].Unsupported = ""
			}
		}
		return
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			l.prepareUntaggedStructuralRead(node, member, seen)
		}
		return
	}
	if element := l.untaggedArrayElement(target); element != nil {
		l.prepareUntaggedStructuralRead(node, element, seen)
		return
	}
	if target.Flags()&checker.TypeFlagsObject == 0 {
		return
	}
	for _, field := range l.checker.GetPropertiesOfType(target) {
		l.prepareUntaggedStructuralRead(node, l.checker.GetTypeOfSymbol(field), seen)
	}
}

// Reuse closureRecords' checker proof and lane 5's code-identity registry. Exact
// logical identity is conservative: physical Closure/Object tags alone do not
// certify higher-order arguments or object results.
func (l *lowering) certifyUntaggedCallableProducers() {
	for id, target := range l.untaggedCallableTargets {
		contract := &l.result.ViewContracts[id-1]
		contract.ProducerCertified = true
		contract.Functions = nil
		for _, producer := range l.closureRecords {
			if checker.Checker_isTypeIdenticalTo(l.checker, producer.proven, target) {
				contract.Functions = append(contract.Functions, producer.function)
			}
		}
	}
}
func certifiedUntaggedCallableRead(program *ir.Program, id ir.ViewContractID) bool {
	if id <= 0 || int(id) > len(program.ViewContracts) {
		return false
	}
	contract := program.ViewContracts[id-1]
	if contract.Unsupported != "" {
		return false
	}
	if contract.Kind == ir.ViewCallable {
		return contract.ProducerCertified
	}
	if contract.Kind != ir.ViewUnion || contract.Of != ir.Closure || len(contract.Members) == 0 {
		return false
	}
	for _, member := range contract.Members {
		if !certifiedUntaggedCallableRead(program, member) {
			return false
		}
	}
	return true
}
