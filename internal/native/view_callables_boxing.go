package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// The code slot identifies the producer's representation. The callable's viewed
// type identifies the caller's representation. The ordinary closure ABI remains
// unchanged; only the values at that boundary are adapted.
func (e *emitter) viewCallableBoxedInvoke(expression ir.CallClosure, received bool) string {
	from := []ir.Type{}
	if received {
		from = append(from, ir.Object)
	}
	for _, argument := range expression.Arguments {
		from = append(from, argument.Type())
	}
	return e.viewCallableBoxedInvokeTypes(from, expression.Returns, received)
}

func (e *emitter) viewCallableBoxedInvokeTypes(from []ir.Type, returns ir.Type, received bool) string {
	needed := returns == ir.Union
	for _, argument := range from {
		needed = needed || argument == ir.Union
	}
	for _, function := range e.program.Functions {
		if !function.Closure {
			continue
		}
		needed = needed || function.Returns == ir.Union
		for _, local := range function.Parameters {
			needed = needed || e.program.Locals[local].Type == ir.Union
		}
	}
	if !needed {
		return "adamic_node_performance_invoke"
	}
	name := e.temporary()
	var body strings.Builder
	body.WriteString("#include <string.h>\n")
	fmt.Fprintf(&body, "static adamic_value %s(adamic_closure *self, adamic_value *arguments, size_t count, bool discard) {\n", name)
	for index, function := range e.program.Functions {
		if !function.Closure || function.Receiver != received {
			continue
		}
		fmt.Fprintf(&body, "if (%s) {\n", e.closureCodeIdentity("self", index))
		call := ir.CallClosure{Closure: ir.MakeClosure{Function: index}, Direct: index + 1}
		layout := e.program.ClosureArgumentLayout(call)
		// Preserve every actual argument and zero every missing producer slot.
		minimum := max(1, len(from), len(layout.Fixed))
		fmt.Fprintf(&body, "size_t capacity = count > %d ? count : %d;\nadamic_value adapted[capacity];\nmemset(adapted, 0, sizeof adapted);\nfor (size_t i = 0; i < count; i++) adapted[i] = arguments[i];\n", minimum, minimum)
		release := []string{}
		for i, local := range function.Parameters {
			if i >= len(from) {
				break
			}
			to := e.program.Locals[local].Type
			if from[i] == to || from[i] != ir.Union && to != ir.Union {
				continue
			}
			value := unslotted(from[i], fmt.Sprintf("arguments[%d].%s", i, member(from[i])))
			if from[i] == ir.Union {
				value = fmt.Sprintf("adamic_union_narrow(%s, %d)", value, to)
			}
			adapted, fresh := converted(from[i], to, value)
			fmt.Fprintf(&body, "if (count > %d) adapted[%d].%s = %s;\n", i, i, member(to), slotted(to, adapted))
			if fresh {
				release = append(release, fmt.Sprintf("if (count > %d) adamic_release(adapted[%d].reference);\n", i, i))
			}
		}
		packed := "adapted"
		adapter := emitter{program: e.program, reuse: e.reuse, indent: 1, scopes: [][]string{{}}}
		slots := []string{}
		for i, of := range layout.Fixed {
			slots = append(slots, fmt.Sprintf("{.%s = %s}", member(of), slotted(of, packedParameter(of, packed, "count", i))))
		}
		adaptedPacked := adapter.closureSlots(call, slots, packed, "count")
		adapter.line("adamic_value result = adamic_closure_call(self, %s, count);", adaptedPacked)
		adapter.releaseScopes(0)
		body.WriteString(adapter.out.String())
		for _, line := range release {
			body.WriteString(line)
		}
		body.WriteString("if (adamic_thrown != NULL) return (adamic_value){.reference = NULL};\n")
		if returns == 0 {
			if function.Returns.IsReference() {
				body.WriteString("adamic_release(result.reference);\n")
			}
			body.WriteString("(void)result; return (adamic_value){.reference = NULL};\n")
		} else if function.Returns != returns && (function.Returns == ir.Union || returns == ir.Union) && function.Returns != 0 {
			value := unslotted(function.Returns, "result."+member(function.Returns))
			if function.Returns == ir.Union {
				value = fmt.Sprintf("adamic_union_narrow(%s, %d)", value, returns)
			}
			adapted, _ := converted(function.Returns, returns, value)
			fmt.Fprintf(&body, "adamic_value converted_result = {.%s = %s};\n", member(returns), slotted(returns, adapted))
			if function.Returns == ir.Union && !returns.IsReference() {
				body.WriteString("adamic_release(result.reference);\n")
			}
			body.WriteString("return converted_result;\n")
		} else {
			body.WriteString("return result;\n")
		}
		body.WriteString("}\n")
	}
	unknownUnion := returns == ir.Union
	for _, of := range from {
		unknownUnion = unknownUnion || of == ir.Union
	}
	if unknownUnion {
		body.WriteString("(void)self; (void)arguments; (void)count; (void)discard; static const char message[] = \"callable ABI adapter: unknown producer signature\"; adamic_panic(message, sizeof message - 1);\n}\n")
	} else {
		body.WriteString("return adamic_node_performance_invoke(self, arguments, count, discard);\n}\n")
	}
	e.declarations = append(e.declarations, body.String())
	return name
}

// Sort's runtime callback also receives adamic_value words. Preserve its signed
// comparison result while adapting the producer's parameters before invocation.
func (e *emitter) viewCallableBoxedComparator(element ir.Type) string {
	invoke := e.viewCallableBoxedInvokeTypes([]ir.Type{element, element}, ir.Number, false)
	if invoke == "adamic_node_performance_invoke" {
		return "adamic_compare_closure"
	}
	name := e.temporary()
	e.declarations = append(e.declarations, fmt.Sprintf("static int %s(adamic_value left, adamic_value right, void *context) { adamic_value arguments[] = {left, right}; adamic_value result = %s((adamic_closure *)context, arguments, 2, false); return result.number < 0 ? -1 : result.number > 0 ? 1 : 0; }", name, invoke))
	return name
}

// Producer identity always compares pointers of the same calling convention.
func (e *emitter) closureCodeIdentity(value string, index int) string {
	if e.program.ClosureConventionNeeded() {
		if e.program.PackedCountNeeded(index) {
			return fmt.Sprintf("(%s->counted && %s->counted_code == %s)", value, value, e.functionName(index))
		}
		return fmt.Sprintf("(!%s->counted && %s->code == %s)", value, value, e.functionName(index))
	}
	return fmt.Sprintf("%s->code == %s", value, e.functionName(index))
}

func (e *emitter) viewCallbackCall(callback string, expression ir.Expression, kind int, from []ir.Type, returns ir.Type, slots ...string) string {
	invoke := e.viewCallableBoxedInvokeTypes(from, returns, false)
	if invoke == "adamic_node_performance_invoke" {
		return e.callbackCall(callback, expression, kind, slots...)
	}
	return fmt.Sprintf("%s(%s, (adamic_value[]){%s}, %d, false)", invoke, callback, strings.Join(slots, ", "), len(slots))
}

func (e *emitter) viewClosureComparator(sort ir.ArraySort) string {
	invoke := e.viewCallableBoxedInvokeTypes([]ir.Type{sort.Element, sort.Element}, ir.Number, false)
	if invoke == "adamic_node_performance_invoke" {
		return e.closureComparator(sort)
	}
	name := e.temporary()
	e.declarations = append(e.declarations, fmt.Sprintf("static int %s(adamic_value left, adamic_value right, void *context) { adamic_value arguments[] = {left, right}; adamic_value result = %s((adamic_closure *)context, arguments, 2, false); return result.number < 0 ? -1 : result.number > 0 ? 1 : 0; }", name, invoke))
	return name
}

// Evaluate once, retaining both the actual argument words and the area layout.
func (e *emitter) viewClosureArguments(call ir.CallClosure) (string, string, string) {
	if len(call.Spread) != 0 {
		source := e.packCallArguments(call.Arguments, call.Spread)
		layout := e.program.ClosureArgumentLayout(call)
		slots := []string{}
		for i, of := range layout.Fixed {
			slots = append(slots, fmt.Sprintf("{.%s = %s}", member(of), slotted(of, packedParameter(of, source+"->elements", source+"->length", i))))
		}
		return e.closureSlots(call, slots, source+"->elements", source+"->length"), source + "->length", source + "->elements"
	}
	slots := []string{}
	for _, argument := range call.Arguments {
		slots = append(slots, fmt.Sprintf("{.%s = %s}", member(argument.Type()), slotted(argument.Type(), e.value(argument))))
	}
	raw := "((adamic_value *)NULL)"
	if len(slots) != 0 {
		raw = "(adamic_value[]){" + strings.Join(slots, ", ") + "}"
	}
	count := fmt.Sprint(len(slots))
	return e.closureSlots(call, slots, raw, count), count, raw
}

// Only programs containing the Node host closures need its identity dispatcher.
func (e *emitter) hostClosuresNeeded() bool {
	needed := false
	walkExpressions(e.program, func(expression ir.Expression) {
		if call, ok := expression.(ir.ProcessCall); ok {
			needed = needed || call.Operation == "performance" || call.Operation == "stdoutHandle"
		}
	})
	return needed
}
