/*
 * phi.go: a phi's operands, an array sorted by predecessor block id with at most one entry each.
 *
 * The Go was a map from predecessor to place, kept for the lookup "what came from this predecessor". A
 * phi has two or three operands, where a short sorted list answers that as fast as a hash and costs far
 * less (#p4h0p54), so it is a sorted list there and here.
 *
 * Where the port differs from the Go: the Go's operand list is a named slice type with methods, and Set
 * and Delete write through a pointer to it. Here they are functions that return the new list, and the
 * caller stores it back in the phi.
 */

import type { BlockIdType, PhiInterface, PhiOperandInterface } from './static_single_assignment.ts';

// operandIndex is where predecessor's entry is in operands, or where it would go, and whether it is
// there.
function operandIndex<P>(
    operands: readonly PhiOperandInterface<P>[],
    predecessor: BlockIdType,
): { readonly index: number; readonly found: boolean } {
    let index = 0;
    for(const operand of operands) {
        if(operand.predecessor >= predecessor) {
            return { index, found: operand.predecessor === predecessor };
        }
        index++;
    }
    return { index, found: false };
}

// phiOperand is the operand from predecessor, or undefined when there is none (Go's Get and At).
export function phiOperand<P>(operands: readonly PhiOperandInterface<P>[], predecessor: BlockIdType): P | undefined {
    const position = operandIndex(operands, predecessor);
    if(!position.found) {
        return undefined;
    }
    return operands[position.index]?.place;
}

// withPhiOperand makes place the operand from predecessor, replacing one already there (Go's Set).
export function withPhiOperand<P>(
    operands: PhiOperandInterface<P>[],
    predecessor: BlockIdType,
    place: P,
): PhiOperandInterface<P>[] {
    const position = operandIndex(operands, predecessor);
    const existing = operands[position.index];
    if(position.found && existing !== undefined) {
        existing.place = place;
        return operands;
    }
    return [...operands.slice(0, position.index), { predecessor, place }, ...operands.slice(position.index)];
}

// withoutPhiOperand removes the operand from predecessor, if there is one (Go's Delete).
export function withoutPhiOperand<P>(
    operands: PhiOperandInterface<P>[],
    predecessor: BlockIdType,
): PhiOperandInterface<P>[] {
    const position = operandIndex(operands, predecessor);
    if(!position.found) {
        return operands;
    }
    return [...operands.slice(0, position.index), ...operands.slice(position.index + 1)];
}

// phiOperandsInOrder returns a phi's predecessor block ids, in ascending block order.
//
// A phi's operands are kept sorted by predecessor block id, so this is their predecessor ids in that
// order, read straight off the list.
export function phiOperandsInOrder<P>(phi: PhiInterface<P>): BlockIdType[] {
    return phi.operands.map((operand) => operand.predecessor);
}
