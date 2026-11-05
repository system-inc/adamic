package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A method read whose only consumer tests for it is an observation, not a detach (#4q1b2er): the
// operand of typeof or of !, a condition (the test of if, while, do...while, for, and of ?:), and the
// left of &&, whose falsy left would be the whole value, while a method value is never falsy, so
// nothing detached escapes it. tsc probes its host this way (`typeof Object.setPrototypeOf ===
// "function"`, `if (array.fill)`, `if (ch.repeat)`). Every other consumer keeps the method value and
// stays refused as a method read as a value (methodValueRefusal): ||, ??, an arm of ?:, an assignment
// or an initializer, an argument, a return, an alias and a comma operand.
//
// Only a method each of whose declarations is a method declaration or signature with no question
// token is admitted. On a value the checker proved to have its type such a method is always there, so
// the test's answer is a constant once the receiver is evaluated, once, for its effects. Through ?.
// the receiver may be missing, and its presence is the answer. An optional method (m?(): T) stays
// refused, since nothing proves it there and nothing yet reads whether it is: its read goes the way an
// optional function-typed field's does (property), and that stops for every method behind it, an
// object literal's as a literal method through a view that erases its receiver, a class's as a class
// method through a view that erases its prototype origin.

// methodPresenceTest reports whether node reads a method that is always there only to test for it.
// A read the library observations already admit (typeof String.raw, typeof Array.prototype.map,
// typeof performance.now) keeps the lowering it has.
func (l *lowering) methodPresenceTest(node *ast.Node) bool {
	if node.Kind != ast.KindPropertyAccessExpression || !l.alwaysPresentMethod(node) || !presenceConsumer(node) {
		return false
	}
	return !l.stringMethodObservation(node) && !l.libraryArrayObservedMethod(node) && !l.nodeProcessMethodObservation(node)
}

// alwaysPresentMethod reports whether a property access reads a method every declaration of which is a
// method declaration or signature that isn't optional.
func (l *lowering) alwaysPresentMethod(node *ast.Node) bool {
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindMethodDeclaration && declaration.Kind != ast.KindMethodSignature || ast.HasQuestionToken(declaration) {
			return false
		}
	}
	return true
}

// presenceConsumer reports whether what consumes node, through any parentheses around it, only tests
// it: typeof, !, a condition, or the left of &&.
func presenceConsumer(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindTypeOfExpression:
		return true
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == node
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == node
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == node
	case ast.KindForStatement:
		return parent.AsForStatement().Condition == node
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == node
	case ast.KindBinaryExpression:
		return parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken && parent.AsBinaryExpression().Left == node
	}
	return false
}

// methodPresenceRead is what testing for a method comes to: the receiver's evaluation, kept for its
// effects, and, read through ?., whether the receiver is there, which is then the answer. present is
// nil when the method is always there.
type methodPresenceRead struct {
	effects []ir.Statement
	present ir.Expression
}

// then is result after the receiver's effects.
func (read methodPresenceRead) then(result ir.Expression) ir.Expression {
	if len(read.effects) == 0 {
		return result
	}
	return ir.Effects{Body: read.effects, Result: result}
}

// methodPresence lowers operand's receiver when operand is a method read only to test for it; handled
// is false for any other operand, which its consumer lowers as it always has.
func (l *lowering) methodPresence(operand *ast.Node) (read methodPresenceRead, handled bool, err error) {
	node := ast.SkipParentheses(operand)
	if !l.methodPresenceTest(node) {
		return methodPresenceRead{}, false, nil
	}
	access := node.AsPropertyAccessExpression()
	receiver := ast.SkipParentheses(access.Expression)
	if access.QuestionDotToken == nil && node.Flags&ast.NodeFlagsOptionalChain != 0 {
		// The rest of a chain after a ?., which short-circuits with it, as property stops it.
		return methodPresenceRead{}, true, l.notYet(node, "a presence test of a method at the end of an optional chain longer than one step")
	}
	if receiver.Kind == ast.KindSuperKeyword {
		return methodPresenceRead{}, true, l.notYet(node, "a presence test of a method read through super")
	}
	if access.QuestionDotToken == nil && l.libraryGlobalReceiver(receiver) {
		// Object, Array, Math: reading a library global has no effects, and Object isn't a value Adamic
		// holds (libraryGlobalValue), so there is nothing to evaluate.
		return methodPresenceRead{}, true, nil
	}
	value, err := l.expression(access.Expression)
	if err != nil {
		return methodPresenceRead{}, true, err
	}
	if access.QuestionDotToken == nil {
		return methodPresenceRead{effects: []ir.Statement{ir.Evaluate{Value: value}}}, true, nil
	}
	switch value.Type() {
	case ir.Object, ir.Array, ir.Map:
		// An object, an array or a Map is falsy only when it's missing (combine's && says why), so its
		// truthiness is ?.'s nullish test, on the receiver's one evaluation.
		return methodPresenceRead{present: ir.Truthy{Value: value}}, true, nil
	}
	return methodPresenceRead{}, true, l.notYet(node, "a presence test of a method read through ?. on a "+typeName(value.Type()))
}

// libraryGlobalReceiver reports whether node names a library global, as isLibraryGlobal does for one
// name.
func (l *lowering) libraryGlobalReceiver(node *ast.Node) bool {
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	return symbol != nil && len(symbol.Declarations) > 0 && load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// methodPresenceTypeOf lowers typeof of a method read only to test for it: "function", or through ?.
// "undefined" where the receiver is missing.
func (l *lowering) methodPresenceTypeOf(node *ast.Node) (ir.Expression, bool, error) {
	read, handled, err := l.methodPresence(node.AsTypeOfExpression().Expression)
	if !handled || err != nil {
		return nil, handled, err
	}
	function := ir.StringConstant{Index: l.constant("function")}
	if read.present == nil {
		return read.then(function), true, nil
	}
	return ir.Conditional{Condition: read.present, WhenTrue: function, WhenNot: ir.StringConstant{Index: l.constant("undefined")}}, true, nil
}

// methodPresenceCondition lowers a condition that is a method read only to test for it: true, or
// through ?. whether the receiver is there.
func (l *lowering) methodPresenceCondition(node *ast.Node) (ir.Expression, bool, error) {
	read, handled, err := l.methodPresence(node)
	if !handled || err != nil {
		return nil, handled, err
	}
	if read.present == nil {
		return read.then(ir.BooleanConstant{Value: true}), true, nil
	}
	return read.present, true, nil
}

// methodPresenceNot lowers ! of a method read only to test for it: false, or through ?. whether the
// receiver is missing.
func (l *lowering) methodPresenceNot(operand *ast.Node) (ir.Expression, bool, error) {
	read, handled, err := l.methodPresence(operand)
	if !handled || err != nil {
		return nil, handled, err
	}
	if read.present == nil {
		return read.then(ir.BooleanConstant{Value: false}), true, nil
	}
	return ir.Unary{Operator: ir.Not, Operand: read.present}, true, nil
}

// methodPresenceAnd lowers left && right whose left is a method read only to test for it: right's
// value, after the receiver's effects.
func (l *lowering) methodPresenceAnd(node *ast.Node) (ir.Expression, bool, error) {
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
		// || and ?? keep the method value: it is their whole value, never right's.
		return nil, false, nil
	}
	read, handled, err := l.methodPresence(binary.Left)
	if !handled || err != nil {
		return nil, handled, err
	}
	if read.present != nil {
		// The receiver's undefined would be the whole value, a slot right's representation may not hold.
		return nil, true, l.notYet(node, "the left of && reading a method through ?., whose missing receiver would be the value")
	}
	right, err := l.expression(binary.Right)
	if err != nil {
		return nil, true, err
	}
	return read.then(right), true, nil
}
