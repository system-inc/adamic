package lower

import "github.com/system-inc/adamic/internal/ir"

// Checker assignability does not adapt packed native arguments. Membership
// certification is sound only if the implementation and view use the same ABI.
func (l *lowering) untaggedCallableABI(index int, contract ir.ViewContract) bool {
	function := l.result.Functions[index]
	if !function.Closure || function.Receiver || function.RestElement != 0 || len(function.Parameters) != len(contract.Parameters) {
		return false
	}
	// The producer registry records only fixed scalar or void results. A
	// discarded result still needs a recorded implementation signature.
	if function.Returns != 0 && function.Returns != ir.Number && function.Returns != ir.Boolean && function.Returns != ir.String {
		return false
	}
	for i, local := range function.Parameters {
		of := l.result.ViewContracts[contract.Parameters[i]-1].Of
		if of == 0 || l.result.Locals[local].Type != of {
			return false
		}
	}
	return contract.DiscardResult || contract.Result != 0 && function.Returns == l.result.ViewContracts[contract.Result-1].Of
}

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

// A known assignable producer with an incompatible ABI is valid TypeScript but
// cannot satisfy this adapter. Diagnose that at the demanded read, before C or JS.
func (l *lowering) callableProducerViewRefusal(graph *allocationFlowGraph, read ir.Property) error {
	root := l.result.ViewContractTypes[read.ViewTypeID]
	if root == 0 {
		return nil
	}
	contract := l.result.ViewContracts[root-1]
	if contract.Kind != ir.ViewCallable && contract.Kind != ir.ViewUnion {
		return nil
	}
	active := map[ir.ViewContractID]bool{}
	var inspect func(ir.ViewContractID, ir.Expression) error
	inspect = func(id ir.ViewContractID, value ir.Expression) error {
		if id == 0 || active[id] {
			return nil
		}
		active[id] = true
		defer delete(active, id)
		contract := l.result.ViewContracts[id-1]
		if contract.Kind == ir.ViewCallable || contract.Kind == ir.ViewUnion && contract.Of == ir.Closure {
			members := contract.Members
			if contract.Kind == ir.ViewCallable {
				members = []ir.ViewContractID{id}
			}
			for _, function := range reachingCallableProducers(graph, value) {
				assignable, compatible := false, false
				for _, producer := range l.closureRecords {
					if producer.function != function {
						continue
					}
					for _, member := range members {
						target := l.untaggedCallableTargets[member]
						if target != nil && l.checker.IsTypeAssignableTo(producer.proven, target) {
							assignable = true
							compatible = compatible || l.untaggedCallableABI(function, l.result.ViewContracts[member-1])
						}
					}
				}
				if assignable && !compatible {
					return &NotYet{Where: read.ViewWhere, What: "a checked callable producer with a different scalar storage, arity, or receiver convention; wrap the producer in an arrow with the view member's exact parameters and result"}
				}
			}
			return nil
		}
		// Tagged object membership checks only its selectors. Unread callable
		// descendants remain deferred until their own checked field read.
		if contract.Kind == ir.ViewObject {
			for _, field := range contract.Fields {
				child := l.result.ViewContracts[field.Contract-1]
				if !field.Optional && child.Kind == ir.ViewScalar && (len(child.Allowed) != 0 || field.Name == "kind" && (child.Of == ir.Number || child.Of == ir.Boolean || child.Of == ir.String)) {
					return nil
				}
			}
		}

		for _, member := range contract.Members {
			if err := inspect(member, value); err != nil {
				return err
			}
		}
		for _, field := range contract.Fields {
			child := l.result.ViewContracts[field.Contract-1]
			projected := ir.Property{Object: value, Name: field.Name, Of: child.Of}
			if err := inspect(field.Contract, projected); err != nil {
				return err
			}
		}
		return nil
	}
	return inspect(root, read)
}
