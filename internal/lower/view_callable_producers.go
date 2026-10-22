package lower

import (
	"github.com/system-inc/adamic/internal/ir"
)

// Use the allocation graph's joins and projections to find known implementations.
// Unknown frontiers retain the existing runtime checks and are never certificates.
func reachingCallableProducers(graph *allocationFlowGraph, expression ir.Expression) []int {
	producers := []int{}
	seenFunctions, seenNodes := map[int]bool{}, map[int]bool{}
	var follow func(ir.Expression, int)
	var demand func(int)
	demand = func(node int) {
		if seenNodes[node] {
			return
		}
		seenNodes[node] = true
		for _, source := range graph.sources[node] {
			follow(source, 0)
		}
	}
	follow = func(value ir.Expression, depth int) {
		if value == nil || depth >= 32 {
			return
		}
		if closure, ok := value.(ir.MakeClosure); ok {
			if !seenFunctions[closure.Function] {
				seenFunctions[closure.Function] = true
				producers = append(producers, closure.Function)
			}
			return
		}
		if sources, _, recognized := graph.projectedSources(value, depth); recognized {
			for _, source := range sources {
				follow(source, depth+1)
			}
			return
		}
		switch wrapper := value.(type) {
		case ir.Conditional:
			follow(wrapper.WhenTrue, depth+1)
			follow(wrapper.WhenNot, depth+1)
			return
		case ir.Box:
			follow(wrapper.Value, depth+1)
			return
		case ir.Narrow:
			follow(wrapper.Value, depth+1)
			return
		case ir.Unwrap:
			follow(wrapper.Value, depth+1)
			return
		case ir.CheckedCast:
			follow(wrapper.Value, depth+1)
			return
		}
		graph.follow(value, func(int) {}, demand, func(string) {})
	}
	follow(expression, 0)
	return producers
}

// A known producer that fits no member of the callable contract it is read through is a cast the checker was told to
// believe. Base refused that read as an unsupported representation conversion, and so does this, located at the read,
// rather than admit a program whose read would only fail at run time where Node succeeds (#aecx10a, bad_callable.a).
func (l *lowering) callableProducerViewRefusal(graph *allocationFlowGraph, read ir.Property) error {
	root := l.result.ViewContractTypes[read.ViewTypeID]
	if root == 0 {
		return nil
	}
	contract := l.result.ViewContracts[root-1]
	members := []ir.ViewContractID{root}
	if contract.Kind == ir.ViewUnion && contract.Of == ir.Closure {
		members = contract.Members
	} else if contract.Kind != ir.ViewCallable {
		return nil
	}
	for _, member := range members {
		if l.untaggedCallableTargets[member] == nil {
			return nil
		}
	}
	for _, function := range reachingCallableProducers(graph, read) {
		known, fits := false, false
		for _, producer := range l.closureRecords {
			if producer.function != function {
				continue
			}
			known = true
			for _, member := range members {
				fits = fits || l.checker.IsTypeAssignableTo(producer.proven, l.untaggedCallableTargets[member])
			}
		}
		if known && !fits {
			return &Refused{Where: read.ViewWhere, What: "checked view read of field " + read.Name + " with unsupported representation conversion contract", Fix: "prove or implement the representation conversion contract before reading this field"}
		}
	}
	return nil
}
