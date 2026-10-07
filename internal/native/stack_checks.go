package native

import "github.com/system-inc/adamic/internal/ir"

// stackCheckCandidates finds recursive components and conservatively includes every
// function when a function-value call is unresolved. These are placement candidates,
// not permission to omit checks: unchecked descendant frames still need a proven
// bound for each build configuration. Until that proof is available, functionBody
// continues checking every function.
func stackCheckCandidates(program *ir.Program) []bool {
	graph := make([][]int, len(program.Functions))
	unresolved := false
	e := &emitter{program: program}
	visit := func(caller int, expression ir.Expression) {
		var targets []int
		switch call := expression.(type) {
		case ir.Call:
			targets = program.CallTargets(call)
		case ir.CallClosure:
			if property, ok := call.Closure.(ir.Property); ok && property.Method {
				if target, known := e.exactReceiverMethod(property.Object, property.Name); known {
					targets = []int{target}
					break
				}
			}
			bounded := program.ClosureTargets(call)
			unresolved = unresolved || bounded.Unknown
			targets = bounded.Functions
		case ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.MapForEach, ir.ArraySort:
			bounded := program.ClosureTargets(call)
			unresolved = unresolved || bounded.Unknown
			targets = bounded.Functions
		}
		for _, target := range targets {
			if target < 0 || target >= len(graph) {
				panic("native: stack call target outside program")
			}
			if caller >= 0 {
				graph[caller] = append(graph[caller], target)
			}
		}
	}
	body := func(caller int, list []ir.Statement) {
		var statements func([]ir.Statement)
		statements = func(list []ir.Statement) {
			for _, statement := range list {
				walkStatement(statement, func(expression ir.Expression) {
					visit(caller, expression)
				}, statements)
			}
		}
		statements(list)
	}
	for caller, function := range program.Functions {
		body(caller, function.Body)
	}
	body(-1, program.Main)
	if unresolved {
		// Unknown is anything in the program, not an empty target set. Keeping
		// all functions also covers everything reached from such a target.
		checks := make([]bool, len(graph))
		for function := range checks {
			checks[function] = true
		}
		return checks
	}
	return recursiveStackComponents(graph)
}

// recursiveStackComponents uses Tarjan's low links. A component recurses when
// it has more than one member, or its sole member has a self edge.
func recursiveStackComponents(graph [][]int) []bool {
	checks := make([]bool, len(graph))
	numbers := make([]int, len(graph))
	low := make([]int, len(graph))
	active := make([]bool, len(graph))
	stack := []int{}
	next := 0
	var visit func(int)
	visit = func(function int) {
		next++
		numbers[function], low[function] = next, next
		stack = append(stack, function)
		active[function] = true
		self := false
		for _, target := range graph[function] {
			self = self || target == function
			if numbers[target] == 0 {
				visit(target)
				low[function] = min(low[function], low[target])
			} else if active[target] {
				low[function] = min(low[function], numbers[target])
			}
		}
		if low[function] != numbers[function] {
			return
		}
		first := len(stack) - 1
		for stack[first] != function {
			first--
		}
		recursive := len(stack)-first > 1 || self
		for _, member := range stack[first:] {
			checks[member] = recursive
			active[member] = false
		}
		stack = stack[:first]
	}
	for function := range graph {
		if numbers[function] == 0 {
			visit(function)
		}
	}
	return checks
}
