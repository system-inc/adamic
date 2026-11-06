package native

import (
	"fmt"
	"slices"
	"strings"

	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
)

// generatorFunction emits a generator (#qk8rztp S1c): its resume function, which runs the body
// from the frame's state to the next way out, its code record, and the function a call calls, which
// makes the frame and runs none of the body. The body is cut into regions at each yield, as an async
// function's is at each await, and its blocks are emitted by the same suspensionBlocks.
//
// The frame is runtime/generator.h's: every local and parameter in a cell, each counted cell owning
// its value or holding NULL, so finish and the drop walk need no state. A yield first hands over its
// value, owned, then kills every counted cell the resume never reads again (flow.LiveIn of the block
// it resumes at), so a suspended generator holds only what it will use. A cell the graph doesn't
// track (one an expression assigns) counts as live and is never killed. Unlike an async frame, no
// cell borrows another's value (asyncBorrowedSlots): its premise, that both slots survive every
// suspension, fails across a yield, whose kill set can release a source while its alias is live.
//
// A nested generator (#qk8rztp S3a) is a closure: it is called the way every closure is (adamic_code),
// through the carrier its call names, and its entry retains that carrier into the frame's self, so
// its resume reads its parent's cells through self->cells, as any closure's body does, at every step,
// however soon whatever made the carrier lets go of it.
func (e *emitter) generatorFunction(index int) string {
	function := &e.program.Functions[index]
	identity := stableName("fn", e.sourceReadable(index), e.functionKey(index, map[int]bool{}))
	graph := flow.Build(e.program, index)
	regions := flow.CutAsync(graph)
	if len(regions.Roots) == 0 || regions.Roots[0] != graph.Entry {
		panic("native: a generator's first region must start at its entry")
	}
	liveIn := flow.LiveIn(graph)
	var sourceLines map[*ir.Statement]ir.SourceLine
	if e.debug {
		sourceLines = asyncSourceLines(function.Body)
	}
	cells := slices.Clone(function.FrameEnvironment)
	e.asyncSlots = map[int]int{}
	e.asyncBorrows = map[int]bool{}
	for position, local := range cells {
		e.asyncSlots[local] = position
		e.program.Locals[local].Borrowed = false
		e.program.Locals[local].Counter = false
	}
	// A cell owns its value when it holds a reference; the walk and the kills read only this flag.
	counted := func(local int) bool {
		return e.program.Locals[local].IterationCell || e.program.Locals[local].Type.IsReference()
	}
	e.function = function
	e.functionIndex = index
	e.indent = 1
	e.scopes = [][]string{nil}
	// Named as functionBody names a function, so a panic and a debugger read the .a name.
	name := function.Name
	if len(function.Source.Declaration) != 0 {
		name = function.Source.Declaration[len(function.Source.Declaration)-1]
	}
	if name == "" {
		name = "<anonymous>"
	}
	var out strings.Builder
	fmt.Fprintf(&out, "static void adamic_generator_resume_%s(adamic_generator *frame, adamic_value *yielded) {\n", identity)
	// The resume runs on its consumer's stack, so it checks it as every emitted function does.
	e.line("%s;", stackCheck(name))
	if function.Closure {
		// The carrier, under the name every captured read and sibling call emits (cellReference,
		// ClosureSelf). A top-level generator declares none, so a read through one can't compile.
		e.line("adamic_closure *self = frame->self; (void)self;")
	}
	e.line("switch (frame->state) {")
	for state, root := range regions.Roots {
		e.line("case %d: goto adamic_generator_block_%d;", state, root)
	}
	e.line("default: adamic_unreachable();")
	e.line("}")
	e.suspensionBlocks(graph, regions, "adamic_generator_block", sourceLines, func(block *flow.BasicBlock, result ir.Expression, waiting string) {
		switch terminal := block.Terminal.(type) {
		case *flow.Yield:
			if waiting == "" {
				panic("native: a yield with no value instruction")
			}
			of := function.Yields
			value := waiting
			if of.IsReference() {
				// Retained before any kill below can release the cell it was read from.
				value = e.kept(waiting)
			}
			e.line("*yielded = (adamic_value){.%s = %s};", member(of), slotted(of, value))
			live := liveIn[terminal.Next]
			for position, local := range cells {
				declared := e.program.Locals[local]
				if counted(local) && !declared.Captured && !declared.ExpressionAssigned && !declared.Global && !live[flow.DeclarationId(local+1)] {
					e.line("adamic_generator_kill(frame, %d);", position)
				}
			}
			e.end()
			e.line("frame->state = %d;", regions.Owner[terminal.Next])
			e.line("frame->status = adamic_generator_status_suspended;")
			e.line("return;")
		case *flow.Return:
			if result != nil {
				panic("native: a generator returning a value")
			}
			e.end()
			e.line("adamic_generator_finish(frame); return;")
		case *flow.Throw:
			// finish releases frame->error, so the error leaves the frame first and is the pending
			// exception the step's caller tests.
			e.line("adamic_object *thrown = frame->error; frame->error = NULL;")
			e.line("adamic_generator_finish(frame); adamic_thrown = thrown; return;")
		default:
			panic(fmt.Sprintf("native: unproved generator terminal %T", terminal))
		}
	})
	if e.debug {
		e.line(sourceLineReset)
	}
	out.WriteString(e.out.String())
	e.out.Reset()
	out.WriteString("}\n")
	fmt.Fprintf(&out, "static const adamic_generator_code adamic_generator_code_%s = {%s, adamic_generator_resume_%s};\n", identity, cString(name), identity)
	fmt.Fprintf(&out, "static %s {\n %s;\n adamic_generator *frame = adamic_generator_new(%d, &adamic_generator_code_%s);\n", e.signature(index), stackCheck(name), len(cells), identity)
	if function.Closure {
		// The frame's own reference: the caller's may be a temporary it lets go of before the first
		// step (Memory's must on #qk8rztp S3a).
		out.WriteString(" frame->self = adamic_retain(self); (void)arguments;\n")
		if e.program.PackedCountNeeded(index) {
			out.WriteString(" (void)argument_count;\n")
		}
	}
	for position, local := range cells {
		fmt.Fprintf(&out, " frame->cells[%d].references = %t;\n", position, counted(local))
	}
	for position, local := range function.Parameters {
		of := e.program.Locals[local].Type
		value := e.localName(local)
		if function.Closure {
			// Packed as a closure's arguments are, as an async closure's entry reads them.
			value = unslotted(of, fmt.Sprintf("arguments[%d].%s", position, member(of)))
			if e.program.PackedCountNeeded(index) {
				value = closureArgument(of, position)
			}
		}
		if of.IsReference() {
			// Every parameter is the frame's own, kept past the call that made it (rulings.md:320).
			value = "adamic_retain(" + value + ")"
		}
		fmt.Fprintf(&out, " frame->cells[%d].value.%s = %s; frame->cells[%d].ready = true;\n", e.asyncSlots[local], member(of), slotted(of, value), e.asyncSlots[local])
	}
	if function.Closure {
		out.WriteString(" return (adamic_value){.reference = frame};\n}\n")
	} else {
		out.WriteString(" return frame;\n}\n")
	}
	e.asyncSlots = nil
	e.asyncBorrows = nil
	e.function = nil
	e.scopes = nil
	return out.String()
}

// generatorNext emits a step: the resume runs to the next way out and its value comes back owned.
// When the body can throw, the pending exception is tested right after, before anything reads done.
func (e *emitter) generatorNext(next ir.GeneratorNext) string {
	generator := generatorPointer(next.Generator, e.value(next.Generator))
	step := fmt.Sprintf("adamic_generator_next(%s).%s", generator, member(next.Of))
	var result string
	if next.Of.IsReference() {
		result = e.own(next.Of, fmt.Sprintf("(%s)%s", cType(next.Of), step))
	} else {
		result = e.snapshot(next.Of, unslotted(next.Of, step))
	}
	if e.program.GeneratorNextMayThrow(next) {
		e.checkThrown()
	}
	return result
}

// generatorPointer is a step's generator as the runtime takes it. An iterable view's dispatch steps
// the object it holds, which IsGenerator has just said is a generator (lower/iterable_views.go).
func generatorPointer(generator ir.Expression, value string) string {
	if generator.Type() == ir.Object {
		return "((adamic_generator *)" + value + ")"
	}
	return value
}
