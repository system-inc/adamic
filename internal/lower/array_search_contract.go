package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Searches Get every index of their captured length, including removed indices.
// Keep admission separate from representation: string and string | undefined
// have identical storage, but only one permits the missing value at this call.
func (l *lowering) arraySearchContract(node *ast.Node, signature *checker.Signature, visit ir.ArrayVisit) (ir.Expression, bool, error) {
	switch visit.Method {
	case "find", "findIndex", "findLast", "findLastIndex":
	default:
		return visit, true, nil
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	receiver := callee.AsPropertyAccessExpression().Expression
	element := l.concrete(l.checker.GetElementTypeOfArrayType(l.checker.GetTypeAtLocation(receiver)))
	visit.SearchElementName = l.checker.TypeToString(element)
	parameters := signature.Parameters()
	if len(parameters) == 0 {
		visit.SearchUndefined = true // No value parameter can use undefined as T.
		return visit, true, nil
	}
	first := l.concrete(l.checker.GetTypeOfSymbol(parameters[0]))
	defaulted := false
	if declaration := parameters[0].ValueDeclaration; declaration != nil && declaration.Kind == ast.KindParameter {
		parameter := declaration.AsParameterDeclaration()
		defaulted = parameter.Initializer != nil
	}
	var known bool
	visit.SearchFirst, known = l.representation(first)
	if !known {
		return nil, true, l.notYet(node, visit.Method+" with an unresolved callback value type "+l.checker.TypeToString(first))
	}
	visit.SearchUndefined = defaulted || l.includesUndefined(first) || first.Flags()&checker.TypeFlagsUnknown != 0
	if defaulted {
		visit.SearchFirst = ir.Maybe(visit.SearchFirst)
	}
	if visit.SearchFirst != visit.Element && visit.SearchFirst != ir.Union && !(visit.SearchFirst.IsMaybe() && visit.SearchFirst.Present() == visit.Element) {
		return nil, true, l.notYet(node, visit.Method+" with an unsupported callback value representation for "+l.checker.TypeToString(first))
	}
	return visit, true, nil
}
