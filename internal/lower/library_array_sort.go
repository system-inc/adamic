package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// SortCompareUserFn in V8 third_party/v8/builtins/array-sort.tq propagates
// abrupt completion before ToNumber. A never-returning callback has no numeric
// result, so carry that proof to the native comparison adapter explicitly.
func (l *lowering) libraryArraySort(node *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, true, l.notYet(node, "sorting without a proven array receiver")
	}
	arrayType := l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression)
	if target := l.weakTarget(arrayType); target != nil {
		arrayType = target
	}
	proven := l.checker.GetElementTypeOfArrayType(l.checker.GetNonNullableType(arrayType))
	if proven == nil || proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return nil, true, l.notYet(node, "sorting array elements whose optionality is not proven")
	}
	if l.includesUndefined(proven) && element != ir.MaybeNumber {
		return nil, true, l.notYet(node, "sorting optional reference elements requires undefined partitioning")
	}
	if len(written) != 1 {
		return l.arraySort(node, array, element)
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(written[0]), checker.SignatureKindCall)
	if len(signatures) != 1 || l.checker.GetReturnTypeOfSignature(signatures[0]).Flags()&checker.TypeFlagsNever == 0 {
		return l.arraySort(node, array, element)
	}
	if len(signatures[0].Parameters()) > 2 {
		return nil, true, l.notYet(node, "a never-returning comparator with more than two parameters")
	}
	for _, parameter := range signatures[0].Parameters() {
		of, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		if !known || of != element {
			return nil, true, l.notYet(node, "a never-returning comparator with incompatible element representations")
		}
	}
	callback, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	if callback.Type() != ir.Closure {
		return nil, true, l.notYet(node, "a never-returning comparator that isn't a represented function")
	}
	return ir.ArraySort{Array: array, Callback: callback, Element: element, CallbackNever: true}, true, nil
}
