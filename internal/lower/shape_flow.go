package lower

import (
	"reflect"
	"sort"

	"github.com/system-inc/adamic/internal/ir"
)

// ReachingAllocations queries the shared graph without selecting graph ownership.
// Unknown frontiers retain checks. Joined assignments, returns and arguments use
// the same edges as the graph-regions analysis.
func (graph *allocationFlowGraph) ReachingAllocations(expression ir.Expression) ir.AllocationSet {
	result := ir.AllocationSet{}
	sites, visited, reasons := map[int]bool{}, map[int]bool{}, map[string]bool{}
	queue := []int{}
	demand := func(node int) {
		if !visited[node] {
			visited[node] = true
			queue = append(queue, node)
		}
	}
	unknown := func(reason string) { result.Unknown = true; reasons[reason] = true }
	follow := func(value ir.Expression) { graph.follow(value, func(site int) { sites[site] = true }, demand, unknown) }
	follow(expression)
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
			follow(source)
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
