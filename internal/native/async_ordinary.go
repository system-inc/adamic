package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
	"slices"
	"strings"
)

// asyncFunction emits only the scheduler protocol and graph terminals. Ordinary instructions
// keep the same expression, statement, store and exception emitters as synchronous functions.
func (e *emitter) asyncFunction(index int) string {
	function := &e.program.Functions[index]
	graph := flow.Build(e.program, index)
	regions := flow.CutAsync(graph)
	cells := slices.Clone(function.FrameEnvironment)
	e.asyncSlots = map[int]int{}
	for position, local := range cells {
		e.asyncSlots[local] = position
		e.program.Locals[local].Borrowed = false
		e.program.Locals[local].Counter = false
	}
	e.function = function
	e.functionIndex = index
	e.indent = 1
	e.scopes = [][]string{nil}
	var out strings.Builder
	fmt.Fprintf(&out, "typedef struct { adamic_async_frame base; adamic_closure *self; adamic_object *error; size_t count; adamic_cell cells[%d]; } adamic_generated_frame_%d;\n", max(1, len(cells)), index)
	fmt.Fprintf(&out, "static void adamic_async_children_%d(adamic_async_frame *base, void (*drop)(void *)) {\n adamic_generated_frame_%d *frame = (void *)base; drop(frame->self); drop(frame->error); adamic_environment_drop_cells(frame->cells, frame->count, drop);\n}\n", index, index)
	fmt.Fprintf(&out, "static void adamic_async_finish_%d(adamic_generated_frame_%d *frame) {\n adamic_async_forget_cleanup(&frame->base);\n", index, index)
	for position, local := range cells {
		if e.program.Locals[local].IterationCell || (e.program.Locals[local].Type.IsReference() && !e.program.Locals[local].Captured) {
			fmt.Fprintf(&out, " void *adamic_slot_%d = frame->cells[%d].value.reference; frame->cells[%d].value.reference = NULL; adamic_release(adamic_slot_%d);\n", position, position, position, position)
		}
	}
	out.WriteString(" void *self = frame->self; frame->self = NULL; adamic_release(self); void *error = frame->error; frame->error = NULL; adamic_release(error); void *output = frame->base.output; frame->base.output = NULL; adamic_release(output);\n}\n")
	fmt.Fprintf(&out, "static void adamic_async_abandon_%d(adamic_async_frame *base) { adamic_async_finish_%d((void *)base); }\n", index, index)
	fmt.Fprintf(&out, "static void adamic_async_resume_%d(adamic_async_frame *base, adamic_value resumed, bool rejected) {\n adamic_generated_frame_%d *frame = (void *)base; adamic_closure *self = frame->self; (void)self; (void)resumed; (void)rejected;\n switch (base->state) {\n", index, index)
	// One state names each region entry. Fulfillment and rejection bind before jumping.
	waits := map[flow.BlockId]*flow.Suspend{}
	for _, block := range graph.Blocks {
		if wait, ok := block.Terminal.(*flow.Suspend); ok {
			waits[wait.Fulfilled] = wait
		}
	}
	for state, root := range regions.Roots {
		e.line("case %d:", state)
		if terminal, ok := waits[root]; ok {
			e.line("if (rejected) { adamic_release(frame->error); frame->error = adamic_retain(resumed.reference); goto adamic_async_block_%d; }", terminal.Rejected)
			if terminal.Local >= 0 {
				value := fmt.Sprintf("(%s)%s", cType(terminal.Of), unslotted(terminal.Of, "resumed."+member(terminal.Of)))
				e.store(terminal.Local, value, false)
				e.line("%s->ready = true;", e.cellReference(terminal.Local))
			}
		}
		e.line("goto adamic_async_block_%d;", root)
	}
	e.line("default: adamic_unreachable();")
	e.line("}")
	for _, blocks := range regions.Blocks {
		for _, id := range blocks {
			block, _ := graph.Block(id)
			e.line("adamic_async_block_%d: {", id)
			e.indent++
			e.asyncHandler = ""
			switch terminal := block.Terminal.(type) {
			case *flow.MayThrow:
				e.asyncHandler = fmt.Sprintf("adamic_async_block_%d", terminal.Handler)
			case *flow.Goto:
				e.asyncHandler = fmt.Sprintf("adamic_async_block_%d", terminal.Block)
			}
			var condition, waiting string
			var result ir.Expression
			for _, instructionID := range block.Instructions {
				instruction := graph.Instructions[instructionID]
				statement := *instruction.At
				if instruction.Part == 1 {
					if attempt, ok := statement.(ir.Try); ok {
						if attempt.CatchLocal >= 0 {
							e.declareLocal(attempt.CatchLocal, "frame->error", true)
						} else {
							e.line("adamic_release(frame->error);")
						}
						e.line("frame->error = NULL;")
					}
					continue
				}
				switch statement := statement.(type) {
				case ir.If, ir.Loop:
					condition = e.snapshot(ir.Boolean, e.decide(instruction.Expression))
				case ir.Return:
					result = statement.Value
				case ir.Declare:
					if _, ok := statement.Value.(ir.Await); ok {
						waiting = e.value(instruction.Expression)
					} else {
						e.statementAt(instruction.At)
					}
				case ir.Evaluate:
					if _, ok := statement.Value.(ir.Await); ok {
						waiting = e.value(instruction.Expression)
					} else {
						e.statementAt(instruction.At)
					}
				default:
					e.statementAt(instruction.At)
				}
			}
			switch terminal := block.Terminal.(type) {
			case *flow.Goto:
				e.line("goto adamic_async_block_%d;", terminal.Block)
			case *flow.MayThrow:
				e.line("goto adamic_async_block_%d;", terminal.Next)
			case *flow.If:
				e.line("if (%s) goto adamic_async_block_%d; else goto adamic_async_block_%d;", condition, terminal.Consequent, terminal.Alternate)
			case *flow.Suspend:
				e.line("base->state = %d;", regions.Owner[terminal.Fulfilled])
				e.line("adamic_async_await(base, %s);", waiting)
				e.end()
				e.line("return;")
			case *flow.Return:
				value, refs := "(adamic_value){.number = 0}", false
				if result != nil {
					value = borrowed(result.Type(), e.value(result))
					refs = result.Type().IsReference()
				}
				e.line("adamic_async_settle(base->output, %s, %t, false);", value, refs)
				e.end()
				e.line("adamic_async_finish_%d(frame); return;", index)
			case *flow.Throw:
				e.line("adamic_async_settle(base->output, (adamic_value){.reference = frame->error}, true, true);")
				e.line("adamic_async_finish_%d(frame); return;", index)
			case *flow.Unreachable:
				e.line("adamic_unreachable();")
			default:
				panic(fmt.Sprintf("native: unproved async terminal %T", terminal))
			}
			e.asyncHandler = ""
			e.indent--
			e.line("}")
		}
	}
	out.WriteString(e.out.String())
	e.out.Reset()
	out.WriteString("}\n")
	fmt.Fprintf(&out, "static %s {\n ADAMIC_CHECK_STACK();\n adamic_generated_frame_%d *frame = adamic_allocate(sizeof *frame, adamic_kind_async_frame);\n memset((char *)frame + sizeof(adamic_heap), 0, sizeof *frame - sizeof(adamic_heap));\n frame->count = %d; adamic_environment_initialize_cells(&frame->base.heap, frame->cells, frame->count);\n adamic_async_promise *output = adamic_async_new(); frame->base.output = adamic_retain(output);\n frame->base.resume = adamic_async_resume_%d; frame->base.children = adamic_async_children_%d;\n", e.signature(index), index, len(cells), index, index)
	fmt.Fprintf(&out, " adamic_async_register_cleanup(&frame->base, adamic_async_abandon_%d);\n", index)
	if function.Closure {
		out.WriteString(" frame->self = adamic_retain(self); (void)arguments;\n")
	}
	for position, local := range cells {
		fmt.Fprintf(&out, " frame->cells[%d].references = %t;\n", position, e.program.Locals[local].Type.IsReference() || e.program.Locals[local].IterationCell)
	}
	for position, local := range function.Parameters {
		value := e.localName(local)
		if function.Closure {
			value = unslotted(e.program.Locals[local].Type, fmt.Sprintf("arguments[%d].%s", position, member(e.program.Locals[local].Type)))
		}
		if e.program.Locals[local].Type.IsReference() {
			value = "adamic_retain(" + value + ")"
		}
		fmt.Fprintf(&out, " frame->cells[%d].value.%s = %s; frame->cells[%d].ready = true;\n", e.asyncSlots[local], member(e.program.Locals[local].Type), slotted(e.program.Locals[local].Type, value), e.asyncSlots[local])
	}
	fmt.Fprintf(&out, " adamic_async_resume_%d(&frame->base, (adamic_value){.number = 0}, false); adamic_release(frame);\n", index)
	if function.Closure {
		out.WriteString(" return (adamic_value){.reference = output};\n}\n")
	} else {
		out.WriteString(" return output;\n}\n")
	}
	e.asyncSlots = nil
	e.function = nil
	e.scopes = nil
	return out.String()
}

func (e *emitter) promiseValue(expression ir.PromiseValue) string {
	value, refs := "(adamic_value){.number = 0}", false
	if expression.Value != nil {
		value = borrowed(expression.Value.Type(), e.value(expression.Value))
		refs = expression.Value.Type().IsReference()
	}
	promise := e.own(ir.Promise, "adamic_async_new()")
	e.line("adamic_async_settle(%s, %s, %t, %t);", promise, value, refs, expression.Reject)
	return promise
}
func (e *emitter) asyncMain() {
	if e.program.AsyncEntry == 0 {
		return
	}
	e.line("adamic_async_promise *root = %s();", e.functionName(e.program.AsyncEntry-1))
	e.line("adamic_async_run();")
	e.line("if (root->rejected) { adamic_thrown = adamic_retain(root->value.reference); adamic_uncaught(); }")
	e.line("adamic_release(root);")
}
