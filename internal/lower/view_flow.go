package lower

import (
	"fmt"
	"reflect"
	"slices"

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
	// running is whether each function could run when the graph was built (runningFunctions): one
	// that couldn't had its writes, returns and calls left out, so checkFlowLiveness holds it to
	// that at the end of lowering.
	running []bool
}

func (graph *allocationFlowGraph) resultNode(function int) int {
	return len(graph.program.Locals) + function + 1
}

func newAllocationFlowGraph(program *ir.Program) *allocationFlowGraph {
	graph := &allocationFlowGraph{program: program, sources: map[int][]ir.Expression{}, unknownParameters: map[int]bool{}}
	// pass hands a call's arguments to its target's parameters, by position. An omitted argument is
	// undefined where the parameter is optional or has a default: OptionalParameters is keyed by the
	// local the argument arrives in, a default's incoming local among them (functions.go's
	// signature), and the default itself is the Coalesce or Conditional the body declares over it.
	// Undefined reaches no allocation, and a parameter that only ever gets it still reaches none. A
	// required or rest parameter's omitted argument stays unknown, and so does any omitted argument
	// where a spread may move what lands at that position.
	pass := func(target int, arguments []ir.Expression, spread []bool) {
		function := &program.Functions[target]
		for i, param := range function.Parameters {
			switch {
			case i < len(arguments):
				graph.sources[param+1] = append(graph.sources[param+1], arguments[i])
			case function.OptionalParameters[param] && !slices.Contains(spread, true):
				graph.sources[param+1] = append(graph.sources[param+1], ir.Undefined{Of: program.Locals[param].Type})
			default:
				graph.unknownParameters[param+1] = true
			}
		}
	}
	collect := func(body []ir.Statement, function int) {
		walk(body, func(node any) bool {
			switch value := node.(type) {
			case ir.BoundClosure:
				for _, parameter := range program.Functions[value.Function].Parameters {
					graph.unknownParameters[parameter+1] = true
				}
			case ir.MakeClosure:
				// Callback invocations are outside the collected call edges.
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
					pass(target, value.Arguments, value.Spread)
				}
			case ir.CallClosure:
				targets := program.ClosureTargets(value)
				if targets.Unknown {
					// Opaque function values are a documented leak-only frontier.
					break
				}
				for _, target := range targets.Functions {
					pass(target, value.Arguments, value.Spread)
				}

			}
			return true
		})
	}
	collect(program.Main, -1)
	graph.running = runningFunctions(program)
	for i, function := range program.Functions {
		// A function that never runs makes none of its writes, returns or calls.
		if !graph.running[i] {
			continue
		}
		collect(function.Body, i)
	}
	// Outside code calls the entry module's exported functions (the JavaScript backend's
	// ModuleExports, wasm32's handleRequest) with anything, so every parameter of one is unknown,
	// rest included, whatever the program's own calls pass it. An export held in a global (Function
	// is -1) is a function value, whose parameters MakeClosure already marks.
	for _, export := range program.Exports {
		if export.Function < 0 || export.Function >= len(program.Functions) {
			continue
		}
		for _, parameter := range program.Functions[export.Function].Parameters {
			graph.unknownParameters[parameter+1] = true
		}
	}
	return graph
}

// flowGraph is newAllocationFlowGraph for a lowering pass: it notes which functions the graph took to
// never run, so checkFlowLiveness can hold lowering to that once nothing more changes the program.
func (l *lowering) flowGraph(program *ir.Program) *allocationFlowGraph {
	graph := newAllocationFlowGraph(program)
	for function, running := range graph.running {
		if !running {
			if l.flowDead == nil {
				l.flowDead = map[int]bool{}
			}
			l.flowDead[function] = true
		}
	}
	return graph
}

// checkFlowLiveness reads, once nothing more changes the program, which functions can run, and
// fails if one a flow graph took to never run (dead, by index) is named now. Liveness is read when a
// graph is built, and a pass after that adding the first reference to a function would leave that
// function's writes out of what the graph's consumers were told: a missed writer, so a compiler bug,
// never a program to emit. A function made after the graph isn't in dead; its body was never in
// that graph at all, as before road 3.
func checkFlowLiveness(program *ir.Program, dead map[int]bool) error {
	running := runningFunctions(program)
	for function := range running {
		if dead[function] && running[function] {
			return fmt.Errorf("lower: compiler bug: function %d was named after the flow graph was built", function)
		}
	}
	return nil
}

// functionReferenceFields are the IR fields that name a function, as type and field, each with what
// its value is offset by: one for a field that holds an index plus one, zero for one that holds the
// index. A count or a slice of them, or a map whose keys or values are, names every function in it.
// TestFunctionReferenceFieldsAreEveryFunctionField holds every IR field whose name says it may hold
// a function to this table or to its own list of fields that don't.
var functionReferenceFields = map[string]int{
	// Calls: a direct or virtual call's function (a virtual call's targets are MethodTargets'),
	// a sibling call's code, a sort's named comparator, a generator's step.
	"Call.Function":          0,
	"CallClosure.Direct":     1,
	"ArraySort.Comparator":   0,
	"GeneratorNext.Function": 0,
	"IteratorNext.Function":  0,
	// Function values and method tables: each is a road an opaque call can run it by.
	"MakeClosure.Function":           0,
	"BoundClosure.Function":          0,
	"BoundMethod.Factory":            0,
	"Method.Function":                0,
	"ObjectLiteral.SpreadGetters":    0,
	"Class.Constructor":              0,
	"Class.Methods":                  0,
	"Accessor.Getter":                0,
	"Accessor.Setter":                0,
	"Local.NestedFunction":           1,
	"Local.ConstantClosure":          1,
	"ViewContract.Functions":         0,
	"Program.MethodTargets":          0,
	"Program.StructuralMethodThunks": 0,
	// Entries outside code calls: the entry module's exports (the JavaScript backend's
	// ModuleExports, wasm32's handleRequest), and the async module body Main moves into.
	"Export.Function":    0,
	"Program.AsyncEntry": 1,
	// Lowering's own links between functions, kept though none of them is a call, since a field
	// that may name a function counts.
	"Function.ForwardsArguments":     1,
	"Function.NestedParent":          1,
	"Function.ForwardedNestedParent": 1,
	"Function.ReferenceParents":      1,
}

// runningFunctions is whether each function can run: whether any field functionReferenceFields
// names, anywhere in the program (Main, every function's body and its own fields, the classes,
// locals, exports, view contracts and the program's tables), names it. One that nothing names is
// no call's target, no function value, no method and no entry, so nothing runs it. An opaque call
// (a CallClosure whose targets are Unknown) runs only a function value that exists, and a function
// value is only ever made by MakeClosure or BoundClosure, or found in a method table, all of them
// counted here, so no opaque call reaches one either. FunctionTypeTargets isn't counted: it is every
// function a checker function type could be by assignability alone, and a function becomes a value
// only through those nodes. A function named only from the body of one that never runs is still
// kept, since the names are read from every body: that keeps more than it must, and is the cheaper
// of the two, with no fixed point to reach.
func runningFunctions(program *ir.Program) []bool {
	running := make([]bool, len(program.Functions))
	var record func(reflect.Value, int)
	record = func(value reflect.Value, offset int) {
		switch value.Kind() {
		case reflect.Int:
			if function := int(value.Int()) - offset; function >= 0 && function < len(running) {
				running[function] = true
			}
		case reflect.Slice, reflect.Array:
			for i := 0; i < value.Len(); i++ {
				record(value.Index(i), offset)
			}
		case reflect.Map:
			entries := value.MapRange()
			for entries.Next() {
				record(entries.Key(), offset)
				record(entries.Value(), offset)
			}
		}
	}
	// plan is, for each struct type met, its fields that name a function and its fields that can hold
	// more of the program, so each field's name is looked up once per type, not once per node.
	type reference struct{ field, offset int }
	type plan struct {
		references []reference
		nested     []int
	}
	plans := map[reflect.Type]*plan{}
	planOf := func(structure reflect.Type) *plan {
		if known, ok := plans[structure]; ok {
			return known
		}
		made := &plan{}
		for i := 0; i < structure.NumField(); i++ {
			if offset, names := functionReferenceFields[structure.Name()+"."+structure.Field(i).Name]; names {
				made.references = append(made.references, reference{field: i, offset: offset})
			}
			if holdsProgram(structure.Field(i).Type) {
				made.nested = append(made.nested, i)
			}
		}
		plans[structure] = made
		return made
	}
	type pointer struct {
		of      reflect.Type
		address uintptr
	}
	seen := map[pointer]bool{}
	var visit func(reflect.Value)
	visit = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface:
			if !value.IsNil() {
				visit(value.Elem())
			}
		case reflect.Pointer:
			if value.IsNil() || seen[pointer{value.Type(), value.Pointer()}] {
				return
			}
			seen[pointer{value.Type(), value.Pointer()}] = true
			visit(value.Elem())
		case reflect.Slice, reflect.Array:
			if holdsProgram(value.Type().Elem()) {
				for i := 0; i < value.Len(); i++ {
					visit(value.Index(i))
				}
			}
		case reflect.Map:
			if holdsProgram(value.Type().Key()) || holdsProgram(value.Type().Elem()) {
				entries := value.MapRange()
				for entries.Next() {
					visit(entries.Key())
					visit(entries.Value())
				}
			}
		case reflect.Struct:
			fields := planOf(value.Type())
			for _, named := range fields.references {
				record(value.Field(named.field), named.offset)
			}
			for _, i := range fields.nested {
				visit(value.Field(i))
			}
		}
	}
	visit(reflect.ValueOf(program).Elem())
	return running
}

// holdsProgram is whether a value of the type can hold a struct, and so a field that names a function.
func holdsProgram(of reflect.Type) bool {
	switch of.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Struct:
		return true
	case reflect.Slice, reflect.Array:
		return holdsProgram(of.Elem())
	case reflect.Map:
		return holdsProgram(of.Key()) || holdsProgram(of.Elem())
	}
	return false
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
	case ir.Undefined, ir.Null:
		// Neither is an allocation, so neither adds a site. A value that's only ever one of them
		// reaches no site, which ReachingAllocations reports as Unknown, never as a proof.
	case ir.Coalesce:
		// Value ?? Fallback is Value when it's present and Fallback otherwise: a default's
		// declaration over its incoming argument (functions.go's lowerBody) is one. A non-null
		// assertion lowers to Value ?? panic (non_null.go), which hands on Value itself or stops,
		// so its Fallback is no value and isn't followed.
		follow(value.Value)
		if value.Panic == nil {
			follow(value.Fallback)
		}
	case ir.Logical:
		// Left || Right and Left && Right give one operand or the other (ir/taste.go): Left where it
		// decides, held as the result's type, which keeps an object as itself, and Right otherwise.
		follow(value.Left)
		follow(value.Right)
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
