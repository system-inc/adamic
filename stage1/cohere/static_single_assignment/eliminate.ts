/*
 * eliminate.go: redundant phi elimination.
 *
 * A phi is redundant when it is not actually a merge: every operand is the same value, or every operand
 * is either the same value or the phi's own result. The second case is what a loop produces when the
 * body does not reassign the variable (the back edge feeds the phi its own output), and it is the reason
 * this cannot be a simple "are all operands equal" test.
 *
 * Braun's construction places a phi whenever a lookup crosses a join, before it knows whether the
 * operands will agree. On a loop it must: the header's phi is minted before the back edge's value
 * exists. So construction cannot avoid producing redundant phis, and a redundant phi is not cosmetic:
 * it is a value that appears to have several reaching definitions when it has one, which is precisely
 * the fact every pass above this will key on.
 *
 * Removal cascades: deleting one phi can make another redundant, since a phi's operand may be the
 * removed phi's result. So this iterates to a fixed point rather than making one pass. Upstream's
 * equivalent is `eliminate_redundant_phi.rs`, 177 lines.
 */

import type { GraphInterface, IdentifierIdType, PhiInterface, PlaceVisitorType } from './static_single_assignment.ts';

/*
 * redundantPhiValue is the single value a phi collapses to, or undefined when it does not collapse.
 *
 * The rule, which is upstream's: ignoring operands that are the phi's own result, if every remaining
 * operand names one value, the phi is that value. A phi with no operands other than itself cannot arise
 * from a reachable block and is treated as not redundant so it stays visible.
 */
function redundantPhiValue<F, B, P>(
    graph: GraphInterface<F, B, P>,
    phi: PhiInterface<P>,
): IdentifierIdType | undefined {
    const result = graph.identifierOf(phi.place);
    let candidate: IdentifierIdType | undefined;
    for(const entry of phi.operands) {
        const operand = graph.identifierOf(entry.place);
        if(operand === result) {
            continue;
        }
        if(candidate === undefined) {
            candidate = operand;
            continue;
        }
        if(operand !== candidate) {
            return undefined;
        }
    }
    return candidate;
}

// applyRewrites rewrites every place in the function through resolve.
function applyRewrites<F, B, P>(
    graph: GraphInterface<F, B, P>,
    fn: F,
    resolve: (id: IdentifierIdType) => IdentifierIdType,
): void {
    // It takes the role it ignores: written (place) => ..., its type could be seen as Graph's
    // withIdentifier, and since it captures graph, stage 0's cycle finder would see the graph holding it
    // (`GAPS.md`, "The cycle rule's cost").
    // eslint-disable-next-line @typescript-eslint/no-unused-vars -- the parameter is what keeps the visitor from being seen as withIdentifier
    const rewrite: PlaceVisitorType<P> = (place, _role) =>
        graph.withIdentifier(place, resolve(graph.identifierOf(place)));
    for(const block of graph.blocks(fn)) {
        for(const phi of graph.phis(block)) {
            phi.place = rewrite(phi.place, 'Define');
            for(const operand of phi.operands) {
                operand.place = rewrite(operand.place, 'Use');
            }
        }
        const count = graph.instructionCount(fn, block);
        for(let index = 0; index < count; index++) {
            graph.eachInstructionPlace(fn, block, index, rewrite);
        }
        graph.eachTerminalPlace(block, rewrite);
    }
    const params = graph.params(fn);
    for(let index = 0; index < params.length; index++) {
        const param = params[index];
        if(param !== undefined) {
            params[index] = rewrite(param, 'Define');
        }
    }
    const returns = graph.returns(fn);
    if(returns !== undefined) {
        graph.setReturns(fn, rewrite(returns, 'Use'));
    }
}

// eliminateRedundantPhis removes phis that are not merges, rewriting every reference to a removed phi's
// result to the value it collapsed to.
//
// Runs to a fixed point over one function; an IR recurses into nested functions itself.
export function eliminateRedundantPhis<F, B, P>(graph: GraphInterface<F, B, P>, fn: F): void {
    for(;;) {
        // rewrites maps a removed phi's result to what it collapsed to.
        const rewrites = new Map<IdentifierIdType, IdentifierIdType>();

        for(const block of graph.blocks(fn)) {
            const kept: PhiInterface<P>[] = [];
            for(const phi of graph.phis(block)) {
                const collapsed = redundantPhiValue(graph, phi);
                if(collapsed !== undefined) {
                    rewrites.set(graph.identifierOf(phi.place), collapsed);
                    continue;
                }
                kept.push(phi);
            }
            graph.setPhis(block, kept);
        }

        if(rewrites.size === 0) {
            return;
        }

        // A removed phi may collapse to another removed phi's result, so chase each chain to its end
        // before rewriting. The chain is acyclic because a phi only collapses to a value that is not
        // itself.
        const resolve = function(start: IdentifierIdType): IdentifierIdType {
            let id = start;
            let seen = 0;
            for(;;) {
                const next = rewrites.get(id);
                if(next === undefined) {
                    return id;
                }
                id = next;
                seen++;
                if(seen > rewrites.size) {
                    // Defensive: a cycle cannot arise from the redundancy test above, but a caller that
                    // hand-built phis could create one, and looping forever is the worst possible response.
                    return id;
                }
            }
        };

        applyRewrites(graph, fn, resolve);
    }
}
