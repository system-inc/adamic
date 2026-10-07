package ir

// AllocationSet is a may-flow result. Unknown disqualifies erasure even when
// Sites is nonempty; an empty set never establishes conformance.
type AllocationSet struct {
	Sites   []int
	Unknown bool
	Reasons []string
}

// FieldTypeCertificate preserves a declared semantic type independently of the
// physical shape slot. CheckerType is scoped to this compiled program. It is
// evidence about permitted values, never evidence that a slot is initialized.
type FieldTypeCertificate struct {
	CheckerType  int
	DeclaredType string
	Optional     bool
}

// ClosureValue exposes the operand for provenance tracing inside a flow query.
// Target readers use ClosureTargetsWithFlow, rather than a target-shaped type.
func ClosureValue(call CallClosure) Expression { return call.Closure }

// ClosureTargetsWithFlow resolves function identities through collected values.
// It is proof-query data only: ordinary ClosureTargets and graph ownership keep
// their existing behavior. The resolver must preserve unknown incoming values.
func (p *Program) ClosureTargetsWithFlow(call CallClosure, reaching func(Expression) FunctionTargets) FunctionTargets {
	if call.Direct != 0 {
		target := call.Direct - 1
		if target < 0 || target >= len(p.Functions) {
			return FunctionTargets{Unknown: true}
		}
		return FunctionTargets{Functions: []int{target}}
	}
	return reaching(ClosureValue(call))
}
