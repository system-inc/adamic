package lower

import (
	"fmt"
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// graphAllocationSites copies the IR tree before assigning unique allocation
// IDs. Pointer and map children must be copied too, so their sites do not depend
// solely on a checker identity. Nil containers preserve their original shape.
func graphAllocationSites(value reflect.Value, next *int) reflect.Value {
	var sites func(reflect.Value) reflect.Value
	sites = func(value reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			mapped := sites(value.Elem())
			result := reflect.New(value.Type()).Elem()
			result.Set(mapped)
			return result
		case reflect.Ptr:
			if value.IsNil() {
				return value
			}
			result := reflect.New(value.Type().Elem())
			result.Elem().Set(sites(value.Elem()))
			return result
		case reflect.Map:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeMapWithSize(value.Type(), value.Len())
			entries := value.MapRange()
			for entries.Next() {
				result.SetMapIndex(sites(entries.Key()), sites(entries.Value()))
			}
			return result
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(sites(value.Field(i)))
			}
			if field := result.FieldByName("GraphTypes"); field.IsValid() && field.Type() == reflect.TypeOf([]int{}) {
				*next--
				field.Set(reflect.ValueOf(append(append([]int{}, field.Interface().([]int)...), *next)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(sites(value.Index(i)))
			}
			return result
		default:
			return value
		}
	}
	return sites(value)
}

// graphFlows works backwards from graph views to the allocations that can supply
// them. Local assignments and function results are joined across all paths; this
// is a may-flow proof, not an execution or a choice of one conditional branch.
// Negative IDs name allocation sites in this IR, independently of fresh checker
// identities. They select the existing allocation emission without runtime work.
func (f *cycleFinder) graphFlows() {
	program := f.l.result
	assignAllocationSites(program)

	var graphType func(*checker.Type) bool
	graphType = func(proven *checker.Type) bool {
		if proven == nil || f.weak(proven) {
			return false
		}
		if program.GraphTypes[int(proven.Id())] {
			return true
		}
		if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range proven.Types() {
				if graphType(member) {
					return true
				}
			}
		}
		return false
	}
	// Positive nodes are locals followed by function results.
	flow := newAllocationFlowGraph(program)
	wanted := map[int]bool{}
	queue := []int{}
	demand := func(node int) {
		if !wanted[node] {
			wanted[node] = true
			queue = append(queue, node)
		}
	}
	for local, proven := range f.l.localTypes {
		if graphType(proven) {
			demand(local + 1)
		}
	}
	for local, types := range f.l.localAlso {
		for _, proven := range types {
			if graphType(proven) {
				demand(local + 1)
			}
		}
	}
	pending := []ir.Expression{}
	// Site retains the checker view of the actual slot being written.
	slot := func(site int, name string, argument int) bool {
		if site <= 0 || site > len(f.l.writeSites) {
			return false
		}
		holder := f.l.writeSites[site-1].holder
		if holder == nil {
			return false
		}
		if name != "" {
			for _, field := range f.fields(holder) {
				if f.l.cycleFieldMatches(holder, field.Name, name) {
					return graphType(f.l.checker.GetTypeOfSymbol(field))
				}
			}
			return false
		}
		args := f.l.checker.GetTypeArguments(holder)
		return argument < len(args) && graphType(args[argument])
	}
	// A counted aggregate can have graph-typed fields/elements. Seed those
	// initializer slots individually, rather than promoting its whole payload.
	returns := map[int]*checker.Type{}
	for symbol, index := range f.l.functions {
		signatures := f.l.checker.GetSignaturesOfType(f.l.checker.GetTypeOfSymbol(symbol), checker.SignatureKindCall)
		if len(signatures) > 0 {
			returns[index] = f.l.checker.GetReturnTypeOfSignature(signatures[0])
		}
	}
	for _, closure := range f.l.closureRecords {
		signatures := f.l.checker.GetSignaturesOfType(closure.proven, checker.SignatureKindCall)
		if len(signatures) > 0 {
			returns[closure.function] = f.l.checker.GetReturnTypeOfSignature(signatures[0])
		}
	}
	var members func(ir.Expression, *checker.Type)
	members = func(expression ir.Expression, proven *checker.Type) {
		if expression == nil || proven == nil || f.weak(proven) {
			return
		}
		if graphType(proven) {
			pending = append(pending, expression)
		}
		switch value := expression.(type) {
		case ir.ObjectLiteral:
			for _, field := range f.fields(proven) {
				for _, initializer := range value.Fields {
					if f.l.cycleFieldMatches(proven, field.Name, initializer.Name) {
						members(initializer.Value, f.l.checker.GetTypeOfSymbol(field))
					}
				}
			}
		case ir.ArrayLiteral:
			arguments := f.l.checker.GetTypeArguments(proven)
			if len(arguments) > 0 {
				for _, element := range value.Elements {
					members(element, arguments[0])
				}
			}
		case ir.Conditional:
			members(value.WhenTrue, proven)
			members(value.WhenNot, proven)
		}
	}
	collect := func(body []ir.Statement, function int) {
		walk(body, func(node any) bool {
			switch value := node.(type) {
			case ir.Declare:
				members(value.Value, f.l.localTypes[value.Local])
			case ir.Assign:
				members(value.Value, f.l.localTypes[value.Local])
			case ir.Return:
				members(value.Value, returns[function])
			case ir.SetProperty:
				if slot(value.Site, value.Name, 0) {
					pending = append(pending, value.Value)
				}
			case ir.SetIndex:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			case ir.ArrayPush:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			case ir.MapSet:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Key)
				}
				if slot(value.Site, "", 1) {
					pending = append(pending, value.Value)
				}
			case ir.SetAdd:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			}
			return true
		})
	}
	collect(program.Main, -1)
	for i, function := range program.Functions {
		collect(function.Body, i)
	}
	follow := func(expression ir.Expression) {
		flow.follow(expression, func(site int) { program.GraphTypes[site] = true }, demand, func(string) {})
	}
	for _, expression := range pending {
		follow(expression)
	}
	for len(queue) > 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		for _, expression := range flow.sources[node] {
			follow(expression)
		}
	}
}

// allocationFlowGraph is the shared may-flow graph used by graph regions and
// shape certification. It never changes ownership when queried.
type allocationFlowGraph struct {
	projection        *allocationProjection
	projectionReads   map[int]ir.AllocationSet
	projectionBusy    map[int]bool
	program           *ir.Program
	sources           map[int][]ir.Expression
	unknownParameters map[int]bool
}

func (graph *allocationFlowGraph) resultNode(function int) int {
	return len(graph.program.Locals) + function + 1
}

func newAllocationFlowGraph(program *ir.Program) *allocationFlowGraph {
	graph := &allocationFlowGraph{program: program, sources: map[int][]ir.Expression{}, unknownParameters: map[int]bool{}}
	collect := func(body []ir.Statement, function int) {
		walk(body, func(node any) bool {
			switch value := node.(type) {
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

func (graph *allocationFlowGraph) follow(expression ir.Expression, allocation func(int), demand func(int), unknown func(string)) {
	program := graph.program
	follow := func(value ir.Expression) { graph.follow(value, allocation, demand, unknown) }
	if expression == nil {
		unknown("missing expression")
		return
	}
	value := reflect.ValueOf(expression)
	if field := value.FieldByName("GraphTypes"); field.IsValid() {
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
