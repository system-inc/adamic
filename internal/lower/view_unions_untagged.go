package lower

import (
	"fmt"

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
		if contract.Kind == ir.ViewUndefined {
			return true
		}
		if seen[id] {
			return true
		}
		seen[id] = true
		defer delete(seen, id)
		if contract.Kind == ir.ViewArray {
			return supported(contract.Element)
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
	if root.Kind != ir.ViewUnion || root.Of != ir.Object || len(root.Members) == 0 {
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
	if l.checker.IsArrayType(target) {
		return l.checker.GetElementTypeOfArrayType(target)
	}
	if target.Flags()&checker.TypeFlagsUnion == 0 {
		return nil
	}
	var elements []*checker.Type
	for _, member := range target.Types() {
		if !l.checker.IsArrayType(member) {
			return nil
		}
		element := l.checker.GetElementTypeOfArrayType(member)
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
