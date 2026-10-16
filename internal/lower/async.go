package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The only async-specific lowering is the await expression and library promise leaves.
// Statements, calls, methods and closures use their ordinary lowerers.
func (l *lowering) awaitExpression(node *ast.Node) (ir.Expression, error) {
	if l.function == nil && ast.GetSourceFileOfNode(node).AsSourceFile() != l.entry {
		return nil, l.notYet(node, "top-level await outside the entry module (asynchronous module evaluation)")
	}
	value, err := l.expression(node.AsAwaitExpression().Expression)
	if err != nil {
		return nil, err
	}
	proven := l.checker.GetTypeAtLocation(node)
	var of ir.Type
	if proven.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined|checker.TypeFlagsNever) == 0 {
		var known bool
		of, known = l.representation(proven)
		if !known {
			return nil, l.notYet(node, "await of an unrepresented payload or thenable")
		}
	}
	if value.Type() != ir.Promise {
		if l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(node.AsAwaitExpression().Expression), "then") != nil {
			return nil, l.notYet(node, "await of thenables")
		}
		if !l.asyncPayload(node.AsAwaitExpression().Expression, map[*ast.Symbol]bool{}) {
			return nil, l.notYet(node, "await of a non-primitive structurally typed value (hidden thenable)")
		}
		value = ir.PromiseValue{Value: value}
	}
	return ir.Await{Value: value, Of: of, Site: l.asyncSite(node)}, nil
}

func (l *lowering) promiseValue(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Promise") {
		return nil, false, nil
	}
	name := callee.Name().Text()
	if (name != "resolve" && name != "reject") || len(call.Arguments.Nodes) > 1 {
		return nil, true, l.notYet(node, "Promise.all and the unproved Promise surface")
	}
	var value ir.Expression
	if len(call.Arguments.Nodes) == 1 {
		var err error
		value, err = l.expression(call.Arguments.Nodes[0])
		if err != nil {
			return nil, true, err
		}
		if l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(call.Arguments.Nodes[0]), "then") != nil {
			return nil, true, l.notYet(node, "Promise thenables")
		}
		if value.Type() == ir.Promise {
			return nil, true, l.notYet(node, "Promise adoption")
		}
		if name == "resolve" && !l.asyncPayload(call.Arguments.Nodes[0], map[*ast.Symbol]bool{}) {
			return nil, true, l.notYet(node, "Promise.resolve of a non-primitive structurally typed value (hidden thenable)")
		}
		if name == "reject" && !l.isLibraryType(l.checker.GetTypeAtLocation(call.Arguments.Nodes[0]), "Error") {
			return nil, true, l.notYet(node, "Promise rejection with a non-Error")
		}
	}
	if name == "reject" && value == nil {
		return nil, true, l.notYet(node, "Promise rejection without an Error")
	}
	return ir.PromiseValue{Value: value, Reject: name == "reject"}, true, nil
}

// Both statement returns and expression arrows need the same adoption boundary.
func (l *lowering) checkAsyncReturn(node *ast.Node, value ir.Expression) error {
	if !l.function.Async {
		return nil
	}
	if value.Type() == ir.Promise {
		return l.notYet(node, "return of a Promise without await (Promise adoption)")
	}
	if l.hasThen(l.checker.GetTypeAtLocation(node)) {
		return l.notYet(node, "return of thenables (thenable adoption)")
	}
	if !l.asyncPayload(node, map[*ast.Symbol]bool{}) {
		return l.notYet(node, "async return of a non-primitive structurally typed value (hidden thenable)")
	}
	return nil
}

func (l *lowering) hasThen(proven *checker.Type) bool {
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if l.hasThen(member) {
				return true
			}
		}
	}
	return l.checker.GetPropertyOfType(proven, "then") != nil
}

// Diagnostic sites are integers in IR, never machine paths or checker pointers.
func (l *lowering) asyncSite(node *ast.Node) int {
	l.asyncSites = append(l.asyncSites, node)
	return len(l.asyncSites)
}

// A structural type can hide then. Only primitive types or a closed allocation
// reached through immutable bindings prove that Promise assimilation does nothing.
func (l *lowering) asyncPayload(node *ast.Node, seen map[*ast.Symbol]bool) bool {
	node = ast.SkipParentheses(node)
	proven := l.checker.GetTypeAtLocation(node)
	if l.hasThen(proven) {
		return false
	}
	var primitive func(*checker.Type) bool
	primitive = func(of *checker.Type) bool {
		if of.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range of.Types() {
				if !primitive(member) {
					return false
				}
			}
			return true
		}
		return of.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid|checker.TypeFlagsNever) != 0
	}
	if primitive(proven) {
		return true
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression:
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind == ast.KindSpreadAssignment && !l.asyncPayload(property.AsSpreadAssignment().Expression, seen) {
				return false
			}
		}
		return true
	case ast.KindArrayLiteralExpression, ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindNewExpression:
		return true
	case ast.KindAwaitExpression:
		// The library Promise's producers are independently checked at resolve/return.
		return l.isLibraryType(l.checker.GetTypeAtLocation(node.AsAwaitExpression().Expression), "Promise")
	case ast.KindConditionalExpression:
		choice := node.AsConditionalExpression()
		return l.asyncPayload(choice.WhenTrue, seen) && l.asyncPayload(choice.WhenFalse, seen)
	case ast.KindIdentifier:
		symbol := l.symbol(node)
		if symbol == nil || seen[symbol] || len(symbol.Declarations) != 1 {
			return false
		}
		seen[symbol] = true
		defer delete(seen, symbol)
		declaration := symbol.Declarations[0]
		if declaration.Kind == ast.KindFunctionDeclaration {
			return true
		}
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Initializer() == nil || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return false
		}
		return l.asyncPayload(declaration.Initializer(), seen)
	}
	return false
}
