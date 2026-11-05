package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
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
// token is admitted, and its answer is a constant only where the receiver's construction proves the
// method there: a library global's library method (Object.setPrototypeOf), a library method on a value
// Adamic holds as its own array or string (array.fill, ch.repeat), and a program class's method on a
// receiver of that class's type. Anything else, a method signature on a program's interface or object
// type above all, may be a checked view or a host value whose real object lacks the method, where Node
// says "undefined": it is not yet (methodPresenceUnproven), never folded. The constant comes once the
// receiver is evaluated, once, for its effects. Through ?. the receiver may be missing, and its
// presence is the answer. An optional method (m?(): T) stays
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
	library := l.libraryDeclaredMethod(node)
	if access.QuestionDotToken == nil && l.libraryGlobalReceiver(receiver) {
		if !library {
			return methodPresenceRead{}, true, l.notYet(node, methodPresenceUnproven)
		}
		// Object, Array, Math: reading a library global has no effects, and Object isn't a value Adamic
		// holds (libraryGlobalValue), so there is nothing to evaluate.
		return methodPresenceRead{}, true, nil
	}
	class := l.programClassMethod(node)
	if !class && !library {
		return methodPresenceRead{}, true, l.notYet(node, methodPresenceUnproven)
	}
	value, err := l.expression(access.Expression)
	if err != nil {
		return methodPresenceRead{}, true, err
	}
	if !class && value.Type() != ir.Array && value.Type() != ir.String {
		// A library method on an object Adamic didn't build as its own array or string: a Map, a host
		// value, a view. Only the object itself can say whether it has the method.
		return methodPresenceRead{}, true, l.notYet(node, methodPresenceUnproven)
	}
	if access.QuestionDotToken == nil {
		return methodPresenceRead{effects: []ir.Statement{ir.Evaluate{Value: value}}}, true, nil
	}
	switch value.Type() {
	case ir.Object, ir.Array:
		// An object or an array is falsy only when it's missing (combine's && says why), so its
		// truthiness is ?.'s nullish test, on the receiver's one evaluation.
		return methodPresenceRead{present: ir.Truthy{Value: value}}, true, nil
	}
	return methodPresenceRead{}, true, l.notYet(node, "a presence test of a method read through ?. on a "+typeName(value.Type()))
}

// methodPresenceUnproven is what a presence test stops as where nothing about how the receiver was made
// proves the method there.
const methodPresenceUnproven = "a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it"

// libraryDeclaredMethod reports whether every declaration of the method node reads is TypeScript's
// bundled library's.
func (l *lowering) libraryDeclaredMethod(node *ast.Node) bool {
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
			return false
		}
	}
	return true
}

// programClassMethod reports whether node reads, off a receiver whose type is one of the program's own
// classes, a method each declaration of which is a class's method with a body: what callOrMethod calls
// directly on such a receiver, by the class's own method. A value of a class type is never a view (every
// view's target is an object type no class is, interfaceCast to speculation), and a cast to one is
// checked by the class's identity (checkedClassCast), so such a receiver is an instance, and the method
// is on its prototype. this is read as its class, the constraint of its polymorphic this type.
func (l *lowering) programClassMethod(node *ast.Node) bool {
	receiver := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	proven := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(receiver)))
	if receiver.Kind == ast.KindThisKeyword && proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(proven); constraint != nil {
			proven = constraint
		}
	}
	if !isClassInstance(proven) {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		class := declaration.Parent
		if declaration.Kind != ast.KindMethodDeclaration || declaration.Body() == nil || class == nil || class.Name() == nil {
			return false
		}
		if _, known := l.classes[l.symbol(class.Name())]; !known {
			return false
		}
	}
	return true
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
