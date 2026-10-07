package lower

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) arrayLengthConstructor(node *ast.Node) (ir.Expression, bool, error) {
	var callee *ast.Node
	var arguments *ast.NodeList
	if node.Kind == ast.KindNewExpression {
		callee, arguments = node.AsNewExpression().Expression, node.AsNewExpression().Arguments
	} else if node.Kind == ast.KindCallExpression {
		callee, arguments = node.AsCallExpression().Expression, node.AsCallExpression().Arguments
	} else {
		return nil, false, nil
	}
	if !l.isLibraryGlobal(callee, "Array") {
		return nil, false, nil
	}
	if arguments == nil || len(arguments.Nodes) != 1 {
		return nil, true, l.notYet(node, "Array constructor other than one numeric length")
	}
	length, err := l.expression(arguments.Nodes[0])
	if err != nil {
		return nil, true, err
	}
	if length.Type() != ir.Number {
		return nil, true, l.notYet(node, "Array constructor with a nonnumeric argument")
	}
	// A bare Array has any[] as its checker type. It can expose length, but
	// elementType continues to refuse reads and writes without a proven element.
	element := ir.Number
	proven := l.checker.GetElementTypeOfArrayType(l.checker.GetTypeAtLocation(node))
	if proven != nil && proven.Flags()&checker.TypeFlagsAny == 0 {
		element, err = l.elementType(node)
		if err != nil {
			return nil, true, err
		}
	}
	if element != ir.Number && element != ir.MaybeNumber {
		return nil, true, l.notYet(node, "holey Array of this element representation")
	}
	return ir.ArrayHoles{Length: length, Element: element}, true, nil
}

// Until each consumer has a hole contract, admitting a constructor makes every
// array in the compilation potentially holey. This includes aliases, fields,
// parameters, returned arrays and callbacks. There is no unsound local-name test.
func (l *lowering) checkArrayHoles() error {
	holey := false
	visit := func(node any) bool {
		if _, ok := node.(ir.ArrayHoles); ok {
			holey = true
		}
		return true
	}
	walk(l.result.Main, visit)
	for _, f := range l.result.Functions {
		walk(f.Body, visit)
	}
	if !holey {
		return nil
	}
	var refused error
	check := func(node any) bool {
		var operation string
		switch node := node.(type) {
		case ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArraySearch,
			ir.ArrayJoin, ir.ArrayPush, ir.ArrayPop, ir.ArraySlice,
			ir.ArraySort, ir.ArrayFill, ir.ArraySplice, ir.ArrayReverse,
			ir.ArrayConcat, ir.ArrayFrom, ir.JSONStringify, ir.ObjectCall,
			ir.ObjectKeys, ir.HasOwn, ir.SetNew, ir.MapNew,
			ir.NodeBufferCall, ir.NodeHostCall, ir.MathCall, ir.StringFromCodes:
			operation = fmt.Sprintf("%T", node)
		case ir.ForOf:
			if node.Iterable.Type() == ir.Array {
				operation = "array for-of"
			}
		case ir.ArrayLiteral:
			for _, spread := range node.Spread {
				if spread {
					operation = "array spread"
				}
			}
		case ir.ArrayIndex:
			if node.Relative {
				operation = "array at"
			}
		}
		if operation != "" && refused == nil {
			refused = &NotYet{Where: l.result.Source, What: operation + " in a program that may contain holey arrays"}
		}
		return true
	}
	walk(l.result.Main, check)
	for _, f := range l.result.Functions {
		walk(f.Body, check)
	}
	return refused
}
