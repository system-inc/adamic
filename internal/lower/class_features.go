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
	// A spread of a possibly absent plain object has synthetic optional slots, not real keys.
	// Class and accessor descriptors provide the public shape for their stored fields.
	proven := l.checker.GetTypeAtLocation(arguments[0])
	// Plain data descriptors and dense arrays use the library own-key proof.
	// Class, static and accessor storage retain their public-shape dispatch.
	if proven.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsStringLike|checker.TypeFlagsNull) != 0 || l.exactObject(arguments[0], 0) || l.checker.IsArrayType(proven) {
		return l.objectCall(node, "keys")
	}
	if !isClassInstance(proven) && l.iteratorMember(proven) != nil {
		return nil, true, l.notYet(node, "Object.keys on a literal with symbol-key storage")
	}
	if !isClassInstance(proven) && !l.isStaticType(proven) && !l.hasAccessorStorage(proven) {
		if l.isLibraryType(proven, "RegExp", "Error") || l.includesUndefined(proven) {
			return l.objectCall(node, "keys")
		}
		for _, property := range l.checker.GetPropertiesOfType(proven) {
			if property.Flags&ast.SymbolFlagsOptional != 0 {
				return l.objectCall(node, "keys")
			}
		}
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
