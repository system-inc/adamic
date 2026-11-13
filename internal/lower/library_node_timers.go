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
// Timeout is a host object (the timer provider's, timers.h) whose methods stay NotYet, unregistered.
// Either global read as a value is nodeTimerValue's.
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
	// head doesn't report. A union may hold one too.
	returned := l.checker.GetReturnTypeOfSignature(signature)
	results := []*checker.Type{returned}
	if returned.Flags()&checker.TypeFlagsUnion != 0 {
		results = returned.Types()
	}
	for _, result := range results {
		if of, known := l.representation(result); known && of == ir.Promise {
			return nil, l.notYet(written[0], nodeSetTimeoutName+" with a callback returning a promise")
		}
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

// The one view each timer's value is admitted in (Adamic's ruling on #995spev, Nov 13): System's
// timer slots as the adapted sys.ts declares them, concrete, so the value's closure has one calling
// convention. Node's setTimeout is variadic and this closure takes two parameters, so a program that
// widens the value afterwards to a view with more of them loses what Node would forward
// (timers_source/widened.a), which is open on #995spev.
const (
	nodeSetTimeoutView   = "(callback: () => void, ms: number) => unknown"
	nodeClearTimeoutView = "(timeoutId: unknown) => void"
)

// nodeTimerValueName is the timer a node reads as a value, by its library declaration: an identifier
// that isn't a call's callee, a property's name, a typeof's operand, a qualified namespace member's
// receiver or part of a type, or a shorthand property's name whose value is the global
// ({ setTimeout }). It's empty for anything else; a direct call is nodeTimer's.
func (l *lowering) nodeTimerValueName(node *ast.Node) string {
	var symbol *ast.Symbol
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment && node.Parent.Name() == node {
		symbol = l.checker.GetShorthandAssignmentValueSymbol(node.Parent)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = l.checker.GetAliasedSymbol(symbol)
		}
	} else {
		if node.Kind != ast.KindIdentifier || !ast.IsExpressionNode(node) || ast.IsPartOfTypeNode(node) || called(node) {
			return ""
		}
		if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node {
			return ""
		}
		use := node
		for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
			use = use.Parent
		}
		if use.Parent == nil || use.Parent.Kind == ast.KindTypeOfExpression || use.Parent.Kind == ast.KindPropertyAccessExpression && use.Parent.Expression() == use && l.namespaceMember(use.Parent) {
			return ""
		}
		symbol = l.symbol(node)
	}
	switch name := nodeLibraryName(symbol); name {
	case nodeSetTimeoutName, nodeClearTimeoutName:
		return name
	}
	return ""
}

// nodeTimerValueRefusal judges a timer read as a value: compared by identity, which its one closure
// keeps, or written where the slot sees it exactly as its view, an optional method's undefined aside.
// Any other slot is NotYet, and so is none, naming what the read is seen as. clearTimeout's value is
// NotYet where its direct call would be for any object (nodeClearTimeout): the program declares an
// _onTimeout method or accessor, which Node reads and the run-time check doesn't.
func (l *lowering) nodeTimerValueRefusal(node *ast.Node, name string) error {
	view := nodeSetTimeoutView
	if name == nodeClearTimeoutName {
		view = nodeClearTimeoutView
	}
	use := node
	for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
		use = use.Parent
	}
	if use.Parent != nil && use.Parent.Kind == ast.KindBinaryExpression {
		operator := use.Parent.AsBinaryExpression().OperatorToken.Kind
		if operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken {
			return nil
		}
	}
	expected := l.slotType(node)
	if expected == nil {
		return l.notYet(node, name+" value needs exactly "+view+", not "+l.checker.TypeToString(l.checker.GetTypeAtLocation(node))+" with no slot")
	}
	if !l.nodeTimerView(expected, name) {
		return l.notYet(node, name+" value needs exactly "+view+", not "+l.checker.TypeToString(expected))
	}
	if name == nodeClearTimeoutName && l.declaresTimeoutMember() {
		return l.notYet(node, nodeClearTimeoutName+" value where the program declares an _onTimeout method or accessor, which Node reads")
	}
	return nil
}

// nodeTimerView says whether a slot sees a timer exactly as its view: setTimeout's callback a
// function taking nothing and returning void, its delay a number, both required, and its result
// unknown; clearTimeout's one required parameter unknown and its result void.
func (l *lowering) nodeTimerView(slot *checker.Type, name string) bool {
	signature, plain := l.nodeTimerSignature(l.concrete(l.checker.GetNonNullableType(slot)))
	if !plain {
		return false
	}
	parameters := signature.Parameters()
	returned := l.concrete(l.checker.GetReturnTypeOfSignature(signature)).Flags()
	if name == nodeClearTimeoutName {
		return len(parameters) == 1 && signature.MinArgumentCount() == 1 && l.concrete(l.checker.GetTypeOfSymbol(parameters[0])).Flags() == checker.TypeFlagsUnknown && returned == checker.TypeFlagsVoid
	}
	if len(parameters) != 2 || signature.MinArgumentCount() != 2 || l.concrete(l.checker.GetTypeOfSymbol(parameters[1])).Flags() != checker.TypeFlagsNumber || returned != checker.TypeFlagsUnknown {
		return false
	}
	callback, plain := l.nodeTimerSignature(l.concrete(l.checker.GetTypeOfSymbol(parameters[0])))
	return plain && len(callback.Parameters()) == 0 && l.concrete(l.checker.GetReturnTypeOfSignature(callback)).Flags() == checker.TypeFlagsVoid
}

// nodeTimerSignature is a function type's one call signature, when it has exactly one, not generic,
// with no this and no rest parameter, and the type has no construct signature or property beside it.
func (l *lowering) nodeTimerSignature(proven *checker.Type) (*checker.Signature, bool) {
	if proven.Flags()&checker.TypeFlagsObject == 0 || len(l.checker.GetPropertiesOfType(proven)) != 0 || len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 {
		return nil, false
	}
	signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, false
	}
	signature := signatures[0]
	if len(signature.TypeParameters()) != 0 || signature.ThisParameter() != nil || signature.Flags()&checker.SignatureFlagsHasRestParameter != 0 {
		return nil, false
	}
	return signature, true
}

// nodeTimerUseRefusal is the refusal visitor's view of a timer read as a value, judged as lowering
// judges it, before the callable namespace's stop (namespaces.go) can name it more vaguely.
func (l *lowering) nodeTimerUseRefusal(node *ast.Node) (bool, error) {
	name := l.nodeTimerValueName(node)
	if name == "" {
		return false, nil
	}
	return true, l.nodeTimerValueRefusal(node, name)
}

// nodeTimerValue lowers a timer read as a value to its one closure, wherever nodeTimerValueRefusal
// admits the read.
func (l *lowering) nodeTimerValue(node *ast.Node) (ir.Expression, bool, error) {
	name := l.nodeTimerValueName(node)
	if name == "" {
		return nil, false, nil
	}
	if err := l.nodeTimerValueRefusal(node, name); err != nil {
		return nil, true, err
	}
	return l.nodeTimerClosure(name), true, nil
}

// nodeTimerClosure is a timer's one closure, made the first time either is read and held by a global
// made before anything runs, as Realpath's value is (nodeRealpathValue), so setTimeout === setTimeout
// holds as on Node. Its body is the timer's direct call on its own parameters, as nodeTimer lowers
// one: setTimeout's Timeout goes back as the view's unknown, a union, and clearTimeout's call is the
// body's one statement. No checker type names the callback's view here, since one closure serves
// every slot that holds it, so the thunk's CallbackType is left unresolved, and the thunk packs its
// arguments for every function value the program makes (ClosureArgumentLayout), as for a callback
// bounded by none. Either guard for the open widening goes here.
func (l *lowering) nodeTimerClosure(name string) ir.Expression {
	role := "host_set_timeout"
	if name == nodeClearTimeoutName {
		role = "host_clear_timeout"
	}
	for index, function := range l.result.Functions {
		if function.Source.Role == role {
			return ir.Read{Local: l.forwarders[index], Of: ir.Closure}
		}
	}
	index := len(l.result.Functions)
	function := ir.Function{Name: role + "_value", Closure: true}
	function.Source.Role = role
	parameter := func(name string, of ir.Type) ir.Expression {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: index})
		function.Parameters = append(function.Parameters, local)
		return ir.Read{Local: local, Of: of}
	}
	// Node's names for the parameters, and its lengths: setTimeout's is 2 and clearTimeout's 1.
	if name == nodeClearTimeoutName {
		timer := parameter("timer", ir.Union)
		function.SourceLength = 1
		function.Body = []ir.Statement{ir.Evaluate{Value: ir.NodeFSFile{Operation: "clear_timeout", Arguments: []ir.Expression{timer}, Of: ir.Number}}}
	} else {
		callback := parameter("callback", ir.Closure)
		after := parameter("after", ir.Number)
		function.SourceLength = 2
		function.Returns = ir.Union
		function.Body = []ir.Statement{ir.Return{Value: fit(ir.NodeFSFile{Operation: "set_timeout", Arguments: []ir.Expression{callback, after}, Of: ir.Object}, ir.Union)}}
	}
	l.result.Functions = append(l.result.Functions, function)
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: role + "_value", Type: ir.Closure, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: held, Value: ir.MakeClosure{Function: index}})
	if l.forwarders == nil {
		l.forwarders = map[int]int{}
	}
	l.forwarders[index] = held
	return ir.Read{Local: held, Of: ir.Closure}
}

func init() { RegisterNodeLibraryMembers(nodeSetTimeoutName, nodeClearTimeoutName) }
