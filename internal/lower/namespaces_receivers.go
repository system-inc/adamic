package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// An arrow inherits its enclosing receiver. Ordinary nested functions, object
// methods and class members bind their own this and are checked by their own
// lowering paths, not the namespace-function receiver rule.
func namespaceOwnThis(declaration *ast.Node) bool {
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil || ast.IsTypeNode(node) {
			return false
		}
		if node != declaration && (ast.IsClassLike(node) || ast.IsFunctionLike(node) && node.Kind != ast.KindArrowFunction) {
			return false
		}
		if node.Kind == ast.KindThisKeyword {
			return true
		}
		return node.ForEachChild(visit)
	}
	return visit(declaration.Body())
}

func namespaceReceiverDeclaration(function *ast.Node) *ast.Node {
	if function.Kind == ast.KindFunctionDeclaration && function.Parent != nil && function.Parent.Kind == ast.KindModuleBlock {
		return function.Parent.Parent
	}
	return nil
}

func namespaceReceiverOwner(node *ast.Node) *ast.Node {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if ast.IsFunctionLike(parent) && parent.Kind != ast.KindArrowFunction {
			return namespaceReceiverDeclaration(parent)
		}
		if ast.IsClassLike(parent) {
			return nil
		}
	}
	return nil
}

// A namespace receiver is erased only after every reference is a direct,
// qualified call on its declaring namespace. Function values and aliases do
// not inherit this proof. Receiver properties must name represented exports;
// identity, writes through this, and computed keys stay explicit boundaries.
func (l *lowering) namespaceReceiverProof(function *ast.Node) error {
	declaration := namespaceReceiverDeclaration(function)
	target := l.symbol(function.Name())
	var found error
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		if ast.IsTypeNode(node) {
			return false
		}
		if l.namespaceValueNode(node) && l.symbol(node) == target {
			if called(node) {
				if node.Kind != ast.KindPropertyAccessExpression || l.namespaceDeclaration(node.Expression()) != declaration {
					found = &Refused{Where: l.program.Where(node), What: "a detached call to namespace function " + function.Name().Text() + "; qualified and detached calls have different receivers", Fix: "call it through its declaring namespace"}
					return true
				}
			} else {
				found = l.notYet(node, "namespace receiver function "+function.Name().Text()+" used as a value; its closed-world receiver is not proven")
				return true
			}
		}
		return node.ForEachChild(visit)
	}
	files, _ := esmModuleOrder(l.checker, l.program.Files()[0])
	for _, file := range files {
		if visit(file.AsNode()) {
			return found
		}
	}
	return nil
}

func (l *lowering) namespaceReceiverRead(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind != ast.KindPropertyAccessExpression || ast.SkipParentheses(node.Expression()).Kind != ast.KindThisKeyword {
		return nil, false, nil
	}
	declaration := namespaceReceiverOwner(node)
	if declaration == nil {
		return nil, false, nil
	}
	symbol := l.checker.GetPropertyOfType(l.checker.GetTypeOfSymbol(l.symbol(declaration.Name())), node.Name().Text())
	if symbol == nil {
		return nil, true, l.notYet(node, "a namespace receiver property without a matching export")
	}
	local, found := l.locals[symbol]
	if !found {
		return nil, true, l.notYet(node, "a namespace receiver property without represented singleton storage")
	}
	if !identicalTypes(l.checker, l.checker.GetTypeOfSymbol(symbol), l.checker.GetTypeAtLocation(node)) {
		return nil, true, l.notYet(node, "a namespace receiver property whose type differs from its export")
	}
	value, err := l.localRead(node, local)
	return value, true, err
}
