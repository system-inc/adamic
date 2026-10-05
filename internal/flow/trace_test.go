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
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range programs(t) {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			program := lowered(t, path)
			graphs := map[int]*Function{}
			points := map[*ir.Statement]map[int]point{}
			for function := -1; function < len(program.Functions); function++ {
				graph := Build(program, function)
				graphs[function] = graph
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
			// Each mark pushes its point's index in this list, or -1 for a point the graph doesn't
			// have, which Node must then never reach.
			var marked []point
			options := javascript.Options{
				Mark: func(at *ir.Statement, part int) string {
					found, ok := points[at][part]
					if !ok {
						return "adamicTrace.push('missing')"
					}
					marked = append(marked, found)
					return fmt.Sprintf("adamicTrace.push(%d)", len(marked)-1)
				},
				Enter: func(function int) string { return fmt.Sprintf("adamicTrace.push('enter %d')", function) },
				Leave: func(function int) string { return fmt.Sprintf("adamicTrace.push('leave %d')", function) },
			}
			directory := t.TempDir()
			trace := filepath.Join(directory, "trace.txt")
			source := fmt.Sprintf("import { writeFileSync as adamicWrite } from 'node:fs';\nconst adamicTrace = [];\nprocess.on('exit', (code) => adamicWrite(%s, [...adamicTrace, `exit ${code}`].join('\\n')));\n%s",
				strconv.Quote(trace), javascript.JavaScriptWith(program, options))
			module := filepath.Join(directory, "program.mjs")
			if err := os.WriteFile(module, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
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
			events := strings.Split(string(contents), "\n")
			for _, problem := range walk(graphs, marked, events) {
				t.Error(problem)
			}
			// Every program runs something; a trace without a point would pass every check here.
			walked := 0
			for _, event := range events {
				if _, err := strconv.Atoi(event); err == nil {
					walked++
				}
			}
			if walked == 0 {
				t.Errorf("Node ran no point at all: %q", contents)
			}
			t.Logf("%d points walked, %d events", walked, len(events))
		})
	}
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
