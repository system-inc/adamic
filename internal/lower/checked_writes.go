package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Admit only field contracts that have a runtime representation. Container invariance and
// callable variance retain their existing refusals. The name summary follows aliases and calls.
func (l *lowering) checkedWidening(node *ast.Node, from, to *checker.Type) bool {
	module := ast.GetSourceFileOfNode(node)
	if !strings.HasSuffix(l.program.FileName(module), ".ts") {
		return false
	}
	fields := map[string]bool{}
	if !l.checkedFieldRelation(from, to, fields, map[[2]*checker.Type]bool{}) {
		return false
	}
	if len(fields) == 0 {
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
			if contract == nil || !viewKnown || viewKind != contract.Kind {
				return false
			}
			a, b := l.checker.GetNonNullableType(own), l.checker.GetNonNullableType(viewed)
			// Reference contracts check presence, not arbitrary structural subtyping.
			if !interfaceScalar(a) && !identicalTypes(l.checker, a, b) {
				return false
			}
			fields[target.Name] = true
		}
		if l.widened(own, viewed, map[[2]*checker.Type]bool{}) != nil && l.structured(l.withoutUndefined(own)) {
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
	if !known || kind < ir.Number || kind > ir.MaybeNumber {
		return nil
	}
	contract := &ir.FieldContract{Kind: kind, Declared: l.checker.TypeToString(declared), Nullable: l.includesUndefined(declared) || l.includesNull(declared)}
	present := l.checker.GetNonNullableType(declared)
	if kind.IsReference() && kind != ir.String && !l.broadReferencePayload(present, map[*checker.Type]bool{}) {
		// Keep the declaration for diagnostics, but fail closed when a hidden generic
		// subtype needs container elements, callable identity, or nested literal domains.
		contract.Kind = 0
	}
	if interfaceScalar(present) && l.checker.TypeToString(present) != "boolean" && !l.openNumericEnumType(present) {
		contract.Allowed = l.viewLiterals(present)
	}
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

// Monomorphized asserted receivers can expose a reference contract narrower than
// the generic constraint. Presence checks cannot prove its structural payload.
func (l *lowering) checkedReferenceWrite(target, value *ast.Node) error {
	if !l.result.CheckedWrites[l.fieldName(target.Name())] {
		return nil
	}
	receiver := ast.SkipParentheses(target.AsPropertyAccessExpression().Expression)
	if receiver.Kind != ast.KindAsExpression {
		return nil
	}
	own := l.concrete(l.checker.GetTypeAtLocation(receiver.AsAsExpression().Expression))
	field := l.checker.GetPropertyOfType(own, l.fieldName(target.Name()))
	if field == nil {
		return nil
	}
	declared := l.contractType(l.checker.GetTypeOfSymbol(field))
	kind, known := l.representation(declared)
	incoming := l.contractType(l.checker.GetTypeAtLocation(value))
	if known && kind.IsReference() && incoming.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsNever) == 0 && !l.checker.IsTypeAssignableTo(l.checker.GetNonNullableType(incoming), l.checker.GetNonNullableType(declared)) {
		return l.notYet(target, "a checked reference write requiring a structural field contract")
	}
	return nil
}

func (l *lowering) broadReferencePayload(declared *checker.Type, seen map[*checker.Type]bool) bool {
	declared = l.contractType(declared)
	if seen[declared] {
		return true
	}
	seen[declared] = true
	if len(l.containers(declared)) != 0 || l.callableViewContract(declared) {
		return false
	}
	if interfaceScalar(declared) {
		return len(l.viewLiterals(declared)) == 0 || l.openNumericEnumType(declared) || l.checker.TypeToString(declared) == "boolean"
	}
	if declared.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range declared.Types() {
			if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
				continue
			}
			if !l.broadReferencePayload(member, seen) {
				return false
			}
		}
		return true
	}
	if !l.structured(declared) {
		return false
	}
	for _, field := range l.checker.GetPropertiesOfType(declared) {
		if !l.broadReferencePayload(l.checker.GetTypeOfSymbol(field), seen) {
			return false
		}
	}
	return true
}
