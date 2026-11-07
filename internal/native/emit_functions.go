// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (nativeEmitter *emitter) functionName(function int) string {
	if nativeEmitter.functionNames == nil {
		nativeEmitter.functionNames = map[int]string{}
	}
	if name, ok := nativeEmitter.functionNames[function]; ok {
		return name
	}
	name := nativeEmitter.namedFunction(function)
	nativeEmitter.functionNames[function] = name
	return name
}

func (nativeEmitter *emitter) signature(function int) string {
	declared := nativeEmitter.program.Functions[function]
	if declared.Closure {
		// Every closure's code is called the same way (adamic_code): its arguments and result as
		// adamic_value, whatever their types.
		count := ""
		if nativeEmitter.program.PackedCountNeeded(function) {
			count = ", size_t argument_count"
		}
		return fmt.Sprintf(
			"adamic_value %s(adamic_closure *self, adamic_value *arguments%s)",
			nativeEmitter.functionName(function),
			count,
		)
	}
	returns := "void"
	if declared.Returns != 0 {
		returns = cType(declared.Returns)
	}
	parameters := []string{}
	name := nativeEmitter.functionName(function)
	if nativeEmitter.inRegion {
		// The version that makes its result in the region it's handed, never NULL (region.go).
		parameters = append(parameters, "adamic_region *region")
		name = nativeEmitter.regionFunctionName(function)
	}
	for _, parameter := range declared.Parameters {
		parameters = append(
			parameters,
			cType(nativeEmitter.program.Locals[parameter].Type)+" "+nativeEmitter.localName(parameter),
		)
	}
	if declared.ArgumentsCount != 0 {
		parameters = append(parameters, "double "+nativeEmitter.localName(declared.ArgumentsCount-1))
	}
	if len(parameters) == 0 {
		parameters = append(parameters, "void")
	}
	return fmt.Sprintf("%s %s(%s)", returns, name, strings.Join(parameters, ", "))
}

// stackCheck is a function's stack guard, naming the function in its overflow panic. The macro
// joins a short name to its message as literals; a message past a C literal's limit is written
// whole, as bytes (cString), since the preprocessor cannot join those.
func stackCheck(name string) string {
	if message := "RangeError: Maximum call stack size exceeded in " + name; len(message) > longestLiteral {
		return "ADAMIC_CHECK_STACK_MESSAGE(" + cString(message) + ")"
	}
	return "ADAMIC_CHECK_STACK_NAMED(" + cString(name) + ")"
}

// functionBody emits a function's body. A string parameter is retained on entry and released on
// every way out, like any local, so a function may reassign it without touching its caller's.
func (nativeEmitter *emitter) functionBody(function ir.Function) {
	nativeEmitter.function = &function
	nativeEmitter.functionDepth = len(nativeEmitter.scopes)
	nativeEmitter.scopes = append(nativeEmitter.scopes, nil)
	// Recursion that runs out of stack panics, as Node's does, rather than crashing (stack.c). The
	// panic names the source declaration, not the instance, so instances sharing a stable symbol
	// emit one body (emit.go).
	name := function.Name
	if len(function.Source.Declaration) != 0 {
		name = function.Source.Declaration[len(function.Source.Declaration)-1]
	}
	if name == "" {
		name = "<anonymous>"
	}
	nativeEmitter.line("%s;", stackCheck(name))
	if function.ArgumentsCount != 0 && !function.Closure {
		nativeEmitter.line("(void)%s;", nativeEmitter.localName(function.ArgumentsCount-1))
	}
	if function.Closure {
		nativeEmitter.line("(void)self;")
		if nativeEmitter.program.PackedCountNeeded(nativeEmitter.functionIndex) {
			nativeEmitter.line("(void)argument_count;")
		}
		nativeEmitter.line("(void)arguments;")
		if function.ArgumentsCount != 0 {
			count := "(double)argument_count"
			if function.Receiver {
				count += " - 1"
			}
			nativeEmitter.line("double %s = %s;", nativeEmitter.localName(function.ArgumentsCount-1), count)
			nativeEmitter.line("(void)%s;", nativeEmitter.localName(function.ArgumentsCount-1))
		}
		for index, parameter := range function.Parameters {
			local := nativeEmitter.program.Locals[parameter]
			if function.RestElement != 0 && index == len(function.Parameters)-1 {
				slot := nativeEmitter.program.RestArgumentSlots[ir.FunctionRestArguments(function)]
				if function.Receiver {
					slot++
				}
				nativeEmitter.line(
					"adamic_array *%s = arguments[%d].reference;",
					nativeEmitter.localName(parameter),
					slot,
				)
				continue
			}
			value := unslotted(local.Type, fmt.Sprintf("arguments[%d].%s", index, member(local.Type)))
			if nativeEmitter.program.PackedCountNeeded(nativeEmitter.functionIndex) {
				value = closureArgument(local.Type, index)
			}
			if local.Type.IsReference() {
				value = fmt.Sprintf("(%s)%s", cType(local.Type), value)
			}
			if nativeEmitter.program.PackedCountNeeded(nativeEmitter.functionIndex) &&
				(local.Type.IsMaybe() || local.Type.IsReference()) {
				value = fmt.Sprintf("(argument_count > %d ? %s : %s)", index, value, absent(local.Type))
			}
			nativeEmitter.line("%s %s = %s;", cType(local.Type), nativeEmitter.localName(parameter), value)
		}
	}
	for index := range function.Body {
		if _, ok := function.Body[index].(ir.AllocateEnvironment); ok {
			nativeEmitter.allocateEnvironment(&function.Body[index])
		}
	}
	for _, parameter := range function.Parameters {
		local := nativeEmitter.program.Locals[parameter]
		if local.Captured {
			if local.EnvironmentCell && local.Type.IsReference() && local.Borrowed {
				panic(fmt.Sprintf("native: a store into the borrowed parameter %s", local.Name))
			}
			// The cell owns the incoming value. No independent raw parameter survives entry.
			nativeEmitter.makeCell(
				parameter,
				nativeEmitter.localName(parameter),
				nativeEmitter.reuse.consumed[parameter],
			)
			continue
		}
		switch {
		case nativeEmitter.reuse.consumed[parameter]:
			// Its caller handed over a reference (reuse.go): it's the callee's to let go of.
			nativeEmitter.hold(nativeEmitter.localName(parameter))
		case nativeEmitter.program.Locals[parameter].Type.IsReference() &&
			!nativeEmitter.program.Locals[parameter].Borrowed:
			// A borrowed parameter is its caller's, kept alive for the whole call.
			nativeEmitter.line("adamic_retain(%s);", nativeEmitter.localName(parameter))
			nativeEmitter.hold(nativeEmitter.localName(parameter))
		}
	}
	for index := range function.Body {
		nativeEmitter.statementAt(&function.Body[index])
	}
	if function.Returns == 0 {
		nativeEmitter.releaseScopes(nativeEmitter.functionDepth)
		if function.Closure {
			nativeEmitter.line("return (adamic_value){.number = 0};")
		}
	} else {
		// Lowering records every permitted implicit return in the IR. C cannot always see
		// that all paths return; reaching this guard still means an IR exit was lost.
		nativeEmitter.line("adamic_unreachable();")
	}
	if nativeEmitter.debug {
		nativeEmitter.line(sourceLineReset)
	}
	nativeEmitter.scopes = nativeEmitter.scopes[:nativeEmitter.functionDepth]
	nativeEmitter.function = nil
}

func (nativeEmitter *emitter) returnStatement(statement ir.Return) {
	if statement.Value == nil {
		nativeEmitter.end()
		nativeEmitter.finallies(0)
		nativeEmitter.releaseScopes(nativeEmitter.functionDepth)
		if nativeEmitter.function != nil && nativeEmitter.function.Closure {
			nativeEmitter.line("return (adamic_value){.number = 0};")
		} else {
			nativeEmitter.line("return;")
		}
		return
	}
	value := ""
	if nativeEmitter.function != nil && nativeEmitter.inRegion {
		// A function that returns fresh makes what it returns in the region it was handed (region.go).
		if literal, isLiteral := statement.Value.(ir.ObjectLiteral); isLiteral && literal.Spread == nil {
			nativeEmitter.regionLiteralDepth = nativeEmitter.depth + 1
		}
		value = nativeEmitter.handRegion(statement.Value, "region")
		if read, ok := statement.Value.(ir.Read); ok {
			if local, known := nativeEmitter.regions.classObjects[nativeEmitter.functionIndex]; known &&
				local == read.Local {
				nativeEmitter.regionValues[value] = true
			}
		}
		nativeEmitter.regionLiteralDepth = 0
	} else {
		value = nativeEmitter.value(statement.Value)
	}
	result := nativeEmitter.temporary()
	if statement.Value.Type().IsReference() && !nativeEmitter.regionValues[value] {
		held := statement.Value.Type()
		if _, isUndefined := statement.Value.(ir.Undefined); isUndefined &&
			nativeEmitter.function != nil &&
			nativeEmitter.function.Returns.IsReference() {
			// undefined is a null pointer of whatever reference the function returns, a string's
			// among them, not the object pointer it's typed as alone.
			held = nativeEmitter.function.Returns
		}
		nativeEmitter.line("%s %s = %s;", cType(held), result, nativeEmitter.kept(value))
	} else {
		nativeEmitter.line("%s %s = %s;", cType(statement.Value.Type()), result, value)
	}
	nativeEmitter.end()
	if len(nativeEmitter.handlers) > 0 {
		// Every finally open runs before the function returns. The result waits in a scope of its
		// own, so a throw from a finally, which replaces the return, lets go of it.
		held := []string{}
		if statement.Value.Type().IsReference() {
			held = append(held, result)
		}
		nativeEmitter.scopes = append(nativeEmitter.scopes, held)
		nativeEmitter.finallies(0)
		nativeEmitter.scopes = nativeEmitter.scopes[:len(nativeEmitter.scopes)-1]
	}
	nativeEmitter.releaseScopes(nativeEmitter.functionDepth)
	if nativeEmitter.function != nil && nativeEmitter.function.Closure {
		nativeEmitter.line(
			"return (adamic_value){.%s = %s};",
			member(statement.Value.Type()),
			slotted(statement.Value.Type(), result),
		)
	} else {
		nativeEmitter.line("return %s;", result)
	}
}

// arguments evaluates a call's arguments in order, each fitted to its parameter: a number or undefined
// made number | undefined's two words. An optional parameter the call leaves out gets undefined. This
// is done here, not in lowering, since only here is every function's signature known: lowering may
// meet a call before the function it calls.
func (nativeEmitter *emitter) arguments(call ir.Call) []string {
	function := nativeEmitter.program.Functions[call.Function]
	if len(call.Spread) != 0 {
		return nativeEmitter.spreadArguments(call)
	}
	if function.RestElement != 0 && !call.RestPacked {
		actual := len(call.Arguments)
		if function.Receiver {
			actual--
		}
		fixed := len(function.Parameters) - 1
		tail := []ir.Expression{}
		if len(call.Arguments) > fixed {
			tail = call.Arguments[fixed:]
		}
		rest := ir.ArrayLiteral{Element: function.RestElement, Elements: tail}
		copy := call
		copy.Arguments = append([]ir.Expression{}, call.Arguments[:min(fixed, len(call.Arguments))]...)
		for len(copy.Arguments) < fixed {
			copy.Arguments = append(copy.Arguments, ir.Undefined{})
		}
		copy.Arguments = append(copy.Arguments, rest)
		copy.RestPacked = true
		copy.ArgumentCount = ir.NumberConstant{Value: float64(actual)}
		return nativeEmitter.arguments(copy)
	}
	parameters := nativeEmitter.program.Functions[call.Function].Parameters
	arguments := make([]string, 0, len(parameters))
	handed := []string{}
	defer func() { nativeEmitter.handedOver(handed) }()
	for index, argument := range call.Arguments {
		if index >= len(parameters) {
			nativeEmitter.value(argument) // Extra arguments still run, before the call.
			continue
		}
		if nativeEmitter.reuse.callConsumes(nativeEmitter.program, call, index) {
			value := nativeEmitter.handOver(argument)
			handed = append(handed, value)
			arguments = append(arguments, value)
			continue
		}
		value := ""
		if lent, ok := nativeEmitter.lentArgument(call, index); ok {
			value = lent
		} else if nativeEmitter.statementRegion != "" && !nativeEmitter.regions.callEscapes(call, index) {
			// A parameter that flows nowhere: a fresh value handed to it lives in the statement's region.
			value = nativeEmitter.handRegion(argument, "&"+nativeEmitter.statementRegion)
		} else {
			value = nativeEmitter.value(argument)
		}
		if index < len(parameters) && nativeEmitter.program.Locals[parameters[index]].Type.IsMaybe() {
			of := nativeEmitter.program.Locals[parameters[index]].Type
			if _, isUndefined := argument.(ir.Undefined); isUndefined {
				value = zero(of)
			} else if argument.Type() == of.Present() {
				value = maybe(of, value)
			}
		}
		if index < len(parameters) &&
			(nativeEmitter.program.Locals[parameters[index]].Type == ir.Union ||
				nativeEmitter.program.Locals[parameters[index]].Type == ir.Weak) &&
			argument.Type() != nativeEmitter.program.Locals[parameters[index]].Type {
			// A number, a boolean or a reference where a union goes, boxed here, where the signature
			// is known; a reference where a Weak goes, its handle.
			takes := nativeEmitter.program.Locals[parameters[index]].Type
			boxed, fresh := converted(argument.Type(), takes, value)
			if _, null := argument.(ir.Null); null && takes == ir.Union {
				boxed, fresh = "&adamic_null", false
			}
			if fresh {
				boxed = nativeEmitter.own(takes, boxed)
			}
			value = boxed
		}
		if index < len(parameters) {
			arguments = append(arguments, value)
		} else {
			nativeEmitter.line("(void)%s;", value)
		}
	}
	for _, parameter := range parameters[min(len(call.Arguments), len(parameters)):] {
		arguments = append(arguments, absent(nativeEmitter.program.Locals[parameter].Type))
	}
	if function.ArgumentsCount != 0 {
		count := fmt.Sprint(len(call.Arguments))
		if function.Receiver {
			count = fmt.Sprint(len(call.Arguments) - 1)
		}
		if call.ArgumentCount != nil {
			count = nativeEmitter.value(call.ArgumentCount)
		}
		if call.ForwardCount {
			count = nativeEmitter.localName(nativeEmitter.function.ArgumentsCount - 1)
		}
		arguments = append(arguments, count)
	}
	return arguments
}

// callThrough calls a function value: closure, or for a call through an interface (ir.Property's
// Method), the receiver's own function value or its class's method, found before the arguments are
// evaluated, as JavaScript reads object.name first.
func (nativeEmitter *emitter) callThrough(expression ir.CallClosure, closure string, receiver string) string {
	method := ""
	exactCount := false
	// What a void call releases of its result: closureArm where it calls closure (or the proven
	// method), methodArm where the object's class's method ran instead.
	closureArm, methodArm := ir.DiscardNothing, ir.DiscardNothing
	if expression.Returns == 0 {
		closureArm = nativeEmitter.program.ClosureDiscard(expression)
	}
	if receiver != "" {
		property := expression.Closure.(ir.Property)
		if function, known := nativeEmitter.exactReceiverMethod(property.Object, property.Name); known {
			// Keep the interface adapter's borrowed-input convention and the same
			// exception and result handling, but call its proven method directly.
			method = nativeEmitter.methodThunk(function)
			exactCount = nativeEmitter.program.PackedCountNeeded(function)
			if expression.Returns == 0 {
				closureArm = nativeEmitter.program.ReturnsDiscard([]int{function})
			}
		} else {
			if expression.Returns == 0 {
				methodArm = nativeEmitter.program.MethodDiscard(expression)
			}
			method = nativeEmitter.temporary()
			if nativeEmitter.program.ClosureConventionNeeded() {
				nativeEmitter.line("adamic_method_entry %s = {0};", method)
			} else {
				nativeEmitter.line("adamic_method %s = NULL;", method)
			}
			if property.View != "" {
				property.Method = false
				var checked bool
				closure, checked = readUnionViewAt(nativeEmitter, property, receiver)
				if !checked {
					panic("compiler bug: unavailable callable view read")
				}
			} else {
				closure = nativeEmitter.own(ir.Closure, fmt.Sprintf(
					"adamic_retain(adamic_object_callee(%s, %s, &%s, &%s))",
					receiver,
					cString(property.Name),
					nativeEmitter.cache(),
					method,
				))
			}
		}
	}
	var packed, count string
	if property, ok := expression.Closure.(ir.Property); ok && property.View != "" {
		packed, count = nativeEmitter.checkedCallableArguments(expression, closure)
	} else {
		packed, count = nativeEmitter.closureArguments(expression)
	}
	call := nativeEmitter.packedClosureCall(expression, closure, packed, count)
	processHost := false
	walkExpressions(nativeEmitter.program, func(value ir.Expression) {
		if host, ok := value.(ir.ProcessCall); ok && host.Operation == "performance" {
			processHost = true
		}
	})

	if expression.Direct > 0 {
		target := expression.Direct - 1
		extra := ""
		if nativeEmitter.program.PackedCountNeeded(target) {
			extra = ", " + count
		}
		call = fmt.Sprintf("%s(%s, %s%s)", nativeEmitter.functionName(target), closure, packed, extra)
	}
	if receiver != "" {
		if closure == "" {
			if exactCount {
				if closure == "" {
					call = fmt.Sprintf("%s(%s, %s, %s)", method, receiver, packed, count)
				}
			} else if nativeEmitter.program.ClosureConventionNeeded() && closure != "" {
				call = fmt.Sprintf("%s.code(%s, %s)", method, receiver, packed)
			} else {
				call = fmt.Sprintf("%s(%s, %s)", method, receiver, packed)
			}
		} else {
			methodCall := fmt.Sprintf("%s(%s, %s)", method, receiver, packed)
			if nativeEmitter.program.ClosureReceiversNeeded() {
				call = fmt.Sprintf(
					"adamic_closure_receiver_call(%s, %s, %s, %s, %d)",
					closure,
					receiver,
					packed,
					count,
					nativeEmitter.packedArgumentSize(expression),
				)
				if nativeEmitter.program.PackedCountNeededFromCall(expression) {
					methodCall = fmt.Sprintf("adamic_method_call(%s, %s, %s, %s)", method, receiver, packed, count)
				} else if nativeEmitter.program.ClosureConventionNeeded() {
					methodCall = fmt.Sprintf("%s.code(%s, %s)", method, receiver, packed)
				}
			}
			if nativeEmitter.program.ClosureConventionNeeded() {
				if nativeEmitter.program.PackedCountNeededFromCall(expression) {
					methodCall = fmt.Sprintf("adamic_method_call(%s, %s, %s, %s)", method, receiver, packed, count)
				} else {
					methodCall = fmt.Sprintf("%s.code(%s, %s)", method, receiver, packed)
				}
			}
			call = fmt.Sprintf("(%s != NULL ? %s : %s)", closure, call, methodCall)
		}
	}

	// Preserve user closures' counted and receiver conventions in mixed programs.
	// Only the five runtime performance closures use padding and void-view cleanup.
	if processHost && closure != "" && expression.Direct == 0 {
		call = fmt.Sprintf(
			"(adamic_node_performance_method(%s) ? adamic_node_performance_invoke(%s, %s, %s, %t) : %s)",
			closure,
			closure,
			packed,
			count,
			expression.Returns == 0,
			call,
		)
	}

	if expression.Returns == 0 && !nativeEmitter.discarding(closureArm) && !nativeEmitter.discarding(methodArm) {
		nativeEmitter.line("%s;", call)
		if nativeEmitter.program.ClosureMayThrow(expression) {
			nativeEmitter.checkThrown()
		}
		return "0"
	}
	if expression.Returns == 0 {
		// A void call's result is released, never read: its function value may hold one that
		// returns a value, a reference owned as every closure's result is. A throw gives back a zero
		// value and leaves before the release.
		result := nativeEmitter.temporary()
		nativeEmitter.line("adamic_value %s = %s;", result, call)
		nativeEmitter.closureThrown()
		if receiver == "" || closure == "" {
			// A function value, or the proven method alone.
			nativeEmitter.line("%s", nativeEmitter.discard(closureArm, closure, result))
			return "0"
		}
		// Whichever arm ran: closure is NULL where the class's method did.
		if nativeEmitter.discarding(closureArm) {
			nativeEmitter.line("if (%s != NULL) {", closure)
			nativeEmitter.line("\t%s", nativeEmitter.discard(closureArm, closure, result))
			if nativeEmitter.discarding(methodArm) {
				nativeEmitter.line("} else {")
				nativeEmitter.line("\t%s", nativeEmitter.discard(methodArm, "", result))
			}
			nativeEmitter.line("}")
		} else {
			nativeEmitter.line("if (%s == NULL) {", closure)
			nativeEmitter.line("\t%s", nativeEmitter.discard(methodArm, "", result))
			nativeEmitter.line("}")
		}
		return "0"
	}
	result := nativeEmitter.temporary()
	nativeEmitter.line("adamic_value %s = %s;", result, call)
	// A throw gives back a zero value, nothing to let go.
	if nativeEmitter.program.ClosureMayThrow(expression) {
		nativeEmitter.checkThrown()
	}
	if expression.Returns == ir.Union {
		return nativeEmitter.callableUnionResult(result, closure, method)
	}
	if expression.Returns.IsReference() {
		// A closure's result comes back owned.
		return nativeEmitter.own(expression.Returns, fmt.Sprintf("(%s)%s.reference", cType(expression.Returns), result))
	}
	return nativeEmitter.snapshot(
		expression.Returns,
		unslotted(expression.Returns, result+"."+member(expression.Returns)),
	)
}

// discarding reports whether a void call has something of its result to let go of (discard).
func (nativeEmitter *emitter) discarding(discard ir.Discard) bool {
	return discard == ir.DiscardReference ||
		discard == ir.DiscardFlagged && nativeEmitter.closureDiscards ||
		discard == ir.DiscardMixed
}

// discard is the C that lets go of result, what a void call gave back, as ir.Discard says: released,
// never read. closure is the function value called, whose own discard field says where the call's
// targets are Unknown or disagree. ADAMIC_CLOSURE_DISCARDS is off only where no such call is made or
// no closure made returns a reference, so such a call's result then holds nothing.
func (nativeEmitter *emitter) discard(discard ir.Discard, closure, result string) string {
	switch discard {
	case ir.DiscardReference:
		return fmt.Sprintf("adamic_release(%s.reference);", result)
	case ir.DiscardFlagged:
		if nativeEmitter.closureDiscards {
			return fmt.Sprintf("adamic_closure_discard(%s, %s);", closure, result)
		}
	case ir.DiscardMixed:
		// Lowering says not yet to such a call (voidDiscards): it never reaches the emitter.
		panic("native: a void call whose targets disagree on releasing their result")
	}
	return ""
}

// closureDiscard stores in a closure just made how a void call that can't know its function lets go
// of its result: by the made function's Returns, whose adamic_value member the type alone fixes.
// adamic_closure_new stores nothing to let go; a canonical closure takes its discard as it's made.
func (nativeEmitter *emitter) closureDiscard(closure string, function int) {
	if discard := nativeEmitter.discardOf(function); nativeEmitter.closureDiscards && discard != "adamic_discard_none" {
		nativeEmitter.line("%s->discard = %s;", closure, discard)
	}
}

// discardOf is the adamic_discard a closure of function is made with.
func (nativeEmitter *emitter) discardOf(function int) string {
	if nativeEmitter.program.Functions[function].Returns.IsReference() {
		return "adamic_discard_reference"
	}
	return "adamic_discard_none"
}
