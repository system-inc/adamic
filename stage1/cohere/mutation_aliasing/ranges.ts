/* eslint-disable max-classes-per-file -- As ranges.go holds them: the range, its table, the definition walk, and the alias graph's node, state and result. */
/*
 * ranges.go: mutable ranges, over which span of a function's evaluation a value is still being written.
 *
 * This is React's `InferMutationAliasingRanges.ts`, run over GraphInterface to produce a MutableRanges
 * table. oxc transcribes the whole pass at
 * `oxc_react_compiler/src/react_compiler_inference/infer_mutation_aliasing_ranges.rs`; where the two
 * disagree React wins. The Go file's package comment holds the measurements behind each rule, taken on
 * cohere's corpus when the pass lived in its high-level IR, and the one divergence from both upstreams,
 * the loop-carried inversion, recorded at the line in the lvalue loop here as there.
 *
 * A value's mutable range is the span of program points over which it is still being written. Once the
 * range is closed, the value is settled: nothing after `end` changes it.
 *
 * A range is an interval, not a set, and that is upstream's decision. React's membership is
 * `id >= range.start && id < range.end`, a half-open interval. A value mutated at instruction 3 and again
 * at 20 has [3, 21), and every instruction between reports as inside it: a reactive scope covering the
 * value has to cover both mutations and everything between them. The interval also carries upstream's
 * invariant: a range is either unset (both zero) or non-empty (end > start). validateMutableRanges is
 * that check.
 *
 * Upstream writes a range at two kinds of site. The definition sites need no effects: every lvalue opens
 * its range at its instruction's order and closes it at order + 1, and the operand loop opens the start
 * of a value whose end is already past the instruction. The extension site is AliasingState.mutate, a
 * worklist over the alias graph the effects build, and StoreContext, which widens from the instruction's
 * shape alone. Upstream runs the widening (its Part 1) entirely before the definition half (its Part 2),
 * and the order is load-bearing: inferMutableRanges says why.
 *
 * Where the port differs from the Go, and why:
 *
 *   - MutableRange is a class whose fields never change, where Go's is a struct copied on read. A pass
 *     that widens a range sets a new one.
 *   - The node's four back-edge maps are Maps, where Go keeps a key slice beside each Go map for its
 *     insertion order; a Map iterates in insertion order already. The identity members are a Set for
 *     the same reason. Go ranges its member map in an order of its own; which members freeze first
 *     doesn't change what ends up frozen.
 *   - Go's definition walk is a struct with a method, made once per call so a visitor handed to a Graph
 *     method doesn't allocate per instruction. Here it's DefinitionWalk, made once per call likewise,
 *     and the one visitor that reads it.
 *   - An effect's from is undefined where Go's is a zero place, whose identifier is 0. The graph builder
 *     reads it as identifier 0, as Go does, though no IR gives a flow effect without one.
 *   - AliasingGraph's state and mutations are readable, as Go's package tests read them.
 *     phiOpensBefore and phiOpenedRange are exported for the same reason.
 *   - insertBackEdge is a method of AliasingNode and freeze's visited set a field of the state, where
 *     Go hands each map in, so a function writes only what it or its receiver holds.
 *   - Each declaration sits above its first read, so inferMutableRanges, first in Go, is last here.
 *   - The file holds six classes, as ranges.go holds the six types they mirror, which the directive at
 *     the top allows.
 */

import { panic } from 'adamic';
import type {
    BlockIdType,
    EvaluationOrderType,
    IdentifierIdType,
    PhiInterface,
    RoleType,
} from '../static_single_assignment/static_single_assignment.ts';
import type { EffectValueKindType, GraphInterface, OptionsInterface } from './mutation_aliasing.ts';

/*
 * MutableRange (Go's MutableRange) is the half-open span of evaluation over which a value is still being
 * written.
 *
 * Membership is `start <= order < end`, which is React's `inRange` exactly. A range with both fields zero
 * is unset: the value has no recorded definition point. That is a real state rather than an error: a
 * function's parameters and context values arrive already defined, so nothing in this pass opens a range
 * for them.
 */
export class MutableRange {
    readonly start: EvaluationOrderType;
    readonly end: EvaluationOrderType;

    constructor(start: EvaluationOrderType, end: EvaluationOrderType) {
        this.start = start;
        this.end = end;
    }

    // isSet reports whether a range was ever opened.
    isSet(): boolean {
        return this.start !== 0 || this.end !== 0;
    }

    // contains reports whether an evaluation position lies inside the range. Half-open, matching
    // React's `inRange`: the start is inside and the end is not. An unset range contains nothing.
    contains(order: EvaluationOrderType): boolean {
        if(!this.isSet()) {
            return false;
        }
        return order >= this.start && order < this.end;
    }

    // isValid reports whether a range satisfies upstream's invariant, React's `validateMutableRange`:
    // `(start === 0 && end === 0) || end > start`.
    isValid(): boolean {
        if(!this.isSet()) {
            return true;
        }
        return this.end > this.start;
    }
}

// unsetRange is the range a value this pass never opened has (Go's zero MutableRange).
const unsetRange = new MutableRange(0, 0);

// The empty fallbacks a lookup ranges over when it finds nothing, typed, since stage 0 can't lower an
// empty array literal's never[] (gap 1).
const noIdentifiers: readonly IdentifierIdType[] = [];
const noMembers: ReadonlySet<IdentifierIdType> = new Set<IdentifierIdType>();

/*
 * RangeGapType (Go's RangeGap) names a range-widening rule this pass cannot apply, for a caller that
 * needs to know.
 *
 * 'LoopCarriedInversion' is a loop-carried value left unset because its two writes produced an inverted
 * interval. A loop-carried value is defined at a high evaluation order by the back-edge store, while the
 * mutation reaching it through the phi happened at a lower one, and upstream's two guards each write
 * their own field. This pass leaves those unset rather than clamping, because the true range starts
 * before the back-edge store and how far before is exactly what the folded phi hides. See the lvalue loop.
 */
export type RangeGapType = 'LoopCarriedInversion';

// rangeGaps (Go's RangeGaps) are the widening rules inferMutableRanges does not apply, returned as a
// value so a test can assert on it, which makes closing a gap a visible event.
export function rangeGaps(): RangeGapType[] {
    return ['LoopCarriedInversion'];
}

/*
 * MutableRanges (Go's MutableRanges) is the table this pass produces: one range per value.
 *
 * A side table rather than a field on an IR's identifier, since each IR has its own identifier type and
 * this module holds neither. A value absent from the table has the unset range, which get returns, so a
 * caller never has to distinguish "not computed" from "no range".
 */
export class MutableRanges {
    private readonly ranges = new Map<IdentifierIdType, MutableRange>();

    // get returns a value's range, or the unset range when this pass never opened one.
    get(id: IdentifierIdType): MutableRange {
        return this.ranges.get(id) ?? unsetRange;
    }

    // length (Go's Len) is how many values the table holds a range for, for measurement.
    length(): number {
        return this.ranges.size;
    }

    // contains reports whether a value is still being written at an evaluation position: React's
    // `inRange` reached through the table. Half-open, and false for an unset range.
    contains(id: IdentifierIdType, order: EvaluationOrderType): boolean {
        return this.get(id).contains(order);
    }

    // set writes a range.
    set(id: IdentifierIdType, range: MutableRange): void {
        this.ranges.set(id, range);
    }

    // invalid (the body of Go's ValidateMutableRanges) is every value whose range breaks upstream's
    // invariant, in ascending order.
    invalid(): IdentifierIdType[] {
        const invalid: IdentifierIdType[] = [];
        for(const [id, range] of this.ranges) {
            if(!range.isValid()) {
                invalid.push(id);
            }
        }
        invalid.sort((left, right) => left - right);
        return invalid;
    }
}

// rangeOf (Go's RangeOf) returns a value's mutable range from a table: the name upstream callers use,
// and the seam that would absorb a later move of the range onto an identifier.
export function rangeOf(ranges: MutableRanges, id: IdentifierIdType): MutableRange {
    return ranges.get(id);
}

// isMutableAt (Go's IsMutableAt) reports whether a value is still being written at an evaluation
// position: React's `inRange`. False can mean "settled" or "we could not see the mutation", and a
// consumer that must not confuse those asks rangeGaps first.
export function isMutableAt(ranges: MutableRanges, id: IdentifierIdType, order: EvaluationOrderType): boolean {
    return ranges.contains(id, order);
}

// validateMutableRanges (Go's ValidateMutableRanges) reports every value whose range breaks upstream's
// invariant, sorted, so a failure message is stable between runs. An empty result is the passing answer.
export function validateMutableRanges(ranges: MutableRanges): IdentifierIdType[] {
    return ranges.invalid();
}

/*
 * phiOpensBefore reports whether a phi's range should open before its block.
 *
 * Upstream's test is `phi.place.identifier.mutableRange.end > firstInstructionIdOfBlock`, which asks
 * whether the phi's value is still being mutated after the block it merges in. That end can only have
 * been widened by mutate.
 */
export function phiOpensBefore(existing: MutableRange, firstOrder: EvaluationOrderType): boolean {
    if(firstOrder === 0) {
        return false;
    }
    return existing.end > firstOrder;
}

/*
 * phiOpenedRange returns a phi's range opened early, or undefined when it doesn't move (Go returns the
 * range and whether it moved).
 *
 * Upstream opens it at `firstInstructionIdOfBlock - 1`, one before the block's first instruction. A
 * phi's value exists at the top of the block, before anything in it runs, so a range starting at the
 * first instruction would exclude the merge point itself.
 */
export function phiOpenedRange(existing: MutableRange, firstOrder: EvaluationOrderType): MutableRange | undefined {
    if(!phiOpensBefore(existing, firstOrder) || existing.start !== 0) {
        return undefined;
    }
    return new MutableRange(firstOrder - 1, existing.end);
}

/*
 * blockFirstOrder (Go's BlockFirstOrder) returns the evaluation position of a block's first instruction.
 *
 * Upstream falls back to the terminal's order for an empty block,
 * `block.instructions.at(0)?.id ?? block.terminal.id`. An empty block is common, every goto-only join
 * block is one, and using zero there would open a phi's range at a position no instruction occupies.
 */
export function blockFirstOrder<F, B, P>(graph: GraphInterface<F, B, P>, fn: F, block: B): EvaluationOrderType {
    const count = graph.instructionCount(fn, block);
    for(let index = 0; index < count; index++) {
        const order = graph.instructionOrder(fn, block, index);
        if(order !== undefined) {
            return order;
        }
    }
    return graph.terminalOrder(block);
}

// maxOrder returns the larger of two evaluation positions. Upstream writes `Math.max(instr.id + 1,
// existing)` inside a branch that has just tested the existing value is zero, so the max is redundant
// there; it is kept because the redundancy is upstream's.
function maxOrder(left: EvaluationOrderType, right: EvaluationOrderType): EvaluationOrderType {
    if(left > right) {
        return left;
    }
    return right;
}

// DefinitionWalk (Go's definitionWalk) is the definition half's one place visitor and what it reads:
// which of the two loops over an instruction's places it is in, and the instruction's order.
class DefinitionWalk<F, B, P> {
    private readonly graph: GraphInterface<F, B, P>;
    private readonly result: MutableRanges;
    order: EvaluationOrderType = 0;
    lvalues = false;

    constructor(graph: GraphInterface<F, B, P>, result: MutableRanges) {
        this.graph = graph;
        this.result = result;
    }

    // visitPlace opens an lvalue's range in the first loop and a widened operand's start in the second.
    visitPlace(place: P, role: RoleType): void {
        if(this.lvalues) {
            if(role === 'Define') {
                this.openLValue(this.graph.identifierOf(place));
            }
            return;
        }
        if(role === 'Define') {
            return;
        }
        const id = this.graph.identifierOf(place);
        const existing = this.result.get(id);
        if(existing.end > this.order && existing.start === 0) {
            this.result.set(id, new MutableRange(this.order, existing.end));
        }
    }

    // openLValue opens the range of a value an instruction defines, the lvalue loop's body.
    private openLValue(id: IdentifierIdType): void {
        const existing = this.result.get(id);
        let start = existing.start;
        let end = existing.end;
        let changed = false;
        if(start === 0) {
            start = this.order;
            changed = true;
        }
        if(end === 0) {
            end = maxOrder(this.order + 1, end);
            changed = true;
        }
        /*
         * Divergence from both upstreams, in the conservative direction, and the only one this pass
         * introduces.
         *
         * Upstream's two guards above are each conditioned on their own field being unset, so a value
         * whose end was widened by a mutation and whose start is opened here keeps both writes
         * independently. On a loop-carried value that produces an interval whose end is at or before its
         * start: the back-edge phi operand is defined at a high evaluation order, while the mutation that
         * reaches it through the phi happened at a lower one. React asserts against exactly this in
         * `validateMutableRange`, behind a flag that defaults to false, so upstream ships the state it
         * declares invalid.
         *
         * Reproducing that would ship intervals on which contains answers false at every position, a
         * confident no at exactly the positions where the value demonstrably was being written. So the
         * range is unset instead, and 'LoopCarriedInversion' names it. Clamping the end to start + 1 was
         * considered and rejected: it writes down a range we don't have, and it satisfies isValid by
         * construction, which would leave validateMutableRanges unable to fail.
         */
        const range = new MutableRange(start, end);
        if(range.isSet() && range.end <= range.start) {
            this.result.set(id, unsetRange);
            return;
        }
        if(changed) {
            this.result.set(id, range);
        }
    }
}

// =============================================================================
// AliasingState: the widening that makes a mutable range wider than one
// =============================================================================

/*
 * AliasingEdgeKindType (Go's aliasingEdgeKind) is the kind of a directed data-flow edge between two
 * values. A 'MaybeAlias' edge downgrades a definite mutation to a conditional one when it is traversed,
 * which is the only place the kind is read during the walk.
 */
export type AliasingEdgeKindType = 'Capture' | 'Alias' | 'MaybeAlias';

/*
 * MutationKindType (Go's mutationKind) is how certain a mutation is: 1 conditional, 2 definite. Ordered,
 * and the order is load-bearing: the worklist's revisit test is `previous >= current -> skip`, so
 * conditional < definite is what lets a definite mutation re-enter a node a conditional one already
 * reached.
 */
export type MutationKindType = 1 | 2;
const mutationKindConditional: MutationKindType = 1;
const mutationKindDefinite: MutationKindType = 2;

/*
 * AliasingEdgeInterface (Go's aliasingEdge) is one directed edge, tagged with the sequence index at which
 * it came into being.
 *
 * index is the whole reason this pass is not a naive graph walk. Every edge and every mutation is
 * stamped with a monotonically increasing counter as the graph is built, and the walk refuses to
 * traverse an edge whose index is at or after the mutation's own index. That is upstream's model of
 * time: a mutation cannot flow backwards through an alias that did not exist yet.
 */
export interface AliasingEdgeInterface {
    readonly index: number;
    readonly node: IdentifierIdType;
    readonly kind: AliasingEdgeKindType;
}

/*
 * AliasingNodeValueType (Go's aliasingNodeValue) distinguishes a phi from an ordinary object. A phi does
 * not propagate a mutation backwards through its alias edges when the mutation arrived travelling
 * forwards, which is the one place the walk consults the node's value at all.
 */
export type AliasingNodeValueType = 'Object' | 'Phi';

// AliasingBackEdgeKindType names a node's four maps of backward edges.
export type AliasingBackEdgeKindType = 'CreatedFrom' | 'Capture' | 'Alias' | 'MaybeAlias';

/*
 * AliasingNode (Go's aliasingNode) is one value in the data-flow graph.
 *
 * The four maps are the backward edges, kept separately because the walk traverses them under different
 * conditions: createdFrom always forces a transitive mutation, captures is followed only for a
 * transitive mutation, and aliases and maybeAliases are followed only when the walk is travelling
 * backwards or the node is not a phi. edges is the single forward list, traversed unconditionally.
 * Insertion order matters for the maps, as it does upstream, and a Map keeps it.
 */
export class AliasingNode {
    readonly id: IdentifierIdType;
    readonly createdFrom = new Map<IdentifierIdType, number>();
    readonly captures = new Map<IdentifierIdType, number>();
    readonly aliases = new Map<IdentifierIdType, number>();
    readonly maybeAliases = new Map<IdentifierIdType, number>();
    readonly edges: AliasingEdgeInterface[] = [];
    readonly value: AliasingNodeValueType;

    constructor(id: IdentifierIdType, value: AliasingNodeValueType) {
        this.id = id;
        this.value = value;
    }

    /*
     * insertBackEdge (Go's insertBackEdge) records a backward edge of a kind, keeping first-wins semantics
     * and insertion order.
     *
     * Upstream writes `entry(from).or_insert(index)` in Rust and `if (!map.has(k)) map.set(k, v)` in
     * JavaScript: the earliest index for a repeated edge is the one kept, which makes the edge visible to
     * the widest set of mutations. Keeping the latest would silently narrow every range that flows through
     * a repeated alias.
     */
    insertBackEdge(kind: AliasingBackEdgeKindType, from: IdentifierIdType, index: number): void {
        const edges = this.backEdges(kind);
        if(edges.has(from)) {
            return;
        }
        edges.set(from, index);
    }

    // backEdges are the backward edges of a kind.
    private backEdges(kind: AliasingBackEdgeKindType): Map<IdentifierIdType, number> {
        switch(kind) {
            case 'CreatedFrom':
                return this.createdFrom;
            case 'Capture':
                return this.captures;
            case 'Alias':
                return this.aliases;
            case 'MaybeAlias':
                return this.maybeAliases;
        }
    }
}

/*
 * AliasingState (Go's aliasingState) is the directed edge graph the widening walks.
 *
 * Both upstreams contain a node type carrying created-from, captures, aliases, maybe-aliases and a
 * forward edge list; four builders called from the ranges pass's own main loop; and mutate as a worklist
 * over that graph. The single widening write in the upstream file is inside that worklist.
 */
export class AliasingState {
    readonly nodes = new Map<IdentifierIdType, AliasingNode>();

    /*
     * identities tracks the identifiers that currently denote the same abstract value. An Assign of a
     * mutable value points the destination at the source's identity, so a later Freeze through either
     * name changes the kind observed through every other name. Alias and MaybeAlias edges describe
     * possible information flow, not shared identity, and freezing across them would suppress real
     * mutations. Each Create starts a fresh identity; CreateFrom keeps a distinct one with a copied kind.
     */
    private readonly identities = new Map<IdentifierIdType, number>();
    private readonly identityMembers = new Map<number, Set<IdentifierIdType>>();
    private nextIdentity = 0;

    // freezeSources are values upstream's abstract value itself points through for freezing: a phi
    // denotes the union of its operands, and freezing a FunctionExpression freezes its captures. They are
    // separate from the Capture edges: an object can capture a value without Freeze(object) freezing it.
    private readonly freezeSources = new Map<IdentifierIdType, readonly IdentifierIdType[]>();

    // freezeSeen are the values the freeze under way has reached (Go's freezeWithSeen takes them as an
    // argument), so each is frozen once however many paths reach it.
    private readonly freezeSeen = new Set<IdentifierIdType>();

    // immutable is upstream's abstract value kind, reduced to what the mutation gate needs: present means
    // "not Mutable or Context", absent means mutable. A value the walk never reached has no entry, and
    // the gate treats that as mutable, the conservative direction for widening.
    readonly immutable = new Map<IdentifierIdType, EffectValueKindType>();

    // markImmutable records that a value is not Mutable or Context, so a conditional mutation of it widens
    // nothing.
    markImmutable(id: IdentifierIdType, kind: EffectValueKindType): void {
        this.immutable.set(id, kind);
    }

    // freeze marks an already-known abstract value as frozen. Upstream's InferenceState.freeze never
    // creates a value for a place the state has not initialized, and node presence is this state's
    // initialization predicate.
    freeze(id: IdentifierIdType): boolean {
        this.freezeSeen.clear();
        return this.freezeWithSeen(id);
    }

    // freezeWithSeen freezes a value the freeze under way hasn't reached yet, and what it points through.
    private freezeWithSeen(id: IdentifierIdType): boolean {
        if(this.freezeSeen.has(id)) {
            return false;
        }
        this.freezeSeen.add(id);
        if(!this.nodes.has(id)) {
            return false;
        }
        const kind = this.immutable.get(id);
        if(kind !== undefined && kind !== 'MaybeFrozen') {
            return false;
        }
        const identity = this.identities.get(id);
        if(identity === undefined) {
            this.markImmutable(id, 'Frozen');
            for(const source of this.freezeSources.get(id) ?? noIdentifiers) {
                this.freezeWithSeen(source);
            }
            return true;
        }
        for(const alias of this.identityMembers.get(identity) ?? noMembers) {
            this.markImmutable(alias, 'Frozen');
            for(const source of this.freezeSources.get(alias) ?? noIdentifiers) {
                this.freezeWithSeen(source);
            }
        }
        return true;
    }

    // notMutable reports whether a value is neither Mutable nor Context: the question `state.mutate`
    // asks, whose conditional arm mutates for those two and returns `none` for everything else.
    notMutable(id: IdentifierIdType): boolean {
        return this.immutable.has(id);
    }

    // deriveImmutable gives into the same mutability as from: upstream's CreateFrom reads the source's
    // kind and initializes the target with it.
    deriveImmutable(from: IdentifierIdType, into: IdentifierIdType): void {
        const kind = this.immutable.get(from);
        if(kind !== undefined) {
            this.immutable.set(into, kind);
            return;
        }
        this.immutable.delete(into);
    }

    // create makes a node, replacing any existing one, as upstream's insert overwrites. A value redefined
    // by a later instruction gets a fresh node with no edges, which is right in single-assignment form
    // because the redefinition is a different value.
    create(id: IdentifierIdType, value: AliasingNodeValueType): void {
        this.nodes.set(id, new AliasingNode(id, value));
        this.immutable.delete(id);
        this.freezeSources.delete(id);
        this.detachIdentity(id);
        this.nextIdentity++;
        this.identities.set(id, this.nextIdentity);
        this.identityMembers.set(this.nextIdentity, new Set<IdentifierIdType>([id]));
    }

    // recordFreezeSources gives a phi or FunctionExpression the values React freezes through when that
    // abstract value is frozen later.
    recordFreezeSources(into: IdentifierIdType, sources: readonly IdentifierIdType[]): void {
        if(sources.length === 0) {
            return;
        }
        this.freezeSources.set(into, sources);
    }

    // shareIdentity makes into denote the same abstract value as from. Called only for Assign from a
    // mutable or context source: Alias, Capture and MaybeAlias change no identity upstream.
    shareIdentity(from: IdentifierIdType, into: IdentifierIdType): void {
        if(from === into) {
            return;
        }
        const identity = this.identities.get(from);
        if(identity === undefined) {
            return;
        }
        this.detachIdentity(into);
        this.identities.set(into, identity);
        (this.identityMembers.get(identity) ?? panic(`no members for identity ${identity}`)).add(into);
    }

    // detachIdentity removes one identifier from its previous abstract value without disturbing the
    // aliases that still denote it.
    private detachIdentity(id: IdentifierIdType): void {
        const identity = this.identities.get(id);
        if(identity === undefined) {
            return;
        }
        this.identities.delete(id);
        const members = this.identityMembers.get(identity);
        if(members === undefined) {
            return;
        }
        members.delete(id);
        if(members.size === 0) {
            this.identityMembers.delete(identity);
        }
    }

    // createFrom is the CreateFrom effect: into is a new value derived from from. Upstream creates the
    // target unconditionally first, then adds a forward alias edge from -> into and records created-from
    // on the target, without checking the source exists.
    createFrom(index: number, from: IdentifierIdType, into: IdentifierIdType): void {
        this.create(into, 'Object');
        // An if rather than a call through ?., which stage 0 can't lower yet (gap 2).
        const fromNode = this.nodes.get(from);
        if(fromNode !== undefined) {
            fromNode.edges.push({ index, node: into, kind: 'Alias' });
        }
        const toNode = this.nodes.get(into);
        if(toNode !== undefined) {
            toNode.insertBackEdge('CreatedFrom', from, index);
        }
    }

    // assign is the Assign and Alias effects: mutating into implies mutating from. Both endpoints must
    // already exist or the edge is dropped entirely, which is why the entry values are created before the
    // walk begins.
    assign(index: number, from: IdentifierIdType, into: IdentifierIdType): void {
        const fromNode = this.nodes.get(from);
        const toNode = this.nodes.get(into);
        if(fromNode === undefined || toNode === undefined) {
            return;
        }
        fromNode.edges.push({ index, node: into, kind: 'Alias' });
        toNode.insertBackEdge('Alias', from, index);
    }

    // capture is the Capture effect: information flows from from into into without aliasing. A capture
    // edge is followed backwards only by a transitive mutation.
    capture(index: number, from: IdentifierIdType, into: IdentifierIdType): void {
        const fromNode = this.nodes.get(from);
        const toNode = this.nodes.get(into);
        if(fromNode === undefined || toNode === undefined) {
            return;
        }
        fromNode.edges.push({ index, node: into, kind: 'Capture' });
        toNode.insertBackEdge('Capture', from, index);
    }

    // maybeAlias is the MaybeAlias effect: a possible aliasing relationship, from an unknown callee.
    // Traversing this edge in either direction downgrades the mutation to conditional.
    maybeAlias(index: number, from: IdentifierIdType, into: IdentifierIdType): void {
        const fromNode = this.nodes.get(from);
        const toNode = this.nodes.get(into);
        if(fromNode === undefined || toNode === undefined) {
            return;
        }
        fromNode.edges.push({ index, node: into, kind: 'MaybeAlias' });
        toNode.insertBackEdge('MaybeAlias', from, index);
    }

    /*
     * mutate walks the alias graph from one mutated value, widening every range it reaches: React
     * 42188-42280 and oxc 213-393, the only function in either implementation that writes a wider end.
     *
     * The termination is not the widening reaching a fixpoint. What bounds the walk is seen, keyed by
     * identifier and valued by mutation kind, with the test `previous >= current -> skip`. A node is
     * processed only when its kind is strictly greater than what seen holds, and processing writes that
     * kind back. The kind has two values and an edge either keeps it or downgrades it to conditional, so
     * each identifier is processed at most twice and the loop runs at most 2N times. Neutralising the
     * comparison makes the pass hang rather than answer wrongly.
     *
     * seen is per call, not per pass: a value reached by two mutations is widened once per call, which is
     * what makes end grow to the last mutation. And the comparison is on the kind alone: including the
     * direction would not terminate on a graph with an alias cycle.
     *
     * The walk skips any edge whose index is at or after the mutation's own, so a mutation only flows
     * through aliases that already existed when it happened. Upstream's forward-edge loop uses break
     * rather than continue on that test, correct only because forward edges are appended in index order.
     */
    mutate(
        ranges: MutableRanges,
        index: number,
        start: IdentifierIdType,
        end: EvaluationOrderType,
        hasEnd: boolean,
        transitive: boolean,
        startKind: MutationKindType,
    ): void {
        const seen = new Map<IdentifierIdType, MutationKindType>();
        const queue: MutationQueueEntryInterface[] = [
            { place: start, transitive, direction: 'Backwards', kind: startKind },
        ];

        for(;;) {
            const entry = queue.pop();
            if(entry === undefined) {
                break;
            }

            const previous = seen.get(entry.place);
            if(previous !== undefined && previous >= entry.kind) {
                continue;
            }
            seen.set(entry.place, entry.kind);

            const node = this.nodes.get(entry.place);
            if(node === undefined) {
                continue;
            }

            // The widening. This is the whole point of the pass, and it is a max rather than an assignment
            // because a value mutated at several instructions keeps the last one.
            if(hasEnd) {
                const existing = ranges.get(node.id);
                if(end > existing.end) {
                    ranges.set(node.id, new MutableRange(existing.start, end));
                }
            }

            // Forward edges: mutating a value mutates what it was aliased or captured into.
            for(const edge of node.edges) {
                if(edge.index >= index) {
                    // Sorted by construction; see the note on break above.
                    break;
                }
                queue.push({
                    place: edge.node,
                    transitive: entry.transitive,
                    direction: 'Forwards',
                    kind: edge.kind === 'MaybeAlias' ? mutationKindConditional : entry.kind,
                });
            }

            // createdFrom always forces transitive, whatever the incoming entry said. Mutating a value
            // derived from another reaches everything that other value transitively holds.
            for(const [alias, aliasIndex] of node.createdFrom) {
                if(aliasIndex >= index) {
                    continue;
                }
                queue.push({ place: alias, transitive: true, direction: 'Backwards', kind: entry.kind });
            }

            // Backward alias edges, suppressed for a phi reached travelling forwards.
            if(entry.direction === 'Backwards' || node.value !== 'Phi') {
                for(const [alias, aliasIndex] of node.aliases) {
                    if(aliasIndex >= index) {
                        continue;
                    }
                    queue.push({
                        place: alias,
                        transitive: entry.transitive,
                        direction: 'Backwards',
                        kind: entry.kind,
                    });
                }
                for(const [alias, aliasIndex] of node.maybeAliases) {
                    if(aliasIndex >= index) {
                        continue;
                    }
                    queue.push({
                        place: alias,
                        transitive: entry.transitive,
                        direction: 'Backwards',
                        kind: mutationKindConditional,
                    });
                }
            }

            // Captures are followed backwards only by a transitive mutation. Mutating a container does not
            // mutate what it holds; mutating it transitively does.
            if(entry.transitive) {
                for(const [capture, captureIndex] of node.captures) {
                    if(captureIndex >= index) {
                        continue;
                    }
                    queue.push({
                        place: capture,
                        transitive: entry.transitive,
                        direction: 'Backwards',
                        kind: entry.kind,
                    });
                }
            }
        }
    }
}

// MutationDirectionType (Go's mutationDirection) is whether the walk arrived at a node travelling
// forwards or backwards. Upstream's condition is `direction === 'backwards' || node.value.kind !== 'Phi'`.
type MutationDirectionType = 'Backwards' | 'Forwards';

// MutationQueueEntryInterface (Go's mutationQueueEntry) is one pending visit in the worklist.
interface MutationQueueEntryInterface {
    readonly place: IdentifierIdType;
    readonly transitive: boolean;
    readonly direction: MutationDirectionType;
    readonly kind: MutationKindType;
}

/*
 * derivePhiImmutable joins the known kinds after every predecessor has been evaluated.
 *
 * Upstream's InferenceState.inferPhi maps the phi to the union of the values its operands name, so a
 * conditional mutation is later checked against their joined kind. Frozen mixed with Mutable yields
 * MaybeFrozen, which suppresses conditional mutation but keeps Assign and CreateFrom identity and alias
 * edges. A back edge or unknown operand leaves the phi mutable: this pass is single-shot, and treating an
 * incomplete union as frozen would be the unsound direction.
 */
function derivePhiImmutable<F, B, P>(
    state: AliasingState,
    graph: GraphInterface<F, B, P>,
    phi: PhiInterface<P>,
    seenBlocks: Set<BlockIdType>,
): void {
    if(phi.operands.length === 0) {
        return;
    }
    let kind: EffectValueKindType = 'Primitive';
    let hasFrozen = false;
    let hasMutable = false;
    for(const operand of phi.operands) {
        if(!seenBlocks.has(operand.predecessor)) {
            return;
        }
        const operandKind = state.immutable.get(graph.identifierOf(operand.place));
        if(operandKind === undefined) {
            hasMutable = true;
            continue;
        }
        // Ifs rather than ||=, which Adamic 0.1 refuses.
        if(operandKind === 'Frozen' || operandKind === 'MaybeFrozen') {
            hasFrozen = true;
        }
        if(operandKind === 'MaybeFrozen') {
            hasMutable = true;
        }
        if(operandKind === 'Frozen' || operandKind === 'MaybeFrozen') {
            kind = operandKind;
        }
        else if(operandKind === 'Global' && kind === 'Primitive') {
            kind = 'Global';
        }
    }
    if(hasMutable) {
        if(!hasFrozen) {
            return;
        }
        kind = 'MaybeFrozen';
    }
    state.markImmutable(graph.identifierOf(phi.place), kind);
}

/*
 * PendingMutationInterface (Go's pendingMutation) is one mutation deferred until the whole graph is
 * built. Upstream collects every mutation during the graph walk and applies them all afterwards: a
 * mutation applied while the graph is half-built would traverse only the edges discovered so far.
 */
export interface PendingMutationInterface {
    readonly index: number;
    readonly end: EvaluationOrderType;
    readonly transitive: boolean;
    readonly kind: MutationKindType;
    readonly place: IdentifierIdType;
}

// PendingPhiOperandInterface is a phi operand whose predecessor hadn't been walked when its phi was,
// applied once it has, with the index it was given then.
interface PendingPhiOperandInterface {
    readonly from: IdentifierIdType;
    readonly into: IdentifierIdType;
    readonly index: number;
}

const noPendingPhiOperands: readonly PendingPhiOperandInterface[] = [];

// AliasingGraph (Go's AliasingGraph) is the alias graph of one function, built and not yet applied: the
// state the walk ended in, and the mutations it collected.
export class AliasingGraph {
    readonly state: AliasingState;
    readonly mutations: readonly PendingMutationInterface[];

    constructor(state: AliasingState, mutations: readonly PendingMutationInterface[]) {
        this.state = state;
        this.mutations = mutations;
    }

    // kind is the abstract kind the walk ended with for a value: an immutable kind, or Mutable for a value
    // it holds none for, which is also what a value the walk never reached reads as.
    kind(id: IdentifierIdType): EffectValueKindType {
        return this.state.immutable.get(id) ?? 'Mutable';
    }

    // mutationCount is how many mutations the walk kept, after the ones upstream drops (a conditional
    // mutation of a value that is not Mutable or Context).
    mutationCount(): number {
        return this.mutations.length;
    }

    /*
     * widen (Go's Widen) is upstream's Part 1 applied: every mutation the graph collected, walked through
     * it, widening the ranges in result.
     *
     * A mutation is applied only when its instruction carries an evaluation order. end is order + 1, so 0
     * can't come from a numbered instruction, and 1 means the order was 0: finalize never reached the
     * block, and widening there would make a value read as mutable at position 0 in a function where
     * nothing has an order at all.
     */
    widen(result: MutableRanges): void {
        for(const mutation of this.mutations) {
            if(mutation.end === 0 || mutation.end === 1) {
                continue;
            }
            this.state.mutate(
                result,
                mutation.index,
                mutation.place,
                mutation.end,
                true,
                mutation.transitive,
                mutation.kind,
            );
        }
    }
}

// isImmutableEnd reports whether a kind makes an Alias or Capture a range no-op at either end: Frozen,
// Primitive or Global.
function isImmutableEnd(kind: EffectValueKindType | undefined): boolean {
    return kind === 'Frozen' || kind === 'Primitive' || kind === 'Global';
}

/*
 * buildAliasingGraph (Go's BuildAliasingGraph) constructs the data-flow graph and returns it with the
 * mutations to apply.
 *
 * This is upstream's Part 1, at React 41736-41850 and oxc 473-645. The walk is over blocks in the order
 * the function's block array holds them, which is reverse postorder, matching upstream's iteration over
 * its block map.
 *
 * The sequence index is the pass's model of time and it is incremented exactly where upstream does.
 * Every edge-creating effect and every mutation takes the current index and increments it. Effects that
 * create a node without an edge (Create) do not increment, and neither do the effects this pass treats as
 * no-ops. Getting those increments wrong shifts every later index, which changes only which edges a
 * mutation can see: a range table subtly narrow or wide with no invariant violated.
 */
export function buildAliasingGraph<F, B, P>(
    graph: GraphInterface<F, B, P>,
    fn: F,
    options: OptionsInterface,
): AliasingGraph {
    const state = new AliasingState();
    const mutations: PendingMutationInterface[] = [];
    let index = 0;

    // Entry values first. Upstream creates nodes for params, context and the return value before walking
    // any block, and the assign, capture and maybeAlias builders drop an edge whose endpoints do not both
    // exist. graph.parametersFrozen is upstream's parameter kind: a component's or hook's parameters are
    // Frozen, since React owns the arguments it passes them, and a nested function expression's Mutable.
    const parametersAreFrozen = graph.parametersFrozen(fn);
    for(const param of graph.params(fn)) {
        const id = graph.identifierOf(param);
        state.create(id, 'Object');
        if(parametersAreFrozen) {
            state.markImmutable(id, 'Frozen');
        }
    }
    for(const contextValue of graph.context(fn)) {
        const id = graph.identifierOf(contextValue);
        state.create(id, 'Object');
        const kind = options.contextKinds.get(id);
        if(kind !== undefined && kind !== 'Mutable') {
            state.markImmutable(id, kind);
        }
    }
    // The return value is an entry value too: upstream creates its node beside the params and context,
    // and adds an alias edge into it at every return terminal. An IR with no returns place (Adamic's) has
    // neither.
    const returns = graph.returns(fn);
    if(returns !== undefined) {
        state.create(graph.identifierOf(returns), 'Object');
    }

    const seenBlocks = new Set<BlockIdType>();
    const pendingPhis = new Map<BlockIdType, PendingPhiOperandInterface[]>();

    for(const block of graph.blocks(fn)) {
        for(const phi of graph.phis(block)) {
            const phiId = graph.identifierOf(phi.place);
            state.create(phiId, 'Phi');
            // A loop rather than a map, since a one-parameter function capturing graph could be seen as
            // its identifierOf, and stage 0's cycle rule refuses it (`GAPS.md`, the cycle rule's cost).
            const operands: IdentifierIdType[] = [];
            for(const operand of phi.operands) {
                operands.push(graph.identifierOf(operand.place));
            }
            state.recordFreezeSources(phiId, operands);
            derivePhiImmutable(state, graph, phi, seenBlocks);
            // Ascending predecessor order, the order a phi's operands are kept in. Assigning indices in an
            // order that varied would change which edges a mutation can see.
            for(let position = 0; position < phi.operands.length; position++) {
                const operand = phi.operands[position] ?? panic('a phi operand past its end');
                const from = operands[position] ?? panic('a phi operand past its end');
                if(!seenBlocks.has(operand.predecessor)) {
                    // A back edge: the predecessor has not been walked yet, so the operand's node may not
                    // exist. Upstream defers these to the predecessor block and still consumes an index
                    // here, so the deferral does not shift the numbering.
                    const pending = pendingPhis.get(operand.predecessor);
                    if(pending === undefined) {
                        pendingPhis.set(operand.predecessor, [{ from, into: phiId, index }]);
                    }
                    else {
                        pending.push({ from, into: phiId, index });
                    }
                }
                else {
                    state.assign(index, from, phiId);
                }
                index++;
            }
        }
        const blockId = graph.id(block);
        seenBlocks.add(blockId);

        const count = graph.instructionCount(fn, block);
        for(let instructionIndex = 0; instructionIndex < count; instructionIndex++) {
            const order = graph.instructionOrder(fn, block, instructionIndex);
            if(order === undefined) {
                continue;
            }
            for(const effect of graph.effects(fn, block, instructionIndex)) {
                // Go reads From's identifier for every effect, and an effect without one has a zero place,
                // identifier 0. Only the flow kinds below read it.
                const from = effect.from === undefined ? 0 : graph.identifierOf(effect.from);
                const into = graph.identifierOf(effect.into);
                switch(effect.kind) {
                    case 'Freeze':
                        // InferenceState.freeze changes the abstract kind immediately, and calls later in
                        // evaluation order consult that kind before keeping a conditional mutation, so this
                        // is not a graph-only no-op even though Freeze itself widens nothing.
                        state.freeze(into);
                        break;

                    case 'Create': {
                        state.create(into, 'Object');
                        // React's rule for closures, GraphInterface.closure: a closure whose captures are
                        // all immutable, and whose body only reads them, is created Frozen. Its captures
                        // are what a later Freeze of it freezes.
                        let value = effect.value;
                        const closure = graph.closure(fn, block, instructionIndex, into, state.immutable);
                        if(closure !== undefined) {
                            // A loop rather than a map, as for a phi's operands above.
                            const sources: IdentifierIdType[] = [];
                            for(const capture of closure.captures) {
                                sources.push(graph.identifierOf(capture));
                            }
                            state.recordFreezeSources(into, sources);
                            if(closure.frozen) {
                                value = 'Frozen';
                            }
                        }
                        if(value !== 'Mutable') {
                            state.markImmutable(into, value);
                        }
                        else {
                            state.immutable.delete(into);
                        }
                        break;
                    }

                    case 'CreateFrom': {
                        const kind = state.immutable.get(from);
                        if(kind !== undefined && kind !== 'MaybeFrozen') {
                            state.create(into, 'Object');
                        }
                        else {
                            state.createFrom(index, from, into);
                        }
                        state.deriveImmutable(from, into);
                        index++;
                        break;
                    }

                    case 'Assign': {
                        // Upstream creates the target if it is absent, which Alias does not do.
                        if(!state.nodes.has(into)) {
                            state.create(into, 'Object');
                        }
                        const sourceIsMutable = !state.notMutable(from) || state.immutable.get(from) === 'MaybeFrozen';
                        // An assignment carries mutability with it, the same way CreateFrom does: upstream's
                        // Assign arm switches on the source's kind, so a frozen value assigned into a
                        // temporary stays frozen.
                        state.deriveImmutable(from, into);
                        if(sourceIsMutable) {
                            state.shareIdentity(from, into);
                            state.assign(index, from, into);
                        }
                        index++;
                        break;
                    }

                    case 'Alias':
                        // Alias and Capture share upstream's refinement. A frozen, primitive or global source
                        // becomes an ImmutableCapture (a range no-op), and a mutable source cannot flow into
                        // a frozen, primitive or global destination.
                        if(!isImmutableEnd(state.immutable.get(from)) && !isImmutableEnd(state.immutable.get(into))) {
                            state.assign(index, from, into);
                        }
                        index++;
                        break;

                    case 'MaybeAlias':
                        // MaybeAlias survives primitive and global sources and immutable destinations, but a
                        // Frozen source is rewritten to ImmutableCapture. MaybeFrozen keeps the edge.
                        if(state.immutable.get(from) !== 'Frozen') {
                            state.maybeAlias(index, from, into);
                        }
                        index++;
                        break;

                    case 'Capture':
                        /*
                         * Upstream switches on the source's kind and only two of its four arms reach
                         * state.capture: Frozen is re-applied as ImmutableCapture, Global and Primitive
                         * are pruned, Context is re-applied as MaybeAlias (represented here by the
                         * conservative Mutable seed), and Mutable or MaybeFrozen is kept for a mutable
                         * destination. A destination that is Frozen, Global or Primitive prunes it too.
                         */
                        if(!isImmutableEnd(state.immutable.get(from)) && !isImmutableEnd(state.immutable.get(into))) {
                            state.capture(index, from, into);
                        }
                        index++;
                        break;

                    case 'MutateTransitive':
                    case 'MutateTransitiveConditionally': {
                        const kind =
                            effect.kind === 'MutateTransitiveConditionally'
                                ? mutationKindConditional
                                : mutationKindDefinite;
                        // Upstream's state.mutate returns none for a conditional mutation of anything that
                        // is not Mutable or Context, and the effect isn't pushed. The index still advances:
                        // a dropped effect must not shift the numbering of the ones that remain.
                        if(kind === mutationKindDefinite || !state.notMutable(into)) {
                            mutations.push({ index, end: order + 1, transitive: true, kind, place: into });
                        }
                        index++;
                        break;
                    }

                    case 'Mutate':
                        mutations.push({
                            index,
                            end: order + 1,
                            transitive: false,
                            kind: mutationKindDefinite,
                            place: into,
                        });
                        index++;
                        break;

                    case 'MutateConditionally':
                        if(!state.notMutable(into)) {
                            mutations.push({
                                index,
                                end: order + 1,
                                transitive: false,
                                kind: mutationKindConditional,
                                place: into,
                            });
                        }
                        index++;
                        break;

                    case 'ImmutableCapture':
                    case 'Apply':
                        // No-ops for range widening, and they consume no index: upstream's `_ => {}` arm,
                        // reproduced deliberately, since an effect reaching here must not shift the
                        // numbering.
                        break;
                }
            }
        }

        // Pending phi operands whose predecessor is this block, now that its values exist. Upstream applies
        // them with the index recorded at deferral rather than a fresh one.
        for(const pending of pendingPhis.get(blockId) ?? noPendingPhiOperands) {
            state.assign(pending.index, pending.from, pending.into);
        }
        pendingPhis.delete(blockId);

        // A return aliases its operand into the function's return value. Upstream does this from the
        // terminal rather than from an effect, because a return carries no instruction effects.
        if(returns !== undefined) {
            const value = graph.returnValue(block);
            if(value !== undefined) {
                state.assign(index, graph.identifierOf(value), graph.identifierOf(returns));
                index++;
            }
        }
    }

    return new AliasingGraph(state, mutations);
}

/*
 * inferMutableRanges (Go's InferMutableRanges) computes the mutable range of every value in fn,
 * returning them as a table keyed by value.
 *
 * Requires evaluation order, which an IR's finalize assigns and its construct re-establishes. Over a
 * graph whose instruction orders are zero every range would open at zero, which is the unset value, so a
 * caller who skips it gets an empty table rather than a wrong one.
 *
 * Nested functions are not included. Upstream analyses a nested function separately and resets every
 * context operand's range to unset afterwards, so ranges are per-function by upstream's own
 * construction, and a table spanning two functions would key two different values under one id.
 *
 * Idempotent: every guard tests the range this pass itself wrote, so a second run over the same function
 * produces an equal table.
 *
 * The order of the two halves is upstream's and it is load-bearing. Every definition-half write is
 * guarded on the field being unset: the lvalue loop opens start only if start is 0 and end only if end is
 * 0. So a value whose end was already widened past its definition point must not have that end
 * overwritten by order + 1, and running the definition half first would do exactly that for any value
 * mutated before its own defining instruction in evaluation order, which single-assignment form makes
 * ordinary through phis and loops.
 */
export function inferMutableRanges<F, B, P>(
    graph: GraphInterface<F, B, P>,
    fn: F,
    options: OptionsInterface,
): MutableRanges {
    const result = new MutableRanges();

    // Part 1: build the alias graph and widen every range a mutation reaches.
    buildAliasingGraph(graph, fn, options).widen(result);

    // Part 2: open a range at every definition point that the widening left unopened.
    //
    // One visitor for every place, made once per call, which reads what to do from the walk.
    const walk = new DefinitionWalk(graph, result);
    const visit = function(place: P, role: RoleType): P {
        walk.visitPlace(place, role);
        return place;
    };
    for(const block of graph.blocks(fn)) {
        // A phi's range opens at the block's first instruction minus one, but only when the phi is
        // mutated after creation, which only the widening can establish. See phiOpensBefore.
        const firstOrder = blockFirstOrder(graph, fn, block);
        for(const phi of graph.phis(block)) {
            const id = graph.identifierOf(phi.place);
            const opened = phiOpenedRange(result.get(id), firstOrder);
            if(opened !== undefined) {
                result.set(id, opened);
            }
        }

        const count = graph.instructionCount(fn, block);
        for(let index = 0; index < count; index++) {
            const order = graph.instructionOrder(fn, block, index);
            if(order === undefined) {
                continue;
            }
            if(order === 0) {
                // Unnumbered, so finalize never reached this block. Opening a range at zero would write
                // the unset value and read back as "no range", which is a wrong answer wearing the shape
                // of a right one. Skipping is the honest response.
                continue;
            }
            walk.order = order;

            // Every lvalue opens its own range here. This is upstream's unconditional loop over
            // `eachInstructionLValue` and it reads no effect: React lines 41916-41927, oxc 805-834,
            // reached through the place visitor and the Define role rather than by matching instruction
            // variants. DefinitionWalk.openLValue is the body.
            walk.lvalues = true;
            graph.eachInstructionPlace(fn, block, index, visit);

            /*
             * The third loop, over operands rather than lvalues: React 41994-41998, oxc 918-926. It opens a
             * range for any operand whose end is already past this instruction but whose start was never
             * opened. The widening writes end on values this function's instructions never define as an
             * lvalue (a parameter, a context value, the return value), and this loop is the only thing
             * that gives those values a start.
             */
            walk.lvalues = false;
            graph.eachInstructionPlace(fn, block, index, visit);

            // StoreContext widens the stored value's end from the instruction shape alone, with no effect
            // consulted: React 42002-42004, oxc 960-967. A store into a captured binding is a mutation the
            // instruction itself proves: the value is written into a cell an enclosing scope can observe.
            // An IR with no captured variables in its graph has no such store (Adamic's).
            const stored = graph.storedContextValue(fn, block, index);
            if(stored !== undefined) {
                const existing = result.get(stored);
                if(existing.end <= order) {
                    result.set(stored, new MutableRange(existing.start, order + 1));
                }
            }
        }
    }

    // OptionsInterface.parametersDefinedOnEntry, Adamic's rule, after both halves: a parameter is defined
    // on entry, before the first instruction, not where it's first read, which is where the operand loop
    // above opens a range nothing defines. So its range starts at the function's first instruction, and
    // one nothing widened is that instruction alone.
    if(options.parametersDefinedOnEntry) {
        for(const parameter of graph.params(fn)) {
            const id = graph.identifierOf(parameter);
            const existing = result.get(id);
            result.set(id, new MutableRange(1, existing.end === 0 ? 2 : existing.end));
        }
    }

    return result;
}
