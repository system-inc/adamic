package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Only representation-preserving array elements and Map values can use these checks.
// Keys, Sets, tuples and function variance retain their existing refusals.
func (l *lowering) containerRelation(from, to *checker.Type) (*checker.Type, *checker.Type) {
	from, to = l.withoutUndefined(l.contractType(from)), l.withoutUndefined(l.contractType(to))
	var own, view *checker.Type
	if l.checker.IsArrayType(from) && l.checker.IsArrayType(to) {
		own, view = l.checker.GetElementTypeOfArrayType(from), l.checker.GetElementTypeOfArrayType(to)
	} else if l.isLibraryType(from, "Map", "ReadonlyMap") && l.isLibraryType(to, "Map", "ReadonlyMap") {
		a, b := l.typeArguments(from), l.typeArguments(to)
		if len(a) != 2 || len(b) != 2 || !identicalTypes(l.checker, a[0], b[0]) {
			return nil, nil
		}
		own, view = a[1], b[1]
	} else {
		return nil, nil
	}
	if checker.IsTupleType(own) || checker.IsTupleType(view) || isClassInstance(own) || isClassInstance(view) {
		return nil, nil
	}
	contract := l.fieldContract(own)
	kind, known := l.representation(view)
	if contract != nil && known && kind == contract.Kind && !l.callableViewContract(own) && !l.callableViewContract(view) {
		return own, view
	}
	return nil, nil
}

func (l *lowering) containerAllocation(node *ast.Node, value ir.Expression) ir.Expression {
	if !l.result.CheckedElements || (value.Type() != ir.Array && value.Type() != ir.Map) {
		return value
	}
	fresh := false
	switch allocation := value.(type) {
	case ir.ArrayLiteral, ir.ArrayMap, ir.ArraySlice, ir.ArrayConcat, ir.ArrayFrom, ir.ArraySplice, ir.MapKeys, ir.MapValues, ir.SetValues, ir.MapEntries, ir.CodePoints, ir.ProgramArguments, ir.MapNew:
		fresh = true
	case ir.ArrayVisit:
		fresh = allocation.Method == "filter"
	case ir.ArrayFill:
		fresh = allocation.Array == nil
	}
	syntax := ast.SkipParentheses(node)
	if syntax.Kind == ast.KindArrayLiteralExpression || syntax.Kind == ast.KindNewExpression {
		fresh = true
	}
	if syntax.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(syntax.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression {
			receiver := callee.AsPropertyAccessExpression().Expression
			if l.checker.IsArrayType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))) {
				switch callee.Name().Text() {
				case "slice", "map", "filter", "concat", "flat", "flatMap", "with", "toSpliced", "toSorted", "toReversed":
					fresh = true
				}
			}
		}
	}
	if !fresh {
		return value
	}
	declared := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if declared == nil {
		declared = l.checker.GetTypeAtLocation(node)
	}
	declared = l.checker.GetNonNullableType(l.contractType(declared))
	var element *checker.Type
	if l.checker.IsArrayType(declared) {
		element = l.checker.GetElementTypeOfArrayType(declared)
	} else if l.isLibraryType(declared, "Map", "ReadonlyMap") {
		arguments := l.typeArguments(declared)
		if len(arguments) == 2 {
			element = arguments[1]
		}
	}
	if element == nil {
		return value
	}
	contract := l.fieldContract(element)
	if contract != nil && !contract.Reference && len(contract.Allowed) == 0 && !contract.NullishOnly {
		contract = nil
	}
	return ir.ContractContainer{Value: value, Contract: contract, AllocationType: l.allocationContractType(node)}
}
