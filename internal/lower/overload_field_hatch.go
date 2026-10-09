package lower

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The ruled TypeScript hatch checks the failed readonly scalar field. Adamic
// sources still require a proof. Other structural relations remain refused.
func (l *lowering) overloadFieldHatch(node *ast.Node, produced, promised *checker.Type) string {
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {
		return ""
	}
	produced, promised = l.concrete(produced), l.concrete(promised)
	if isClassInstance(promised) || l.callableViewContract(promised) {
		return ""
	}
	path := l.overloadResultPath(produced, promised, "result", map[[2]*checker.Type]bool{})
	if path == "result.kind" && l.fieldLiteral(promised, "kind") != nil && produced.Flags()&checker.TypeFlagsUnion != 0 {
		target := l.checker.GetPropertyOfType(promised, "kind")
		if target == nil || !l.checker.IsReadonlySymbol(target) || accessorSymbol(target) || target.Flags&ast.SymbolFlagsOptional != 0 {
			return ""
		}
		found := false
		for _, member := range produced.Types() {
			if l.censusRelated(member, promised) {
				found = true
				continue
			}
			field := l.checker.GetPropertyOfType(member, "kind")
			if isClassInstance(member) || field == nil || field.Flags&ast.SymbolFlagsOptional != 0 || accessorSymbol(field) || !l.checker.IsReadonlySymbol(field) || !l.overloadScalar(l.checker.GetTypeOfSymbol(field)) {
				return ""
			}
		}
		if found {
			return "kind"
		}
	}
	if path != "result.value" || produced.Flags()&checker.TypeFlagsObject == 0 || promised.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(produced) || isClassInstance(promised) {
		return ""
	}
	for _, target := range l.checker.GetPropertiesOfType(promised) {
		source := l.checker.GetPropertyOfType(produced, target.Name)
		if source == nil || accessorSymbol(source) || accessorSymbol(target) || source.Flags&ast.SymbolFlagsOptional != target.Flags&ast.SymbolFlagsOptional {
			return ""
		}
		from, to := l.checker.GetTypeOfSymbol(source), l.checker.GetTypeOfSymbol(target)
		if target.Name == "value" {
			if target.Flags&ast.SymbolFlagsOptional != 0 || !l.checker.IsReadonlySymbol(source) || !l.checker.IsReadonlySymbol(target) || !l.overloadScalar(from) || !l.overloadScalar(to) {
				return ""
			}
			a, ak := l.representation(from)
			b, bk := l.representation(to)
			if !ak || !bk || a != b && !(a == ir.Union && b == ir.String) {
				return ""
			}
		} else if l.checker.IsReadonlySymbol(source) != l.checker.IsReadonlySymbol(target) || !checker.Checker_isTypeIdenticalTo(l.checker, from, to) {
			return ""
		}
	}
	return "value"
}

func (l *lowering) recordOverloadFieldCheck(call *ast.CallExpression, implementation *ast.Node, ordinal int, field, reason string) {
	counts := &l.result.PredicateChecks
	counts.Checked++
	counts.Sites = append(counts.Sites, ir.PredicateCallCheck{Where: l.program.Where(call.AsNode()), Function: implementation.Name().Text(), Overload: ordinal, Directions: []ir.PredicateDirectionCheck{{Direction: "result." + field, Status: "checked", Reason: reason}}})
}

func (l *lowering) overloadFieldMessage(call *ast.CallExpression, implementation *ast.Node, ordinal int, field string, produced, promised *checker.Type) string {
	return fmt.Sprintf("overload %d of %s at %s field result.%s: implementation result %s cannot serve overload result %s", ordinal, implementation.Name().Text(), l.program.Where(call.AsNode()), field, l.checker.TypeToString(produced), l.checker.TypeToString(promised))
}
