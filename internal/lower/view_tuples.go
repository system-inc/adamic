package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

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
