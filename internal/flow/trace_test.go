package flow

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
const adamicFrames = [{ last: -1, held: new Map() }];
const adamicEnter = () => { adamicFrames.push({ last: -1, held: new Map() }); };
const adamicLeave = () => { adamicFrames.pop(); };
const adamicPoint = (point, variables) => {
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
				return "adamicTrace.push('missing')"
			}
			result.marked = append(result.marked, found)
			return fmt.Sprintf("adamicPoint(%d, %s)", len(result.marked)-1, variables[found.function])
		},
		Enter: func(function int) string {
			return fmt.Sprintf("(adamicEnter(), adamicTrace.push('enter %d'))", function)
		},
		Leave: func(function int) string {
			return fmt.Sprintf("(adamicLeave(), adamicTrace.push('leave %d'))", function)
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
	if _, returns := from.Terminal.(*Return); returns {
		return ""
	}
	if reaches(graph, from, nil) {
		return ""
	}
	return fmt.Sprintf("the call ended at bb%d, which doesn't lead to a return", from.Id)
}

// reaches reports whether a block's terminal leads, through blocks that run nothing, to a block
// with instructions that accept takes, or (accept nil) to a return.
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
			if _, returns := successor.Terminal.(*Return); returns && accept == nil {
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
