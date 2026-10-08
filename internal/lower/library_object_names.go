package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) objectNamesCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	if name != "keys" && name != "getOwnPropertyNames" {
		return nil, false, nil
	}
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) != 1 || hasSpread(node) {
		return nil, true, l.notYet(node, "Object."+name+" with these arguments")
	}
	argument := written[0]
	proven := l.checker.GetTypeAtLocation(argument)
	if proven.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
		return l.objectNamesFailure(argument)
	}
	dense := !checker.IsTupleType(proven) && l.objectNamesDenseArray(argument, 0)
	primitive := proven.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsStringLike) != 0
	if !primitive && !dense && !l.exactObject(argument, 0) {
		return nil, true, l.notYet(argument, "Object."+name+" without a proven complete plain shape or primitive (array holes, descriptors and host objects are not represented)")
	}
	value, err := l.expression(argument)
	if err != nil {
		return nil, true, err
	}
	return ir.ObjectCall{Method: name, Arguments: []ir.Expression{fit(value, ir.Union)}, Returns: ir.Array}, true, nil
}

// A let literal's type describes its whole shape only until the binding is replaced. Reject every
// possible target use, including destructuring and loop assignments. Field writes are rejected too:
// this is deliberately more conservative than a whole-program shape-flow proof.
func (l *lowering) objectBindingAssigned(declaration *ast.Node, symbol *ast.Symbol) bool {
	assigned := false
	var target ast.Visitor
	target = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && l.symbol(node) == symbol {
			assigned = true
		}
		return node.ForEachChild(target)
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if assigned {
			return true
		}
		switch node.Kind {
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindEqualsToken {
				target(binary.Left)
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			target(node.AsForInOrOfStatement().Initializer)
		}
		return node.ForEachChild(visit)
	}
	ast.GetSourceFileOfNode(declaration).AsNode().ForEachChild(visit)
	return assigned
}

// Only literal arrays and bindings used solely by these own-key queries enter
// here. No alias, callback, field write or collection operation can introduce
// sparse elements or named properties between construction and observation.
func (l *lowering) objectNamesDenseArray(node *ast.Node, depth int) bool {
	if depth > 16 {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindArrayLiteralExpression {
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			if element.Kind == ast.KindOmittedExpression || element.Kind == ast.KindSpreadElement {
				return false
			}
		}
		return true
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	variable := declaration.AsVariableDeclaration()
	if variable.Type != nil || variable.Initializer == nil || !l.objectNamesDenseArray(variable.Initializer, depth+1) {
		return false
	}
	safe := true
	var visit ast.Visitor
	visit = func(read *ast.Node) bool {
		if !safe {
			return true
		}
		if ast.IsIdentifier(read) && l.symbol(read) == symbol && read != declaration.Name() {
			parent := read.Parent
			if parent.Kind != ast.KindCallExpression {
				safe = false
				return true
			}
			call := parent.AsCallExpression()
			if len(call.Arguments.Nodes) != 1 || call.Arguments.Nodes[0] != read {
				safe = false
				return true
			}
			callee := ast.SkipParentheses(call.Expression)
			if callee.Kind != ast.KindPropertyAccessExpression {
				safe = false
				return true
			}
			property := callee.AsPropertyAccessExpression()
			name := property.Name().Text()
			if !l.isLibraryGlobal(property.Expression, "Object") || (name != "keys" && name != "getOwnPropertyNames") {
				safe = false
				return true
			}
		}
		return read.ForEachChild(visit)
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return safe
}

func (l *lowering) objectNamesFailure(argument *ast.Node) (ir.Expression, bool, error) {
	var value ir.Expression = ir.Undefined{}
	if ast.SkipParentheses(argument).Kind != ast.KindNullKeyword {
		var err error
		value, err = l.expression(argument)
		if err != nil {
			return nil, true, err
		}
	}
	function := len(l.result.Functions)
	parameter := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "own_keys_receiver", Type: value.Type(), Function: function})
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "object_own_keys_error", Parameters: []int{parameter}, Returns: ir.Array, Body: []ir.Statement{ir.Throw{Value: ir.ObjectCall{Method: "typeError", Arguments: []ir.Expression{ir.StringConstant{Index: l.constant("Cannot convert undefined or null to object")}}, Returns: ir.Object}}}})
	return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.Array}, true, nil
}
