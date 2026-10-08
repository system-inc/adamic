package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The only async-specific lowering is the await expression and library promise leaves.
// Statements, calls, methods and closures use their ordinary lowerers.
func (l *lowering) awaitExpression(node *ast.Node) (ir.Expression, error) {
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
		value = ir.PromiseValue{Value: value}
	}
	return ir.Await{Value: value, Of: of}, nil
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
