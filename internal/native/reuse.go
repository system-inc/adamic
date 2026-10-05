package native

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
)

// Reuse in place (Perceus, docs/memory.md): a value whose count is 1, consumed to build one of the
// same shape, takes over its memory instead of being freed while a copy is allocated.
//
// In stage 0 the one place a value is consumed to build one of its own shape is the spread,
// { ...source, field: value }, whose result always has the source's shape. Uniqueness is checked at
// runtime, as Perceus does: the count is exactly 1. What the compiler proves is that taking the
// memory is invisible, and it proves that here, where the plan is made:
//
//   - The source is a variable this function owns a count of: a local, or a consumed parameter
//     (below). A borrowed parameter's count is its caller's, so a count of 1 there means the caller
//     still holds it, and it's never reused.
//   - The source is dead after the instruction (liveness over internal/flow's graph), or the
//     instruction itself gives its variable a new value: nothing reads the old object again.
//   - In the instruction, the source is read only by the spread and by reads of its fields inside
//     the literal, and each field the literal replaces is read at most once. So while the fields'
//     values are evaluated, nothing but this literal can reach the object (its count is 1, and it
//     isn't handed to anything), nothing can change it, and JavaScript's order (the spread's fields
//     read first) holds without a copy. A replaced field's one read can then move its value out
//     instead of retaining it, which is what lets a recursive call down a uniquely held tree reuse
//     the next node too.
//
// A parameter that's such a source is consumed: its caller hands over a reference (a statement's
// temporary is handed over as it is, anything else retained), and the callee lets go of it on every
// way out but the one where its memory became the result. An argument a call takes that way is
// moved rather than retained when nothing reads its variable after the call: a local that's dead
// there, or a global that the statement is assigning anew and that nothing the call can reach
// touches, which is program 9's tree = insert(tree, value).
type reusePlan struct {
	// consumed are the parameters a caller hands over a reference to.
	consumed map[int]bool

	// spreads are, per statement, the variables whose spread there takes over their object when it's
	// unique.
	spreads map[*ir.Statement]map[int]bool

	// moves are, per statement, the variables whose one read there, as an argument a consumed
	// parameter takes, moves their value out.
	moves map[*ir.Statement]map[int]bool

	// arrays are, per statement, the variables whose array there, when it's unique, is taken over:
	// by a map that writes its results in place, or by a literal that spreads it first and appends
	// the rest to it.
	arrays map[*ir.Statement]map[int]bool
}

// planReuse makes the plan for a program.
func planReuse(program *ir.Program) *reusePlan {
	plan := &reusePlan{consumed: map[int]bool{}, spreads: map[*ir.Statement]map[int]bool{}, moves: map[*ir.Statement]map[int]bool{}, arrays: map[*ir.Statement]map[int]bool{}}
	comparators := map[int]bool{}
	walkExpressions(program, func(expression ir.Expression) {
		if sort, ok := expression.(ir.ArraySort); ok {
			comparators[sort.Comparator] = true
		}
	})
	type function struct {
		index int
		graph *flow.Function
		live  map[flow.InstructionId]map[flow.DeclarationId]bool
	}
	functions := []function{}
	for index := -1; index < len(program.Functions); index++ {
		graph := flow.Build(program, index)
		functions = append(functions, function{index, graph, flow.LiveOut(graph)})
	}
	// Spreads first: which parameters are consumed depends on them, and moves on that.
	for _, each := range functions {
		forEachInstruction(each.graph, func(instruction *flow.Instruction) {
			for _, literal := range spreadsOf(evaluated(instruction)) {
				source := literal.Spread.(ir.Read).Local
				if !plan.reusable(program, each.index, comparators, instruction, each.live[instruction.Id], source) {
					continue
				}
				if plan.spreads[instruction.At] == nil {
					plan.spreads[instruction.At] = map[int]bool{}
				}
				plan.spreads[instruction.At][source] = true
				if program.Locals[source].Borrowed {
					plan.consumed[source] = true
				}
			}
			for _, source := range arraysTaken(program, evaluated(instruction)) {
				if !plan.owned(program, each.index, comparators, instruction, each.live[instruction.Id], source, ir.Array) ||
					readsOf(evaluated(instruction), source) != 1 {
					continue
				}
				if plan.arrays[instruction.At] == nil {
					plan.arrays[instruction.At] = map[int]bool{}
				}
				plan.arrays[instruction.At][source] = true
				if program.Locals[source].Borrowed {
					plan.consumed[source] = true
				}
			}
		})
	}
	for _, each := range functions {
		forEachInstruction(each.graph, func(instruction *flow.Instruction) {
			walk(evaluated(instruction), func(expression ir.Expression) {
				call, ok := expression.(ir.Call)
				if !ok {
					return
				}
				parameters := program.Functions[call.Function].Parameters
				for index, argument := range call.Arguments {
					read, isRead := argument.(ir.Read)
					if !isRead || index >= len(parameters) || !plan.consumed[parameters[index]] {
						continue
					}
					if plan.movable(program, instruction, each.live[instruction.Id], read, call.Function) {
						if plan.moves[instruction.At] == nil {
							plan.moves[instruction.At] = map[int]bool{}
						}
						plan.moves[instruction.At][read.Local] = true
					}
				}
			})
		})
	}
	return plan
}

// reusable reports whether a spread of source in an instruction may take over source's object.
func (plan *reusePlan) reusable(program *ir.Program, function int, comparators map[int]bool, instruction *flow.Instruction, live map[flow.DeclarationId]bool, source int) bool {
	if !plan.owned(program, function, comparators, instruction, live, source, ir.Object) {
		return false
	}
	return plan.readOnlyInside(instruction, source)
}

// owned reports whether source is a variable of the type given that this function owns a count of
// (a local, or a parameter that can be consumed), dead after the instruction or given a new value
// by it.
func (plan *reusePlan) owned(program *ir.Program, function int, comparators map[int]bool, instruction *flow.Instruction, live map[flow.DeclarationId]bool, source int, valueType ir.Type) bool {
	local := program.Locals[source]
	if local.Global || local.Captured || local.Type != valueType {
		return false
	}
	if local.Function != function {
		// A variable of an enclosing function: a closure reaches it through a cell, never here.
		return false
	}
	if isParameter(program, source) {
		// Only a parameter nothing assigns, of a named function its callers call directly, can be
		// handed over: a closure is called from runtime loops, and a comparator from the sort.
		if !local.Borrowed || program.Functions[function].Closure || comparators[function] {
			return false
		}
	}
	return !live[flow.DeclarationId(source+1)] || defines(instruction, source)
}

// readOnlyInside reports whether, in an instruction, source is read only by one object spread and by
// reads of its fields inside that literal, a replaced field read at most once.
func (plan *reusePlan) readOnlyInside(instruction *flow.Instruction, source int) bool {
	expression := evaluated(instruction)
	spreads := 0
	var literal *ir.ObjectLiteral
	for _, each := range spreadsOf(expression) {
		if each.Spread.(ir.Read).Local == source {
			spreads++
			copied := each
			literal = &copied
		}
	}
	if spreads != 1 {
		return false
	}
	replaced := map[string]bool{}
	for _, field := range literal.Fields {
		replaced[field.Name] = true
	}
	inside := 0
	fieldReads := map[string]int{}
	for _, field := range literal.Fields {
		walk(field.Value, func(expression ir.Expression) {
			if property, ok := expression.(ir.Property); ok {
				if read, ok := property.Object.(ir.Read); ok && read.Local == source {
					inside++
					fieldReads[property.Name]++
				}
			}
		})
	}
	for name, count := range fieldReads {
		if replaced[name] && count > 1 {
			return false
		}
	}
	// The spread's read, plus the field reads inside the literal, must be all of them.
	return readsOf(expression, source) == 1+inside
}

// movable reports whether an argument that reads a variable may move its value into a consumed
// parameter: nothing reads the variable's current value after.
func (plan *reusePlan) movable(program *ir.Program, instruction *flow.Instruction, live map[flow.DeclarationId]bool, read ir.Read, callee int) bool {
	local := program.Locals[read.Local]
	if local.Captured || read.Checked || readsOf(evaluated(instruction), read.Local) != 1 {
		return false
	}
	if !local.Global {
		return !live[flow.DeclarationId(read.Local+1)] || defines(instruction, read.Local)
	}
	// A global: only when this statement assigns it anew, and nothing the call can reach reads or
	// writes it, or calls something this can't see into (a closure, or a callback).
	assign, ok := (*instruction.At).(ir.Assign)
	return ok && assign.Local == read.Local && !touches(program, callee, read.Local, map[int]bool{})
}

// touches reports whether a function, or anything it calls, reads or writes a global, or calls a
// function value (which might).
func touches(program *ir.Program, function int, global int, seen map[int]bool) bool {
	if seen[function] {
		return false
	}
	seen[function] = true
	found := false
	var statement func(statements []ir.Statement)
	statement = func(statements []ir.Statement) {
		for _, each := range statements {
			if assign, ok := each.(ir.Assign); ok && assign.Local == global {
				found = true
			}
			if declare, ok := each.(ir.Declare); ok && declare.Local == global {
				found = true
			}
			walkStatement(each, func(expression ir.Expression) {
				switch expression := expression.(type) {
				case ir.Read:
					if expression.Local == global {
						found = true
					}
				case ir.Call:
					if touches(program, expression.Function, global, seen) {
						found = true
					}
				case ir.CallClosure, ir.MakeClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.ArraySort:
					found = true
				}
			}, statement)
		}
	}
	statement(program.Functions[function].Body)
	return found
}

func isParameter(program *ir.Program, local int) bool {
	function := program.Locals[local].Function
	if function < 0 {
		return false
	}
	for _, parameter := range program.Functions[function].Parameters {
		if parameter == local {
			return true
		}
	}
	return false
}

// defines reports whether an instruction gives a variable a new value: a declaration or an
// assignment (a global's included, which the graph doesn't track).
func defines(instruction *flow.Instruction, local int) bool {
	switch statement := (*instruction.At).(type) {
	case ir.Assign:
		return statement.Local == local
	case ir.Declare:
		return statement.Local == local
	}
	return false
}

// evaluated is what an instruction evaluates: its expression, or, for a statement with several
// (a field or element write), the statement itself.
func evaluated(instruction *flow.Instruction) any {
	if instruction.Expression != nil {
		return instruction.Expression
	}
	switch statement := (*instruction.At).(type) {
	case ir.SetIndex, ir.SetProperty:
		return statement
	}
	return nil
}

func forEachInstruction(graph *flow.Function, visit func(*flow.Instruction)) {
	for _, block := range graph.Blocks {
		for _, id := range block.Instructions {
			visit(graph.Instructions[id])
		}
	}
}

// spreadsOf is every object literal in a node whose spread is a variable read.
func spreadsOf(node any) []ir.ObjectLiteral {
	var literals []ir.ObjectLiteral
	walk(node, func(expression ir.Expression) {
		if literal, ok := expression.(ir.ObjectLiteral); ok {
			if _, isRead := literal.Spread.(ir.Read); isRead {
				literals = append(literals, literal)
			}
		}
	})
	return literals
}

// readsOf counts the reads of a variable in a node.
func readsOf(node any, local int) int {
	count := 0
	walk(node, func(expression ir.Expression) {
		if read, ok := expression.(ir.Read); ok && read.Local == local {
			count++
		}
	})
	return count
}

// walk visits every expression in a node, outermost first, by reflection: a switch over the IR's
// expressions would miss one added later, and a missed read here is a reuse that isn't invisible.
func walk(node any, visit func(ir.Expression)) {
	walkStatement(node, visit, nil)
}

// walkStatement is walk, handing any statements it meets (a block's, a loop's) to statements, or
// skipping them when that's nil.
func walkStatement(node any, visit func(ir.Expression), statements func([]ir.Statement)) {
	var each func(value reflect.Value)
	each = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return
			}
			if expression, ok := value.Interface().(ir.Expression); ok {
				visit(expression)
			}
			each(value.Elem())
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				each(value.Field(index))
			}
		case reflect.Slice:
			if inner, ok := value.Interface().([]ir.Statement); ok {
				if statements != nil {
					statements(inner)
				}
				return
			}
			for index := 0; index < value.Len(); index++ {
				each(value.Index(index))
			}
		case reflect.Array:
			for index := 0; index < value.Len(); index++ {
				each(value.Index(index))
			}
		}
	}
	if node == nil {
		return
	}
	// The node itself first: each visits what it meets inside, and the node isn't inside itself.
	if expression, ok := node.(ir.Expression); ok {
		visit(expression)
	}
	each(reflect.ValueOf(node))
}

// walkExpressions visits every expression in a program.
func walkExpressions(program *ir.Program, visit func(ir.Expression)) {
	var statements func([]ir.Statement)
	statements = func(list []ir.Statement) {
		for _, statement := range list {
			walkStatement(statement, visit, statements)
		}
	}
	for _, function := range program.Functions {
		statements(function.Body)
	}
	statements(program.Main)
}

// taking is a spread being reused: while its literal's fields are evaluated, a read of one of the
// fields it replaces moves the value out when the object is unique (unique names the C flag).
type taking struct {
	source   int
	unique   string
	replaced map[string]bool
}

// reused emits a spread the plan reuses: when its object's count is 1, the object itself, with the
// literal's fields written over its own; otherwise a copy, as any spread. Either way the result is a
// reference the statement owns.
func (e *emitter) reused(literal ir.ObjectLiteral) (string, bool) {
	read, ok := literal.Spread.(ir.Read)
	if !ok || !e.reuse.spreads[e.at][read.Local] {
		return "", false
	}
	source := e.localName(read.Local)
	unique := e.temporary()
	// Checked before any field's value is evaluated, which can't make anything else hold it: the plan
	// saw that nothing in this instruction but the literal reads the source, and only its fields.
	e.line("bool %s = %s->heap.references == 1;", unique, source)
	object := e.own(ir.Object, fmt.Sprintf("(%s ? adamic_retain(%s) : adamic_object_copy(%s))", unique, source, source))
	replaced := map[string]bool{}
	for _, field := range literal.Fields {
		replaced[field.Name] = true
	}
	outer := e.taking
	e.taking = &taking{source: read.Local, unique: unique, replaced: replaced}
	values := make([]string, 0, len(literal.Fields))
	for _, field := range literal.Fields {
		values = append(values, e.value(field.Value))
	}
	e.taking = outer
	for index, field := range literal.Fields {
		slot := e.temporary()
		e.line("adamic_value *%s = adamic_object_field(%s, %s, &%s);", slot, object, cString(field.Name), e.cache())
		if field.Value.Type().IsReference() {
			// A field moved out of a unique object left NULL behind, and releasing that is nothing.
			e.line("adamic_release(%s->reference);", slot)
			e.line("%s->reference = adamic_retain(%s);", slot, values[index])
		} else {
			e.line("%s->%s = %s;", slot, member(field.Value.Type()), slotted(field.Value.Type(), values[index]))
		}
	}
	return object, true
}

// take emits a read of a field a reused spread replaces: moved out of the object when it's unique,
// since the literal writes over it and nothing else can see it, and retained otherwise.
func (e *emitter) take(property ir.Property) (string, bool) {
	if e.taking == nil || property.Optional || !property.Of.IsReference() || property.Of == ir.Union {
		return "", false
	}
	read, ok := property.Object.(ir.Read)
	if !ok || read.Local != e.taking.source || !e.taking.replaced[property.Name] {
		return "", false
	}
	slot := e.temporary()
	e.line("adamic_value *%s = adamic_object_field(%s, %s, &%s);", slot, e.localName(read.Local), cString(property.Name), e.cache())
	value := e.own(property.Of, fmt.Sprintf("(%s)(%s ? %s->reference : adamic_retain(%s->reference))", cType(property.Of), e.taking.unique, slot, slot))
	e.line("if (%s) {", e.taking.unique)
	e.line("\t%s->reference = NULL;", slot)
	e.line("}")
	return value, true
}

// handOver evaluates an argument a consumed parameter takes, as a reference handed to the callee:
// moved out of its variable where the plan says nothing reads it after, a statement's temporary as
// it is (the statement no longer releases it), and anything else retained.
func (e *emitter) handOver(argument ir.Expression) string {
	if read, ok := argument.(ir.Read); ok && e.reuse.moves[e.at][read.Local] {
		variable := e.localName(read.Local)
		moved := e.temporary()
		e.line("%s %s = %s;", cType(read.Of), moved, variable)
		e.line("%s = NULL;", variable)
		return moved
	}
	value := e.value(argument)
	if index := slices.Index(e.owned, value); index >= 0 {
		e.owned = slices.Delete(e.owned, index, index+1)
		return value
	}
	return "adamic_retain(" + value + ")"
}

// arraysTaken is every variable whose array an expression could take over: the array a map maps,
// when its callback is written there and takes no third argument (the array, which would see its
// elements replaced as it went) and its results are held as the elements are; and the array a
// literal spreads first.
func arraysTaken(program *ir.Program, node any) []int {
	var sources []int
	walk(node, func(expression ir.Expression) {
		switch expression := expression.(type) {
		case ir.ArrayMap:
			read, isRead := expression.Array.(ir.Read)
			closure, isClosure := expression.Callback.(ir.MakeClosure)
			if isRead && isClosure && len(program.Functions[closure.Function].Parameters) < 3 && sameSlots(expression.Element, expression.Result) {
				sources = append(sources, read.Local)
			}
		case ir.ArrayLiteral:
			if len(expression.Spread) > 0 && expression.Spread[0] {
				if read, isRead := expression.Elements[0].(ir.Read); isRead {
					sources = append(sources, read.Local)
				}
			}
		}
	})
	return sources
}

// sameSlots reports whether values of two types sit in an array's slots alike: the same type, or
// both references held as pointers.
func sameSlots(left, right ir.Type) bool {
	return left == right || lendable(left) && lendable(right)
}

// mapped emits a map the plan reuses: when its array is unique, each result is written over the
// element it came from, in the array itself; otherwise a new array, as any map.
func (e *emitter) mapped(expression ir.ArrayMap) (string, bool) {
	read, ok := expression.Array.(ir.Read)
	if !ok || !e.reuse.arrays[e.at][read.Local] {
		return "", false
	}
	source := e.localName(read.Local)
	unique := e.temporary()
	e.line("bool %s = %s->heap.references == 1;", unique, source)
	callback := e.value(expression.Callback)
	mapped := e.own(ir.Array, fmt.Sprintf("(%s ? adamic_retain(%s) : adamic_array_new(%s->length, %t))", unique, source, source, expression.Result.IsReference()))
	count, index, result := e.temporary(), e.temporary(), e.temporary()
	e.line("size_t %s = %s->length;", count, source)
	e.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
	e.line("\tif (%s >= %s->length) {", index, source)
	e.line("\t\tstatic const char message[] = \"map: the array shrank while it was being mapped\";")
	e.line("\t\tadamic_panic(message, sizeof message - 1);")
	e.line("\t}")
	e.line("\tadamic_value %s = %s->code(%s, (adamic_value[]){%s->elements[%s], {.number = (double)%s}, {.reference = %s}});", result, callback, callback, source, index, index, source)
	e.line("\tif (%s) {", unique)
	// The callback is done with the element it was handed: the result takes its place.
	if expression.Element.IsReference() {
		e.line("\t\tadamic_release(%s->elements[%s].reference);", source, index)
	}
	e.line("\t\t%s->elements[%s] = %s;", source, index, result)
	e.line("\t} else {")
	e.line("\t\tadamic_array_push(%s, %s);", mapped, result)
	e.line("\t}")
	e.line("}")
	return mapped, true
}

// spreadArray emits an array literal the plan reuses, one that spreads a variable's array first:
// when that array is unique, the literal is the array itself, the rest appended to it; otherwise a
// new array, as any literal.
func (e *emitter) spreadArray(literal ir.ArrayLiteral) (string, bool) {
	if len(literal.Spread) == 0 || !literal.Spread[0] {
		return "", false
	}
	read, ok := literal.Elements[0].(ir.Read)
	if !ok || !e.reuse.arrays[e.at][read.Local] {
		return "", false
	}
	source := e.localName(read.Local)
	unique := e.temporary()
	e.line("bool %s = %s->heap.references == 1;", unique, source)
	array := e.own(ir.Array, fmt.Sprintf("(%s ? adamic_retain(%s) : adamic_array_new(0, %t))", unique, source, literal.Element.IsReference()))
	e.line("if (!%s) {", unique)
	e.line("\tadamic_array_append(%s, %s);", array, source)
	e.line("}")
	for index, element := range literal.Elements[1:] {
		value := e.value(element)
		switch {
		case literal.Spread[index+1]:
			e.line("adamic_array_append(%s, %s);", array, value)
		case literal.Element.IsReference():
			e.line("adamic_array_push(%s, (adamic_value){.reference = adamic_retain(%s)});", array, value)
		default:
			e.line("adamic_array_push(%s, (adamic_value){.%s = %s});", array, member(literal.Element), slotted(literal.Element, value))
		}
	}
	return array, true
}
