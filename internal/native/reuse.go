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

	// lending are the arrays a variable borrows an element from (planElementBorrows), never moved.
	lending map[int]bool
}

// planReuse makes the plan for a program.
func planReuse(program *ir.Program, lending map[int]bool) *reusePlan {
	plan := &reusePlan{consumed: map[int]bool{}, spreads: map[*ir.Statement]map[int]bool{}, moves: map[*ir.Statement]map[int]bool{}, arrays: map[*ir.Statement]map[int]bool{}, lending: lending}
	comparators := map[int]bool{}
	walkExpressions(program, func(expression ir.Expression) {
		if sort, ok := expression.(ir.ArraySort); ok {
			targets := program.ClosureTargets(sort)
			if targets.Unknown {
				for target := range program.Functions {
					comparators[target] = true
				}
			}
			for _, target := range targets.Functions {
				comparators[target] = true
			}
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
				source, _ := variableRead(literal.Spread)
				if !plan.reusable(program, each.index, comparators, instruction, each.live[instruction.Id], source.Local) {
					continue
				}
				if plan.spreads[instruction.At] == nil {
					plan.spreads[instruction.At] = map[int]bool{}
				}
				plan.spreads[instruction.At][source.Local] = true
				if program.Locals[source.Local].Borrowed {
					plan.consumed[source.Local] = true
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
	// Virtual implementations share an ownership convention. If one consumes a position,
	// every implementation takes a count, including those that only read and then release it.
	for changed := true; changed; {
		changed = false
		for signature, targets := range program.MethodTargets {
			for position, parameter := range program.Functions[signature].Parameters {
				consumed := plan.consumed[parameter]
				for _, target := range targets {
					consumed = consumed || plan.consumed[program.Functions[target].Parameters[position]]
				}
				if !consumed {
					continue
				}
				for _, target := range append([]int{signature}, targets...) {
					local := program.Functions[target].Parameters[position]
					if !plan.consumed[local] {
						plan.consumed[local] = true
						changed = true
					}
				}
			}
		}
	}
	for _, each := range functions {
		forEachInstruction(each.graph, func(instruction *flow.Instruction) {
			walk(evaluated(instruction), func(expression ir.Expression) {
				call, ok := expression.(ir.Call)
				if !ok || program.CallExpandsArguments(call) {
					return
				}
				for index, argument := range call.Arguments {
					read, isRead := variableRead(argument)
					if !isRead || !plan.callConsumes(program, call, index) {
						continue
					}
					if plan.movable(program, instruction, each.live[instruction.Id], read) {
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
	} else if local.Borrowed {
		// A variable borrowed from an array (element_borrow.go): its count is the array's.
		return false
	}
	return !live[flow.DeclarationId(source+1)] || overwrites(program, instruction, source)
}

// overwrites reports whether an instruction gives a variable a new value on every way it ends, so
// the old value is dead after it even though the variable is live: a declaration or an assignment
// that can't throw. One that can throw leaves the old value in place for the catch, the finally or
// the caller that takes the throw.
func overwrites(program *ir.Program, instruction *flow.Instruction, local int) bool {
	return defines(instruction, local) && !flow.CanThrow(program, instruction)
}

// readOnlyInside reports whether, in an instruction, source is read only by one object spread and by
// reads of its fields inside that literal, a replaced field read at most once.
func (plan *reusePlan) readOnlyInside(instruction *flow.Instruction, source int) bool {
	expression := evaluated(instruction)
	spreads := 0
	var literal *ir.ObjectLiteral
	for _, each := range spreadsOf(expression) {
		if read, _ := variableRead(each.Spread); read.Local == source {
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
			// A field read only reads. A method's callee (Property.Method) calls code with the source
			// as this, which can keep it or read a field take has moved out, so it stays a read of its
			// own and the source isn't reused (reuse_spread_method.a, reuse_spread_method_alias.a).
			if property, ok := expression.(ir.Property); ok && !property.Method {
				if read, ok := variableRead(property.Object); ok && read.Local == source {
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
func (plan *reusePlan) movable(program *ir.Program, instruction *flow.Instruction, live map[flow.DeclarationId]bool, read ir.Read) bool {
	local := program.Locals[read.Local]
	if local.Captured || read.Checked || readsOf(evaluated(instruction), read.Local) != 1 {
		return false
	}
	if plan.lending[read.Local] {
		// A variable borrows an element of this array, with no count of its own: moving the array
		// to a callee that lets go of it would free the element under the borrower.
		return false
	}
	if local.Borrowed && !plan.consumed[read.Local] {
		// A borrowed parameter's count is its caller's: there's nothing here to move, and handing
		// it on as if there were gives the caller's object to a callee that may reuse or free it.
		return false
	}
	if !local.Global {
		return !live[flow.DeclarationId(read.Local+1)] || overwrites(program, instruction, read.Local)
	}
	if flow.CanThrow(program, instruction) {
		// A global moved out and then a throw: whoever takes it, in this function or a caller, may
		// read the global and find it NULL.
		return false
	}
	// A global: only when this statement assigns it anew, and nothing that runs while it's moved out
	// reads or writes it, or calls something this can't see into (a closure, or a callback). That's
	// every call in the statement, not only the callee: another argument, evaluated after this one
	// is moved, may read the global.
	assign, ok := (*instruction.At).(ir.Assign)
	if !ok || assign.Local != read.Local {
		return false
	}
	reached := false
	walk(evaluated(instruction), func(expression ir.Expression) {
		switch expression := expression.(type) {
		case ir.Call:
			for _, target := range program.CallTargets(expression) {
				if touches(program, target, read.Local, map[int]bool{}) {
					reached = true
				}
			}
		case ir.CallClosure, ir.MakeClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.ArraySort:
			// A Map's or Set's forEach is void, so it's never in the value an assignment evaluates;
			// in a function this calls, touches finds it.
			reached = true
		}
	})
	return !reached
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
					for _, target := range program.CallTargets(expression) {
						if touches(program, target, global, seen) {
							found = true
						}
					}
				case ir.CallClosure, ir.MakeClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.ArraySort, ir.MapForEach:
					found = true
				}
			}, statement)
		}
	}
	statement(program.Functions[function].Body)
	return found
}

// variableRead is the variable an expression reads: a read, or a read checked for undefined
// (ir.Defined), whose check the emitter makes when it evaluates the expression.
func variableRead(expression ir.Expression) (ir.Read, bool) {
	if defined, ok := expression.(ir.Defined); ok {
		expression = defined.Value
	}
	read, ok := expression.(ir.Read)
	return read, ok
}

// variable emits the check for undefined of an expression variableRead took apart, when it has one,
// and is the variable's own name: what reuse takes over or moves is the variable, never a copy of it
// (a global's read retains one).
func (e *emitter) variable(expression ir.Expression, read ir.Read) string {
	name := e.localName(read.Local)
	if read.Readiness != "" {
		e.checkReadyRead(read.Local, read.Readiness)
	}
	if defined, ok := expression.(ir.Defined); ok {
		e.checkDefined(name, defined.Message)
	}
	return name
}

// checkDefined emits ir.Defined's check of a value: NULL panics with the message.
func (e *emitter) checkDefined(value string, message string) {
	e.line("if (%s == NULL) {", value)
	e.line("\tstatic const char message[] = %s;", cString(message))
	e.line("\tadamic_panic(message, sizeof message - 1);")
	e.line("}")
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
			if _, isRead := variableRead(literal.Spread); isRead && !literal.NoReuse {
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
	read, ok := variableRead(literal.Spread)
	if !ok || !e.reuse.spreads[e.at][read.Local] {
		return "", false
	}
	source := e.variable(literal.Spread, read)
	unique := e.temporary()
	// Checked before any field's value is evaluated, which can't make anything else hold it: the plan
	// saw that nothing in this instruction but the literal reads the source, and only its fields.
	held := fmt.Sprintf("(%s && !%s->frozen)", uniquelyHeld(source), source)
	if literal.SpreadMaybeUndefined {
		// Undefined is nothing to take over: the object is made as JavaScript's {} is (spreadCopy).
		held = fmt.Sprintf("(%s != NULL && %s)", source, held)
	}
	e.line("bool %s = %s;", unique, held)
	object := e.own(ir.Object, fmt.Sprintf("(%s ? adamic_retain(%s) : %s)", unique, source, e.spreadCopy(literal, source)))
	e.emptySpread(literal, source, object)
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
		cache := e.cache()
		e.line("adamic_value *%s = adamic_object_field(%s, %s, &%s);", slot, object, cString(field.Name), cache)
		e.line("adamic_object_field_types(%s)[%s.index] = %d;", object, cache, field.Value.Type())
		e.line("adamic_object_initialized(%s)[%s.index] = %d;", object, cache, map[bool]int{true: 0, false: 1}[field.Uninitialized])
		if field.Value.Type().IsReference() {
			// A field moved out of a unique object left NULL behind, and releasing that is nothing.
			e.line("adamic_release(%s->reference);", slot)
			e.line("%s->reference = %s;", slot, e.kept(values[index]))
		} else {
			e.line("%s->%s = %s;", slot, member(field.Value.Type()), slotted(field.Value.Type(), values[index]))
		}
	}
	// A spread produces a plain object, even when its source allocation is reused.
	e.line("%s->class = NULL;", object)
	return object, true
}

// spreadCopy is the object a spread that isn't reused starts from: a copy of the source's, or, when
// the source may be undefined and is, a new object with the source type's other fields and the
// literal's own, as JavaScript's { ...undefined } is {} with the literal's fields written in.
func (e *emitter) spreadCopy(literal ir.ObjectLiteral, source string) string {
	copy := fmt.Sprintf("adamic_object_copy(%s)", source)
	if literal.SpreadReadiness != "" {
		copy = fmt.Sprintf("adamic_object_copy_checked(%s, %s)", source, cString(literal.SpreadReadiness))
	}
	if !literal.SpreadMaybeUndefined {
		return copy
	}
	return fmt.Sprintf("(%s != NULL ? %s : adamic_object_new(&%s))", source, copy, e.shape(emptyFields(literal)))
}

// emptySpread gives the fields of the object spreadCopy made for an undefined source the value
// undefined has in each: a new object's slots are zero, which is undefined only for a reference.
func (e *emitter) emptySpread(literal ir.ObjectLiteral, source string, object string) {
	if !literal.SpreadMaybeUndefined {
		return
	}
	lines := []string{}
	for index, field := range literal.Empty {
		if !field.Value.Type().IsReference() {
			lines = append(lines, fmt.Sprintf("\t%s->slots[%d].%s = %s;", object, index, member(field.Value.Type()), slotted(field.Value.Type(), e.value(field.Value))))
		}
	}
	if len(lines) == 0 {
		return
	}
	e.line("if (%s == NULL) {", source)
	for _, line := range lines {
		e.line("%s", line)
	}
	e.line("}")
}

// emptyFields is the layout of the object an undefined spread makes: the source type's fields the
// literal doesn't give, then the literal's own, which are written by name after.
func emptyFields(literal ir.ObjectLiteral) []ir.Field {
	return append(slices.Clone(literal.Empty), literal.Fields...)
}

// take emits a read of a field a reused spread replaces: moved out of the object when it's unique,
// since the literal writes over it and nothing else can see it, and retained otherwise.
func (e *emitter) take(property ir.Property) (string, bool) {
	if e.taking == nil || property.Optional || !property.Of.IsReference() || property.Of == ir.Union {
		return "", false
	}
	read, ok := variableRead(property.Object)
	if !ok || read.Local != e.taking.source || !e.taking.replaced[property.Name] {
		return "", false
	}
	object := e.variable(property.Object, read)
	slot := e.temporary()
	e.line("adamic_value *%s = adamic_object_field(%s, %s, &%s);", slot, object, cString(property.Name), e.cache())
	value := e.own(property.Of, fmt.Sprintf("(%s)(%s ? %s->reference : adamic_retain(%s->reference))", cType(property.Of), e.taking.unique, slot, slot))
	e.line("if (%s) {", e.taking.unique)
	e.line("\t%s->reference = NULL;", slot)
	e.line("}")
	return value, true
}

// handOver evaluates an argument a consumed parameter takes, as a reference handed to the callee:
// moved out of its variable where the plan says nothing reads it after, a statement's temporary as
// it is, and anything else retained. A moved or handed-over temporary stays the statement's to let go
// of until every argument is evaluated (handedOver): an argument after it may throw, and then the call
// is never made.
func (e *emitter) handOver(argument ir.Expression) string {
	if read, ok := variableRead(argument); ok && e.reuse.moves[e.at][read.Local] {
		variable := e.variable(argument, read)
		moved := e.temporary()
		e.line("%s %s = %s;", cType(read.Of), moved, variable)
		e.line("%s = NULL;", variable)
		e.owned = append(e.owned, moved)
		return moved
	}
	value := e.value(argument)
	if slices.Contains(e.owned, value) {
		return value
	}
	return retained(value)
}

// handedOver gives up the statement's temporaries a call's consumed parameters take, once every
// argument is evaluated and the call is about to be made: the callee lets go of them from here.
func (e *emitter) handedOver(arguments []string) {
	for _, argument := range arguments {
		if index := slices.Index(e.owned, argument); index >= 0 {
			e.owned = slices.Delete(e.owned, index, index+1)
		}
	}
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
			read, isRead := variableRead(expression.Array)
			closure, isClosure := expression.Callback.(ir.MakeClosure)
			// Not when the callback can throw: the loop that writes over the array in place (mapped)
			// has no way out halfway, with some elements results and the rest not, so a callback that
			// can throw gets the loop that makes a new array and tests for the throw after each call.
			if isRead && isClosure && len(program.Functions[closure.Function].Parameters) < 3 && sameSlots(expression.Element, expression.Result) && !program.Functions[closure.Function].MayThrow {
				sources = append(sources, read.Local)
			}
		case ir.ArrayLiteral:
			if len(expression.Spread) > 0 && expression.Spread[0] {
				if read, isRead := variableRead(expression.Elements[0]); isRead {
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
	read, ok := variableRead(expression.Array)
	if !ok || !e.reuse.arrays[e.at][read.Local] {
		return "", false
	}
	source := e.variable(expression.Array, read)
	unique := e.temporary()
	e.line("bool %s = %s;", unique, uniquelyHeld(source))
	callback := e.value(expression.Callback)
	mapped := e.own(ir.Array, fmt.Sprintf("(%s ? adamic_retain(%s) : adamic_array_new(%s->length, %t))", unique, source, source, expression.Result.IsReference()))
	count, index, result := e.temporary(), e.temporary(), e.temporary()
	e.line("size_t %s = %s->length;", count, source)
	e.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
	e.line("\tif (%s >= %s->length) {", index, source)
	e.line("\t\tstatic const char message[] = \"map: the array shrank while it was being mapped\";")
	e.line("\t\tadamic_panic(message, sizeof message - 1);")
	e.line("\t}")
	call := e.callbackCall(callback, expression.Callback, expression.CallbackType, fmt.Sprintf("%s->elements[%s]", source, index), fmt.Sprintf("{.number = (double)%s}", index), fmt.Sprintf("{.reference = %s}", source))
	e.line("\tadamic_value %s = %s;", result, call)
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
	read, ok := variableRead(literal.Elements[0])
	if !ok || !e.reuse.arrays[e.at][read.Local] {
		return "", false
	}
	source := e.variable(literal.Elements[0], read)
	unique := e.temporary()
	e.line("bool %s = %s;", unique, uniquelyHeld(source))
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
			e.line("adamic_array_push(%s, (adamic_value){.reference = %s});", array, retained(value))
		default:
			e.line("adamic_array_push(%s, (adamic_value){.%s = %s});", array, member(literal.Element), slotted(literal.Element, value))
		}
	}
	return array, true
}

// uniquelyHeld is the C test that a value nothing else can reach: its count is 1, and no Weak points
// at it. A Weak doesn't count, so a count of 1 alone leaves the Weak's holder able to read or write
// the value while it's being taken over, and to find the new value at the old one's place after.
func uniquelyHeld(value string) string {
	return fmt.Sprintf("(%s->heap.references == 1 && !adamic_weak_held(%s))", value, value)
}

// callConsumes requires a count to be handed over at this position for every
// implementation. MethodTargets joins conventions before moves are planned.
func (plan *reusePlan) callConsumes(program *ir.Program, call ir.Call, position int) bool {
	for _, target := range program.CallTargets(call) {
		parameters := program.Functions[target].Parameters
		if position >= len(parameters) || !plan.consumed[parameters[position]] {
			return false
		}
	}
	return true
}
