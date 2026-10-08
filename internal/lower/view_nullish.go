package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Null and undefined are alternatives, not permission to reinterpret a present
// payload. Reserve before recursing so nullable recursive objects terminate.
func (l *lowering) nullishViewContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	target = l.concrete(target)
	id := ir.ViewContractID(len(l.result.ViewContracts) + 1)
	contract := ir.ViewContract{Kind: ir.ViewNullable, Name: l.checker.TypeToString(target), Null: l.includesNull(target), Undefined: l.includesUndefined(target)}
	contract.Of, _ = l.representation(target)
	l.result.ViewContracts = append(l.result.ViewContracts, contract)
	l.result.ViewContractTypes[int(target.Id())] = id
	present := l.checker.GetNonNullableType(target)
	if present.Flags()&checker.TypeFlagsNever == 0 {
		child, err := l.viewContract(node, present)
		if err != nil {
			return 0, err
		}
		contract.Element = child
		presentContract := l.result.ViewContracts[child-1]
		contract.Unsupported = presentContract.Unsupported
		if contract.Unsupported == "" && presentContract.Kind == ir.ViewUnion {
			contract.Unsupported = nullableUnionUnsupported(l.result, presentContract)
		}
	}
	l.result.ViewContracts[id-1] = contract
	return id, nil
}

// Logical present kinds come from checker types, never physical slot evidence.
// The runtime independently checks the actual producer tag before conversion.
func (l *lowering) nullishViewKinds(target *checker.Type) uint32 {
	target = l.concrete(target)
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		var mask uint32
		for _, member := range target.Types() {
			mask |= l.nullishViewKinds(member)
		}
		return mask
	}
	if target.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
		return 0
	}
	of, known := l.representation(target)
	if !known || of == 0 || of == ir.Union {
		return 0
	}
	// Records and fixed objects have the same logical heap kind; storage stays certified separately.
	if of == ir.Record {
		of = ir.Object
	}
	return 1 << of
}

// Distinct runtime kinds select a member without reading its payload. Object
// unions with a common finite tag use the existing tag adapter. Ambiguous array
// or opaque callable alternatives retain a named read refusal.
func nullableUnionUnsupported(program *ir.Program, contract ir.ViewContract) string {
	kinds := map[ir.Type]int{}
	for _, id := range contract.Members {
		member := program.ViewContracts[id-1]
		if member.Kind == ir.ViewUndefined || member.Kind == ir.ViewNull {
			continue
		}
		if member.Unsupported != "" {
			return member.Unsupported
		}
		switch member.Kind {
		case ir.ViewScalar, ir.ViewObject, ir.ViewArray, ir.ViewMap:
		default:
			return "nullable union member selection"
		}
		if member.Kind != ir.ViewScalar {
			kinds[member.Of]++
		}
	}
	for of, count := range kinds {
		if count > 1 && !(of == ir.Object && contract.Of == ir.Object) {
			return "nullable union member selection"
		}
	}
	return ""
}
