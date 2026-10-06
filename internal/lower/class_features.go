package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func memberKey(name *ast.Node, class int) string {
	if name.Kind == ast.KindPrivateIdentifier {
		return name.Text() + "@" + strconv.Itoa(class)
	}
	return name.Text()
}

func (l *lowering) hasPrivateStorage(proven *checker.Type) bool {
	proven = l.concrete(proven)
	for _, property := range l.checker.GetPropertiesOfType(proven) {
		for _, declaration := range property.Declarations {
			if declaration.Kind == ast.KindPropertyDeclaration && declaration.Name().Kind == ast.KindPrivateIdentifier {
				return true
			}
		}
	}
	return false
}

func (l *lowering) objectKeys(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "keys" || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
		return nil, false, nil
	}
	arguments := nodesOf(node.AsCallExpression().Arguments)
	if len(arguments) != 1 {
		return nil, true, l.notYet(node, "Object.keys without exactly one argument")
	}
	value, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	if value.Type() != ir.Object {
		return nil, true, l.notYet(node, "Object.keys on a non-object")
	}
	return ir.ObjectKeys{Object: value}, true, nil
}
