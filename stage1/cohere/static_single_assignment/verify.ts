/*
 * verify.go: verification of single static assignment form, the property checked rather than asserted.
 *
 * construct is only useful if what it produces is actually single assignment form, and "the tests pass" is weak evidence
 * when the tests were written by the same person as the construction. So the defining property is
 * checked directly, over whatever function is handed in:
 *
 *  1. Every value is defined at most once.
 *  2. Every use is dominated by its definition.
 *
 * The second is the real one. A construction that places a phi at the wrong block, or renames a use to
 * a definition sitting on a sibling branch, breaks exactly this and breaks nothing that a structural
 * well-formedness check would notice.
 *
 * Phi operands are checked against the predecessor, not the phi's own block. A phi's operand for
 * predecessor P is read on the edge from P, so the definition must dominate P's exit. Checking it the
 * naive way reports a false failure on every correct loop phi, since the back edge's value is defined
 * below the header.
 *
 * Dominance is computed here, over this graph, with the Cooper-Harvey-Kennedy fixed point over a reverse
 * postorder of the real edges that computeDominance makes for itself (it says why the block array isn't
 * one). Being a second implementation is acceptable for a checker: a checker that shares an
 * implementation with the thing it checks can agree with it and both be wrong.
 *
 * Where the port differs from the Go: Go's violation kind is an integer enum; here it is the union of
 * its names. A violation's String is its detail, which a caller reads off the field. Go's
 * realReversePostorder and entryPredecessors return several results; here each returns one object, or
 * undefined where the Go's last result is false.
 */

import { panic } from 'adamic';
import { edgesOf } from './graph.ts';
import type { BlockIdType, GraphInterface, IdentifierIdType, PlaceVisitorType } from './static_single_assignment.ts';

// Dominance is the immediate-dominator array over a function's blocks, by position in the order
// computeDominance made.
export class Dominance {
    // position maps a block id to its index in the order computeDominance made: a reverse postorder of
    // the real edges from the entry, then the blocks they don't reach.
    private readonly position: Map<BlockIdType, number>;
    // immediate[i] is the index of block i's immediate dominator; the entry is its own, and a block the
    // real edges don't reach has none (-1), dominated by nothing but itself.
    private readonly immediate: number[];

    constructor(position: Map<BlockIdType, number>, immediate: number[]) {
        this.position = position;
        this.immediate = immediate;
    }

    // dominates reports whether every path from the entry to block passes through dominator. A block
    // dominates itself, the standard convention.
    dominates(dominator: BlockIdType, block: BlockIdType): boolean {
        const target = this.position.get(dominator);
        if(target === undefined) {
            return false;
        }
        let index = this.position.get(block);
        if(index === undefined) {
            return false;
        }
        for(;;) {
            if(index === target) {
                return true;
            }
            const immediate: number = this.immediate[index] ?? -1;
            if(index === 0 || immediate === -1) {
                return false;
            }
            index = immediate;
        }
    }
}

/*
 * computeDominance runs Cooper-Harvey-Kennedy over the function's blocks.
 *
 * Why it orders the blocks itself rather than reading the block array: the algorithm needs every block
 * after some predecessor of it, so that each immediate dominator comes before its block and intersect's
 * walk up the tree ends where it should. A reverse postorder of the real edges guarantees that. The
 * block array is not always one: reversePostorder visits a structural fallthrough first, as upstream
 * does, and keeps a block where the fallthrough first reached it, so a loop that a fallthrough reaches
 * before its back edge sits ahead of every real predecessor it has. Over the array, the fixed point then
 * settled on too few dominators: 8 of 2,000 generated graphs disagreed with dominance computed as plain
 * set intersection (#6v4a54x). The array's order is what evaluation order and every other pass rely on,
 * so it stays, and this pass walks the real edges itself.
 *
 * Blocks the real edges don't reach (a fallthrough's placeholder) follow the reverse postorder, in the
 * array's order, with no dominator but themselves.
 */
export function computeDominance<F, B, P>(graph: GraphInterface<F, B, P>, fn: F): Dominance {
    const ordered = realReversePostorder(graph, fn);
    const blocks = ordered.blocks;
    const position = new Map<BlockIdType, number>();
    for(let index = 0; index < blocks.length; index++) {
        const block = blocks[index];
        if(block !== undefined) {
            position.set(graph.id(block), index);
        }
    }
    const immediate = Array.from({ length: blocks.length }, () => -1);
    if(ordered.reached === 0) {
        return new Dominance(position, immediate);
    }
    immediate[0] = 0;

    const immediateOf = (index: number): number => immediate[index] ?? -1;
    const intersect = function(first: number, second: number): number {
        let left = first;
        let right = second;
        while(left !== right) {
            while(left > right) {
                left = immediateOf(left);
            }
            while(right > left) {
                right = immediateOf(right);
            }
        }
        return left;
    };

    let changed = true;
    while(changed) {
        changed = false;
        for(let index = 1; index < ordered.reached; index++) {
            const block = blocks[index];
            if(block === undefined) {
                continue;
            }
            let newImmediate = -1;
            for(const predecessorId of graph.predecessors(block)) {
                const predecessorIndex = position.get(predecessorId);
                if(predecessorIndex === undefined || immediateOf(predecessorIndex) === -1) {
                    continue;
                }
                newImmediate = newImmediate === -1 ? predecessorIndex : intersect(predecessorIndex, newImmediate);
            }
            if(newImmediate !== -1 && immediateOf(index) !== newImmediate) {
                immediate[index] = newImmediate;
                changed = true;
            }
        }
    }
    return new Dominance(position, immediate);
}

// RealReversePostorder is realReversePostorder's answer: the blocks in its order, and how many of them,
// from the first, the real edges reach.
interface RealReversePostorderInterface<B> {
    readonly blocks: B[];
    readonly reached: number;
}

// RealFrame is one block on realReversePostorder's walk stack: its real successors, and the index of the
// one it follows next.
interface RealFrameInterface<B> {
    readonly block: B;
    readonly successors: BlockIdType[];
    next: number;
}

// realReversePostorder is the function's blocks in a reverse postorder of the real edges from its entry,
// then the blocks those edges don't reach, in the block array's order; and how many the edges reach.
// Fallthroughs are not edges. Exceptional edges are real ones, as everywhere in this module.
function realReversePostorder<F, B, P>(graph: GraphInterface<F, B, P>, fn: F): RealReversePostorderInterface<B> {
    const visited = new Map<BlockIdType, boolean>();
    const postorder: B[] = [];
    const stack: RealFrameInterface<B>[] = [];
    const entry = enterReal(graph, fn, graph.entry(fn), visited);
    if(entry !== undefined) {
        stack.push(entry);
    }
    while(stack.length > 0) {
        const top = stack[stack.length - 1] ?? panic('an empty walk stack');
        if(top.next === top.successors.length) {
            postorder.push(top.block);
            stack.pop();
            continue;
        }
        const next = top.successors[top.next] ?? panic('a frame reading past its successors');
        top.next++;
        const child = enterReal(graph, fn, next, visited);
        if(child !== undefined) {
            stack.push(child);
        }
    }

    const blocks: B[] = [];
    for(let index = postorder.length - 1; index >= 0; index--) {
        blocks.push(postorder[index] ?? panic('a postorder index past its end'));
    }
    for(const block of graph.blocks(fn)) {
        if(visited.get(graph.id(block)) !== true) {
            blocks.push(block);
        }
    }
    return { blocks, reached: postorder.length };
}

// enterReal is realReversePostorder's frame for a block it hasn't visited, with the block's real
// successors in the order its terminal names them, or undefined for a block it has visited or that isn't
// there. The Go's enter is a closure over the walk; this takes the walk's visited set instead.
function enterReal<F, B, P>(
    graph: GraphInterface<F, B, P>,
    fn: F,
    id: BlockIdType,
    visited: Map<BlockIdType, boolean>,
): RealFrameInterface<B> | undefined {
    const block = graph.block(fn, id);
    if(block === undefined || visited.get(id) === true) {
        return undefined;
    }
    visited.set(id, true);
    const successors: BlockIdType[] = [];
    for(const edge of edgesOf(graph, block)) {
        if(edge.edge !== 'Fallthrough') {
            successors.push(edge.successor);
        }
    }
    return { block, successors, next: 0 };
}

// EnteredEntry is an entry block some edge enters, and the blocks with an edge into it.
export interface EnteredEntryInterface {
    readonly entry: BlockIdType;
    readonly predecessors: BlockIdType[];
}

// entryPredecessors is the entry block and its predecessors when it has any, which graph.entry rules
// out: construct's lookup walks back through predecessors until a block has none, and an entry some edge
// enters can put it on a cycle that never ends. It is undefined for an entry no edge enters.
export function entryPredecessors<F, B, P>(
    graph: GraphInterface<F, B, P>,
    fn: F,
): EnteredEntryInterface | undefined {
    const entry = graph.entry(fn);
    const block = graph.block(fn, entry);
    if(block === undefined) {
        return undefined;
    }
    const predecessors = graph.predecessors(block);
    if(predecessors.length === 0) {
        return undefined;
    }
    return { entry, predecessors: [...predecessors] };
}

// SingleAssignmentViolationKindType (Go's SSAViolationKind) is which invariant a violation breaks: one of
// the two single assignment states (a value written more than once, or a use the definition does not
// dominate), or the precondition construct rests on (an entry block some edge enters, which graph.entry
// rules out and construct refuses).
export type SingleAssignmentViolationKindType = 'MultipleDefinitions' | 'UseNotDominated' | 'EntryHasPredecessors';

// SingleAssignmentViolationInterface (Go's SSAViolation) is one broken invariant.
export interface SingleAssignmentViolationInterface {
    // kind is which invariant broke.
    readonly kind: SingleAssignmentViolationKindType;
    // identifier is the value involved.
    readonly identifier: IdentifierIdType;
    // block is where the violating use or duplicate definition sits.
    readonly block: BlockIdType;
    // detail is a readable explanation naming both ends.
    readonly detail: string;
}

/*
 * verifySingleAssignment checks that a function is in single static assignment form and returns every
 * violation (Go's VerifySSA).
 *
 * An empty result means the property holds. It does not mean the function was worth checking: a
 * function whose variable references all lowered to globals has almost nothing to check and passes
 * trivially. collectSingleAssignmentStats is how a caller tells a real pass from a vacuous one.
 *
 * Nested functions are not descended into; call it per function.
 */
export function verifySingleAssignment<F, B, P>(
    graph: GraphInterface<F, B, P>,
    fn: F,
): SingleAssignmentViolationInterface[] {
    const blocks = graph.blocks(fn);
    const violations: SingleAssignmentViolationInterface[] = [];
    if(blocks.length === 0) {
        return violations;
    }

    const entered = entryPredecessors(graph, fn);
    if(entered !== undefined) {
        violations.push({
            kind: 'EntryHasPredecessors',
            // Go's violation leaves its identifier zero here: no value is involved.
            identifier: 0,
            block: entered.entry,
            detail: `the entry block bb${entered.entry} has predecessors [${entered.predecessors.map((id) => `${id}`).join(' ')}], and no edge may enter it`,
        });
    }
    const dominance = computeDominance(graph, fn);

    // Where each value is defined. A parameter is defined at the entry block.
    const definedIn = new Map<IdentifierIdType, BlockIdType>();
    // Position within the block, so a use earlier in the same block than its definition is caught.
    const definedAt = new Map<IdentifierIdType, number>();

    const recordDefinition = function(id: IdentifierIdType, blockId: BlockIdType, position: number): void {
        const previous = definedIn.get(id);
        if(previous !== undefined) {
            violations.push({
                kind: 'MultipleDefinitions',
                identifier: id,
                block: blockId,
                detail: `value ${graph.placeString(fn, id)} is defined in bb${previous} and again in bb${blockId}`,
            });
            return;
        }
        definedIn.set(id, blockId);
        definedAt.set(id, position);
    };

    const entry = graph.entry(fn);
    for(const param of graph.params(fn)) {
        recordDefinition(graph.identifierOf(param), entry, -1);
    }

    // The visitors read the block and position being walked from these.
    let blockId: BlockIdType = 0;
    let position = 0;
    let count = 0;
    let isContextStore = false;
    let currentBlock: B | undefined;
    const recordInstruction: PlaceVisitorType<P> = function(place, role) {
        if(role !== 'Define') {
            return place;
        }
        if(
            isContextStore &&
            currentBlock !== undefined &&
            !graph.contextStoreDefines(fn, currentBlock, position, place)
        ) {
            return place;
        }
        recordDefinition(graph.identifierOf(place), blockId, position);
        return place;
    };
    const recordTerminal: PlaceVisitorType<P> = function(place, role) {
        if(role === 'Define') {
            recordDefinition(graph.identifierOf(place), blockId, count);
        }
        return place;
    };

    for(const block of blocks) {
        blockId = graph.id(block);
        currentBlock = block;
        // A phi's result is defined at the top of its block, before every instruction.
        for(const phi of graph.phis(block)) {
            recordDefinition(graph.identifierOf(phi.place), blockId, -1);
        }
        count = graph.instructionCount(fn, block);
        for(position = 0; position < count; position++) {
            /*
             * A context binding is written many times, and upstream's invariant never sees it.
             * `AssertConsistentIdentifiers.ts:43` asserts "Expected lvalues to be assigned exactly once"
             * against the instruction's own temporary; a `StoreContext` writes its binding through a
             * different field, which is never added to that set. This walk records every definition, so
             * it is narrowed to match upstream's subject: an ordinary store still reports, and only the
             * place a context store writes through is exempt.
             */
            isContextStore = graph.isContextStore(fn, block, position);
            graph.eachInstructionPlace(fn, block, position, recordInstruction);
        }
        graph.eachTerminalPlace(block, recordTerminal);
    }

    // checkUse verifies one use sitting in useBlock at usePosition.
    const checkUse = function(id: IdentifierIdType, useBlock: BlockIdType, usePosition: number, what: string): void {
        const definitionBlock = definedIn.get(id);
        if(definitionBlock === undefined) {
            // Never defined in this function: a global, an import, or a capture. Not a violation; single
            // assignment says nothing about values this function does not define.
            return;
        }
        if(definitionBlock === useBlock) {
            if((definedAt.get(id) ?? 0) > usePosition) {
                violations.push({
                    kind: 'UseNotDominated',
                    identifier: id,
                    block: useBlock,
                    detail: `${what} in bb${useBlock} reads ${graph.placeString(fn, id)} before it is defined in the same block`,
                });
            }
            return;
        }
        if(!dominance.dominates(definitionBlock, useBlock)) {
            violations.push({
                kind: 'UseNotDominated',
                identifier: id,
                block: useBlock,
                detail: `${what} in bb${useBlock} reads ${graph.placeString(fn, id)} defined in bb${definitionBlock}, which does not dominate bb${useBlock}`,
            });
        }
    };

    const checkInstruction: PlaceVisitorType<P> = function(place, role) {
        if(role !== 'Define') {
            checkUse(graph.identifierOf(place), blockId, position, 'instruction');
        }
        return place;
    };
    const checkTerminal: PlaceVisitorType<P> = function(place, role) {
        if(role !== 'Define') {
            checkUse(graph.identifierOf(place), blockId, count, 'terminal');
        }
        return place;
    };

    for(const block of blocks) {
        blockId = graph.id(block);
        // A phi operand is read on the edge from its predecessor, so it must be dominated by the
        // predecessor's exit rather than by the phi's own block. See the file comment.
        for(const phi of graph.phis(block)) {
            for(const operand of phi.operands) {
                const predecessor = graph.block(fn, operand.predecessor);
                if(predecessor === undefined) {
                    continue;
                }
                checkUse(
                    graph.identifierOf(operand.place),
                    operand.predecessor,
                    graph.instructionCount(fn, predecessor),
                    `phi operand for bb${operand.predecessor}`,
                );
            }
        }
        count = graph.instructionCount(fn, block);
        for(position = 0; position < count; position++) {
            graph.eachInstructionPlace(fn, block, position, checkInstruction);
        }
        graph.eachTerminalPlace(block, checkTerminal);
    }

    return violations;
}

/*
 * SingleAssignmentStatsInterface (Go's SSAStats) counts what a verification actually had to look at.
 *
 * This exists because an empty violation list is exactly what a vacuous check returns. A function with
 * no phis and no renamed values passes verification perfectly while proving nothing. A caller reporting
 * a clean run should report these numbers alongside it.
 */
export interface SingleAssignmentStatsInterface {
    // phis is how many merge points were placed.
    readonly phis: number;
    // namedValues is how many defined values carry a source name rather than being a temporary. Zero
    // means no source variable was ever resolved, so nothing was renamed.
    readonly namedValues: number;
    // uses is how many use sites were checked for dominance.
    readonly uses: number;
}

// collectSingleAssignmentStats measures one function (Go's CollectSSAStats).
export function collectSingleAssignmentStats<F, B, P>(
    graph: GraphInterface<F, B, P>,
    fn: F,
): SingleAssignmentStatsInterface {
    let phis = 0;
    let namedValues = 0;
    let uses = 0;
    const seen = new Map<IdentifierIdType, boolean>();
    const note = function(place: P): void {
        const id = graph.identifierOf(place);
        if(seen.get(id) === true) {
            return;
        }
        seen.set(id, true);
        if(graph.named(fn, id)) {
            namedValues++;
        }
    };
    const instruction: PlaceVisitorType<P> = function(place, role) {
        if(role === 'Define') {
            note(place);
            return place;
        }
        uses++;
        return place;
    };
    const terminal: PlaceVisitorType<P> = function(place, role) {
        if(role !== 'Define') {
            uses++;
        }
        return place;
    };
    for(const block of graph.blocks(fn)) {
        const blockPhis = graph.phis(block);
        phis += blockPhis.length;
        for(const phi of blockPhis) {
            note(phi.place);
        }
        const count = graph.instructionCount(fn, block);
        for(let index = 0; index < count; index++) {
            graph.eachInstructionPlace(fn, block, index, instruction);
        }
        graph.eachTerminalPlace(block, terminal);
    }
    return { phis, namedValues, uses };
}
