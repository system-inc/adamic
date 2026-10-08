package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Upstream numeric IDs have a required any brand. A number cannot carry an
// own brand property; the existing prototype-name inventory proves this name
// absent, so its value is undefined, which inhabits any. Keep the number kind
// check. This does not make any usable: a demanded any-valued read still refuses.
// Written Adamic declarations retain phantomRefusal's void-only rule.
func (l *lowering) tupleNumericCarrierField(primitive *checker.Type, field *ast.Symbol) bool {
	if primitive.Flags()&checker.TypeFlagsNumberLike == 0 || field.Flags&ast.SymbolFlagsOptional != 0 || l.checker.GetTypeOfSymbol(field).Flags()&checker.TypeFlagsAny == 0 || len(field.Declarations) == 0 {
		return false
	}
	return ast.GetSourceFileOfNode(field.Declarations[0]).IsDeclarationFile && !primitiveMember(primitive.Flags(), field.Name)
}

func fixedViewTuple(target *checker.Type) bool {
	return checker.IsTupleType(target) && checker.TupleType_combinedFlags(target.TargetTupleType())&checker.ElementFlagsNonRequired == 0
}

// Unrelated tuple positions share numeric field names. Preserve the existing
// constructor's exact supported child instead of an unrelated fallback refusal.
func (l *lowering) viewTuplePositionChecks(receiver ir.ViewContractID, field string, child ir.ViewContractID) bool {
	if receiver == 0 || child == 0 {
		return false
	}
	contract := l.result.ViewContracts[receiver-1]
	if !contract.FixedTuple {
		return false
	}
	for _, position := range contract.Fields {
		if position.Name == field && position.Contract == child {
			return true
		}
	}
	return false
}

// Only scalar carriers and existing fixed tuple certificates are planned here.
func (l *lowering) tupleScalarUnionType(target *checker.Type) bool {
	if target == nil || target.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	tuples := 0
	for _, member := range target.Types() {
		if member.Flags()&checker.TypeFlagsUndefined != 0 {
			continue
		}
		if fixedViewTuple(member) {
			tuples++
			continue
		}
		if base := l.phantomBase(member); base != nil {
			member = base
		}
		if !interfaceScalar(member) {
			return false
		}
	}
	return tuples > 0
}
