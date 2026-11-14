package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// An optional call of an optional method the program declares, object?.method?.(...) as tsc's
// performance.ts writes system?.cpuProfilingEnabled?.(), and object.method?.(...), as ruled on #tvq1eqm
// (docs/rulings.md): the call short-circuits the whole chain. The receiver is evaluated once, by the
// call itself, and where it's undefined or null the chain is undefined and no argument is evaluated
// (ir.Property's Optional, the ?. before the method, which native tests as the null pointer both are
// in an object's slot). Otherwise the method is looked up once, before the arguments: where the object
// has none by that name, or its field holds undefined, the call is undefined too, with no argument
// evaluated (ir.CallClosure's Optional), and where it has one, that one is called with the receiver as
// this. The result is R | undefined, held as that type's undefined, never a zero.
//
// The receiver isn't a temporary of lowering's: a local declared inside an expression is held to the
// end of the statement around it, so in a short-circuit's arm (tsc's own, the right of ||) its C
// declaration sits in the arm's block and its release outside it. The call's own receiver is a C
// temporary of the emitter's, scoped with the arm.
//
// The call ends the chain: more of it after the call (o?.m?.().x) is property's NotYet, and a call
// through ?. of anything else (a function value, a library method, a method behind a longer chain or
// in parentheses) is optionalCall's. A present value that isn't a function, check 49, reaches the slot
// only through a cast to a checked view, whose real object may hold anything there, and such a read is
// stopped (unemittableCheckedViewRead), so the check has nothing to fire on; in .a the types prove it
// away.

// optionalMethod is an optional method call, lowered.
type optionalMethod struct {
	call ir.CallClosure
}

// optionalMethodCall lowers node when it's object?.method?.(...) or object.method?.(...) on an optional
// method the program declares, or callback?.(...) on a function value (optionalValueSignature); handled
// is false for any other call, which lowers as it always has.
func (l *lowering) optionalMethodCall(node *ast.Node) (optionalMethod, bool, error) {
	signature := l.optionalMethodSignature(node)
	if signature == nil {
		if signature = l.optionalValueSignature(node); signature != nil {
			return l.optionalValueCall(node, signature)
		}
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
	// ir.MethodCallee, so a narrowed callee ('m' in o, then o.m?.()) is the same method call: the emitters
	// test the call's own ?. on what the lookup finds before the narrowing's check (#6na4x1b).
	property, _, isMethod := ir.MethodCallee(call.Closure)
	if !isMethod || property.Object.Type() != ir.Object || !optionalMethodResult(call.Returns) {
		return optionalMethod{}, true, l.notYet(node, optionalCallNotYet)
	}
	if l.checker.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsVoid != 0 {
		// The checker can type an optional call of a void method void | undefined, which callClosure
		// reads as an object's result. The method gives back nothing to read, and a read of the call is
		// refused as a read of a void value, so it's a void call.
		call.Returns = 0
	}
	call.Optional = true
	return optionalMethod{call: call}, true, nil
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

// An optional call of a function value, callback?.(...), as tsc's sys.ts pollWatchedFileQueue calls its
// optional parameter callbackOnWatchFileStat?.(...) and emitter.ts's printer its captured handlers
// (#036ve6w). The callee is read once, before the arguments (ir.CallClosure's Optional on a plain
// callee, which both emitters hold in one temporary): where it's undefined the call is undefined and no
// argument is evaluated, so an argument that writes the binding still calls what was read. Otherwise
// it's an ordinary call of that function value, with no this.

// optionalValueSignature is the one call signature of the function value node calls when node is
// callback?.(...), written so, with no parentheses around the callee, on an identifier naming a
// parameter or a variable (captured or not), whose type holds undefined and no null, and whose other
// members are one function type with one call signature and no construct signature. It's nil for any
// other call: a field's o.f?.() is optionalMethodSignature's, and an element, a chain, a call's result,
// a null-holding or a union callee stay optionalCall's NotYet.
func (l *lowering) optionalValueSignature(node *ast.Node) *checker.Signature {
	call := node.AsCallExpression()
	if call.QuestionDotToken == nil || call.Expression.Kind != ast.KindIdentifier {
		return nil
	}
	symbol := l.symbol(call.Expression)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsVariable == 0 || len(symbol.Declarations) == 0 {
		return nil
	}
	declared := l.checker.GetTypeAtLocation(call.Expression)
	if l.includesNull(declared) || !l.includesUndefined(declared) {
		return nil
	}
	present := l.checker.GetNonNullableType(declared)
	if present.Flags()&checker.TypeFlagsUnion != 0 || len(l.checker.GetSignaturesOfType(present, checker.SignatureKindConstruct)) != 0 {
		return nil
	}
	signatures := l.checker.GetSignaturesOfType(present, checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil
	}
	return signatures[0]
}

// optionalValueCall lowers callback?.(...) once optionalValueSignature has said it qualifies: the call
// as callClosure lowers it, held as a plain function value and marked Optional, with its result's
// undefined given a place (optionalMethodResult) and a void callback's call made a void call.
func (l *lowering) optionalValueCall(node *ast.Node, signature *checker.Signature) (optionalMethod, bool, error) {
	lowered, err := l.callClosure(node)
	if err != nil {
		return optionalMethod{}, true, err
	}
	call, isCall := lowered.(ir.CallClosure)
	if !isCall || call.Bound || call.Direct > 0 || call.Closure.Type() != ir.Closure || !optionalMethodResult(call.Returns) {
		return optionalMethod{}, true, l.notYet(node, optionalCallNotYet)
	}
	if _, _, isMethod := ir.MethodCallee(call.Closure); isMethod {
		return optionalMethod{}, true, l.notYet(node, optionalCallNotYet)
	}
	if l.checker.GetReturnTypeOfSignature(signature).Flags()&checker.TypeFlagsVoid != 0 {
		call.Returns = 0
	}
	call.Optional = true
	return optionalMethod{call: call}, true, nil
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

// value is the call as a value, undefined where it stops; a void call is undefined after it's made,
// as expression lowers every void call.
func (optional optionalMethod) value() ir.Expression {
	if optional.call.Returns == 0 {
		return ir.Effects{Body: optional.statements(), Result: ir.Undefined{}}
	}
	return optional.call
}

// statements is the call as a statement, its value dropped.
func (optional optionalMethod) statements() []ir.Statement {
	return []ir.Statement{ir.Evaluate{Value: optional.call}}
}
