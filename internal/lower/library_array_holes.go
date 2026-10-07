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
	if proven != nil && proven.Flags()&checker.TypeFlagsAny != 0 {
		contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
		if contextual != nil && l.checker.IsArrayType(contextual) {
			proven = l.checker.GetElementTypeOfArrayType(contextual)
		}
	}
	if proven != nil && proven.Flags()&checker.TypeFlagsAny == 0 {
		var known bool
		element, known = l.kept(proven)
		if !known || slotless(element) {
			return nil, true, l.notYet(node, "Array length constructor with an unsupported element representation")
		}
	}
	if proven != nil && proven.Flags()&checker.TypeFlagsAny != 0 {
		parent := node.Parent
		for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			parent = parent.Parent
		}
		if parent == nil || (parent.Kind != ast.KindVariableDeclaration && !(parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "length")) {
			return nil, true, l.notYet(node, "an untyped length Array escaping without a proven element representation")
		}
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
		if _, ok := node.(ir.ArraySetLength); ok {
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
	if err := l.arrayHolesSourceRefusal(); err != nil {
		return err
	}
	var refused error
	check := func(node any) bool {
		var operation string
		switch node := node.(type) {
		case ir.ArrayPush, ir.ArrayPop, ir.ArraySlice,
			ir.ArraySort, ir.ArrayFill, ir.ArraySplice, ir.ArrayReverse,
			ir.ArrayConcat, ir.ArrayFrom, ir.JSONStringify, ir.ObjectCall,
			ir.ObjectKeys, ir.HasOwn, ir.SetNew, ir.MapNew,
			ir.NodeBufferCall, ir.NodeHostCall, ir.UnionToString:
			operation = fmt.Sprintf("%T", node)
		case ir.Length:
			if read, ok := node.Array.(ir.Read); ok && l.arrayLocalUntyped(read.Local) {
				return false
			}
		case ir.Read:
			if node.Of == ir.Array {
				if proven := l.localTypes[node.Local]; proven != nil && l.checker.IsArrayType(proven) {
					element := l.checker.GetElementTypeOfArrayType(proven)
					if element != nil && element.Flags()&checker.TypeFlagsAny != 0 {
						operation = "an untyped array escaping its length observation"
					}
				}
			}
		case ir.MathCall:
			if node.Spread != nil {
				operation = "Math array spread"
			}
		case ir.StringFromCodes:
			if node.Spread != nil {
				operation = "String array spread"
			}
		case ir.ArrayVisit:
			if node.Method == "find" || node.Method == "findIndex" || node.Method == "findLast" || node.Method == "findLastIndex" {
				operation = "array " + node.Method
			}
		case ir.ArrayJoin:
			if node.Depth != 0 {
				operation = "nested array join"
			}
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

func (l *lowering) arrayLengthWrite(node *ast.Node) ([]ir.Statement, bool, error) {
	binary := node.AsBinaryExpression()
	target := ast.SkipParentheses(binary.Left)
	if binary.OperatorToken.Kind != ast.KindEqualsToken || target.Kind != ast.KindPropertyAccessExpression || target.Name().Text() != "length" {
		return nil, false, nil
	}
	receiver := target.AsPropertyAccessExpression().Expression
	of, known := l.representation(l.checker.GetTypeAtLocation(receiver))
	if !known || of != ir.Array {
		return nil, false, nil
	}
	array, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	length, err := l.expression(binary.Right)
	if err != nil {
		return nil, true, err
	}
	if length.Type() != ir.Number {
		return nil, true, l.notYet(node, "array length assignment without a number")
	}
	return []ir.Statement{ir.Evaluate{Value: ir.ArraySetLength{Array: array, Length: length}}}, true, nil
}

func (l *lowering) arrayLocalUntyped(local int) bool {
	proven := l.localTypes[local]
	if proven == nil || !l.checker.IsArrayType(proven) {
		return false
	}
	element := l.checker.GetElementTypeOfArrayType(proven)
	return element != nil && element.Flags()&checker.TypeFlagsAny != 0
}

func (l *lowering) arrayHolesSourceRefusal() error {
	denied := map[string]bool{
		"hasOwnProperty": true, "at": true, "values": true, "entries": true, "push": true, "pop": true, "slice": true, "concat": true,
		"splice": true, "reverse": true, "sort": true, "fill": true,
		"find": true, "findIndex": true, "findLast": true, "findLastIndex": true,
		"flat": true, "flatMap": true, "copyWithin": true, "with": true,
		"toReversed": true, "toSorted": true, "toSpliced": true,
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if node.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression {
				access := callee.AsPropertyAccessExpression()
				name := access.Name().Text()
				if denied[name] {
					receiver, known := l.representation(l.checker.GetTypeAtLocation(access.Expression))
					if known && receiver == ir.Array {
						found = l.notYet(node, "array "+name+" on a potentially holey array")
						return true
					}
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.ForEachChild(visit)
	}
	return found
}
