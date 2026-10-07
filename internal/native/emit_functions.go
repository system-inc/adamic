// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) functionName(function int) string {
	return fmt.Sprintf("adamic_function_%d_%s", function, cIdentifier.ReplaceAllString(e.program.Functions[function].Name, ""))
}

func (e *emitter) signature(function int) string {
	declared := e.program.Functions[function]
	if declared.Closure {
		// Every closure's code is called the same way (adamic_code): its arguments and result as
		// adamic_value, whatever their types.
		return fmt.Sprintf("adamic_value %s(adamic_closure *self, adamic_value *arguments, size_t argument_count)", e.functionName(function))
	}
	returns := "void"
	if declared.Returns != 0 {
		returns = cType(declared.Returns)
	}
	parameters := []string{}
	name := e.functionName(function)
	if e.inRegion {
		// The version that makes its result in the region it's handed, never NULL (region.go).
		parameters = append(parameters, "adamic_region *region")
		name = e.regionFunctionName(function)
	}
	for _, parameter := range declared.Parameters {
		parameters = append(parameters, cType(e.program.Locals[parameter].Type)+" "+e.localName(parameter))
	}
	if len(parameters) == 0 {
		parameters = append(parameters, "void")
	}
	return fmt.Sprintf("%s %s(%s)", returns, name, strings.Join(parameters, ", "))
}

// functionBody emits a function's body. A string parameter is retained on entry and released on
// every way out, like any local, so a function may reassign it without touching its caller's.
func (e *emitter) functionBody(function ir.Function) {
	e.function = &function
	e.functionDepth = len(e.scopes)
	e.scopes = append(e.scopes, nil)
	// Recursion that runs out of stack panics, as Node's does, rather than crashing (stack.c).
	e.line("ADAMIC_CHECK_STACK();")
	if function.Closure {
		e.line("(void)self;")
		e.line("(void)arguments;")
		e.line("(void)argument_count;")
		for index, parameter := range function.Parameters {
			local := e.program.Locals[parameter]
			value := closureArgument(local.Type, index)
			if local.Type.IsReference() {
				value = fmt.Sprintf("(%s)%s", cType(local.Type), value)
			}
			e.line("%s %s = %s;", cType(local.Type), e.localName(parameter), value)
		}
	}
	e.allocateEnvironment(function.FrameEnvironment)
	for _, parameter := range function.Parameters {
		switch {
		case e.reuse.consumed[parameter]:
			// Its caller handed over a reference (reuse.go): it's the callee's to let go of.
			e.hold(e.localName(parameter))
		case e.program.Locals[parameter].Type.IsReference() && !e.program.Locals[parameter].Borrowed:
			// A borrowed parameter is its caller's, kept alive for the whole call.
			e.line("adamic_retain(%s);", e.localName(parameter))
			e.hold(e.localName(parameter))
		}
		if e.program.Locals[parameter].Uninitialized && !e.program.Locals[parameter].Captured {
			e.line("bool %s = true;", readyName(parameter))
		}
		if e.program.Locals[parameter].Captured {
			// A closure captured this parameter: from here on it lives in a cell.
			e.makeCell(parameter, e.localName(parameter), false)
		}
	}
	for index := range function.Body {
		e.statementAt(&function.Body[index])
	}
	if function.Returns == 0 {
		e.releaseScopes(e.functionDepth)
		if function.Closure {
			e.line("return (adamic_value){.number = 0};")
		}
	} else {
		// Lowering records every permitted implicit return in the IR. C cannot always see
		// that all paths return; reaching this guard still means an IR exit was lost.
		e.line("adamic_unreachable();")
	}
	e.scopes = e.scopes[:e.functionDepth]
	e.function = nil
}

func (e *emitter) returnStatement(statement ir.Return) {
	if statement.Value == nil {
		e.end()
		e.finallies(0)
		e.releaseScopes(e.functionDepth)
		if e.function != nil && e.function.Closure {
			e.line("return (adamic_value){.number = 0};")
		} else {
			e.line("return;")
		}
		return
	}
	value := ""
	if e.function != nil && e.inRegion {
		// A function that returns fresh makes what it returns in the region it was handed (region.go).
		if literal, isLiteral := statement.Value.(ir.ObjectLiteral); isLiteral && literal.Spread == nil {
			e.regionLiteralDepth = e.depth + 1
		}
		value = e.handRegion(statement.Value, "region")
		if read, ok := statement.Value.(ir.Read); ok {
			if local, known := e.regions.classObjects[e.functionIndex]; known && local == read.Local {
				e.regionValues[value] = true
			}
		}
		e.regionLiteralDepth = 0
	} else {
		value = e.value(statement.Value)
	}
	result := e.temporary()
	if statement.Value.Type().IsReference() && !e.regionValues[value] {
		held := statement.Value.Type()
		if _, isUndefined := statement.Value.(ir.Undefined); isUndefined && e.function != nil && e.function.Returns.IsReference() {
			// undefined is a null pointer of whatever reference the function returns, a string's
			// among them, not the object pointer it's typed as alone.
			held = e.function.Returns
		}
		e.line("%s %s = %s;", cType(held), result, e.kept(value))
	} else {
		e.line("%s %s = %s;", cType(statement.Value.Type()), result, value)
	}
	e.end()
	if len(e.handlers) > 0 {
		// Every finally open runs before the function returns. The result waits in a scope of its
		// own, so a throw from a finally, which replaces the return, lets go of it.
		held := []string{}
		if statement.Value.Type().IsReference() {
			held = append(held, result)
		}
		e.scopes = append(e.scopes, held)
		e.finallies(0)
		e.scopes = e.scopes[:len(e.scopes)-1]
	}
	e.releaseScopes(e.functionDepth)
	if e.function != nil && e.function.Closure {
		e.line("return (adamic_value){.%s = %s};", member(statement.Value.Type()), slotted(statement.Value.Type(), result))
	} else {
		e.line("return %s;", result)
	}
}

// arguments evaluates a call's arguments in order, each fitted to its parameter: a number or undefined
// made number | undefined's two words. An optional parameter the call leaves out gets undefined. This
// is done here, not in lowering, since only here is every function's signature known: lowering may
// meet a call before the function it calls.
func (e *emitter) arguments(call ir.Call) []string {
	parameters := e.program.Functions[call.Function].Parameters
	arguments := make([]string, 0, len(parameters))
	handed := []string{}
	defer func() { e.handedOver(handed) }()
	for index, argument := range call.Arguments {
		if index >= len(parameters) {
			e.value(argument) // Extra arguments still run, before the call.
			continue
		}
		if e.reuse.callConsumes(e.program, call, index) {
			value := e.handOver(argument)
			handed = append(handed, value)
			arguments = append(arguments, value)
			continue
		}
		value := ""
		if lent, ok := e.lentArgument(call, index); ok {
			value = lent
		} else if e.statementRegion != "" && !e.regions.callEscapes(call, index) {
			// A parameter that flows nowhere: a fresh value handed to it lives in the statement's region.
			value = e.handRegion(argument, "&"+e.statementRegion)
		} else {
			value = e.value(argument)
		}
		if index < len(parameters) && e.program.Locals[parameters[index]].Type.IsMaybe() {
			of := e.program.Locals[parameters[index]].Type
			if _, isUndefined := argument.(ir.Undefined); isUndefined {
				value = zero(of)
			} else if argument.Type() == of.Present() {
				value = maybe(of, value)
			}
		}
		if index < len(parameters) && (e.program.Locals[parameters[index]].Type == ir.Union || e.program.Locals[parameters[index]].Type == ir.Weak) && argument.Type() != e.program.Locals[parameters[index]].Type {
			// A number, a boolean or a reference where a union goes, boxed here, where the signature
			// is known; a reference where a Weak goes, its handle.
			takes := e.program.Locals[parameters[index]].Type
			boxed, fresh := converted(argument.Type(), takes, value)
			if fresh {
				boxed = e.own(takes, boxed)
			}
			value = boxed
		}
		arguments = append(arguments, value)
	}
	for _, parameter := range parameters[min(len(call.Arguments), len(parameters)):] {
		if of := e.program.Locals[parameter].Type; of.IsMaybe() {
			arguments = append(arguments, zero(of))
		} else {
			arguments = append(arguments, "NULL")
		}
	}
	return arguments
}

// callThrough calls a function value: closure, or for a call through an interface (ir.Property's
// Method), the receiver's own function value or its class's method, found before the arguments are
// evaluated, as JavaScript reads object.name first.
func (e *emitter) callThrough(expression ir.CallClosure, closure string, receiver string) string {
	method := ""
	if receiver != "" {
		property := expression.Closure.(ir.Property)
		if function, known := e.exactReceiverMethod(property.Object, property.Name); known {
			// Keep the interface adapter's borrowed-input convention and the same
			// exception and result handling, but call its proven method directly.
			method = e.methodThunk(function)
		} else {
			method = e.temporary()
			e.line("adamic_method %s = NULL;", method)
			if property.View != "" {
				closure = e.emitViewCallableRead(property, receiver, method)
			} else {
				closure = e.own(ir.Closure, fmt.Sprintf("adamic_retain(adamic_object_callee(%s, %s, &%s, &%s))", receiver, cString(property.Name), e.cache(), method))
			}
		}
	}
	arguments := []string{}
	for _, argument := range expression.Arguments {
		arguments = append(arguments, fmt.Sprintf("{.%s = %s}", member(argument.Type()), slotted(argument.Type(), e.value(argument))))
	}
	packed := "NULL"
	if len(arguments) > 0 {
		packed = "(adamic_value[]){" + strings.Join(arguments, ", ") + "}"
	}
	call := fmt.Sprintf("adamic_node_performance_invoke(%s, %s, %d, %t)", closure, packed, len(arguments), expression.Returns == 0)
	if expression.Direct > 0 {
		call = fmt.Sprintf("%s(%s, %s, %d)", e.functionName(expression.Direct-1), closure, packed, len(arguments))
	}
	if receiver != "" {
		if closure == "" {
			call = fmt.Sprintf("%s(%s, %s, %d)", method, receiver, packed, len(arguments))
		} else {
			received := "(adamic_value[]){ {.reference = " + receiver + "}"
			if len(arguments) > 0 {
				received += ", " + strings.Join(arguments, ", ")
			}
			received += "}"
			call = fmt.Sprintf("(%s != NULL ? adamic_node_performance_invoke(%s, %s->receiver ? %s : %s, %d + (%s->receiver ? 1 : 0), %t) : %s(%s, %s, %d))", closure, closure, closure, received, packed, len(arguments), closure, expression.Returns == 0, method, receiver, packed, len(arguments))
		}
	}
	if expression.Returns == 0 {
		e.line("%s;", call)
		e.closureThrown()
		return "0"
	}
	result := e.temporary()
	e.line("adamic_value %s = %s;", result, call)
	// A throw gives back a zero value, nothing to let go.
	e.closureThrown()
	if expression.Returns.IsReference() {
		// A closure's result comes back owned.
		return e.own(expression.Returns, fmt.Sprintf("(%s)%s.reference", cType(expression.Returns), result))
	}
	return e.snapshot(expression.Returns, unslotted(expression.Returns, result+"."+member(expression.Returns)))
}
