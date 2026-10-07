package lower

import (
	"math"

	"github.com/system-inc/adamic/internal/ir"
)

const maximumStringLength = 536870888

// guardRuntimeRanges inserts ordinary throws before runtime points that cannot
// unwind. Only functions reachable from a try need a catchable stack check.
// String arguments are evaluated once, before length validation and allocation.
func (l *lowering) guardRuntimeRanges() {
	p := l.result
	if len(l.tries) == 0 {
		return
	}
	// Register the possible guard error before collecting class dispatch and
	// field facts; a newly discovered descendant must not escape either proof.
	instance := l.errorInstance("RangeError")
	l.finishClassCalls()
	stackFailure := ir.ObjectLiteral{Class: instance.class, Methods: instance.methodList(), Fields: []ir.Field{{Name: "name", Value: ir.StringConstant{Index: l.constant("RangeError")}}, {Name: "message", Value: ir.StringConstant{Index: l.constant("Maximum call stack size exceeded")}}, {Name: "cause", Value: ir.Undefined{Of: ir.Union}}}}
	candidates := map[int]bool{}
	for _, closure := range l.closureRecords {
		candidates[closure.function] = true
	}
	for _, instance := range l.instances {
		for _, method := range instance.methodList() {
			candidates[method.Function] = true
		}
	}
	targets := func(body []ir.Statement) []int {
		found := map[int]bool{}
		walk(body, func(node any) bool {
			switch node := node.(type) {
			case ir.Call:
				for _, target := range p.CallTargets(node) {
					found[target] = true
				}
			case ir.Read:
				if node.Checked {
					if error, ok := p.ReadyErrors[node.Local]; ok {
						found[error.Function] = true
					}
				}
			case ir.Assign:
				if node.Checked {
					if error, ok := p.ReadyErrors[node.Local]; ok {
						found[error.Function] = true
					}
				}
			case ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.MapForEach:
				for target := range candidates {
					found[target] = true
				}
			case ir.ArraySort:
				if node.Callback == nil {
					found[node.Comparator] = true
				} else {
					for target := range candidates {
						found[target] = true
					}
				}
			}
			return true
		})
		result := []int{}
		for target := range p.Functions {
			if found[target] {
				result = append(result, target)
			}
		}
		return result
	}
	edges := make([][]int, len(p.Functions))
	for index, function := range p.Functions {
		edges[index] = targets(function.Body)
	}
	reachable := map[int]bool{}
	var reach func(int)
	reach = func(index int) {
		if reachable[index] {
			return
		}
		reachable[index] = true
		for _, next := range edges[index] {
			reach(next)
		}
	}
	for _, record := range l.tries {
		for _, target := range targets(record.body) {
			reach(target)
		}
	}
	recursive := map[int]bool{}
	for start := range reachable {
		seen := map[int]bool{}
		var cycle func(int) bool
		cycle = func(index int) bool {
			if index == start {
				return true
			}
			if seen[index] {
				return false
			}
			seen[index] = true
			for _, next := range edges[index] {
				if cycle(next) {
					return true
				}
			}
			return false
		}
		for _, next := range edges[start] {
			if cycle(next) {
				recursive[start] = true
				break
			}
		}
	}
	bound := l.stringLengthBounds(stackFailure.Fields...)
	guard := func(node any) any {
		switch node := node.(type) {
		case ir.Concat:
			if bound(node) <= maximumStringLength {
				return node
			}
			return l.checkedLibrary(nil, node.Parts, func(args []ir.Expression) (ir.Expression, ir.Expression, ir.Expression) {
				var length ir.Expression = ir.NumberConstant{}
				for _, arg := range args {
					length = ir.Binary{Operator: ir.Add, Left: length, Right: ir.StringLength{Value: arg}}
				}
				return ir.Concat{Parts: args}, ir.Binary{Operator: ir.Greater, Left: length, Right: ir.NumberConstant{Value: maximumStringLength}}, ir.StringConstant{Index: l.constant("Invalid string length")}
			})
		case ir.StringCall:
			if node.Method != "padStart" && node.Method != "padEnd" {
				return node
			}
			if count, ok := node.Arguments[0].(ir.NumberConstant); ok && (math.IsNaN(count.Value) || math.Trunc(count.Value) <= maximumStringLength) {
				return node
			}
			if fill, ok := node.Arguments[1].(ir.StringConstant); ok && len(p.Strings[fill.Index]) == 0 {
				return node
			}
			return l.checkedLibrary(nil, append([]ir.Expression{node.Value}, node.Arguments...), func(args []ir.Expression) (ir.Expression, ir.Expression, ir.Expression) {
				tooLong := ir.Binary{Operator: ir.Greater, Left: ir.MathCall{Function: "trunc", Arguments: []ir.Expression{args[1]}}, Right: ir.NumberConstant{Value: maximumStringLength}}
				nonempty := ir.Binary{Operator: ir.Greater, Left: ir.StringLength{Value: args[2]}, Right: ir.NumberConstant{}}
				return ir.StringCall{Method: node.Method, Value: args[0], Arguments: args[1:]}, ir.Binary{Operator: ir.And, Left: tooLong, Right: nonempty}, ir.StringConstant{Index: l.constant("Invalid string length")}
			})
		}
		return node
	}
	protected := func(node any) any {
		if tried, ok := node.(ir.Try); ok {
			tried.Body = rewriteChecks(tried.Body, guard).([]ir.Statement)
			// A catch's failure must run this try's finally. Its finally has no
			// handler here; an enclosing protected body guards it if needed.
			if tried.HasFinally {
				tried.Catch = rewriteChecks(tried.Catch, guard).([]ir.Statement)
			}
			return tried
		}
		return node
	}
	// Capture the original function count: helpers are already guarded and must
	// not recursively wrap their own operation.
	for index := 0; index < len(edges); index++ {
		if p.Functions[index].LibraryGuarded {
			// Its existing library operation is already checked.
		} else if reachable[index] {
			p.Functions[index].Body = rewriteChecks(p.Functions[index].Body, guard).([]ir.Statement)
		} else {
			p.Functions[index].Body = rewriteChecks(p.Functions[index].Body, protected).([]ir.Statement)
		}

	}
	// A top-level try has no function entry. Rewrite its protected body, including
	// nested handlers, without charging unrelated top-level string operations.
	p.Main = rewriteChecks(p.Main, protected).([]ir.Statement)
	// Rebuild after adding range helpers: their prologues and newly introduced
	// error constructors must unwind too when called on the last recursive frame.
	finalEdges := make([][]int, len(p.Functions))
	for index, function := range p.Functions {
		finalEdges[index] = targets(function.Body)
	}
	// An acyclic callee can cross the margin before the next recursive entry.
	// Guard every descendant too, including library wrappers and allocators;
	// otherwise a helper's fatal prologue could still bypass the handler.
	stackGuarded := map[int]bool{}
	var guardDescendants func(int)
	guardDescendants = func(index int) {
		if stackGuarded[index] {
			return
		}
		stackGuarded[index] = true
		for _, next := range finalEdges[index] {
			guardDescendants(next)
		}
	}
	for index := range recursive {
		guardDescendants(index)
	}

	for index := range p.Functions {
		if stackGuarded[index] {
			p.Functions[index].StackGuarded = true
			p.Functions[index].Body = append([]ir.Statement{ir.If{Condition: ir.StackExceeded{}, Then: []ir.Statement{ir.Throw{Value: stackFailure}}}}, p.Functions[index].Body...)
		}
	}

}

// stringLengthBounds uses all closed-program stores and direct-call arguments.
// Cycles, unknown producers and indirect parameters use the runtime limit, never
// an empty initial fact. UTF-8 bytes conservatively bound UTF-16 units.
func (l *lowering) stringLengthBounds(additional ...ir.Field) func(ir.Expression) float64 {
	p := l.result
	stores := map[int][]ir.Expression{}
	fields := map[string][]ir.Expression{}
	// Future stack throws are not in the body yet. Include their actual fields
	// so a catch cannot prove a join from only earlier, shorter error messages.
	for _, field := range additional {
		fields[field.Name] = append(fields[field.Name], field.Value)
	}

	unknownFields := false
	indirect := map[int]bool{}
	for _, function := range p.Functions {
		if function.Closure {
			for _, local := range function.Parameters {
				indirect[local] = true
			}
		}
	}
	for _, instance := range l.instances {
		for _, method := range instance.methodList() {
			for _, local := range p.Functions[method.Function].Parameters {
				indirect[local] = true
			}
		}
	}
	note := func(node any) bool {
		switch node := node.(type) {
		case ir.Declare:
			stores[node.Local] = append(stores[node.Local], node.Value)
		case ir.Assign:
			stores[node.Local] = append(stores[node.Local], node.Value)
		case ir.Call:
			for _, target := range p.CallTargets(node) {
				for index, parameter := range p.Functions[target].Parameters {
					if index < len(node.Arguments) {
						stores[parameter] = append(stores[parameter], node.Arguments[index])
					}
				}
			}
		case ir.ObjectLiteral:
			if node.Spread != nil {
				unknownFields = true
			}
			for _, field := range append(append([]ir.Field{}, node.Fields...), node.Empty...) {
				fields[field.Name] = append(fields[field.Name], field.Value)
			}
		case ir.SetProperty:
			fields[node.Name] = append(fields[node.Name], node.Value)
		case ir.ObjectCall:
			if node.Method == "assign" || (node.Returns == ir.Object && node.Method != "freeze") {
				unknownFields = true
			}
		case ir.Read, ir.Property, ir.CallClosure, ir.ArrayIndex, ir.MapGet, ir.WeakTarget,
			ir.Defined, ir.Narrow, ir.Unwrap, ir.Conditional, ir.Coalesce, ir.Undefined, ir.Null:
			// Aliases of objects already accounted for by their producers.
		default:
			if value, ok := node.(ir.Expression); ok && value.Type() == ir.Object {
				unknownFields = true
			}
		}
		return true
	}
	// Uncalled constructors have unconstrained parameters but cannot store into
	// a live object. Follow the closed call graph instead of letting those
	// unused bodies poison every field with an artificial unknown store.
	used := map[int]bool{}
	var visitBody func([]ir.Statement)
	var visitFunction func(int)
	visitFunction = func(index int) {
		if used[index] {
			return
		}
		used[index] = true
		visitBody(p.Functions[index].Body)
	}
	visitValues := func() {
		for index, function := range p.Functions {
			if function.Closure {
				visitFunction(index)
			}
		}
		for _, closure := range l.closureRecords {
			visitFunction(closure.function)
		}
		for _, instance := range l.instances {
			for _, method := range instance.methodList() {
				visitFunction(method.Function)
			}
		}
	}
	visitBody = func(body []ir.Statement) {
		walk(body, func(node any) bool {
			switch node := node.(type) {
			case ir.Call:
				for _, target := range p.CallTargets(node) {
					visitFunction(target)
				}
			case ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.MapForEach:
				visitValues()
			case ir.ArraySort:
				if node.Callback == nil {
					visitFunction(node.Comparator)
				} else {
					visitValues()
				}
			}
			return true
		})
	}
	visitBody(p.Main)
	for _, error := range p.ReadyErrors {
		visitFunction(error.Function)
		walk(error, note)
	}
	walk(p.Main, note)
	for index, function := range p.Functions {
		if used[index] {
			walk(function.Body, note)
		}
	}
	var bound func(ir.Expression, map[int]bool, int) float64
	bound = func(value ir.Expression, seen map[int]bool, depth int) float64 {
		if value == nil || depth > 24 {
			return maximumStringLength
		}
		maximum := func(values []ir.Expression) float64 {
			result := float64(0)
			if len(values) == 0 {
				return maximumStringLength
			}
			for _, value := range values {
				result = math.Max(result, bound(value, seen, depth+1))
			}
			return result
		}
		switch value := value.(type) {
		case ir.StringConstant:
			return float64(len(p.Strings[value.Index]))
		case ir.Undefined:
			return 0
		case ir.NumberToString:
			return 32
		case ir.BooleanToString:
			return 5
		case ir.ToFixed:
			return 330
		case ir.NumberFormat:
			return 1100
		case ir.Read:
			if seen[value.Local] || indirect[value.Local] {
				return maximumStringLength
			}
			seen[value.Local] = true
			result := maximum(stores[value.Local])
			delete(seen, value.Local)
			return result
		case ir.Property:
			if !unknownFields {
				return maximum(fields[value.Name])
			}
		case ir.Concat:
			result := float64(0)
			for _, part := range value.Parts {
				result += bound(part, seen, depth+1)
			}
			return result
		case ir.Conditional:
			return math.Max(bound(value.WhenTrue, seen, depth+1), bound(value.WhenNot, seen, depth+1))
		case ir.Coalesce:
			return math.Max(bound(value.Value, seen, depth+1), bound(value.Fallback, seen, depth+1))
		case ir.StringCall:
			input := bound(value.Value, seen, depth+1)
			switch value.Method {
			case "repeat":
				if count, ok := value.Arguments[0].(ir.NumberConstant); ok && count.Value >= 0 && !math.IsInf(count.Value, 0) {
					return math.Min(maximumStringLength, input*math.Trunc(count.Value))
				}
			case "slice", "substring", "trim", "trimStart", "trimEnd":
				return input
			case "padStart", "padEnd":
				if count, ok := value.Arguments[0].(ir.NumberConstant); ok && !math.IsNaN(count.Value) {
					return math.Min(maximumStringLength, math.Max(input, math.Trunc(count.Value)))
				}
			}
		case ir.Call:
			values := []ir.Expression{}
			for _, target := range p.CallTargets(value) {
				walk(p.Functions[target].Body, func(node any) bool {
					if returned, ok := node.(ir.Return); ok {
						values = append(values, returned.Value)
					}
					return true
				})
			}
			return maximum(values)
		}
		return maximumStringLength
	}
	return func(value ir.Expression) float64 { return bound(value, map[int]bool{}, 0) }
}
