package lower

import (
	"fmt"
	"reflect"

	"github.com/system-inc/adamic/internal/ir"
)

// allocationFlowGraph is the shared may-flow graph used by graph regions and
// shape certification. It never changes ownership when queried.
type allocationFlowGraph struct {
	projection      *allocationProjection
	projectionReads map[int]ir.AllocationSet
	projectionBusy  map[int]bool
	// projectionFrames memoizes a projected query that isn't a bare Read, keyed by the expression, its
	// depth and the Reads busy around it (shape_flow.go's projectionFrame).
	projectionFrames  map[string]ir.AllocationSet
	program           *ir.Program
	sources           map[int][]ir.Expression
	unknownParameters map[int]bool
	// entry, when set, names a value an expression is where it enters a view: the flow stops there
	// and reaches the site entry gives, which no allocation has (view_widened.go).
	entry func(ir.Expression) int
}

func (graph *allocationFlowGraph) resultNode(function int) int {
	return len(graph.program.Locals) + function + 1
}

func newAllocationFlowGraph(program *ir.Program) *allocationFlowGraph {
	return collectAllocationFlowGraph(program, nil)
}

// newIterableViewFlowGraph is the graph the iterable view proof reads (checkIterableViews): the
// shared graph, but a nested function onlyCalledDirectly holds for takes its parameters from its
// calls' edges alone, as a module function does from its ir.Calls, since no call outside those edges
// can run it (#qk8rztp S3c-1). The exemption is the view proof's alone (Compiler's scope, which Lower
// took on #qk8rztp S3c): graph regions, shapes, library_spread and the other view passes build theirs
// with newAllocationFlowGraph and read what they read before. Moving it into the shared graph is a
// head of its own, with each consumer's delta and #0d8s5ff's inversion test.
func newIterableViewFlowGraph(program *ir.Program) *allocationFlowGraph {
	return collectAllocationFlowGraph(program, closureCallsOf(program))
}

// collectAllocationFlowGraph collects the graph's edges. calls, when given, says which nested
// functions are only ever called directly, whose closures the MakeClosure arm then leaves alone.
func collectAllocationFlowGraph(program *ir.Program, calls *closureCalls) *allocationFlowGraph {
	graph := &allocationFlowGraph{program: program, sources: map[int][]ir.Expression{}, unknownParameters: map[int]bool{}}
	collect := func(body []ir.Statement, function int) {
		walk(body, func(node any) bool {
			switch value := node.(type) {
			case ir.BoundClosure:
				for _, parameter := range program.Functions[value.Function].Parameters {
					graph.unknownParameters[parameter+1] = true
				}
			case ir.MakeClosure:
				// Callback invocations are outside the collected call edges, unless every call of the
				// closure is one of them (onlyCalledDirectly, the view proof's alone).
				if calls != nil && calls.onlyCalledDirectly(value.Function) {
					break
				}
				for _, parameter := range program.Functions[value.Function].Parameters {
					graph.unknownParameters[parameter+1] = true
				}
			case ir.Declare:
				graph.sources[value.Local+1] = append(graph.sources[value.Local+1], value.Value)
			case ir.Assign:
				graph.sources[value.Local+1] = append(graph.sources[value.Local+1], value.Value)
			case ir.Return:
				if function >= 0 {
					graph.sources[graph.resultNode(function)] = append(graph.sources[graph.resultNode(function)], value.Value)
				}
			case ir.Call:
				for _, target := range program.CallTargets(value) {
					if target < 0 || target >= len(program.Functions) {
						continue
					}
					for i, param := range program.Functions[target].Parameters {
						if i < len(value.Arguments) {
							graph.sources[param+1] = append(graph.sources[param+1], value.Arguments[i])
						} else {
							graph.unknownParameters[param+1] = true
						}
					}
				}
			case ir.CallClosure:
				targets := program.ClosureTargets(value)
				if targets.Unknown {
					// Opaque function values are a documented leak-only frontier.
					break
				}
				for _, target := range targets.Functions {
					for i, param := range program.Functions[target].Parameters {
						if i < len(value.Arguments) {
							graph.sources[param+1] = append(graph.sources[param+1], value.Arguments[i])
						} else {
							graph.unknownParameters[param+1] = true
						}
					}
				}

			}
			return true
		})
	}
	collect(program.Main, -1)
	for i, function := range program.Functions {
		collect(function.Body, i)
	}
	return graph
}

// closureCalls is what onlyCalledDirectly reads off a lowered program, in one walk of it: for each
// function, whether it is one the exemption may take, whether some node runs it outside any call's
// edges, how many times its closure value is used, and how many of those uses are accounted for as
// a direct call of it or its one binding.
type closureCalls struct {
	candidates map[int]bool
	escapes    map[int]bool
	uses       map[int]int
	accounted  map[int]int
}

// onlyCalledDirectly says that every call that can run function is an edge the graph collects, so its
// parameters' sources are its calls' arguments and nothing else (Lower's predicate on #qk8rztp S3c,
// scoped to the view proof by Compiler's ruling). It is #0d8s5ff's concern turned around: the
// MakeClosure arm marks a closure's parameters unknown because something the graph doesn't name might
// call it, and this answers yes only where nothing can, so the inversion #0d8s5ff owes isn't needed
// first. It holds for a function declared inside a function by name, nestedDeclarations' or a generic
// one's instance, that neither suspends nor takes a rest parameter, when:
//
//   - every MakeClosure of it, and every Read of the local its declaration binds it to, is the Closure
//     operand of a CallClosure whose ClosureTargets is exactly that function, never Unknown, with no
//     spread, not a bound closure's kept call and not optional;
//   - or is that binding's own Declare;
//   - every ClosureSelf in its body is the carrier of a CallClosure whose targets are bounded: its own
//     recursion, or a sibling's call through the group's shared cells, each an edge the graph
//     collects;
//   - the binding is written nowhere but its Declare;
//   - and nothing else names it as code to run: no BoundClosure, BoundMethod, method table, getter or
//     sort comparator.
//
// Any other node holding one of those values, an Effects or an argument among them, is a use not
// accounted for, so the answer is no and the parameters stay unknown, today's answer, which fails
// closed. An arrow or a function expression keeps today's answer too: a program arrow kept in a
// local stays a function value whose callers the proof doesn't see (iterable_views_test.go).
func (calls *closureCalls) onlyCalledDirectly(function int) bool {
	return calls.candidates[function] && !calls.escapes[function] && calls.uses[function] == calls.accounted[function]
}

// closureCallsOf walks program once for onlyCalledDirectly. A nested declaration's binding is the
// Declare of its local to its MakeClosure, which nestedDeclarations' prologue makes.
func closureCallsOf(program *ir.Program) *closureCalls {
	calls := &closureCalls{candidates: map[int]bool{}, escapes: map[int]bool{}, uses: map[int]int{}, accounted: map[int]int{}}
	for index, function := range program.Functions {
		calls.candidates[index] = function.NestedParent != 0 && function.RestElement == 0 && !function.Suspends()
	}
	bindings := map[int]int{}
	bound := func(node any) bool {
		if declared, isDeclare := node.(ir.Declare); isDeclare {
			if made, isMade := declared.Value.(ir.MakeClosure); isMade && program.Locals[declared.Local].NestedFunction == made.Function+1 {
				bindings[declared.Local] = made.Function
			}
		}
		return true
	}
	walk(program.Main, bound)
	for _, function := range program.Functions {
		walk(function.Body, bound)
	}
	writes := map[int]int{}
	visit := func(function int) func(node any) bool {
		return func(node any) bool {
			localWrites(node, func(local int) {
				if _, isBinding := bindings[local]; isBinding {
					writes[local]++
				}
			})
			switch value := node.(type) {
			case ir.MakeClosure:
				calls.uses[value.Function]++
			case ir.Read:
				if target, isBinding := bindings[value.Local]; isBinding {
					calls.uses[target]++
				}
			case ir.ClosureSelf:
				if function >= 0 {
					calls.uses[function]++
				}
			case ir.Declare:
				if made, isMade := value.Value.(ir.MakeClosure); isMade {
					if target, isBinding := bindings[value.Local]; isBinding && target == made.Function {
						calls.accounted[target]++
					}
				}
			case ir.CallClosure:
				targets := program.ClosureTargets(value)
				if (len(value.Spread) != 0 || value.Bound || value.Optional) && !targets.Unknown {
					// Its edges feed parameters by position, which a spread doesn't keep.
					for _, target := range targets.Functions {
						calls.escapes[target] = true
					}
				}
				switch closure := value.Closure.(type) {
				case ir.ClosureSelf:
					if function >= 0 && !targets.Unknown {
						calls.accounted[function]++
					}
				case ir.MakeClosure:
					if !targets.Unknown && len(targets.Functions) == 1 && targets.Functions[0] == closure.Function {
						calls.accounted[closure.Function]++
					}
				case ir.Read:
					if target, isBinding := bindings[closure.Local]; isBinding && !targets.Unknown && len(targets.Functions) == 1 && targets.Functions[0] == target {
						calls.accounted[target]++
					}
				}
			case ir.BoundClosure:
				calls.escapes[value.Function] = true
			case ir.BoundMethod:
				calls.escapes[value.Factory] = true
			case ir.ObjectLiteral:
				for _, method := range value.Methods {
					calls.escapes[method.Function] = true
				}
				for _, getter := range value.SpreadGetters {
					calls.escapes[getter] = true
				}
			case ir.ArraySort:
				if value.Callback == nil {
					// A named comparator, run by the sort with no closure of its own.
					for _, target := range program.ClosureTargets(value).Functions {
						calls.escapes[target] = true
					}
				}
			}
			return true
		}
	}
	walk(program.Main, visit(-1))
	for index, function := range program.Functions {
		walk(function.Body, visit(index))
	}
	for local, function := range bindings {
		if writes[local] != 1 {
			calls.escapes[function] = true
		}
	}
	return calls
}

func (graph *allocationFlowGraph) follow(expression ir.Expression, allocation func(int), demand func(int), unknown func(string)) {
	program := graph.program
	follow := func(value ir.Expression) { graph.follow(value, allocation, demand, unknown) }
	if expression == nil {
		unknown("missing expression")
		return
	}
	value := reflect.ValueOf(expression)
	if field := value.FieldByName("GraphTypes"); field.IsValid() {
		// Ownership metadata also labels collection operations. Views only
		// prove the literal layouts indexed by their projection analysis.
		switch expression.(type) {
		case ir.ObjectLiteral, ir.ArrayLiteral:
		default:
			unknown(fmt.Sprintf("untracked allocation source %T", expression))
			return
		}
		ids := field.Interface().([]int)
		if len(ids) == 0 || ids[len(ids)-1] >= 0 {
			unknown("allocation has no site identity")
		} else {
			allocation(ids[len(ids)-1])
		}
		return
	}
	switch value := expression.(type) {
	case ir.Read:
		demand(value.Local + 1)
	case ir.Call:
		if len(program.CallTargets(value)) == 0 {
			unknown("call has no tracked target")
		}
		for _, target := range program.CallTargets(value) {
			demand(graph.resultNode(target))
		}
	case ir.CallClosure:
		targets := program.ClosureTargets(value)
		if targets.Unknown {
			unknown("opaque closure")
			return
		}
		for _, target := range targets.Functions {
			demand(graph.resultNode(target))
		}
	case ir.Conditional:
		follow(value.WhenTrue)
		follow(value.WhenNot)
	case ir.Box:
		follow(value.Value)
	case ir.Narrow:
		follow(value.Value)
	case ir.Unwrap:
		follow(value.Value)
	case ir.CheckedCast:
		follow(value.Value)
	default:
		unknown(fmt.Sprintf("untracked source %T", expression))
	}
}

// assignAllocationSites uses graph regions' numbering for programs without cycle
// seeds too. Existing numbered programs are left unchanged.
func assignAllocationSites(program *ir.Program) {
	numbered := false
	inspect := func(node any) bool {
		value := reflect.ValueOf(node)
		if value.IsValid() && value.Kind() == reflect.Struct {
			if field := value.FieldByName("GraphTypes"); field.IsValid() {
				ids := field.Interface().([]int)
				numbered = numbered || len(ids) > 0 && ids[len(ids)-1] < 0
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if numbered {
		return
	}
	next := 0
	program.Main = graphAllocationSites(reflect.ValueOf(program.Main), &next).Interface().([]ir.Statement)
	for i := range program.Functions {
		program.Functions[i].Body = graphAllocationSites(reflect.ValueOf(program.Functions[i].Body), &next).Interface().([]ir.Statement)
	}
}
