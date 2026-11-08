package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// An optional call of an optional method the program declares, object?.method?.(...) as tsc's
// performance.ts writes system?.cpuProfilingEnabled?.(), and object.method?.(...), as ruled on #tvq1eqm
// (docs/rulings.md): the call short-circuits the whole chain. The receiver is evaluated once, into a
// temporary, and where it's undefined or null the chain is undefined and no argument is evaluated.
// Otherwise the method is looked up once, before the arguments, through the temporary: where the
// object has none by that name, or its field holds undefined, the call is undefined too, with no
// argument evaluated (ir.CallClosure's Optional), and where it has one, that one is called with the
// temporary as this. The result is R | undefined, held as that type's undefined, never a zero.
//
// The call ends the chain: more of it after the call (o?.m?.().x) is property's NotYet, and a call
// through ?. of anything else (a function value, a library method, a method behind a longer chain or
// in parentheses) is optionalCall's. A present value that isn't a function, check 49, reaches the slot
// only through a cast to a checked view, whose real object may hold anything there, and such a read is
// stopped (unemittableCheckedViewRead), so the check has nothing to fire on; in .a the types prove it
// away.

// optionalMethod is an optional method call, lowered: receiver declares the temporary the receiver is
// evaluated into and present tests it, both nil for object.method?.(...), whose receiver isn't tested.
type optionalMethod struct {
	receiver []ir.Statement
	present  ir.Expression
	call     ir.CallClosure
}

// optionalMethodCall lowers node when it's object?.method?.(...) or object.method?.(...) on an optional
// method the program declares; handled is false for any other call, which lowers as it always has.
func (l *lowering) optionalMethodCall(node *ast.Node) (optionalMethod, bool, error) {
	signature := l.optionalMethodSignature(node)
	if signature == nil {
		return optionalMethod{}, false, nil
	}
	lowered, err := l.callClosure(node)
	if err != nil {
		return optionalMethod{}, true, err
	}
	call, isCall := lowered.(ir.CallClosure)
	if !isCall {
		return optionalMethod{}, true, l.notYet(node, optionalCallNotYet)
	}
	property, isProperty := call.Closure.(ir.Property)
	if !isProperty || !property.Method || property.Object.Type() != ir.Object || !optionalMethodResult(call.Returns) {
		return optionalMethod{}, true, l.notYet(node, optionalCallNotYet)
	}
	if l.checker.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsVoid != 0 && call.Returns != 0 {
		// A void method gives back nothing to read as R | undefined.
		return optionalMethod{}, true, l.notYet(node, optionalCallNotYet)
	}
	call.Optional = true
	if node.AsCallExpression().Expression.AsPropertyAccessExpression().QuestionDotToken == nil {
		return optionalMethod{call: call}, true, nil
	}
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optional_call_receiver", Type: ir.Object, Function: function, ExpressionAssigned: true})
	held := ir.Read{Local: local, Of: ir.Object}
	receiver := []ir.Statement{ir.Declare{Local: local, Value: property.Object}}
	property.Object, property.Optional = held, false
	call.Closure = property
	present := ir.Unary{Operator: ir.Not, Operand: ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: held}, Right: ir.IsNull{Value: held}}}
	return optionalMethod{receiver: receiver, present: present, call: call}, true, nil
}

// optionalMethodSignature is the one call signature of the method node calls when node is
// object?.method?.(...) or object.method?.(...), written so, with no parentheses around the callee and
// a receiver that isn't itself in an optional chain, on a method the program declares once, in an
// interface or an object type, as a method signature or a field holding a function, marked ?, that
// can't be null. It's nil for any other call.
func (l *lowering) optionalMethodSignature(node *ast.Node) *checker.Signature {
	call := node.AsCallExpression()
	if call.QuestionDotToken == nil || call.Expression.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	receiver := call.Expression.AsPropertyAccessExpression().Expression
	if receiver.Flags&ast.NodeFlagsOptionalChain != 0 || ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword {
		return nil
	}
	method := l.checker.GetSymbolAtLocation(call.Expression)
	if method == nil || method.Flags&ast.SymbolFlagsOptional == 0 || len(method.Declarations) != 1 {
		return nil
	}
	declaration := method.Declarations[0]
	if (declaration.Kind != ast.KindMethodSignature && declaration.Kind != ast.KindPropertySignature) || !ast.HasQuestionToken(declaration) || load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
		return nil
	}
	declared := l.checker.GetTypeOfSymbol(method)
	if l.includesNull(declared) {
		return nil
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetNonNullableType(declared), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil
	}
	return signatures[0]
}

// optionalMethodResult is whether a call's undefined has a place in its result's representation: void
// has none to give, a number or a boolean has its missing pair, and a string, an object, an array, a
// map or a function its null reference. A Union's undefined isn't built here.
func optionalMethodResult(of ir.Type) bool {
	switch of {
	case 0, ir.MaybeNumber, ir.MaybeBoolean, ir.String, ir.Object, ir.Array, ir.Map, ir.Closure:
		return true
	}
	return false
}

// optionalCallNotYet is what a call through ?. that isn't lowered stops as.
const optionalCallNotYet = "a call through ?. (an optional call)"

// value is the call as a value: undefined, with nothing after the receiver evaluated, where the
// receiver isn't there.
func (optional optionalMethod) value() ir.Expression {
	if optional.call.Returns == 0 {
		return ir.Effects{Body: optional.statements(), Result: ir.Undefined{}}
	}
	if optional.present == nil {
		return optional.call
	}
	return ir.Effects{Body: optional.receiver, Result: ir.Conditional{
		Condition: optional.present, WhenTrue: optional.call, WhenNot: fit(ir.Undefined{}, optional.call.Returns), Of: optional.call.Returns,
	}}
}

// statements is the call as a statement, its value dropped.
func (optional optionalMethod) statements() []ir.Statement {
	evaluate := ir.Evaluate{Value: optional.call}
	if optional.present == nil {
		return []ir.Statement{evaluate}
	}
	return append(append([]ir.Statement{}, optional.receiver...), ir.If{Condition: optional.present, Then: []ir.Statement{evaluate}})
}
