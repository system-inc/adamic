package lower

import "github.com/system-inc/adamic/internal/ir"

// Tuples have indexed native storage, without an object length slot. Refuse
// only a demanded object contract over a known tuple, before backend emission.
func tupleObjectViewRead(program *ir.Program, graph *allocationFlowGraph, index *allocationProjection, read ir.Property) error {
	id := program.ViewContractTypes[read.ViewTypeID]
	if id == 0 {
		return nil
	}
	contract := program.ViewContracts[id-1]
	if contract.Kind != ir.ViewObject || contract.FixedTuple {
		return nil
	}
	length := false
	for _, field := range contract.Fields {
		length = length || field.Name == "length"
	}
	if !length {
		return nil
	}
	for _, site := range graph.ReachingAllocations(read).Sites {
		if literal, ok := index.records[site]; ok && literal.Tuple {
			return &NotYet{Where: read.ViewWhere, What: "a tuple through an object view with a length field; expose a scalar length field on the view and copy pair.length into that field"}
		}
	}
	return nil
}
