package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Lane 1 calls this hook only after wiring member selection at reads.
var viewMixedUnionContractHook viewContractHook = internMixedUnionViewContract

// Intern a complete union graph using the shared member builder. This is not an
// admission hook until lane 1 wires selection and representation conversion.
// Reserve before recursion, but discard every newly reserved id on failure:
// a later lookup must not mistake a partial graph for a complete contract.
func internMixedUnionViewContract(l *lowering, node *ast.Node, target *checker.Type, build viewContractBuilder) (id ir.ViewContractID, err error) {
	if target.Flags()&checker.TypeFlagsUnion == 0 {
		return 0, l.notYet(node, "a mixed union contract for a non-union type")
	}
	of, known := l.representation(target)
	if !known {
		return 0, l.notYet(node, "checked union representation for "+l.checker.TypeToString(target))
	}
	if build == nil {
		return 0, l.notYet(node, "a checked union without a recursive member builder")
	}
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	if cached := l.result.ViewContractTypes[int(target.Id())]; cached != 0 {
		return cached, nil
	}
	start := len(l.result.ViewContracts)
	id = ir.ViewContractID(start + 1)
	contract := ir.ViewContract{Kind: ir.ViewUnion, Name: l.checker.TypeToString(target), Of: of}
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id
	defer func() {
		if err == nil {
			return
		}
		l.result.ViewContracts = l.result.ViewContracts[:start]
		for key, value := range l.result.ViewContractTypes {
			if int(value) > start {
				delete(l.result.ViewContractTypes, key)
			}
		}
		id = 0
	}()
	for _, member := range target.Types() {
		var child ir.ViewContractID
		child, err = build(member)
		if err != nil {
			return 0, err
		}
		if child <= 0 || int(child) > len(l.result.ViewContracts) || l.result.ViewContracts[int(child)-1].Kind == ir.ViewUnknown {
			return 0, l.notYet(node, "an unavailable checked union member contract for "+l.checker.TypeToString(member))
		}
		contract.Members = append(contract.Members, child)
	}
	l.result.ViewContracts[int(id)-1] = contract
	return id, nil
}

// Approved phantom-void alternatives in a string representation use the scalar
// reader's undefined payload certificate, independently of optional absence.
func (l *lowering) viewBrandedStringUndefined(target *checker.Type) bool {
	if l.includesNull(target) {
		return false
	}
	of, known := l.representation(target)
	if !known || of != ir.String {
		return false
	}
	if l.phantomUndefined(target) {
		return true
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if l.phantomUndefined(member) {
				return true
			}
		}
	}
	return false
}
