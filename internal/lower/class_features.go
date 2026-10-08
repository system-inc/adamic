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
	if !isClassInstance(proven) && l.iteratorMember(proven) != nil {
		return nil, true, l.notYet(node, "Object.keys on a literal with symbol-key storage")
	}
	// A structural view does not remove the iterable literal's hidden symbol slot.
	// Fresh explicit string-keyed literals cannot carry storage erased by a view.
	argument := ast.SkipParentheses(arguments[0])
	fresh := argument.Kind == ast.KindObjectLiteralExpression
	if fresh {
		for _, property := range argument.AsObjectLiteralExpression().Properties.Nodes {
			fresh = fresh && property.Kind != ast.KindSpreadAssignment
		}
	}
	if !fresh && !isClassInstance(proven) && !l.isStaticType(proven) && l.iteratorMember(proven) == nil {
		modules, err := l.moduleOrder(l.program.Files()[0])
		if err != nil {
			return nil, true, err
		}
		var hazard error
		var visit ast.Visitor
		visit = func(candidate *ast.Node) bool {
			if candidate.Kind == ast.KindObjectLiteralExpression {
				shape := l.checker.GetTypeAtLocation(candidate)
				if l.iteratorMember(shape) != nil && l.iterationShapeFits(shape, proven) {
					hazard = &Refused{Where: l.program.Where(node), What: "a structural Object.keys view that can hide an iterable literal's symbol-key storage (adamic/symbol-key-view)", Fix: "pass a new plain object containing only the desired string-keyed fields to Object.keys"}
					return true
				}
			}
			return candidate.ForEachChild(visit)
		}
		for _, module := range modules {
			module.AsNode().ForEachChild(visit)
		}
		if hazard != nil {
			return nil, true, hazard
		}
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
