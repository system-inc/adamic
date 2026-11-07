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
func (nativeEmitter *emitter) block(statements []ir.Statement, after func()) {
	nativeEmitter.scopes = append(nativeEmitter.scopes, nil)
	for index := range statements {
		nativeEmitter.statementAt(&statements[index])
	}
	if after != nil {
		after()
	}
	nativeEmitter.releaseScopes(len(nativeEmitter.scopes) - 1)
	nativeEmitter.scopes = nativeEmitter.scopes[:len(nativeEmitter.scopes)-1]
}

// decide evaluates a condition and releases its temporaries before anything branches on it,
// returning the C expression holding the answer.
func (nativeEmitter *emitter) decide(condition ir.Expression) string {
	value := nativeEmitter.value(condition)
	if len(nativeEmitter.owned) == 0 {
		return value
	}
	name := nativeEmitter.temporary()
	nativeEmitter.line("bool %s = %s;", name, value)
	nativeEmitter.end()
	return name
}

// statementAt emits a statement where it stands in its slice, the address the reuse plan names it by.
func (nativeEmitter *emitter) statementAt(at *ir.Statement) {
	outer := nativeEmitter.at
	nativeEmitter.at = at
	if !nativeEmitter.regionStatement(at) {
		nativeEmitter.statement(*at)
	}
	nativeEmitter.at = outer
}

func (nativeEmitter *emitter) statement(statement ir.Statement) {
	// Each statement's checks share a ladder of their own (exceptions.go).
	frame := nativeEmitter.openLadder()
	defer nativeEmitter.closeLadder(frame)
	switch statement := statement.(type) {
	case ir.SourceLine:
		if nativeEmitter.debug {
			nativeEmitter.line("#line %d %s", statement.Line, cLiteral(statement.File))
		}
	case ir.Debugger:
		// No debugger is attached to native builds, so emit no instruction or trap.
	case ir.WriteLine:
		stream := "adamic_stdout"
		if statement.Stream == ir.Stderr {
			stream = "adamic_stderr"
		}
		value := nativeEmitter.value(statement.Value)
		nativeEmitter.line("adamic_write_line(%s, %s);", stream, value)
		nativeEmitter.end()
	case ir.AllocateEnvironment:
		// Emitted at function entry before captured parameters are initialized.
	case ir.Declare:
		if statement.CopyCell {
			cell := nativeEmitter.cellReference(statement.Local)
			nativeEmitter.replaceIterationCell(statement.Local, nativeEmitter.asyncSlots[statement.Local], unslotted(
				nativeEmitter.program.Locals[statement.Local].Type,
				cell+"->value."+member(nativeEmitter.program.Locals[statement.Local].Type),
			), false)
			return
		}
		if statement.Uninitialized && !nativeEmitter.program.Locals[statement.Local].Uninitialized {
			if nativeEmitter.program.Locals[statement.Local].EnvironmentCell {
				return
			}
			local := nativeEmitter.program.Locals[statement.Local]
			if local.IterationCell {
				nativeEmitter.replaceIterationCell(
					statement.Local,
					nativeEmitter.asyncSlots[statement.Local],
					zero(local.Type),
					true,
				)
				nativeEmitter.line("%s->ready = false;", nativeEmitter.cellReference(statement.Local))
				return
			}
			if local.Captured {
				nativeEmitter.makeCell(statement.Local, zero(local.Type), true)
				nativeEmitter.line("%s->ready = false;", nativeEmitter.cellReference(statement.Local))
			}
			return
		}
		if nativeEmitter.elementBorrows[nativeEmitter.at] {
			nativeEmitter.borrowElement(statement)
			return
		}
		local := nativeEmitter.program.Locals[statement.Local]
		value := zero(local.Type)
		if statement.Value != nil {
			value = nativeEmitter.value(statement.Value)
		}
		// The declaration is the statement's last write: what it owns, the variable can take.
		owned := nativeEmitter.taken(value)
		if local.Global {
			nativeEmitter.store(statement.Local, value, owned)
			nativeEmitter.line(
				"%s = %t;",
				nativeEmitter.readyName(statement.Local),
				!statement.Uninitialized && (!local.NamespaceState || statement.Value != nil),
			)
			if local.Uninitialized {
				nativeEmitter.line("%s_declared = true;", nativeEmitter.readyName(statement.Local))
			}
			nativeEmitter.initialized = append(nativeEmitter.initialized, statement.Local)
		} else {
			nativeEmitter.declareLocal(statement.Local, value, owned)
			if local.Uninitialized {
				if local.Captured {
					nativeEmitter.line("%s = %t;", nativeEmitter.localReady(statement.Local), !statement.Uninitialized)
				} else {
					nativeEmitter.line(
						"bool %s = %t; (void)%s;",
						nativeEmitter.readyName(statement.Local),
						!statement.Uninitialized,
						nativeEmitter.readyName(statement.Local),
					)
				}
			}
		}
		nativeEmitter.end()
	case ir.Assign:
		if nativeEmitter.integerCounter(statement.Local) {
			// The loop's update, counter + step or counter - step, is the only write a counter has, and
			// its step whole and invariant within 2^53 (lower/counters.go), which the cast keeps exactly.
			sum, isSum := statement.Value.(ir.Binary)
			read, isRead := sum.Left.(ir.Read)
			if !isSum ||
				(sum.Operator != ir.Add && sum.Operator != ir.Subtract) ||
				!isRead ||
				read.Local != statement.Local {
				panic(fmt.Sprintf(
					"native: compiler bug: counter %s written other than by a step",
					nativeEmitter.program.Locals[statement.Local].Name,
				))
			}
			operator := "+="
			if sum.Operator == ir.Subtract {
				operator = "-="
			}
			nativeEmitter.line(
				"%s %s (int64_t)(%s);",
				nativeEmitter.localName(statement.Local),
				operator,
				nativeEmitter.value(sum.Right),
			)
			return
		}
		if parts, appends := nativeEmitter.appendsTo(statement); appends {
			nativeEmitter.appendTo(statement.Local, parts)
			nativeEmitter.end()
			break
		}
		value := nativeEmitter.value(statement.Value)
		if nativeEmitter.program.Locals[statement.Local].Uninitialized &&
			nativeEmitter.program.Locals[statement.Local].Global &&
			!nativeEmitter.program.Locals[statement.Local].Hoisted {
			// After the value, as JavaScript does: the right side runs, then the write throws.
			nativeEmitter.deadZone(statement, statement.Local, nativeEmitter.readyName(statement.Local)+"_declared")
		}
		if statement.Checked {
			nativeEmitter.checkReady(statement, statement.Local)
		}
		// The store is the statement's last write: nothing after it can assign the variable again
		// while the statement still reads the value, so what the statement owns, the variable takes.
		nativeEmitter.store(statement.Local, value, nativeEmitter.taken(value))
		if nativeEmitter.program.Locals[statement.Local].Uninitialized {
			nativeEmitter.line("%s = %t;", nativeEmitter.localReady(statement.Local), !statement.Uninitialized)
		} else if nativeEmitter.program.Locals[statement.Local].NamespaceState {
			nativeEmitter.line("%s = true;", nativeEmitter.readyName(statement.Local))
		}
		nativeEmitter.end()
	case ir.Evaluate:
		if splice, isSplice := statement.Value.(ir.ArraySplice); isSplice {
			// What a splice removes, nothing here uses: it's let go of, and no array is made for it.
			nativeEmitter.line("adamic_array_remove(%s);", nativeEmitter.spliceArguments(splice))
			nativeEmitter.end()
			break
		}
		if call, isCall := statement.Value.(ir.Call); isCall && call.Returns == 0 {
			if call.Accessor != "" {
				nativeEmitter.accessorCall(call)
			} else {
				nativeEmitter.line("%s;", nativeEmitter.callCode(call, nativeEmitter.arguments(call)))
			}
			if nativeEmitter.program.CallMayThrow(call) {
				nativeEmitter.checkThrown()
			}
		} else {
			// The value goes, but whatever making it did stays: a push's append is its effect.
			nativeEmitter.line("(void)%s;", nativeEmitter.value(statement.Value))
		}
		nativeEmitter.end()
	case ir.Return:
		nativeEmitter.returnStatement(statement)
	case ir.SetIndex:
		array := nativeEmitter.value(statement.Array)
		index := nativeEmitter.value(statement.Index)
		value := nativeEmitter.value(statement.Value)
		if statement.Array.Type().IsTypedArray() {
			store := "adamic_typed_array_set"
			if read, isRead := statement.Index.(ir.Read); isRead && nativeEmitter.integerCounter(read.Local) {
				store, index = "adamic_typed_array_set_integer", nativeEmitter.localName(read.Local)
			}
			nativeEmitter.line("%s(%s, %s, %s);", store, array, index, value)
			nativeEmitter.end()
			break
		}
		if statement.Element.IsReference() {
			value = nativeEmitter.heldReferenceIn(array, value)
		}
		nativeEmitter.line(
			"adamic_array_set(%s, %s, (adamic_value){.%s = %s});",
			array,
			index,
			member(statement.Element),
			slotted(statement.Element, value),
		)
		nativeEmitter.end()
	case ir.SetProperty:
		if statement.Record {
			nativeEmitter.recordWrite(statement)
			nativeEmitter.end()
			break
		}
		object := nativeEmitter.value(statement.Object)
		value := nativeEmitter.value(statement.Value)
		nativeEmitter.writeProperty(statement, object, value)
		nativeEmitter.end()
	case ir.Panic:
		// The program ends here, so nothing it holds needs letting go.
		message := nativeEmitter.value(statement.Message)
		nativeEmitter.line("adamic_panic((%s)->bytes, (%s)->length);", message, message)
		nativeEmitter.owned = nil
	case ir.If:
		condition := nativeEmitter.decide(statement.Condition)
		nativeEmitter.line("if (%s) {", unwrap(condition))
		nativeEmitter.nested(statement.Then, nil)
		if len(statement.Else) > 0 {
			nativeEmitter.line("} else {")
			nativeEmitter.nested(statement.Else, nil)
		}
		nativeEmitter.line("}")
	case ir.Labeled:
		nativeEmitter.labeled(statement)
	case ir.Block:
		if nativeEmitter.counterBlock(statement) {
			break
		}
		if statement.Closes != 0 {
			if nativeEmitter.closes == nil {
				nativeEmitter.closes = map[string]bool{}
			}
			nativeEmitter.closes[nativeEmitter.localName(statement.Closes-1)] = true
		}
		nativeEmitter.line("{")
		nativeEmitter.nested(statement.Body, nil)
		nativeEmitter.line("}")
	case ir.Loop:
		nativeEmitter.loop(statement)
	case ir.ForOf:
		nativeEmitter.forOf(statement)
	case ir.Switch:
		nativeEmitter.switchStatement(statement)
	case ir.Break:
		if statement.Label != "" {
			nativeEmitter.labeledJump(statement.Label, false)
			break
		}
		target := nativeEmitter.breakables[len(nativeEmitter.breakables)-1-statement.Depth]
		nativeEmitter.finallies(nativeEmitter.innerHandlers(target.depth))
		nativeEmitter.releaseScopes(target.depth)
		if statement.Depth == 0 {
			nativeEmitter.line("break;")
		} else {
			target.broken = true
			nativeEmitter.line("goto %s;", target.label)
		}
	case ir.Continue:
		if statement.Label != "" {
			nativeEmitter.labeledJump(statement.Label, true)
			break
		}
		var current *loop
		for index := len(nativeEmitter.loops) - 1; index >= 0; index-- {
			if nativeEmitter.loops[index].sourceName == "" {
				current = nativeEmitter.loops[index]
				break
			}
		}
		if current == nil {
			panic("native: continue outside a loop")
		}
		current.continued = true
		nativeEmitter.finallies(nativeEmitter.innerHandlers(current.depth))
		nativeEmitter.releaseScopes(current.depth)
		nativeEmitter.line("goto %s;", current.label)
	case ir.Throw:
		nativeEmitter.throwStatement(statement)
	case ir.Try:
		nativeEmitter.tryStatement(statement)
	default:
		// Lowering only produces statements this switch knows. Reaching this is a compiler bug.
		panic(fmt.Sprintf("native: no C for %T", statement))
	}
}

// writeProperty writes value into object's field, both already evaluated, as object.name = value does:
// through a checked view, the write is checked and a scalar going into a union slot is boxed. holds
// are what a caller's aside owns, let go too if the write throws.
func (nativeEmitter *emitter) writeProperty(statement ir.SetProperty, object, value string, holds ...string) {
	if statement.Object.Type() == ir.Array {
		object = nativeEmitter.snapshot(ir.Object, "("+object+")->metadata")
	}
	// The object may be undefined where the checker narrowed it away and a call since put it back
	// (ir.Defined): JavaScript throws at the write, after the value, and so does this.
	nativeEmitter.line("if (%s == NULL) {", object)
	nativeEmitter.line(
		"\tstatic const char message[] = %s;",
		cArray("TypeError: Cannot set properties of undefined (setting '"+statement.Name+"')"),
	)
	nativeEmitter.line("\tadamic_panic(message, sizeof message - 1);")
	nativeEmitter.line("}")
	nativeEmitter.line("adamic_object_check_data_write(%s, %s);", object, cString(statement.Name))
	// A frozen object leaves Node's TypeError pending, tested before the value is kept, so the value
	// is still the statement's to let go (ir.Program's FrozenWriteMayThrow).
	if nativeEmitter.program.FrozenWriteMayThrow(statement) {
		nativeEmitter.checkThrown(holds...)
	}
	records := nativeEmitter.hasRecordStorage()
	keptValue := value
	if statement.Value.Type().IsReference() {
		keptValue = nativeEmitter.keptIn(object, value)
	}
	if records {
		// A record reaches a fixed object's write here; under a name some define hides, the write is
		// checked and throws on a read-only key, as recordSet's are (ir.Program's RecordWriteMayThrow).
		nativeEmitter.line("if (adamic_record_is(%s)) {", object)
		boxed := keptValue
		if statement.Value.Type() == ir.Number {
			boxed = fmt.Sprintf("adamic_box_number(%s)", value)
		}
		if statement.Value.Type() == ir.Boolean {
			boxed = fmt.Sprintf("(%s ? &adamic_box_true : &adamic_box_false)", value)
		}
		if statement.Value.Type() == ir.MaybeNumber {
			boxed = fmt.Sprintf("((%s).present ? adamic_box_number((%s).number) : NULL)", value, value)
		}
		if statement.Value.Type() == ir.MaybeBoolean {
			boxed = fmt.Sprintf(
				"((%s).present ? ((%s).boolean ? &adamic_box_true : &adamic_box_false) : NULL)",
				value,
				value,
			)
		}
		key := nativeEmitter.recordKey(statement.Name)
		nativeEmitter.recordSet(
			object,
			key,
			fmt.Sprintf("(adamic_value){.reference = %s}", boxed),
			key,
			nativeEmitter.program.RecordWriteMayThrow(statement),
		)
		nativeEmitter.line("} else {")
	}
	slot := nativeEmitter.temporary()
	cache := ""
	if nativeEmitter.fieldTypesNeeded() {
		cache = nativeEmitter.cache()
	}
	if nativeEmitter.program.CheckedField(statement.ViewReceiverTypeID, statement.Name) {
		expected := statement.DeclaredType
		if expected == "" {
			expected = map[ir.Type]string{
				ir.Number:  "number",
				ir.Boolean: "boolean",
				ir.String:  "string",
				ir.Object:  "object",
				ir.Array:   "array",
				ir.Map:     "Map",
				ir.Closure: "function",
				ir.Union:   "unknown",
			}[statement.Value.Type()]
			if expected == "" {
				expected = "unknown"
			}
		}
		nativeEmitter.line(
			"adamic_object_view_write(%s, %s, &%s, %d, %s, %s);",
			object,
			cString(statement.Name),
			cache,
			fieldRepresentation(statement.Value),
			cString(expected),
			cString("<write>."+statement.Name),
		)
	}
	if nativeEmitter.fieldTypesNeeded() {
		nativeEmitter.line(
			"adamic_value *%s = adamic_object_write_field(%s, %s, &%s);",
			slot,
			object,
			cString(statement.Name),
			cache,
		)
	} else {
		nativeEmitter.line(
			"adamic_value *%s = %s;",
			slot,
			nativeEmitter.writeFieldSlot(object, statement.Name, statement.Class),
		)
	}
	converted := nativeEmitter.program.CheckedField(statement.ViewReceiverTypeID, statement.Name) &&
		statement.Value.Type() <= ir.Boolean
	if converted {
		nativeEmitter.line(
			"if (adamic_object_field_types(%s)[adamic_slot_index(%s, %s)] == 10) {",
			object,
			object,
			slot,
		)
		boxed := fmt.Sprintf("adamic_box_number(%s)", value)
		if statement.Value.Type() == ir.Boolean {
			boxed = fmt.Sprintf("(%s ? &adamic_box_true : &adamic_box_false)", value)
		}
		nativeEmitter.line("adamic_heap *view_new_value = (adamic_heap *)%s;", boxed)
		nativeEmitter.line("adamic_release(%s->reference);", slot)
		nativeEmitter.line("%s->reference = view_new_value;", slot)
		nativeEmitter.line("} else {")
	}
	if statement.Value.Type().IsReference() {
		// The new reference is taken before the old is let go: they may be the same.
		old := nativeEmitter.temporary()
		nativeEmitter.line("void *%s = %s->reference;", old, slot)
		nativeEmitter.line("%s->reference = %s;", slot, keptValue)
		if len(nativeEmitter.program.GraphTypes) != 0 {
			nativeEmitter.dropIn(object, old)
		} else {
			nativeEmitter.line("if (%s != NULL) adamic_release(%s);", old, old)
		}
	} else {
		nativeEmitter.line("%s->%s = %s;", slot, member(statement.Value.Type()), slotted(statement.Value.Type(), value))
	}
	if nativeEmitter.fieldTypesNeeded() {
		nativeEmitter.line(
			"adamic_object_field_types(%s)[adamic_slot_index(%s, %s)] = %d;",
			object,
			object,
			slot,
			fieldRepresentation(statement.Value),
		)
	}
	if converted {
		nativeEmitter.line("}")
	}
	if nativeEmitter.fieldReadinessNeeded(statement.Name) || nativeEmitter.constructionNeeded() {
		nativeEmitter.line(
			"adamic_object_set_initialized(%s, %s, %t);",
			object,
			cString(statement.Name),
			!statement.Uninitialized || statement.Unset,
		)
	}
	if records {
		nativeEmitter.line("}")
	}
}

// nested emits statements one level in, in a scope of their own.
func (nativeEmitter *emitter) nested(statements []ir.Statement, after func()) {
	nativeEmitter.indent++
	nativeEmitter.block(statements, after)
	nativeEmitter.indent--
}

// loop emits every loop the same way: for (;;), the condition checked as an if that breaks, the
// body in its own scope, then a continue label and the update. A continue releases the body's
// locals and jumps to the label, so it runs the update and the check, as JavaScript's does.
func (nativeEmitter *emitter) loop(statement ir.Loop) {
	current := &loop{label: nativeEmitter.temporary() + "_continue"}
	check := func() {
		condition := nativeEmitter.decide(statement.Condition)
		nativeEmitter.line("if (!(%s)) {", unwrap(condition))
		nativeEmitter.line("\tbreak;")
		nativeEmitter.line("}")
	}
	nativeEmitter.line("for (;;) {")
	nativeEmitter.indent++
	if !statement.CheckAfter {
		check()
	}
	// The body is emitted aside first, since only then is it known whether anything continues.
	saved := nativeEmitter.out
	nativeEmitter.out = strings.Builder{}
	current.depth = len(nativeEmitter.scopes)
	nativeEmitter.linkLabels(statement.Labels, current)
	nativeEmitter.loops = append(nativeEmitter.loops, current)
	target := &breakable{depth: current.depth, label: nativeEmitter.temporary()}
	nativeEmitter.breakables = append(nativeEmitter.breakables, target)
	nativeEmitter.line("{")
	nativeEmitter.nested(statement.Body, nil)
	nativeEmitter.line("}")
	nativeEmitter.breakables = nativeEmitter.breakables[:len(nativeEmitter.breakables)-1]
	nativeEmitter.loops = nativeEmitter.loops[:len(nativeEmitter.loops)-1]
	body := nativeEmitter.out.String()
	nativeEmitter.out = saved
	nativeEmitter.out.WriteString(body)
	if current.continued {
		nativeEmitter.line("%s:;", current.label)
	}
	// Each iteration's let bindings are its own (ECMA-262, CreatePerIterationEnvironment): a captured
	// one gets a fresh cell holding the current value before the update runs, so a closure made in
	// this iteration keeps this iteration's.
	for _, local := range statement.PerIteration {
		if !nativeEmitter.program.Locals[local].Captured {
			continue
		}
		cell := nativeEmitter.cellName(local)
		fresh := nativeEmitter.temporary()
		nativeEmitter.line("adamic_cell *%s = adamic_cell_new(%s->value, %s->references);", fresh, cell, cell)
		nativeEmitter.line("%s->ready = %s->ready;", fresh, cell)
		nativeEmitter.line("if (%s->references) {", fresh)
		nativeEmitter.line("\tadamic_retain(%s->value.reference);", fresh)
		nativeEmitter.line("}")
		nativeEmitter.adoptGraph(fresh, "sizeof *"+fresh, nativeEmitter.program.Locals[local].GraphCell)
		nativeEmitter.line("adamic_release(%s);", cell)
		nativeEmitter.line("%s = %s;", cell, fresh)
	}
	for index := range statement.Update {
		nativeEmitter.statementAt(&statement.Update[index])
	}
	if statement.CheckAfter {
		check()
	}
	nativeEmitter.indent--
	nativeEmitter.line("}")
	if target.broken {
		nativeEmitter.line("%s:;", target.label)
	}
}

// forOf emits for (const element of array). Unless the element borrow proof keeps the array
// alive in its variable, the iterator holds its own count even if that variable is reassigned.
// Its length is read again before each pass. Over a map, an iterator holds the map and keeps it
// from compacting until every way out of the loop has let go of it.
func (nativeEmitter *emitter) forOf(statement ir.ForOf) {
	nativeEmitter.line("{")
	nativeEmitter.indent++
	nativeEmitter.scopes = append(nativeEmitter.scopes, nil)
	iterable := nativeEmitter.value(statement.Iterable)
	held := nativeEmitter.temporary()
	overString := statement.Iterable.Type() == ir.String
	overMap := statement.MapPart != ""
	overRegex := statement.RegexIterator
	overTyped := statement.Iterable.Type().IsTypedArray()
	if overTyped {
		// Typed-array length is fixed. Hold the original view even if the body
		// reassigns its source, and read live storage through the integer helper.
		nativeEmitter.line("adamic_typed_array *%s = %s;", held, nativeEmitter.kept(iterable))
	} else if overMap {
		nativeEmitter.line("adamic_map_iterator *%s = adamic_map_iterate(%s);", held, iterable)
	} else if nativeEmitter.elementBorrows[nativeEmitter.at] {
		// The same proof lends both the element and the array. Neither owns a count here.
		nativeEmitter.line("%s %s = %s;", cType(statement.Iterable.Type()), held, iterable)
	} else {
		nativeEmitter.line("%s %s = %s;", cType(statement.Iterable.Type()), held, nativeEmitter.kept(iterable))
	}
	if overTyped || !nativeEmitter.elementBorrows[nativeEmitter.at] {
		nativeEmitter.hold(held)
	}
	nativeEmitter.end()
	index := nativeEmitter.temporary()
	size := nativeEmitter.temporary()
	current := &loop{label: nativeEmitter.temporary() + "_continue"}
	entryKey, entryValue := nativeEmitter.temporary(), nativeEmitter.temporary()
	if overTyped {
		nativeEmitter.line("for (size_t %s = 0; %s < %s->length; %s++) {", index, index, held, index)
	} else if overMap {
		nativeEmitter.line("adamic_value %s, %s;", entryKey, entryValue)
		nativeEmitter.line("while (adamic_map_iterator_next(%s, &%s, &%s)) {", held, entryKey, entryValue)
	} else if overRegex {
		nativeEmitter.line("adamic_array *%s;", entryValue)
		nativeEmitter.line("while ((%s = adamic_regex_iterator_step(%s)) != NULL) {", entryValue, held)
	} else if overString {
		// A code point at a time: size is its byte length, and the element is a string of it.
		nativeEmitter.line(
			"for (size_t %s = 0, %s = 0; %s < %s->length; %s += %s) {",
			index,
			size,
			index,
			held,
			index,
			size,
		)
		nativeEmitter.line("\t%s = adamic_string_next(%s, %s);", size, held, index)
	} else {
		nativeEmitter.line("for (size_t %s = 0; %s < %s->length; %s++) {", index, index, held, index)
	}
	nativeEmitter.indent++
	saved := nativeEmitter.out
	nativeEmitter.out = strings.Builder{}
	current.depth = len(nativeEmitter.scopes)
	nativeEmitter.linkLabels(statement.Labels, current)
	nativeEmitter.loops = append(nativeEmitter.loops, current)
	target := &breakable{depth: current.depth, label: nativeEmitter.temporary()}
	nativeEmitter.breakables = append(nativeEmitter.breakables, target)
	nativeEmitter.line("{")
	nativeEmitter.indent++
	nativeEmitter.scopes = append(nativeEmitter.scopes, nil)
	element := unslotted(statement.Element, fmt.Sprintf("%s->elements[%s].%s", held, index, member(statement.Element)))
	// bindEntry declares a local from the step's key or value, retained, since the body may delete
	// the entry.
	bindEntry := func(local int, slot string, of ir.Type) {
		reading := unslotted(of, slot+"."+member(of))
		if of.IsReference() {
			reading = fmt.Sprintf("(%s)%s", cType(of), reading)
		}
		nativeEmitter.declareLocal(local, reading, false)
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
			bound := nativeEmitter.program.Locals[binding.Local]
			field := unslotted(bound.Type, fmt.Sprintf(
				"adamic_object_field(%s, %s, &%s)->%s",
				element,
				cString(binding.Field),
				nativeEmitter.cache(),
				member(bound.Type),
			))
			if bound.Type.IsReference() {
				field = fmt.Sprintf("(%s)%s", cType(bound.Type), field)
			}
			nativeEmitter.declareLocal(binding.Local, field, false)
		}
	} else if overTyped {
		nativeEmitter.declareLocal(
			statement.Local,
			fmt.Sprintf("adamic_typed_array_get_integer(%s, (int64_t)%s).number", held, index),
			false,
		)
	} else if overRegex {
		nativeEmitter.declareLocal(statement.Local, entryValue, true)
	} else if overString {
		nativeEmitter.declareLocal(
			statement.Local,
			fmt.Sprintf("adamic_string_slice_bytes(%s, %s, %s)", held, index, size),
			true,
		)
	} else {
		if statement.Element.IsReference() {
			element = fmt.Sprintf("(%s)%s", cType(statement.Element), element)
		}
		if nativeEmitter.elementBorrows[nativeEmitter.at] {
			nativeEmitter.line(
				"%s %s = %s;",
				cType(statement.Element),
				nativeEmitter.localName(statement.Local),
				element,
			)
		} else {
			nativeEmitter.declareLocal(statement.Local, element, false)
		}
	}
	for index := range statement.Body {
		nativeEmitter.statementAt(&statement.Body[index])
	}
	nativeEmitter.releaseScopes(len(nativeEmitter.scopes) - 1)
	nativeEmitter.scopes = nativeEmitter.scopes[:len(nativeEmitter.scopes)-1]
	nativeEmitter.indent--
	nativeEmitter.line("}")
	nativeEmitter.breakables = nativeEmitter.breakables[:len(nativeEmitter.breakables)-1]
	nativeEmitter.loops = nativeEmitter.loops[:len(nativeEmitter.loops)-1]
	body := nativeEmitter.out.String()
	nativeEmitter.out = saved
	nativeEmitter.out.WriteString(body)
	if current.continued {
		nativeEmitter.line("%s:;", current.label)
	}
	nativeEmitter.indent--
	nativeEmitter.line("}")
	if target.broken {
		nativeEmitter.line("%s:;", target.label)
	}
	nativeEmitter.releaseScopes(len(nativeEmitter.scopes) - 1)
	nativeEmitter.scopes = nativeEmitter.scopes[:len(nativeEmitter.scopes)-1]
	nativeEmitter.indent--
	nativeEmitter.line("}")
}

// switchStatement emits a switch as do { if / else } while (0), so a break inside a case leaves the
// switch exactly as JavaScript's does: the do is the innermost thing a C break leaves.
func (nativeEmitter *emitter) switchStatement(statement ir.Switch) {
	nativeEmitter.line("{")
	nativeEmitter.indent++
	nativeEmitter.scopes = append(nativeEmitter.scopes, nil)
	value := nativeEmitter.value(statement.Value)
	held := nativeEmitter.temporary()
	if statement.Value.Type().IsReference() {
		nativeEmitter.line("%s %s = %s;", cType(statement.Value.Type()), held, retained(value))
		nativeEmitter.hold(held)
	} else {
		nativeEmitter.line("%s %s = %s;", cType(statement.Value.Type()), held, value)
	}
	nativeEmitter.end()
	nativeEmitter.line("do {")
	nativeEmitter.indent++
	target := &breakable{depth: len(nativeEmitter.scopes), label: nativeEmitter.temporary()}
	nativeEmitter.breakables = append(nativeEmitter.breakables, target)
	depth := 0
	for _, matched := range statement.Cases {
		condition := nativeEmitter.switchTests(statement.Value.Type(), held, matched.Tests)
		nativeEmitter.line("if (%s) {", condition)
		nativeEmitter.nested(matched.Body, nil)
		nativeEmitter.line("} else {")
		nativeEmitter.indent++
		depth++
	}
	nativeEmitter.block(statement.Default, nil)
	for ; depth > 0; depth-- {
		nativeEmitter.indent--
		nativeEmitter.line("}")
	}
	nativeEmitter.breakables = nativeEmitter.breakables[:len(nativeEmitter.breakables)-1]
	nativeEmitter.indent--
	nativeEmitter.line("} while (0);")
	if target.broken {
		nativeEmitter.line("%s:;", target.label)
	}
	nativeEmitter.releaseScopes(len(nativeEmitter.scopes) - 1)
	nativeEmitter.scopes = nativeEmitter.scopes[:len(nativeEmitter.scopes)-1]
	nativeEmitter.indent--
	nativeEmitter.line("}")
}
