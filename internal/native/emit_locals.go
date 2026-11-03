// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"regexp"
)

// localName uses the owning declaration and the local's spelling.
func (e *emitter) localName(local int) string {
	return e.namedLocal(local)
}

// readyName is the flag that says a global's declaration has run.
func (e *emitter) readyName(local int) string {
	return e.localName(local) + "_ready"
}

var cIdentifier = regexp.MustCompile(`[^A-Za-z0-9_]`)

// store gives a local a value; a string's new reference is taken before the old one is let go,
// since they may be the same string.
func (e *emitter) store(local int, value string, owned bool) {
	if e.program.Locals[local].Borrowed && e.cellSlot(local) == "" {
		// Storing would release the old value, which is the caller's. Lowering never borrows a
		// parameter anything assigns, so reaching this is a compiler bug, said out loud.
		panic(fmt.Sprintf("native: a store into the borrowed parameter %s", e.program.Locals[local].Name))
	}
	name := e.localName(local)
	if slot := e.cellSlot(local); slot != "" {
		if !e.program.Locals[local].Type.IsReference() {
			e.line("%s.%s = %s;", slot, member(e.program.Locals[local].Type), slotted(e.program.Locals[local].Type, value))
			return
		}
		if e.asyncBorrows[local] {
			// The source slot owns this alias until the whole frame finishes.
			e.line("%s.reference = %s;", slot, value)
			if owned {
				e.line("adamic_release(%s);", value)
			}
			return
		}
		if e.program.Locals[local].GraphCell {
			old := e.temporary()
			e.line("void *%s = %s.reference;", old, slot)
			e.line("%s.reference = adamic_graph_hold(%s, %s);", slot, e.cellReference(local), value)
			e.dropIn(e.cellReference(local), old)
			if owned {
				e.line("adamic_release(%s);", value)
			}
			return
		}
		if !owned {
			value = retained(value)
		}
		old := e.temporary()
		e.line("void *%s = %s.reference;", old, slot)
		e.line("%s.reference = %s;", slot, value)
		if _, framed := e.asyncSlots[local]; framed {
			e.line("if (%s != NULL) adamic_release(%s);", old, old)
		} else {
			e.line("adamic_release(%s);", old)
		}
		return
	}
	if !e.program.Locals[local].Type.IsReference() {
		e.line("%s = %s;", name, value)
		return
	}
	if !owned {
		value = retained(value)
	}
	old := e.temporary()
	e.line("%s %s = %s;", cType(e.program.Locals[local].Type), old, name)
	e.line("%s = %s;", name, value)
	e.line("adamic_release(%s);", old)
}

// checkReady throws JavaScript's ReferenceError when node, a read or a write of local, touches the
// binding in its temporal dead zone: its readiness flag is clear.
func (e *emitter) checkReady(node any, local int) { e.deadZone(node, local, e.localReady(local)) }

// deadZone emits the test of ready and, when it fails, the ReferenceError Node throws, "Cannot access
// 'name' before initialization", named ReferenceError with that message alone, so String(error) reads
// as Node's and a catch takes it (#bxmdk73). A node only throws here when ir.Program's DeadZoneMayThrow
// counts it, since that is what lower's throwsOut and flow's CanThrow read: a throw they don't count is
// one no caller tests and no try lands, so emitting one is a compiler bug, said out loud.
func (e *emitter) deadZone(node any, local int, ready string) {
	if !e.program.DeadZoneMayThrow(node) {
		panic(fmt.Sprintf("native: a temporal-dead-zone test of %s on a node ir.Program's DeadZoneMayThrow doesn't count", e.program.Locals[local].Name))
	}
	if len(e.handlers) == 0 && e.function != nil && !e.function.Async && !e.function.MayThrow {
		// The throw would leave a function its callers believe can't throw, so none would test it.
		panic(fmt.Sprintf("native: a temporal-dead-zone throw of %s leaves %s, which lowering didn't mark as one that can throw", e.program.Locals[local].Name, e.function.Name))
	}
	e.line("if (!%s) {", ready)
	e.indent++
	e.throwNamed("ReferenceError", fmt.Sprintf("Cannot access '%s' before initialization", e.program.Locals[local].Name))
	e.indent--
	e.line("}")
}

func (e *emitter) localReady(local int) string {
	if binding := e.program.Locals[local].Ready; binding != 0 {
		return e.read(ir.Read{Local: binding - 1, Of: ir.Boolean})
	}
	if e.program.Locals[local].Captured && !e.program.Locals[local].Global {
		return e.cellReference(local) + "->ready"
	}
	return e.readyName(local)
}

// checkReadyRead panics when a read the readiness pass couldn't clear reaches a variable before its
// assignment: Adamic's own stop, never JavaScript's throw, so it's no DeadZoneMayThrow node.
func (e *emitter) checkReadyRead(local int, expression string) {
	message := fmt.Sprintf("read before assignment: variable '%s' in %s", e.program.Locals[local].Name, expression)
	e.line("if (!%s) {", e.localReady(local))
	e.line("\tstatic const char message[] = %s;", cArray(message))
	e.line("\tadamic_panic(message, sizeof message - 1);")
	e.line("}")
}

// read reads a local. A global may change under a later call in the same statement, so its value is
// copied (a string retained) the moment JavaScript would read it; and from inside a function it's
// checked against the temporal dead zone first.
func (e *emitter) read(read ir.Read) string {
	if read.Unset && e.program.Locals[read.Local].Global && !e.program.Locals[read.Local].Hoisted {
		e.deadZone(read, read.Local, e.readyName(read.Local)+"_declared")
	}
	name := e.localName(read.Local)
	if read.Checked {
		e.checkReady(read, read.Local)
	}
	// The shared readiness pass clears checks it proves unnecessary.
	if read.Readiness != "" {
		e.checkReadyRead(read.Local, read.Readiness)
	}
	if e.integerCounter(read.Local) {
		// Read as the double it stands for, which every value it can hold is exactly.
		return "((double)" + name + ")"
	}
	// A reference its consumer lends needs no count: nothing can run before it's used (borrow.go).
	lent := e.lendable && lendable(read.Of)
	if slot := e.cellSlot(read.Local); slot != "" {
		if read.Checked {
			e.checkReady(read, read.Local)
		}
		// A captured variable may change under a call later in the statement (a closure that
		// writes it), so, like a global, it's copied the moment JavaScript reads it.
		value := unslotted(read.Of, slot+"."+member(read.Of))
		if lent {
			e.self = true
			return e.snapshot(read.Of, fmt.Sprintf("(%s)%s", cType(read.Of), value))
		}
		if read.Of.IsReference() {
			// An uncaptured suspension slot has the same statement lifetime as
			// an ordinary local: this resume holds the frame, and no callee can
			// replace the binding. Expression writes still require a snapshot.
			if _, framed := e.asyncSlots[read.Local]; framed && !e.program.Locals[read.Local].Captured && !e.program.Locals[read.Local].ExpressionAssigned {
				return fmt.Sprintf("((%s)%s)", cType(read.Of), value)
			}
			return e.own(read.Of, fmt.Sprintf("adamic_retain((%s)%s)", cType(read.Of), value))
		}
		snapshot := e.temporary()
		e.line("%s %s = %s;", cType(read.Of), snapshot, value)
		return snapshot
	}
	if !e.program.Locals[read.Local].Global && !e.program.Locals[read.Local].ExpressionAssigned {
		return name
	}
	if lent {
		e.self = true
		return e.snapshot(read.Of, name)
	}
	if read.Of.IsReference() {
		return e.own(read.Of, fmt.Sprintf("adamic_retain(%s)", name))
	}
	snapshot := e.temporary()
	e.line("%s %s = %s;", cType(read.Of), snapshot, name)
	return snapshot
}

// declareLocal declares a local with its first value. A reference is retained unless owned says the
// value is already the local's. A captured local is declared straight into a cell.
func (e *emitter) declareLocal(local int, value string, owned bool) {
	declared := e.program.Locals[local]
	if position, ok := e.asyncSlots[local]; ok {
		if declared.IterationCell && !declared.Preallocated {
			e.replaceIterationCell(local, position, value, owned)
			return
		}
		e.store(local, value, owned)
		e.line("%s->ready = true;", e.cellReference(local))
		return
	}
	if e.integerCounter(local) {
		// A proven whole-number start within 2^53, which an integer holds exactly.
		e.line("int64_t %s = (int64_t)%s;", e.localName(local), value)
		return
	}
	if declared.Captured && declared.Preallocated {
		e.store(local, value, owned)
		e.line("%s->ready = true;", e.cellReference(local))
		return
	}
	if declared.Captured {
		e.makeCell(local, value, owned)
		return
	}
	name := e.localName(local)
	if declared.Type.IsReference() {
		if !owned {
			value = retained(value)
		}
		e.line("%s %s = %s;", cType(declared.Type), name, value)
		e.hold(name)
		return
	}
	e.line("%s %s = %s;", cType(declared.Type), name, value)
}

// makeCell declares a captured local's cell, holding value (retained unless owned).
func (e *emitter) makeCell(local int, value string, owned bool) {
	declared := e.program.Locals[local]
	if declared.EnvironmentCell {
		e.store(local, value, owned)
		e.line("%s->ready = true;", e.cellReference(local))
		return
	}
	cell := e.cellName(local)
	if declared.GraphCell && !declared.ProgramRegion {
		e.line("adamic_cell *%s = adamic_cell_new((adamic_value){.number = 0}, %t);", cell, declared.Type.IsReference())
		e.adoptGraph(cell, "sizeof *"+cell, true)
		e.hold(cell)
		e.store(local, value, owned)
		return
	}
	if declared.Type.IsReference() && !owned {
		value = retained(value)
	}
	if declared.ProgramRegion {
		e.line("adamic_cell *%s = adamic_cell_new((adamic_value){0}, %t);", cell, declared.Type.IsReference())
		e.adoptProgram(cell, "sizeof *"+cell, true)
		e.line("%s->value.%s = %s;", cell, member(declared.Type), slotted(declared.Type, value))
	} else {
		e.line("adamic_cell *%s = adamic_cell_new((adamic_value){.%s = %s}, %t);", cell, member(declared.Type), slotted(declared.Type, value), declared.Type.IsReference())
	}
	e.hold(cell)
}

func (e *emitter) cellName(local int) string {
	return e.localName(local) + "_cell"
}

// cellSlot is the adamic_value a captured local lives in, from the current function: a cell of its
// environment, or a cell it declared. It's "" for a local that isn't in a cell.
func (e *emitter) cellSlot(local int) string {
	if reference := e.cellReference(local); reference != "" {
		return reference + "->value"
	}
	return ""
}

// cellReference is the cell a captured local lives in, from the current function, or "".
func (e *emitter) cellReference(local int) string {
	if position, ok := e.asyncSlots[local]; ok {
		if e.program.Locals[local].IterationCell {
			return fmt.Sprintf("((adamic_cell *)frame->cells[%d].value.reference)", position)
		}
		return fmt.Sprintf("(&frame->cells[%d])", position)
	}
	declared := e.program.Locals[local]
	if declared.Global || !declared.Captured {
		return ""
	}
	if e.function != nil {
		for index, captured := range e.function.Environment {
			if captured == local {
				return fmt.Sprintf("self->cells[%d]", index)
			}
		}
	}
	return e.cellName(local)
}

// allocateEnvironment emits the one IR frame site; slot names borrow its storage.
func (e *emitter) allocateEnvironment(site *ir.Statement) {
	cells := (*site).(ir.AllocateEnvironment).Cells
	if len(cells) == 0 {
		return
	}
	environment := e.temporary()
	switch e.regions.environments[site] {
	case environmentStack:
		// A fixed C aggregate owns both the header and its complete slot vector.
		// The pointer keeps one slot representation across all placements.
		storage := e.temporary()
		e.line("struct { adamic_environment record; adamic_cell cells[%d]; } %s;", len(cells), storage)
		e.line("adamic_environment *%s = &%s.record;", environment, storage)
		e.line("adamic_environment_init(%s, %s.cells, %d);", environment, storage, len(cells))
		e.regions.cleanups[environment] = "adamic_environment_end(" + environment + ");"
		e.hold(environment)
	case environmentRegion:
		region := e.temporary()
		e.line("adamic_region %s = ADAMIC_REGION;", region)
		e.line("adamic_environment *%s = adamic_environment_new_in(&%s, %d);", environment, region, len(cells))
		e.regions.cleanups[region] = "adamic_region_end(&" + region + ");"
		e.hold(region)
	default:
		e.line("adamic_environment *%s = adamic_environment_new(%d);", environment, len(cells))
		graph := false
		for _, local := range cells {
			graph = graph || e.program.Locals[local].GraphCell
		}
		member := false
		for _, local := range cells {
			member = member || e.program.Locals[local].ProgramRegion
		}
		e.adoptGraph(environment, fmt.Sprintf("sizeof *%s + %d * sizeof(adamic_cell)", environment, len(cells)), graph && !member)
		e.adoptProgram(environment, fmt.Sprintf("sizeof *%s + %d * sizeof %s->cells[0]", environment, len(cells), environment), member)
		e.hold(environment)
	}
	for position, local := range cells {
		e.line("adamic_cell *%s = &%s->cells[%d];", e.cellName(local), environment, position)
		e.line("%s->references = %t;", e.cellName(local), e.program.Locals[local].Type.IsReference())
	}
}

// The frame owns the current binding; escaped closures own older bindings independently.
func (e *emitter) replaceIterationCell(local, position int, value string, owned bool) {
	declared := e.program.Locals[local]
	if declared.Type.IsReference() && !owned {
		value = retained(value)
	}
	fresh := e.temporary()
	e.line("adamic_cell *%s = adamic_cell_new((adamic_value){.%s = %s}, %t);", fresh, member(declared.Type), slotted(declared.Type, value), declared.Type.IsReference())
	e.line("adamic_release(frame->cells[%d].value.reference);", position)
	e.line("frame->cells[%d].value.reference = %s;", position, fresh)
}
