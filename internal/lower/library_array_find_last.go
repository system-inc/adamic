package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Port of V8's FindLast / FindLastIndex loop in array-findlast.tq and
// array-findlastindex.tq. Snapshot length, Get each descending index, and keep
// the value across the callback. Get on a removed dense index yields undefined.
func (l *lowering) libraryArrayFindLast(node *ast.Node, array ir.Expression, element ir.Type, name string) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes[0]
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(written), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, true, l.notYet(node, name+" with an overloaded callback")
	}
	returns, known := l.representation(l.checker.GetReturnTypeOfSignature(signatures[0]))
	if !known || returns != ir.Boolean {
		return nil, true, l.notYet(node, name+" requires a represented boolean callback")
	}
	parameters := signatures[0].Parameters()
	first := ir.Maybe(element)
	if len(parameters) > 0 {
		if l.includesNull(l.checker.GetTypeOfSymbol(parameters[0])) {
			return nil, true, l.notYet(node, name+" callback requires distinct null and missing undefined tags (compiler representation support)")
		}
		var known bool
		first, known = l.representation(l.checker.GetTypeOfSymbol(parameters[0]))
		if !known {
			return nil, true, l.notYet(node, name+" with an unrepresented callback element")
		}
	}
	// A scalar T callback cannot receive undefined in its one-word ABI. Preserve
	// the existing dense fast path only when its inline body cannot remove elements.
	if ir.Maybe(element) != element && first == element {
		if !l.libraryArrayFindLastCannotShrink(written) {
			return nil, true, l.notYet(node, name+" callback may shrink the array but its scalar element excludes undefined; include undefined in the callback parameter type")
		}
		return l.arrayVisit(node, array, element, name)
	}
	if element.IsReference() && len(parameters) > 0 && !l.includesUndefined(l.checker.GetTypeOfSymbol(parameters[0])) {
		// Case 2 uses the compiler's search contract and its named exit-70 check
		// before the callback call. Do not pass a missing reference as T.
		return l.arrayVisit(node, array, element, name)
	}
	callback, err := l.expression(written)
	if err != nil {
		return nil, true, err
	}
	if callback.Type() != ir.Closure {
		return nil, true, l.notYet(node, name+" callback is not a closure")
	}
	if len(parameters) == 0 && slotless(ir.Maybe(element)) {
		return nil, true, l.notYet(node, name+" with an ignored optional element needs compiler closure ABI support")
	}
	result, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	b := l.libraryArrayBuilder([]ir.Expression{array, callback})
	source := b.read(b.parameters[0])
	index := b.declare("index", ir.Binary{Operator: ir.Subtract, Left: ir.Length{Array: source}, Right: ir.NumberConstant{Value: 1}})
	item := b.local("element", ir.Maybe(element))
	input := ir.Expression(ir.ArrayIndex{Array: source, Index: b.read(index), Element: element})
	arguments := []ir.Expression{b.read(item), b.read(index), source}
	for i, parameter := range parameters {
		takes, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		if !known || slotless(takes) {
			return nil, true, l.notYet(node, name+" with an unrepresented callback parameter")
		}
		arguments[i] = fit(arguments[i], takes)
		if arguments[i].Type() != takes {
			return nil, true, l.notYet(node, name+" callback cannot represent a missing element")
		}
	}
	selected := ir.CallClosure{Closure: b.read(b.parameters[1]), Arguments: arguments, Returns: ir.Boolean}
	found := ir.Expression(b.read(index))
	final := ir.Expression(ir.NumberConstant{Value: -1})
	if name == "findLast" {
		found = fit(b.read(item), result)
		final = fit(ir.Undefined{}, result)
	}
	b.body = append(b.body, ir.Loop{Condition: ir.Binary{Operator: ir.GreaterOrEqual, Left: b.read(index), Right: ir.NumberConstant{}}, Body: []ir.Statement{
		ir.Declare{Local: item, Value: input},
		ir.If{Condition: selected, Then: []ir.Statement{ir.Return{Value: found}}},
	}, Update: []ir.Statement{ir.Assign{Local: index, Value: ir.Binary{Operator: ir.Subtract, Left: b.read(index), Right: ir.NumberConstant{Value: 1}}}}})
	return b.finish("array_"+name, final), true, nil
}

// This proof only admits inline scalar callbacks with no unknown calls. Writes
// and pushes do not shrink a dense array; pop, shift and splice can. An external
// callback needs the compiler's interprocedural effect/optional-parameter proof.
func (l *lowering) libraryArrayFindLastCannotShrink(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindArrowFunction && node.Kind != ast.KindFunctionExpression {
		return false
	}
	safe := true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		// Accessors, implicit iteration and construction can call user code too.
		switch node.Kind {
		case ast.KindNewExpression, ast.KindTaggedTemplateExpression, ast.KindForOfStatement, ast.KindSpreadElement, ast.KindDeleteExpression:
			safe = false
		case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression, ast.KindTemplateSpan:
			var operand *ast.Node
			switch node.Kind {
			case ast.KindPrefixUnaryExpression:
				operand = node.AsPrefixUnaryExpression().Operand
			case ast.KindPostfixUnaryExpression:
				operand = node.AsPostfixUnaryExpression().Operand
			default:
				operand = node.AsTemplateSpan().Expression
			}
			of, known := l.representation(l.checker.GetTypeAtLocation(operand))
			if !known || (of.IsReference() && of != ir.String) {
				safe = false
			}
		case ast.KindPropertyAccessExpression:
			parent := node.Parent
			if parent != nil && (parent.Kind == ast.KindPrefixUnaryExpression || parent.Kind == ast.KindPostfixUnaryExpression || (parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == node)) {
				safe = false
			}
			called := node.Parent != nil && node.Parent.Kind == ast.KindCallExpression && node.Parent.AsCallExpression().Expression == node
			of, known := l.representation(l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression))
			if !called && !(node.Name().Text() == "length" && known && (of == ir.Array || of == ir.String)) {
				safe = false
			}
		case ast.KindElementAccessExpression:
			of, known := l.representation(l.checker.GetTypeAtLocation(node.AsElementAccessExpression().Expression))
			if !known || (of != ir.Array && of != ir.String) {
				safe = false
			}
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			operator := binary.OperatorToken.Kind
			if operator != ast.KindEqualsToken && operator != ast.KindEqualsEqualsEqualsToken && operator != ast.KindExclamationEqualsEqualsToken {
				for _, operand := range []*ast.Node{binary.Left, binary.Right} {
					of, known := l.representation(l.checker.GetTypeAtLocation(operand))
					if !known || (of.IsReference() && of != ir.String) {
						safe = false
					}
				}
			}
		}
		if node.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if callee.Kind != ast.KindPropertyAccessExpression {
				safe = false
				return true
			}
			name := callee.Name().Text()
			switch name {
			case "push":
				receiver, known := l.representation(l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression))
				if !known || receiver != ir.Array || !l.libraryMember(callee) {
					safe = false
				}
			case "log":
				if !l.isConsole(callee) {
					safe = false
				}
			default:
				safe = false
			}
		}
		return node.ForEachChild(visit)
	}
	node.ForEachChild(visit)
	return safe
}
