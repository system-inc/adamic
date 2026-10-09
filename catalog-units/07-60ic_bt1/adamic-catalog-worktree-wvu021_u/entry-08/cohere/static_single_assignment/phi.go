package static_single_assignment

import "slices"

// Phi is a merge: the value of Place at the top of a block, given which predecessor control came from.
//
// Operands holds one entry per entry in the block's predecessors, kept sorted by predecessor block id,
// so ranging it is deterministic and is the order PhiOperandsInOrder has always given.
//
// It was a `map[BlockId]Place`, kept for the lookup "what came from THIS predecessor". A phi has two or
// three operands, where a map's smallest allocation holds eight and a linear scan beats a hash, so the
// map cost 41 MB over 533K phis on a cold ahra run for nothing a short sorted slice cannot answer
// (#p4h0p54). The lookup is `Operands.At` or `Operands.Get`.
type Phi[P any] struct {
	Place    P
	Operands PhiOperands[P]
}

// PhiOperand is the value a phi takes when control arrives from Predecessor.
type PhiOperand[P any] struct {
	Predecessor BlockId
	Place       P
}

// PhiOperands is a phi's operands, sorted by predecessor block id with at most one entry each.
type PhiOperands[P any] []PhiOperand[P]

// index is where predecessor's entry is, or where it would go, and whether it is there.
func (operands PhiOperands[P]) index(predecessor BlockId) (int, bool) {
	for index, operand := range operands {
		if operand.Predecessor >= predecessor {
			return index, operand.Predecessor == predecessor
		}
	}
	return len(operands), false
}

// Get is the operand from predecessor, and whether there is one.
func (operands PhiOperands[P]) Get(predecessor BlockId) (P, bool) {
	if index, found := operands.index(predecessor); found {
		return operands[index].Place, true
	}
	var none P
	return none, false
}

// At is the operand from predecessor, or the zero place when there is none, as the map's index was.
func (operands PhiOperands[P]) At(predecessor BlockId) P {
	place, _ := operands.Get(predecessor)
	return place
}

// Set makes place the operand from predecessor, replacing one already there.
func (operands *PhiOperands[P]) Set(predecessor BlockId, place P) {
	index, found := operands.index(predecessor)
	if found {
		(*operands)[index].Place = place
		return
	}
	*operands = slices.Insert(*operands, index, PhiOperand[P]{Predecessor: predecessor, Place: place})
}

// Delete removes the operand from predecessor, if there is one.
func (operands *PhiOperands[P]) Delete(predecessor BlockId) {
	if index, found := operands.index(predecessor); found {
		*operands = slices.Delete(*operands, index, index+1)
	}
}

// PhiOperandsInOrder returns a phi's predecessor block ids, in ascending block order.
//
// `Phi.Operands` is kept sorted by predecessor block id, so this is its predecessor ids in that order,
// read straight off the slice.
func PhiOperandsInOrder[P any](phi *Phi[P]) []BlockId {
	blocks := make([]BlockId, 0, len(phi.Operands))
	for _, operand := range phi.Operands {
		blocks = append(blocks, operand.Predecessor)
	}
	return blocks
}
