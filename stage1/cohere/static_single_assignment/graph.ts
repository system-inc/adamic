/*
 * graph.go: graph maintenance, the passes an IR runs to bring a freshly built function into the state
 * every other pass assumes.
 *
 * Three invariants are established here and every consumer may rely on them:
 *
 *  1. The block array is in reverse postorder, and unreachable blocks have been removed.
 *  2. Every block's predecessors are exactly the set of blocks with a real edge into it.
 *  3. Every instruction and terminal has a nonzero, monotonically increasing EvaluationOrder.
 *
 * An IR re-establishes them by calling the three passes in that order (its finalize), which any pass
 * that restructures the graph must do.
 *
 * Where the port differs from the Go: the Go keeps each pass's working memory in a sync.Pool between
 * calls, and hands Graph callbacks made once with it, because each was an allocation per lowered
 * function on a cold ahra run (#rwsffzm, #p4h0p54). Those are Go's costs, measured on Go; here each
 * pass makes its working memory per call, and the traversal is unchanged.
 */

import { panic } from 'adamic';
import type { BlockIdType, EdgeType, GraphInterface } from './static_single_assignment.ts';

// inRange reports whether an id is under the function's block bound, which every set is as long as.
function inRange(id: BlockIdType, bound: number): boolean {
    return id < bound;
}

/*
 * reversePostorder reorders the function's blocks into reverse postorder and drops unreachable blocks.
 *
 * Why unreachable blocks are dropped here and kept in controlflow: `controlflow` keeps them, because a
 * rule may want to ask what would have run after a `return`. This drops them, because a value defined
 * only in a block control never reaches is a value no analysis should see: single-assignment
 * construction over an unreachable definition produces a phi operand from a predecessor that cannot
 * execute, and every effect derived from it is fiction.
 *
 * The traversal is upstream's `getReversePostorderedBlocks`: visit a structural fallthrough first, then
 * the real successors in reverse order. Postorder is reversed at the end, so that puts loop bodies and
 * conditional arms before their continuation in the final array. That ordering is more than
 * presentation. Evaluation order is assigned from this array, and mutable-range inference must see a
 * mutation in a loop body before a read after the loop.
 *
 * A fallthrough is visited initially as structural rather than executable. If no real edge reaches it,
 * upstream retains its block id as an empty unreachable block; if a real edge reaches it later, the
 * same block is revisited as executable. Keeping those two states separate prevents a fallthrough link
 * from manufacturing a control-flow edge while still preserving the structured terminal's required
 * target. An IR whose terminals have only real edges reports no 'Fallthrough', and the walk is then a
 * plain reverse postorder.
 */
export function reversePostorder<F, B, P>(graph: GraphInterface<F, B, P>, fn: F): void {
    // The three sets are dense arrays over block ids, bounded by blockBound, and the walk keeps its
    // frames and their successors in two flat stacks rather than a frame and two lists per block.
    const bound = graph.blockBound(fn);
    const visited = Array.from({ length: bound }, () => false);
    const used = Array.from({ length: bound }, () => false);
    const usedFallthroughs = Array.from({ length: bound }, () => false);

    const postorder: BlockIdType[] = [];
    // successors are every frame's edges still to follow, a frame's from its start to its end. A
    // child's are pushed after its parent's and popped when the child is, so the array is a stack in
    // step with the frames.
    const successors: ReversePostorderSuccessorInterface[] = [];
    const stack: ReversePostorderFrameInterface[] = [];

    // real and the fallthrough are what one block's edges gave, gathered by collect.
    let real: BlockIdType[] = [];
    let fallthroughBlock = 0;
    let hasFallthrough = false;
    const collect = function(next: BlockIdType, edge: EdgeType): void {
        if(edge === 'Fallthrough') {
            fallthroughBlock = next;
            hasFallthrough = true;
            return;
        }
        real.push(next);
    };

    const enter = function(id: BlockIdType, isUsed: boolean): ReversePostorderFrameInterface | undefined {
        if(!inRange(id, bound)) {
            return undefined;
        }
        const wasUsed = used[id] === true;
        const wasVisited = visited[id] === true;
        visited[id] = true;
        if(isUsed) {
            used[id] = true;
        }
        if(wasVisited && (wasUsed || !isUsed)) {
            return undefined;
        }

        const block = graph.block(fn, id);
        if(block === undefined) {
            return undefined;
        }
        const start = successors.length;
        real = [];
        hasFallthrough = false;
        graph.eachEdge(block, collect);
        if(hasFallthrough) {
            if(isUsed && inRange(fallthroughBlock, bound)) {
                usedFallthroughs[fallthroughBlock] = true;
            }
            successors.push({ id: fallthroughBlock, isUsed: false });
        }
        for(let index = real.length - 1; index >= 0; index--) {
            successors.push({ id: real[index] ?? panic('an edge index past its block'), isUsed });
        }
        return { block: id, start, end: successors.length, next: start, ownsAppend: !wasVisited };
    };

    const entry = enter(graph.entry(fn), true);
    if(entry === undefined) {
        return;
    }
    stack.push(entry);

    while(stack.length > 0) {
        const top = stack[stack.length - 1] ?? panic('an empty walk stack');
        if(top.next === top.end) {
            if(top.ownsAppend) {
                postorder.push(top.block);
            }
            while(successors.length > top.start) {
                successors.pop();
            }
            stack.pop();
            continue;
        }
        const next = successors[top.next] ?? panic('a frame reading past its successors');
        top.next++;
        const child = enter(next.id, next.isUsed);
        if(child !== undefined) {
            stack.push(child);
        }
    }

    const blocks: B[] = [];
    for(let index = postorder.length - 1; index >= 0; index--) {
        const id = postorder[index] ?? panic('a postorder index past its end');
        if(inRange(id, bound) && used[id] === true) {
            const block = graph.block(fn, id);
            if(block !== undefined) {
                blocks.push(block);
            }
        }
        else if(inRange(id, bound) && usedFallthroughs[id] === true) {
            const block = graph.block(fn, id);
            if(block !== undefined) {
                blocks.push(graph.placeholder(fn, block));
            }
        }
    }
    graph.setBlocks(fn, blocks);
    graph.retain(fn, (id) => inRange(id, bound) && (used[id] === true || usedFallthroughs[id] === true));
}

// ReversePostorderSuccessor is one edge reversePostorder will follow, and whether it is executable.
interface ReversePostorderSuccessorInterface {
    readonly id: BlockIdType;
    readonly isUsed: boolean;
}

// ReversePostorderFrame is one block on reversePostorder's walk stack. Its successors are the shared
// stack's entries from start to end, and next is the one it follows next.
interface ReversePostorderFrameInterface {
    readonly block: BlockIdType;
    readonly start: number;
    readonly end: number;
    next: number;
    readonly ownsAppend: boolean;
}

// EdgeTo is one gathered edge.
export interface EdgeToInterface {
    readonly successor: BlockIdType;
    readonly edge: EdgeType;
}

// edgesOf gathers a block's edges, for a pass that reads them in a loop of its own rather than in a
// callback.
export function edgesOf<F, B, P>(graph: GraphInterface<F, B, P>, block: B): EdgeToInterface[] {
    const edges: EdgeToInterface[] = [];
    graph.eachEdge(block, function(successor, edge) {
        edges.push({ successor, edge });
    });
    return edges;
}

/*
 * markPredecessors recomputes every block's predecessors from the real edges.
 *
 * Fallthroughs are not edges and do not produce a predecessor; exceptional edges are real ones.
 *
 * This is the field single-assignment construction needs: Braun's algorithm reads a value at a block
 * join by asking each predecessor what it holds, and a missing or spurious predecessor is a wrong phi
 * rather than a crash.
 */
export function markPredecessors<F, B, P>(graph: GraphInterface<F, B, P>, fn: F): void {
    const blocks = graph.blocks(fn);
    for(const block of blocks) {
        graph.setPredecessors(block, []);
    }
    for(const block of blocks) {
        const id = graph.id(block);
        const seen = new Map<BlockIdType, boolean>();
        for(const edge of edgesOf(graph, block)) {
            if(edge.edge === 'Fallthrough' || seen.get(edge.successor) === true) {
                continue;
            }
            seen.set(edge.successor, true);
            const successor = graph.block(fn, edge.successor);
            if(successor === undefined) {
                continue;
            }
            graph.setPredecessors(successor, [...graph.predecessors(successor), id]);
        }
    }
}

/*
 * markEvaluationOrder assigns each instruction and terminal its position in evaluation order.
 *
 * Numbering follows the block array, which is reverse postorder, so a forward analysis sees increasing
 * numbers along any acyclic path. Across a back edge the number decreases, which is the correct and
 * expected signal that a loop was traversed.
 *
 * Starts at 1: zero means unassigned, and a pass that reads an order of zero has found a block the
 * finalizer did not reach.
 */
export function markEvaluationOrder<F, B, P>(graph: GraphInterface<F, B, P>, fn: F): void {
    let order = 1;
    for(const block of graph.blocks(fn)) {
        const count = graph.instructionCount(fn, block);
        for(let index = 0; index < count; index++) {
            graph.setInstructionOrder(fn, block, index, order);
            order++;
        }
        graph.setTerminalOrder(block, order);
        order++;
    }
}
