package lower

import (
	"reflect"
	"sort"
	"strconv"

	"github.com/system-inc/adamic/internal/ir"
)

// ReachingAllocations queries the shared graph without selecting graph ownership.
// Unknown frontiers retain checks. Joined assignments, returns and arguments use
// the same edges as the graph-regions analysis.
func (graph *allocationFlowGraph) ReachingAllocations(expression ir.Expression) ir.AllocationSet {
	return graph.reachingAllocations(expression, true, 0)
}

func (graph *allocationFlowGraph) reachingAllocations(expression ir.Expression, project bool, depth int) (result ir.AllocationSet) {
	if read, ok := expression.(ir.Read); ok && project {
		if graph.projectionReads == nil {
			graph.projectionReads = map[int]ir.AllocationSet{}
			graph.projectionBusy = map[int]bool{}
		}
		if cached, ok := graph.projectionReads[read.Local]; ok {
			return cached
		}
		if graph.projectionBusy[read.Local] {
			return ir.AllocationSet{Unknown: true, Reasons: []string{"recursive projection receiver not certified"}}
		}
		graph.projectionBusy[read.Local] = true
		defer func() { delete(graph.projectionBusy, read.Local); graph.projectionReads[read.Local] = result }()
	}
	result = ir.AllocationSet{}
	sites, visited, reasons := map[int]bool{}, map[int]bool{}, map[string]bool{}
	queue := []int{}
	demand := func(node int) {
		if !visited[node] {
			visited[node] = true
			queue = append(queue, node)
		}
	}
	unknown := func(reason string) { result.Unknown = true; reasons[reason] = true }
	var follow func(ir.Expression, int)
	follow = func(value ir.Expression, level int) {
		if project {
			sources, reasons, recognized := graph.projectedSources(value, level)
			if recognized {
				for _, reason := range reasons {
					unknown(reason)
				}
				for _, source := range sources {
					follow(source, level+1)
				}
				return
			}
			switch wrapper := value.(type) {
			case ir.Conditional:
				follow(wrapper.WhenTrue, level)
				follow(wrapper.WhenNot, level)
				return
			case ir.Box:
				follow(wrapper.Value, level)
				return
			case ir.Narrow:
				follow(wrapper.Value, level)
				return
			case ir.Unwrap:
				follow(wrapper.Value, level)
				return
			case ir.CheckedCast:
				follow(wrapper.Value, level)
				return
			}
		}
		graph.follow(value, func(site int) { sites[site] = true }, demand, unknown)
	}
	follow(expression, depth)
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if graph.unknownParameters[node] {
			unknown("parameter may receive untracked or omitted arguments")
		}
		sources := graph.sources[node]
		if len(sources) == 0 {
			unknown("local or result has no tracked producer")
		}
		for _, source := range sources {
			follow(source, depth)
		}
	}
	if len(sites) == 0 {
		unknown("no reaching allocation established")
	}
	for site := range sites {
		result.Sites = append(result.Sites, site)
	}
	for reason := range reasons {
		result.Reasons = append(result.Reasons, reason)
	}
	sort.Ints(result.Sites)
	sort.Strings(result.Reasons)
	return result
}

// rewriteShapeStatements preserves the IR tree's containers and rewrites only
// interface-held nodes, so metadata never becomes an executable expression.
func rewriteShapeStatements(body []ir.Statement, rewrite func(any) any) []ir.Statement {
	var transform func(reflect.Value) reflect.Value
	transform = func(value reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			mapped := transform(value.Elem())
			result := reflect.New(value.Type()).Elem()
			result.Set(reflect.ValueOf(rewrite(mapped.Interface())))
			return result
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(transform(value.Field(i)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(transform(value.Index(i)))
			}
			return result
		default:
			return value
		}
	}
	return transform(reflect.ValueOf(body)).Interface().([]ir.Statement)
}

// Projection is queried only by shape proofs. Graph-regions traversal stays unchanged.
type allocationProjection struct {
	records       map[int]ir.ObjectLiteral
	arrays        map[int]ir.ArrayLiteral
	stores        map[int]map[string][]ir.Expression
	opaque        map[string]bool
	opaqueSources map[string][]ir.Expression
	arrayMutation bool
	opaqueCalls   bool
}

func (graph *allocationFlowGraph) projectionIndex() *allocationProjection {
	if graph.projection != nil {
		return graph.projection
	}
	index := &allocationProjection{records: map[int]ir.ObjectLiteral{}, arrays: map[int]ir.ArrayLiteral{}, stores: map[int]map[string][]ir.Expression{}, opaque: map[string]bool{}, opaqueSources: map[string][]ir.Expression{}}
	graph.projection = index
	stores := []ir.SetProperty{}
	collect := func(node any) bool {
		switch value := node.(type) {
		case ir.ObjectLiteral:
			if len(value.GraphTypes) > 0 {
				index.records[value.GraphTypes[len(value.GraphTypes)-1]] = value
			}
		case ir.ArrayLiteral:
			if len(value.GraphTypes) > 0 {
				index.arrays[value.GraphTypes[len(value.GraphTypes)-1]] = value
			}
		case ir.SetProperty:
			stores = append(stores, value)
		case ir.SetIndex, ir.ArrayPush, ir.ArraySplice, ir.ArrayPop:
			index.arrayMutation = true
		case ir.CallClosure, ir.ObjectCall:
			index.opaqueCalls = true
		}
		return true
	}
	walk(graph.program.Main, collect)
	for _, function := range graph.program.Functions {
		walk(function.Body, collect)
	}
	for _, store := range stores {
		receiver := graph.reachingAllocations(store.Object, false, 0)
		if receiver.Unknown || store.Uninitialized {
			index.opaque[store.Name] = true
			index.opaqueSources[store.Name] = append(index.opaqueSources[store.Name], store.Object, store.Value)
		}
		for _, site := range receiver.Sites {
			if index.stores[site] == nil {
				index.stores[site] = map[string][]ir.Expression{}
			}
			index.stores[site][store.Name] = append(index.stores[site][store.Name], store.Value)
		}
	}
	return index
}

// projectedSources joins initial fields and all visible stores across all paths.
// An opaque receiver can alias any allocation; its field name stays unknown.
func (graph *allocationFlowGraph) projectedSources(expression ir.Expression, depth int) ([]ir.Expression, []string, bool) {
	var receiver ir.Expression
	name := ""
	array := false
	switch value := expression.(type) {
	case ir.Property:
		receiver, name = value.Object, value.Name
	case ir.ArrayIndex:
		receiver = value.Array
		array = true
		key, ok := value.Index.(ir.NumberConstant)
		if !ok || key.Value < 0 || key.Value != float64(int(key.Value)) || value.Relative {
			return nil, []string{"dynamic or relative element key not certified"}, true
		}
		name = strconv.Itoa(int(key.Value))
	default:
		return nil, nil, false
	}
	if depth >= 32 {
		return nil, []string{"recursive projection depth not certified"}, true
	}
	index := graph.projectionIndex()
	sources := []ir.Expression{}
	reasons := []string{}
	if index.opaqueCalls {
		reasons = append(reasons, "opaque call may mutate projected fields")
	}
	if index.opaque[name] {
		reasons = append(reasons, "field store has opaque receiver or clears readiness: "+name)
	}
	if array && index.arrayMutation {
		reasons = append(reasons, "array mutation prevents element certificate")
	}
	bases := graph.reachingAllocations(receiver, true, depth+1)
	reasons = append(reasons, bases.Reasons...)
	for _, site := range bases.Sites {
		found := false
		if literal, ok := index.records[site]; ok {
			if literal.Spread != nil {
				reasons = append(reasons, "spread field projection not certified")
			}
			for _, field := range literal.Fields {
				if field.Name == name {
					sources = append(sources, field.Value)
					found = true
				}
			}
		}
		if literal, ok := index.arrays[site]; ok && array {
			key, _ := strconv.Atoi(name)
			if len(literal.Spread) > 0 {
				reasons = append(reasons, "spread element projection not certified")
			}
			if key < len(literal.Elements) {
				sources = append(sources, literal.Elements[key])
				found = true
			}
		}
		if !found {
			reasons = append(reasons, "projected field absent or allocation not modeled: "+name)
		}
		sources = append(sources, index.stores[site][name]...)
	}
	return sources, reasons, true
}
