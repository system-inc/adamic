package flow

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

// point is one instruction of one graph: the function (-1 for the top level) and its id.
type point struct {
	function    int
	instruction InstructionId
}

// Every path Node takes through a program is a path in the graph. Reaching definitions and the
// verifier run over the graph Build made, so they agree with a wrong edge; this holds the edges to
// what runs. The JavaScript backend marks every point the graph has an instruction for, and each
// call's sequence of marks must walk the graph: from one instruction to the next in its block, or
// from a block's last to the first of a block its terminal reaches through blocks that run nothing.
// A program that finishes must end every call where the graph returns.
func TestEveryPathNodeTakesIsInTheGraph(t *testing.T) {
	t.Parallel()
	for _, path := range programs(t) {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			run := traced(t, path)
			for _, problem := range walk(run.graphs, run.marked, run.events) {
				t.Error(problem)
			}
			// Every program runs something; a trace without a point would pass every check here.
			walked := 0
			for _, event := range run.events {
				if _, err := strconv.Atoi(event); err == nil {
					walked++
				}
			}
			if walked == 0 {
				t.Errorf("Node ran no point at all: %q", run.events)
			}
			t.Logf("%d points walked, %d events", walked, len(run.events))
		})
	}
}

// run is one program run on Node with every point marked.
type run struct {
	graphs map[int]*Function
	marked []point
	events []string
}

// traceRuntime is what a traced program runs with, before its own code. adamicPoint records a point
// and, first, what the instruction before it in the same call did to the objects the call's tracked
// variables hold: one that holds the same object as it did at that point, which now prints
// differently, was mutated by that instruction. Printing follows every field, element and map entry,
// so a mutation of anything a variable's object reaches counts; a closure prints as itself, since
// what its cells hold isn't a tracked value. Calls are frames, and each compares only its own.
const traceRuntime = `
const adamicTrace = [];
const adamicIdentities = new WeakMap();
let adamicNextIdentity = 1;
const adamicIdentity = (value) => {
	if (!adamicIdentities.has(value)) adamicIdentities.set(value, adamicNextIdentity++);
	return adamicIdentities.get(value);
};
const adamicPrint = (value, seen) => {
	if (typeof value === 'number') return Object.is(value, -0) ? '-0' : String(value);
	if (typeof value === 'string') return JSON.stringify(value);
	if (value === null || typeof value !== 'object') return String(value);
	if (seen.has(value)) return '@' + adamicIdentity(value);
	seen.add(value);
	if (typeof value.code === 'function') return 'closure' + adamicIdentity(value);
	if (value instanceof Map) return 'map(' + [...value].map(([key, entry]) => adamicPrint(key, seen) + '=>' + adamicPrint(entry, seen)).join(',') + ')';
	if (Array.isArray(value)) return '[' + value.map((element) => adamicPrint(element, seen)).join(',') + ']';
	return '{' + Object.keys(value).map((key) => key + ':' + adamicPrint(value[key], seen)).join(',') + '}';
};
// A panic on Node is a throw a catch can take, though natively it ends the program where it stands, so
// what runs after one isn't a path of the program: the runtime silences stdout when it panics, and
// from the first point after that, the trace says so and records nothing more.
const adamicStdoutWrite = process.stdout.write;
let adamicStopped = false;
const adamicPanicked = () => {
	if (!adamicStopped && process.stdout.write !== adamicStdoutWrite) {
		adamicStopped = true;
		adamicTrace.push('panicked');
	}
	return adamicStopped;
};
const adamicFrames = [{ last: -1, held: new Map() }];
const adamicEnter = (function_) => {
	if (adamicPanicked()) return;
	adamicFrames.push({ last: -1, held: new Map() });
	adamicTrace.push('enter ' + function_);
};
const adamicLeave = (function_) => {
	if (adamicPanicked()) return;
	adamicFrames.pop();
	adamicTrace.push('leave ' + function_);
};
const adamicMissing = () => {
	if (!adamicPanicked()) adamicTrace.push('missing');
};
const adamicPoint = (point, variables) => {
	if (adamicPanicked()) return;
	const frame = adamicFrames[adamicFrames.length - 1];
	for (const [local, read] of variables) {
		let value;
		try { value = read(); } catch { frame.held.delete(local); continue; }
		if (value === null || typeof value !== 'object') { frame.held.delete(local); continue; }
		const identity = adamicIdentity(value), print = adamicPrint(value, new Set());
		const before = frame.held.get(local);
		if (before !== undefined && before.identity === identity && before.print !== print && frame.last >= 0) {
			adamicTrace.push('mutated ' + local + ' ' + frame.last);
		}
		frame.held.set(local, { identity, print });
	}
	frame.last = point;
	adamicTrace.push(point);
};
`

// traced runs a program on Node, through the JavaScript backend with every point marked, and returns
// its graphs (each in single assignment form) and the trace.
func traced(t *testing.T, path string) run {
	t.Helper()
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	result := run{graphs: map[int]*Function{}}
	points := map[*ir.Statement]map[int]point{}
	// variables is, per function, the JavaScript that reads each of its tracked variables that can
	// hold an object.
	variables := map[int]string{}
	for function := -1; function < len(program.Functions); function++ {
		graph := Build(program, function)
		Construct(graph)
		result.graphs[function] = graph
		reads := []string{}
		seen := map[DeclarationId]bool{}
		for _, identifier := range graph.Identifiers {
			local := int(identifier.Declaration) - 1
			if identifier.Declaration == 0 || seen[identifier.Declaration] || !mutable(program.Locals[local].Type) {
				continue
			}
			seen[identifier.Declaration] = true
			reads = append(reads, fmt.Sprintf("[%d, () => %s]", local, javascript.Name(program, local)))
		}
		variables[function] = "[" + strings.Join(reads, ", ") + "]"
		for _, block := range graph.Blocks {
			for _, id := range block.Instructions {
				instruction := graph.Instructions[id]
				if points[instruction.At] == nil {
					points[instruction.At] = map[int]point{}
				}
				points[instruction.At][instruction.Part] = point{function, id}
			}
		}
	}
	options := javascript.Options{
		Mark: func(at *ir.Statement, part int) string {
			found, ok := points[at][part]
			if !ok {
				return "adamicMissing()"
			}
			result.marked = append(result.marked, found)
			return fmt.Sprintf("adamicPoint(%d, %s)", len(result.marked)-1, variables[found.function])
		},
		Enter: func(function int) string {
			return fmt.Sprintf("adamicEnter(%d)", function)
		},
		Leave: func(function int) string {
			return fmt.Sprintf("adamicLeave(%d)", function)
		},
	}
	directory := t.TempDir()
	trace := filepath.Join(directory, "trace.txt")
	source := fmt.Sprintf("import { writeFileSync as adamicWrite } from 'node:fs';\n%s\nprocess.on('exit', (code) => adamicWrite(%s, [...adamicTrace, `exit ${code}`].join('\\n')));\n%s",
		traceRuntime, strconv.Quote(trace), javascript.JavaScriptWith(program, options))
	module := filepath.Join(directory, "program.mjs")
	if err := os.WriteFile(module, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", runner, module)
	command.Dir = filepath.Dir(path)
	if output, err := command.CombinedOutput(); err != nil {
		if _, exited := err.(*exec.ExitError); !exited {
			t.Fatalf("node: %v\n%s", err, output)
		}
	}
	contents, err := os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	result.events = strings.Split(string(contents), "\n")
	return result
}

// walk follows a trace through the graphs and says everything about it that isn't a path.
func walk(graphs map[int]*Function, marked []point, events []string) []string {
	type frame struct {
		function int
		last     *point
	}
	var problems, ends []string
	stack := []frame{{function: -1}}
	end := func(current frame) {
		if problem := finishes(graphs[current.function], current.last); problem != "" {
			ends = append(ends, fmt.Sprintf("function %d: %s", current.function, problem))
		}
	}
	exit := ""
	for _, event := range events {
		top := &stack[len(stack)-1]
		switch {
		case event == "missing":
			problems = append(problems, fmt.Sprintf("function %d ran a point the graph says can't be reached", top.function))
		case strings.HasPrefix(event, "enter "):
			function, _ := strconv.Atoi(strings.TrimPrefix(event, "enter "))
			stack = append(stack, frame{function: function})
		case strings.HasPrefix(event, "leave "):
			end(*top)
			stack = stack[:len(stack)-1]
		case strings.HasPrefix(event, "exit "):
			exit = strings.TrimPrefix(event, "exit ")
		case event == "panicked":
			// Nothing after a panic is a path of the program (traceRuntime); the trace records only its
			// exit from here, and a program that panicked isn't held to its calls' ends.
		case strings.HasPrefix(event, "mutated "):
			// The ranges' business (ranges_test.go), not the path's.
		default:
			index, err := strconv.Atoi(event)
			if err != nil {
				problems = append(problems, fmt.Sprintf("a trace line that isn't an event: %q", event))
				continue
			}
			next := marked[index]
			if next.function != top.function {
				problems = append(problems, fmt.Sprintf("a point of function %d ran inside a call of function %d", next.function, top.function))
				continue
			}
			if problem := follows(graphs[top.function], top.last, next.instruction); problem != "" {
				problems = append(problems, fmt.Sprintf("function %d (%s): %s", top.function, graphs[top.function].Name, problem))
			}
			top.last = &next
		}
	}
	if exit == "" {
		return append(problems, "the trace has no exit: Node never finished")
	}
	// A call a panic cut short ends wherever it stood, as the program does, so only a program that
	// finished is held to ending every call where the graph returns.
	if exit == "0" {
		end(stack[0])
		problems = append(problems, ends...)
	}
	return problems
}

// follows says why next can't run right after last (nil: the call's first point), or "".
func follows(graph *Function, last *point, next InstructionId) string {
	if last != nil {
		block, position := locate(graph, last.instruction)
		if position+1 < len(block.Instructions) {
			if block.Instructions[position+1] == next {
				return ""
			}
			return fmt.Sprintf("instruction %d ran after %d, which bb%d continues with %d", next, last.instruction, block.Id, block.Instructions[position+1])
		}
		if reaches(graph, block, func(candidate *BasicBlock) bool { return candidate.Instructions[0] == next }) {
			return ""
		}
		return fmt.Sprintf("instruction %d ran after %d, and no edge from bb%d leads there", next, last.instruction, block.Id)
	}
	entry, _ := graph.Block(graph.Entry)
	if len(entry.Instructions) > 0 {
		if entry.Instructions[0] == next {
			return ""
		}
	} else if reaches(graph, entry, func(candidate *BasicBlock) bool { return candidate.Instructions[0] == next }) {
		return ""
	}
	return fmt.Sprintf("instruction %d ran first, and the entry doesn't lead there", next)
}

// finishes says why a call whose last point was last can't end there, or "".
func finishes(graph *Function, last *point) string {
	from, _ := graph.Block(graph.Entry)
	if last != nil {
		block, position := locate(graph, last.instruction)
		if position+1 < len(block.Instructions) {
			return fmt.Sprintf("the call ended after instruction %d, in the middle of bb%d", last.instruction, block.Id)
		}
		from = block
	}
	if leaves(from.Terminal) {
		return ""
	}
	if reaches(graph, from, nil) {
		return ""
	}
	return fmt.Sprintf("the call ended at bb%d, which doesn't lead to a return", from.Id)
}

// leaves reports whether a terminal ends the call: a return, or a throw out of the function (a call
// a throw left ends there too, its caller's catch or finally taking it).
func leaves(terminal Terminal) bool {
	switch terminal.(type) {
	case *Return, *Throw:
		return true
	}
	return false
}

// reaches reports whether a block's terminal leads, through blocks that run nothing, to a block
// with instructions that accept takes, or (accept nil) to a return or a throw out.
func reaches(graph *Function, from *BasicBlock, accept func(*BasicBlock) bool) bool {
	seen := map[BlockId]bool{}
	queue := []*BasicBlock{from}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		found := false
		EachSuccessor(current.Terminal, func(id BlockId) {
			if found || seen[id] {
				return
			}
			seen[id] = true
			successor, ok := graph.Block(id)
			if !ok {
				return
			}
			if len(successor.Instructions) > 0 {
				found = accept != nil && accept(successor)
				return
			}
			if leaves(successor.Terminal) && accept == nil {
				found = true
				return
			}
			queue = append(queue, successor)
		})
		if found {
			return true
		}
	}
	return false
}

// locate is the block an instruction is in, and its position there.
func locate(graph *Function, instruction InstructionId) (*BasicBlock, int) {
	for _, block := range graph.Blocks {
		for position, id := range block.Instructions {
			if id == instruction {
				return block, position
			}
		}
	}
	panic(fmt.Sprintf("flow: instruction %d is in no block", instruction))
}

// Liveness holds on every path Node takes: wherever, in a call's sequence of points, the next thing
// to touch a variable after an instruction is a read of it, the variable is live after that
// instruction. Reuse in place takes a dead value's memory, so a variable called dead that is read
// again is an object rewritten under a reader.
func TestLivenessHoldsOnEveryPath(t *testing.T) {
	t.Parallel()
	checked := 0
	var lock sync.Mutex
	t.Run("programs", func(t *testing.T) {
		for _, path := range programs(t) {
			t.Run(path, func(t *testing.T) {
				t.Parallel()
				run := traced(t, path)
				live := map[int]map[InstructionId]map[DeclarationId]bool{}
				for function, graph := range run.graphs {
					live[function] = LiveOut(graph)
				}
				count := 0
				for _, sequence := range frames(run) {
					graph := run.graphs[sequence.function]
					throwers := throwingInstructions(graph)
					// Backward: readNext says whether the next thing to touch a variable is a read.
					readNext := map[DeclarationId]bool{}
					for position := len(sequence.points) - 1; position >= 0; position-- {
						id := sequence.points[position]
						for variable, read := range readNext {
							count++
							if read && !live[sequence.function][id][variable] {
								t.Errorf("function %d (%s): a variable (declaration %d) is read after instruction %d, where liveness says it's dead", sequence.function, graph.Name, variable, id)
							}
						}
						instruction := graph.Instructions[id]
						if !threw(graph, throwers, sequence.points, position) {
							for _, define := range instruction.Defines {
								readNext[graph.Identifiers[define.Identifier].Declaration] = false
							}
						}
						for _, use := range instruction.Uses {
							readNext[graph.Identifiers[use.Identifier].Declaration] = true
						}
					}
				}
				lock.Lock()
				checked += count
				lock.Unlock()
			})
		}
	})
	if checked == 0 {
		t.Errorf("nothing was checked")
	}
	t.Logf("%d variable-and-point pairs checked", checked)
}

// throwingInstructions is, for each instruction that ends a block by maybe throwing, its terminal.
func throwingInstructions(graph *Function) map[InstructionId]*MayThrow {
	throwers := map[InstructionId]*MayThrow{}
	for _, block := range graph.Blocks {
		if throws, ok := block.Terminal.(*MayThrow); ok && len(block.Instructions) > 0 {
			throwers[block.Instructions[len(block.Instructions)-1]] = throws
		}
	}
	return throwers
}

// threw reports whether the point at position, when it can throw, may have on this run: the point
// after it is one its handler leads to, or one its next doesn't, or there is none (the call left by
// a throw). A throw never gives the instruction's variables their values. When both ways lead to the
// same point, either may have happened, and liveness must hold for the throw too.
func threw(graph *Function, throwers map[InstructionId]*MayThrow, points []InstructionId, position int) bool {
	throws, ok := throwers[points[position]]
	if !ok {
		return false
	}
	if position+1 == len(points) {
		return true
	}
	next := points[position+1]
	return leadsTo(graph, throws.Handler, next) || !leadsTo(graph, throws.Next, next)
}

// leadsTo reports whether control entering a block runs next first: the block's own first
// instruction, or, through blocks that run nothing, the first of a block after it.
func leadsTo(graph *Function, id BlockId, next InstructionId) bool {
	block, ok := graph.Block(id)
	if !ok {
		return false
	}
	if len(block.Instructions) > 0 {
		return block.Instructions[0] == next
	}
	return reaches(graph, block, func(candidate *BasicBlock) bool { return candidate.Instructions[0] == next })
}

// sequence is one call's points, in the order they ran.
type sequence struct {
	function int
	points   []InstructionId
}

// frames splits a trace into its calls' sequences.
func frames(run run) []sequence {
	var done []sequence
	stack := []sequence{{function: -1}}
	for _, event := range run.events {
		switch {
		case strings.HasPrefix(event, "enter "):
			function, _ := strconv.Atoi(strings.TrimPrefix(event, "enter "))
			stack = append(stack, sequence{function: function})
		case strings.HasPrefix(event, "leave "):
			done = append(done, stack[len(stack)-1])
			stack = stack[:len(stack)-1]
		default:
			if index, err := strconv.Atoi(event); err == nil {
				top := &stack[len(stack)-1]
				top.points = append(top.points, run.marked[index].instruction)
			}
		}
	}
	return append(done, stack...)
}
