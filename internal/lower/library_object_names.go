package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) objectNamesCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	return l.objectNamesCallArguments(node, name, node.AsCallExpression().Arguments.Nodes)
}

func (l *lowering) objectNamesCallArguments(node *ast.Node, name string, written []*ast.Node) (ir.Expression, bool, error) {
	if name != "keys" && name != "getOwnPropertyNames" {
		return nil, false, nil
	}
	if len(written) != 1 || hasSpread(node) {
		return nil, true, l.notYet(node, "Object."+name+" with these arguments")
	}
	argument := written[0]
	proven := l.checker.GetTypeAtLocation(argument)
	primitive := proven.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsStringLike) != 0
	if !primitive && !l.exactObject(argument, 0) {
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
