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
func (nativeEmitter *emitter) asyncFunction(index int) string {
	function := &nativeEmitter.program.Functions[index]
	identity := stableName("fn", nativeEmitter.sourceReadable(index), nativeEmitter.functionKey(index, map[int]bool{}))
	graph := flow.Build(nativeEmitter.program, index)
	regions := flow.CutAsync(graph)
	var sourceLines map[*ir.Statement]ir.SourceLine
	if nativeEmitter.debug {
		sourceLines = asyncSourceLines(function.Body)
	}
	cells := slices.Clone(function.FrameEnvironment)
	nativeEmitter.asyncSlots = map[int]int{}
	nativeEmitter.asyncBorrows = asyncBorrowedSlots(nativeEmitter.program, index)
	for position, local := range cells {
		nativeEmitter.asyncSlots[local] = position
		nativeEmitter.program.Locals[local].Borrowed = false
		nativeEmitter.program.Locals[local].Counter = false
	}
	nativeEmitter.function = function
	nativeEmitter.functionIndex = index
	nativeEmitter.indent = 1
	nativeEmitter.scopes = [][]string{nil}
	var out strings.Builder
	fmt.Fprintf(
		&out,
		"typedef struct { adamic_async_frame base; adamic_closure *self; adamic_object *error; size_t count; adamic_cell cells[%d]; } adamic_generated_frame_%s;\n",
		max(1, len(cells)),
		identity,
	)
	fmt.Fprintf(
		&out,
		"static void adamic_async_children_%s(adamic_async_frame *base, void (*drop)(void *)) {\n adamic_generated_frame_%s *frame = (void *)base; if (frame->self != NULL) drop(frame->self); if (frame->error != NULL) drop(frame->error); adamic_environment_drop_cells(frame->cells, frame->count, drop);\n}\n",
		identity,
		identity,
	)
	fmt.Fprintf(
		&out,
		"static void adamic_async_finish_%s(adamic_generated_frame_%s *frame) {\n adamic_async_forget_cleanup(&frame->base);\n",
		identity,
		identity,
	)
	for position, local := range cells {
		if nativeEmitter.program.Locals[local].IterationCell ||
			(nativeEmitter.program.Locals[local].Type.IsReference() && !nativeEmitter.program.Locals[local].Captured) {
			if nativeEmitter.asyncBorrows[local] && !nativeEmitter.program.Locals[local].IterationCell {
				fmt.Fprintf(&out, " frame->cells[%d].value.reference = NULL;\n", position)
			} else {
				fmt.Fprintf(
					&out,
					" void *adamic_slot_%d = frame->cells[%d].value.reference; frame->cells[%d].value.reference = NULL; if (adamic_slot_%d != NULL) adamic_release(adamic_slot_%d);\n",
					position,
					position,
					position,
					position,
					position,
				)
			}
			fmt.Fprintf(&out, " frame->cells[%d].references = false;\n", position)
		}
	}
	out.WriteString(
		" void *self = frame->self; frame->self = NULL; if (self != NULL) adamic_release(self); void *error = frame->error; frame->error = NULL; if (error != NULL) adamic_release(error); void *output = frame->base.output; frame->base.output = NULL; adamic_release(output);\n}\n",
	)
	fmt.Fprintf(
		&out,
		"static void adamic_async_abandon_%s(adamic_async_frame *base) { adamic_async_finish_%s((void *)base); }\n",
		identity,
		identity,
	)
	fmt.Fprintf(
		&out,
		"static void adamic_async_resume_%s(adamic_async_frame *base, adamic_value resumed, bool rejected) {\n adamic_generated_frame_%s *frame = (void *)base; adamic_closure *self = frame->self; (void)self; (void)resumed; (void)rejected;\n switch (base->state) {\n",
		identity,
		identity,
	)
	// One state names each region entry. Fulfillment and rejection bind before jumping.
	waits := map[flow.BlockId]*flow.Suspend{}
	for _, block := range graph.Blocks {
		if wait, ok := block.Terminal.(*flow.Suspend); ok {
			waits[wait.Fulfilled] = wait
		}
	}
	for state, root := range regions.Roots {
		nativeEmitter.line("case %d:", state)
		if terminal, ok := waits[root]; ok {
			nativeEmitter.line(
				"if (rejected) { adamic_release(frame->error); frame->error = adamic_retain(resumed.reference); goto adamic_async_block_%d; }",
				terminal.Rejected,
			)
			if terminal.Local >= 0 {
				value := fmt.Sprintf(
					"(%s)%s",
					cType(terminal.Of),
					unslotted(terminal.Of, "resumed."+member(terminal.Of)),
				)
				nativeEmitter.store(terminal.Local, value, false)
				nativeEmitter.line("%s->ready = true;", nativeEmitter.cellReference(terminal.Local))
			}
		}
		nativeEmitter.line("goto adamic_async_block_%d;", root)
	}
	nativeEmitter.line("default: adamic_unreachable();")
	nativeEmitter.line("}")
	nativeEmitter.suspensionBlocks(graph, regions, "adamic_async_block", sourceLines, func(
		block *flow.BasicBlock,
		result ir.Expression,
		waiting string,
	) {
		switch terminal := block.Terminal.(type) {
		case *flow.Suspend:
			nativeEmitter.line("base->state = %d;", regions.Owner[terminal.Fulfilled])
			nativeEmitter.line("adamic_async_await(base, %s);", waiting)
			nativeEmitter.end()
			nativeEmitter.line("return;")
		case *flow.Return:
			value, refs := "(adamic_value){.number = 0}", false
			if result != nil {
				value = borrowed(result.Type(), nativeEmitter.value(result))
				refs = result.Type().IsReference()
			}
			nativeEmitter.line("adamic_async_settle(base->output, %s, %t, false);", value, refs)
			nativeEmitter.end()
			nativeEmitter.line("adamic_async_finish_%s(frame); return;", identity)
		case *flow.Throw:
			nativeEmitter.line(
				"adamic_async_settle(base->output, (adamic_value){.reference = frame->error}, true, true);",
			)
			nativeEmitter.line("adamic_async_finish_%s(frame); return;", identity)
		default:
			panic(fmt.Sprintf("native: unproved async terminal %T", terminal))
		}
	})
	if nativeEmitter.debug {
		nativeEmitter.line(sourceLineReset)
	}
	out.WriteString(nativeEmitter.out.String())
	nativeEmitter.out.Reset()
	out.WriteString("}\n")
	fmt.Fprintf(
		&out,
		"static %s {\n ADAMIC_CHECK_STACK();\n adamic_generated_frame_%s *frame = adamic_allocate(sizeof *frame, adamic_kind_async_frame);\n memset((char *)frame + sizeof(adamic_heap), 0, sizeof *frame - sizeof(adamic_heap));\n frame->count = %d; adamic_environment_initialize_cells(&frame->base.heap, frame->cells, frame->count);\n adamic_async_promise *output = adamic_async_new(); frame->base.output = adamic_retain(output);\n frame->base.resume = adamic_async_resume_%s; frame->base.children = adamic_async_children_%s;\n",
		nativeEmitter.signature(index),
		identity,
		len(cells),
		identity,
		identity,
	)
	fmt.Fprintf(&out, " adamic_async_register_cleanup(&frame->base, adamic_async_abandon_%s);\n", identity)
	if function.Closure {
		out.WriteString(" frame->self = adamic_retain(self); (void)arguments;\n")
	}
	for position, local := range cells {
		fmt.Fprintf(
			&out,
			" frame->cells[%d].references = %t;\n",
			position,
			nativeEmitter.program.Locals[local].IterationCell ||
				(nativeEmitter.program.Locals[local].Type.IsReference() && !nativeEmitter.asyncBorrows[local]),
		)
	}
	if function.ArgumentsCount != 0 {
		local := function.ArgumentsCount - 1
		count := nativeEmitter.localName(local)
		if function.Closure {
			count = "(double)argument_count"
			if function.Receiver {
				count += " - 1"
			}
		}
		fmt.Fprintf(
			&out,
			" frame->cells[%d].value.number = %s; frame->cells[%d].ready = true;\n",
			nativeEmitter.asyncSlots[local],
			count,
			nativeEmitter.asyncSlots[local],
		)
	}
	for position, local := range function.Parameters {
		value := nativeEmitter.localName(local)
		if function.Closure {
			of := nativeEmitter.program.Locals[local].Type
			value = unslotted(of, fmt.Sprintf("arguments[%d].%s", position, member(of)))
			if function.RestElement != 0 && position == len(function.Parameters)-1 {
				slot := nativeEmitter.program.RestArgumentSlots[ir.FunctionRestArguments(*function)]
				if function.Receiver {
					slot++
				}
				value = fmt.Sprintf("arguments[%d].reference", slot)
			} else if nativeEmitter.program.PackedCountNeeded(index) {
				value = closureArgument(of, position)
			}
		}
		if nativeEmitter.program.Locals[local].Type.IsReference() {
			value = "adamic_retain(" + value + ")"
		}
		fmt.Fprintf(
			&out,
			" frame->cells[%d].value.%s = %s; frame->cells[%d].ready = true;\n",
			nativeEmitter.asyncSlots[local],
			member(nativeEmitter.program.Locals[local].Type),
			slotted(nativeEmitter.program.Locals[local].Type, value),
			nativeEmitter.asyncSlots[local],
		)
	}
	fmt.Fprintf(
		&out,
		" adamic_async_resume_%s(&frame->base, (adamic_value){.number = 0}, false); adamic_release(frame);\n",
		identity,
	)
	if function.Closure {
		out.WriteString(" return (adamic_value){.reference = output};\n}\n")
	} else {
		out.WriteString(" return output;\n}\n")
	}
	nativeEmitter.asyncSlots = nil
	nativeEmitter.asyncBorrows = nil
	nativeEmitter.function = nil
	nativeEmitter.scopes = nil
	return out.String()
}

func (nativeEmitter *emitter) promiseValue(expression ir.PromiseValue) string {
	value, refs := "(adamic_value){.number = 0}", false
	if expression.Value != nil {
		value = borrowed(expression.Value.Type(), nativeEmitter.value(expression.Value))
		refs = expression.Value.Type().IsReference()
	}
	promise := nativeEmitter.own(ir.Promise, "adamic_async_new()")
	nativeEmitter.line("adamic_async_settle(%s, %s, %t, %t);", promise, value, refs, expression.Reject)
	return promise
}
func (nativeEmitter *emitter) asyncMain() {
	if nativeEmitter.program.AsyncEntry == 0 {
		return
	}
	nativeEmitter.line(
		"adamic_async_promise *root = %s();",
		nativeEmitter.functionName(nativeEmitter.program.AsyncEntry-1),
	)
	nativeEmitter.line("adamic_async_run();")
	nativeEmitter.line("if (root->rejected) { adamic_thrown = root->value.reference; adamic_uncaught(); }")
	nativeEmitter.line("adamic_release(root);")
}

// suspensionBlocks emits every region's blocks of a suspending function, an async function's or a
// generator's, each under label and its block id: the instructions, with the same expression,
// statement, store and exception emitters as a synchronous function, and the jumps between blocks.
// The terminals that leave the function (a suspension, a return, a throw) are its protocol's, and
// leave is handed the block, its return's value and the value its last instruction made for a
// suspension. Under --debug each instruction carries its source line (#jhyfmk4).
func (nativeEmitter *emitter) suspensionBlocks(
	graph *flow.Function,
	regions flow.AsyncRegions,
	label string,
	sourceLines map[*ir.Statement]ir.SourceLine,
	leave func(block *flow.BasicBlock, result ir.Expression, waiting string),
) {
	var lastSourceLine ir.SourceLine
	for _, blocks := range regions.Blocks {
		for _, id := range blocks {
			block, _ := graph.Block(id)
			nativeEmitter.line("%s_%d: {", label, id)
			nativeEmitter.indent++
			nativeEmitter.asyncHandler = ""
			switch terminal := block.Terminal.(type) {
			case *flow.MayThrow:
				nativeEmitter.asyncHandler = fmt.Sprintf("%s_%d", label, terminal.Handler)
			case *flow.Goto:
				// Only an explicit throw uses its normal Goto as an exception edge.
				if len(block.Instructions) > 0 {
					last := graph.Instructions[block.Instructions[len(block.Instructions)-1]]
					if _, throwing := (*last.At).(ir.Throw); throwing {
						nativeEmitter.asyncHandler = fmt.Sprintf("%s_%d", label, terminal.Block)
					}
				}
			}
			var condition, waiting string
			var result ir.Expression
			for _, instructionID := range block.Instructions {
				instruction := graph.Instructions[instructionID]
				if source, ok := sourceLines[instruction.At]; ok && source != lastSourceLine {
					nativeEmitter.statement(source)
					lastSourceLine = source
				}
				statement := *instruction.At
				if attempt, ok := statement.(ir.Try); ok && instruction.Part == 1 {
					if attempt.CatchLocal >= 0 {
						nativeEmitter.declareLocal(attempt.CatchLocal, "frame->error", true)
					} else {
						nativeEmitter.line("adamic_release(frame->error);")
					}
					nativeEmitter.line("frame->error = NULL;")
					continue
				}
				switch statement := statement.(type) {
				case ir.Switch:
					if instruction.Part > 0 {
						condition = nativeEmitter.snapshot(ir.Boolean, nativeEmitter.decide(
							ir.Binary{Operator: ir.Equal, Left: statement.Value, Right: instruction.Expression},
						))
					}
				case ir.If, ir.Loop:
					condition = nativeEmitter.snapshot(ir.Boolean, nativeEmitter.decide(instruction.Expression))
				case ir.Return:
					result = statement.Value
				case ir.Declare:
					if _, ok := statement.Value.(ir.Await); ok {
						if instruction.Part == 1 {
							continue
						}
						waiting = nativeEmitter.value(instruction.Expression)
					} else {
						nativeEmitter.statementAt(instruction.At)
					}
				case ir.Evaluate:
					_, awaits := statement.Value.(ir.Await)
					_, yields := statement.Value.(ir.Yield)
					if awaits || yields {
						if instruction.Part == 1 {
							continue
						}
						waiting = nativeEmitter.value(instruction.Expression)
					} else {
						nativeEmitter.statementAt(instruction.At)
					}
				default:
					nativeEmitter.statementAt(instruction.At)
				}
			}
			switch terminal := block.Terminal.(type) {
			case *flow.Goto:
				nativeEmitter.line("goto %s_%d;", label, terminal.Block)
			case *flow.MayThrow:
				nativeEmitter.line("goto %s_%d;", label, terminal.Next)
			case *flow.If:
				nativeEmitter.line(
					"if (%s) goto %s_%d; else goto %s_%d;",
					condition,
					label,
					terminal.Consequent,
					label,
					terminal.Alternate,
				)
			case *flow.Unreachable:
				nativeEmitter.line("adamic_unreachable();")
			default:
				leave(block, result, waiting)
			}
			nativeEmitter.asyncHandler = ""
			nativeEmitter.indent--
			nativeEmitter.line("}")
		}
	}
}
