package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
)

// Scope fallback families to target contracts carried by actual reaching
// allocations. This queries lane 3's graph and projection index; it does not
// introduce another allocation or alias solver. Untracked origins and opaque
// projections retain the original conservative field-name fallback.
func (l *lowering) scopedViewFieldFamilies(graph *allocationFlowGraph) (map[int]map[string]string, bool) {
	program := l.result
	casts := []ir.CheckedCast{}
	collect := func(node any) bool {
		if cast, ok := node.(ir.CheckedCast); ok && cast.ViewContract != 0 {
			casts = append(casts, cast)
		}
		return true
	}
	walk(program.Main, collect)
	for _, f := range program.Functions {
		walk(f.Body, collect)
	}
	type demand struct {
		site     int
		contract ir.ViewContractID
	}
	queue := []demand{}
	unknown := false
	add := func(value ir.Expression, contract ir.ViewContractID) {
		set := graph.ReachingAllocations(value)
		unknown = unknown || set.Unknown
		for _, site := range set.Sites {
			queue = append(queue, demand{site, contract})
		}
	}
	for _, origin := range program.ViewOrigins {
		matched := false
		for _, cast := range casts {
			if reflect.DeepEqual(origin, cast.Value) {
				matched = true
				add(origin, cast.ViewContract)
			}
		}
		if !matched {
			unknown = true
		}
	}
	index := graph.projectionIndex()
	unknown = unknown || index.opaqueCalls || index.arrayMutation
	scopes := map[int]map[string]string{}
	seen := map[demand]bool{}
	for len(queue) != 0 {
		next := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if seen[next] {
			continue
		}
		seen[next] = true
		if next.contract <= 0 || int(next.contract) > len(program.ViewContracts) {
			unknown = true
			continue
		}
		contract := program.ViewContracts[next.contract-1]
		if scopes[next.site] == nil {
			scopes[next.site] = map[string]string{}
		}
		if contract.Unsupported != "" {
			scopes[next.site]["*"] = contract.Unsupported
		}
		for _, member := range contract.Members {
			queue = append(queue, demand{next.site, member})
		}
		if contract.ObjectPresent != 0 {
			queue = append(queue, demand{next.site, contract.ObjectPresent})
		}
		literal, record := index.records[next.site]
		if record && literal.Spread != nil {
			unknown = true
		}
		for _, field := range contract.Fields {
			if field.Contract <= 0 || int(field.Contract) > len(program.ViewContracts) {
				unknown = true
				continue
			}
			child := program.ViewContracts[field.Contract-1]
			if child.Unsupported != "" {
				scopes[next.site][field.Name] = child.Unsupported
			}
			if index.opaque[field.Name] {
				unknown = true
			}
			if record && viewScopeAggregate(program, field.Contract, map[ir.ViewContractID]bool{}) {
				for _, allocated := range literal.Fields {
					if allocated.Name == field.Name && viewAggregate(allocated.Value) {
						add(allocated.Value, field.Contract)
					}
				}
			}
			for _, value := range index.stores[next.site][field.Name] {
				if viewAggregate(value) && viewScopeAggregate(program, field.Contract, map[ir.ViewContractID]bool{}) {
					add(value, field.Contract)
				}
			}
		}
		if contract.Element != 0 && viewScopeAggregate(program, contract.Element, map[ir.ViewContractID]bool{}) {
			if array, ok := index.arrays[next.site]; ok {
				if len(array.Spread) > 0 {
					unknown = true
				}
				for _, element := range array.Elements {
					if viewAggregate(element) {
						add(element, contract.Element)
					}
				}
			} else {
				unknown = true
			}
		}
	}
	return scopes, unknown
}

// Nullish storage is not an aggregate allocation. Primitive contracts need no
// descendant scope; wrong primitive payloads remain checked at the field read.
func viewScopeAggregate(program *ir.Program, id ir.ViewContractID, seen map[ir.ViewContractID]bool) bool {
	if id <= 0 || int(id) > len(program.ViewContracts) || seen[id] {
		return false
	}
	seen[id] = true
	contract := program.ViewContracts[id-1]
	switch contract.Kind {
	case ir.ViewObject, ir.ViewArray, ir.ViewMap, ir.ViewCallable:
		return true
	}
	if contract.ObjectPresent != 0 && viewScopeAggregate(program, contract.ObjectPresent, seen) {
		return true
	}
	for _, member := range contract.Members {
		if viewScopeAggregate(program, member, seen) {
			return true
		}
	}
	return false
}
