package lower

import (
	"fmt"
	"reflect"
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
	// each with why (localWrites): a for-of's binding takes each element, and a catch's what was
	// thrown. A local that is also assigned keeps this, so a query reaching it is Unknown rather than
	// only what the assignments write.
	unknownWrites map[int]string
	// entry, when set, names a value an expression is where it enters a view: the flow stops there
	// and reaches the site entry gives, which no allocation has (view_widened.go).
	entry func(ir.Expression) int
}

// unknownAt is why a node's value may come from somewhere the graph doesn't name, if it may: a
// parameter some call the graph can't see may pass anything, and a local a node writes with a value
// the walk can't name (localWrites) holds that too. Every reader of the graph's sources asks this
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
	graph := &allocationFlowGraph{program: program, sources: map[int][]ir.Expression{}, unknownParameters: map[int]bool{}, unknownWrites: map[int]string{}}
	collect := func(body []ir.Statement, function int) {
		walk(body, func(node any) bool {
			graph.noteWrites(reflect.ValueOf(node))
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

// localWrites is how collect takes every IR field that writes a local, by type and field: empty for
// one whose node's Value is the local's source (Declare and Assign, collect's own cases), and
// otherwise why the local is an unknown write. Its rows are exactly local_uses.go's localFields'
// write rows (TestLocalWritesAreLocalFieldsWrites), so an IR node that comes to write a local is
// classified here before it can leave a local with only some of its writers.
var localWrites = map[string]string{
	"Declare.Local": "",
	"Assign.Local":  "",
	// Each step writes the element into Local, or destructures it into Pattern's locals, a Binding
	// each, Local unused then. A for-in is a ForOf too, over its object's keys.
	"ForOf.Local":    "a for-of binding takes each element",
	"Binding.Local":  "a for-of binding takes each element",
	"Try.CatchLocal": "a catch binding takes what was thrown",
}

// writeFields is, for each struct type noteWrites has met, which of its fields write a local by a
// row of localWrites with a reason, and which hold plain structs to look into, worked out once a type.
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
		case field.Type.Kind() == reflect.Int && localWrites[name] != "":
			fields.writes = append(fields.writes, writeField{index: index, name: name, why: localWrites[name]})
		case field.Type.Kind() == reflect.Struct && !irNode(field.Type),
			field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct && !irNode(field.Type.Elem()):
			fields.inner = append(fields.inner, index)
		}
	}
	writeFields.Store(of, fields)
	return fields
}

// noteWrites records each local the fields of value write by a row of localWrites with a reason,
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
