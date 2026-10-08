// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

type loop struct {
	sourceName string
	breakLabel string
	broken     bool
	iteration  *loop
	depth      int
	label      string
	continued  bool
}

// A break to an outer target uses a goto, emitted only when needed, after the target's loop.
// Cleanup runs before the jump, including every finally and local scope it leaves.
type breakable struct {
	depth  int
	label  string
	broken bool
}

// block emits statements in a scope of their own.
func (e *emitter) block(statements []ir.Statement, after func()) {
	e.scopes = append(e.scopes, nil)
	for index := range statements {
		e.statementAt(&statements[index])
	}
	if after != nil {
		after()
	}
	e.releaseScopes(len(e.scopes) - 1)
	e.scopes = e.scopes[:len(e.scopes)-1]
}

// decide evaluates a condition and releases its temporaries before anything branches on it,
// returning the C expression holding the answer.
func (e *emitter) decide(condition ir.Expression) string {
	value := e.value(condition)
	if len(e.owned) == 0 {
		return value
	}
	name := e.temporary()
	e.line("bool %s = %s;", name, value)
	e.end()
	return name
}

// statementAt emits a statement where it stands in its slice, the address the reuse plan names it by.
func (e *emitter) statementAt(at *ir.Statement) {
	outer := e.at
	e.at = at
	if !e.regionStatement(at) {
		e.statement(*at)
	}
	e.at = outer
}

func (e *emitter) statement(statement ir.Statement) {
	switch statement := statement.(type) {
	case ir.Debugger:
		// No debugger is attached to native builds, so emit no instruction or trap.
	case ir.WriteLine:
		stream := "adamic_stdout"
		if statement.Stream == ir.Stderr {
			stream = "adamic_stderr"
		}
		value := e.value(statement.Value)
		e.line("adamic_write_line(%s, %s);", stream, value)
		e.end()
	case ir.AllocateEnvironment:
		// Emitted at function entry before captured parameters are initialized.
	case ir.Declare:
		if statement.Uninitialized && (e.program.Locals[statement.Local].Captured || e.program.Locals[statement.Local].Preallocated) {
			if e.program.Locals[statement.Local].EnvironmentCell {
				return
			}
			local := e.program.Locals[statement.Local]
			if local.Captured {
				e.makeCell(statement.Local, zero(local.Type), true)
				e.line("%s->ready = false;", e.cellReference(statement.Local))
			}
			return
		}
		if e.elementBorrows[e.at] {
			e.borrowElement(statement)
			return
		}
		local := e.program.Locals[statement.Local]
		value := zero(local.Type)
		if statement.Value != nil {
			value = e.value(statement.Value)
		}
		// The declaration is the statement's last write: what it owns, the variable can take.
		owned := e.taken(value)
		if local.Global {
			e.store(statement.Local, value, owned)
			e.line("%s = %t;", readyName(statement.Local), !statement.Uninitialized)
			if local.Uninitialized {
				e.line("%s_declared = true;", readyName(statement.Local))
			}
			e.initialized = append(e.initialized, statement.Local)
		} else {
			e.declareLocal(statement.Local, value, owned)
			if local.Uninitialized {
				if local.Captured {
					e.line("%s = %t;", e.localReady(statement.Local), !statement.Uninitialized)
				} else {
					e.line("bool %s = %t; (void)%s;", readyName(statement.Local), !statement.Uninitialized, readyName(statement.Local))
				}
			}
		}
		e.end()
	case ir.Assign:
		if e.program.Locals[statement.Local].Counter {
			// The loop's update, counter + step or counter - step, is the only write a counter has, and
			// its step a whole constant within 2^53 (lower/counters.go), which the cast keeps exactly.
			sum, isSum := statement.Value.(ir.Binary)
			read, isRead := sum.Left.(ir.Read)
			if !isSum || (sum.Operator != ir.Add && sum.Operator != ir.Subtract) || !isRead || read.Local != statement.Local {
				panic(fmt.Sprintf("native: compiler bug: counter %s written other than by a step", e.program.Locals[statement.Local].Name))
			}
			operator := "+="
			if sum.Operator == ir.Subtract {
				operator = "-="
			}
			e.line("%s %s (int64_t)(%s);", e.localName(statement.Local), operator, e.value(sum.Right))
			return
		}
		if parts, appends := e.appendsTo(statement); appends {
			e.appendTo(statement.Local, parts)
			e.end()
			break
		}
		value := e.value(statement.Value)
		if e.program.Locals[statement.Local].Uninitialized && e.program.Locals[statement.Local].Global && !e.program.Locals[statement.Local].Hoisted {
			e.line("if (!%s_declared) {", readyName(statement.Local))
			message := "ReferenceError: Cannot access '" + e.program.Locals[statement.Local].Name + "' before initialization"
			e.line("\tadamic_panic(%s, %d);", cString(message), len(message))
			e.line("}")
		}
		if statement.Checked {
			// After the value, as JavaScript does: the right side runs, then the write throws.
			e.checkReady(statement.Local)
		}
		// The store is the statement's last write: nothing after it can assign the variable again
		// while the statement still reads the value, so what the statement owns, the variable takes.
		e.store(statement.Local, value, e.taken(value))
		if e.program.Locals[statement.Local].Uninitialized {
			e.line("%s = %t;", e.localReady(statement.Local), !statement.Uninitialized)
		}
		e.end()
	case ir.Evaluate:
		if splice, isSplice := statement.Value.(ir.ArraySplice); isSplice {
			// What a splice removes, nothing here uses: it's let go of, and no array is made for it.
			e.line("adamic_array_remove(%s);", e.spliceArguments(splice))
			e.end()
			break
		}
		if call, isCall := statement.Value.(ir.Call); isCall && call.Returns == 0 {
			if call.Accessor != "" {
				e.accessorCall(call)
			} else {
				e.line("%s;", e.callCode(call, e.arguments(call)))
			}
			if e.program.CallMayThrow(call) {
				e.checkThrown()
			}
		} else {
			// The value goes, but whatever making it did stays: a push's append is its effect.
			e.line("(void)%s;", e.value(statement.Value))
		}
		e.end()
	case ir.Return:
		e.returnStatement(statement)
	case ir.SetIndex:
		array := e.value(statement.Array)
		index := e.value(statement.Index)
		value := e.value(statement.Value)
		if statement.Array.Type().IsTypedArray() {
			e.line("adamic_typed_array_set(%s, %s, %s);", array, index, value)
			e.end()
			break
		}
		if statement.Element.IsReference() {
			value = e.heldReferenceIn(array, value)
		}
		setter := "adamic_array_set"
		if e.hasArrayHoles() {
			setter = "adamic_array_holes_set"
		}
		e.line("%s(%s, %s, (adamic_value){.%s = %s});", setter, array, index, member(statement.Element), slotted(statement.Element, value))
		e.end()
	case ir.SetProperty:
		object := e.value(statement.Object)
		value := e.value(statement.Value)
		// The object may be undefined where the checker narrowed it away and a call since put it back
		// (ir.Defined): JavaScript throws at the write, after the value, and so does this.
		e.line("if (%s == NULL) {", object)
		e.line("\tstatic const char message[] = %s;", cArray("TypeError: Cannot set properties of undefined (setting '"+statement.Name+"')"))
		e.line("\tadamic_panic(message, sizeof message - 1);")
		e.line("}")
		e.line("adamic_object_check_data_write(%s, %s);", object, cString(statement.Name))
		slot := e.temporary()
		cache := e.cache()
		if e.program.CheckedFields[statement.Name] {
			e.line("adamic_object_view_write(%s, %s, &%s, %d, %s, %s);", object, cString(statement.Name), cache, statement.Value.Type(), cString(map[ir.Type]string{ir.Number: "number", ir.Boolean: "boolean", ir.String: "string"}[statement.Value.Type()]), cString("<write>."+statement.Name))
		}
		if e.program.CheckedFields[statement.Name] {
			e.line("adamic_value *%s = adamic_object_write_field(%s, %s, &%s);", slot, object, cString(statement.Name), cache)
		} else {
			e.line("adamic_value *%s = %s;", slot, e.writeFieldSlot(object, statement.Name, statement.Class))
		}
		converted := e.program.CheckedFields[statement.Name] && statement.Value.Type() <= ir.Boolean
		if converted {
			e.line("if (adamic_object_field_types(%s)[(size_t)(%s - %s->slots)] == 10) {", object, slot, object)
			boxed := fmt.Sprintf("adamic_box_number(%s)", value)
			if statement.Value.Type() == ir.Boolean {
				boxed = fmt.Sprintf("(%s ? &adamic_box_true : &adamic_box_false)", value)
			}
			e.line("adamic_heap *view_new_value = (adamic_heap *)%s;", boxed)
			e.line("adamic_release(%s->reference);", slot)
			e.line("%s->reference = view_new_value;", slot)
			e.line("} else {")
		}
		if statement.Value.Type().IsReference() {
			// The new reference is taken before the old is let go: they may be the same.
			old := e.temporary()
			e.line("void *%s = %s->reference;", old, slot)
			e.line("%s->reference = %s;", slot, e.keptIn(object, value))
			if len(e.program.GraphTypes) != 0 {
				e.dropIn(object, old)
			} else {
				e.line("if (%s != NULL) adamic_release(%s);", old, old)
			}
		} else {
			e.line("%s->%s = %s;", slot, member(statement.Value.Type()), slotted(statement.Value.Type(), value))
		}
		e.line("adamic_object_field_types(%s)[(size_t)(%s - %s->slots)] = %d;", object, slot, object, statement.Value.Type())
		if converted {
			e.line("}")
		}
		// Direct-slot stores can bypass write_field; readiness must follow the store too.
		e.line("adamic_object_set_initialized(%s, %s, %t);", object, cString(statement.Name), !statement.Uninitialized)
		e.end()
	case ir.Panic:
		// The program ends here, so nothing it holds needs letting go.
		message := e.value(statement.Message)
		e.line("adamic_panic((%s)->bytes, (%s)->length);", message, message)
		e.owned = nil
	case ir.If:
		condition := e.decide(statement.Condition)
		e.line("if (%s) {", unwrap(condition))
		e.nested(statement.Then, nil)
		if len(statement.Else) > 0 {
			e.line("} else {")
			e.nested(statement.Else, nil)
		}
		e.line("}")
	case ir.Labeled:
		e.labeled(statement)
	case ir.Block:
		e.line("{")
		e.nested(statement.Body, nil)
		e.line("}")
	case ir.Loop:
		e.loop(statement)
	case ir.ForOf:
		e.forOf(statement)
	case ir.Switch:
		e.switchStatement(statement)
	case ir.Break:
		if statement.Label != "" {
			e.labeledJump(statement.Label, false)
			break
		}
		target := e.breakables[len(e.breakables)-1-statement.Depth]
		e.finallies(e.innerHandlers(target.depth))
		e.releaseScopes(target.depth)
		if statement.Depth == 0 {
			e.line("break;")
		} else {
			target.broken = true
			e.line("goto %s;", target.label)
		}
	case ir.Continue:
		if statement.Label != "" {
			e.labeledJump(statement.Label, true)
			break
		}
		var current *loop
		for index := len(e.loops) - 1; index >= 0; index-- {
			if e.loops[index].sourceName == "" {
				current = e.loops[index]
				break
			}
		}
		if current == nil {
			panic("native: continue outside a loop")
		}
		current.continued = true
		e.finallies(e.innerHandlers(current.depth))
		e.releaseScopes(current.depth)
		e.line("goto %s;", current.label)
	case ir.Throw:
		e.throwStatement(statement)
	case ir.Try:
		e.tryStatement(statement)
	default:
		// Lowering only produces statements this switch knows. Reaching this is a compiler bug.
		panic(fmt.Sprintf("native: no C for %T", statement))
	}
}

// nested emits statements one level in, in a scope of their own.
func (e *emitter) nested(statements []ir.Statement, after func()) {
	e.indent++
	e.block(statements, after)
	e.indent--
}

// loop emits every loop the same way: for (;;), the condition checked as an if that breaks, the
// body in its own scope, then a continue label and the update. A continue releases the body's
// locals and jumps to the label, so it runs the update and the check, as JavaScript's does.
func (e *emitter) loop(statement ir.Loop) {
	e.temporaries++
	current := &loop{label: fmt.Sprintf("adamic_continue_%d", e.temporaries)}
	check := func() {
		condition := e.decide(statement.Condition)
		e.line("if (!(%s)) {", unwrap(condition))
		e.line("\tbreak;")
		e.line("}")
	}
	e.line("for (;;) {")
	e.indent++
	if !statement.CheckAfter {
		check()
	}
	// The body is emitted aside first, since only then is it known whether anything continues.
	saved := e.out
	e.out = strings.Builder{}
	current.depth = len(e.scopes)
	e.linkLabels(statement.Labels, current)
	e.loops = append(e.loops, current)
	target := &breakable{depth: current.depth, label: e.temporary()}
	e.breakables = append(e.breakables, target)
	e.line("{")
	e.nested(statement.Body, nil)
	e.line("}")
	e.breakables = e.breakables[:len(e.breakables)-1]
	e.loops = e.loops[:len(e.loops)-1]
	body := e.out.String()
	e.out = saved
	e.out.WriteString(body)
	if current.continued {
		e.line("%s:;", current.label)
	}
	// Each iteration's let bindings are its own (ECMA-262, CreatePerIterationEnvironment): a captured
	// one gets a fresh cell holding the current value before the update runs, so a closure made in
	// this iteration keeps this iteration's.
	for _, local := range statement.PerIteration {
		if !e.program.Locals[local].Captured {
			continue
		}
		cell := e.cellName(local)
		fresh := e.temporary()
		e.line("adamic_cell *%s = adamic_cell_new(%s->value, %s->references);", fresh, cell, cell)
		e.line("%s->ready = %s->ready;", fresh, cell)
		e.line("if (%s->references) {", fresh)
		e.line("\tadamic_retain(%s->value.reference);", fresh)
		e.line("}")
		e.adoptGraph(fresh, "sizeof *"+fresh, e.program.Locals[local].GraphCell)
		e.line("adamic_release(%s);", cell)
		e.line("%s = %s;", cell, fresh)
	}
	for index := range statement.Update {
		e.statementAt(&statement.Update[index])
	}
	if statement.CheckAfter {
		check()
	}
	e.indent--
	e.line("}")
	if target.broken {
		e.line("%s:;", target.label)
	}
}

// forOf emits for (const element of array). Unless the element borrow proof keeps the array
// alive in its variable, the iterator holds its own count even if that variable is reassigned.
// Its length is read again before each pass. Over a map, an iterator holds the map and keeps it
// from compacting until every way out of the loop has let go of it.
func (e *emitter) forOf(statement ir.ForOf) {
	e.line("{")
	e.indent++
	e.scopes = append(e.scopes, nil)
	iterable := e.value(statement.Iterable)
	held := e.temporary()
	overString := statement.Iterable.Type() == ir.String
	overMap := statement.MapPart != ""
	overRegex := statement.RegexIterator
	overTyped := statement.Iterable.Type().IsTypedArray()
	if overTyped {
		e.line("adamic_typed_array_iterator *%s = adamic_typed_array_iterate(%s);", held, iterable)
	} else if overMap {
		e.line("adamic_map_iterator *%s = adamic_map_iterate(%s);", held, iterable)
	} else if e.elementBorrows[e.at] {
		// The same proof lends both the element and the array. Neither owns a count here.
		e.line("%s %s = %s;", cType(statement.Iterable.Type()), held, iterable)
	} else {
		e.line("%s %s = %s;", cType(statement.Iterable.Type()), held, e.kept(iterable))
	}
	if !e.elementBorrows[e.at] {
		e.hold(held)
	}
	e.end()
	index := e.temporary()
	size := e.temporary()
	e.temporaries++
	current := &loop{label: fmt.Sprintf("adamic_continue_%d", e.temporaries)}
	entryKey, entryValue := e.temporary(), e.temporary()
	if overTyped {
		e.line("double %s;", entryValue)
		e.line("while (adamic_typed_array_iterator_next(%s, &%s)) {", held, entryValue)
	} else if overMap {
		e.line("adamic_value %s, %s;", entryKey, entryValue)
		e.line("while (adamic_map_iterator_next(%s, &%s, &%s)) {", held, entryKey, entryValue)
	} else if overRegex {
		e.line("adamic_array *%s;", entryValue)
		e.line("while ((%s = adamic_regex_iterator_step(%s)) != NULL) {", entryValue, held)
	} else if overString {
		// A code point at a time: size is its byte length, and the element is a string of it.
		e.line("for (size_t %s = 0, %s = 0; %s < %s->length; %s += %s) {", index, size, index, held, index, size)
		e.line("\t%s = adamic_string_next(%s, %s);", size, held, index)
	} else {
		e.line("for (size_t %s = 0; %s < %s->length; %s++) {", index, index, held, index)
	}
	e.indent++
	saved := e.out
	e.out = strings.Builder{}
	current.depth = len(e.scopes)
	e.linkLabels(statement.Labels, current)
	e.loops = append(e.loops, current)
	target := &breakable{depth: current.depth, label: e.temporary()}
	e.breakables = append(e.breakables, target)
	e.line("{")
	e.indent++
	e.scopes = append(e.scopes, nil)
	element := unslotted(statement.Element, fmt.Sprintf("%s->elements[%s].%s", held, index, member(statement.Element)))
	// bindEntry declares a local from the step's key or value, retained, since the body may delete
	// the entry.
	bindEntry := func(local int, slot string, of ir.Type) {
		reading := unslotted(of, slot+"."+member(of))
		if of.IsReference() {
			reading = fmt.Sprintf("(%s)%s", cType(of), reading)
		}
		e.declareLocal(local, reading, false)
	}
	if overMap {
		switch statement.MapPart {
		case "keys":
			bindEntry(statement.Local, entryKey, statement.Key)
		case "values":
			bindEntry(statement.Local, entryValue, statement.Value)
		default:
			for _, binding := range statement.Pattern {
				if binding.Field == "0" {
					bindEntry(binding.Local, entryKey, statement.Key)
				} else {
					bindEntry(binding.Local, entryValue, statement.Value)
				}
			}
		}
	} else if statement.Pattern != nil {
		// Each name reads its field of the tuple, which the array holds while the body runs.
		for _, binding := range statement.Pattern {
			bound := e.program.Locals[binding.Local]
			field := unslotted(bound.Type, fmt.Sprintf("adamic_object_field(%s, %s, &%s)->%s", element, cString(binding.Field), e.cache(), member(bound.Type)))
			if bound.Type.IsReference() {
				field = fmt.Sprintf("(%s)%s", cType(bound.Type), field)
			}
			e.declareLocal(binding.Local, field, false)
		}
	} else if overTyped {
		e.declareLocal(statement.Local, entryValue, false)
	} else if overRegex {
		e.declareLocal(statement.Local, entryValue, true)
	} else if overString {
		e.declareLocal(statement.Local, fmt.Sprintf("adamic_string_slice_bytes(%s, %s, %s)", held, index, size), true)
	} else {
		if statement.Element.IsReference() {
			element = fmt.Sprintf("(%s)%s", cType(statement.Element), element)
		}
		if e.elementBorrows[e.at] {
			e.line("%s %s = %s;", cType(statement.Element), e.localName(statement.Local), element)
		} else {
			e.declareLocal(statement.Local, element, false)
		}
	}
	for index := range statement.Body {
		e.statementAt(&statement.Body[index])
	}
	e.releaseScopes(len(e.scopes) - 1)
	e.scopes = e.scopes[:len(e.scopes)-1]
	e.indent--
	e.line("}")
	e.breakables = e.breakables[:len(e.breakables)-1]
	e.loops = e.loops[:len(e.loops)-1]
	body := e.out.String()
	e.out = saved
	e.out.WriteString(body)
	if current.continued {
		e.line("%s:;", current.label)
	}
	e.indent--
	e.line("}")
	if target.broken {
		e.line("%s:;", target.label)
	}
	e.releaseScopes(len(e.scopes) - 1)
	e.scopes = e.scopes[:len(e.scopes)-1]
	e.indent--
	e.line("}")
}

// switchStatement emits a switch as do { if / else } while (0), so a break inside a case leaves the
// switch exactly as JavaScript's does: the do is the innermost thing a C break leaves.
func (e *emitter) switchStatement(statement ir.Switch) {
	e.line("{")
	e.indent++
	e.scopes = append(e.scopes, nil)
	value := e.value(statement.Value)
	held := e.temporary()
	if statement.Value.Type().IsReference() {
		e.line("%s %s = %s;", cType(statement.Value.Type()), held, retained(value))
		e.hold(held)
	} else {
		e.line("%s %s = %s;", cType(statement.Value.Type()), held, value)
	}
	e.end()
	e.line("do {")
	e.indent++
	target := &breakable{depth: len(e.scopes), label: e.temporary()}
	e.breakables = append(e.breakables, target)
	depth := 0
	for _, matched := range statement.Cases {
		condition := e.switchTests(statement.Value.Type(), held, matched.Tests)
		e.line("if (%s) {", condition)
		e.nested(matched.Body, nil)
		e.line("} else {")
		e.indent++
		depth++
	}
	e.block(statement.Default, nil)
	for ; depth > 0; depth-- {
		e.indent--
		e.line("}")
	}
	e.breakables = e.breakables[:len(e.breakables)-1]
	e.indent--
	e.line("} while (0);")
	if target.broken {
		e.line("%s:;", target.label)
	}
	e.releaseScopes(len(e.scopes) - 1)
	e.scopes = e.scopes[:len(e.scopes)-1]
	e.indent--
	e.line("}")
}
