package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Searches Get every index of their captured length, including removed indices.
// Keep admission separate from representation: string and string | undefined
// have identical storage, but only one permits the missing value at this call.
func (lowering *lowering) arraySearchContract(
	node *ast.Node,
	signature *checker.Signature,
	visit ir.ArrayVisit,
) (ir.Expression, bool, error) {
	switch visit.Method {
	case "find", "findIndex", "findLast", "findLastIndex":
	default:
		return visit, true, nil
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	receiver := callee.AsPropertyAccessExpression().Expression
	element := lowering.concrete(lowering.elementTypeOfArray(lowering.checker.GetTypeAtLocation(receiver)))
	visit.SearchElementName = lowering.checker.TypeToString(element)
	parameters := signature.Parameters()
	if len(parameters) == 0 {
		visit.SearchUndefined = true // No value parameter can use undefined as T.
		return visit, true, nil
	}
	first := lowering.concrete(lowering.checker.GetTypeOfSymbol(parameters[0]))
	defaulted := false
	if declaration := parameters[0].ValueDeclaration; declaration != nil && declaration.Kind == ast.KindParameter {
		parameter := declaration.AsParameterDeclaration()
		defaulted = parameter.Initializer != nil
	}
	var known bool
	visit.SearchFirst, known = lowering.representation(first)
	if !known {
		return nil, true, lowering.notYet(
			node,
			visit.Method+" with an unresolved callback value type "+lowering.checker.TypeToString(first),
		)
	}
	visit.SearchNull = visit.Element.IsReference() &&
		visit.Element != ir.Union &&
		visit.SearchFirst == ir.Union &&
		lowering.includesNull(element) &&
		!lowering.includesUndefined(element)
	visit.SearchUndefined = defaulted ||
		lowering.includesUndefined(first) ||
		first.Flags()&checker.TypeFlagsUnknown != 0
	if defaulted {
		visit.SearchFirst = ir.Maybe(visit.SearchFirst)
	}
	if visit.SearchFirst != visit.Element &&
		visit.SearchFirst != ir.Union &&
		!(visit.SearchFirst.IsMaybe() && visit.SearchFirst.Present() == visit.Element) {
		return nil, true, lowering.notYet(
			node,
			visit.Method+" with an unsupported callback value representation for "+lowering.checker.TypeToString(first),
		)
	}
	return visit, true, nil
}
