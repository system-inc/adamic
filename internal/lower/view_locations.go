package lower

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) recordViewCastOrigin(node *ast.Node, value ir.Expression) {
	l.result.ViewOrigins = append(l.result.ViewOrigins, value)
	cast := node
	if cast.Kind != ast.KindAsExpression {
		cast = nil
		var visit ast.Visitor
		visit = func(child *ast.Node) bool {
			if cast != nil {
				return true
			}
			if child.Kind == ast.KindAsExpression {
				cast = child
				return true
			}
			child.ForEachChild(visit)
			return false
		}
		node.ForEachChild(visit)
	}
	if cast != nil {
		l.result.ViewCastLocations = append(l.result.ViewCastLocations, ir.ViewCastLocation{Value: value, Where: l.program.Where(cast)})
	}
}

// Diagnostic candidates use the existing shared flow graph and projection
// index. Unknown provenance never pretends that one candidate is certain.
func (l *lowering) recordViewReadCastLocations(graph *allocationFlowGraph) {
	program := l.result
	if len(program.ViewCastLocations) == 0 {
		return
	}
	index := graph.projectionIndex()
	type origin struct {
		where   string
		sites   map[int]bool
		unknown bool
	}
	origins := []origin{}
	for _, cast := range program.ViewCastLocations {
		next := origin{where: filepath.Base(cast.Where), sites: map[int]bool{}}
		queue := []int{}
		add := func(value ir.Expression) {
			set := graph.ReachingAllocations(value)
			next.unknown = next.unknown || set.Unknown
			for _, site := range set.Sites {
				if !next.sites[site] {
					next.sites[site] = true
					queue = append(queue, site)
				}
			}
		}
		add(cast.Value)
		for len(queue) > 0 {
			site := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			if object, ok := index.records[site]; ok {
				for _, field := range object.Fields {
					if viewAggregate(field.Value) {
						add(field.Value)
					}
				}
				if object.Spread != nil {
					add(object.Spread)
				}
			}
			if array, ok := index.arrays[site]; ok {
				for _, value := range array.Elements {
					if viewAggregate(value) {
						add(value)
					}
				}
			}
			for _, values := range index.stores[site] {
				for _, value := range values {
					if viewAggregate(value) {
						add(value)
					}
				}
			}
		}
		origins = append(origins, next)
	}
	candidates := map[string]map[string]bool{}
	record := func(label string, receiver ir.Expression) {
		if label == "" || receiver == nil {
			return
		}
		set := graph.ReachingAllocations(receiver)
		if candidates[label] == nil {
			candidates[label] = map[string]bool{}
		}
		for _, cast := range origins {
			reaches := set.Unknown || cast.unknown
			for _, site := range set.Sites {
				reaches = reaches || cast.sites[site]
			}
			if reaches {
				candidates[label][cast.where] = true
			}
		}
		// Program-wide checked-field fallback can guard an ordinary receiver too.
		// Name the possible origins rather than inventing an exact source alias.
		if len(candidates[label]) == 0 {
			for _, cast := range origins {
				candidates[label][cast.where] = true
			}
		}
	}
	inspect := func(node any) bool {
		switch read := node.(type) {
		case ir.Property:
			record(read.View, read.Object)
		case ir.ArrayIndex:
			record(read.View, read.Array)
		case ir.ArrayMap:
			record(read.ViewRead.View, read.Array)
		case ir.ArrayVisit:
			record(read.ViewRead.View, read.Array)
		case ir.ArrayReduce:
			record(read.ViewRead.View, read.Array)
		case ir.ArrayPop:
			record(read.ViewRead.View, read.Array)
		case ir.ArrayJoin:
			record(read.ViewRead.View, read.Array)
		case ir.ArraySearch:
			record(read.ViewRead.View, read.Array)
		case ir.ForOf:
			record(read.ViewRead.View, read.Iterable)
		case ir.RecordCall:
			if read.DictionaryRead != nil {
				record(read.DictionaryRead.View, read.Arguments[0])
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	program.ViewReadCastLocations = map[string]string{}
	for label, sites := range candidates {
		names := []string{}
		for site := range sites {
			names = append(names, site)
		}
		sort.Strings(names)
		prefix := "view cast at "
		if len(names) > 1 {
			prefix = "possible view casts at "
		}
		program.ViewReadCastLocations[label] = prefix + strings.Join(names, ", ")
	}
}
