package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// setTimeout and clearTimeout called directly (#995spev), in the shapes tsc's sys.ts calls them:
// host.setTimeout(callback, interval, label, queue) at :377, whose extra arguments reach the
// callback, setTimeout(pollQueue, PollingInterval.High, "pollQueue") at :490, whose callback takes
// fewer than it's handed, and clearTimeout(timer) then timer = setTimeout(...) at :767 and :770. The
// Timeout is a host object (the timer provider's, timers.h) whose methods stay NotYet, unregistered,
// and reading either global as a value is a later head's.
const (
	nodeSetTimeoutName   = "node:timers.setTimeout"
	nodeClearTimeoutName = "node:timers.clearTimeout"
	nodeTimeoutName      = "node:timers.Timeout"
)

func (l *lowering) nodeTimer(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind != ast.KindCallExpression {
		return nil, false, nil
	}
	name := l.nodeLibraryMember(node)
	if name != nodeSetTimeoutName && name != nodeClearTimeoutName {
		return nil, false, nil
	}
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return nil, true, l.notYet(node, name+" through an optional call; use a direct call")
	}
	written := node.AsCallExpression().Arguments.Nodes
	for _, argument := range written {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(argument, name+" with spread arguments; write the arguments explicitly")
		}
	}
	if name == nodeClearTimeoutName {
		value, err := l.nodeClearTimeout(node, written)
		return value, true, err
	}
	value, err := l.nodeSetTimeout(node, written)
	return value, true, err
}

// nodeSetTimeout admits a callback that's one plain function the program can call, a delay that's a
// number or absent, and extra arguments each made what the callback's parameter at its position
// takes, as a call through the function value makes them. One past the callback's parameters is
// still evaluated and handed over, as Node hands it, and the callback ignores it.
func (l *lowering) nodeSetTimeout(node *ast.Node, written []*ast.Node) (ir.Expression, error) {
	if len(written) == 0 {
		return nil, l.notYet(node, nodeSetTimeoutName+" without a callback")
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(written[0]), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, l.notYet(written[0], nodeSetTimeoutName+" with a callback that isn't one plain function")
	}
	signature := signatures[0]
	if len(signature.TypeParameters()) != 0 || signature.ThisParameter() != nil || signature.HasRestParameter() {
		return nil, l.notYet(written[0], nodeSetTimeoutName+" with a generic, this-typed or rest-parameter callback")
	}
	// Node drops what a callback returns; a promise's rejection would be an unhandled one, which this
	// head doesn't report.
	if returned, known := l.representation(l.checker.GetReturnTypeOfSignature(signature)); known && returned == ir.Promise {
		return nil, l.notYet(written[0], nodeSetTimeoutName+" with a callback returning a promise")
	}
	callback, err := l.expression(written[0])
	if err != nil {
		return nil, err
	}
	if callback.Type() != ir.Closure {
		return nil, l.notYet(written[0], nodeSetTimeoutName+" with a callback that isn't a function value")
	}
	delay := ir.Expression(ir.NumberConstant{Value: 1})
	if len(written) > 1 {
		if delay, err = l.expression(written[1]); err != nil {
			return nil, err
		}
		if delay.Type() != ir.Number {
			return nil, l.notYet(written[1], nodeSetTimeoutName+" with a delay that isn't a number")
		}
	}
	arguments := []ir.Expression{callback, delay}
	parameters := signature.Parameters()
	for index, argument := range written[min(len(written), 2):] {
		value, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		if index < len(parameters) {
			parameter := parameters[index]
			takes, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
			if !known {
				return nil, l.notYet(argument, nodeSetTimeoutName+" with a callback whose parameter "+parameter.Name+" can't be held")
			}
			// The checker describes a defaulted parameter by its present type, but its incoming slot
			// must preserve undefined until the body runs.
			if len(parameter.Declarations) > 0 && parameter.Declarations[0].Kind == ast.KindParameter && parameter.Declarations[0].AsParameterDeclaration().Initializer != nil {
				takes = ir.Maybe(takes)
			}
			value = fit(value, takes)
		}
		if censusCallableSlotless(value.Type()) {
			return nil, l.notYet(argument, nodeSetTimeoutName+" passing "+typeName(value.Type())+" to its callback")
		}
		arguments = append(arguments, value)
	}
	return ir.NodeFSFile{Operation: "set_timeout", Arguments: arguments, Of: ir.Object, CallbackType: int(l.concrete(l.checker.GetTypeAtLocation(written[0])).Id())}, nil
}

// nodeClearTimeout takes any value (Adamic's ruling on #b5w3ycg, point 2, and host_axioms.json's
// row for the declaration): a Timeout is cancelled, and anything else is a no-op that doesn't throw,
// natively by the Timeout's shape. Node unenrolls a value with a truthy _onTimeout whatever made it,
// writing its fields, which a no-op wouldn't, so an object whose type has one is NotYet here, and one
// whose type hides it stops at run time on both backends (Adamic's ruling on #995spev, Nov 7;
// timers_impl.h). Node's clearTimeout returns undefined, so the call is a statement.
func (l *lowering) nodeClearTimeout(node *ast.Node, written []*ast.Node) (ir.Expression, error) {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent == nil || outer.Parent.Kind != ast.KindExpressionStatement {
		return nil, l.notYet(node, nodeClearTimeoutName+" used as a value; call it as a statement")
	}
	if len(written) != 1 {
		return nil, l.notYet(node, nodeClearTimeoutName+" outside clearTimeout(timeout)")
	}
	proven := l.checker.GetTypeAtLocation(written[0])
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	opaque := false
	for _, member := range members {
		timeout := nodeLibraryName(member.Symbol()) == nodeTimeoutName
		if member.Flags()&checker.TypeFlagsObject != 0 && !timeout && l.checker.GetPropertyOfType(member, "_onTimeout") != nil {
			return nil, l.notYet(written[0], nodeClearTimeoutName+" of an object with an _onTimeout property, which Node unenrolls")
		}
		opaque = opaque || !timeout && member.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsNonPrimitive|checker.TypeFlagsUnknown|checker.TypeFlagsTypeParameter|checker.TypeFlagsIntersection) != 0
	}
	// The run-time check reads an object's own fields and a record's keys by name. A method or an
	// accessor named _onTimeout is neither, natively, though Node reads it (calling a getter), so where
	// the program declares one and the value might be an object, it's NotYet here instead.
	if opaque && l.declaresTimeoutMember() {
		return nil, l.notYet(written[0], nodeClearTimeoutName+" of a value that may be an object with an _onTimeout method or accessor, which Node reads")
	}
	value, err := l.expression(written[0])
	if err != nil {
		return nil, err
	}
	return ir.NodeFSFile{Operation: "clear_timeout", Arguments: []ir.Expression{value}, Of: ir.Number}, nil
}

// declaresTimeoutMember says whether the program declares a method, getter or setter named _onTimeout,
// on a class or an object literal.
func (l *lowering) declaresTimeoutMember() bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier && name.Text() == "_onTimeout" {
				found = true
			}
		}
		return found || node.ForEachChild(visit)
	}
	for _, file := range l.program.Files() {
		if file.AsNode().ForEachChild(visit) {
			return true
		}
	}
	return found
}

func init() { RegisterNodeLibraryMembers(nodeSetTimeoutName, nodeClearTimeoutName) }
