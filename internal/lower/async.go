package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The only async-specific lowering is the await expression and library promise leaves.
// Statements, calls, methods and closures use their ordinary lowerers.
func (lowering *lowering) awaitExpression(node *ast.Node) (ir.Expression, error) {
	if lowering.function == nil && ast.GetSourceFileOfNode(node).AsSourceFile() != lowering.entry {
		return nil, lowering.notYet(node, "top-level await outside the entry module (asynchronous module evaluation)")
	}
	value, expressionError := lowering.expression(node.AsAwaitExpression().Expression)
	if expressionError != nil {
		return nil, expressionError
	}
	proven := lowering.checker.GetTypeAtLocation(node)
	var of ir.Type
	if proven.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined|checker.TypeFlagsNever) == 0 {
		var known bool
		of, known = lowering.representation(proven)
		if !known {
			return nil, lowering.notYet(node, "await of an unrepresented payload or thenable")
		}
	}
	if value.Type() != ir.Promise {
		if lowering.checker.GetPropertyOfType(
			lowering.checker.GetTypeAtLocation(node.AsAwaitExpression().Expression),
			"then",
		) != nil {
			return nil, lowering.notYet(node, "await of thenables")
		}
		if !lowering.asyncPayload(node.AsAwaitExpression().Expression) {
			lowering.asyncPayloadFailures = append(
				lowering.asyncPayloadFailures,
				lowering.notYet(node, "await of a non-primitive structurally typed value (hidden thenable)"),
			)
		}
		value = ir.PromiseValue{Value: value}
	}
	return ir.Await{Value: value, Of: of, Site: lowering.asyncSite(node)}, nil
}

func (lowering *lowering) promiseValue(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression ||
		!lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Promise") {
		return nil, false, nil
	}
	name := callee.Name().Text()
	if (name != "resolve" && name != "reject") || len(call.Arguments.Nodes) > 1 {
		return nil, true, lowering.notYet(node, "Promise.all and the unproved Promise surface")
	}
	var value ir.Expression
	if len(call.Arguments.Nodes) == 1 {
		var expressionError error
		value, expressionError = lowering.expression(call.Arguments.Nodes[0])
		if expressionError != nil {
			return nil, true, expressionError
		}
		if lowering.checker.GetPropertyOfType(
			lowering.checker.GetTypeAtLocation(call.Arguments.Nodes[0]),
			"then",
		) != nil {
			return nil, true, lowering.notYet(node, "Promise thenables")
		}
		if value.Type() == ir.Promise {
			return nil, true, lowering.notYet(node, "Promise adoption")
		}
		if name == "resolve" && !lowering.asyncPayload(call.Arguments.Nodes[0]) {
			lowering.asyncPayloadFailures = append(
				lowering.asyncPayloadFailures,
				lowering.notYet(node, "Promise.resolve of a non-primitive structurally typed value (hidden thenable)"),
			)
		}
		if name == "reject" &&
			!lowering.isLibraryType(lowering.checker.GetTypeAtLocation(call.Arguments.Nodes[0]), "Error") {
			return nil, true, lowering.notYet(node, "Promise rejection with a non-Error")
		}
	}
	if name == "reject" && value == nil {
		return nil, true, lowering.notYet(node, "Promise rejection without an Error")
	}
	return ir.PromiseValue{Value: value, Reject: name == "reject"}, true, nil
}

// Both statement returns and expression arrows need the same adoption boundary.
func (lowering *lowering) checkAsyncReturn(node *ast.Node, value ir.Expression) error {
	if !lowering.function.Async {
		return nil
	}
	if value.Type() == ir.Promise {
		return lowering.notYet(node, "return of a Promise without await (Promise adoption)")
	}
	if literal := ast.SkipParentheses(node); literal.Kind == ast.KindObjectLiteralExpression {
		for _, property := range literal.AsObjectLiteralExpression().Properties.Nodes {
			if name := property.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
				return lowering.notYet(name, "async return with a computed field name (possible thenable adoption)")
			}
		}
	}
	if lowering.hasThen(lowering.checker.GetTypeAtLocation(node)) {
		return lowering.notYet(node, "return of thenables (thenable adoption)")
	}
	if !lowering.asyncPayload(node) {
		lowering.asyncPayloadFailures = append(
			lowering.asyncPayloadFailures,
			lowering.notYet(node, "async return of a non-primitive structurally typed value (hidden thenable)"),
		)
	}
	return nil
}

func (lowering *lowering) hasThen(proven *checker.Type) bool {
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if lowering.hasThen(member) {
				return true
			}
		}
	}
	return lowering.checker.GetPropertyOfType(proven, "then") != nil
}

// Diagnostic sites are integers in IR, never machine paths or checker pointers.
func (lowering *lowering) asyncSite(node *ast.Node) int {
	lowering.asyncSites = append(lowering.asyncSites, node)
	return len(lowering.asyncSites)
}

// A structural type can hide then. A const binding does not freeze its object:
// Object.assign or an escaped alias can add then before adoption. Only primitives
// or fresh allocations at the adoption point prove assimilation does nothing.
func (lowering *lowering) asyncPayload(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	proven := lowering.checker.GetTypeAtLocation(node)
	if lowering.hasThen(proven) {
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
			// A record's computed key may be "then" even when its structural
			// type exposes no named property to the checker.
			if name := property.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
				return false
			}
			if property.Kind == ast.KindSpreadAssignment &&
				!lowering.asyncPayload(property.AsSpreadAssignment().Expression) {
				return false
			}
		}
		return true
	case ast.KindArrayLiteralExpression, ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindFunctionDeclaration:
		return true
	case ast.KindNewExpression:
		callee := ast.SkipParentheses(node.AsNewExpression().Expression)
		if lowering.isLibraryGlobal(callee, "Error") ||
			lowering.isLibraryGlobal(callee, "Map") ||
			lowering.isLibraryGlobal(callee, "Set") {
			return true
		}
		// A class with an index signature, its own or a base's, can write this[key] with key
		// "then", which hasThen can't see: the checker's properties never list index keys. Today
		// such a class stops earlier, at the index signature's own NotYet (indexSignatureStorageGap), so
		// this holds the boundary for the day classes may carry one.
		symbol := lowering.symbol(callee)
		return symbol != nil &&
			len(symbol.Declarations) == 1 &&
			symbol.Declarations[0].Kind == ast.KindClassDeclaration &&
			len(lowering.checker.GetIndexInfosOfType(proven)) == 0
	case ast.KindAwaitExpression:
		// The library Promise's producers are independently checked at resolve/return.
		return lowering.isLibraryType(
			lowering.checker.GetTypeAtLocation(node.AsAwaitExpression().Expression),
			"Promise",
		)
	case ast.KindConditionalExpression:
		choice := node.AsConditionalExpression()
		return lowering.asyncPayload(choice.WhenTrue) && lowering.asyncPayload(choice.WhenFalse)

	}
	return false
}
