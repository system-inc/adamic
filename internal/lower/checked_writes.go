package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Admit only field contracts that have a runtime representation. Compatible callbacks
// follow instantiated parameter and result field views. Every receiving overload
// must fit the producer. The name summary follows aliases and calls.
func (l *lowering) checkedWidening(node *ast.Node, from, to *checker.Type) bool {
	module := ast.GetSourceFileOfNode(node)
	if !strings.HasSuffix(l.program.FileName(module), ".ts") {
		return false
	}
	if l.checkedNeverArray(from, to) {
		l.result.CheckedElements = true
		return true
	}
	fields := map[string]bool{}
	if !l.checkedFieldRelation(from, to, fields, map[[2]*checker.Type]bool{}) {
		return false
	}
	if len(fields) == 0 && !l.result.CheckedElements {
		return false
	}
	if l.result.CheckedWrites == nil {
		l.result.CheckedWrites = map[string]bool{}
	}
	for name := range fields {
		l.result.CheckedWrites[name] = true
	}
	return true
}

func (l *lowering) checkedFieldRelation(from, to *checker.Type, fields map[string]bool, seen map[[2]*checker.Type]bool) bool {
	from, to = l.concrete(from), l.concrete(to)
	if from.Flags()&checker.TypeFlagsUndefined != 0 && l.includesUndefined(to) {
		return true
	}
	if from == to || seen[[2]*checker.Type{from, to}] {
		return true
	}
	seen[[2]*checker.Type{from, to}] = true
	if from.Flags()&checker.TypeFlagsTypeParameter != 0 {
		from = l.checker.GetBaseConstraintOfType(from)
	}
	if to.Flags()&checker.TypeFlagsTypeParameter != 0 {
		to = l.checker.GetBaseConstraintOfType(to)
	}
	if from == nil || to == nil {
		return false
	}
	from, to = l.withoutUndefined(from), l.withoutUndefined(to)
	if members := l.definedMembers(from); len(members) > 1 {
		for _, member := range members {
			if !l.checkedFieldRelation(member, to, fields, seen) {
				return false
			}
		}
		return true
	}
	if members := l.definedMembers(to); len(members) > 1 {
		for _, member := range members {
			if l.checker.IsTypeAssignableTo(from, member) && !l.checkedFieldRelation(from, member, fields, seen) {
				return false
			}
		}
		return true
	}

	// A compatible callback passes its viewed argument into the actual parameter.
	// Check that reversed field relation, and the result in its forward direction.
	fromCalls := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	toCalls := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(fromCalls) != 0 || len(toCalls) != 0 {
		if len(fromCalls) != 1 || len(toCalls) == 0 {
			return false
		}
		for _, receiving := range toCalls {
			producing, receiving := l.checkedSignaturePair(fromCalls[0], receiving)
			a, b := producing.Parameters(), receiving.Parameters()
			for i, parameter := range a {
				takes := l.censusCallableParameterType(parameter)
				given := l.checker.GetUndefinedType()
				if i < len(b) {
					given = l.censusCallableParameterType(b[i])
				}
				if !l.checker.IsTypeAssignableTo(given, takes) || !l.enumAssignable(given, takes) || !l.checkedFieldRelation(given, takes, fields, seen) {
					return false
				}
			}
			if !l.checkedFieldRelation(l.checker.GetReturnTypeOfSignature(producing), l.checker.GetReturnTypeOfSignature(receiving), fields, seen) {
				return false
			}
		}
		return true
	}

	if own, view := l.containerRelation(from, to); own != nil {
		if l.widened(own, view, map[[2]*checker.Type]bool{}) != nil && !interfaceScalar(l.checker.GetNonNullableType(own)) && !l.checkedFieldRelation(own, view, fields, seen) {
			return false
		}
		l.result.CheckedElements = true
		return true
	}
	if len(l.containers(from)) != 0 || len(l.containers(to)) != 0 || !l.structured(from) || !l.structured(to) || l.callableViewContract(from) || l.callableViewContract(to) || isClassInstance(to) {
		return false
	}
	for _, target := range l.checker.GetPropertiesOfType(to) {
		source := l.checker.GetPropertyOfType(from, target.Name)
		if source == nil {
			continue
		}
		own, viewed := l.contractType(l.checker.GetTypeOfSymbol(source)), l.contractType(l.checker.GetTypeOfSymbol(target))
		if accessorSymbol(source) || accessorSymbol(target) || source.Flags&ast.SymbolFlagsMethod != 0 || target.Flags&ast.SymbolFlagsMethod != 0 {
			return false
		}
		if !l.checker.IsReadonlySymbol(target) && (l.checker.IsReadonlySymbol(source) || !l.checker.IsTypeAssignableTo(viewed, own) || !l.enumAssignable(viewed, own)) {
			contract := l.fieldContract(own)
			viewKind, viewKnown := l.representation(viewed)
			if contract == nil || !viewKnown || (viewKind != contract.Kind && !(contract.Kind == ir.Number && viewKind == ir.MaybeNumber) && !((contract.Kind == ir.Boolean || contract.Kind == ir.MaybeBoolean) && (viewKind == ir.Boolean || viewKind == ir.MaybeBoolean)) && !(contract.NullishOnly && viewKind >= ir.String && viewKind <= ir.Map)) {
				return false
			}
			a, b := l.checker.GetNonNullableType(own), l.checker.GetNonNullableType(viewed)
			// A present reference needs a directional type proof or a reifiable allocation shape.
			if !interfaceScalar(a) && !identicalTypes(l.checker, a, b) && !contract.Structural && !(contract.Kind == ir.Object && contract.Reference) {
				return false
			}
			fields[target.Name] = true
		}
		if l.widened(own, viewed, map[[2]*checker.Type]bool{}) != nil && (l.structured(l.withoutUndefined(own)) || len(l.definedMembers(own)) > 1 && !interfaceScalar(l.checker.GetNonNullableType(own))) {
			if !l.checkedFieldRelation(own, viewed, fields, seen) {
				return false
			}
		}
	}
	return true
}

func (l *lowering) fieldContract(declared *checker.Type) *ir.FieldContract {
	declared = l.contractType(declared)
	kind, known := l.representation(declared)
	if !known || kind < ir.Number || (kind > ir.MaybeNumber && kind != ir.MaybeBoolean) {
		return nil
	}
	contract := &ir.FieldContract{NullishOnly: declared.Flags()&checker.TypeFlagsUndefined != 0, Kind: kind, Declared: l.checker.TypeToString(declared), Nullable: l.includesUndefined(declared) || l.includesNull(declared)}
	present := l.checker.GetNonNullableType(declared)
	if interfaceScalar(present) && l.checker.TypeToString(present) != "boolean" && !l.openNumericEnumType(present) {
		contract.Allowed = l.viewLiterals(present)
	}
	l.referenceContract(contract, declared)
	return contract
}

func (l *lowering) literalFieldContract(node *ast.Node, name string) *ir.FieldContract {
	// Contextual declarations take precedence over the initializer's literal or flow type.
	declared := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if declared == nil {
		declared = l.checker.GetTypeAtLocation(node)
	}
	if field := l.checker.GetPropertyOfType(declared, name); field != nil {
		return l.fieldContract(l.checker.GetTypeOfSymbol(field))
	}
	return nil
}

func (l *lowering) checkedWrite(target *ast.Node) string {
	name := l.fieldName(target.Name())
	if !l.result.CheckedWrites[name] {
		return ""
	}
	expression := sourceExpression(target)
	l.result.WriteChecks = append(l.result.WriteChecks, ir.WriteCheck{Where: l.program.Where(target), Expression: expression})
	return expression
}

func (l *lowering) contractType(declared *checker.Type) *checker.Type {
	declared = l.concrete(declared)
	if declared.Flags()&(checker.TypeFlagsTypeParameter|checker.TypeFlagsIndexedAccess) != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(declared); constraint != nil {
			return l.concrete(constraint)
		}
	}
	return declared
}

// A never-element allocation remains empty through every alias. Its representation
// is immaterial until a store, which the allocation contract always rejects.
func (l *lowering) checkedNeverArray(from, to *checker.Type) bool {
	from, to = l.withoutUndefined(l.contractType(from)), l.withoutUndefined(l.contractType(to))
	if !l.checker.IsArrayType(from) || !l.checker.IsArrayType(to) {
		return false
	}
	own := l.checker.GetElementTypeOfArrayType(from)
	return own.Flags()&checker.TypeFlagsNever != 0
}

func (l *lowering) elementWriteOrigin(node *ast.Node) ir.WriteCheck {
	return ir.WriteCheck{Where: l.program.Where(node), Expression: sourceExpression(node) + "[]"}
}

// Each branch keeps the contract attached to its original allocation. A reduced
// conditional type is a receiving view, not evidence about either allocation.
func (l *lowering) checkedBranchWidening(node *ast.Node, own, target *checker.Type) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindConditionalExpression {
		return l.checkedWidening(node, own, target)
	}
	conditional := node.AsConditionalExpression()
	for _, branch := range []*ast.Node{conditional.WhenTrue, conditional.WhenFalse} {
		source := l.checker.GetTypeAtLocation(branch)
		if l.freshOrWidened(branch, source, target) != nil && !l.checkedBranchWidening(branch, source, target) {
			return false
		}
	}
	return true
}
