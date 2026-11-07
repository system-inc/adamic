// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"regexp"
)

// localName uses the owning declaration and the local's spelling.
func (nativeEmitter *emitter) localName(local int) string {
	return nativeEmitter.namedLocal(local)
}

// readyName is the flag that says a global's declaration has run.
func (nativeEmitter *emitter) readyName(local int) string {
	return nativeEmitter.localName(local) + "_ready"
}

var cIdentifier = regexp.MustCompile(`[^A-Za-z0-9_]`)

// store gives a local a value; a string's new reference is taken before the old one is let go,
// since they may be the same string.
func (nativeEmitter *emitter) store(local int, value string, owned bool) {
	if nativeEmitter.program.Locals[local].Borrowed && nativeEmitter.cellSlot(local) == "" {
		// Storing would release the old value, which is the caller's. Lowering never borrows a
		// parameter anything assigns, so reaching this is a compiler bug, said out loud.
		panic(fmt.Sprintf("native: a store into the borrowed parameter %s", nativeEmitter.program.Locals[local].Name))
	}
	name := nativeEmitter.localName(local)
	if slot := nativeEmitter.cellSlot(local); slot != "" {
		if !nativeEmitter.program.Locals[local].Type.IsReference() {
			nativeEmitter.line(
				"%s.%s = %s;",
				slot,
				member(nativeEmitter.program.Locals[local].Type),
				slotted(nativeEmitter.program.Locals[local].Type, value),
			)
			return
		}
		if nativeEmitter.asyncBorrows[local] {
			// The source slot owns this alias until the whole frame finishes.
			nativeEmitter.line("%s.reference = %s;", slot, value)
			if owned {
				nativeEmitter.line("adamic_release(%s);", value)
			}
			return
		}
		if nativeEmitter.program.Locals[local].GraphCell {
			old := nativeEmitter.temporary()
			nativeEmitter.line("void *%s = %s.reference;", old, slot)
			nativeEmitter.line(
				"%s.reference = adamic_graph_hold(%s, %s);",
				slot,
				nativeEmitter.cellReference(local),
				value,
			)
			nativeEmitter.dropIn(nativeEmitter.cellReference(local), old)
			if owned {
				nativeEmitter.line("adamic_release(%s);", value)
			}
			return
		}
		if !owned {
			value = retained(value)
		}
		old := nativeEmitter.temporary()
		nativeEmitter.line("void *%s = %s.reference;", old, slot)
		nativeEmitter.line("%s.reference = %s;", slot, value)
		if _, framed := nativeEmitter.asyncSlots[local]; framed {
			nativeEmitter.line("if (%s != NULL) adamic_release(%s);", old, old)
		} else {
			nativeEmitter.line("adamic_release(%s);", old)
		}
		return
	}
	if !nativeEmitter.program.Locals[local].Type.IsReference() {
		nativeEmitter.line("%s = %s;", name, value)
		return
	}
	if !owned {
		value = retained(value)
	}
	old := nativeEmitter.temporary()
	nativeEmitter.line("%s %s = %s;", cType(nativeEmitter.program.Locals[local].Type), old, name)
	nativeEmitter.line("%s = %s;", name, value)
	nativeEmitter.line("adamic_release(%s);", old)
}

// checkReady throws JavaScript's ReferenceError when node, a read or a write of local, touches the
// binding in its temporal dead zone: its readiness flag is clear.
func (nativeEmitter *emitter) checkReady(node any, local int) {
	nativeEmitter.deadZone(node, local, nativeEmitter.localReady(local))
}

// deadZone emits the test of ready and, when it fails, the ReferenceError Node throws, "Cannot access
// 'name' before initialization", named ReferenceError with that message alone, so String(error) reads
// as Node's and a catch takes it (#bxmdk73). A node only throws here when ir.Program's DeadZoneMayThrow
// counts it, since that is what lower's throwsOut and flow's CanThrow read: a throw they don't count is
// one no caller tests and no try lands, so emitting one is a compiler bug, said out loud.
func (nativeEmitter *emitter) deadZone(node any, local int, ready string) {
	if !nativeEmitter.program.DeadZoneMayThrow(node) {
		panic(fmt.Sprintf(
			"native: a temporal-dead-zone test of %s on a node ir.Program's DeadZoneMayThrow doesn't count",
			nativeEmitter.program.Locals[local].Name,
		))
	}
	if len(nativeEmitter.handlers) == 0 &&
		nativeEmitter.function != nil &&
		!nativeEmitter.function.Suspends() &&
		!nativeEmitter.function.MayThrow {
		// The throw would leave a function its callers believe can't throw, so none would test it.
		panic(fmt.Sprintf(
			"native: a temporal-dead-zone throw of %s leaves %s, which lowering didn't mark as one that can throw",
			nativeEmitter.program.Locals[local].Name,
			nativeEmitter.function.Name,
		))
	}
	nativeEmitter.line("if (!%s) {", ready)
	nativeEmitter.indent++
	nativeEmitter.throwNamed(
		"ReferenceError",
		fmt.Sprintf("Cannot access '%s' before initialization", nativeEmitter.program.Locals[local].Name),
	)
	nativeEmitter.indent--
	nativeEmitter.line("}")
}

func (nativeEmitter *emitter) localReady(local int) string {
	if binding := nativeEmitter.program.Locals[local].Ready; binding != 0 {
		return nativeEmitter.read(ir.Read{Local: binding - 1, Of: ir.Boolean})
	}
	if nativeEmitter.program.Locals[local].Captured && !nativeEmitter.program.Locals[local].Global {
		return nativeEmitter.cellReference(local) + "->ready"
	}
	return nativeEmitter.readyName(local)
}

// checkReadyRead panics when a read the readiness pass couldn't clear reaches a variable before its
// assignment: Adamic's own stop, never JavaScript's throw, so it's no DeadZoneMayThrow node.
func (nativeEmitter *emitter) checkReadyRead(local int, expression string) {
	message := fmt.Sprintf(
		"read before assignment: variable '%s' in %s",
		nativeEmitter.program.Locals[local].Name,
		expression,
	)
	nativeEmitter.line("if (!%s) {", nativeEmitter.localReady(local))
	nativeEmitter.line("\tstatic const char message[] = %s;", cArray(message))
	nativeEmitter.line("\tadamic_panic(message, sizeof message - 1);")
	nativeEmitter.line("}")
}

// read reads a local. A global may change under a later call in the same statement, so its value is
// copied (a string retained) the moment JavaScript would read it; and from inside a function it's
// checked against the temporal dead zone first.
func (nativeEmitter *emitter) read(read ir.Read) string {
	if read.Unset &&
		nativeEmitter.program.Locals[read.Local].Global &&
		!nativeEmitter.program.Locals[read.Local].Hoisted {
		nativeEmitter.deadZone(read, read.Local, nativeEmitter.readyName(read.Local)+"_declared")
	}
	name := nativeEmitter.localName(read.Local)
	if read.Checked {
		nativeEmitter.checkReady(read, read.Local)
	}
	// The shared readiness pass clears checks it proves unnecessary.
	if read.Readiness != "" {
		nativeEmitter.checkReadyRead(read.Local, read.Readiness)
	}
	if nativeEmitter.integerCounter(read.Local) {
		// Read as the double it stands for, which every value it can hold is exactly.
		return "((double)" + name + ")"
	}
	// A reference its consumer lends needs no count: nothing can run before it's used (borrow.go).
	lent := nativeEmitter.lendable && lendable(read.Of)
	if slot := nativeEmitter.cellSlot(read.Local); slot != "" {
		if read.Checked {
			nativeEmitter.checkReady(read, read.Local)
		}
		// A captured variable may change under a call later in the statement (a closure that
		// writes it), so, like a global, it's copied the moment JavaScript reads it.
		value := unslotted(read.Of, slot+"."+member(read.Of))
		if lent {
			nativeEmitter.self = true
			return nativeEmitter.snapshot(read.Of, fmt.Sprintf("(%s)%s", cType(read.Of), value))
		}
		if read.Of.IsReference() {
			// An uncaptured suspension slot has the same statement lifetime as
			// an ordinary local: this resume holds the frame, and no callee can
			// replace the binding. Expression writes still require a snapshot.
			if _, framed := nativeEmitter.asyncSlots[read.Local]; framed &&
				!nativeEmitter.program.Locals[read.Local].Captured &&
				!nativeEmitter.program.Locals[read.Local].ExpressionAssigned {
				return fmt.Sprintf("((%s)%s)", cType(read.Of), value)
			}
			return nativeEmitter.own(read.Of, fmt.Sprintf("adamic_retain((%s)%s)", cType(read.Of), value))
		}
		snapshot := nativeEmitter.temporary()
		nativeEmitter.line("%s %s = %s;", cType(read.Of), snapshot, value)
		return snapshot
	}
	if !nativeEmitter.program.Locals[read.Local].Global &&
		!nativeEmitter.program.Locals[read.Local].ExpressionAssigned {
		return name
	}
	if lent {
		nativeEmitter.self = true
		return nativeEmitter.snapshot(read.Of, name)
	}
	if read.Of.IsReference() {
		return nativeEmitter.own(read.Of, fmt.Sprintf("adamic_retain(%s)", name))
	}
	snapshot := nativeEmitter.temporary()
	nativeEmitter.line("%s %s = %s;", cType(read.Of), snapshot, name)
	return snapshot
}

// declareLocal declares a local with its first value. A reference is retained unless owned says the
// value is already the local's. A captured local is declared straight into a cell.
func (nativeEmitter *emitter) declareLocal(local int, value string, owned bool) {
	declared := nativeEmitter.program.Locals[local]
	if position, ok := nativeEmitter.asyncSlots[local]; ok {
		if declared.IterationCell && !declared.Preallocated {
			nativeEmitter.replaceIterationCell(local, position, value, owned)
			return
		}
		nativeEmitter.store(local, value, owned)
		nativeEmitter.line("%s->ready = true;", nativeEmitter.cellReference(local))
		return
	}
	if nativeEmitter.integerCounter(local) {
		// A proven whole-number start within 2^53, which an integer holds exactly.
		nativeEmitter.line("int64_t %s = (int64_t)%s;", nativeEmitter.localName(local), value)
		return
	}
	if declared.Captured && declared.Preallocated {
		nativeEmitter.store(local, value, owned)
		nativeEmitter.line("%s->ready = true;", nativeEmitter.cellReference(local))
		return
	}
	if declared.Captured {
		nativeEmitter.makeCell(local, value, owned)
		return
	}
	name := nativeEmitter.localName(local)
	if declared.Type.IsReference() {
		if !owned {
			value = retained(value)
		}
		nativeEmitter.line("%s %s = %s;", cType(declared.Type), name, value)
		nativeEmitter.hold(name)
		return
	}
	nativeEmitter.line("%s %s = %s;", cType(declared.Type), name, value)
}

// makeCell declares a captured local's cell, holding value (retained unless owned).
func (nativeEmitter *emitter) makeCell(local int, value string, owned bool) {
	declared := nativeEmitter.program.Locals[local]
	if declared.EnvironmentCell {
		nativeEmitter.store(local, value, owned)
		nativeEmitter.line("%s->ready = true;", nativeEmitter.cellReference(local))
		return
	}
	cell := nativeEmitter.cellName(local)
	if declared.GraphCell && !declared.ProgramRegion {
		nativeEmitter.line(
			"adamic_cell *%s = adamic_cell_new((adamic_value){.number = 0}, %t);",
			cell,
			declared.Type.IsReference(),
		)
		nativeEmitter.adoptGraph(cell, "sizeof *"+cell, true)
		nativeEmitter.hold(cell)
		nativeEmitter.store(local, value, owned)
		return
	}
	if declared.Type.IsReference() && !owned {
		value = retained(value)
	}
	if declared.ProgramRegion {
		nativeEmitter.line(
			"adamic_cell *%s = adamic_cell_new((adamic_value){0}, %t);",
			cell,
			declared.Type.IsReference(),
		)
		nativeEmitter.adoptProgram(cell, "sizeof *"+cell, true)
		nativeEmitter.line("%s->value.%s = %s;", cell, member(declared.Type), slotted(declared.Type, value))
	} else {
		nativeEmitter.line(
			"adamic_cell *%s = adamic_cell_new((adamic_value){.%s = %s}, %t);",
			cell,
			member(declared.Type),
			slotted(declared.Type, value),
			declared.Type.IsReference(),
		)
	}
	nativeEmitter.hold(cell)
}

func (nativeEmitter *emitter) cellName(local int) string {
	return nativeEmitter.localName(local) + "_cell"
}

// cellSlot is the adamic_value a captured local lives in, from the current function: a cell of its
// environment, or a cell it declared. It's "" for a local that isn't in a cell.
func (nativeEmitter *emitter) cellSlot(local int) string {
	if reference := nativeEmitter.cellReference(local); reference != "" {
		return reference + "->value"
	}
	return ""
}

// cellReference is the cell a captured local lives in, from the current function, or "".
func (nativeEmitter *emitter) cellReference(local int) string {
	if position, ok := nativeEmitter.asyncSlots[local]; ok {
		if nativeEmitter.program.Locals[local].IterationCell {
			return fmt.Sprintf("((adamic_cell *)frame->cells[%d].value.reference)", position)
		}
		return fmt.Sprintf("(&frame->cells[%d])", position)
	}
	declared := nativeEmitter.program.Locals[local]
	if declared.Global || !declared.Captured {
		return ""
	}
	if nativeEmitter.function != nil {
		for index, captured := range nativeEmitter.function.Environment {
			if captured == local {
				return fmt.Sprintf("self->cells[%d]", index)
			}
		}
	}
	return nativeEmitter.cellName(local)
}

// allocateEnvironment emits the one IR frame site; slot names borrow its storage.
func (nativeEmitter *emitter) allocateEnvironment(site *ir.Statement) {
	cells := (*site).(ir.AllocateEnvironment).Cells
	if len(cells) == 0 {
		return
	}
	environment := nativeEmitter.temporary()
	switch nativeEmitter.regions.environments[site] {
	case environmentStack:
		// A fixed C aggregate owns both the header and its complete slot vector.
		// The pointer keeps one slot representation across all placements.
		storage := nativeEmitter.temporary()
		nativeEmitter.line("struct { adamic_environment record; adamic_cell cells[%d]; } %s;", len(cells), storage)
		nativeEmitter.line("adamic_environment *%s = &%s.record;", environment, storage)
		nativeEmitter.line("adamic_environment_init(%s, %s.cells, %d);", environment, storage, len(cells))
		nativeEmitter.regions.cleanups[environment] = "adamic_environment_end(" + environment + ");"
		nativeEmitter.hold(environment)
	case environmentRegion:
		region := nativeEmitter.temporary()
		nativeEmitter.line("adamic_region %s = ADAMIC_REGION;", region)
		nativeEmitter.line(
			"adamic_environment *%s = adamic_environment_new_in(&%s, %d);",
			environment,
			region,
			len(cells),
		)
		nativeEmitter.regions.cleanups[region] = "adamic_region_end(&" + region + ");"
		nativeEmitter.hold(region)
	default:
		nativeEmitter.line("adamic_environment *%s = adamic_environment_new(%d);", environment, len(cells))
		graph := false
		for _, local := range cells {
			graph = graph || nativeEmitter.program.Locals[local].GraphCell
		}
		member := false
		for _, local := range cells {
			member = member || nativeEmitter.program.Locals[local].ProgramRegion
		}
		nativeEmitter.adoptGraph(
			environment,
			fmt.Sprintf("sizeof *%s + %d * sizeof(adamic_cell)", environment, len(cells)),
			graph && !member,
		)
		nativeEmitter.adoptProgram(
			environment,
			fmt.Sprintf("sizeof *%s + %d * sizeof %s->cells[0]", environment, len(cells), environment),
			member,
		)
		nativeEmitter.hold(environment)
	}
	for position, local := range cells {
		nativeEmitter.line("adamic_cell *%s = &%s->cells[%d];", nativeEmitter.cellName(local), environment, position)
		nativeEmitter.line(
			"%s->references = %t;",
			nativeEmitter.cellName(local),
			nativeEmitter.program.Locals[local].Type.IsReference(),
		)
	}
}

// The frame owns the current binding; escaped closures own older bindings independently.
func (nativeEmitter *emitter) replaceIterationCell(local, position int, value string, owned bool) {
	declared := nativeEmitter.program.Locals[local]
	if declared.Type.IsReference() && !owned {
		value = retained(value)
	}
	fresh := nativeEmitter.temporary()
	nativeEmitter.line(
		"adamic_cell *%s = adamic_cell_new((adamic_value){.%s = %s}, %t);",
		fresh,
		member(declared.Type),
		slotted(declared.Type, value),
		declared.Type.IsReference(),
	)
	nativeEmitter.line("adamic_release(frame->cells[%d].value.reference);", position)
	nativeEmitter.line("frame->cells[%d].value.reference = %s;", position, fresh)
}
