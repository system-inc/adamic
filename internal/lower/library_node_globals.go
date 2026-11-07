package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Error.captureStackTrace is a function on the global Error that sets its target's .stack in Node
// and does nothing else. Adamic refuses reading .stack (nodeStackTraceLimitMember), so it's bound as
// a function whose body is empty: a call evaluates its arguments and nothing more, and a read gives
// one function value made before anything runs, so it's truthy and === itself, as Node's is. As with
// stackTraceLimit the member isn't registered: a form nodeGlobalsUseRefusal misses falls to the
// generic NotYet.
const nodeCaptureStackTraceMember = "node:globals.ErrorConstructor.captureStackTrace"

// global.gc is Node's collector, there only when Node runs with --expose-gc. Neither the oracle nor
// a native or wasm binary does, so a read of it is undefined wherever it's read, and a call throws
// the TypeError Node throws there. global is read only as that access's receiver.
const (
	nodeGarbageCollectorMember = "node:globals.gc"
	nodeGlobalMember           = "node:globals.global"
)

// nodeGarbageCollectorMissing is what Node's TypeError says when global.gc() runs without --expose-gc.
const nodeGarbageCollectorMissing = "global.gc is not a function"

// nodeCaptureStackTraceRole is the Source.Role of the one function Error.captureStackTrace is bound to.
const nodeCaptureStackTraceRole = "host_capture_stack_trace"

// nodeGlobalsOuter is node with the parentheses around it.
func nodeGlobalsOuter(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node
}

// nodeGlobalsCall is the call node is the callee of, parentheses and a non-null assertion aside, or
// nil when it's none's.
func nodeGlobalsCall(node *ast.Node) *ast.Node {
	outer := node
	for outer.Parent != nil &&
		(outer.Parent.Kind == ast.KindParenthesizedExpression || outer.Parent.Kind == ast.KindNonNullExpression) {
		outer = outer.Parent
	}
	if outer.Parent == nil ||
		outer.Parent.Kind != ast.KindCallExpression ||
		outer.Parent.AsCallExpression().Expression != outer {
		return nil
	}
	return outer.Parent
}

// nodeGlobalsCallee is the member access a call calls, parentheses and a non-null assertion aside.
func nodeGlobalsCallee(call *ast.Node) *ast.Node {
	callee := call.AsCallExpression().Expression
	for callee.Kind == ast.KindParenthesizedExpression || callee.Kind == ast.KindNonNullExpression {
		callee = callee.Expression()
	}
	return callee
}

// nodeGlobalsWritten names a member access written or deleted, or is "" for a read.
func nodeGlobalsWritten(node *ast.Node) string {
	outer := nodeGlobalsOuter(node)
	if outer.Parent != nil && outer.Parent.Kind == ast.KindDeleteExpression {
		return "deleted"
	}
	if ast.IsAssignmentTarget(outer) {
		return "written"
	}
	return ""
}

// nodeCaptureStackTraceRefusal names every use of Error.captureStackTrace but a read and a call as a
// statement, both through the global Error identifier.
func (lowering *lowering) nodeCaptureStackTraceRefusal(node *ast.Node) error {
	fix := " (fix: read or call Error.captureStackTrace)"
	if node.Kind != ast.KindPropertyAccessExpression {
		return lowering.notYet(node, nodeCaptureStackTraceMember+" outside a property access"+fix)
	}
	if !lowering.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Error") {
		return lowering.notYet(node, nodeCaptureStackTraceMember+" through a receiver other than the global Error"+fix)
	}
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return lowering.notYet(node, nodeCaptureStackTraceMember+" read through an optional chain"+fix)
	}
	if use := nodeGlobalsWritten(node); use != "" {
		return lowering.notYet(node, nodeCaptureStackTraceMember+" "+use+fix)
	}
	call := nodeGlobalsCall(node)
	if call == nil {
		return nil
	}
	if call.Flags&ast.NodeFlagsOptionalChain != 0 {
		return lowering.notYet(call, nodeCaptureStackTraceMember+" called through an optional chain"+fix)
	}
	for _, argument := range call.AsCallExpression().Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return lowering.notYet(
				argument,
				nodeCaptureStackTraceMember+" called with a spread argument (fix: pass its arguments one by one)",
			)
		}
	}
	if outer := nodeGlobalsOuter(call); outer.Parent == nil || outer.Parent.Kind != ast.KindExpressionStatement {
		return lowering.notYet(
			call,
			nodeCaptureStackTraceMember+" with its undefined result used as a value (fix: call it as a statement)",
		)
	}
	return nil
}

// nodeGarbageCollectorRefusal names every use of global.gc but a read and a call, both through the
// global identifier.
func (lowering *lowering) nodeGarbageCollectorRefusal(node *ast.Node) error {
	fix := " (fix: read or call global.gc)"
	if node.Kind != ast.KindPropertyAccessExpression {
		return lowering.notYet(node, nodeGarbageCollectorMember+" outside a property access of global"+fix)
	}
	receiver := node.AsPropertyAccessExpression().Expression
	if !ast.IsIdentifier(receiver) || lowering.nodeLibraryMember(receiver) != nodeGlobalMember {
		return lowering.notYet(
			node,
			nodeGarbageCollectorMember+" through a receiver other than the global identifier"+fix,
		)
	}
	if node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return lowering.notYet(node, nodeGarbageCollectorMember+" read through an optional chain"+fix)
	}
	if use := nodeGlobalsWritten(node); use != "" {
		return lowering.notYet(node, nodeGarbageCollectorMember+" "+use+fix)
	}
	if call := nodeGlobalsCall(node); call != nil && call.Flags&ast.NodeFlagsOptionalChain != 0 {
		return lowering.notYet(call, nodeGarbageCollectorMember+" called through an optional chain"+fix)
	}
	if call := nodeGlobalsCall(node); call != nil {
		for _, argument := range call.AsCallExpression().Arguments.Nodes {
			if argument.Kind == ast.KindSpreadElement {
				return lowering.notYet(
					argument,
					nodeGarbageCollectorMember+" called with a spread argument (fix: pass its arguments one by one)",
				)
			}
		}
		// The call always throws, so its result is never had; a slot that would hold it isn't lowered.
		if outer := nodeGlobalsOuter(call); outer.Parent == nil || outer.Parent.Kind != ast.KindExpressionStatement {
			return lowering.notYet(call, nodeGarbageCollectorMember+
				" with its result used as a value, which the call's TypeError never gives (fix: call it as a statement)")
		}
	}
	return nil
}

// nodeGlobalsUseRefusal is the refusal visitor's view of Error.captureStackTrace, global.gc and
// global: a property's name and a call's callee are judged with their access, global is judged as
// global.gc's receiver, and global anywhere else is NotYet.
func (lowering *lowering) nodeGlobalsUseRefusal(node *ast.Node) (bool, error) {
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node {
		switch lowering.nodeLibraryMember(node.Parent) {
		case nodeCaptureStackTraceMember:
			return true, lowering.nodeCaptureStackTraceRefusal(node.Parent)
		case nodeGarbageCollectorMember:
			return true, lowering.nodeGarbageCollectorRefusal(node.Parent)
		}
	}
	member := node
	switch node.Kind {
	case ast.KindCallExpression:
		member = nodeGlobalsCallee(node)
	case ast.KindNewExpression:
		member = ast.SkipParentheses(node.AsNewExpression().Expression)
	case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
		// The checker names what's inside, which is judged as itself.
		for member.Kind == ast.KindParenthesizedExpression || member.Kind == ast.KindNonNullExpression {
			member = member.Expression()
		}
	}
	switch lowering.nodeLibraryMember(member) {
	case nodeCaptureStackTraceMember:
		return true, lowering.nodeCaptureStackTraceRefusal(member)
	case nodeGarbageCollectorMember:
		return true, lowering.nodeGarbageCollectorRefusal(member)
	case nodeGlobalMember:
		if node.Kind != ast.KindIdentifier {
			// A call or new of global itself; the checker refuses both, and global is judged on its own visit.
			return true, lowering.notYet(
				node,
				nodeGlobalMember+" other than as global.gc's receiver (fix: name the global itself, without global.)",
			)
		}
		if node.Parent != nil &&
			node.Parent.Kind == ast.KindPropertyAccessExpression &&
			node.Parent.AsPropertyAccessExpression().Expression == node &&
			lowering.nodeLibraryMember(node.Parent) == nodeGarbageCollectorMember {
			// The access is judged on its own visit.
			return true, nil
		}
		return true, lowering.notYet(
			node,
			nodeGlobalMember+" other than as global.gc's receiver (fix: name the global itself, without global.)",
		)
	}
	return false, nil
}

// nodeCaptureStackTraceRead says whether node is a read of Error.captureStackTrace that
// nodeCaptureStackTraceRefusal admits. Its function never reads this, so unlike a method read as a
// value it loses nothing.
func (lowering *lowering) nodeCaptureStackTraceRead(node *ast.Node) bool {
	return node.Kind == ast.KindPropertyAccessExpression &&
		lowering.nodeLibraryMember(node) == nodeCaptureStackTraceMember &&
		nodeGlobalsCall(node) == nil &&
		lowering.nodeCaptureStackTraceRefusal(node) == nil
}

// nodeCaptureStackTraceValue is the one function value Error.captureStackTrace reads as: a closure of
// a function with Node's two parameters and an empty body, held by a global made before anything
// runs, as a module function read as a value is (functionValue). A slot the value is written into
// must see its parameters as they're declared and its result as void.
func (lowering *lowering) nodeCaptureStackTraceValue(node *ast.Node) (ir.Expression, error) {
	signatures := lowering.checker.GetSignaturesOfType(
		lowering.checker.GetTypeAtLocation(node),
		checker.SignatureKindCall,
	)
	if len(signatures) != 1 {
		return nil, lowering.notYet(node, nodeCaptureStackTraceMember+" as a value with other than one signature")
	}
	parameters := []ir.Type{}
	for _, parameter := range signatures[0].Parameters() {
		of, known := lowering.censusCallableParameter(parameter)
		if !known || censusCallableSlotless(of) {
			return nil, lowering.notYet(node, nodeCaptureStackTraceMember+
				" as a value taking a parameter "+
				parameter.Name+
				" with no slot representation")
		}
		parameters = append(parameters, of)
	}
	if expected := lowering.slotType(node); expected != nil {
		viewed := lowering.checker.GetSignaturesOfType(lowering.concrete(expected), checker.SignatureKindCall)
		valid := len(viewed) == 1 &&
			viewed[0].ThisParameter() == nil &&
			viewed[0].Flags()&checker.SignatureFlagsHasRestParameter == 0 &&
			len(viewed[0].TypeParameters()) == 0 &&
			len(viewed[0].Parameters()) <= len(parameters)
		if valid {
			for index, parameter := range viewed[0].Parameters() {
				of, known := lowering.censusCallableParameter(parameter)
				valid = valid && known && of == parameters[index]
			}
			valid = valid &&
				lowering.concrete(
					lowering.checker.GetReturnTypeOfSignature(viewed[0]),
				).Flags()&checker.TypeFlagsVoid != 0
		}
		if !valid {
			return nil, lowering.notYet(node, nodeCaptureStackTraceMember+
				" as a value in a slot that sees it as "+
				lowering.checker.TypeToString(expected)+
				" (fix: let the slot see (targetObject: object, constructorOpt?: Function) => void)")
		}
	}
	for index, function := range lowering.result.Functions {
		if function.Source.Role == nodeCaptureStackTraceRole {
			return ir.Read{Local: lowering.forwarders[index], Of: ir.Closure}, nil
		}
	}
	index := len(lowering.result.Functions)
	// Node's Error.captureStackTrace.length is 2.
	function := ir.Function{Name: nodeCaptureStackTraceRole + "_value", SourceLength: 2, Closure: true}
	function.Source.Role = nodeCaptureStackTraceRole
	for position, of := range parameters {
		local := len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: signatures[0].Parameters()[position].Name, Type: of, Function: index},
		)
		function.Parameters = append(function.Parameters, local)
	}
	lowering.result.Functions = append(lowering.result.Functions, function)
	lowering.closureRecords = append(
		lowering.closureRecords,
		closureRecord{proven: lowering.concrete(lowering.checker.GetTypeAtLocation(node)), function: index, node: node},
	)
	held := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: nodeCaptureStackTraceRole + "_value", Type: ir.Closure, Global: true, Function: -1},
	)
	lowering.forwarderValues = append(
		lowering.forwarderValues,
		ir.Declare{Local: held, Value: ir.MakeClosure{Function: index}},
	)
	if lowering.forwarders == nil {
		lowering.forwarders = map[int]int{}
	}
	lowering.forwarders[index] = held
	return ir.Read{Local: held, Of: ir.Closure}, nil
}

// nodeGlobalsValue lowers a read of Error.captureStackTrace to its one function value and a call of
// it to its arguments' evaluation in order, a read of global.gc to undefined, and a call of global.gc
// to its arguments' evaluation and then Node's TypeError, thrown as Node throws it, so a catch around
// it sees it.
func (lowering *lowering) nodeGlobalsValue(node *ast.Node) (ir.Expression, bool, error) {
	member := node
	switch node.Kind {
	case ast.KindCallExpression:
		member = nodeGlobalsCallee(node)
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindIdentifier:
	default:
		return nil, false, nil
	}
	name := lowering.nodeLibraryMember(member)
	if name != nodeCaptureStackTraceMember && name != nodeGarbageCollectorMember {
		return nil, false, nil
	}
	refusal := lowering.nodeCaptureStackTraceRefusal
	if name == nodeGarbageCollectorMember {
		refusal = lowering.nodeGarbageCollectorRefusal
	}
	if refusalError := refusal(member); refusalError != nil {
		return nil, true, refusalError
	}
	if node.Kind != ast.KindCallExpression {
		if nodeGlobalsCall(node) != nil {
			// The callee is lowered with its call, never apart from it.
			return nil, true, lowering.notYet(node, name+" as a callee lowered apart from its call")
		}
		if name == nodeCaptureStackTraceMember {
			value, nodeCaptureStackTraceValueError := lowering.nodeCaptureStackTraceValue(node)
			return value, true, nodeCaptureStackTraceValueError
		}
		of, known := lowering.representation(lowering.checker.GetTypeAtLocation(node))
		if !known || !of.IsReference() {
			return nil, true, lowering.notYet(
				node,
				nodeGarbageCollectorMember+" read where its type has no reference representation",
			)
		}
		return ir.Undefined{Of: of}, true, nil
	}
	effects := ir.Effects{Result: ir.Undefined{}}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		value, expressionError := lowering.expression(argument)
		if expressionError != nil {
			return nil, true, expressionError
		}
		effects.Body = append(effects.Body, ir.Evaluate{Value: value})
	}
	if name == nodeGarbageCollectorMember {
		index := len(lowering.result.Functions)
		failure := ir.MakeError{
			Message: ir.StringConstant{Index: lowering.constant(nodeGarbageCollectorMissing)},
			Name:    ir.StringConstant{Index: lowering.constant("TypeError")},
		}
		lowering.result.Functions = append(
			lowering.result.Functions,
			ir.Function{Name: "global_gc", Body: []ir.Statement{ir.Throw{Value: failure}}},
		)
		effects.Body = append(effects.Body, ir.Evaluate{Value: ir.Call{Function: index}})
	}
	return effects, true, nil
}
