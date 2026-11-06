package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// The global object as a value is not yet (#00qn5az head 3, Adamic's ruling): any expression whose type
// is typeof globalThis, whatever name reaches it (globalThis, Node's global, globalThis.globalThis, in
// parentheses), is admitted only as the receiver of a property or element access, which a read, a member
// call and a write all are, and a write there is judged by rule 50 (hostBindingGlobalThisRefusal) against
// the host-binding table. typeof of the object is admitted only where a comparison with a literal string
// is all that reads it. Every other use stops here, one not-yet: an alias, an argument (Reflect.set,
// Object.assign, Object.defineProperty), an initializer, a return, a spread, a widening to object, {} or
// a record view, an element of a container, a condition, a comparison.
//
// It isn't refused: passing the global object around is valid JavaScript Node runs, and no type hatch.
// But a value of it can install a global, Reflect.set(globalThis, "onProfilerEvent", f) above all, and
// an Absent host binding folds typeof to "undefined" natively while Node then says "function": a silent
// divergence. So lifting any part of this stop has to answer the host-binding fold first (#00qn5az,
// docs/rulings.md's entry "A `.ts` file's ambient value declaration is a host binding";
// host_bindings.go). TestGlobalObjectAsAValueIsNotYet pins every road it closes today.
//
// The rule runs at ordinal 60, before every rule but rule 50, and judges a node's children as well as the
// node itself, so it speaks at the parent's visit: ahead of recordStorageView (2000), which stops a record
// view at the declaration that holds the value, and of nodeLibraryRefusal (800) and namespaceRefusal
// (2100), which meet global and globalThis on their own visits; "reading globalThis" is lowering's, after
// every rule. Unknown reach fails closed: every expression is judged by its type, whatever its kind, save
// the kinds whose type can never be the global object's.

// globalObjectValueStop is what the global object as a value stops as.
const globalObjectValueStop = "the global object as a value (only a property read's or member call's receiver is admitted)"

// globalObject reports whether an expression's type is typeof globalThis: the checker's own globalThis
// symbol is that type's symbol, whatever name the expression reads it by.
func (l *lowering) globalObject(node *ast.Node) bool {
	global := l.checker.GetGlobalSymbol("globalThis", ast.SymbolFlagsModule, nil)
	return global != nil && l.checker.GetTypeAtLocation(node).Symbol() == global
}

// globalObjectValueRefusal stops the first use of the global object as a value among node's children,
// then node itself, which a walk that starts at node (refuseNode) reaches with no parent's visit first.
func (l *lowering) globalObjectValueRefusal(node *ast.Node) error {
	var stop error
	node.ForEachChild(func(child *ast.Node) bool {
		stop = l.globalObjectValueUse(child)
		return stop != nil
	})
	if stop != nil {
		return stop
	}
	return l.globalObjectValueUse(node)
}

// globalObjectValueUse is the stop for a node that is the global object used as a value, or nil.
func (l *lowering) globalObjectValueUse(node *ast.Node) error {
	parent := node.Parent
	if parent == nil || globalObjectAdmitted(node) {
		return nil
	}
	if parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node {
		// { globalThis } puts the value in an object; the name is no expression of its own to the checker.
		global := l.checker.GetGlobalSymbol("globalThis", ast.SymbolFlagsModule, nil)
		value := l.checker.GetShorthandAssignmentValueSymbol(parent)
		if global != nil && value != nil && (value == global || l.checker.GetTypeOfSymbolAtLocation(value, node).Symbol() == global) {
			return l.notYet(node, globalObjectValueStop)
		}
		return nil
	}
	if !ast.IsExpressionNode(node) || ast.IsPartOfTypeNode(node) || inTypeQuery(node) || parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Name() == node {
		return nil
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindTemplateExpression, ast.KindTypeOfExpression, ast.KindVoidExpression,
		ast.KindDeleteExpression, ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression, ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression, ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindClassExpression:
		// Their type is a primitive, or an object or function they make: never the global object.
		return nil
	}
	if !l.globalObject(node) {
		return nil
	}
	return l.notYet(node, globalObjectValueStop)
}

// inTypeQuery reports whether node names a value only for a type: `typeof globalThis` written as a type,
// which the checker counts as an expression but which nothing runs.
func inTypeQuery(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && (parent.Kind == ast.KindQualifiedName || parent.Kind == ast.KindPropertyAccessExpression) {
		parent = parent.Parent
	}
	return parent != nil && parent.Kind == ast.KindTypeQuery
}

// globalObjectAdmitted reports whether what consumes a value of the global object is one of the uses
// admitted: the receiver of a property or element access, parentheses (judged as the value itself, by
// their own consumer), and typeof whose answer only a comparison with a literal string reads.
func globalObjectAdmitted(value *ast.Node) bool {
	parent := value.Parent
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Expression == value
	case ast.KindElementAccessExpression:
		return parent.AsElementAccessExpression().Expression == value
	case ast.KindParenthesizedExpression:
		return true
	case ast.KindTypeOfExpression:
		outer := parent
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		comparison := outer.Parent
		if comparison == nil || comparison.Kind != ast.KindBinaryExpression {
			return false
		}
		binary := comparison.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
		default:
			return false
		}
		other := binary.Right
		if other == outer {
			other = binary.Left
		}
		other = ast.SkipParentheses(other)
		return other.Kind == ast.KindStringLiteral || other.Kind == ast.KindNoSubstitutionTemplateLiteral
	}
	return false
}

func init() {
	refusalRules.register(60, "globalObjectValueRefusal", (*lowering).globalObjectValueRefusal)
}
