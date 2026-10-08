// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"regexp"
)

// localName is a local's C name: its index, which is unique, and its name as written, for reading.
func (e *emitter) localName(local int) string {
	prefix := "adamic_local"
	if e.program.Locals[local].Global {
		prefix = "adamic_global"
	}
	return fmt.Sprintf("%s_%d_%s", prefix, local, cIdentifier.ReplaceAllString(e.program.Locals[local].Name, ""))
}

// readyName is the flag that says a global's declaration has run.
func readyName(local int) string {
	return fmt.Sprintf("adamic_ready_%d", local)
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
		e.line("adamic_release(%s);", old)
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

// checkReady panics as JavaScript throws when a global is touched before its declaration has run.
func (e *emitter) checkReady(local int) { e.checkReadyRead(local, "") }

func (e *emitter) localReady(local int) string {
	if e.program.Locals[local].Captured && !e.program.Locals[local].Global {
		return e.cellReference(local) + "->ready"
	}
	return readyName(local)
}

func (e *emitter) checkReadyRead(local int, expression string) {
	message := fmt.Sprintf("ReferenceError: Cannot access '%s' before initialization", e.program.Locals[local].Name)
	if expression != "" {
		message = fmt.Sprintf("read before assignment: variable '%s' in %s", e.program.Locals[local].Name, expression)
	}
	if failure, found := e.program.ReadyErrors[local]; found {
		e.line("if (!%s) {", e.localReady(local))
		e.indent++
		owned := append([]string(nil), e.owned...)
		value := e.value(failure)
		e.line("adamic_thrown = adamic_retain(%s);", value)
		e.checkThrown()
		e.owned = owned
		e.indent--
		e.line("}")
		return
	}
	// Hand-built legacy IR has no nominal error allocator.
	e.line("if (!%s) {", e.localReady(local))
	e.line("\tstatic const char message[] = %s;", cString(message))
	e.line("\tadamic_panic(message, sizeof message - 1);")
	e.line("}")
}

// read reads a local. A global may change under a later call in the same statement, so its value is
// copied (a string retained) the moment JavaScript would read it; and from inside a function it's
// checked against the temporal dead zone first.
func (e *emitter) read(read ir.Read) string {
	name := e.localName(read.Local)
	if read.Readiness != "" {
		e.checkReadyRead(read.Local, read.Readiness)
	}
	if e.program.Locals[read.Local].Counter {
		// Read as the double it stands for, which every value it can hold is exactly.
		return "((double)" + name + ")"
	}
	// A reference its consumer lends needs no count: nothing can run before it's used (borrow.go).
	lent := e.lendable && lendable(read.Of)
	if slot := e.cellSlot(read.Local); slot != "" {
		if read.Checked {
			e.checkReady(read.Local)
		}
		// A captured variable may change under a call later in the statement (a closure that
		// writes it), so, like a global, it's copied the moment JavaScript reads it.
		value := unslotted(read.Of, slot+"."+member(read.Of))
		if lent {
			e.self = true
			return e.snapshot(read.Of, fmt.Sprintf("(%s)%s", cType(read.Of), value))
		}
		if read.Of.IsReference() {
			return e.own(read.Of, fmt.Sprintf("adamic_retain((%s)%s)", cType(read.Of), value))
		}
		snapshot := e.temporary()
		e.line("%s %s = %s;", cType(read.Of), snapshot, value)
		return snapshot
	}
	if !e.program.Locals[read.Local].Global && !e.program.Locals[read.Local].ExpressionAssigned {
		return name
	}
	if read.Checked {
		e.checkReady(read.Local)
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
	if _, ok := e.asyncSlots[local]; ok {
		e.store(local, value, owned)
		e.line("%s->ready = true;", e.cellReference(local))
		return
	}
	if declared.Counter {
		// A whole-number constant within 2^53, which an integer holds exactly (lower/counters.go).
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
	if declared.GraphCell {
		e.line("adamic_cell *%s = adamic_cell_new((adamic_value){.number = 0}, %t);", cell, declared.Type.IsReference())
		e.adoptGraph(cell, "sizeof *"+cell, true)
		e.hold(cell)
		e.store(local, value, owned)
		return
	}
	if declared.Type.IsReference() && !owned {
		value = retained(value)
	}
	e.line("adamic_cell *%s = adamic_cell_new((adamic_value){.%s = %s}, %t);", cell, member(declared.Type), slotted(declared.Type, value), declared.Type.IsReference())
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
func (e *emitter) allocateEnvironment(cells []int) {
	if len(cells) == 0 {
		return
	}
	environment := e.temporary()
	e.line("adamic_environment *%s = adamic_environment_new(%d);", environment, len(cells))
	graph := false
	for _, local := range cells {
		graph = graph || e.program.Locals[local].GraphCell
	}
	e.adoptGraph(environment, fmt.Sprintf("sizeof *%s + %d * sizeof(adamic_cell)", environment, len(cells)), graph)
	e.hold(environment)
	for position, local := range cells {
		e.line("adamic_cell *%s = &%s->cells[%d];", e.cellName(local), environment, position)
		e.line("%s->references = %t;", e.cellName(local), e.program.Locals[local].Type.IsReference())
	}
}
