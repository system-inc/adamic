package lower

import (
	"fmt"
	"reflect"
	"slices"
	"sync"

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
	// unknownWrites are the locals a node writes with a value the walk can't name, by local plus one,
	// each with why (flowLocalWrites): a for-of's binding takes each element, and a catch's what was
	// thrown. A local that is also assigned keeps this, so a query reaching it is Unknown rather than
	// only what the assignments write.
	unknownWrites map[int]string
	// entry, when set, names a value an expression is where it enters a view: the flow stops there
	// and reaches the site entry gives, which no allocation has (view_widened.go).
	entry func(ir.Expression) int
	// running is whether each function could run when the graph was built (runningFunctions): one
	// that couldn't had its writes, returns and calls left out, so checkFlowLiveness holds it to
	// that at the end of lowering.
	running []bool
}

// unknownAt is why a node's value may come from somewhere the graph doesn't name, if it may: a
// parameter some call the graph can't see may pass anything, and a local a node writes with a value
// the walk can't name (flowLocalWrites) holds that too. Every reader of the graph's sources asks this
// before it trusts them, rather than reading either map.
func (graph *allocationFlowGraph) unknownAt(node int) (string, bool) {
	if graph.unknownParameters[node] {
		return "parameter may receive untracked or omitted arguments", true
	}
	if why, written := graph.unknownWrites[node]; written {
		return why, true
	}
	return "", false
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

// flowGraph is newAllocationFlowGraph for a lowering pass, and iterableViewFlowGraph is
// newIterableViewFlowGraph for one: each notes which functions the graph took to never run
// (noteFlowDead), so checkFlowLiveness can hold lowering to that once nothing more changes the
// program. Every graph lowering builds is built through one of them.
func (l *lowering) flowGraph(program *ir.Program) *allocationFlowGraph {
	return l.noteFlowDead(newAllocationFlowGraph(program))
}

func (l *lowering) iterableViewFlowGraph(program *ir.Program) *allocationFlowGraph {
	return l.noteFlowDead(newIterableViewFlowGraph(program))
}

// noteFlowDead adds the functions graph took to never run to flowDead, and hands graph back.
func (l *lowering) noteFlowDead(graph *allocationFlowGraph) *allocationFlowGraph {
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

// collectAllocationFlowGraph collects the graph's edges. calls, when given, says which nested
// functions are only ever called directly, whose closures the MakeClosure arm then leaves alone.
func collectAllocationFlowGraph(program *ir.Program, calls *closureCalls) *allocationFlowGraph {
	graph := &allocationFlowGraph{program: program, sources: map[int][]ir.Expression{}, unknownParameters: map[int]bool{}, unknownWrites: map[int]string{}}
	// pass hands a call's arguments to its target's parameters, by position. An omitted argument is
	// undefined where the parameter is optional or has a default: OptionalParameters is keyed by the
	// local the argument arrives in, a default's incoming local among them (functions.go's signature),
	// and the default itself is the Coalesce or Conditional the body declares over it. Undefined reaches
	// no allocation, and a parameter that only ever gets it still reaches none. A required or rest
	// parameter's omitted argument stays unknown, and so does any omitted argument where a spread may
	// move what lands at that position.
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
			graph.noteWrites(reflect.ValueOf(node))
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
	// A function that never runs makes none of its writes, returns or calls (runningFunctions). This
	// and calls' exemption agree, so neither needs anything of the other. A call can run a function
	// onlyCalledDirectly holds for only through a MakeClosure of it, a Read of its binding, whose
	// Declare holds that MakeClosure, or a ClosureSelf in its group, each of whose members is made the
	// same way; a MakeClosure names its function, so any function such a call can run is running. A
	// call the exemption accounts for in a body left out here never runs, so the edge it would have
	// collected passes nothing.
	graph.running = runningFunctions(program)
	for i, function := range program.Functions {
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
	mark := func(held int64, offset int) {
		if function := held - int64(offset); function >= 0 && function < int64(len(running)) {
			running[function] = true
		}
	}
	// record reads an index of any integer width, signed or not, so a field whose type is narrower
	// than int, or a named integer type, still names its function.
	var record func(reflect.Value, int)
	record = func(value reflect.Value, offset int) {
		switch value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			mark(value.Int(), offset)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			// Past len(running)+offset it names no function, and it could overflow int64.
			if held := value.Uint(); held <= uint64(len(running)+offset) {
				mark(int64(held), offset)
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
				if (slices.Contains(value.Spread, true) || value.Bound || value.Optional) && !targets.Unknown {
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

// flowLocalWrites is how collect takes every IR field that writes a local, by type and field: empty for
// one whose node's Value is the local's source (Declare and Assign, collect's own cases), and
// otherwise why the local is an unknown write. Its rows are exactly local_uses.go's localFields'
// write rows (TestLocalWritesAreLocalFieldsWrites), so an IR node that comes to write a local is
// classified here before it can leave a local with only some of its writers.
var flowLocalWrites = map[string]string{
	"Declare.Local": "",
	"Assign.Local":  "",
	// Each step writes the element into Local, or destructures it into Pattern's locals, a Binding
	// each, Local unused then. A for-in is a ForOf too, over its object's keys.
	"ForOf.Local":    "a for-of binding takes each element",
	"Binding.Local":  "a for-of binding takes each element",
	"Try.CatchLocal": "a catch binding takes what was thrown",
}

// writeFields is, for each struct type noteWrites has met, which of its fields write a local by a
// row of flowLocalWrites with a reason, and which hold plain structs to look into, worked out once a type.
var writeFields sync.Map // reflect.Type to writeFieldsOf

type writeFieldsOf struct {
	writes []writeField
	inner  []int
}

type writeField struct {
	index int
	name  string
	why   string
}

func writeFieldsFor(of reflect.Type) writeFieldsOf {
	if known, ok := writeFields.Load(of); ok {
		return known.(writeFieldsOf)
	}
	fields := writeFieldsOf{}
	for index := 0; index < of.NumField(); index++ {
		field := of.Field(index)
		name := of.Name() + "." + field.Name
		switch {
		case field.Type.Kind() == reflect.Int && flowLocalWrites[name] != "":
			fields.writes = append(fields.writes, writeField{index: index, name: name, why: flowLocalWrites[name]})
		case field.Type.Kind() == reflect.Struct && !irNode(field.Type),
			field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct && !irNode(field.Type.Elem()):
			fields.inner = append(fields.inner, index)
		}
	}
	writeFields.Store(of, fields)
	return fields
}

// noteWrites records each local the fields of value write by a row of flowLocalWrites with a reason,
// looking into the plain structs it holds (a ForOf's Bindings) and leaving the statements and
// expressions it holds to walk. A ForOf with a pattern leaves Local unused, and a Try's CatchLocal is
// -1 when the catch binds nothing, as localUseCount reads them.
func (graph *allocationFlowGraph) noteWrites(value reflect.Value) {
	if value.Kind() != reflect.Struct {
		return
	}
	fields := writeFieldsFor(value.Type())
	if len(fields.writes) == 0 && len(fields.inner) == 0 {
		return
	}
	pattern := value.Type() == reflect.TypeOf(ir.ForOf{}) && value.FieldByName("Pattern").Len() != 0
	for _, write := range fields.writes {
		local := int(value.Field(write.index).Int())
		if write.name == "ForOf.Local" && pattern || local < 0 {
			continue
		}
		graph.unknownWrites[local+1] = write.why
	}
	for _, index := range fields.inner {
		field := value.Field(index)
		if field.Kind() == reflect.Struct {
			graph.noteWrites(field)
			continue
		}
		for element := 0; element < field.Len(); element++ {
			graph.noteWrites(field.Index(element))
		}
	}
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
		// x || y and x && y each give back one operand or the other (ir/taste.go), so the value reaches what
		// either does. Where lowering converts the left operand to the result's type, an object is held as
		// itself; the fresh values that conversion makes are scalars and falsy lefts, which reach no site.
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
